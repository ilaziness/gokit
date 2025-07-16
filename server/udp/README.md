# UDP Server

一个功能完善的UDP服务器实现，支持可选的TLS加密（DTLS）。

## 特性

- 🚀 高性能UDP服务器
- 🔒 可选TLS加密支持（DTLS）
- 🔧 中间件支持
- 📦 自定义协议编解码
- 🎯 操作码路由
- 🔄 连接池和工作协程池
- 📊 内置ping/pong支持
- 🛡️ 优雅关闭

## 快速开始

### 基本使用

```go
package main

import (
    "github.com/ilaziness/gokit/config"
    "github.com/ilaziness/gokit/server/udp"
)

func main() {
    // 配置UDP服务器
    cfg := &config.UDPServer{
        Debug:     true,
        Address:   ":8080",
        WorkerNum: 1000,
    }

    // 创建服务器
    server := udp.NewDefaultUDP(cfg)

    // 添加处理器
    server.AddHandler(1000, func(ctx *udp.Context) {
        // 处理业务逻辑
        ctx.Write([]byte("Hello from UDP server"))
    })

    // 启动服务器
    server.Start()
}
```

### 启用DTLS加密

```go
cfg := &config.UDPServer{
    Debug:    true,
    Address:  ":8080",
    CertFile: "server.crt",  // DTLS证书文件
    KeyFile:  "server.key",  // DTLS私钥文件
}

// 创建DTLS服务器
server := udp.NewDefaultUDP(cfg)
server.Start() // 自动检测证书文件并启用DTLS
```

### 生成测试证书

```bash
# Linux/macOS
./scripts/generate_dtls_certs.sh

# Windows
scripts\generate_dtls_certs.bat
```

## 协议格式

UDP服务器使用自定义的二进制协议：

```
+--------+--------+--------+--------+
|      Len (4 bytes)               |  包总长度
+--------+--------+--------+--------+
|      SQID (4 bytes)              |  请求序号
+--------+--------+--------+--------+
| OpCode | Version|                |  操作码(2字节) + 版本(2字节)
+--------+--------+--------+--------+
|      Payload (variable)          |  数据载荷
+--------+--------+--------+--------+
```

### 保留操作码

- `0`: 请求成功响应
- `1`: 服务器错误
- `2`: Ping
- `3`: Pong
- `4`: 未找到处理器

业务操作码建议从1000开始使用。

## 中间件

### 内置中间件

- **Ping中间件**: 自动处理ping/pong消息

### 自定义中间件

```go
server.AddMiddleware(func(ctx *udp.Context) {
    // 前置处理
    log.Printf("Processing request from %s", ctx.GetRemoteAddr())
    
    // 继续执行下一个中间件或处理器
    ctx.Next()
    
    // 后置处理
    log.Printf("Request processed")
})
```

## 处理器

```go
server.AddHandler(1001, func(ctx *udp.Context) {
    // 获取客户端地址
    clientAddr := ctx.GetRemoteAddr()
    
    // 获取请求数据
    payload := ctx.Payload
    
    // 发送响应
    ctx.Write([]byte("response data"))
    
    // 或者发送带操作码的响应
    ctx.WriteWithOpCode(1002, []byte("custom response"))
    
    // 发送错误响应
    ctx.ServerErr()
    
    // 中断处理链
    ctx.Abort()
})
```

## 配置选项

```go
type UDPServer struct {
    Debug     bool   // 调试模式
    Address   string // 监听地址，如 ":8080"
    WorkerNum int    // 工作协程数量
    CertFile  string // TLS证书文件路径
    KeyFile   string // TLS私钥文件路径
}
```

## 客户端示例

### 普通UDP客户端

```go
package main

import (
    "net"
    "github.com/ilaziness/gokit/server/udp"
)

func main() {
    conn, _ := net.Dial("udp", "localhost:8080")
    defer conn.Close()

    codec := udp.NewPackCodec()
    
    // 构造请求包
    pack := &udp.Pack{
        Head: udp.PackHead{
            SQID:    1,
            OpCode:  1000,
            Version: udp.Version1,
        },
        Payload: []byte("Hello Server"),
    }
    
    // 编码并发送
    data, _ := codec.Encode(pack)
    conn.Write(data)
    
    // 读取响应
    buffer := make([]byte, 1024)
    n, _ := conn.Read(buffer)
    
    // 解码响应
    response, _ := codec.Decode(buffer[:n])
    fmt.Printf("Response: %s\n", response.Payload)
}
```

### DTLS客户端

```go
package main

import (
    "context"
    "time"
    "github.com/pion/dtls/v3"
    "github.com/ilaziness/gokit/server/udp"
)

func main() {
    // DTLS客户端配置
    config := &dtls.Config{
        InsecureSkipVerify: true, // 仅用于测试
        ConnectContextMaker: func() (context.Context, func()) {
            return context.WithTimeout(context.Background(), 30*time.Second)
        },
    }

    // 连接到DTLS服务器
    conn, _ := dtls.Dial("udp", "localhost:8080", config)
    defer conn.Close()

    codec := udp.NewPackCodec()
    
    // 构造请求包
    pack := &udp.Pack{
        Head: udp.PackHead{
            SQID:    1,
            OpCode:  1000,
            Version: udp.Version1,
        },
        Payload: []byte("Hello DTLS Server"),
    }
    
    // 编码并发送
    data, _ := codec.Encode(pack)
    conn.Write(data)
    
    // 读取响应
    buffer := make([]byte, 1024)
    n, _ := conn.Read(buffer)
    
    // 解码响应
    response, _ := codec.Decode(buffer[:n])
    fmt.Printf("DTLS Response: %s\n", response.Payload)
}
```

## 性能特性

- **连接池**: 复用Context对象减少GC压力
- **工作协程池**: 控制并发处理数量
- **零拷贝**: 高效的数据包处理
- **异步处理**: 非阻塞消息处理

## 安全特性

- **TLS支持**: 可选的DTLS加密
- **包大小限制**: 防止过大数据包攻击
- **超时控制**: 防止资源耗尽
- **优雅关闭**: 安全的服务器关闭

## 注意事项

1. UDP是无连接协议，不保证消息送达
2. 单个UDP包最大65507字节
3. TLS over UDP (DTLS) 需要特殊处理
4. 建议在生产环境中启用TLS加密
5. 合理设置WorkerNum以平衡性能和资源使用

## 示例代码

查看 `example/` 目录获取完整的服务器和客户端示例代码。