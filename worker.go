package cah

import (
	"math/rand"
	"time"
)

// goWorker 表示一个工作协程 (goWorker represents a worker goroutine)
type goWorker struct {
	pool *Pool // 所属的协程池 (The pool to which it belongs)
}

// run 将一个goWorker启动 (Start a goWorker)
func (w *goWorker) run() {
	w.pool.addRunning(1) // 增加正在运行的工作者计数 (Increment the count of running workers)
	go func() {
		// defer 用于确保在出现问题后，执行必要的清理和回收工作 (Ensure necessary cleanup and recycling in case of a problem)
		defer func() {
			w.pool.addRunning(-1)         // 减少正在运行的工作者计数 (Decrement the count of running workers)
			w.pool.reusablePool.Put(w)    // 将当前工作者放入回收池中 (Put the current worker into the reusable pool)
			if r := recover(); r != nil { // 捕获并处理恐慌 (Recover and handle panic)
				w.pool.loggingError(r) // 记录错误 (Log the error)
			}
		}()

		for {
			// 从任务管道获取一个任务 (Get a task from the task pipeline)
			f := w.getTask()
			if f == nil { // 如果没有任务 (If there are no tasks)
				for {
					// 检查是否有等待的任务 (Check for waiting tasks)
					if w.pool.reWaiting() > 0 && w.pool.reRunning() == 1 {
						w.pool.cond.Broadcast() // 广播所有等待的工作者 (Broadcast to all waiting workers)
						return
					} else if w.pool.reWaiting() == 0 && w.pool.reRunning() == 1 {
						return // 如果没有等待任务且当前仅有一个运行中的工作者 (Return if no waiting tasks and only one running worker)
					} else {
						time.Sleep(500 * time.Millisecond) // 休眠一段时间 (Sleep for a while)
						continue
					}
				}
			}

			if w.pool.reRunning() < w.pool.capacity { // 如果运行中的工作者小于池的容量 (If running workers are less than the pool capacity)
				if w.pool.running > 30000 { // 如果正在运行的工作者数量超过30000 (If the running workers exceed 30000)
					// 当运行的goroutine达到一定数量时，采取随机抽取策略放行 (Adopt a random extraction strategy to release when the number of running goroutines reaches a certain level)
					if w.lottery() {
						w.check() // 检查条件 (Check conditions)
					}
				}
				w.check() // 再次检查 (Check again)
			}

			// 执行任务 (Execute the task)
			f()
			w.pool.addTaskWaiting(-1) // 减少等待任务计数 (Decrement the waiting task count)

			// 判断是否需要继续工作 (Check if the worker needs to continue working)
			if w.pool.reRunning() < w.pool.reTaskWaiting() {
				continue
			}

			if w.pool.IsClosed() { // 如果池已经关闭 (If the pool is closed)
				return
			}

			// goWorker挂起，等待被重新唤醒 (Suspend the goWorker and wait to be awakened)
			w.pool.Lock.Lock()           // 加锁 (Lock)
			if w.pool.reRunning() == 1 { // 如果当前只有一个正在运行的工作者 (If currently only one running worker)
				w.pool.Lock.Unlock() // 解锁 (Unlock)
				continue
			}

			// 阻塞挂起 (Block and suspend)
			w.wait()           // 等待 (Wait)
			w.pool.cond.Wait() // 等待条件 (Wait for condition)
			w.wake()           // 唤醒 (Wake up)

			// 唤醒后续操作 (Subsequent operations after waking up)
			if w.pool.IsClosed() {
				w.pool.Lock.Unlock() // 解锁 (Unlock)
				return
			}
			// 判断存在的goWorker是否大于等待的任务 (Check if existing goWorkers exceed waiting tasks)
			if w.pool.reRunning()+w.pool.reWaiting() > w.pool.reTaskWaiting() {
				w.pool.Lock.Unlock() // 解锁 (Unlock)
				return
			}
			w.pool.Lock.Unlock() // 解锁 (Unlock)
		}
	}()
}

// getTask 从任务管道获取任务 (Get a task from the task pipeline)
func (w *goWorker) getTask() func() {
	defer func() {
		if r := recover(); r != nil { // 捕获并处理恐慌 (Recover and handle panic)
			w.pool.loggingError(r) // 记录错误 (Log the error)
		}
	}()
	// 尝试从任务管道取出任务并执行 (Try to take out tasks from the task pipeline)
	for {
		select {
		case t := <-w.pool.urgentChan: // 从紧急通道获取任务 (Get task from the urgent channel)
			return t // 返回任务 (Return the task)
		default:
			select {
			case t := <-w.pool.vipChan: // 从VIP通道获取任务 (Get task from the VIP channel)
				return t // 返回任务 (Return the task)
			default:
				select {
				case t := <-w.pool.commonChan: // 从普通通道获取任务 (Get task from the common channel)
					return t // 返回任务 (Return the task)
				default:
					continue // 如果没有任务，继续循环 (If no task, continue the loop)
				}
			}
		}
	}
}

// lottery 随机抽取 (Random extraction)
func (w *goWorker) lottery() bool {
	if w.pool.capacity-w.pool.running <= 0 { // 如果当前没有可用的容量 (If there is no available capacity)
		return false // 返回 false (Return false)
	}
	// 使用概率进行判断 (Determine using probability)
	return rand.Float64() < float64(w.pool.running)/float64(w.pool.capacity-w.pool.running)
}

// check 检查是否要创建或者唤醒一个goWorker (Check whether to create or wake up a goWorker)
func (w *goWorker) check() {
	w.pool.Lock.Lock() // 加锁 (Lock)
	// 如果有等待的任务并且没有正在运行工作者且当前工作者数量小于池的容量 (If there are waiting tasks, no running workers, and the current worker count is less than the pool capacity)
	if w.pool.reTaskWaiting() > w.pool.reRunning() && w.pool.reWaiting() == 0 && w.pool.reRunning()+w.pool.reWaiting() < w.pool.capacity {
		newW := w.pool.reusablePool.Get().(*goWorker) // 从回收池中获取一个新的工作者 (Get a new worker from the reusable pool)
		newW.run()                                    // 启动新的工作者 (Run the new worker)
		w.pool.Lock.Unlock()                          // 解锁 (Unlock)
		return
	} else if w.pool.waiting > 0 { // 如果有等待的工作者 (If there are waiting workers)
		w.pool.cond.Signal() // 唤醒一个等待的工作者 (Wake up one waiting worker)
	}
	w.pool.Lock.Unlock() // 解锁 (Unlock)
}

// wait 挂起前操作 (Operations before suspension)
func (w *goWorker) wait() {
	w.pool.addRunning(-1) // 减少正在运行的工作者计数 (Decrement the count of running workers)
	w.pool.addWaiting(1)  // 增加等待的工作者计数 (Increment the count of waiting workers)
}

// wake 唤醒后操作 (Operations after waking up)
func (w *goWorker) wake() {
	w.pool.addRunning(1)  // 增加正在运行的工作者计数 (Increment the count of running workers)
	w.pool.addWaiting(-1) // 减少等待的工作者计数 (Decrement the count of waiting workers)
}
