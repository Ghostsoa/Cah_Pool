package cah

// Config 配置类
type Config struct {
	Capacity      int32 // 协程池中的工作池大小 (Size of the worker pool in the goroutine pool)
	CommonQueue   int32 // 普通任务队列数 (Number of common task queues)
	VIPQueue      int32 // VIP 任务队列数 (Number of VIP task queues)
	UrgentQueue   int32 // 紧急任务队列数 (Number of urgent task queues)
	EnableCluster bool  // 是否开启集群模式 (Whether to enable cluster mode)
	NodeCapacity  int32 // 集群中单个节点的容量 (Capacity of a single node in the cluster)
}
