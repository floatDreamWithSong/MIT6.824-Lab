package raft

// rf = Make(...)  										  创建一个新的Raft服务器
// rf.Start(command interface{}) (index, term, isleader)  开始一个新的日志条目的共识
// rf.GetState() (term, isLeader)					      询问Raft当前的term，以及它是否认为自己是leader
// ApplyMsg   						每次新的条目被提交到日志时，每个Raft Peer应该发送一个ApplyMsg到同一个服务器。

import (
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"../labrpc"
)

// import "bytes"
// import "../labgob"

// 随着每个Raft Peer意识到后续的日志条目被提交，Peer应该通过传递给Make()的applyCh发送一个ApplyMsg到同一个服务器。
// 设置CommandValid为true以指示ApplyMsg包含一个新提交的日志条目。
//
// 在Lab 3中，您可能希望在applyCh上发送其他类型的消息（例如快照）；
// 此时您可以向ApplyMsg添加字段，但对于其他用途，请将CommandValid设置为false。

const (
	// 状态
	FOLLOWER = iota
	CANDIDATE
	LEADER
)

const (
	// 常量
	HEARTBEAT_TIMEOUT = 50
	APPLY_TIMEOUT = 30
)

type ApplyMsg struct {
	CommandValid bool
	Command      interface{}
	CommandIndex int
}

type LogEntry struct {
	Term int
	Command interface{}
}

type Raft struct {
	mu        sync.Mutex          // Peer状态锁
	peers     []*labrpc.ClientEnd // Peer 的RPC终端
	persister *Persister          // Peer的状态持久化
	me        int                 // 当前的Peer的index
	dead      int32               // set by Kill()

	// 您的数据在这里（2A，2B，2C）。
	// 查看论文的Figure 2以了解Raft服务器必须维护的状态。

	// 非易失性状态（在响应RPC前就已经被持久化）
	currentTerm int // 当前term
	votedFor int 	// 投票给谁
	log []LogEntry  // 日志条目
	// 易失性状态
	commitIndex int // 已经commit的最后日志index
	lastApplied int // 已经应用到状态机的最后日志index
	// leader 易失性状态, 每次选举后重新初始化
	nextIndex []int // 对所有的服务器，下一个要发送的日志index, 初始化为leader最后一个log的index + 1
	matchIndex []int// 对所有的服务器，已知的最新commit的日志index
	// candidate 状态
	// 选举定时器
	electionTimeOut time.Time
	// 心跳定时器
	heartbeatTimer int

	// 其他
	applyCh chan ApplyMsg
	stage int // 0:follower 1:candidate 2:leader
}

func (rf* Raft) setElectionTimeOut(server int64) {
	// 设置基础时间
    t:= time.Now().Add(time.Millisecond*800)
	// 设置随机种子
    r := rand.New(rand.NewSource(time.Now().Unix() + server))
	// 获取随机波动的时间范围
	ms:=r.Int63() %300
	// 设置超时时间
	t = t.Add(time.Duration(ms)*time.Millisecond)
	rf.electionTimeOut = t
	DPrintf("%d 的选举超时时间是 %v", rf.me, rf.electionTimeOut)
}

func (rf *Raft) changeStage(to int,reset bool) {
	if to == CANDIDATE {
		/*
			candidate的服务器规则：
			1.转变为选举人之后开始选举
			2.currentTerm自增
			3.给自己投票
			4.重置选举计时器
		*/
		rf.stage = CANDIDATE
		rf.currentTerm += 1
		rf.votedFor = rf.me
		rf.setElectionTimeOut(int64(rf.me))
		rf.joinElection()
	}
}

// 如果一段时间没有收到心跳，就会发起选举
func (rf *Raft) ticker() {
	for rf.killed() == false {
		time.Sleep(HEARTBEAT_TIMEOUT*time.Millisecond)
		rf.mu.Lock()
		// 如果是leader就发送心跳，否则超时就发起选举
		if rf.stage == LEADER {
			rf.leaderAppend()
		}else {
			if time.Now().After(rf.electionTimeOut){
				rf.changeStage(CANDIDATE, true)
			}
		}
		rf.mu.Unlock()
	}
}

func (rf* Raft) joinElection () {
	voteCount := 1 // 自己开始投给自己有一票
	for i:=0;i<len(rf.peers);i++ {
		if i == rf.me {
			continue
		}
		go func(server int) {
			rf.mu.Lock()
			args := & RequestVoteArgs{
				Term: rf.currentTerm,
				CandidateId: rf.me,
				LastLogIndex: rf.getLastLogIndex(),
				LastLogTerm: rf.log[rf.getLastLogIndex()].Term,
			}
			reply := & RequestVoteReply{}
			rf.mu.Unlock()
			ok := rf.sendRequestVote(server, args, reply)
			if ok {
				rf.mu.Lock()
				// 如果自己不是候选者，或者已经出现了新的term
				if rf.stage != CANDIDATE || reply.Term < rf.currentTerm {
					rf.mu.Unlock()
					return
				}
				if reply.VoteGranted && args.Term == rf.currentTerm {
					voteCount++
					// 如果超过半数的服务器投票给了自己
					if voteCount > len(rf.peers)/2 + 1 {
						rf.changeStage(LEADER, true)
					}
					rf.mu.Unlock()
					return
				}
				// 如果自己过期了
				if reply.Term > args.Term {
					if reply.Term > rf.currentTerm {
						rf.currentTerm = reply.Term
					}
					rf.changeStage(FOLLOWER, false)
				}
				rf.mu.Unlock()
				return
			}
		}(i)
	}
}

// 返回当前term和该服务器是否认为自己是leader。

func (rf *Raft) GetState() (int, bool) {

	var term int
	var isleader bool
	// Your code here (2A).
	return term, isleader
}

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
// 恢复先前持久化的状态。
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

/**
* <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<
**/

// 示例RequestVote RPC参数结构。字段名称必须以大写字母开头！
type RequestVoteArgs struct {
	// Your data here (2A, 2B).
	Term int // 候选人现在的term
	CandidateId int // 候选人的ID
	LastLogIndex int // 候选人最后的log index
	LastLogTerm int // 候选人最后log的term
}
type AppendEntriesArgs struct {
	Term int // leader的term
	LeaderId int // leader的ID
	PrevLogIndex int // leader的前一个log的index
	PrevLogTerm int // leader的前一个log的term
	Entries []LogEntry // 要发送的log entries
	LeaderCommit int // leader commit的index
}

// 示例RequestVote RPC回复结构。字段名称必须以大写字母开头！
type RequestVoteReply struct {
	// Your data here (2A).
	Term int // 投票者的term，用于更新候选人term
	VoteGranted bool // 候选人是否得到选票 
}
type AppendEntriesReply struct {
	Term int // follower的term，用于更新候选人term
	Success bool // 如果匹配了prevLogIndex和prevLogTerm，则为true
}

// RPC 处理程序。
func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	// Your code here (2A, 2B).
}
func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {

}

// RPC调用。
func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, reply *RequestVoteReply) bool {
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply) // 调用目的服务器的RequestVote方法
	return ok
}
func (rf *Raft) sendAppendEntries(server int, args *AppendEntriesArgs, reply *AppendEntriesReply) bool {
	ok := rf.peers[server].Call("Raft.AppendEntries", args, reply)
	return ok
}

/**
* <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<
**/

// 这个使用Raft的服务（例如，一个k/v服务器）希望start agreement on the next command to be appended to Raft's log。
// 如果这个服务器不是leader，返回false。否则， start the agreement
// 并立即返回。并不保证此命令最终将被提交到Raft日志中，因为领导者可能会失败或失去选举。
// 即使Raft实例已被杀死，此函数也应优雅地返回。

// 如果命令最终committed，则第一个返回值是命令将出现的index。第二个返回值是当前term。
// 第三个返回值是这个服务器是否相信它是leader。
//
func (rf *Raft) Start(command interface{}) (int, int, bool) {
	index := -1
	term := -1
	isLeader := true

	// Your code here (2B).


	return index, term, isLeader
}


// 测试器不会在每个测试之后停止由Raft创建的goroutine，但它会调用Kill()方法。
// 您的代码可以使用killed()来检查是否已调用Kill()。原子避免了需要锁的需要。
//
// 长期运行的goroutine使用内存并可能消耗CPU时间，可能导致后续测试失败并生成令人困惑的调试输出。
// 任何具有长时间运行循环的goroutine都应该调用killed()来检查它是否应该停止。

func (rf *Raft) Kill() {
	atomic.StoreInt32(&rf.dead, 1)
	// Your code here, if desired.
}

func (rf *Raft) killed() bool {
	z := atomic.LoadInt32(&rf.dead)
	return z == 1
}


// 这个服务或测试程序想要创建一个Raft服务器。所有Raft服务器的端口（包括这个）都在peers[]中。
// 这个服务器的端口是peers[me]。所有服务器的peers[]数组都有相同的顺序。
// persister是这个服务器用来保存其持久化状态的地方，并且最初保存了最近的状态，如果有。
// applyCh是一个通道，在这个通道上，server或tester期望Raft发送ApplyMsg消息。
// Make()必须快速返回，因此它应该为任何长时间运行的工作启动goroutine。

func Make(peers []*labrpc.ClientEnd, me int,
	persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{}
	rf.peers = peers
	rf.persister = persister
	rf.me = me

	// Your initialization code here (2A, 2B, 2C).
	rf.mu.Lock()
	rf.stage = FOLLOWER
	rf.currentTerm = 0
	rf.votedFor = -1
	rf.commitIndex = 0
	rf.lastApplied = 0

	rf.log = []LogEntry{}
	rf.log = append(rf.log, LogEntry{Term: 0, Command: nil})
	rf.applyCh = applyCh
	rf.setElectionTimeOut(int64(rf.me))
	rf.mu.Unlock()

	// 恢复已被持久化的状态
	rf.readPersist(persister.ReadRaftState())

	DPrintf("init server %d, term %d",rf.me,rf.currentTerm)

	go rf.ticker()
	go rf.apppliedTicker()

	return rf
}
