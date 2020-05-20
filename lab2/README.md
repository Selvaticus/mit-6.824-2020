# Lab 2: Raft

## Introduction

This is the first in a series of labs in which we will build a fault-tolerant key/value storage system.

* In this lab we will implement Raft, a replicated state machine protocol.
* In the next lab (lab3) we will build a key/value service on top of Raft.
* In the last lab (lab4) we will “shard” the service over multiple replicated state machines for higher performance.

A replicated service achieves fault tolerance by storing complete copies of its state (i.e., data) on multiple replica servers. Replication allows the service to continue operating even if some of its servers experience failures (crashes or a broken or flaky network). The challenge is that failures may cause the replicas to hold differing copies of the data.

We will be following the design in the [extended Raft paper](https://pdos.csail.mit.edu/6.824/papers/raft-extended.pdf), with particular attention to Figure 2.

This lab is has three parts:

* 2A - Raft leader election and heartbeats.
* 2B - Append new log entries to leader and follower nodes.
* 2C - Persistent state to survive a reboot.

**Notes:**

* This document only has summary of directions given for this lab, for full information about this lab, by the teaching staff, go [here](https://pdos.csail.mit.edu/6.824/labs/lab-raft.html)

## Getting Started

Most of this lab is too build the Raft protocol, so there is little to run at this moment, more will happen when we start to build systems on top of this work in the coming labs.

For now all we need to worry is that the provided test bed passes for the different parts of this lab.

We can check this by running:

```sh
$ cd lab2/raft
$ go test
Test (2A): initial election ...
--- FAIL: TestInitialElection2A (5.04s)
        config.go:326: expected one leader, got none
Test (2A): election after network failure ...
--- FAIL: TestReElection2A (5.03s)
        config.go:326: expected one leader, got none
...
$
```

## The Job

Implement Raft by adding code to `raft/raft.go`. There is just skeleton code in that file.

The implementation must have the following interface, which the tests and (eventually) the other labs will use. More details in comments in `raft.go`.

```go
// create a new Raft server instance:
rf := Make(peers, me, persister, applyCh)

// start agreement on a new log entry:
rf.Start(command interface{}) (index, term, isleader)

// ask a Raft for its current term, and whether it thinks it is leader
rf.GetState() (term, isLeader)

// each time a new entry is committed to the log, each Raft peer
// should send an ApplyMsg to the service (or tester).
type ApplyMsg
```

**Important:** The Raft peers should exchange RPCs using the *labrpc* Go package (source in `src/labrpc`). The tester can tell labrpc to delay RPCs, re-order them, and discard them to simulate various network failures.


### 2A - Raft leader election and heartbeats

In the first part (2A) of this lab we must implement the leader election and heartbeats parts of the protocol.

The goal for Part 2A is for a single leader to be elected, for the leader to remain the leader if there are no failures, and for a new leader to take over if the old leader fails or if packets to/from the old leader are lost.

Run `go test -run 2A` to test this part code.

A good run would look something like this:

```sh
$ go test -run 2A
Test (2A): initial election ...
  ... Passed --   4.0  3   32    9170    0
Test (2A): election after network failure ...
  ... Passed --   6.1  3   70   13895    0
PASS
ok      raft    10.187s
$
```

Each "Passed" line contains five numbers; these are the time that the test took in seconds, the number of Raft peers (usually 3 or 5), the number of RPCs sent during the test, the total number of bytes in the RPC messages, and the number of log entries that Raft reports were committed.
We can use this numbers to help sanity-check the number of RPCs that our implementation sends.

For all of labs 2, 3, and 4, we should assume that if it takes more than 600 seconds for all of the tests (go test), or if any individual test takes more than 120 seconds, its a failure mark.

## Dev log

### 04/05/2020

Initial code provided by the staff was pulled. (Check this commit for the code that was provided versus the one I wrote)

### 11/05/2020

*Working on part 2A.*

First test for 2A pass, but not the second. So the first leader is elected, but no new leader gets elected after failure in the cluster.

Will need to debug the above, and the code will need some proper refactor its quite ugly as of now.

### 20/05/2020

*Still working on part 2A.*

Code is still ugly, but I'm focusing on getting the reelection working.

* Fixed the RequestVote RFC to handle out of date state properly. Now we should be resetting to follower whenever the current raft is out of date.

* Changed the way we check for reelection timeouts. Use the reelection timeout and lastComms time to trigger reelection. (insted of previous 2 * heartbeat)

Finally found a nasty deadlock bug with mutex locks and channels. I've fixed the bug and now all tests for part 2A pass. Code is still pretty ugly and needs some care

Things to improve:

* Handle the timers better

* Implement reset of election timer (This can possibly cause liveliness issues)

* Implement better print debug function (More information provided)

* Handle the "kill signal" better

* Clean up the fix for the deadlock bug between mutex locks and channels
