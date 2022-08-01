package xclient

import (
	"AoiFramework/aoirpc"
	"context"
	"reflect"
	"sync"
)

// XClient 客户端负载均衡
type XClient struct {
	d       Discovery
	mode    SelectMode
	opt     *aoirpc.Option
	mu      sync.Mutex
	clients map[string]*aoirpc.Client //缓存
}

func (xc *XClient) Close() error {
	xc.mu.Lock()
	defer xc.mu.Unlock()
	for k, v := range xc.clients {
		_ = v.Close()
		delete(xc.clients, k)
	}
	return nil
}

func NewXClient(d Discovery, mode SelectMode, opt *aoirpc.Option) *XClient {
	return &XClient{d: d, mode: mode, opt: opt, clients: make(map[string]*aoirpc.Client)}
}

func (xc *XClient) Call(ctx context.Context, serviceMethod string, args, reply interface{}) error {
	addr, err := xc.d.Get(xc.mode)
	if err != nil {
		return err
	}
	return xc.call(addr, ctx, serviceMethod, args, reply)
}

func (xc *XClient) call(addr string, ctx context.Context, method string, args interface{}, reply interface{}) error {
	//尝试获取client
	var client *aoirpc.Client
	var err error
	client, err = xc.dial(addr)
	if err != nil {
		return err
	}
	return client.Call(ctx, method, args, reply)
}

func (xc *XClient) dial(addr string) (*aoirpc.Client, error) {
	xc.mu.Lock()
	defer xc.mu.Unlock()
	client, ok := xc.clients[addr]
	if ok && !client.IsAvailable() {
		_ = client.Close()
		delete(xc.clients, addr)
		client = nil
	}
	//找不到对应的client
	if client == nil {
		var err error
		client, err = aoirpc.XDial(addr, xc.opt)
		if err != nil {
			return nil, err
		}
		xc.clients[addr] = client
	}
	return client, nil
}

// Broadcast 广播,所有的server都要执行
func (xc *XClient) Broadcast(ctx context.Context, serviceMethod string, args, reply interface{}) error {
	all, err := xc.d.GetAll()
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	var mu sync.Mutex
	var e error
	replyDone := reply == nil //标识是否需要设置reply
	ctx, cancelFunc := context.WithCancel(ctx)
	for _, server := range all {
		wg.Add(1)
		go func(addr string) {
			defer wg.Done()
			var cloneReply interface{}
			//选择是否需要设置
			if !replyDone {
				cloneReply = reflect.New(reflect.ValueOf(reply).Elem().Type()).Interface()
			}
			err := xc.Call(ctx, serviceMethod, args, cloneReply)
			mu.Lock()
			defer mu.Unlock()
			if err != nil && e != nil {
				e = err
				cancelFunc()
			}
			if err == nil && !replyDone {
				reflect.ValueOf(reply).Elem().Set(reflect.ValueOf(cloneReply).Elem())
				replyDone = true
			}

		}(server)
	}
	wg.Wait()
	return e
}
