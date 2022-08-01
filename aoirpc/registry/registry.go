package registry

import (
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

type ServerItem struct {
	Addr  string    //地址
	start time.Time //超时时间
}

type AoiRegister struct {
	timeout time.Duration //默认超时时间
	mu      sync.Mutex
	servers map[string]*ServerItem
}

const (
	defaultPath    = "/aoirpc/registry"
	defaultTimeout = 5 * time.Minute
	AoiKey         = "X-Aoirpc-Servers"
)

func New(duration time.Duration) *AoiRegister {
	return &AoiRegister{
		timeout: duration,
		mu:      sync.Mutex{},
		servers: make(map[string]*ServerItem),
	}
}

var DefaultRegistry = New(defaultTimeout)

//注册
func (r *AoiRegister) putServer(addr string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	server := r.servers[addr]
	if server == nil {
		r.servers[addr] = &ServerItem{
			Addr:  addr,
			start: time.Now(),
		}
	} else {
		server.start = time.Now()
	}
}
func (r *AoiRegister) aliveServers() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	var alive []string
	for addr, item := range r.servers {
		if r.timeout == 0 || item.start.Add(r.timeout).After(time.Now()) {
			alive = append(alive, addr)
		} else {
			delete(r.servers, addr)
		}
	}
	sort.Strings(alive)
	return alive
}

func (r *AoiRegister) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	//根据method类型判断
	switch req.Method {
	case http.MethodGet:
		w.Header().Set("X-Aoirpc-Servers", strings.Join(r.aliveServers(), ","))
	case http.MethodPost:
		addr := req.Header.Get("X-Aoirpc-Servers")
		if addr == "" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		r.putServer(addr)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (r *AoiRegister) HandleHTTP(registryPath string) {
	http.Handle(registryPath, r)
	log.Println("rpc registry path:", registryPath)
}

func HandleHTTP() {
	DefaultRegistry.HandleHTTP(defaultPath)
}

//心跳检测
func sendHeartbeat(registry, addr string) error {
	log.Println(addr, "send heart beat to registry", registry)
	httpclient := &http.Client{}
	request, _ := http.NewRequest(http.MethodPost, registry, nil)
	request.Header.Set("X-Aoirpc-Servers", addr)
	if _, err := httpclient.Do(request); err != nil {
		log.Println("rpc server: heart beat err:", err)
		return err
	}
	return nil
}

func Heartbeat(registry, addr string, duration time.Duration) {
	//检查时间周期
	if duration == 0 {
		duration = defaultTimeout - 1*time.Minute
	}
	var err error
	err = sendHeartbeat(registry, addr)
	go func() {
		ticker := time.NewTicker(duration)
		defer ticker.Stop()
		for err == nil {
			select {
			case <-ticker.C:
				err = sendHeartbeat(registry, addr)
			}
		}
	}()
}
