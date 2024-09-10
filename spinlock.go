package cah

import (
	"runtime"
	"sync/atomic"
)

// spinLock 是 ants 项目中实现的自旋锁 (spinLock is a spinlock implemented from the ants project).
// 它用于在多协程环境中保护共享资源，通过自旋等待来减少上下文切换的开销。 (It is used to protect shared resources in a multi-goroutine environment, reducing context-switching overhead through spin-waiting.)

type spinLock uint32

const maxBackoff = 16 // 最大回退次数 (Maximum backoff count)

// Lock 自旋锁 (Lock the spinlock)
func (sl *spinLock) Lock() {
	backoff := 1
	for !atomic.CompareAndSwapUint32((*uint32)(sl), 0, 1) { // 尝试获取锁 (Attempt to acquire the lock)
		for i := 0; i < backoff; i++ { // 进行自旋等待 (Perform spin-waiting)
			runtime.Gosched() // 让出当前协程的执行权 (Yield the execution of the current goroutine)
		}
		if backoff < maxBackoff { // 增加回退时间 (Increase backoff time)
			backoff <<= 1
		}
	}
}

// Unlock 解锁自旋锁 (Unlock the spinlock)
func (sl *spinLock) Unlock() {
	atomic.StoreUint32((*uint32)(sl), 0) // 将锁状态重置为未锁定 (Reset the lock state to unlocked)
}
