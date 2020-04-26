package mr

import "fmt"
import "log"
import "net/rpc"
import "hash/fnv"
import "io/ioutil"
import "os"

import "encoding/json"
import "strconv"
import "path/filepath"
import "sort"


//
// Map functions return a slice of KeyValue.
//
type KeyValue struct {
	Key   string
	Value string
}

// for sorting by key.
type ByKey []KeyValue

// for sorting by key.
func (a ByKey) Len() int           { return len(a) }
func (a ByKey) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByKey) Less(i, j int) bool { return a[i].Key < a[j].Key }

//
// use ihash(key) % NReduce to choose the reduce
// task number for each KeyValue emitted by Map.
//
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}


//
// main/mrworker.go calls this function.
//
func Worker(mapf func(string, string) []KeyValue,
	reducef func(string, []string) string) {

	// Your worker implementation here.

	taskArgs := GetTaskArgs{}
	taskReply := GetTaskReply{}

	// Ask for the first task
	runTask := call("Master.GetTask", &taskArgs, &taskReply)

	for runTask == true && !taskReply.AllDone {
		if taskReply.IsMap {
			PrintDebug("Gonna run the map function")
			PrintDebugf("Task: %+v", taskReply)
			_, err := runMapPhase(mapf, &taskReply)

			// Report on the map job status
			reportArgs := ReportOnMapJobArgs{}
			reportReply := ReportOnMapJobReply{}
			reportArgs.TaskId = taskReply.TaskId
			if err == nil {
				// If the map job ran into errors report a failed job to master
				reportArgs.Status = 0
			} else {
				reportArgs.Status = 1
			}
			// If the map job ran fine, report a good run back to master
			call("Master.ReportOnMapJob", &reportArgs, &reportReply)
		} else {
			// If it is not a map job, then must be a reduce job

			PrintDebug("Gonna run the reduce function")
			PrintDebugf("Task: %+v", taskReply)
			_, err := runReducePhase(reducef, &taskReply)


			// Report on the reduce job status
			reportArgs := ReportOnReduceJobArgs{}
			reportReply := ReportOnReduceJobReply{}
			reportArgs.TaskId = taskReply.TaskId
			if err == nil {
				// If the map job ran into errors report a failed job to master
				reportArgs.Status = 0
			} else {
				reportArgs.Status = 1
			}

			call("Master.ReportOnReduceJob", &reportArgs, &reportReply)
		}

		taskArgs = GetTaskArgs{}
		taskReply = GetTaskReply{}

		// Ask for the next task
		runTask = call("Master.GetTask", &taskArgs, &taskReply)
	}


	// uncomment to send the Example RPC to the master.
	// CallExample()

}

func runMapPhase(mapf func(string, string) []KeyValue, task *GetTaskReply)  ([]string, error) {
	// Read input
	byteContents, err := ioutil.ReadFile(task.Filename)
	if err != nil {
		return nil, err
	}

	contents := string(byteContents[:])

	// Call the app map function
	mapResult := mapf(task.Filename, contents)

	intermediateFiles := make(map[int]*os.File)

	var intermediateFilesNames []string

	for _, kv := range mapResult {
		// Go through the results from the Map function
		bucket := ihash(kv.Key) % task.Reducers
		file, ok := intermediateFiles[bucket]
		if !ok {
			// If no file descriptor exists for the bucket, create a new one
			// Create a temp file to be renamed after
			file, err = ioutil.TempFile("", "mr-")
			// file, err = os.Create(filename)
			if err != nil {
				// If we fail to create a file, return err
				return nil, err
			}
			// Update map with file descriptiors and list of files written
			intermediateFiles[bucket] = file
			// intermediateFilesNames = append(intermediateFilesNames, filename)
		}
		// write the result Key/Value to the file
		enc := json.NewEncoder(file)
		enc.Encode(&kv)

	}

	// Renaming all temp files and close file handlers
	for key, value := range intermediateFiles {
		old_filename := value.Name()
		// Close file handler
		value.Close()
		new_filename := fmt.Sprintf("mr-%v-%v", strconv.Itoa(task.TaskId), strconv.Itoa(key))
		err := os.Rename(old_filename, new_filename)
		if err != nil {
			log.Fatalf("Failed to rename file %v to %v", old_filename, new_filename)
		}
		intermediateFilesNames = append(intermediateFilesNames, new_filename)
	}


	return intermediateFilesNames, nil
}

func runReducePhase(reducef func(string, []string) string, task *GetTaskReply)  (string, error) {

	// Assumption that the format will always going to be mr-MAP_TASK_ID-REDUCE_TASK_ID
	matches, err := filepath.Glob("mr-*-"+strconv.Itoa(task.BucketId))

	if err != nil {
		return "", err
	}

	// read all files for this reducer from all the map intermediate results
	intermediate := []KeyValue{}
	for _, filename := range(matches)  {
		file, err := os.Open(filename)
		if err != nil {
			log.Fatalf("cannot open %v", filename)
		}
		// content, err := ioutil.ReadAll(file)
		dec := json.NewDecoder(file)
		for {
			var kv KeyValue
			if err := dec.Decode(&kv); err != nil {
				break
			}
			intermediate = append(intermediate, kv)
		}
		file.Close()
	}

	sort.Sort(ByKey(intermediate))

	// oname := "mr-out-"+strconv.Itoa(task.BucketId)
	// ofile, _ := os.Create(oname)

	ofile, err := ioutil.TempFile("", "mr-out")

	// call Reduce on each distinct key in intermediate[], and print the result
	i := 0
	for i < len(intermediate) {
		j := i + 1
		for j < len(intermediate) && intermediate[j].Key == intermediate[i].Key {
			j++
		}
		values := []string{}
		for k := i; k < j; k++ {
			values = append(values, intermediate[k].Value)
		}
		output := reducef(intermediate[i].Key, values)

		// this is the correct format for each line of Reduce output.
		fmt.Fprintf(ofile, "%v %v\n", intermediate[i].Key, output)

		i = j
	}

	oname := ofile.Name()

	ofile.Close()

	new_name := "mr-out-"+strconv.Itoa(task.BucketId)

	err = os.Rename(oname, new_name)
	if err != nil {
		log.Fatalf("Failed to rename file %v to %v", oname, new_name)
	}

	return new_name, nil
}

//
// example function to show how to make an RPC call to the master.
//
// the RPC argument and reply types are defined in rpc.go.
//
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

//
// send an RPC request to the master, wait for the response.
// usually returns true.
// returns false if something goes wrong.
//
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