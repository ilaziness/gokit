package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"time"

	quicserver "github.com/ilaziness/gokit/server/quic"
	"github.com/quic-go/quic-go"
)

func main() {
	// TLS配置 - 在生产环境中应该验证证书
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true, // 仅用于测试
		NextProtos:         []string{"quic-server"},
	}

	// QUIC配置
	quicConfig := &quic.Config{
		MaxIdleTimeout:  30 * time.Second,
		KeepAlivePeriod: 15 * time.Second,
		EnableDatagrams: true,
		Allow0RTT:       false,
	}

	// 连接到QUIC服务器
	conn, err := quic.DialAddr(context.Background(), "localhost:8443", tlsConfig, quicConfig)
	if err != nil {
		log.Fatal("Failed to connect:", err)
	}
	defer conn.CloseWithError(0, "client shutdown")

	fmt.Println("Connected to QUIC server")

	// 测试数据报通信
	go testDatagrams(conn)

	// 测试流通信
	go testStreams(conn)

	// 保持连接
	time.Sleep(10 * time.Second)
	fmt.Println("Client shutting down")
}

func testDatagrams(conn quic.Connection) {
	codec := quicserver.NewPackCodec()

	for i := 0; i < 5; i++ {
		// 创建请求包
		pack := &quicserver.Pack{
			Head: quicserver.PackHead{
				SQID:    uint32(i + 1),
				OpCode:  1000, // 对应服务器的数据报处理器
				Version: 1,
			},
			Payload: []byte(fmt.Sprintf("Datagram message %d", i+1)),
		}

		// 编码并发送
		data, err := codec.Encode(pack)
		if err != nil {
			log.Printf("Encode error: %v", err)
			continue
		}

		if err := conn.SendDatagram(data); err != nil {
			log.Printf("Send datagram error: %v", err)
			continue
		}

		fmt.Printf("Sent datagram %d\n", i+1)

		// 接收响应
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			respData, err := conn.ReceiveDatagram(ctx)
			if err != nil {
				log.Printf("Receive datagram error: %v", err)
				return
			}

			respPack, err := codec.Decode(respData)
			if err != nil {
				log.Printf("Decode response error: %v", err)
				return
			}

			fmt.Printf("Received datagram response: %s\n", string(respPack.Payload))
		}()

		time.Sleep(1 * time.Second)
	}
}

func testStreams(conn quic.Connection) {
	codec := quicserver.NewPackCodec()

	for i := 0; i < 3; i++ {
		// 打开新流
		stream, err := conn.OpenStreamSync(context.Background())
		if err != nil {
			log.Printf("Open stream error: %v", err)
			continue
		}

		go func(streamID int, s quic.Stream) {
			defer s.Close()

			// 创建请求包
			pack := &quicserver.Pack{
				Head: quicserver.PackHead{
					SQID:    uint32(streamID + 100),
					OpCode:  2000, // 对应服务器的流处理器
					Version: 1,
				},
				Payload: []byte(fmt.Sprintf("Stream message %d", streamID)),
			}

			// 编码并发送
			data, err := codec.Encode(pack)
			if err != nil {
				log.Printf("Encode stream error: %v", err)
				return
			}

			if _, err := s.Write(data); err != nil {
				log.Printf("Write stream error: %v", err)
				return
			}

			fmt.Printf("Sent stream message %d\n", streamID)

			// 读取响应
			buffer := make([]byte, 1024)
			n, err := s.Read(buffer)
			if err != nil {
				// 服务器可能已经关闭了流，这是预期的行为
				if err.Error() == "EOF" {
					log.Println("Stream closed by server")
					return
				}
				log.Printf("Read stream error: %v", err)
				return
			}

			respPack, err := codec.Decode(buffer[:n])
			if err != nil {
				log.Printf("Decode stream response error: %v", err)
				return
			}

			fmt.Printf("Received stream response: %s\n", string(respPack.Payload))
		}(i+1, stream)

		time.Sleep(1 * time.Second)
	}
}
