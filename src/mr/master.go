package mr

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sync"
)

type Master struct {
	// Your definitions here.
	files            []string
	nReduce          int
	mapTasks         []Task
	reduceTasks      []Task
	completedTaskNum int
	mu               sync.Mutex
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
	startTime int64
	status    TaskStatus
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
	m.mu.Lock()         // 锁定互斥锁
	defer func(){
		m.mu.Unlock() // 确保在函数结束时解锁
		// 输出reply的内容
		fmt.Printf("reply: %+v\n", reply)
	}()
	// 若有map任务处于NOT_STARTED状态，将其状态修改为IN_PROGRESS，并返回任务
	for index, task := range m.mapTasks {
		if task.status == NOT_STARTED {
			task.status = IN_PROGRESS
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
	// 既没有待处理的任务，也没有处于IN_PROGRESS的任务，就进行reduce任务的分配
	for index, task := range m.reduceTasks {
		if task.status == NOT_STARTED {
			task.status = IN_PROGRESS
			m.reduceTasks[index] = task
			*reply = task.RequestTaskReply
			return nil
		}
	}
	return nil
}

func (m *Master) HandleSubmitTask(args *SubmitTaskArgs, reply *SubmitTaskReply) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	// 如果是Map任务，将任务状态修改为COMPLETED,生成reduce任务，并将work生成的中间文件加入到files中
	if args.TaskType == MAP {
		for index, task := range m.mapTasks {
			if task.TaskId == args.TaskId {
				m.mapTasks[index].status = COMPLETED
				reduceTask := Task{
					RequestTaskReply: RequestTaskReply{
						TaskType: REDUCE,
						Filename: args.Filename,
						TaskId:   args.TaskId,
						NReduce:  m.nReduce,
					},
					status: NOT_STARTED,
				}
				m.reduceTasks = append(m.reduceTasks, reduceTask)
				break
			}
		}
	} else if args.TaskType == REDUCE {
		for index, task := range m.reduceTasks {
			if task.TaskId == args.TaskId {
				m.reduceTasks[index].status = COMPLETED
				m.completedTaskNum++
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
}

// main/mrmaster.go calls Done() periodically to find out
// if the entire job has finished.
// 这个会被定期调用用来检验工作是否完成
func (m *Master) Done() bool {
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
			startTime:        0,
		}
		task.TaskType = MAP
		task.TaskId = index
		task.Filename = append(task.Filename, file)
		task.status = NOT_STARTED
		task.NReduce = nReduce
		m.mapTasks = append(m.mapTasks, task)
	}

	m.server()
	return &m
}
