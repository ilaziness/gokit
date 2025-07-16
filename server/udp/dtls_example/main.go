package main

import (
	"fmt"
	"log"

	"github.com/ilaziness/gokit/config"
	"github.com/ilaziness/gokit/server/udp"
)

func main() {
	// DTLS UDP服务器配置
	cfg := &config.UDPServer{
		Debug:     true,
		Address:   ":8080",
		WorkerNum: 1000,
		// DTLS配置 - 需要提供证书文件
		CertFile: "server.crt",
		KeyFile:  "server.key",
	}

	// 创建DTLS UDP服务器
	server := udp.NewDefaultUDP(cfg)

	// 添加测试处理器
	server.AddHandler(1000, func(ctx *udp.Context) {
		fmt.Printf("Received message from %s: %s\n",
			ctx.GetRemoteAddr(), string(ctx.Payload))

		// 回复消息
		response := fmt.Sprintf("Echo: %s", string(ctx.Payload))
		err := ctx.Write([]byte(response))
		if err != nil {
			log.Printf("Failed to send response: %v", err)
		}
	})

	// 启动服务器
	fmt.Println("Starting DTLS UDP server on :8080")
	fmt.Println("Make sure you have server.crt and server.key files")
	server.Start()
}
