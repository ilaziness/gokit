# QUIC 服务器

为 gokit 框架提供的生产级 QUIC 服务器实现，采用现代最佳实践和全面的功能构建。

## 特性

- **完整的 QUIC 支持**：基于 `quic-go` 构建，支持 QUIC 协议并兼容 HTTP/3
- **双通信模式**：
  - 不可靠数据报用于低延迟通信
  - 可靠流用于保证传输
- **生产就绪**：
  - 连接池和工作线程管理
  - 全面的错误处理和恢复机制
  - 可配置的超时和限制
- **安全性**：
  - 要求 TLS 1.3（QUIC 标准）
  - 可配置的密码套件和曲线
  - 基于证书的认证
- **中间件支持**：
  - 内置 ping/pong、日志记录、速率限制和恢复中间件
  - 易于扩展自定义中间件
- **性能优化**：
  - 上下文对象池
  - 高效的二进制协议
  - 经过基准测试和验证

## 快速开始

### 1. 生成证书

QUIC 需要 TLS 证书。生成测试证书：

```bash
# Linux/macOS
chmod +x scripts/generate_certs.sh
./scripts/generate_certs.sh

# Windows
scripts\generate_certs.bat
```

### 2. 基础服务器

```go
package main

import (
    "github.com/ilaziness/gokit/config"
    "github.com/ilaziness/gokit/server/quic"
)

func main() {
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

    server := quic.NewDefaultQUIC(cfg)

    // 添加数据报处理器
    server.AddHandler(1000, func(ctx *quic.Context) {
        response := "Echo: " + string(ctx.Payload)
        ctx.Write([]byte(response))
    })

    // 添加流处理器
    server.AddStreamHandler(2000, func(ctx *quic.StreamContext) {
        response := "Stream Echo: " + string(ctx.Payload)
        ctx.Write([]byte(response))
    })

    server.Start()
}
```

### 3. 增强版服务器

增强版服务器提供了更多企业级功能，如服务注册发现、可观测性、多协议支持和高级中间件：

```go
package main

import (
    "fmt"
    "log"
    "time"
    
    "github.com/ilaziness/gokit/config"
    "github.com/ilaziness/gokit/server/quic"
    "go.opentelemetry.io/otel/metric/noop"
    "go.opentelemetry.io/otel/trace/noop"
)

// 简单的速率限制器实现
type SimpleRateLimiter struct {
    limits map[string]int
    max    int
}

func NewSimpleRateLimiter(max int) *SimpleRateLimiter {
    return &SimpleRateLimiter{
        limits: make(map[string]int),
        max:    max,
    }
}

func (r *SimpleRateLimiter) Allow(key string) bool {
    if r.limits[key] >= r.max {
        return false
    }
    r.limits[key]++
    return true
}

func (r *SimpleRateLimiter) Reset(key string) {
    r.limits[key] = 0
}

// 简单的认证器实现
type SimpleAuthenticator struct{}

func (a *SimpleAuthenticator) Authenticate(token string) (*quic.User, error) {
    // 在实际应用中，这里应该验证令牌并从数据库或缓存中获取用户
    if token == "valid-token" {
        return &quic.User{
            ID:       "user-123",
            Username: "testuser",
            Roles:    []string{"user", "admin"},
        }, nil
    }
    return nil, fmt.Errorf("invalid token")
}

func main() {
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

    // 创建增强版服务器
    server := quic.NewEnhancedServer(cfg)
    
    // 创建指标收集器（实际应用中应使用真实的指标收集器）
    metrics := &quic.ServerMetrics{
        RequestCount:      noop.Int64Counter{},
        RequestDuration:   noop.Float64Histogram{},
        ActiveConnections: noop.Int64UpDownCounter{},
        ErrorCount:        noop.Int64Counter{},
        BytesReceived:     noop.Int64Counter{},
        BytesSent:         noop.Int64Counter{},
    }
    
    // 设置可观测性组件
    server.WithMetrics(metrics)
    server.WithTracer(noop.Tracer{})
    
    // 添加增强中间件
    rateLimiter := NewSimpleRateLimiter(100) // 每个客户端最多100个请求
    authenticator := &SimpleAuthenticator{}
    
    server.AddEnhancedMiddleware(quic.NewTracingMiddleware(noop.Tracer{}))
    server.AddEnhancedMiddleware(quic.NewMetricsMiddleware(metrics))
    server.AddEnhancedMiddleware(quic.NewRateLimitMiddleware(rateLimiter))
    server.AddEnhancedMiddleware(quic.NewAuthMiddleware(authenticator))
    
    // 添加基础中间件
    server.AddMiddleware(quic.Logger())
    server.AddMiddleware(quic.Recovery())
    
    // 添加数据报处理器
    server.AddHandler(1000, func(ctx *quic.Context) {
        log.Printf("处理来自 %s 的请求", ctx.GetRemoteAddr())
        
        // 在实际应用中，这里可以使用 JSON 编解码器处理结构化数据
        response := fmt.Sprintf("增强响应: %s", string(ctx.Payload))
        ctx.Write([]byte(response))
    })
    
    fmt.Println("启动增强版 QUIC 服务器，监听地址:", cfg.Address)
    server.Start()
}
```

### 4. 运行示例

```bash
# 启动服务器
cd example
go run main.go

# 在另一个终端中运行客户端
cd client_example
go run main.go
```

## 配置

### QUICServer 配置

```go
type QUICServer struct {
    Debug            bool   // 启用调试日志
    Address          string // 服务器地址（例如 "localhost:8443"）
    WorkerNum        int    // 最大并发工作线程数
    CertFile         string // TLS 证书文件路径
    KeyFile          string // TLS 私钥文件路径
    IdleTimeout      int    // 连接空闲超时（秒）
    KeepAlive        int    // 保活间隔（秒）
    HandshakeTimeout int    // TLS 握手超时（秒）
    MaxStreams       int    // 每个连接的最大并发流数
    Allow0RTT        bool   // 启用 0-RTT（早期数据）
}
```

### 默认值

- `WorkerNum`: 100,000
- `IdleTimeout`: 30 秒
- `KeepAlive`: 15 秒
- `HandshakeTimeout`: 10 秒
- `MaxStreams`: 1000
- `Allow0RTT`: false

## 协议

服务器使用自定义二进制协议，包结构如下：

```
+--------+--------+--------+--------+
|           长度 (4 字节)           |
+--------+--------+--------+--------+
|           序列号 (4 字节)         |
+--------+--------+--------+--------+
|  操作码 (2 字节) | 版本号 (2 字节) |
+--------+--------+--------+--------+
|              负载数据             |
|             （可变长度）          |
+--------+--------+--------+--------+
```

### 保留操作码

- `0`: 成功响应
- `1`: 服务器错误
- `2`: Ping
- `3`: Pong
- `4`: 未找到

业务逻辑应使用操作码 >= 1000。

## 通信模式

### 数据报（不可靠）

适用于：
- 低延迟通信
- 即发即弃消息
- 实时更新
- 游戏、物联网传感器

```go
server.AddHandler(1000, func(ctx *quic.Context) {
    // 处理数据报
    ctx.Write([]byte("response"))
})
```

### 流（可靠）

适用于：
- 文件传输
- 需要保证传输的 API 调用
- 大数据传输
- 请求-响应模式

```go
server.AddStreamHandler(2000, func(ctx *quic.StreamContext) {
    // 处理流
    streamID := ctx.GetStreamID()
    ctx.Write([]byte("response"))
})
```

## 中间件

### 内置中间件

```go
server := quic.NewDefaultQUIC(cfg) // 包含 Ping 中间件

// 添加额外中间件
server.AddMiddleware(quic.Logger())
server.AddMiddleware(quic.Recovery())
server.AddMiddleware(quic.RateLimiter(100, 60)) // 100 请求/分钟
```

### 自定义中间件

```go
func CustomAuth() quic.Handler {
    return func(ctx *quic.Context) {
        // 认证逻辑
        if !isAuthenticated(ctx) {
            ctx.ServerErr()
            ctx.Abort()
            return
        }
        ctx.Next()
    }
}

server.AddMiddleware(CustomAuth())
```

## 性能

### 基准测试

运行基准测试以测量性能：

```bash
go test -bench=. -benchmem
```

在现代硬件上的典型结果：
- 编码：~500ns/op，1 次内存分配
- 解码：~400ns/op，2 次内存分配
- 往返：~900ns/op，3 次内存分配

### 优化建议

1. **连接池**：尽可能重用连接
2. **批量操作**：在流中发送多条消息
3. **负载大小**：保持数据报大小在 1200 字节以下以获得最佳性能
4. **工作线程调整**：根据工作负载调整 `WorkerNum`
5. **缓冲区大小**：根据预期数据量配置接收窗口

## 安全考虑

### TLS 配置

服务器强制使用 TLS 1.3 并使用安全的密码套件：

```go
CipherSuites: []uint16{
    tls.TLS_AES_128_GCM_SHA256,
    tls.TLS_AES_256_GCM_SHA384,
    tls.TLS_CHACHA20_POLY1305_SHA256,
}
```

### 证书管理

- 在生产环境中使用来自受信任 CA 的证书
- 实现证书轮换
- 监控证书过期
- 考虑使用 Let's Encrypt 进行自动续期

### 网络安全

- 使用防火墙限制访问
- 实现速率限制
- 监控 DDoS 攻击
- 考虑使用支持 QUIC 的负载均衡器

## 监控和可观测性

### 日志记录

服务器提供不同级别的结构化日志：

```go
// 启用调试日志
cfg.Debug = true

// 处理器中的自定义日志
log.Info(ctx, "Processing request from %s", ctx.GetRemoteAddr())
```

### 指标

与您的监控系统集成：

```go
// Prometheus 示例
func MetricsMiddleware() quic.Handler {
    return func(ctx *quic.Context) {
        start := time.Now()
        ctx.Next()
        duration := time.Since(start)
        
        requestDuration.WithLabelValues(
            fmt.Sprintf("%d", ctx.OpCode),
        ).Observe(duration.Seconds())
    }
}
```

## 生产部署

### Docker 示例

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o quic-server ./example

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/quic-server .
COPY --from=builder /app/server.crt .
COPY --from=builder /app/server.key .
EXPOSE 8443/udp
CMD ["./quic-server"]
```

### Kubernetes 部署

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: quic-server
spec:
  replicas: 3
  selector:
    matchLabels:
      app: quic-server
  template:
    metadata:
      labels:
        app: quic-server
    spec:
      containers:
      - name: quic-server
        image: your-registry/quic-server:latest
        ports:
        - containerPort: 8443
          protocol: UDP
        env:
        - name: QUIC_ADDRESS
          value: "0.0.0.0:8443"
        - name: QUIC_CERT_FILE
          value: "/etc/certs/tls.crt"
        - name: QUIC_KEY_FILE
          value: "/etc/certs/tls.key"
        volumeMounts:
        - name: certs
          mountPath: /etc/certs
          readOnly: true
      volumes:
      - name: certs
        secret:
          secretName: quic-server-certs
---
apiVersion: v1
kind: Service
metadata:
  name: quic-server-service
spec:
  selector:
    app: quic-server
  ports:
  - port: 8443
    targetPort: 8443
    protocol: UDP
  type: LoadBalancer
```

## 故障排除

### 常见问题

1. **证书错误**
   - 确保证书有效且未过期
   - 检查文件权限
   - 验证证书链

2. **连接超时**
   - 检查防火墙设置
   - 验证 UDP 端口是否开放
   - 调整超时配置

3. **性能问题**
   - 监控工作线程池使用情况
   - 检查内存泄漏
   - 分析 CPU 使用情况

### 调试模式

启用调试日志以获取详细信息：

```go
cfg.Debug = true
```

这将记录：
- 连接事件
- 数据包详情
- 错误消息
- 性能指标

## 贡献

1. Fork 仓库
2. 创建功能分支
3. 为新功能添加测试
4. 确保所有测试通过
5. 提交拉取请求

### 运行测试

```bash
# 单元测试
go test ./...

# 基准测试
go test -bench=. -benchmem

# 竞态检测
go test -race ./...

# 覆盖率
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 许可证

本项目使用与父项目 gokit 相同的许可证。