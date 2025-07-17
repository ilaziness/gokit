package quic

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/ilaziness/gokit/config"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// EnhancedServer 增强版QUIC服务器，更符合主流架构
type EnhancedServer struct {
	*Server // 嵌入基础服务器

	// 服务发现
	registry ServiceRegistry

	// 可观测性
	tracer  trace.Tracer
	metrics *ServerMetrics

	// 协议支持
	codecs map[string]EnhancedCodec

	// 中间件增强
	middlewareChain MiddlewareChain
}

// ServiceRegistry 服务注册发现接口
type ServiceRegistry interface {
	Register(ctx context.Context, service *ServiceInfo) error
	Deregister(ctx context.Context, serviceID string) error
	Discover(ctx context.Context, serviceName string) ([]*ServiceInfo, error)
	Watch(ctx context.Context, serviceName string) (<-chan []*ServiceInfo, error)
}

// ServiceInfo 服务信息
type ServiceInfo struct {
	ID      string            `json:"id"`
	Name    string            `json:"name"`
	Address string            `json:"address"`
	Port    int               `json:"port"`
	Tags    []string          `json:"tags"`
	Meta    map[string]string `json:"meta"`
	Health  HealthCheck       `json:"health"`
	Weight  int               `json:"weight"`
}

// HealthCheck 健康检查配置
type HealthCheck struct {
	Enabled  bool          `json:"enabled"`
	Interval time.Duration `json:"interval"`
	Timeout  time.Duration `json:"timeout"`
	Path     string        `json:"path"`
}

// ServerMetrics 服务器指标
type ServerMetrics struct {
	RequestCount      metric.Int64Counter
	RequestDuration   metric.Float64Histogram
	ActiveConnections metric.Int64UpDownCounter
	ErrorCount        metric.Int64Counter
	BytesReceived     metric.Int64Counter
	BytesSent         metric.Int64Counter
}

// MiddlewareChain 中间件链
type MiddlewareChain struct {
	middlewares []EnhancedMiddleware
	mu          sync.RWMutex
}

// EnhancedMiddleware 增强中间件接口
type EnhancedMiddleware interface {
	Name() string
	Priority() int
	Handle(ctx *EnhancedContext) error
}

// EnhancedContext 增强上下文
type EnhancedContext struct {
	*Context

	// 链路追踪
	Span trace.Span

	// 请求ID
	RequestID string

	// 协议信息
	Protocol      string
	EnhancedCodec EnhancedCodec

	// 元数据
	Metadata map[string]string

	// 指标
	startTime time.Time
}

// EnhancedCodec 增强编解码器接口
type EnhancedCodec interface {
	Encode(v interface{}) ([]byte, error)
	Decode(data []byte, v interface{}) error
	ContentType() string
}

// MultiCodec 多协议编解码器
type MultiCodec struct {
	codecs map[string]EnhancedCodec
	mu     sync.RWMutex
}

// JSONCodec JSON编解码器
type JSONCodec struct{}

func (c *JSONCodec) Encode(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (c *JSONCodec) Decode(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func (c *JSONCodec) ContentType() string {
	return "application/json"
}

// ProtobufCodec Protobuf编解码器（示例）
type ProtobufCodec struct{}

func (c *ProtobufCodec) Encode(v interface{}) ([]byte, error) {
	// 实际实现需要protobuf支持
	return nil, fmt.Errorf("protobuf codec not implemented")
}

func (c *ProtobufCodec) Decode(data []byte, v interface{}) error {
	return fmt.Errorf("protobuf codec not implemented")
}

func (c *ProtobufCodec) ContentType() string {
	return "application/protobuf"
}

// NewEnhancedServer 创建增强版服务器
func NewEnhancedServer(config *config.QUICServer) *EnhancedServer {
	baseServer := NewQUIC(config)

	enhanced := &EnhancedServer{
		Server: baseServer,
		codecs: make(map[string]EnhancedCodec),
		middlewareChain: MiddlewareChain{
			middlewares: make([]EnhancedMiddleware, 0),
		},
	}

	// 注册默认编解码器
	enhanced.RegisterCodec("application/json", &JSONCodec{})
	enhanced.RegisterCodec("application/protobuf", &ProtobufCodec{})

	return enhanced
}

// RegisterCodec 注册编解码器
func (s *EnhancedServer) RegisterCodec(contentType string, codec EnhancedCodec) {
	s.codecs[contentType] = codec
}

// WithServiceRegistry 设置服务注册中心
func (s *EnhancedServer) WithServiceRegistry(registry ServiceRegistry) *EnhancedServer {
	s.registry = registry
	return s
}

// WithTracer 设置链路追踪
func (s *EnhancedServer) WithTracer(tracer trace.Tracer) *EnhancedServer {
	s.tracer = tracer
	return s
}

// WithMetrics 设置指标收集
func (s *EnhancedServer) WithMetrics(metrics *ServerMetrics) *EnhancedServer {
	s.metrics = metrics
	return s
}

// AddEnhancedMiddleware 添加增强中间件
func (s *EnhancedServer) AddEnhancedMiddleware(middleware EnhancedMiddleware) {
	s.middlewareChain.Add(middleware)
}

// Add 添加中间件到链
func (mc *MiddlewareChain) Add(middleware EnhancedMiddleware) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	// 按优先级插入
	inserted := false
	for i, m := range mc.middlewares {
		if middleware.Priority() > m.Priority() {
			mc.middlewares = append(mc.middlewares[:i],
				append([]EnhancedMiddleware{middleware}, mc.middlewares[i:]...)...)
			inserted = true
			break
		}
	}

	if !inserted {
		mc.middlewares = append(mc.middlewares, middleware)
	}
}

// Execute 执行中间件链
func (mc *MiddlewareChain) Execute(ctx *EnhancedContext) error {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	for _, middleware := range mc.middlewares {
		if err := middleware.Handle(ctx); err != nil {
			return fmt.Errorf("middleware %s failed: %w", middleware.Name(), err)
		}
	}
	return nil
}

// 内置增强中间件

// TracingMiddleware 链路追踪中间件
type TracingMiddleware struct {
	tracer trace.Tracer
}

func NewTracingMiddleware(tracer trace.Tracer) *TracingMiddleware {
	return &TracingMiddleware{tracer: tracer}
}

func (m *TracingMiddleware) Name() string {
	return "tracing"
}

func (m *TracingMiddleware) Priority() int {
	return 1000 // 高优先级，最先执行
}

func (m *TracingMiddleware) Handle(ctx *EnhancedContext) error {
	if m.tracer != nil {
		_, span := m.tracer.Start(ctx.Context, "quic.request")
		ctx.Span = span

		// 设置span属性
		// span.SetAttributes(
		// 	attribute.String("quic.opcode", fmt.Sprintf("%d", ctx.OpCode)),
		// 	attribute.String("quic.remote_addr", ctx.GetRemoteAddr()),
		// )
	}
	return nil
}

// MetricsMiddleware 指标收集中间件
type MetricsMiddleware struct {
	metrics *ServerMetrics
}

func NewMetricsMiddleware(metrics *ServerMetrics) *MetricsMiddleware {
	return &MetricsMiddleware{metrics: metrics}
}

func (m *MetricsMiddleware) Name() string {
	return "metrics"
}

func (m *MetricsMiddleware) Priority() int {
	return 900 // 高优先级
}

func (m *MetricsMiddleware) Handle(ctx *EnhancedContext) error {
	if m.metrics != nil {
		ctx.startTime = time.Now()

		// 增加请求计数
		m.metrics.RequestCount.Add(ctx.Context, 1)

		// 记录字节数
		m.metrics.BytesReceived.Add(ctx.Context, int64(len(ctx.Payload)))
	}
	return nil
}

// RateLimitMiddleware 限流中间件
type RateLimitMiddleware struct {
	limiter RateLimiterInterface
}

// RateLimiterInterface 限流器接口
type RateLimiterInterface interface {
	Allow(key string) bool
	Reset(key string)
}

func NewRateLimitMiddleware(limiter RateLimiterInterface) *RateLimitMiddleware {
	return &RateLimitMiddleware{limiter: limiter}
}

func (m *RateLimitMiddleware) Name() string {
	return "rate_limit"
}

func (m *RateLimitMiddleware) Priority() int {
	return 800
}

func (m *RateLimitMiddleware) Handle(ctx *EnhancedContext) error {
	if m.limiter != nil {
		clientAddr := ctx.GetRemoteAddr()
		if !m.limiter.Allow(clientAddr) {
			return fmt.Errorf("rate limit exceeded for %s", clientAddr)
		}
	}
	return nil
}

// AuthMiddleware 认证中间件
type AuthMiddleware struct {
	authenticator Authenticator
}

type Authenticator interface {
	Authenticate(token string) (*User, error)
}

type User struct {
	ID       string
	Username string
	Roles    []string
}

func NewAuthMiddleware(auth Authenticator) *AuthMiddleware {
	return &AuthMiddleware{authenticator: auth}
}

func (m *AuthMiddleware) Name() string {
	return "auth"
}

func (m *AuthMiddleware) Priority() int {
	return 700
}

func (m *AuthMiddleware) Handle(ctx *EnhancedContext) error {
	if m.authenticator != nil {
		token := ctx.Metadata["authorization"]
		if token == "" {
			return fmt.Errorf("missing authorization token")
		}

		user, err := m.authenticator.Authenticate(token)
		if err != nil {
			return fmt.Errorf("authentication failed: %w", err)
		}

		// 将用户信息存储到上下文
		ctx.Metadata["user_id"] = user.ID
		ctx.Metadata["username"] = user.Username
	}
	return nil
}

// 使用示例
func ExampleEnhancedServer() {
	cfg := &config.QUICServer{
		Address:  "localhost:8443",
		CertFile: "server.crt",
		KeyFile:  "server.key",
	}

	server := NewEnhancedServer(cfg)

	// 添加增强中间件
	// server.AddEnhancedMiddleware(NewTracingMiddleware(tracer))
	// server.AddEnhancedMiddleware(NewMetricsMiddleware(metrics))
	// server.AddEnhancedMiddleware(NewRateLimitMiddleware(rateLimiter))
	// server.AddEnhancedMiddleware(NewAuthMiddleware(authenticator))

	// 设置服务注册
	// server.WithServiceRegistry(consulRegistry)

	// 添加处理器
	server.AddHandler(1000, func(ctx *Context) {
		// 业务逻辑
		ctx.Write([]byte("enhanced response"))
	})

	// 启动服务器
	// server.Start()
}
