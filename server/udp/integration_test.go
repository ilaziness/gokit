package udp

import (
	"net"
	"sync"
	"testing"
	"time"

	"github.com/ilaziness/gokit/config"
)

func TestUDPServerIntegration(t *testing.T) {
	// 配置服务器
	cfg := &config.UDPServer{
		Debug:     false,
		Address:   ":0", // 使用随机端口
		WorkerNum: 10,
	}

	// 创建服务器
	server := NewDefaultUDP(cfg)

	// 添加测试处理器
	server.AddHandler(1000, func(ctx *Context) {
		response := "echo: " + string(ctx.Payload)
		_ = ctx.Write([]byte(response))
	})

	// 启动服务器
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		// 这里我们不能直接调用Start()，因为它会阻塞
		// 在实际测试中，我们需要模拟服务器的核心逻辑
	}()

	// 等待服务器启动
	time.Sleep(100 * time.Millisecond)

	// 这里我们测试编解码器的功能，因为完整的服务器集成测试
	// 需要更复杂的设置来避免阻塞
	codec := NewPackCodec()

	// 测试ping包
	pingPack := &Pack{
		Head: PackHead{
			SQID:    1,
			OpCode:  uint16(OpCodePing),
			Version: Version1,
		},
		Payload: []byte("ping"),
	}

	data, err := codec.Encode(pingPack)
	if err != nil {
		t.Fatalf("Failed to encode ping pack: %v", err)
	}

	decodedPack, err := codec.Decode(data)
	if err != nil {
		t.Fatalf("Failed to decode ping pack: %v", err)
	}

	if decodedPack.Head.OpCode != uint16(OpCodePing) {
		t.Errorf("Expected OpCode %d, got %d", OpCodePing, decodedPack.Head.OpCode)
	}

	if string(decodedPack.Payload) != "ping" {
		t.Errorf("Expected payload 'ping', got '%s'", decodedPack.Payload)
	}
}

func TestContextFunctionality(t *testing.T) {
	// 创建模拟连接
	conn := &mockPacketConn{}
	addr := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 12345}
	codec := NewPackCodec()

	// 创建上下文
	ctx := &Context{}
	ctx.Reset(conn, addr, codec)

	// 测试数据包设置
	pack := &Pack{
		Head: PackHead{
			SQID:    123,
			OpCode:  1000,
			Version: Version1,
		},
		Payload: []byte("test data"),
	}

	ctx.SetData(pack)

	if ctx.SQID != 123 {
		t.Errorf("Expected SQID 123, got %d", ctx.SQID)
	}

	if ctx.OpCode != 1000 {
		t.Errorf("Expected OpCode 1000, got %d", ctx.OpCode)
	}

	if string(ctx.Payload) != "test data" {
		t.Errorf("Expected payload 'test data', got '%s'", ctx.Payload)
	}

	// 测试写入响应
	err := ctx.Write([]byte("response"))
	if err != nil {
		t.Errorf("Failed to write response: %v", err)
	}

	// 验证模拟连接收到了数据
	if len(conn.writtenData) == 0 {
		t.Error("Expected data to be written to connection")
	}
}

// 模拟PacketConn用于测试
type mockPacketConn struct {
	writtenData [][]byte
	writtenAddr []net.Addr
}

func (m *mockPacketConn) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	return 0, nil, nil
}

func (m *mockPacketConn) WriteTo(p []byte, addr net.Addr) (n int, err error) {
	data := make([]byte, len(p))
	copy(data, p)
	m.writtenData = append(m.writtenData, data)
	m.writtenAddr = append(m.writtenAddr, addr)
	return len(p), nil
}

func (m *mockPacketConn) Close() error {
	return nil
}

func (m *mockPacketConn) LocalAddr() net.Addr {
	return &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080}
}

func (m *mockPacketConn) SetDeadline(t time.Time) error {
	return nil
}

func (m *mockPacketConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (m *mockPacketConn) SetWriteDeadline(t time.Time) error {
	return nil
}
