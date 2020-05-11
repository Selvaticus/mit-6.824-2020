package raft

//
// this is an outline of the API that raft must expose to
// the service (or tester). see comments below for
// each of these functions for more details.
//
// rf = Make(...)
//   create a new Raft server.
// rf.Start(command interface{}) (index, term, isleader)
//   start agreement on a new log entry
// rf.GetState() (term, isLeader)
//   ask a Raft for its current term, and whether it thinks it is leader
// ApplyMsg
//   each time a new entry is committed to the log, each Raft peer
//   should send an ApplyMsg to the service (or tester)
//   in the same server.
//

import "sync"
import "sync/atomic"
import "../labrpc"

// import "bytes"
// import "../labgob"

import "math/rand"
import "time"


// Raft server states
const (
	LEADER = 0
	FOLLOWER = 1
	CANDIDATE = 2
)

const HEARTBEAT = 10 * time.Millisecond

type ServerState int

type LogEntry struct {
	term int
	command interface{}
}


//
// as each Raft peer becomes aware that successive log entries are
// committed, the peer should send an ApplyMsg to the service (or
// tester) on the same server, via the applyCh passed to Make(). set
// CommandValid to true to indicate that the ApplyMsg contains a newly
// committed log entry.
//
// in Lab 3 you'll want to send other kinds of messages (e.g.,
// snapshots) on the applyCh; at that point you can add fields to
// ApplyMsg, but set CommandValid to false for these other uses.
//
type ApplyMsg struct {
	CommandValid bool
	Command      interface{}
	CommandIndex int
}

//
// A Go object implementing a single Raft peer.
//
type Raft struct {
	mu        sync.Mutex          // Lock to protect shared access to this peer's state
	peers     []*labrpc.ClientEnd // RPC end points of all peers
	persister *Persister          // Object to hold this peer's persisted state
	me        int                 // this peer's index into peers[]
	dead      int32               // set by Kill()

	// Your data here (2A, 2B, 2C).
	// Look at the paper's Figure 2 for a description of what
	// state a Raft server must maintain.

	// State for leader election (2A)
	currentTerm int
	votedFor int
	state ServerState

	log []LogEntry

	// Last comms from leader
	lastComm time.Time

}

// return currentTerm and whether this server
// believes it is the leader.
func (rf *Raft) GetState() (int, bool) {

	var term int
	var isleader bool
	// Your code here (2A).
	term = rf.currentTerm
	isleader = rf.state == LEADER

	return term, isleader
}

//
// save Raft's persistent state to stable storage,
// where it can later be retrieved after a crash and restart.
// see paper's Figure 2 for a description of what should be persistent.
//
func (rf *Raft) persist() {
	// Your code here (2C).
	// Example:
	// w := new(bytes.Buffer)
	// e := labgob.NewEncoder(w)
	// e.Encode(rf.xxx)
	// e.Encode(rf.yyy)
	// data := w.Bytes()
	// rf.persister.SaveRaftState(data)
}


//
// restore previously persisted state.
//
func (rf *Raft) readPersist(data []byte) {
	if data == nil || len(data) < 1 { // bootstrap without any state?
		return
	}
	// Your code here (2C).
	// Example:
	// r := bytes.NewBuffer(data)
	// d := labgob.NewDecoder(r)
	// var xxx
	// var yyy
	// if d.Decode(&xxx) != nil ||
	//    d.Decode(&yyy) != nil {
	//   error...
	// } else {
	//   rf.xxx = xxx
	//   rf.yyy = yyy
	// }
}




//
// example RequestVote RPC arguments structure.
// field names must start with capital letters!
//
type RequestVoteArgs struct {
	// Your data here (2A, 2B).
	Term int
	CandidateId int
}

//
// example RequestVote RPC reply structure.
// field names must start with capital letters!
//
type RequestVoteReply struct {
	// Your data here (2A).
	Term int
	VoteGranted bool
}

//
// example RequestVote RPC handler.
//
func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	// Your code here (2A, 2B).
	if args.Term < rf.currentTerm {
		// Candidate term is out of date. Don't vote and send current term so it can update itself
		reply.Term = rf.currentTerm
		reply.VoteGranted = false
	} else {
		rf.mu.Lock()
		defer rf.mu.Unlock()

		// Update current term as it will be at least the same as the candidate
		rf.currentTerm = args.Term
		
		reply.Term = rf.currentTerm

		if rf.votedFor == -1 || rf.votedFor == args.CandidateId {
			// If this server hasnt voted yet or already voted for that candidate, give vote to the candidate
			reply.VoteGranted = true

			// Update which candidate we voted for
			rf.votedFor = args.CandidateId
		} else {
			reply.VoteGranted = false
		}
	}
}

type AppendEntriesArgs struct {
	Term int
	LeaderId int
}

type AppendEntriesReply struct {
	Term int
	Success bool
}

func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	if args.Term < rf.currentTerm {
		// Term is out of date. Respond with current term
		reply.Success = false
	} else {
		// Make sure the instance has the latest term and reset the FOLLOWER status
		rf.currentTerm = args.Term
		rf.state = FOLLOWER
		rf.votedFor = -1

		// Record the time of this call
		rf.lastComm = time.Now()
		
		reply.Success = true
	}

	reply.Term = rf.currentTerm
}

//
// example code to send a RequestVote RPC to a server.
// server is the index of the target server in rf.peers[].
// expects RPC arguments in args.
// fills in *reply with RPC reply, so caller should
// pass &reply.
// the types of the args and reply passed to Call() must be
// the same as the types of the arguments declared in the
// handler function (including whether they are pointers).
//
// The labrpc package simulates a lossy network, in which servers
// may be unreachable, and in which requests and replies may be lost.
// Call() sends a request and waits for a reply. If a reply arrives
// within a timeout interval, Call() returns true; otherwise
// Call() returns false. Thus Call() may not return for a while.
// A false return can be caused by a dead server, a live server that
// can't be reached, a lost request, or a lost reply.
//
// Call() is guaranteed to return (perhaps after a delay) *except* if the
// handler function on the server side does not return.  Thus there
// is no need to implement your own timeouts around Call().
//
// look at the comments in ../labrpc/labrpc.go for more details.
//
// if you're having trouble getting RPC to work, check that you've
// capitalized all field names in structs passed over RPC, and
// that the caller passes the address of the reply struct with &, not
// the struct itself.
//
func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, reply *RequestVoteReply) bool {
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
	return ok
}


func (rf *Raft) sendAppendEntries(server int, args *AppendEntriesArgs, reply *AppendEntriesReply) bool {
	ok := rf.peers[server].Call("Raft.AppendEntries", args, reply)
	return ok
}


//
// the service using Raft (e.g. a k/v server) wants to start
// agreement on the next command to be appended to Raft's log. if this
// server isn't the leader, returns false. otherwise start the
// agreement and return immediately. there is no guarantee that this
// command will ever be committed to the Raft log, since the leader
// may fail or lose an election. even if the Raft instance has been killed,
// this function should return gracefully.
//
// the first return value is the index that the command will appear at
// if it's ever committed. the second return value is the current
// term. the third return value is true if this server believes it is
// the leader.
//
func (rf *Raft) Start(command interface{}) (int, int, bool) {
	index := -1
	term := -1
	isLeader := true

	// Your code here (2B).


	return index, term, isLeader
}

//
// the tester doesn't halt goroutines created by Raft after each test,
// but it does call the Kill() method. your code can use killed() to
// check whether Kill() has been called. the use of atomic avoids the
// need for a lock.
//
// the issue is that long-running goroutines use memory and may chew
// up CPU time, perhaps causing later tests to fail and generating
// confusing debug output. any goroutine with a long-running loop
// should call killed() to check whether it should stop.
//
func (rf *Raft) Kill() {
	atomic.StoreInt32(&rf.dead, 1)
	// Your code here, if desired.
}

func (rf *Raft) killed() bool {
	z := atomic.LoadInt32(&rf.dead)
	return z == 1
}

//
// the service or tester wants to create a Raft server. the ports
// of all the Raft servers (including this one) are in peers[]. this
// server's port is peers[me]. all the servers' peers[] arrays
// have the same order. persister is a place for this server to
// save its persistent state, and also initially holds the most
// recent saved state, if any. applyCh is a channel on which the
// tester or service expects Raft to send ApplyMsg messages.
// Make() must return quickly, so it should start goroutines
// for any long-running work.
//
func Make(peers []*labrpc.ClientEnd, me int,
	persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{}
	rf.peers = peers
	rf.persister = persister
	rf.me = me

	// Your initialization code here (2A, 2B, 2C).
	rf.currentTerm = 0
	rf.votedFor = -1 // Using -1 as no vote casted.
	rf.state = FOLLOWER

	rf.log = make([]LogEntry, 1)

	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())

	// Start long-runnong goroutines
	go rf.heartbeat()
	go rf.election()

	return rf
}


// Time driven activities

// This routine sends heartbeats to followers if this server instance is the leader
func (rf *Raft) heartbeat() {
	// Keep working whilst the raft server is not dead
	// Our test bed doesn't kill processes or goroutines, just marks them as killed
	// as such if we detect the raft server as been killed we should stop all work
	for !rf.killed() {
		
		time.Sleep(HEARTBEAT)

		if _, isleader := rf.GetState(); isleader {
			// This instance is the leader and should send heartbeats
			var args = &AppendEntriesArgs{}
			args.Term = rf.currentTerm
			args.LeaderId = rf.me

			for i := 0; i < len(rf.peers); i++ {
				if i != rf.me {
					// If not the current server instance
					reply := &AppendEntriesReply{}
					rf.sendAppendEntries(i, args, reply)
					// go func() {

					// }
				}
			}
		}
	}
}

// This routine checks if it should start a new election process
func (rf *Raft) election() {
	// Keep working whilst the raft server is not dead
	// Our test bed doesn't kill processes or goroutines, just marks them as killed
	// as such if we detect the raft server as been killed we should stop all work

	for !rf.killed() {
		// Get a randon election timout between 250 - 500 ms
		var electionTimeout = (rand.Float32() * 250) + 250

		time.Sleep(time.Duration(electionTimeout) * time.Millisecond)

		rf.mu.Lock()
		if rf.state == FOLLOWER {
			// We are a follower, check if we received comms from the leader

			if time.Since(rf.lastComm) > 2 * HEARTBEAT {
				// If we didnt received any comms or the last comm was longer than 2 heartbeats ago we need to start an election

				rf.currentTerm += 1
				rf.state = CANDIDATE

				// Vote for ourselves
				rf.votedFor = rf.me

				// SEND VOTE REQUESTS
				var args = &RequestVoteArgs{rf.currentTerm, rf.me}

				var votes = 0

				for i := 0; i < len(rf.peers); i++ {
					if i != rf.me {
						// If not the current server instance
						reply := &RequestVoteReply{}
						if rf.sendRequestVote(i, args, reply) && reply.VoteGranted{
							votes++
						}

						if votes > len(rf.peers)/2 {
							// this server instance won the majority

							rf.state = LEADER
							break;
						}
					}
				}
			}
		}
		rf.mu.Unlock()
	}
}