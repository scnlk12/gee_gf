package singleflight

import "sync"

// 代表正在进行中 或已经结束的请求 使用sync.WaitGroup锁避免重入
type call struct {
	wg sync.WaitGroup
	val interface{}
	err error
}

// Group 是singleflight的主数据结构，管理不同key的请求
type Group struct {
	mu sync.Mutex // protects m
	m map[string]*call
}

// 针对相同的key 无论Do被调用多少次，函数fn只会被调用一次，等fn调用结束了，返回返回值或错误
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock() // 保护Group的成员变量m不被并发读写而加的锁
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		c.wg.Wait() // 如果请求正在进行中，则等待 阻塞 直到锁被释放
		return c.val, c.err // 请求结束，返回结果
	}
	c := new(call)
	c.wg.Add(1) // 发起请求前加锁 锁加1
	g.m[key] = c // 添加到g.m 表明key已经有对应请求在处理
	g.mu.Unlock()
	
	c.val, c.err = fn() // 调用fn，发起请求
	c.wg.Done() // 请求结束 锁减1

	g.mu.Lock()
	delete(g.m, key) // 更新g.m
	g.mu.Unlock()

	return c.val, c.err
}