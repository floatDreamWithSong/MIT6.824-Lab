# MIT6.824实验
## Lab1: MapReduce

[lab1-link](http://nil.csail.mit.edu/6.824/2020/labs/lab-mr.html)

### Introduction
In this lab you'll build a MapReduce system.
You'll implement **a worker process** that **calls** application Map and Reduce functions
and **handles reading and writing** files,
and **a coordinator process** that **hands out tasks** to workers 
and **copes with failed workers**.
You'll be building something similar to the MapReduce paper.
(Note: this lab uses "coordinator" instead of the paper's "master".)

实现一个和MapReduce论文类似的机制，也就是数单词个数Word Count。
用于测试的文件在src/main目录下，以pg-*.txt形式命名。每个pg-*.txt文件都是一本电子书，非常长。
我们的任务是统计出所有电子书中出现过的单词，以及它们的出现次数。

> 注意，由于课程作业的GO版本较老，现在的Go使用模块机制，导致不能使用相对路径
> 执行`go env -w GO111MODULE=off`来关闭模块机制
> 如果提示无法修改，因为和环境变量冲突，请先`unset GO111MOUDLE`，再设置

```shell
# cd src/main 
go build -buildmode=plugin ../mrapps/wc.go
go run mrsequential.go wc.so pg*.txt
```

mrsequential.go leaves its output in the file mr-out-0.
The input is from the text files named pg-xxx.txt.

Feel free to borrow code from mrsequential.go. You should also have a look at mrapps/wc.go to see what MapReduce application code looks like.

任务难度：中/难
你的任务是实现一个分布式的 MapReduce，包含两种程序：coordinator 和 worker。
一个 coordinator 进程和多个并行运行的 worker 进程。
在实际系统中，workers 会运行在不同的机器上，但在本实验中，你将在同一台机器上运行它们。
workers 通过 RPC 与 coordinator 通信。
每个 worker 进程将循环执行以下操作：
向 coordinator 请求任务，读取任务输入文件，执行任务，将任务输出写入一个或多个文件，然后再次向 coordinator 请求新任务。
coordinator 应该能够检测到在合理时间内未完成任务的 worker（本实验中为 10 秒），并将任务重新分配给其他 worker。

我们为你提供了一些起始代码。main/mrcoordinator.go 和 main/mrworker.go 中的“main”例程不能更改。
你应该将实现放在 mr/coordinator.go、mr/worker.go 和 mr/rpc.go 中。

> 这里的coordinator就是master

```shell
rm mr-out*
go run mrcoordinator.go pg-*.txt
```

pg-*.txt 是输入文件，每个文件对应一个“split”，是 Map 任务的输入。
在其他窗口中运行一些 workers：

```shell
go run mrworker.go wc.so
```

当 workers 和 coordinator 完成后，查看 mr-out-* 文件中的输出。
完成实验后，将所有 reduce 任务的输出文件按排序合并，应与顺序版本的输出匹配：

> $ cat mr-out-* | sort | more
> A 509
> ABOUT 2
> ACT 8
> ...

一个测试脚本 main/test-mr.sh，用于检查 wc 和 indexer MapReduce 应用程序的输出是否正确，
以及你的实现是否并行运行 Map 和 Reduce 任务，并从 worker 崩溃中恢复。

如果你现在运行测试脚本，它会卡住，因为 coordinator 不会退出。
你可以在 mr/coordinator.go 中的 Done 函数中将 ret := false 改为 true，使 coordinator 立即退出。然后：

```text
$ bash test-mr.sh
*** Starting wc test.
--- wc test: FAIL
$
```

测试脚本期望输出文件名为 mr-out-X，每个 reduce 任务一个文件。
空实现的 mr/coordinator.go 和 mr/worker.go 不会产生这些文件（也不会做太多其他事情），所以测试失败。

### 规则
Map 阶段应将中间键值对分成 nReduce 个 reduce 任务的桶，其中 nReduce 是 main/mrcoordinator.go 传递给 MakeCoordinator() 的参数。每个 mapper 应创建 nReduce 个中间文件，供 reduce 任务使用。
worker 实现应将 reduce 任务的输出放在 mr-out-X 文件中。
reduce 函数的输出应使用 Go 的 %v %v 格式生成，与 main/mrsequential.go 中的格式一致。测试脚本会检查格式是否正确。
你可以修改 mr/worker.go、mr/coordinator.go 和 mr/rpc.go。可以临时修改其他文件进行测试，但确保你的代码与原始版本兼容，因为我们将在原始版本上进行测试。
worker 应将中间 Map 输出放在当前目录的文件中，以便后续的 Reduce 任务可以读取。
main/mrcoordinator.go 期望 mr/coordinator.go 实现一个 Done() 方法，当 MapReduce 作业完全完成后返回 true，此时 mrcoordinator.go 将退出。
当作业完成后，worker 进程应退出。一种实现方式是检查 call() 的返回值：
如果 worker 无法联系到 coordinator，可以假设 coordinator 已经因为作业完成而退出，因此 worker 可以终止。根据设计，你也可以让 coordinator 给 worker 发送“退出”伪任务。

### 提示
开发和调试指南 有一些开发和调试的提示。
一种开始方式是修改 mr/worker.go 中的 Worker()，通过 RPC 向 coordinator 请求任务。然后修改 coordinator 以响应未启动的 map 任务的文件名。然后修改 worker 以读取该文件并像 mrsequential.go 中那样调用 Map 函数。
应用程序的 Map 和 Reduce 函数在运行时通过 Go 的插件 包加载，文件名以 .so 结尾。
如果你更改了 mr/ 目录中的任何内容，可能需要重新构建使用的任何 MapReduce 插件，例如 go build -buildmode=plugin ../mrapps/wc.go。
本实验依赖于 workers 共享文件系统。在同一台机器上运行所有 workers 时这很简单，但如果 workers 运行在不同机器上，则需要全局文件系统（如 GFS）。
中间文件的合理命名约定是 mr-X-Y，其中 X 是 Map 任务号，Y 是 reduce 任务号。
worker 的 Map 任务代码需要一种方法来存储中间键值对，以便 Reduce 任务可以正确读取。可以使用 Go 的 encoding/json 包。例如，将键值对写入文件：

```go
// 键值对写入
enc := json.NewEncoder(file)
for _, kv := ... {
    err := enc.Encode(&kv)
}
// 读取
dec := json.NewDecoder(file)
for {
var kv KeyValue
if err := dec.Decode(&kv); err != nil {
break
}
kva = append(kva, kv)
}
```

Map 阶段可以使用 ihash(key) 函数（在 worker.go 中）来选择键对应的 reduce 任务。
你可以从 mrsequential.go 中借用一些代码，用于读取 Map 输入文件、在 Map 和 Reduce 之间排序中间键值对，以及将 Reduce 输出存储在文件中。
coordinator 作为 RPC 服务器，需要处理共享数据的锁。
使用 Go 的 race 检测器（go run -race）。test-mr.sh 中有关于如何使用 -race 的注释。我们测试你的代码时不会使用 race 检测器，但如果你的代码有 race，即使没有检测器，也可能失败。
workers 经常需要等待，例如，在所有 Map 完成之前 Reduce 不能开始。一种方法是让 worker 周期性地通过 time.Sleep() 向 coordinator 请求工作，或者让 coordinator 的 RPC 处理程序在等待时使用 time.Sleep() 或 sync.Cond。Go 在其自己的线程中运行每个 RPC 处理程序，因此一个处理程序的等待不会阻止 coordinator 处理其他 RPC。
coordinator 难以区分崩溃的 worker、alive 但卡住的 worker 和执行缓慢的 worker。最好的办法是让 coordinator 等待一段时间（本实验中为 10 秒），然后假设 worker 已死亡。当然，它可能没有。
如果你实现备份任务（第 3.6 节），请注意，当 workers 在不崩溃的情况下执行任务时，不应安排额外的任务。备份任务应在相对较长时间（例如 10 秒）后才调度。
你可以使用 mrapps/crash.go 应用程序插件来测试崩溃恢复。它会在 Map 和 Reduce 函数中随机退出。
为了防止在崩溃情况下观察到部分写入的文件，MapReduce 论文提到了使用临时文件并在完全写入后原子重命名的技巧。可以使用 ioutil.TempFile 或 os.CreateTemp（如果你使用的是 Go 1.17 或更高版本）来创建临时文件，并使用 os.Rename 进行原子重命名。

## Lab2: Raft

[lab2-link](http://nil.csail.mit.edu/6.824/2020/labs/lab-raft.html)

你将做一系列的实验区构建一个具有容错能力的K/V存储系统。你需要实现一个Raft(一个replicated state machine, 一个复制状态机)。下一个实验将基于Raft构建一个K/V服务。

raft将客户端请求收集为一个序列，称为log,并确保所有发备份服务器看到相同的log，每一个副本根据log的顺序执行客户端的请求，保持所有副本的状态相同。如果一个服务器挂了但是等会儿恢复，Raft将帮助它跟进log。

In this lab you'll implement Raft as a Go object type with associated methods, meant to be used as a module in a larger service. A set of Raft instances talk to each other with RPC to maintain replicated logs. Your Raft interface will support an indefinite sequence of numbered commands, also called log entries. The entries are numbered with index numbers. The log entry with a given index will eventually be committed. At that point, your Raft should send the log entry to the larger service for it to execute.

在这个实验中，你将实现一个Raft对象类型，它有一组关联的方法，用作一个更大的服务的模块。一组Raft实例通过RPC与每个其他实例通信，以维护复制日志。你的Raft接口将支持一个无限序列的编号命令，也称为日志条目。具有给定索引的日志条目最终将被提交。在那时，你的Raft应该将日志条目发送给更大的服务，让它执行。

You should follow the design in the extended Raft paper, with particular attention to Figure 2. You'll implement most of what's in the paper, including saving persistent state and reading it after a node fails and then restarts. You will not implement cluster membership changes (Section 6). You'll implement log compaction / snapshotting (Section 7) in a later lab.

你应该遵循[扩展的Raft论文的设计](http://nil.csail.mit.edu/6.824/2020/papers/raft-extended.pdf)，特别关注图2。你将实现大部分论文中的内容，包括保存持久状态并在节点失败后重新启动后读取它。你不会实现集群成员更改（第6节）。你将在以后的实验中实现日志压缩/快照（第7节）。

You may find this guide useful, as well as this advice about locking and structure for concurrency. For a wider perspective, have a look at Paxos, Chubby, Paxos Made Live, Spanner, Zookeeper, Harp, Viewstamped Replication, and Bolosky et al.

你可能会发现这个[指南](https://thesquareplanet.com/blog/students-guide-to-raft/)很有用，以及关于并发的[锁定](http://nil.csail.mit.edu/6.824/2020/labs/raft-locking.txt)和[结构](http://nil.csail.mit.edu/6.824/2020/labs/raft-structure.txt)的建议。对于更广泛的视角，看看Paxos，Chubby，Paxos Made Live，Spanner，Zookeeper，Harp，Viewstamped Replication和[Bolosky等人](http://static.usenix.org/event/nsdi11/tech/full_papers/Bolosky.pdf)。

We supply you with skeleton code src/raft/raft.go. We also supply a set of tests, which you should use to drive your implementation efforts, and which we'll use to grade your submitted lab. The tests are in src/raft/test_test.go.

我们提供了src/raft/raft.go的框架代码，你就在这儿写代码。我们还提供了一组测试，您应该使用它们来驱动您的实现努力，我们将使用它们来评分您提交的实验。测试位于src/raft/test_test.go。

运行：

```shell
cd src/raft
go test
```

A service calls Make(peers,me,…) to create a Raft peer. The peers argument is an array of network identifiers of the Raft peers (including this one), for use with RPC. The me argument is the index of this peer in the peers array. Start(command) asks Raft to start the processing to append the command to the replicated log. Start() should return immediately, without waiting for the log appends to complete. The service expects your implementation to send an ApplyMsg for each newly committed log entry to the applyCh channel argument to Make().

通过调用`Make(peers,me,…)`来创建一个Raft对等点。peers参数是Raft对等点的网络标识符数组（包括这个），用于RPC。me参数是peers数组中这个对等点的索引。Start(command)要求Raft开始将命令附加到复制日志。Start()应该立即返回，而不等待日志追加完成。service你的代码可以为每个新提交的日志条目发送一个ApplyMsg到applyCh通道参数。

raft.go contains example code that sends an RPC (sendRequestVote()) and that handles an incoming RPC (RequestVote()). Your Raft peers should exchange RPCs using the labrpc Go package (source in src/labrpc). The tester can tell labrpc to delay RPCs, re-order them, and discard them to simulate various network failures. While you can temporarily modify labrpc, make sure your Raft works with the original labrpc, since that's what we'll use to test and grade your lab. Your Raft instances must interact only with RPC; for example, they are not allowed to communicate using shared Go variables or files.

raft.go包含一个例子，它发送一个RPC（sendRequestVote()）和一个传入的RPC（RequestVote()）。你的Raft对等点应该使用labrpc Go包（源在src/labrpc）来交换RPC。tester可以告诉labrpc延迟RPC，重新排序它们，并且丢弃它们来模拟各种网络故障。虽然你可以暂时修改labrpc，但确保你的Raft与原始的labrpc一起工作，因为这就是我们将用来测试和评分你的实验。你的Raft实例必须只与RPC交互；例如，它们不允许使用共享的Go变量或文件进行通信。

Subsequent labs build on this lab, so it is important to give yourself enough time to write solid code.

后续实验基于这个实验，所以给你足够的时间来编写坚实的代码。

### Part 2A

**Task**

Implement Raft leader election and heartbeats (AppendEntries RPCs with no log entries). The goal for Part 2A is for a single leader to be elected, for the leader to remain the leader if there are no failures, and for a new leader to take over if the old leader fails or if packets to/from the old leader are lost. Run go test -run 2A to test your 2A code.

实现Raft领导者选举和心跳（带有空日志条目的AppendEntries RPC）。Part 2A的目标是选举出一个领导者，领导者在没有故障的情况下保持领导者身份，并且如果旧领导者失败或与旧领导者通信的数据包丢失，则新领导者接管。运行`go test -run 2A`来测试你的2A代码。

``` shell
go test -run 2A
```

预计输出
``` shell
Test (2A): initial election ...
  ... Passed --   4.0  3   32    9170    0
Test (2A): election after network failure ...
  ... Passed --   6.1  3   70   13895    0
PASS
ok      raft    10.187s
```

Each "Passed" line contains five numbers; these are the time that the test took in seconds, the number of Raft peers (usually 3 or 5), the number of RPCs sent during the test, the total number of bytes in the RPC messages, and the number of log entries that Raft reports were committed. Your numbers will differ from those shown here. You can ignore the numbers if you like, but they may help you sanity-check the number of RPCs that your implementation sends. For all of labs 2, 3, and 4, the grading script will fail your solution if it takes more than 600 seconds for all of the tests (go test), or if any individual test takes more than 120 seconds.

每个带有“Passed”的输出行包含5个数字，它们分别为：
1. 测试用时
2. Raft peers
3. RPC发送数
4. RPC消耗字节数
5. 提交的日志数
每个测试点的用时不超过2分钟。

Hint: 提示
【核心实现步骤】

选举机制实现:
- 参照论文图2设计状态机
- 在raft.go中添加选举相关的状态字段
- 定义日志条目结构体
- 实现RequestVoteArgs/Reply结构体
- 创建后台goroutine定期触发选举
- 编写RequestVote RPC处理程序

心跳机制实现:
- 定义AppendEntries RPC结构体
- leader定期发送心跳包
- 编写重置选举超时的处理程序

【关键注意事项】

• 选举超时随机化:
- 不同节点设置不同超时时间(建议150-300ms范围)
- 实际设置需考虑测试约束(心跳频率≤10次/秒)
- 保证5秒内完成选举(即使需要多轮投票)

• 实现细节:
- 使用time.Sleep而非Timer/Ticker
- 遵循图2的完整选举逻辑
- 实现GetState()方法
- 处理rf.Kill()状态

【调试建议】

消息跟踪调试:
- 使用DPrintf记录消息收发
- 收集输出: go test -run 2A > out
- 分析out文件中的消息流

并发检测:
- 使用go test -race检查数据竞争
- 合理使用锁机制(参考锁定指南)

常见问题排查:
- 检查字段命名是否符合RPC规范
- 验证选举超时随机化实现
- 确保心跳频率不超过限制
- 处理网络分区等边界情况