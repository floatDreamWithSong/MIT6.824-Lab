package mr

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"net/rpc"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"
)

// Map functions return a slice of KeyValue.
type KeyValue struct {
	Key   string
	Value string
}

// use ihash(key) % NReduce to choose the reduce
// task number for each KeyValue emitted by Map.
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

type ByKey []KeyValue

func (a ByKey) Len() int           { return len(a) }
func (a ByKey) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByKey) Less(i, j int) bool { return a[i].Key < a[j].Key }

// main/mrworker.go calls this function.
func Worker(mapf func(string, string) []KeyValue,
	reducef func(string, []string) string) {

	// Your worker implementation here.
	// 此处添加worker实现
	// uncomment to send the Example RPC to the master.
	// CallExample()
	for {
		args := RequestTaskArgs{}
		reply := RequestTaskReply{}
		call("Master.HandleRequestTask", &args, &reply)
		// 打印出任务信息
		fmt.Printf("TaskType: %v, Filename: %v, TaskId: %v, NReduce: %v\n", reply.TaskType, reply.Filename, reply.TaskId, reply.NReduce)
		// 若任务类型为MAP，说明有任务需要处理
		if reply.TaskType == WAIT {
			time.Sleep(1 * time.Second)
		} else if reply.TaskType == MAP {
			// map任务: 读取文件，将文件内容传入mapf，获取K/V对，遍历K/V对数组，根据K值映射到reduce任务，reduce任务的编号为ihash%nReduce
			// 此外还要用map记录每个编号对应的K/V对数组所生成的临时文件TempFile。
			// 然后遍历map，重命名为真正的文件，其中格式为mr-tmp-taskId-reduceTaskId，通过ihash%nReduce选择reduce任务，将K/V对Json化写入TempFile
			// 全部写入成功后，再Rename创建真正的文件，然后上报给master
			intermediate := []KeyValue{}
			for _, filename := range reply.Filename {
				// 根据文件名获取文件对象
				file, err := os.Open(filename)
				if err != nil {
					log.Fatalf("cannot open %v", filename)
				}
				// 读取文件对象的内容
				content, err := ioutil.ReadAll(file)
				if err != nil {
					log.Fatalf("cannot read %v", filename)
				}
				// 释放
				file.Close()
				// 传入文件名，内容字符，获取K/V对
				kva := mapf(filename, string(content))
				// 解构扩展K/V中间体
				intermediate = append(intermediate, kva...)
			}
			// sort.Sort(ByKey(intermediate))
			kvpair_map := make(map[int][]KeyValue)
			// 选择reduce任务
			for _, kv := range intermediate {
				// 选择reduce任务
				reduceTaskId := ihash(kv.Key) % reply.NReduce
				// 写入map
				kvpair_map[reduceTaskId] = append(kvpair_map[reduceTaskId], kv)
			}
			// 遍历map，生成中间文件
			// 生成reduce任务编号对应TempFile的map
			reduceTaskId_TempFile := make(map[int]*os.File)
			for reduceTaskId, kvpair := range kvpair_map {
				sort.Sort(ByKey(kvpair))
				// 生成中间文件
				file, err := ioutil.TempFile("./", "mr-tmp-*")
				defer file.Close()
				if err != nil {
					log.Fatalf("cannot create %v", file.Name())
				}
				// json
				enc := json.NewEncoder(file)
				for _, kv := range kvpair {
					err := enc.Encode(&kv)
					if err != nil {
						log.Fatalf("cannot encode %v", kv)
					}
				}
				// 记录reduce任务编号对应TempFile
				reduceTaskId_TempFile[reduceTaskId] = file
			}
			// 遍历map，重命名为真正的文件
			for reduceTaskId, file := range reduceTaskId_TempFile {
				// 重命名
				filename := "mr-" + strconv.Itoa(reply.TaskId) + "-" + strconv.Itoa(reduceTaskId)
				err := os.Rename(filepath.Join(file.Name()), filename)
				if err != nil {
					log.Fatalf("cannot rename %v", file.Name())
				}
				file.Close()
			}
		}
	}

}

// example function to show how to make an RPC call to the master.
//
// the RPC argument and reply types are defined in rpc.go.
func CallExample() {

	// declare an argument structure.
	args := ExampleArgs{}

	// fill in the argument(s).
	args.X = 99

	// declare a reply structure.
	reply := ExampleReply{}

	// send the RPC request, wait for the reply.
	call("Master.Example", &args, &reply)

	// reply.Y should be 100.
	fmt.Printf("reply.Y %v\n", reply.Y)
}

// send an RPC request to the master, wait for the response.
// usually returns true.
// returns false if something goes wrong.
func call(rpcname string, args interface{}, reply interface{}) bool {
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
	sockname := masterSock()
	c, err := rpc.DialHTTP("unix", sockname)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	fmt.Println(err)
	return false
}
