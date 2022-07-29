package aoirpc

import (
	"AoiFramework/aoirpc/codec"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"reflect"
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

type Server struct{}

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
	//暂时认为是string
	req.argv = reflect.New(reflect.TypeOf(""))
	if err = cc.ReadBody(req.argv.Interface()); err != nil {
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
	log.Println(req.h, req.argv.Elem()) //获取实例对象
	//简单回复字符串
	req.replyv = reflect.ValueOf(fmt.Sprintf("aoiServer resp %d", req.h.Seq))
	server.sendResponse(c, req.h, req.replyv.Interface(), lock)
}
