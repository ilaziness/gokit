package quic

import (
	"context"

	"github.com/quic-go/quic-go"
)

// Context 数据报上下文
type Context struct {
	context.Context

	index     int
	isAbort   bool
	handler   []Handler
	packCodec Codec
	conn      quic.Connection
	Pack      *Pack
	SQID      uint32
	OpCode    OpCode
	Payload   []byte
}

// Reset 重置上下文
func (c *Context) Reset(conn quic.Connection, pc Codec) {
	c.index = -1
	c.isAbort = false
	c.handler = nil
	c.conn = conn
	c.packCodec = pc
}

// Next 运行中间件
func (c *Context) Next() {
	c.index++
	l := len(c.handler)
	for ; c.index < l; c.index++ {
		if c.isAbort {
			// 中断执行
			return
		}
		c.handler[c.index](c)
	}
}

// SetData 设置数据
func (c *Context) SetData(pack *Pack) {
	c.SQID = pack.Head.SQID
	c.OpCode = OpCode(pack.Head.OpCode)
	c.Payload = pack.Payload
	c.Pack = pack
}

// Write 写入一般响应数据
func (c *Context) Write(data []byte) error {
	pack := &Pack{
		Head: PackHead{
			SQID:    c.SQID,
			Version: c.Pack.Head.Version,
		},
		Payload: data,
	}
	return c.sendDatagram(pack)
}

// WriteWithOpCode 写入指定操作码响应数据
func (c *Context) WriteWithOpCode(opcode OpCode, data []byte) error {
	pack := &Pack{
		Head: PackHead{
			OpCode:  uint16(opcode),
			SQID:    c.SQID,
			Version: c.Pack.Head.Version,
		},
		Payload: data,
	}
	return c.sendDatagram(pack)
}

// ServerErr 写入服务错误响应
func (c *Context) ServerErr() error {
	return c.WriteWithOpCode(OpCodeServerErr, nil)
}

// WriteNotFound 写入404错误响应
func (c *Context) WriteNotFound() error {
	return c.WriteWithOpCode(OpCodeNotFound, nil)
}

// Abort 中断处理
func (c *Context) Abort() {
	c.isAbort = true
}

// GetRemoteAddr 获取客户端地址
func (c *Context) GetRemoteAddr() string {
	return c.conn.RemoteAddr().String()
}

// GetConnectionState 获取连接状态
func (c *Context) GetConnectionState() quic.ConnectionState {
	return c.conn.ConnectionState()
}

// sendDatagram 发送数据报
func (c *Context) sendDatagram(pack *Pack) error {
	data, err := c.packCodec.Encode(pack)
	if err != nil {
		return err
	}
	// 使用 quic-go 的 SendDatagram 方法替代 SendMessage
	return c.conn.SendDatagram(data)
}

// StreamContext 流上下文
type StreamContext struct {
	context.Context

	packCodec Codec
	conn      quic.Connection
	stream    quic.Stream
	Pack      *Pack
	SQID      uint32
	OpCode    OpCode
	Payload   []byte
}

// Reset 重置流上下文
func (sc *StreamContext) Reset(conn quic.Connection, stream quic.Stream, pc Codec) {
	sc.conn = conn
	sc.stream = stream
	sc.packCodec = pc
}

// SetData 设置数据
func (sc *StreamContext) SetData(pack *Pack) {
	sc.SQID = pack.Head.SQID
	sc.OpCode = OpCode(pack.Head.OpCode)
	sc.Payload = pack.Payload
	sc.Pack = pack
}

// Write 写入一般响应数据
func (sc *StreamContext) Write(data []byte) error {
	pack := &Pack{
		Head: PackHead{
			SQID:    sc.SQID,
			Version: sc.Pack.Head.Version,
		},
		Payload: data,
	}
	return sc.sendStream(pack)
}

// WriteWithOpCode 写入指定操作码响应数据
func (sc *StreamContext) WriteWithOpCode(opcode OpCode, data []byte) error {
	pack := &Pack{
		Head: PackHead{
			OpCode:  uint16(opcode),
			SQID:    sc.SQID,
			Version: sc.Pack.Head.Version,
		},
		Payload: data,
	}
	return sc.sendStream(pack)
}

// ServerErr 写入服务错误响应
func (sc *StreamContext) ServerErr() error {
	return sc.WriteWithOpCode(OpCodeServerErr, nil)
}

// WriteNotFound 写入404错误响应
func (sc *StreamContext) WriteNotFound() error {
	return sc.WriteWithOpCode(OpCodeNotFound, nil)
}

// GetRemoteAddr 获取客户端地址
func (sc *StreamContext) GetRemoteAddr() string {
	return sc.conn.RemoteAddr().String()
}

// GetStreamID 获取流ID
func (sc *StreamContext) GetStreamID() quic.StreamID {
	return sc.stream.StreamID()
}

// GetConnectionState 获取连接状态
func (sc *StreamContext) GetConnectionState() quic.ConnectionState {
	return sc.conn.ConnectionState()
}

// sendStream 通过流发送数据
func (sc *StreamContext) sendStream(pack *Pack) error {
	data, err := sc.packCodec.Encode(pack)
	if err != nil {
		return err
	}
	_, err = sc.stream.Write(data)
	return err
}
