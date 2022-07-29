package codec

import (
	"bufio"
	"encoding/gob"
	"io"
	"log"
)

// GobCodec 实现gob编码
type GobCodec struct {
	conn io.ReadWriteCloser //接受的链接
	buf  *bufio.Writer      //提升性能
	dec  *gob.Decoder
	enc  *gob.Encoder
}

func NewGobCodec(closer io.ReadWriteCloser) Codec {
	var buf = bufio.NewWriter(closer)
	return &GobCodec{
		conn: closer,
		buf:  buf,                    //缓冲避免阻塞
		dec:  gob.NewDecoder(closer), //解码器
		enc:  gob.NewEncoder(buf),    //编码器
	}

}

func (g *GobCodec) Close() error {
	return g.conn.Close()
}

func (g *GobCodec) ReadHeader(header *Header) error {
	return g.dec.Decode(header)
}

func (g *GobCodec) ReadBody(i interface{}) error {
	return g.dec.Decode(i)
}

func (g *GobCodec) Write(header *Header, body interface{}) (err error) {
	//写入数据
	defer func() {
		_ = g.buf.Flush()
		if err != nil {
			g.Close()
		}
	}()
	if err = g.enc.Encode(header); err != nil {
		log.Println("rpc codec: gob error encoding header:", err)
		return
	}
	if err = g.enc.Encode(body); err != nil {
		log.Println("rpc codec: gob error encoding body:", err)
		return
	}
	return nil
}
