# MIT6.824实验
## Lab1: MapReduce
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

## 规则
Map 阶段应将中间键值对分成 nReduce 个 reduce 任务的桶，其中 nReduce 是 main/mrcoordinator.go 传递给 MakeCoordinator() 的参数。每个 mapper 应创建 nReduce 个中间文件，供 reduce 任务使用。
worker 实现应将 reduce 任务的输出放在 mr-out-X 文件中。
reduce 函数的输出应使用 Go 的 %v %v 格式生成，与 main/mrsequential.go 中的格式一致。测试脚本会检查格式是否正确。
你可以修改 mr/worker.go、mr/coordinator.go 和 mr/rpc.go。可以临时修改其他文件进行测试，但确保你的代码与原始版本兼容，因为我们将在原始版本上进行测试。
worker 应将中间 Map 输出放在当前目录的文件中，以便后续的 Reduce 任务可以读取。
main/mrcoordinator.go 期望 mr/coordinator.go 实现一个 Done() 方法，当 MapReduce 作业完全完成后返回 true，此时 mrcoordinator.go 将退出。
当作业完成后，worker 进程应退出。一种实现方式是检查 call() 的返回值：
如果 worker 无法联系到 coordinator，可以假设 coordinator 已经因为作业完成而退出，因此 worker 可以终止。根据设计，你也可以让 coordinator 给 worker 发送“退出”伪任务。

## 提示
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
## 无学分挑战练习
实现你自己的 MapReduce 应用程序（参见 mrapps/* 中的示例），例如分布式 Grep（参见 MapReduce 论文的第 2.3 节）。
让你的 MapReduce coordinator 和 workers 在单独的机器上运行，就像在实际中一样。
你需要设置你的 RPC 通过 TCP/IP 而不是 Unix 套接字进行通信（参见 Coordinator.server() 中的注释行），并使用共享文件系统进行文件读写。
例如，你可以在 MIT 的 Athena 集群 上的多台机器上使用 AFS，或者在 AWS 上租用几台实例并使用 S3 进行存储。