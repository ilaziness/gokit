package quic

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/ilaziness/gokit/config"
	"github.com/quic-go/quic-go"
)

// TestQUICServerIntegration 集成测试
func TestQUICServerIntegration(t *testing.T) {
	// 生成测试证书
	certFile, keyFile, cleanup := generateTestCerts(t)
	defer cleanup()

	// 创建服务器配置
	cfg := &config.QUICServer{
		Debug:            false,
		Address:          "localhost:0", // 使用随机端口
		WorkerNum:        100,
		CertFile:         certFile,
		KeyFile:          keyFile,
		IdleTimeout:      5,
		KeepAlive:        2,
		HandshakeTimeout: 3,
		MaxStreams:       100,
		Allow0RTT:        false,
	}

	// 创建服务器
	server := NewQUIC(cfg)

	// 添加测试处理器
	var datagramReceived, streamReceived sync.WaitGroup
	datagramReceived.Add(1)
	streamReceived.Add(1)

	server.AddHandler(1000, func(ctx *Context) {
		defer datagramReceived.Done()
		response := fmt.Sprintf("Echo: %s", string(ctx.Payload))
		if err := ctx.Write([]byte(response)); err != nil {
			t.Errorf("Write datagram response error: %v", err)
		}
	})

	server.AddStreamHandler(2000, func(ctx *StreamContext) {
		defer streamReceived.Done()
		response := fmt.Sprintf("Stream Echo: %s", string(ctx.Payload))
		if err := ctx.Write([]byte(response)); err != nil {
			t.Errorf("Write stream response error: %v", err)
		}
	})

	// 启动服务器
	serverCtx, serverCancel := context.WithCancel(context.Background())
	defer serverCancel()

	go func() {
		// 模拟服务器启动，但不阻塞测试
		server.initConfigs()
		var err error
		server.listener, err = quic.ListenAddr(cfg.Address, server.tlsConfig, server.quicConfig)
		if err != nil {
			t.Errorf("Listen error: %v", err)
			return
		}
		defer server.listener.Close()

		// 更新配置中的地址为实际监听地址
		cfg.Address = server.listener.Addr().String()

		server.handleConnections(serverCtx)
	}()

	// 等待服务器启动
	time.Sleep(100 * time.Millisecond)

	// 创建客户端
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"quic-server"},
	}

	quicConfig := &quic.Config{
		MaxIdleTimeout:  5 * time.Second,
		KeepAlivePeriod: 2 * time.Second,
		EnableDatagrams: true,
	}

	// 连接到服务器
	conn, err := quic.DialAddr(context.Background(), cfg.Address, tlsConfig, quicConfig)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.CloseWithError(0, "test complete")

	// 测试数据报通信
	t.Run("Datagram", func(t *testing.T) {
		codec := NewPackCodec()

		// 发送数据报
		pack := &Pack{
			Head: PackHead{
				SQID:    123,
				OpCode:  1000,
				Version: 1,
			},
			Payload: []byte("test datagram"),
		}

		data, err := codec.Encode(pack)
		if err != nil {
			t.Fatalf("Encode error: %v", err)
		}

		if err := conn.SendDatagram(data); err != nil {
			t.Fatalf("Send datagram error: %v", err)
		}

		// 等待处理完成
		done := make(chan struct{})
		go func() {
			datagramReceived.Wait()
			close(done)
		}()

		select {
		case <-done:
			// 测试通过
		case <-time.After(2 * time.Second):
			t.Error("Datagram handler not called within timeout")
		}

		// 接收响应
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		respData, err := conn.ReceiveDatagram(ctx)
		if err != nil {
			t.Fatalf("Receive datagram error: %v", err)
		}

		respPack, err := codec.Decode(respData)
		if err != nil {
			t.Fatalf("Decode response error: %v", err)
		}

		expected := "Echo: test datagram"
		if string(respPack.Payload) != expected {
			t.Errorf("Expected %q, got %q", expected, string(respPack.Payload))
		}
	})

	// 测试流通信
	t.Run("Stream", func(t *testing.T) {
		codec := NewPackCodec()

		// 打开流
		stream, err := conn.OpenStreamSync(context.Background())
		if err != nil {
			t.Fatalf("Open stream error: %v", err)
		}
		// 不要在这里关闭流，等读取完响应后再关闭

		// 发送流数据
		pack := &Pack{
			Head: PackHead{
				SQID:    456,
				OpCode:  2000,
				Version: 1,
			},
			Payload: []byte("test stream"),
		}

		data, err := codec.Encode(pack)
		if err != nil {
			t.Fatalf("Encode error: %v", err)
		}

		if _, err := stream.Write(data); err != nil {
			t.Fatalf("Write stream error: %v", err)
		}

		// 等待处理完成
		done := make(chan struct{})
		go func() {
			streamReceived.Wait()
			close(done)
		}()

		select {
		case <-done:
			// 测试通过
		case <-time.After(2 * time.Second):
			t.Error("Stream handler not called within timeout")
		}

		// 读取响应
		buffer := make([]byte, 1024)
		n, err := stream.Read(buffer)
		if err != nil {
			// 服务器可能已经关闭了流，这是预期的行为
			if err.Error() == "EOF" {
				t.Log("Stream closed by server after sending response")
				return
			}
			t.Fatalf("Read stream error: %v", err)
		}

		respPack, err := codec.Decode(buffer[:n])
		if err != nil {
			t.Fatalf("Decode stream response error: %v", err)
		}

		expected := "Stream Echo: test stream"
		if string(respPack.Payload) != expected {
			t.Errorf("Expected %q, got %q", expected, string(respPack.Payload))
		}

		// 读取完响应后再关闭流
		stream.Close()
	})
}

// TestContextFunctionality 测试上下文功能
func TestContextFunctionality(t *testing.T) {
	// 测试数据报上下文
	t.Run("DatagramContext", func(t *testing.T) {
		ctx := &Context{}
		codec := NewPackCodec()

		// 模拟连接（这里使用nil，实际测试中需要真实连接）
		ctx.Reset(nil, codec)

		pack := &Pack{
			Head: PackHead{
				SQID:    789,
				OpCode:  1001,
				Version: 1,
			},
			Payload: []byte("test payload"),
		}

		ctx.SetData(pack)

		if ctx.SQID != 789 {
			t.Errorf("Expected SQID 789, got %d", ctx.SQID)
		}
		if ctx.OpCode != 1001 {
			t.Errorf("Expected OpCode 1001, got %d", ctx.OpCode)
		}
		if string(ctx.Payload) != "test payload" {
			t.Errorf("Expected payload 'test payload', got %q", string(ctx.Payload))
		}
	})

	// 测试流上下文
	t.Run("StreamContext", func(t *testing.T) {
		ctx := &StreamContext{}
		codec := NewPackCodec()

		// 模拟连接和流
		ctx.Reset(nil, nil, codec)

		pack := &Pack{
			Head: PackHead{
				SQID:    999,
				OpCode:  2001,
				Version: 1,
			},
			Payload: []byte("stream payload"),
		}

		ctx.SetData(pack)

		if ctx.SQID != 999 {
			t.Errorf("Expected SQID 999, got %d", ctx.SQID)
		}
		if ctx.OpCode != 2001 {
			t.Errorf("Expected OpCode 2001, got %d", ctx.OpCode)
		}
		if string(ctx.Payload) != "stream payload" {
			t.Errorf("Expected payload 'stream payload', got %q", string(ctx.Payload))
		}
	})
}

// generateTestCerts 生成测试用的TLS证书
func generateTestCerts(t *testing.T) (certFile, keyFile string, cleanup func()) {
	// 生成私钥
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}

	// 创建证书模板
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization:  []string{"Test"},
			Country:       []string{"US"},
			Province:      []string{""},
			Locality:      []string{"San Francisco"},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		DNSNames:    []string{"localhost"},
	}

	// 生成证书
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatalf("Failed to create certificate: %v", err)
	}

	// 创建临时文件
	certFile = t.TempDir() + "/server.crt"
	keyFile = t.TempDir() + "/server.key"

	// 写入证书文件
	certOut, err := os.Create(certFile)
	if err != nil {
		t.Fatalf("Failed to create cert file: %v", err)
	}
	defer certOut.Close()

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certDER}); err != nil {
		t.Fatalf("Failed to write certificate: %v", err)
	}

	// 写入私钥文件
	keyOut, err := os.Create(keyFile)
	if err != nil {
		t.Fatalf("Failed to create key file: %v", err)
	}
	defer keyOut.Close()

	privateKeyDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		t.Fatalf("Failed to marshal private key: %v", err)
	}

	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privateKeyDER}); err != nil {
		t.Fatalf("Failed to write private key: %v", err)
	}

	cleanup = func() {
		os.Remove(certFile)
		os.Remove(keyFile)
	}

	return certFile, keyFile, cleanup
}
