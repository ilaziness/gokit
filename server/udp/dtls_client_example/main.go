package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/ilaziness/gokit/server/udp"
	"github.com/pion/dtls/v3"
)

func main() {
	// 解析服务器地址
	addr, err := net.ResolveUDPAddr("udp", "localhost:8080")
	if err != nil {
		panic(err)
	}

	// DTLS客户端配置
	config := &dtls.Config{
		InsecureSkipVerify: true, // 仅用于测试，生产环境应该验证证书
	}

	// 连接到DTLS服务器
	conn, err := dtls.Dial("udp", addr, config)
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to DTLS server: %v", err))
	}
	defer conn.Close()

	fmt.Println("Connected to DTLS UDP server")

	codec := udp.NewPackCodec()
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("DTLS Echo Client")
	fmt.Println("Type messages to send to server (type 'quit' to exit):")

	sqid := uint32(1)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}
		if input == "quit" {
			break
		}

		// 创建数据包
		pack := &udp.Pack{
			Head: udp.PackHead{
				SQID:    sqid,
				OpCode:  1000, // 使用自定义操作码
				Version: udp.Version1,
			},
			Payload: []byte(input),
		}

		// 编码并发送
		data, err := codec.Encode(pack)
		if err != nil {
			fmt.Printf("Encode error: %v\n", err)
			continue
		}

		_, err = conn.Write(data)
		if err != nil {
			fmt.Printf("Send error: %v\n", err)
			continue
		}

		// 读取响应
		buffer := make([]byte, 1024)
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Printf("Read error: %v\n", err)
			continue
		}

		// 解码响应
		responsePack, err := codec.Decode(buffer[:n])
		if err != nil {
			fmt.Printf("Decode error: %v\n", err)
			continue
		}

		fmt.Printf("Server: %s\n", string(responsePack.Payload))
		sqid++
	}

	fmt.Println("Goodbye!")
}
