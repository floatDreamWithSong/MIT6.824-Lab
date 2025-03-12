package mr

//
// RPC definitions.
//
// remember to capitalize all names.
//

import (
	"os"
	"strconv"
)

//
// example to show how to declare the arguments
// and reply for an RPC.
//

type ExampleArgs struct {
	X int64
}

type ExampleReply struct {
	Y int64
}

// Add your RPC definitions here.
// 在此处添加RPC代码

type TaskType int

const (
	MAP TaskType = iota
	REDUCE
	WAIT
	CLOSE
)

type RequestTaskArgs struct{}

type RequestTaskReply struct {
	TaskType  TaskType
	Filename  []string
	TaskId    int
	NReduce   int
	StartTime int64
}

type SubmitTaskArgs struct {
	TaskType  TaskType
	Filename  []string
	TaskId    int
	StartTime int64
}

type SubmitTaskReply struct{}

// Cook up a unique-ish UNIX-domain socket name
// in /var/tmp, for the master.
// Can't use the current directory since
// Athena AFS doesn't support UNIX-domain sockets.
func masterSock() string {
	s := "/var/tmp/824-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}
