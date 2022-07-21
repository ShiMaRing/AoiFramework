package aoicache

import "C"
import (
	"AoiFramework/aoicache/lru"
	"sync"
)

//cache 封装并屏蔽底层数据结构
type cache struct {
	mu         sync.Mutex //持有互斥锁
	lru        *lru.Cache //底层数据结构
	cacheBytes int64      //已存储的字节数
}

func (c *cache) add(key string, value Data) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		c.lru = lru.New(c.cacheBytes, nil)
	}
	c.lru.Add(key, value)
}

func (c *cache) get(key string) (value Data, exist bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil { //没有执行add操作
		return
	}
	if v, ok := c.lru.Get(key); ok {
		return v.(Data), ok
	}
	return
}
