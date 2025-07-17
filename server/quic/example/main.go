package main

import (
	"fmt"
	"log"

	"github.com/ilaziness/gokit/config"
	"github.com/ilaziness/gokit/server/quic"
)

func main() {
	// QUIC服务器配置
	cfg := &config.QUICServer{
		Debug:            true,
		Address:          "localhost:8443",
		WorkerNum:        1000,
		CertFile:         "server.crt",
		KeyFile:          "server.key",
		IdleTimeout:      30,
		KeepAlive:        15,
		HandshakeTimeout: 10,
		MaxStreams:       1000,
		Allow0RTT:        false,
	}

	// 创建QUIC服务器
	server := quic.NewDefaultQUIC(cfg)

	// 添加数据报处理器
	server.AddHandler(1000, func(ctx *quic.Context) {
		fmt.Printf("Received datagram from %s: %s\n", ctx.GetRemoteAddr(), string(ctx.Payload))

		response := fmt.Sprintf("Echo: %s", string(ctx.Payload))
		if err := ctx.Write([]byte(response)); err != nil {
			log.Printf("Write response error: %v", err)
		}
	})

	// 添加流处理器
	server.AddStreamHandler(2000, func(ctx *quic.StreamContext) {
		fmt.Printf("Received stream data from %s (stream %d): %s\n",
			ctx.GetRemoteAddr(), ctx.GetStreamID(), string(ctx.Payload))

		response := fmt.Sprintf("Stream Echo: %s", string(ctx.Payload))
		if err := ctx.Write([]byte(response)); err != nil {
			log.Printf("Write stream response error: %v", err)
		}
	})

	// 添加自定义中间件
	server.AddMiddleware(quic.Logger())
	server.AddMiddleware(quic.Recovery())
	server.AddMiddleware(quic.RateLimiter(100, 60)) // 每分钟最多100个请求

	fmt.Println("Starting QUIC server on", cfg.Address)
	fmt.Println("Make sure you have server.crt and server.key files in the current directory")

	// 启动服务器
	server.Start()
}
