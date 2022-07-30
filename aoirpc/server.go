package aoirpc

import (
	"AoiFramework/aoirpc/codec"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"reflect"
	"strings"
	"sync"
)

const MagicNumber = 0x3bef5c //magic number marks a aoirpc request

type Option struct {
	MagicNumber int        //标识请求
	CodecType   codec.Type //请求类型
}

var DefaultOption = &Option{
	MagicNumber: MagicNumber,
	CodecType:   codec.GobType,
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
		go server.handleReq(c, req, lock, wg)
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

func (server *Server) handleReq(c codec.Codec, req *request, lock *sync.Mutex, wg *sync.WaitGroup) {
	defer wg.Done()

	err := req.svc.call(req.mtype, req.argv, req.replyv)

	if err != nil {
		req.h.Error = err.Error()
		server.sendResponse(c, req.h, invalidRequest, lock)
		return
	}

	fmt.Printf("get rpc %v %v \n", req.argv, *req.replyv.Interface().(*int))

	server.sendResponse(c, req.h, req.replyv.Interface(), lock)
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
