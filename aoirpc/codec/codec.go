package codec

import "io"

// Header 消息头信息
type Header struct {
	ServiceMethod string
	Seq           uint64
	Error         string
}

// Codec 抽象接口，提供数据的编解码操作
type Codec interface {
	io.Closer
	ReadHeader(*Header) error
	ReadBody(interface{}) error
	Write(*Header, interface{}) error
}

//返回构造函数

type NewCodecFunc func(closer io.ReadWriteCloser) Codec

type Type string

const (
	GobType  Type = "application/gob"
	JsonType Type = "application/json"
)

// NewCodecFuncType 注册器
var NewCodecFuncType map[Type]NewCodecFunc

func init() {
	NewCodecFuncType = make(map[Type]NewCodecFunc)
	NewCodecFuncType[GobType] = NewGobCodec
}
