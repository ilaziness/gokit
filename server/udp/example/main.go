package main

import (
	"fmt"
	"log"

	"github.com/ilaziness/gokit/config"
	"github.com/ilaziness/gokit/server/udp"
)

func main() {
	// UDP服务器配置
	cfg := &config.UDPServer{
		Debug:     true,
		Address:   ":8080",
		WorkerNum: 1000,
		// 可选的TLS配置
		// CertFile: "server.crt",
		// KeyFile:  "server.key",
	}

	// 创建UDP服务器
	server := udp.NewDefaultUDP(cfg)

	// 添加自定义处理器
	server.AddHandler(1000, func(ctx *udp.Context) {
		fmt.Printf("Received from %s: %s\n", ctx.GetRemoteAddr(), string(ctx.Payload))

		// 回复消息
		response := fmt.Sprintf("Echo: %s", string(ctx.Payload))
		if err := ctx.Write([]byte(response)); err != nil {
			log.Printf("Failed to send response: %v", err)
		}
	})

	// 添加自定义中间件
	server.AddMiddleware(func(ctx *udp.Context) {
		fmt.Printf("Middleware: Processing OpCode %d from %s\n", ctx.OpCode, ctx.GetRemoteAddr())
		ctx.Next()
	})

	// 启动服务器
	fmt.Println("Starting UDP server on :8080")
	server.Start()
}
