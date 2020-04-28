# Lab 1: MapReduce

## Introduction

We will build a MapReduce system. We will implement a worker process that calls application Map and Reduce functions and handles reading and writing files, and a master process that hands out tasks to workers and copes with failed workers.

We will be building something similar to the [MapReduce paper](http://research.google.com/archive/mapreduce-osdi04.pdf).

**Notes:**

* This document only has summary of directions given for this lab, for full information about this lab, by the teaching staff, go [here](https://pdos.csail.mit.edu/6.824/labs/lab-mr.html)

## Getting Started

A simple sequential mapreduce implementation is supplied in `main/mrsequential.go`. It runs the maps and reduces one at a time, in a single process. Also provided are a couple of MapReduce applications: word-count in `mrapps/wc.go`, and a text indexer in `mrapps/indexer.go`. You can run word count sequentially as follows:

```sh
$ cd lab1/main
$ go build -buildmode=plugin mrapps/wc.go
$ rm mr-out*
$ go run mrsequential.go wc.so pg*.txt
$ more mr-out-0
A 509
ABOUT 2
ACT 8
...
```

`mrsequential.go` leaves its output in the file `mr-out-0`. The input is from the text files named `pg-xxx.txt`.

From the staff:
> Feel free to borrow code from `mrsequential.go`. You should also have a look at `mrapps/wc.go` to see what MapReduce application code looks like

## The Job

The job is to implement a distributed MapReduce, consisting of two programs, the master and the worker. There will be just one master process, and one or more worker processes executing in parallel. 

In a real system the workers would run on a bunch of different machines, but for this lab we will run them all on a single machine. The workers will talk to the master via RPC. Each worker process will ask the master for a task, read the task's input from one or more files, execute the task, and write the task's output to one or more files.

The master should notice if a worker hasn't completed its task in a reasonable amount of time (for this lab, use ten seconds), and give the same task to a different worker.

The staff has given a little code to start off. The "main" routines for the master and worker are in `main/mrmaster.go` and `main/mrworker.go`; don't change these files. We should put our implementation in `mr/master.go`, `mr/worker.go`, and `mr/rpc.go`.

*Note: `master.go`, `worker.go`, and `rpc.go` also have barebone implementations, so just a fraction of the code will be implemented by us. (should be highlighted in the code comments)*

### How to run

First we need to make a plugin of whatever mrapp we are going to run, below there's an example of the word counter app.

```sh
go build -buildmode=plugin ../mrapps/wc.go
```

Second we need to start the master with the inputs to process

```sh
In the main directory, run the master.
rm mr-out*
go run mrmaster.go pg-*.txt
```

The `pg-*.txt` arguments to `mrmaster.go` are the input files; each file corresponds to one "split", and is the input to one Map task.

Third we need to start some workers with the app that is to run

```sh
go run mrworker.go wc.so
```

When the workers and master have finished, look at the output in `mr-out-*`. When you've completed the lab, the sorted union of the output files should match the sequential output, like this:

```sh
$ cat mr-out-* | sort | more
A 509
ABOUT 2
ACT 8
...
```

### How to test

A test script has been supplied, by the staff, in `main/test-mr.sh`. The tests check that the wc and indexer MapReduce applications produce the correct output when given the `pg-xxx.txt` files as input. The tests also check that your implementation runs the Map and Reduce tasks in parallel, and that your implementation recovers from workers that crash while running tasks.

The output needs to be in files named `mr-out-X`, one for each reduce task. The empty implementations of `mr/master.go` and `mr/worker.go` don't produce those files (or do much of anything else), so the test fails.

When you've finished, the test script output should look like this:

```sh
$ sh ./test-mr.sh
*** Starting wc test.
--- wc test: PASS
*** Starting indexer test.
--- indexer test: PASS
*** Starting map parallelism test.
--- map parallelism test: PASS
*** Starting reduce parallelism test.
--- reduce parallelism test: PASS
*** Starting crash test.
--- crash test: PASS
*** PASSED ALL TESTS
```

## Rules

* The map phase should divide the intermediate keys into buckets for **nReduce** reduce tasks, where **nReduce** is the argument that `main/mrmaster.go` passes to `MakeMaster()`.

* The worker implementation should put the output of the X'th reduce task in the file `mr-out-X`.

* A `mr-out-X` file should contain one line per Reduce function output. The line should be generated with the Go **"%v %v"** format, called with the key and value. Have a look in `main/mrsequential.go` for the line commented "this is the correct format". The test script will fail if your implementation deviates too much from this format.

* The worker should put intermediate Map output in files in the current directory, where your worker can later read them as input to Reduce tasks.

* `main/mrmaster.go` expects `mr/master.go` to implement a `Done()` method that returns true when the MapReduce job is completely finished; at that point, `mrmaster.go` will exit.

* When the job is completely finished, the worker processes should exit. A simple way to implement this is to use the return value from `call()`: if the worker fails to contact the master, it can assume that the master has exited because the job is done, and so the worker can terminate too. Depending on your design, you might also find it helpful to have a "please exit" pseudo-task that the master can give to workers.

## Dev log

### 26/04/2020

Although the implementation solves the Map/Reduce problem, I belive this is not done on an efficient manner, also the code needs to be refactored.

This points me to the conclusion that a cleaner more efficient structure can be found on master to managed all the workloads.

As of today the code passes all tests except the crash test when ran per the test script.

### 28/04/2020

As of today the implementations passes all given tests, if they are done separatly, when done in sequence the *reduce parallelism* test is really slow, and the *crash* test fails. I believe this is due to the way I'm handling the intermediate files.

This might also be the reason for the performance issues perceived before.

So further work needs to be done.