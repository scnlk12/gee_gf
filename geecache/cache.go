package geecache

import (
	"geecache/lru"
	"sync"
)

type cache struct {
	mu sync.Mutex
	lru *lru.Cache
	cacheBytes int64
}

func (c *cache) add(key string, value Byteview) {
	c.mu.Lock()
	defer c.mu.Unlock()
	// 判断c.lru是否为nil 延迟初始化Lazy Initialization
	// 一个对象的延迟初始化意味着该对象的创建将会延迟至第一次使用该对象时
	// 用于提高性能 并减少程序的内存要求
	if c.lru == nil {
		c.lru = lru.New(c.cacheBytes, nil)
	}
	c.lru.Add(key, value)
}

func (c *cache) get(key string) (value Byteview, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		return
	}
	if v, ok := c.lru.Get(key); ok {
		return v.(Byteview), ok
	}
	return
}