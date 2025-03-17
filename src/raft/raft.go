package raft

//
// this is an outline of the API that raft must expose to
// the service (or tester). see comments below for
// each of these functions for more details.
//
// rf = Make(...)
//   create a new Raft server.
//   创建一个新的Raft服务器
// rf.Start(command interface{}) (index, term, isleader)
//   start agreement on a new log entry
//   开始一个新的日志条目的共识
// rf.GetState() (term, isLeader)
//   ask a Raft for its current term, and whether it thinks it is leader
//   询问Raft当前的term，以及它是否认为自己是leader
// ApplyMsg
//   each time a new entry is committed to the log, each Raft peer
//   should send an ApplyMsg to the service (or tester)
//   in the same server.
//   每次新的条目被提交到日志时，每个Raft对等方
//   应该向服务（或测试器）发送一个ApplyMsg到同一个服务器。

import "sync"
import "sync/atomic"
import "../labrpc"

// import "bytes"
// import "../labgob"



//
// as each Raft peer becomes aware that successive log entries are
// committed, the peer should send an ApplyMsg to the service (or
// tester) on the same server, via the applyCh passed to Make(). set
// CommandValid to true to indicate that the ApplyMsg contains a newly
// committed log entry.
//
// 随着每个Raft对等方意识到后续的日志条目被提交，对等方应该向服务（或测试器）
// 通过传递给Make()的applyCh发送一个ApplyMsg到同一个服务器。
// 设置CommandValid为true以指示ApplyMsg包含一个新提交的日志条目。
//
// in Lab 3 you'll want to send other kinds of messages (e.g.,
// snapshots) on the applyCh; at that point you can add fields to
// ApplyMsg, but set CommandValid to false for these other uses.
//
// 在Lab 3中，您可能希望在applyCh上发送其他类型的消息（例如快照）；
// 此时您可以向ApplyMsg添加字段，但对于其他用途，请将CommandValid设置为false。

type ApplyMsg struct {
	CommandValid bool
	Command      interface{}
	CommandIndex int
}

//
// A Go object implementing a single Raft peer.
// 一个Go对象实现了一个Raft对等方。
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

	// 您的数据在这里（2A，2B，2C）。
	// 查看论文的Figure 2以了解Raft服务器必须维护的状态。

}

// return currentTerm and whether this server
// believes it is the leader.
//
// 返回当前term和该服务器是否认为自己是leader。
func (rf *Raft) GetState() (int, bool) {

	var term int
	var isleader bool
	// Your code here (2A).
	return term, isleader
}

//
// save Raft's persistent state to stable storage,
// where it can later be retrieved after a crash and restart.
// see paper's Figure 2 for a description of what should be persistent.
//
// 将Raft的持久化状态保存到稳定存储中，这个存储使得其他可以在崩溃和重新启动之后检索。
// 查看论文的Figure 2以了解应该持久化的内容。
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
// 恢复先前持久化的状态。
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
// 示例RequestVote RPC参数结构。
// 字段名称必须以大写字母开头！
type RequestVoteArgs struct {
	// Your data here (2A, 2B).
}

//
// example RequestVote RPC reply structure.
// field names must start with capital letters!
//
type RequestVoteReply struct {
	// Your data here (2A).
}

//
// example RequestVote RPC handler.
//
// RPC 处理程序。
func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	// Your code here (2A, 2B).
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
// RPC调用。
func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, reply *RequestVoteReply) bool {
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
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
// 这个使用Raft的服务（例如，一个k/v服务器）希望在Raft的日志中开始
// 下一个命令的租约。如果这个服务器不是领导者，返回false。否则，启动
// 租约并立即返回。并不保证此命令最终将被提交到Raft日志中，因为领导者可能会失败或失去选举。
// 即使Raft实例已被杀死，此函数也应优雅地返回。
//
// the first return value is the index that the command will appear at
// if it's ever committed. the second return value is the current
// term. the third return value is true if this server believes it is
// the leader.
//
// 如果命令最终被提交，则第一个返回值是命令将出现的索引。第二个返回值是当前任期。
// 第三个返回值是true，如果这个服务器相信它是领导者。
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
// 测试器不会在每个测试之后停止由Raft创建的goroutine，但它会调用Kill()方法。
// 您的代码可以使用killed()来检查是否已调用Kill()。原子避免了需要锁的需要。
//
// the issue is that long-running goroutines use memory and may chew
// up CPU time, perhaps causing later tests to fail and generating
// confusing debug output. any goroutine with a long-running loop
// should call killed() to check whether it should stop.
//
// 长期运行的goroutine使用内存并可能消耗CPU时间，可能导致后续测试失败并生成令人困惑的调试输出。
// 任何具有长时间运行循环的goroutine都应该调用killed()来检查它是否应该停止。
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
// 这个服务或测试程序想要创建一个Raft服务器。所有Raft服务器的端口（包括这个）都在peers[]中。
// 这个服务器的端口是peers[me]。所有服务器的peers[]数组都有相同的顺序。
// persister是这个服务器用来保存其持久化状态的地方，并且最初保存了最近的状态，如果有。
// applyCh是一个通道，在这个通道上，server或tester期望Raft发送ApplyMsg消息。
// Make()必须快速返回，因此它应该为任何长时间运行的工作启动goroutine。
//
func Make(peers []*labrpc.ClientEnd, me int,
	persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{}
	rf.peers = peers
	rf.persister = persister
	rf.me = me

	// Your initialization code here (2A, 2B, 2C).

	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())


	return rf
}
