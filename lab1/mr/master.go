package mr

import "log"
import "net"
import "os"
import "net/rpc"
import "net/http"
import "sync"
import "time"


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
	started time.Time
}

type ReduceTask struct {
	taskId int
	taskBucket int
	// 0 will be not started
	// 1 running
	// 2 done
	stage int
	started time.Time
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
		for i, v := range(m.mapTasks) {
			if v.stage == 0 {
				reply.Filename = v.filename
				reply.TaskId = v.taskId
				reply.AllDone = false
				reply.TaskType = Map_Task
				reply.Reducers = m.reducers		

				// Marked as running, aka assigned to worker
				// and record when it started
				v.stage = 1
				v.started = time.Now()
				m.mapTasks[i] = v

				PrintDebugf("Sending job: %+v", *reply)
				PrintDebugf("Updated task: %+v", v)
				return nil
			}
		}
	} else if !m.reducePhaseDone() {
		PrintDebug("Reduce phase not done yet...")
		// Send reduce task
		for i, v := range(m.reduceTasks) {
			if v.stage == 0 {
				reply.BucketId = v.taskBucket
				reply.TaskId = v.taskId
				reply.AllDone = false
				reply.TaskType = Reduce_Task

				// Marked as running, aka assigned to worker
				// and record when it started
				v.stage = 1
				v.started = time.Now()
				m.reduceTasks[i] = v

				PrintDebugf("Sending job: %+v", reply)
				PrintDebugf("Updated task: %+v", v)
				return nil
			}
		}
	} else {
		PrintDebug("All done.")
		reply.AllDone = true
		return nil
	}

	// Hacky way to let the worker no there is no task at the moment
	// but the job is not done yet either
	reply.TaskType = No_Task
	
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

// Scheduler function to reassign long running tasks to other workers
func (m *Master) checkForLongRunningTasks() {
	// Make sure no one with mutate our data while we figure out if tasks need reassigment
	PrintDebug("Checking for long running tasks...")
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.mapPhaseDone() {
		for i, v := range(m.mapTasks) {
			PrintDebugf("Checking task: %+v", v)
			if v.stage == 1 {
				elapsed := time.Since(v.started)
				if elapsed.Seconds() > 10 {
					PrintDebug("It is long running, making it available to assigment")
					// Task as been running for too long, needs to be made available again
					v.stage = 0
					m.mapTasks[i] = v
					
				}
			}
		}
	} else if !m.reducePhaseDone() {
		for i, v := range(m.reduceTasks) {
			PrintDebugf("Checking task: %+v", v)
			if v.stage == 1 {
				elapsed := time.Since(v.started)
				if elapsed.Seconds() > 10 {
					PrintDebug("It is long running, making it available to assigment")
					// Task as been running for too long, needs to be made available again
					v.stage = 0
					m.reduceTasks[i] = v
				}
			}
		}
	}
	PrintDebug("Checking for long running tasks done.")
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
	m.mu.Lock()
	defer m.mu.Unlock()
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

	// initialise a map task per file
	m.mapTasks = make([]MapTask, len(files))
	for i, v := range(files) {
		m.mapTasks[i] = MapTask{
			taskId: i,
			filename: v,
			stage: 0,
		}
	}

	// initialise a reduce task per reducer
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

	// Schedule a function to reasign task if they taking too long
	tenSecondTimer := time.NewTicker(10 * time.Second)
	go func() {
		// wait for tick
		for {
			PrintDebug("Waiting for timer")
			<-tenSecondTimer.C
			m.checkForLongRunningTasks()
		}
	}()

	PrintDebugf("MAP TASKS: %+v", m.mapTasks)
	PrintDebugf("REDUCE TASKS: %+v", m.reduceTasks)

	m.server()
	return &m
}