package cah

import (
	"sync"
	"testing"
	"time"
)

// BenchmarkPool 的基准测试
func Benchmark_cah(b *testing.B) {
	pool, err := NewOnePool(Config{UrgentQueue: 50, Capacity: 30000})
	if err != nil {
		b.Fatalf("Error creating Pool1: %v", err)
	}

	b.ReportAllocs() // 报告内存分配统计
	b.ResetTimer()   // 重置计时器

	var wg sync.WaitGroup

	// 使用两个 goroutine 来处理提交的任务
	for i := 0; i < b.N; i++ {
		wg.Add(1)

		var err error
		err = pool.Submit(func() {
			defer wg.Done()
			// 模拟任务执行
			time.Sleep(10 * time.Millisecond)
		})

		if err != nil {
			b.Errorf("Error submitting task %d: %v", i, err)
		}
	}

	wg.Wait() // 等待所有任务完成

	b.StopTimer() // 停止计时器

	pool.Close()

}

// 集群测试
func Benchmark_cah_Cluster(b *testing.B) {
	// 配置集群参数
	config := Config{
		EnableCluster: true,   // 启用集群
		Capacity:      200000, // 总容量
		NodeCapacity:  50000,
	}

	cluster, err := NewCluster(config)
	if err != nil {
		b.Fatalf("Error creating cluster: %v", err)
	}
	defer cluster.Close() // 确保测试完成后关闭集群

	b.ReportAllocs() // 报告内存分配统计
	b.ResetTimer()   // 重置计时器

	var wg sync.WaitGroup

	// 使用两个 goroutine 来处理提交的任务
	for i := 0; i < b.N; i++ {
		wg.Add(1)

		err := cluster.Submit(func() {
			defer wg.Done()
			// 模拟任务执行
			time.Sleep(10 * time.Millisecond)
		})

		if err != nil {
			b.Errorf("Error sending task %d: %v", i, err)
		}
	}

	wg.Wait()     // 等待所有任务完成
	b.StopTimer() // 停止计时器
}
