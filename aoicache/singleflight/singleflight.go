package singleflight

import "sync"

//call 代表正在进行中，或已经结束的请求。使用 sync.WaitGroup 锁避免重入
type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

//Group 是 singleflight 的主数据结构，管理不同 key 的请求(call)。
type Group struct {
	mu sync.Mutex
	m  map[string]*call
}

func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	//如果取到了值，说明已经完成了加载
	if v, ok := g.m[key]; ok {
		g.mu.Unlock()
		v.wg.Wait()
		return v.val, v.err
	}
	c := new(call)
	c.wg.Add(1)

	g.m[key] = c
	g.mu.Unlock()

	c.val, c.err = fn()
	c.wg.Done()

	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()

	return c.val, c.err
}
