package xclient

import (
	"AoiFramework/aoirpc/registry"
	"log"
	"net/http"
	"strings"
	"time"
)

// AoiRegistryDiscovery 实现相关接口，从服务注册服务器获取服务
type AoiRegistryDiscovery struct {
	*MultiServerDiscovery
	registry   string
	timeout    time.Duration
	lastUpdate time.Time
}

//最长更新时间为10秒,十秒更一次
const defaultUpdateTimeout = 10 * time.Second

// NewAoiRegistryDiscovery 注册中心地址，超时时间
func NewAoiRegistryDiscovery(registerAddr string, timeout time.Duration) *AoiRegistryDiscovery {
	if timeout == 0 {
		timeout = defaultUpdateTimeout
	}
	d := &AoiRegistryDiscovery{
		MultiServerDiscovery: NewMultiServerDiscovery(make([]string, 0)),
		registry:             registerAddr,
		timeout:              timeout,
	}
	return d
}

func (d *AoiRegistryDiscovery) Update(servers []string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.servers = servers
	d.lastUpdate = time.Now()
	return nil
}

func (d *AoiRegistryDiscovery) Refresh() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	//检查是否需要检查
	if d.lastUpdate.Add(d.timeout).After(time.Now()) {
		return nil
	}
	log.Println("rpc registry: refresh servers from registry", d.registry)

	resp, err := http.Get(d.registry) //获取存活列表
	if err != nil {
		log.Println("rpc registry refresh err:", err)
		return err
	}

	//去除掉两侧空格
	servers := strings.Split(resp.Header.Get(registry.AoiKey), ",")
	d.servers = make([]string, 0, len(servers))
	for _, server := range servers {
		if strings.TrimSpace(server) != "" {
			d.servers = append(d.servers, strings.TrimSpace(server))
		}
	}
	d.lastUpdate = time.Now()
	return nil
}

func (d *AoiRegistryDiscovery) Get(mode SelectMode) (string, error) {
	if err := d.Refresh(); err != nil {
		return "", err
	}
	return d.MultiServerDiscovery.Get(mode)
}

func (d *AoiRegistryDiscovery) GetAll() ([]string, error) {
	if err := d.Refresh(); err != nil {
		return nil, err
	}
	return d.MultiServerDiscovery.GetAll()
}
