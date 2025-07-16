package udp

import (
	"encoding/binary"
	"fmt"
)

type Codec interface {
	Decode([]byte) (*Pack, error)
	Encode(*Pack) ([]byte, error)
}

var (
	ErrPayloadLenErr = fmt.Errorf("payload length error")
	ErrPackTooSmall  = fmt.Errorf("packet too small")
)

const (
	packHeadLen        = 12    // 包头长度
	Version1    uint16 = 1     // 协议版本v1
	MaxUDPSize         = 65507 // UDP最大数据包大小

	OpCodeResOK     OpCode = 0 // 请求成功
	OpCodeServerErr OpCode = 1 // 服务端错误
	OpCodePing      OpCode = 2 // ping
	OpCodePong      OpCode = 3 // pong
	OpCodeNotFound  OpCode = 4 // 请求handler未找到
)

// Pack 包结构
type Pack struct {
	Head    PackHead
	Payload []byte
}

// PackHead 包头，固定长度packHeadLen
type PackHead struct {
	Len     uint32 // 包长度
	SQID    uint32 // 请求序号，客户端自增，用来标识一对请求和响应
	OpCode  uint16 // 操作码
	Version uint16 // 协议版本
}

// PackCodec 包编码解码器
type PackCodec struct{}

func NewPackCodec() *PackCodec {
	return &PackCodec{}
}

// Decode 解码包
func (p *PackCodec) Decode(data []byte) (*Pack, error) {
	if len(data) < packHeadLen {
		return nil, ErrPackTooSmall
	}

	// 解析包头
	pl := binary.BigEndian.Uint32(data[0:4])
	sqid := binary.BigEndian.Uint32(data[4:8])
	opCode := binary.BigEndian.Uint16(data[8:10])
	version := binary.BigEndian.Uint16(data[10:12])

	// 验证包长度
	if int(pl) != len(data) {
		return nil, fmt.Errorf("packet length mismatch: expected %d, got %d", pl, len(data))
	}

	// 计算 payload 长度
	payloadLen := int(pl) - packHeadLen
	if payloadLen < 0 {
		return nil, ErrPayloadLenErr
	}

	// 提取 payload 数据
	var payload []byte
	if payloadLen > 0 {
		payload = make([]byte, payloadLen)
		copy(payload, data[packHeadLen:])
	}

	// 构造 Pack 对象并返回
	pk := &Pack{
		Head: PackHead{
			Len:     pl,
			SQID:    sqid,
			OpCode:  opCode,
			Version: version,
		},
		Payload: payload,
	}
	return pk, nil
}

// Encode 编码包
func (p *PackCodec) Encode(pack *Pack) ([]byte, error) {
	pack.Head.Len = uint32(len(pack.Payload) + packHeadLen)
	if pack.Head.OpCode == 0 {
		pack.Head.OpCode = uint16(OpCodeResOK)
	}

	// 检查包大小是否超过UDP限制
	if pack.Head.Len > MaxUDPSize {
		return nil, fmt.Errorf("packet size %d exceeds UDP limit %d", pack.Head.Len, MaxUDPSize)
	}

	// 创建缓冲区
	buf := make([]byte, pack.Head.Len)

	// 编码包头
	binary.BigEndian.PutUint32(buf[0:4], pack.Head.Len)
	binary.BigEndian.PutUint32(buf[4:8], pack.Head.SQID)
	binary.BigEndian.PutUint16(buf[8:10], pack.Head.OpCode)
	binary.BigEndian.PutUint16(buf[10:12], pack.Head.Version)

	// 复制 payload
	if len(pack.Payload) > 0 {
		copy(buf[packHeadLen:], pack.Payload)
	}

	return buf, nil
}
