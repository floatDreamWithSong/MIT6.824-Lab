package mr

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Master struct {
	// Your definitions here.
	files                 []string
	nReduce               int
	mapTasks              []Task
	reduceTasks           []Task
	completedTaskNum      int
	completedTaskNumMutex sync.Mutex
	mu                    sync.Mutex
}

type TaskStatus int

const (
	NOT_STARTED TaskStatus = iota
	IN_PROGRESS
	COMPLETED
)

// Your code here -- RPC handlers for the worker to call.
type Task struct {
	RequestTaskReply
	status TaskStatus
}

func (m *Master) checkTaskStatus() {
	// 持续遍历任务队列，对于处于IN_PROGRESS的任务，检查任务是否超时(10s)，如果超时，将任务状态修改为NOT_STARTED，然后重新分配任务

	for {
		time.Sleep(500 * time.Millisecond)
		m.mu.Lock()
		for i, task := range m.mapTasks {
			if task.status == IN_PROGRESS && time.Since(time.Unix(task.StartTime, 0)) > 10*time.Second {
				log.Default().Printf("map task %d timeout", task.TaskId)
				m.mapTasks[i].status = NOT_STARTED
			}
		}
		for i, task := range m.reduceTasks {
			if task.status == IN_PROGRESS && time.Since(time.Unix(task.StartTime, 0)) > 10*time.Second {
				log.Default().Printf("reduce task %d timeout", task.TaskId)
				m.reduceTasks[i].status = NOT_STARTED
			}
		}
		m.mu.Unlock()
	}
}

// an example RPC handler.
//
// the RPC argument and reply types are defined in rpc.go.
func (m *Master) Example(args *ExampleArgs, reply *ExampleReply) error {
	reply.Y = args.X + 1
	return nil
}

func (m *Master) HandleRequestTask(args *RequestTaskArgs, reply *RequestTaskReply) error {
	// 如果有待处理的map任务，线程安全地将任务状态修改为IN_PROGRESS，然后返回任务
	m.mu.Lock() // 锁定互斥锁
	defer func() {
		m.mu.Unlock() // 确保在函数结束时解锁
		// 输出reply的内容
		fmt.Printf("reply: %+v\n", reply)
	}()
	// 若有map任务处于NOT_STARTED状态，将其状态修改为IN_PROGRESS，并返回任务
	for index, task := range m.mapTasks {
		if task.status == NOT_STARTED {
			task.status = IN_PROGRESS
			// 记录任务开始时间
			startTime := time.Now().Unix()
			task.StartTime = startTime
			m.mapTasks[index] = task
			*reply = task.RequestTaskReply
			return nil
		}
	}
	// 若没有map任务还有处于IN_PROGRESS的任务，让worker处于等待
	for _, task := range m.mapTasks {
		if task.status == IN_PROGRESS {
			reply.TaskType = WAIT
			reply.TaskId = -1
			return nil
		}
	}
	// 既没有待处理的任务，也没有处于IN_PROGRESS的map任务，就进行reduce任务的分配
	for index, task := range m.reduceTasks {
		if task.status == NOT_STARTED {
			task.status = IN_PROGRESS
			startTime := time.Now().Unix()
			task.StartTime = startTime
			m.reduceTasks[index] = task
			*reply = task.RequestTaskReply
			return nil
		}
	}
	// 如果还有未处理完的reduce任务，让worker处于等待
	for _, task := range m.reduceTasks {
		if task.status == IN_PROGRESS {
			reply.TaskType = WAIT
			reply.TaskId = -1
			return nil
		}
	}
	// 否则所有任务都处理完了，返回CLOSE
	reply.TaskType = CLOSE
	reply.TaskId = -2
	return nil
}

func (m *Master) HandleSubmitTask(args *SubmitTaskArgs, reply *SubmitTaskReply) error {
	m.mu.Lock()
	defer func() {
		m.mu.Unlock()
		fmt.Printf("receive: %+v\n", args)
	}()
	// 如果是Map任务，检查任务时间戳，将任务状态修改为COMPLETED，并根据work生成的中间文件的末尾数字，将其加入到对应reduceId的文件列表中,中间文件的格式为mr-<mapId>-<reduceId>
	if args.TaskType == MAP {
		for index, task := range m.mapTasks {
			if task.TaskId == args.TaskId && task.StartTime == args.StartTime {
				m.mapTasks[index].status = COMPLETED
				for _, filename := range args.Filename {
					// 解析文件名，以-为分隔符，获取reduceId
					parts := strings.Split(filename, "-")
					if len(parts) != 3 {
						log.Printf("Invalid intermediate file name: %s", filename)
						continue
					}
					reduceId, err := strconv.Atoi(parts[2])
					if err != nil || reduceId >= m.nReduce {
						log.Printf("Failed to parse reduce ID from filename: %s", filename)
						continue
					}
					// 将文件名加入到对应reduceId的文件列表中
					m.reduceTasks[reduceId].Filename = append(m.reduceTasks[reduceId].Filename, filename)
				}
				break
			}
		}
	} else if args.TaskType == REDUCE {
		for index, task := range m.reduceTasks {
			if task.TaskId == args.TaskId && task.StartTime == args.StartTime {
				m.reduceTasks[index].status = COMPLETED
				m.completedTaskNumMutex.Lock()
				m.completedTaskNum++
				m.completedTaskNumMutex.Unlock()
				break
			}
		}
	}
	return nil
}

// start a thread that listens for RPCs from worker.go
// 套接字监听
func (m *Master) server() {
	rpc.Register(m)
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	sockname := masterSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	fmt.Println("master start at ", sockname)
	go http.Serve(l, nil)
	go m.checkTaskStatus()
}

// main/mrmaster.go calls Done() periodically to find out
// if the entire job has finished.
// 这个会被定期调用用来检验工作是否完成
func (m *Master) Done() bool {
	m.completedTaskNumMutex.Lock()
	defer m.completedTaskNumMutex.Unlock()
	// Your code here.
	return m.completedTaskNum >= len(m.files)
}

// create a Master.
// main/mrmaster.go calls this function.
// nReduce is the number of reduce tasks to use.
func MakeMaster(files []string, nReduce int) *Master {
	m := Master{}

	// Your code here.
	// 初始化Master
	fmt.Println("receive files:")
	for _, file := range files {
		fmt.Println(file)
	}
	m.files = files
	fmt.Println("receive nReduce:", nReduce)
	m.nReduce = nReduce

	// 初始化map任务
	for index, file := range files {
		task := Task{
			RequestTaskReply: RequestTaskReply{},
		}
		task.TaskType = MAP
		task.TaskId = index
		task.Filename = append(task.Filename, file)
		task.status = NOT_STARTED
		task.NReduce = nReduce
		m.mapTasks = append(m.mapTasks, task)
	}
	// 初始化reduce任务
	for index := 0; index < nReduce; index++ {
		task := Task{
			RequestTaskReply: RequestTaskReply{},
		}

		task.TaskType = REDUCE
		task.TaskId = index
		task.status = NOT_STARTED
		task.NReduce = nReduce
		m.reduceTasks = append(m.reduceTasks, task)
	}
	m.server()
	return &m
}
