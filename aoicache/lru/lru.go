package lru

//最近最少访，维护一个map以及一个双向链表
//一旦某一节点被访问了，就移动到链表队首处，定时淘汰链表队尾元素即可

import (
	"container/list"
)

//Cache 缓存的数据结构
type Cache struct {
	maxBytes  int64                         //最大存储数量
	nBytes    int64                         //当前存储数量
	ll        *list.List                    //链表标头，element为节点
	cache     map[string]*list.Element      //映射
	OnEvicted func(key string, value Value) //删除节点时的回调函数
}

func New(maxBytes int64, onEvicted func(key string, value Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

//entry 键值对，存储key以及值
type entry struct {
	key   string
	value Value
}

// Value 该接口需要指明存储的字节数，返回占用的内存大小
type Value interface {
	Len() int
}

// Get 获取对应的值，并将相对应的节点推送至队首，
func (c *Cache) Get(key string) (value Value, ok bool) {
	element, ok := c.cache[key]
	if !ok {
		return nil, ok
	}
	c.ll.MoveToFront(element)
	e := element.Value.(*entry)
	return e.value, ok
}

//RemoveOldest 缓存更新算法，删除末尾元素
func (c *Cache) RemoveOldest() {
	ele := c.ll.Back()
	if ele != nil { //防止没有添加元素
		c.ll.Remove(ele)
		etr := ele.Value.(*entry)
		delete(c.cache, etr.key)
		c.nBytes -= int64(etr.value.Len() + len(etr.key)) //减去存储值长以及键长
		if c.OnEvicted != nil {
			c.OnEvicted(etr.key, etr.value)
		}
	}
}

//Add 添加方法
//如果键存在，则更新对应节点的值，并将该节点移到队尾。
//不存在则是新增场景，首先队尾添加新节点 &entry{key, value}, 并字典中添加 key 和节点的映射关系。
//更新 c.nbytes，如果超过了设定的最大值 c.maxBytes，则移除最少访问的节点。
func (c *Cache) Add(key string, value Value) {
	if etr, ok := c.cache[key]; ok {
		//更新值并将其置于最后
		e := etr.Value.(*entry)
		prev := e.value.Len()
		e.value = value //应当复用原先的而非重建，节约内存
		c.nBytes += int64(value.Len() - prev)
		c.ll.MoveToFront(etr)
	} else {
		if value == nil {
			panic("invalid value (cant be nil)")
		}
		e := &entry{
			key:   key,
			value: value,
		}
		etr = c.ll.PushFront(e)
		c.cache[key] = etr
		c.nBytes += int64(len(key) + value.Len())
	}
	//直到内存足够之前，不断删除旧元素
	for c.maxBytes != 0 && c.maxBytes < c.nBytes {
		c.RemoveOldest()
	}
}

func (c *Cache) Len() int {
	return c.ll.Len()
}
