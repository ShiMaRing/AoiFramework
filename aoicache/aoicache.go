package aoicache

import (
	"fmt"
	"log"
	"sync"
)

// Group 核心代码处，提供外部交互方法
type Group struct {
	name      string //group名称
	getter    Getter //回调函数，查询不到数据时执行
	mainCache cache  //提供数据支持
}

var (
	mu     sync.RWMutex //读写锁，保护groups
	groups = make(map[string]*Group)
)

// NewGroup 直接覆盖
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil getter !")
	}
	mu.Lock()
	defer mu.Unlock()
	group := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
	}
	groups[name] = group
	return group
}

func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

//核心方法
func (g *Group) Get(key string) (Data, error) {
	if key == "" {
		return Data{}, fmt.Errorf("key is required") //不允许空键
	}
	get, exist := g.mainCache.get(key)
	if exist {
		log.Println("cache hit") //如果能够直接拿到的话说明缓存命中
		return get, nil
	}
	return g.load(key)
}

func (g *Group) load(key string) (Data, error) {
	return g.getByGetter(key)
}

func (g *Group) getByGetter(key string) (Data, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return Data{}, err
	}
	//反之则拿到了数据
	data := Data{bytes: clone(bytes)} //避免修改底层数据，返回克隆值

	g.mainCache.add(key, data) //得到的新值添加到cache中

	return data, nil
}
