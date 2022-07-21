package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// Hash 哈希算法
type Hash func([]byte) uint32

type Map struct {
	hash    Hash
	dup     int   //一个节点对应多少个虚拟节点
	keys    []int //存储映射
	hashMap map[int]string
}

func New(replicas int, fn Hash) *Map {
	m := &Map{
		dup:     replicas,
		hash:    fn,
		hashMap: make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// Add 添加节点
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.dup; i++ {
			hash := m.hash([]byte(strconv.Itoa(i) + key))
			m.keys = append(m.keys, int(hash))
			m.hashMap[int(hash)] = key //映射的值都指向key
		}
	}
	sort.Ints(m.keys) //整理形成环
}

// Get 先映射
func (m *Map) Get(key string) string {
	res := int(m.hash([]byte(key)))
	//找到第一个比res大的
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= res
	}) //找到最小的满足该条件的i的位置,可能会出现在最后一个，此时需要环成0
	return m.hashMap[m.keys[idx%len(m.keys)]]
}
