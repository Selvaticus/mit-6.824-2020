package mr

//
// RPC definitions.
//
// remember to capitalize all names.
//

import "os"
import "strconv"

//
// example to show how to declare the arguments
// and reply for an RPC.
//

type ExampleArgs struct {
	X int
}

type ExampleReply struct {
	Y int
}

// Add your RPC definitions here.

const (
	Map_Task = 1
	Reduce_Task = 2
	No_Task = 0
)

type GetTaskArgs struct {
}

type GetTaskReply struct {
	TaskId int
	Filename string
	AllDone bool
	TaskType int
	Reducers int
	BucketId int
}

type ReportOnMapJobArgs struct {
	TaskId int
	Status int
}

type ReportOnMapJobReply struct {
}

type ReportOnReduceJobArgs struct {
	TaskId int
	Status int
}

type ReportOnReduceJobReply struct {
	TaskId int
	Status int
}


// Cook up a unique-ish UNIX-domain socket name
// in /var/tmp, for the master.
// Can't use the current directory since
// Athena AFS doesn't support UNIX-domain sockets.
func masterSock() string {
	s := "/var/tmp/824-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}