package cah

// inputScheduler 入口任务分发调度 (Entry task dispatch scheduler)
func (p *Pool) inputScheduler(t task) error {
	if t.f == nil { // 如果任务函数为 nil，返回 nil (Return nil if the task function is nil)
		return nil
	}
	// 优先级处理 (Priority handling)
	switch t.p {
	case 3: // 紧急任务 (Urgent task)
		p.urgentChan <- t.f // 将任务添加到紧急通道 (Add the task to the urgent channel)
	case 2: // VIP任务 (VIP task)
		p.vipChan <- t.f // 将任务添加到VIP通道 (Add the task to the VIP channel)
	default: // 普通任务 (Common task)
		p.commonChan <- t.f // 将任务添加到普通通道 (Add the task to the common channel)
	}
	p.addTaskWaiting(1) // 增加等待任务计数 (Increment the waiting task counter)
	return nil          // 返回 nil (Return nil)
}
