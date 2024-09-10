package cah

import (
	"errors"
	"log"
	"sync"
	"sync/atomic"
)

// Task 结构体来封装任务和优先级
type task struct {
	f func() // 要执行的任务 (The function to be executed)
	p int32  // 任务的优先级 (Priority of the task)
}

// Pool 结构体表示一个协程池
type Pool struct {
	capacity     int32       // 协程池的最大容量，即允许同时运行的工作协程的最大数量 (Maximum capacity of the pool, i.e., maximum number of worker goroutines running simultaneously)
	running      int32       // 当前正在运行的协程数 (Current number of running goroutines)
	waiting      int32       // 等待执行的协程数 (Number of goroutines waiting to execute)
	taskWaiting  int32       // 等待执行的任务 (Number of tasks waiting to execute)
	start        bool        // 记录协程池是否已启动 (Flag indicating whether the goroutine pool has started)
	vipChan      chan func() // 优先通道 (Channel for VIP tasks)
	commonChan   chan func() // 普通通道 (Channel for common tasks)
	urgentChan   chan func() // 紧急通道 (Channel for urgent tasks)
	worker       goWorker    // 工作协程 (Worker goroutine)
	reusablePool sync.Pool   // 可重用的回收池 (Pool for reusable objects)
	cond         sync.Cond   // 条件变量 (Condition variable)
	once         *sync.Once  // 确保某些初始化操作只执行一次 (Ensures certain initialization operations are executed only once)
	Lock         spinLock    // 自定义自旋锁 (Custom spin lock)
	config       Config      // 自定义配置 (Custom configuration)
}

// NewOnePool 创建一个新的协程池实例 (Creates a new instance of the goroutine pool)
func NewOnePool(config Config) (*Pool, error) {

	// 初始化协程池 (Initialize the goroutine pool)
	p := &Pool{
		capacity: config.Capacity, // 设置池的最大容量 (Set the maximum capacity of the pool)
		once:     &sync.Once{},    // 初始化 sync.Once 以确保某些操作只执行一次 (Initialize sync.Once to ensure some operations are executed only once)
		start:    true,            // 设置协程池为已启动 (Set the goroutine pool as started)
		config:   config,          // 存储用户配置 (Store user configuration)
	}

	// 初始化条件变量 (Initialize the condition variable)
	p.cond = sync.Cond{L: &p.Lock}

	// 设置可重用对象池中的新对象创建方法 (Set the method for creating new objects in the reusable pool)
	p.reusablePool.New = func() interface{} {
		return &goWorker{
			pool: p, // 将当前池传递给工作器 (Pass the current pool to the worker)
		}
	}

	// 初始化普通任务管道 (Initialize the common task channel)
	if config.CommonQueue == 0 {
		p.commonChan = make(chan func(), 3000000) // 默认容量为3000000 (Default capacity is 3,000,000)
	} else {
		p.commonChan = make(chan func(), config.CommonQueue) // 使用用户指定的容量 (Use user-specified capacity)
	}

	// 初始化VIP任务管道 (Initialize the VIP task channel)
	if config.VIPQueue == 0 {
		p.vipChan = make(chan func(), 1000000) // 默认容量为1000000 (Default capacity is 1,000,000)
	} else {
		p.vipChan = make(chan func(), config.VIPQueue) // 使用用户指定的容量 (Use user-specified capacity)
	}

	// 初始化紧急任务管道 (Initialize the urgent task channel)
	if config.UrgentQueue == 0 {
		p.urgentChan = make(chan func(), 500) // 默认容量为500 (Default capacity is 500)
	} else {
		p.urgentChan = make(chan func(), config.UrgentQueue) // 使用用户指定的容量 (Use user-specified capacity)
	}

	// 初始话goWorker (Initialize the goWorker)
	p.worker = goWorker{pool: p}

	// 启动一个goWorker (Start a goWorker)
	p.worker.run()

	return p, nil // 返回协程池实例 (Return the goroutine pool instance)
}

// Submit 发送任务接口 (Submit task interface)
func (p *Pool) Submit(taskIn func(), priority ...int) error {

	// 判断是否已经关闭协程池 (Check if the goroutine pool is closed)
	if p.IsClosed() {
		return errors.New("pool is closed") // 返回错误信息 (Return error message)
	}

	// 判断是否传入任务优先级标识，并格式化标识 (Check if a priority identifier is provided and format it)
	// 如果未开启优先级管道，将不会初始化优先级管道，同时标识无效 (If priority channels are not enabled, it will not initialize and the identifier will be invalid)
	if len(priority) == 0 {
		t := task{f: taskIn, p: 0} // 默认优先级为0 (Default priority is 0)
		err := p.inputScheduler(t) // 将任务输入调度器 (Input task to the scheduler)
		if err != nil {
			return err // 返回调度错误 (Return scheduling error)
		}
		return nil // 返回成功 (Return success)
	}

	pry := int32(priority[0])    // 获取优先级 (Get priority)
	t := task{f: taskIn, p: pry} // 创建任务 (Create task)
	err := p.inputScheduler(t)   // 将任务输入调度器 (Input task to the scheduler)
	if err != nil {
		return err // 返回调度错误 (Return scheduling error)
	}

	return nil // 返回成功 (Return success)
}

// IsClosed 检查协程池是否关闭 (Check if the goroutine pool is closed)
func (p *Pool) IsClosed() bool {
	if p.start {
		return false // 如果协程池已启动，返回false (If the pool has started, return false)
	}
	return true // 否则返回true (Otherwise, return true)
}

// 增加运行中的协程计数器 (Increment the running goroutine counter)
func (p *Pool) addRunning(number int32) {
	atomic.AddInt32(&p.running, number) // 原子地增加正在运行的协程计数 (Atomically increment the running goroutine count)
}

// Running 返回运行中的协程计数 (Return the count of running goroutines)
func (p *Pool) reRunning() int32 {
	return atomic.LoadInt32(&p.running) // 原子地加载当前正在运行的协程计数 (Atomically load the current running goroutine count)
}

// 增加等待任务的计数器 (Increment the waiting task counter)
func (p *Pool) addWaiting(number int32) {
	atomic.AddInt32(&p.waiting, number) // 原子地增加等待任务的计数 (Atomically increment the waiting task count)
}

// 增加等待任务的计数器 (Increment the waiting tasks counter)
func (p *Pool) addTaskWaiting(number int32) {
	atomic.AddInt32(&p.taskWaiting, number) // 原子地增加等待执行的任务计数 (Atomically increment the waiting tasks count)
}

// 返回等待执行任务的计数 (Return the count of waiting tasks)
func (p *Pool) reTaskWaiting() int32 {
	return atomic.LoadInt32(&p.taskWaiting) // 原子地加载当前等待执行的任务计数 (Atomically load the current waiting tasks count)
}

// 返回等待的协程计数 (Return the count of waiting goroutines)
func (p *Pool) reWaiting() int32 {
	return atomic.LoadInt32(&p.waiting) // 原子地加载当前等待的协程计数 (Atomically load the current waiting goroutines count)
}

// loggingError 捕获 panic 并输出日志 (loggingError captures panic and logs it)
func (p *Pool) loggingError(r interface{}) {
	// 使用标准库的 log 包或其他日志库 (Use the standard library log package or another logging library)
	log.Printf("Worker panicked: %v", r) // 输出 panic 信息 (Log panic information)
}

// Close 安全关闭协程池 (Safely close the goroutine pool)
func (p *Pool) Close() {
	p.once.Do(func() { // 确保只执行一次 (Ensure this runs only once)
		p.start = false     // 设置协程池状态为关闭 (Set the pool state to closed)
		p.urgentChan <- nil // 发送一个 nil，通知工作者不再接受新任务 (Send a nil to notify workers not to accept new tasks)
		close(p.urgentChan) // 关闭紧急任务管道 (Close the urgent task channel)
		close(p.vipChan)    // 关闭VIP任务管道 (Close the VIP task channel)
		close(p.commonChan) // 关闭普通任务管道 (Close the common task channel)
		return
	})
}

// Cluster 表示一个协程池集群 (Cluster represents a cluster of goroutine pools)
type Cluster struct {
	Pools       []*Pool // 存储节奏中的协程池 (Stores the pools in the cluster)
	nextPoolIdx int32   // 下一个池的索引 (Index of the next pool)
}

// NewCluster 创建一个新的集群 (Create a new cluster)
func NewCluster(config Config) (*Cluster, error) {
	if config.EnableCluster { // 如果启用了集群 (If clustering is enabled)
		nodeCapacity := config.NodeCapacity                             // 获取每个节点的容量 (Get the capacity of each node)
		numNodes := (config.Capacity + nodeCapacity - 1) / nodeCapacity // 计算节点数 (Calculate the number of nodes)
		eachNodeCapacity := config.Capacity / numNodes                  // 计算每个节点的容量 (Calculate the capacity for each node)

		pools := make([]*Pool, numNodes) // 初始化协程池数组 (Initialize the pools array)
		for i := int32(0); i < numNodes; i++ {
			poolConfig := Config{ // 配置每个节点的配置 (Configure each node)
				UrgentQueue: config.UrgentQueue,
				VIPQueue:    config.VIPQueue,
				CommonQueue: config.CommonQueue,
				Capacity:    eachNodeCapacity, // 设置每个节点的容量 (Set each node's capacity)
			}
			pool, err := NewOnePool(poolConfig) // 创建新的协程池 (Create a new goroutine pool)
			if err != nil {
				return nil, err // 返回错误 (Return error)
			}
			pools[i] = pool // 存储创建的协程池 (Store the created goroutine pool)
		}
		return &Cluster{Pools: pools}, nil // 返回集群实例 (Return the cluster instance)
	} else { // 如果未启用集群 (If clustering is not enabled)
		pool, err := NewOnePool(config) // 创建单个协程池 (Create a single goroutine pool)
		if err != nil {
			return nil, err // 返回错误 (Return error)
		}
		return &Cluster{Pools: []*Pool{pool}}, nil // 返回集群实例 (Return the cluster instance)
	}
}

// Submit 将任务提交到集群中的某个池 (Submit a task to one of the pools in the cluster)
func (c *Cluster) Submit(taskIn func(), priority ...int) error {
	// 确保集群不为空 (Ensure the cluster is not empty)
	if len(c.Pools) == 0 {
		return errors.New("no pools available in the cluster") // 返回错误信息 (Return error message)
	}
	// 计算要发送的池的索引 (Calculate the index of the pool to send to)
	poolIndex := int(atomic.AddInt32(&c.nextPoolIdx, 1)-1) % len(c.Pools)

	// 获取对应的池并提交任务 (Get the corresponding pool and submit the task)
	pool := c.Pools[poolIndex]
	return pool.Submit(taskIn, priority...) // 提交任务到池 (Submit the task to the pool)
}

// Close 安全关闭集群中的所有协程池 (Safely close all goroutine pools in the cluster)
func (c *Cluster) Close() {
	for i := len(c.Pools) - 1; i >= 0; i-- { // 从后向前关闭池 (Close the pools from back to front)
		c.Pools[i].Close() // 调用每个池的 Close 方法 (Call the Close method on each pool)
	}
}
