package tcp

import (
	"encoding/binary"
	"fmt"
	"io"
)

type Codec interface {
	Decode(io.ReadWriter) (*Pack, error)
	Encode(io.ReadWriter, *Pack) error
}

var (
	ErrPayloadLenErr = fmt.Errorf("read payload length error")
)

const (
	OpCodeResOK     OpCode = 0 // 请求成功
	OpCodeServerErr OpCode = 1 // 服务端错误
	OpCodePing      OpCode = 2 // ping
	OpCodePong      OpCode = 3 // ping
	OpCodeNotFund   OpCode = 4 // 请求handler未找到
)

// Pack 包结构
type Pack struct {
	Head    PackHead
	Payload []byte
}

// PackHead 包头，固定长度10byte
type PackHead struct {
	Len    uint32 // 包长度
	OpCode uint16 // 操作码
	SQID   uint32 // 请求序号，客户端自增，用来标识一对请求和响应
}

// PackCodec 包编码解码器
type PackCodec struct{}

func NewPackCodec() *PackCodec {
	return &PackCodec{}
}

// Decode 解码包
func (p *PackCodec) Decode(conn io.ReadWriter) (*Pack, error) {
	var (
		pl     uint32
		opCode uint16
		sqid   uint32
	)

	// 读取 Len 字段 (4 bytes)
	if err := binary.Read(conn, binary.BigEndian, &pl); err != nil {
		return nil, err
	}

	// 读取 OpCode 字段 (2 bytes)
	if err := binary.Read(conn, binary.BigEndian, &opCode); err != nil {
		return nil, err
	}

	// 读取 SQID 字段 (4 bytes)
	if err := binary.Read(conn, binary.BigEndian, &sqid); err != nil {
		return nil, err
	}

	// 计算 payload 长度：总长度 - 包头长度 (10 bytes)
	payloadLen := int(pl) - 10
	if payloadLen < 0 {
		return nil, io.ErrShortBuffer
	}

	// 读取 payload 数据
	payload := make([]byte, payloadLen)
	n, err := io.ReadFull(conn, payload)
	if err != nil {
		return nil, err
	}
	// 增加对实际读取长度的校验
	if n != payloadLen {
		return nil, ErrPayloadLenErr
	}

	// 构造 Pack 对象并返回
	pk := &Pack{
		Head: PackHead{
			Len:    pl,
			OpCode: opCode,
			SQID:   sqid,
		},
		Payload: payload,
	}
	return pk, nil
}

// Encode 编码包
// pack len会重新计算覆盖
func (p *PackCodec) Encode(conn io.ReadWriter, pack *Pack) error {
	pack.Head.Len = uint32(len(pack.Payload) + 10)
	if pack.Head.OpCode == 0 {
		pack.Head.OpCode = uint16(OpCodeResOK)
	}

	// 编码包头
	headerBuf := make([]byte, 10)
	binary.BigEndian.PutUint32(headerBuf[0:4], pack.Head.Len)
	binary.BigEndian.PutUint16(headerBuf[4:6], pack.Head.OpCode)
	binary.BigEndian.PutUint32(headerBuf[6:10], pack.Head.SQID)

	// 发送包头和 Payload
	if _, err := conn.Write(headerBuf); err != nil {
		return err
	}
	if len(pack.Payload) == 0 {
		return nil
	}
	if _, err := conn.Write(pack.Payload); err != nil {
		return err
	}

	return nil
}
