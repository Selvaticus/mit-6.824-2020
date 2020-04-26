package mr

import "log"
import "net"
import "os"
import "net/rpc"
import "net/http"
import "sync"


type Master struct {
	// Your definitions here.
	mapTasks []MapTask
	reduceTasks []ReduceTask
	reducers int
	nFiles int

	mu sync.Mutex
}

type MapTask struct {
	taskId int
	filename string
	// 0 will be not started
	// 1 running
	// 2 done
	stage int
}

type ReduceTask struct {
	taskId int
	taskBucket int
	// 0 will be not started
	// 1 running
	// 2 done
	stage int
}

type TaskStage struct {
	
	stage int
}

// Your code here -- RPC handlers for the worker to call.
func (m *Master) GetTask(args *GetTaskArgs, reply *GetTaskReply) error {
	// Make sure only one RPC call is mutating the Master struct
	PrintDebug("Received request for job")
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.mapPhaseDone() {
		PrintDebug("Map phase not done yet...")
		// Send map task
		for _, v := range(m.mapTasks) {
			if v.stage == 0 {
				reply.Filename = v.filename
				reply.TaskId = v.taskId
				reply.AllDone = false
				reply.IsMap = true
				reply.Reducers = m.reducers		

				// Marked as running, aka assigned to worker
				v.stage = 1
				PrintDebugf("Sending job: %+v", reply)
				return nil
			}
		}
	} else if !m.reducePhaseDone() {
		PrintDebug("Reduce phase not done yet...")
		// Send reduce task
		for _, v := range(m.reduceTasks) {
			if v.stage == 0 {
				reply.BucketId = v.taskBucket
				reply.TaskId = v.taskId
				reply.AllDone = false
				reply.IsMap = false

				// Marked as running, aka assigned to worker
				v.stage = 1

				PrintDebugf("Sending job: %+v", reply)
				return nil
			}
		}
	} else {
		PrintDebug("All done.")
		reply.AllDone = true
	}
	
	return nil
}

func (m *Master) ReportOnMapJob(args *ReportOnMapJobArgs, reply *ReportOnMapJobReply) error {
	// Make sure only one RPC call is mutating the Master struct
	m.mu.Lock()
	defer m.mu.Unlock()

	if args.Status == 0 {
		m.mapTasks[args.TaskId].stage = 2
	} else {
		m.mapTasks[args.TaskId].stage = 0
	}
	return nil
}

func (m *Master) ReportOnReduceJob(args *ReportOnReduceJobArgs, reply *ReportOnReduceJobReply) error {
	// Make sure only one RPC call is mutating the Master struct
	m.mu.Lock()
	defer m.mu.Unlock()

	if args.Status == 0 {
		m.reduceTasks[args.TaskId].stage = 2
	} else {
		m.reduceTasks[args.TaskId].stage = 0
	}
	return nil
}


// Not an RPC handler just a helper function
func (m *Master) mapPhaseDone() bool {

	for _, v := range(m.mapTasks) {
		if v.stage != 2 {
			return false
		}
	}
	
	return true;
}

func (m *Master) reducePhaseDone() bool {

	for _, v := range(m.reduceTasks) {
		if v.stage != 2 {
			return false
		}
	}
	
	return true;
}

//
// an example RPC handler.
//
// the RPC argument and reply types are defined in rpc.go.
//
func (m *Master) Example(args *ExampleArgs, reply *ExampleReply) error {
	reply.Y = args.X + 1
	return nil
}


//
// start a thread that listens for RPCs from worker.go
//
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
	go http.Serve(l, nil)
}

//
// main/mrmaster.go calls Done() periodically to find out
// if the entire job has finished.
//
func (m *Master) Done() bool {
	ret := false

	// Your code here.
	if m.mapPhaseDone() && m.reducePhaseDone() {
		ret = true
	}

	return ret
}

//
// create a Master.
// main/mrmaster.go calls this function.
// nReduce is the number of reduce tasks to use.
//
func MakeMaster(files []string, nReduce int) *Master {
	m := Master{}

	// Your code here.
	m.mapTasks = make([]MapTask, len(files))
	for i, v := range(files) {
		m.mapTasks[i] = MapTask{
			taskId: i,
			filename: v,
			stage: 0,
		}
	}
	m.reduceTasks = make([]ReduceTask, nReduce)
	for i := 0; i < nReduce; i++ {
		m.reduceTasks[i] = ReduceTask{
			taskId: i,
			taskBucket: i,
			stage: 0,
		}
	}
	m.reducers = nReduce
	m.nFiles = len(files)

	PrintDebugf("MAP TASKS: %+v", m.mapTasks)
	PrintDebugf("REDUCE TASKS: %+v", m.reduceTasks)

	m.server()
	return &m
}