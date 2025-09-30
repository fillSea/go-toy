package workerpool

import (
	"errors"
	"fmt"
	"sync"
)

// Goroutine 池的主要代码

var (
	ErrNoIdleWorkerInPool = errors.New("no idle worker in pool") // workerpool中任务已满，没有空闲goroutine用于处理新任务
	ErrWorkerPoolFreed    = errors.New("workerpool freed")       // workerpool已终止运行
)

type Task func()

type Pool struct {
	capacity int  // workerpool大小
	preAlloc bool // 是否在创建pool的时候就预创建workers，默认值为：false
	// 当pool满的情况下，新的Schedule调用是否阻塞当前goroutine。默认值：true
	// 如果block = false，则Schedule返回ErrNoWorkerAvailInPool
	block  bool
	active chan struct{}  // active channel
	tasks  chan Task      // task channel
	wg     sync.WaitGroup // 用于在pool销毁时等待所有worker退出
	quit   chan struct{}  // 用于通知各个worker退出的信号channel
}

const (
	defaultCapacity = 100
	maxCapacity     = 10000
)

// New 创建一个新的workerpool
// capacity 为workerpool的大小
// opts 为workerpool的选项
func New(capacity int, opts ...Option) *Pool {
	if capacity <= 0 {
		capacity = defaultCapacity
	}
	if capacity > maxCapacity {
		capacity = maxCapacity
	}

	p := &Pool{
		capacity: capacity,
		preAlloc: false,
		block:    true,
		tasks:    make(chan Task),
		quit:     make(chan struct{}),
		active:   make(chan struct{}, capacity),
	}

	for _, opt := range opts {
		opt(p)
	}

	fmt.Printf("workerpool start(preAlloc=%t)\n", p.preAlloc)

	// 如果preAlloc为true，则在创建pool的时候就预创建workers
	if p.preAlloc {
		// create all goroutines and send into works channel
		for i := 0; i < p.capacity; i++ {
			p.newWorker(i + 1)
			p.active <- struct{}{}
		}
	}

	go p.run() // 启动workerpool

	return p
}

// returnTask 将任务返回给任务channel
func (p *Pool) returnTask(t Task) {
	go func() {
		p.tasks <- t
	}()
}

// run 启动workerpool
func (p *Pool) run() {
	idx := len(p.active)

	// 非预分配模式下的动态创建逻辑
	// 如果没有预先分配所有 worker，则在运行时根据任务需求动态创建
	if !p.preAlloc {
	loop:
		for t := range p.tasks {
			// 将取出的任务重新放回任务队列，避免任务丢失
			// 这一步只是为了检测任务存在，而不是实际处理任务
			p.returnTask(t)
			select {
			case <-p.quit:
				return
			case p.active <- struct{}{}:
				idx++
				p.newWorker(idx)
			// 如果 active 通道已满（达到容量上限），则跳出循环
			// 此时已创建了最大数量的 worker
			default:
				break loop
			}
		}
	}

	// 主循环：持续监听退出信号和 worker 创建机会
	// 一旦创建了足够的 worker 或完成了初始任务检测，进入此循环
	for {
		select {
		case <-p.quit:
			return
		case p.active <- struct{}{}: // 发送空的结构体到active channel，用于通知有新的worker可用
			// create a new worker
			idx++
			p.newWorker(idx)
		}
	}
}

// newWorker 创建一个新的worker
func (p *Pool) newWorker(i int) {
	p.wg.Add(1)
	go func() {
		defer func() {
			// 处理worker panic
			if err := recover(); err != nil {
				fmt.Printf("worker[%03d]: recover panic[%s] and exit\n", i, err)
				<-p.active
			}
			p.wg.Done()
		}()

		fmt.Printf("worker[%03d]: start...\n", i)

		// 处理任务
		for {
			select {
			case <-p.quit:
				fmt.Printf("worker[%03d]: exit\n", i)
				<-p.active
				return
			case t := <-p.tasks:
				fmt.Printf("worker[%03d]: receive a task\n", i)
				t()
			}
		}
	}()
}

// Schedule 向workerpool提交一个任务
// 如果workerpool已满，返回ErrNoIdleWorkerInPool错误
func (p *Pool) Schedule(t Task) error {
	select {
	case <-p.quit:
		return ErrWorkerPoolFreed
	case p.tasks <- t:
		return nil
	default:
		if p.block {
			p.tasks <- t
			return nil
		}
		return ErrNoIdleWorkerInPool
	}
}

// Free 释放workerpool
// 关闭quit channel，通知所有worker退出
// 等待所有worker退出
// 打印workerpool已释放的消息
func (p *Pool) Free() {
	close(p.quit) // make sure all worker and p.run exit and schedule return error
	p.wg.Wait()
	fmt.Printf("workerpool freed(preAlloc=%t)\n", p.preAlloc)
}
