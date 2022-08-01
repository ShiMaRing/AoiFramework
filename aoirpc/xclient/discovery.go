package xclient

import (
	"errors"
	"math"
	"math/rand"
	"sync"
	"time"
)

type SelectMode int

const (
	RandomMode SelectMode = iota
	RoundRobinMode
)

// Discovery 服务发现接口
type Discovery interface {
	Refresh() error
	Update(servers []string) error
	Get(mode SelectMode) (string, error)
	GetAll() ([]string, error)
}

// MultiServerDiscovery 服务发现接口
type MultiServerDiscovery struct {
	r       *rand.Rand   //随机数生成
	mu      sync.RWMutex //锁
	servers []string     //服务列表
	index   int          //当前index
}

func (m *MultiServerDiscovery) Refresh() error {
	return nil
}

func (m *MultiServerDiscovery) Update(servers []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.servers = servers
	return nil
}

func (m *MultiServerDiscovery) Get(mode SelectMode) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	length := len(m.servers)
	if length == 0 {
		return "", errors.New("rpc discovery: no available servers")
	}
	switch mode {
	case RandomMode:
		return m.servers[m.r.Intn(length)], nil
	case RoundRobinMode:
		s := m.servers[m.index%length]
		m.index = (m.index + 1) % length
		return s, nil
	default:
		return "", errors.New("rpc discovery: not supported select mode")
	}
}

func (m *MultiServerDiscovery) GetAll() ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	servers := make([]string, len(m.servers), len(m.servers))
	copy(servers, m.servers)
	return servers, nil
}

func NewMultiServerDiscovery(servers []string) *MultiServerDiscovery {
	m := &MultiServerDiscovery{
		r:       rand.New(rand.NewSource(time.Now().Unix())),
		servers: servers,
	}
	m.index = m.r.Intn(math.MaxInt32 - 1)
	return m
}
