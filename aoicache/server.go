package aoicache

import (
	"AoiFramework/aoicache/consistenthash"
	"fmt"
	"google.golang.org/protobuf/proto"
	"log"
	"net/http"
	"strings"
	"sync"
)

const defaultPath = "/aoi/"
const defaultReplicas = 20

type HTTPPool struct {
	self        string
	basePath    string
	mu          sync.Mutex             //读写锁保护
	peers       *consistenthash.Map    //保存对应的映射关系，用来查询
	httpGetters map[string]*httpGetter //每一个链接对应的请求参数
}

func NewHTTPPool(selfPath string) *HTTPPool {
	return &HTTPPool{
		self:     selfPath,
		basePath: defaultPath,
	}
}

func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

//ServeHTTP 实现http handler接口,根据请求参数由本地获取对应的值
func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//先判断是否属于正确服务
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("HTTPPool serving unexpected path: " + r.URL.Path))
		return
	}
	p.Log("%s %s", r.Method, r.URL.Path)

	//尝试获取参数
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		//说明参数缺失
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	groupName := parts[0]
	key := parts[1]

	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group: "+groupName, http.StatusNotFound)
		return
	}

	data, err := group.Get(key)

	body, err := proto.Marshal(&Response{Value: data.ByteSlice()})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(body) //返回复制的数组元素
}

// Set 根据传入的主机地址注册,避免读写同时进行，
func (p *HTTPPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock() //避免读写并发操作
	p.peers = consistenthash.New(defaultReplicas, nil)
	p.httpGetters = make(map[string]*httpGetter)

	p.peers.Add(peers...) //注册对应的主机地址
	for _, v := range peers {
		p.httpGetters[v] = &httpGetter{baseURL: v + p.basePath}
	}
}

//PickPeer 寻找远程服务器
func (p *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock() //避免读写并发操作

	peer := p.peers.Get(key)
	//如果没有set就直接发起get请求将会造成获取到“”空值
	if peer != "" && peer != p.self {
		p.Log("Pick peer %s", peer)
		return p.httpGetters[peer], true
	}
	return nil, false
}
