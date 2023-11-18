package util

import (
	"sync"
)

// GoroutinePool 是一个用于管理并发执行的goroutines的池
type GoroutinePool struct {
	concurrency int
	semaphore   chan struct{}
	wg          sync.WaitGroup
}

// NewGoroutinePool 创建一个新的GoroutinePool
func NewGoroutinePool(concurrency int) *GoroutinePool {
	return &GoroutinePool{
		concurrency: concurrency,
		semaphore:   make(chan struct{}, concurrency),
	}
}

// Add 将任务添加到池中并开始执行
func (p *GoroutinePool) Add(task func()) {
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		p.semaphore <- struct{}{}        // 获取信号量资源
		defer func() { <-p.semaphore }() // 释放信号量资源
		task()
	}()
}

// Wait 阻塞直到所有任务完成
func (p *GoroutinePool) Wait() {
	p.wg.Wait()
}
