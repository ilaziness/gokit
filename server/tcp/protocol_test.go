package tcp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPackCodec_Decode(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte // 模拟输入数据
		expected *Pack  // 期望输出
		err      error  // 期望错误
	}{
		{
			name: "Normal Case",
			input: func() []byte {
				var buf bytes.Buffer
				var data = []byte(`{"key": "value"}`)
				_ = binary.Write(&buf, binary.BigEndian, uint32(16+packHeadLen)) // Len (4 bytes)
				_ = binary.Write(&buf, binary.BigEndian, uint32(123))            // SQID (4 bytes)
				_ = binary.Write(&buf, binary.BigEndian, uint16(0))              // OpCode (2 bytes)
				_ = binary.Write(&buf, binary.BigEndian, uint16(1))              // Version (2 bytes)
				buf.Write(data)
				return buf.Bytes()
			}(),
			expected: &Pack{
				Head: PackHead{
					Len:     16 + packHeadLen,
					OpCode:  0,
					SQID:    123,
					Version: Version1,
				},
				Payload: []byte(`{"key": "value"}`),
			},
			err: nil,
		},
		{
			name: "Empty Payload",
			input: func() []byte {
				var buf bytes.Buffer
				_ = binary.Write(&buf, binary.BigEndian, uint32(packHeadLen)) // Len (4 bytes)
				_ = binary.Write(&buf, binary.BigEndian, uint32(123))         // SQID (4 bytes)
				_ = binary.Write(&buf, binary.BigEndian, uint16(0))           // OpCode (2 bytes)
				_ = binary.Write(&buf, binary.BigEndian, uint16(1))           // Version (2 bytes)
				return buf.Bytes()
			}(),
			expected: &Pack{
				Head: PackHead{
					Len:    packHeadLen,
					OpCode: 0,
					SQID:   123,
				},
				Payload: []byte{},
			},
			err: nil,
		},
		{
			name:     "Invalid Length - Incomplete Header",
			input:    make([]byte, 9), // 包头不完整（需要10字节）
			expected: nil,
			err:      io.ErrUnexpectedEOF,
		},
		{
			name: "Negative Payload Length",
			input: func() []byte {
				var buf bytes.Buffer
				_ = binary.Write(&buf, binary.BigEndian, uint32(5))   // Len (4 bytes)
				_ = binary.Write(&buf, binary.BigEndian, uint32(123)) // SQID (4 bytes)
				_ = binary.Write(&buf, binary.BigEndian, uint16(0))   // OpCode (2 bytes)
				_ = binary.Write(&buf, binary.BigEndian, uint16(1))   // Version (2 bytes)
				return buf.Bytes()
			}(),
			expected: nil,
			err:      io.ErrShortBuffer,
		},
		{
			name:     "Nil Input",
			input:    nil,
			expected: nil,
			err:      io.EOF,
		},
		{
			name: "Large Payload",
			input: func() []byte {
				var buf bytes.Buffer
				_ = binary.Write(&buf, binary.BigEndian, uint32(1024*1024+packHeadLen)) // 1MB payload + header
				_ = binary.Write(&buf, binary.BigEndian, uint32(123))                   // SQID (4 bytes)
				_ = binary.Write(&buf, binary.BigEndian, uint16(0))                     // OpCode (2 bytes)
				_ = binary.Write(&buf, binary.BigEndian, uint16(1))                     // Version (2 bytes)
				// 填充1MB的数据
				payload := make([]byte, 1024*1024)
				for i := range payload {
					payload[i] = byte(i % 256)
				}
				buf.Write(payload)
				return buf.Bytes()
			}(),
			expected: &Pack{
				Head: PackHead{
					Len:     1024*1024 + packHeadLen,
					OpCode:  0,
					SQID:    123,
					Version: Version1,
				},
				Payload: func() []byte {
					payload := make([]byte, 1024*1024)
					for i := range payload {
						payload[i] = byte(i % 256)
					}
					return payload
				}(),
			},
			err: nil,
		},
		{
			name: "Partial Read",
			input: func() []byte {
				var buf bytes.Buffer
				_ = binary.Write(&buf, binary.BigEndian, uint32(26)) // Len (4 bytes)，故意比实际长度小
				return buf.Bytes()
			}(),
			expected: nil,
			err:      io.EOF,
		},
		{
			name:     "Zero Length Buffer",
			input:    make([]byte, 0),
			expected: nil,
			err:      io.EOF,
		},
		{
			name:     "Buffer Too Small for Header",
			input:    make([]byte, 5), // 小于包头最小长度(10 bytes)
			expected: nil,
			err:      io.ErrUnexpectedEOF,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn := bytes.NewBuffer(tt.input)
			codec := NewPackCodec()
			pack, err := codec.Decode(conn)
			if tt.err != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.err, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected.Head.Len, pack.Head.Len)
				require.Equal(t, tt.expected.Head.OpCode, pack.Head.OpCode)
				require.Equal(t, tt.expected.Head.SQID, pack.Head.SQID)
				require.Equal(t, tt.expected.Payload, pack.Payload)
			}
		})
	}
}

func TestPackCodec_Encode(t *testing.T) {
	tests := []struct {
		name     string
		sqid     uint32
		data     []byte // 修改字段类型为 []byte 以匹配 Encode 接口定义
		expected *Pack  // 期望输出的完整数据（包头 + Payload）
		err      error
	}{
		{
			name: "Normal Case",
			sqid: 123,
			data: []byte(`{"key": "value"}`),
			expected: &Pack{
				Head: PackHead{
					Len:     16 + packHeadLen,
					OpCode:  0,
					SQID:    123,
					Version: Version1,
				},
				Payload: []byte(`{"key": "value"}`),
			},
			err: nil,
		},
		{
			name:     "Write Error Simulation",
			sqid:     123,
			data:     []byte("test data"),
			expected: nil,
			err:      io.ErrClosedPipe,
		},
		{
			name: "Nil Data",
			sqid: 123,
			data: nil,
			expected: &Pack{
				Head: PackHead{
					Len:     packHeadLen, // 仅包头
					OpCode:  0,
					SQID:    123,
					Version: Version1,
				},
				Payload: []byte{}, // Payload 为 nil
			},
			err: nil,
		},
		{
			name: "Empty Data",
			sqid: 123,
			data: []byte{},
			expected: &Pack{
				Head: PackHead{
					Len:     packHeadLen, // 仅包头
					OpCode:  0,
					SQID:    123,
					Version: Version1,
				},
				Payload: []byte{}, // Payload 为空字节切片
			},
			err: nil,
		},
		{
			name: "Large Payload",
			sqid: 123,
			data: func() []byte {
				// 创建一个1MB大小的有效载荷
				largePayload := make([]byte, 1024*1024)
				for i := range largePayload {
					largePayload[i] = byte(i % 256)
				}
				return largePayload
			}(),
			expected: &Pack{
				Head: PackHead{
					Len:     1024*1024 + packHeadLen, // Len = payload length + header length
					OpCode:  0,
					SQID:    123,
					Version: Version1,
				},
				Payload: func() []byte {
					// 创建一个1MB大小的有效载荷
					largePayload := make([]byte, 1024*1024)
					for i := range largePayload {
						largePayload[i] = byte(i % 256)
					}
					return largePayload
				}(),
			},
			err: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var conn io.ReadWriter
			if errors.Is(tt.err, io.ErrClosedPipe) { // 模拟写入错误
				conn = &errorWriter{err: tt.err}
			} else {
				conn = new(bytes.Buffer)
			}

			codec := NewPackCodec()
			pack := &Pack{
				Head: PackHead{
					OpCode: 0,
					SQID:   tt.sqid,
				},
				Payload: tt.data,
			}
			// 执行 Encode 方法
			err := codec.Encode(conn, pack)

			// 验证错误
			if tt.err != nil {
				assert.Error(t, err)
				// 使用 ErrorIs 精确匹配错误链中的原始错误
				assert.ErrorIs(t, err, tt.err)
			} else {
				assert.NoError(t, err)

				// 使用Decode解析生成的数据包
				buf, ok := conn.(*bytes.Buffer)
				if !ok {
					t.Fatalf("expected *bytes.Buffer, got %T", conn)
				}

				// 通过Decode方法解析生成的数据包
				pack, err := codec.Decode(buf)
				assert.NoError(t, err)
				// 验证解析后的数据与预期一致
				require.Equal(t, tt.expected.Head.Len, pack.Head.Len)
				require.Equal(t, tt.expected.Head.OpCode, pack.Head.OpCode)
				require.Equal(t, tt.expected.Head.SQID, pack.Head.SQID)
				require.Equal(t, tt.expected.Payload, pack.Payload)
			}
		})
	}
}

// errorWriter 是一个模拟写入错误的自定义 Writer 实现
type errorWriter struct {
	err error
}

func (w *errorWriter) Write(_ []byte) (n int, err error) {
	return 0, w.err
}

func (w *errorWriter) Read(_ []byte) (n int, err error) {
	return 0, io.EOF
}
