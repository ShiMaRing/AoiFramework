package aoirpc

import (
	"AoiFramework/aoirpc/codec"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
)

/*the method’s type is exported.
the method is exported.
the method has two arguments, both exported (or builtin) types.
the method’s second argument is a pointer.
the method has return type error.*/

// Call 定义请求结构体
type Call struct {
	Seq           uint64
	ServiceMethod string
	Args          interface{}
	Reply         interface{}
	Error         error
	Done          chan *Call // Strobes when call is complete.
}

//将自己添加至队列
func (call *Call) done() {
	call.Done <- call
}

// Client 定义客户端
type Client struct {
	cc       codec.Codec
	opt      *Option
	sending  sync.Mutex
	header   codec.Header
	mu       sync.Mutex
	seq      uint64
	pending  map[uint64]*Call
	closing  bool // user has called Close
	shutdown bool // server has told us to stop
}

var ErrShutdown = errors.New("connection is shut down")

func (c *Client) Close() error {
	//关闭操作
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closing {
		return ErrShutdown
	}
	c.closing = true
	return c.cc.Close()
}

func (c *Client) IsAvailable() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return !c.closing && !c.shutdown
}

//接下来提供请求call的注册删除等方法
func (c *Client) registerCall(call *Call) (uint64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	//查看是否合法，此时不可以调用判断函数会死锁
	if c.closing || c.shutdown {
		return 0, ErrShutdown
	}
	call.Seq = c.seq
	c.pending[c.seq] = call
	c.seq++
	return call.Seq, nil
}

//保护map并发
func (c *Client) removeCall(seq uint64) *Call {
	c.mu.Lock()
	defer c.mu.Unlock()
	call := c.pending[seq]
	delete(c.pending, seq)
	return call
}

//通知所有错误消息
func (c *Client) terminateCalls(err error) {
	c.sending.Lock()
	defer c.sending.Unlock()
	c.mu.Lock()
	defer c.mu.Unlock()
	//获取锁，
	c.shutdown = true
	for _, call := range c.pending {
		call.Error = err
		call.done()
	}
}

//接收参数
func (c *Client) receive() {
	var err error
	for err == nil {
		var h codec.Header
		//读取头信息就出错
		if err = c.cc.ReadHeader(&h); err != nil {
			break
		}
		//否则就认为正确收到了头返回值
		call := c.removeCall(h.Seq)

		/*		fmt.Println("sql:", h.Seq)*/
		switch {
		case call == nil:
			//这是一个错误的call传值
			err = c.cc.ReadBody(nil)
		case h.Error != "":
			call.Error = errors.New(h.Error)
			err = c.cc.ReadBody(nil)
			call.done()
		default:
			err = c.cc.ReadBody(call.Reply)
			if err != nil {
				call.Error = err
			}
			call.done()
		}
	}
	c.terminateCalls(err)
}

func NewClient(conn net.Conn, opt *Option) (*Client, error) {
	f := codec.NewCodecFuncType[opt.CodecType]
	if f == nil {
		err := fmt.Errorf("invalid codec type %s", opt.CodecType)
		log.Println("rpc client: codec error:", err)
		return nil, err
	}
	if err := json.NewEncoder(conn).Encode(opt); err != nil {
		log.Println("rpc client: options error: ", err)
		_ = conn.Close()
		return nil, err
	}
	return newClientCodec(f(conn), opt), nil

}

func newClientCodec(f codec.Codec, opt *Option) *Client {
	client := &Client{
		cc:      f,
		opt:     opt,
		sending: sync.Mutex{},
		mu:      sync.Mutex{},
		pending: make(map[uint64]*Call),
	}
	go client.receive()
	return client
}

func parseOptions(opts ...*Option) (*Option, error) {

	if len(opts) == 0 || opts[0] == nil {
		return DefaultOption, nil
	}

	if len(opts) != 1 {
		return nil, errors.New("number of options is more than 1")
	}

	opt := opts[0]
	opt.MagicNumber = DefaultOption.MagicNumber

	if opt.CodecType == "" {
		opt.CodecType = DefaultOption.CodecType
	}

	return opt, nil
}

// Dial 拨号方法
func Dial(network, address string, opts ...*Option) (client *Client, err error) {
	option, err := parseOptions(opts...)
	if err != nil {
		return nil, err
	}
	dial, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	defer func() {
		if client == nil {
			_ = dial.Close()
		}

	}()
	return NewClient(dial, option)
}

// Call 异步等待
func (c *Client) Call(serviceMethod string, args, reply interface{}) error {
	var call *Call
	call = <-c.Go(serviceMethod, args, reply, make(chan *Call, 1)).Done
	return call.Error
}

func (c *Client) Go(method string, args interface{}, reply interface{}, calls chan *Call) *Call {
	//用户也可以直接访问该方法
	if calls == nil {
		calls = make(chan *Call, 10)
	} else if cap(calls) == 0 {
		log.Panic("rpc client: done channel is unbuffered")
	}
	call := &Call{
		ServiceMethod: method,
		Args:          args,
		Reply:         reply,
		Done:          calls,
	}
	c.send(call)
	return call
}

//发送数据
func (c *Client) send(call *Call) {
	c.sending.Lock()
	defer c.sending.Unlock()

	registerCall, err := c.registerCall(call)

	if err != nil {
		call.Error = err
		call.done()
		return
	}

	c.header.Seq = registerCall
	c.header.ServiceMethod = call.ServiceMethod
	c.header.Error = ""

	if err := c.cc.Write(&c.header, call.Args); err != nil {
		call = c.removeCall(registerCall)
		if call != nil {
			call.Error = err
			call.done()
		}
	}
}
