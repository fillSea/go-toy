本项目采用完全基于 channel+select 的实现方案，不使用其他数据结构，也不使用 sync 包提供的各种同步结构，比如 Mutex、RWMutex，以及 Cond 等。

# 1. 架构
workerpool 的实现主要分为三个部分：
- pool 的创建与销毁。
- pool 中 worker（Goroutine）的管理。
- task 的提交与调度。
工作流程如下图所示，
![alt text](image/%E6%9E%B6%E6%9E%84%E5%9B%BE.png)
- capacity 是 pool 的一个属性，代表整个 pool 中 worker 的最大容量。
- active channel：是一个带缓冲的 channel，作为 worker 的“计数器”。
	- 当 active channeql 可写时，我们就创建一个 worker，用于处理用户通过 Schedule 函数提交的待处理的请求。
	- 当 active channel 满了的时候，pool 就会停止 worker 的创建，直到某个 worker 因故退出， active channel 又空出一个位置时，pool 才会创建新的 worker 填补那个空位。
- Task：是抽象的用户要提交给 workerpool 执行的请求。Task 通过 Schedule 函数提交到一个 task channel 中，已经创建的 worker 将从这个 task channel 中读取 task 并执行。
workerpool 包对外主要提供三个 API，它们分别是：
- workerpool. New：用于创建一个 pool 类型实例，并将 pool 池的 worker 管理机制运行起来。
- workerpool. Free：用于销毁一个 pool 池，停掉所有 pool 池中的 worker。
- Pool. Schedule：这是 Pool 类型的一个导出方法，workerpool 包的用户通过该方法向 pool 池提交待执行的任务（Task）。
# 2. 测试代码
```go
func main() {
	p := workerpool.New(5, workerpool.WithPreAllocWorkers(false), workerpool.WithBlock(false))

	time.Sleep(2 * time.Second)
	for i := 0; i < 10; i++ {
		err := p.Schedule(func() {
			time.Sleep(time.Second * 3)
		})
		if err != nil {
			fmt.Printf("task[%d]: error: %s\n", i, err.Error())
		}
	}

	p.Free()
}
```