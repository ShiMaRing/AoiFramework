package aoirpc

import (
	"AoiFramework/aoirpc/codec"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"time"
)

const (
	MagicNumber      = 0x3bef5c //magic number marks a aoirpc request
	connected        = "200 Connected to Aoi RPC"
	defaultRPCPath   = "/aoirpc"
	defaultDebugPath = "/debug/aoirpc"
)

type Option struct {
	MagicNumber       int        //标识请求
	CodecType         codec.Type //请求类型
	ConnectionTimeout time.Duration
	HandleTimeout     time.Duration
}

var DefaultOption = &Option{
	MagicNumber:       MagicNumber,
	CodecType:         codec.GobType,
	ConnectionTimeout: 5 * time.Second,
	HandleTimeout:     5 * time.Second,
}

type Server struct {
	serviceMap sync.Map
}

func NewServer() *Server {
	return &Server{}
}

var DefaultServer = NewServer()

func (server *Server) Accept(lis net.Listener) {
	//规定监听接口
	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Println("rpc server: accept error:", err)
			return
		}
		go server.ServeConn(conn)
	}

}

// Accept accepts connections on the listener and serves requests
// for each incoming connection.
func Accept(lis net.Listener) {
	DefaultServer.Accept(lis)
}

var handleTimeout time.Duration

func (server *Server) ServeConn(conn net.Conn) {
	defer func() {
		_ = conn.Close()
	}()
	//使用json解析器解析
	var opt Option
	if err := json.NewDecoder(conn).Decode(&opt); err != nil {
		log.Println("rpc server: options error: ", err)
		return
	}
	handleTimeout = opt.HandleTimeout
	if opt.MagicNumber != MagicNumber {
		log.Printf("rpc server: invalid magic number %x", opt.MagicNumber)
		return
	}
	if opt.CodecType != codec.JsonType && opt.CodecType != codec.GobType {
		log.Printf("rpc server: invalid codec type %s", opt.CodecType)
		return
	}
	server.serveCode((codec.NewCodecFuncType[opt.CodecType])(conn))
}

//获得解码器
func (server *Server) serveCode(c codec.Codec) { //设置锁，避免重复
	var lock = new(sync.Mutex)
	var wg = new(sync.WaitGroup)
	for true {
		req, err := server.readRequest(c)
		if err != nil {
			if req == nil { //说明请求无法恢复
				break
			}
			req.h.Error = err.Error() //提示eof
			server.sendResponse(c, req.h, invalidRequest, lock)
			continue
		}
		wg.Add(1)
		go server.handleReq(c, req, lock, wg, handleTimeout)
	}
}

var invalidRequest = struct{}{}

type request struct {
	h            *codec.Header // header of request
	argv, replyv reflect.Value // argv and replyv of request

	mtype *methodType //指示方法
	svc   *service    //服务
}

func (server *Server) readReqHeader(c codec.Codec) (*codec.Header, error) {
	var header codec.Header
	err := c.ReadHeader(&header)
	if err != nil {
		if err != io.EOF && err != io.ErrUnexpectedEOF {
			//说明不是读完了,是其他错误
			log.Println("rpc server: read header error:", err)
			return nil, err
		}
	}
	return &header, nil
}

func (server *Server) readRequest(cc codec.Codec) (*request, error) {
	//读取完整请求
	header, err := server.readReqHeader(cc)
	if err != nil {
		return nil, err
	}
	req := &request{
		h: header,
	}
	//header中保存有相关的信息
	findService, mtype, err := server.findService(req.h.ServiceMethod)
	req.svc = findService
	req.mtype = mtype
	if err != nil {
		return req, err
	}
	req.argv = mtype.newArgv()
	req.replyv = mtype.newReplyv()

	argvi := req.argv.Interface()

	if req.argv.Type().Kind() != reflect.Pointer {
		argvi = req.argv.Addr().Interface()
	}

	if err = cc.ReadBody(argvi); err != nil {
		log.Println("rpc server: read argv err:", err)
		return req, err
	}
	return req, nil
}

func (server *Server) sendResponse(c codec.Codec, h *codec.Header, r interface{}, lock *sync.Mutex) {
	lock.Lock()
	defer lock.Unlock()
	if err := c.Write(h, r); err != nil {
		log.Println("rpc server: write response error:", err)
	}
}

func (server *Server) handleReq(c codec.Codec, req *request, lock *sync.Mutex, wg *sync.WaitGroup, handleTimeout time.Duration) {
	defer wg.Done()
	//定义called和sent chan
	var called = make(chan struct{})
	var sent = make(chan struct{})

	var finish = make(chan struct{})
	defer close(finish)

	go func() {
		err := req.svc.call(req.mtype, req.argv, req.replyv)
		select {
		case <-finish:
			close(called)
			close(sent)
			return
		case called <- struct{}{}:
			if err != nil {
				req.h.Error = err.Error()
				server.sendResponse(c, req.h, invalidRequest, lock)
				sent <- struct{}{}
				return
			}
			server.sendResponse(c, req.h, req.replyv.Interface(), lock)
			sent <- struct{}{}
		}
	}()

	if handleTimeout == 0 {
		<-called
		<-sent
		return
	}

	select {
	case <-time.After(handleTimeout):
		req.h.Error = fmt.Sprintf("rpc server: request handle timeout: expect within %s", handleTimeout)
		server.sendResponse(c, req.h, invalidRequest, lock)
	case <-called:
		<-sent
	}
}

func (server *Server) Register(rcvr interface{}) error {
	s := newService(rcvr)
	if _, dup := server.serviceMap.LoadOrStore(s.name, s); dup {
		return errors.New("rpc: service already defined: " + s.name)
	}
	return nil
}

func Register(rcvr interface{}) error {
	return DefaultServer.Register(rcvr)
}

func (server *Server) findService(serviceMethod string) (svc *service, mtype *methodType, err error) {
	//根据传入的A.B()获取函数
	index := strings.LastIndex(serviceMethod, ".")
	if index == -1 {
		err = errors.New("rpc server: service/method request ill-formed: " + serviceMethod)
		return
	}
	serviceName, methodName := serviceMethod[:index], serviceMethod[index+1:]
	load, ok := server.serviceMap.Load(serviceName)
	if !ok {
		err = errors.New("rpc server: can't find service " + serviceName)
		return
	}
	svc = load.(*service)
	mtype = svc.methods[methodName]
	if mtype == nil {
		err = errors.New("rpc server: can't find method " + methodName)
	}

	return
}

//实现handler接口
func (server *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	//检测是否为connect方法请求
	if req.Method != http.MethodConnect {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusMethodNotAllowed)
		_, _ = io.WriteString(w, "405 must CONNECT\n")
		return
	}
	//劫持这个响应流，不再自动关闭，不再具有http方法
	conn, _, err := w.(http.Hijacker).Hijack()
	if err != nil {
		log.Print("rpc hijacking ", req.RemoteAddr, ": ", err.Error())
		return
	}
	_, _ = io.WriteString(conn, "HTTP/1.0 "+connected+"\n\n")
	server.ServeConn(conn)
}

func (server *Server) HandleHttp() {
	http.Handle(defaultDebugPath, server)
}

func HandleHTTP() {
	DefaultServer.HandleHTTP()
}
