package quic

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/ilaziness/gokit/config"
	"github.com/ilaziness/gokit/hook"
	"github.com/ilaziness/gokit/log"
	"github.com/ilaziness/gokit/process"
	"github.com/quic-go/quic-go"
)

const (
	defaultWorkerNum        = 100000
	defaultBufferSize       = 65536
	defaultIdleTimeout      = 30 * time.Second
	defaultKeepAlive        = 15 * time.Second
	defaultHandshakeTimeout = 10 * time.Second
)

type (
	// Handler 处理函数
	Handler func(ctx *Context)
	// OpCode 操作码, 业务使用>=1000，1000以下作为保留
	OpCode uint
	// StreamHandler 流处理函数
	StreamHandler func(ctx *StreamContext)
)

// Server QUIC服务器
type Server struct {
	config         *config.QUICServer
	workerSem      chan struct{} // 控制同时执行的请求数
	handlers       map[OpCode]Handler
	streamHandlers map[OpCode]StreamHandler
	middlewares    []Handler
	ctxPool        sync.Pool
	streamCtxPool  sync.Pool
	packCodec      Codec
	listener       *quic.Listener
	tlsConfig      *tls.Config
	quicConfig     *quic.Config
}

// NewQUIC 创建一个QUIC服务，不含任何中间件
func NewQUIC(config *config.QUICServer) *Server {
	return newBaseQUIC(config)
}

// NewDefaultQUIC 创建一个包含默认中间件的QUIC服务
func NewDefaultQUIC(config *config.QUICServer) *Server {
	srv := newBaseQUIC(config)
	srv.setDefaultMiddleware()
	return srv
}

func newBaseQUIC(config *config.QUICServer) *Server {
	workerNum := defaultWorkerNum
	if config.WorkerNum > 0 {
		workerNum = config.WorkerNum
	}

	return &Server{
		config:         config,
		workerSem:      make(chan struct{}, workerNum),
		handlers:       make(map[OpCode]Handler),
		streamHandlers: make(map[OpCode]StreamHandler),
		packCodec:      NewPackCodec(),
		middlewares:    []Handler{},
		ctxPool: sync.Pool{
			New: func() any {
				return &Context{
					index:   -1,
					Context: context.Background(),
				}
			},
		},
		streamCtxPool: sync.Pool{
			New: func() any {
				return &StreamContext{
					Context: context.Background(),
				}
			},
		},
	}
}

// AddHandler 添加数据报处理函数
func (s *Server) AddHandler(oc OpCode, h Handler) {
	s.handlers[oc] = h
}

// AddStreamHandler 添加流处理函数
func (s *Server) AddStreamHandler(oc OpCode, h StreamHandler) {
	s.streamHandlers[oc] = h
}

// AddMiddleware 添加中间件
func (s *Server) AddMiddleware(ms ...Handler) {
	s.middlewares = append(s.middlewares, ms...)
}

func (s *Server) setDefaultMiddleware() {
	s.AddMiddleware(Ping)
}

// Start 启动QUIC服务器
func (s *Server) Start() {
	if s.config.Debug {
		log.SetLevel(log.ModeDebug)
	} else {
		log.SetLevel(log.ModeRelease)
	}

	ctx, stop := context.WithCancel(context.Background())

	// 初始化TLS和QUIC配置
	if err := s.initConfigs(); err != nil {
		panic(fmt.Sprintf("init configs error: %v", err))
	}

	// 创建QUIC监听器
	var err error
	s.listener, err = quic.ListenAddr(s.config.Address, s.tlsConfig, s.quicConfig)
	if err != nil {
		panic(fmt.Sprintf("listen QUIC error: %v", err))
	}

	log.Info(ctx, "QUIC server start at: %s", s.listener.Addr())

	// 启动连接处理协程
	process.SafeGo(func() {
		s.handleConnections(ctx)
	})

	// 等待退出信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	stop()
	hook.Exit.Trigger()

	// 关闭监听器
	if err = s.listener.Close(); err != nil {
		log.Warn(ctx, "close QUIC listener error: %s", err)
	}

	log.Logger.Infoln("Shutdown Server ...")
	time.Sleep(time.Second * 2)
	log.Logger.Infoln("Server Shutdown")
}

// initConfigs 初始化TLS和QUIC配置
func (s *Server) initConfigs() error {
	// 初始化TLS配置
	s.tlsConfig = s.createTLSConfig()
	if s.tlsConfig == nil {
		return errors.New("TLS configuration is required for QUIC")
	}

	// 初始化QUIC配置
	s.quicConfig = s.createQUICConfig()
	return nil
}

// createTLSConfig 创建TLS配置
func (s *Server) createTLSConfig() *tls.Config {
	if s.config.CertFile == "" || s.config.KeyFile == "" {
		log.Error(context.Background(), "cert_file and key_file are required for QUIC server")
		return nil
	}

	cert, err := tls.LoadX509KeyPair(s.config.CertFile, s.config.KeyFile)
	if err != nil {
		log.Error(context.Background(), "load TLS config error: %s", err)
		return nil
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		NextProtos:   []string{"quic-server"},
		MinVersion:   tls.VersionTLS13, // QUIC requires TLS 1.3
		CurvePreferences: []tls.CurveID{
			tls.X25519MLKEM768,
			tls.CurveP256,
			tls.X25519,
		},
		CipherSuites: []uint16{
			tls.TLS_AES_128_GCM_SHA256,
			tls.TLS_AES_256_GCM_SHA384,
			tls.TLS_CHACHA20_POLY1305_SHA256,
		},
	}
}

// createQUICConfig 创建QUIC配置
func (s *Server) createQUICConfig() *quic.Config {
	idleTimeout := defaultIdleTimeout
	if s.config.IdleTimeout > 0 {
		idleTimeout = time.Duration(s.config.IdleTimeout) * time.Second
	}

	keepAlive := defaultKeepAlive
	if s.config.KeepAlive > 0 {
		keepAlive = time.Duration(s.config.KeepAlive) * time.Second
	}

	handshakeTimeout := defaultHandshakeTimeout
	if s.config.HandshakeTimeout > 0 {
		handshakeTimeout = time.Duration(s.config.HandshakeTimeout) * time.Second
	}

	maxStreams := int64(1000)
	if s.config.MaxStreams > 0 {
		maxStreams = int64(s.config.MaxStreams)
	}

	return &quic.Config{
		MaxIdleTimeout:                 idleTimeout,
		MaxIncomingStreams:             maxStreams,
		MaxIncomingUniStreams:          maxStreams,
		HandshakeIdleTimeout:           handshakeTimeout,
		KeepAlivePeriod:                keepAlive,
		InitialStreamReceiveWindow:     defaultBufferSize * 10,
		MaxStreamReceiveWindow:         defaultBufferSize * 100,
		InitialConnectionReceiveWindow: defaultBufferSize * 10,
		MaxConnectionReceiveWindow:     defaultBufferSize * 100,
		Allow0RTT:                      s.config.Allow0RTT,
		EnableDatagrams:                true,
	}
}

// handleConnections 处理QUIC连接
func (s *Server) handleConnections(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			conn, err := s.listener.Accept(ctx)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return
				}
				log.Warn(ctx, "accept QUIC connection error: %s", err)
				continue
			}

			log.Debug(ctx, "accepted QUIC connection from: %s", conn.RemoteAddr())

			// 为每个连接启动处理协程
			process.SafeGo(func() {
				s.handleConnection(ctx, conn)
			})
		}
	}
}

// handleConnection 处理单个QUIC连接
func (s *Server) handleConnection(ctx context.Context, conn quic.Connection) {
	defer func() {
		if err := conn.CloseWithError(0, "server shutdown"); err != nil {
			log.Debug(ctx, "close connection error: %s", err)
		}
	}()

	// 启动数据报处理协程
	process.SafeGo(func() {
		s.handleDatagrams(ctx, conn)
	})

	// 处理流连接
	for {
		select {
		case <-ctx.Done():
			return
		case <-conn.Context().Done():
			return
		default:
			stream, err := conn.AcceptStream(ctx)
			if err != nil {
				if errors.Is(err, context.Canceled) ||
					errors.Is(err, io.EOF) {
					return
				}
				log.Warn(ctx, "accept stream error: %s", err)
				continue
			}

			log.Debug(ctx, "accepted stream %d from: %s", stream.StreamID(), conn.RemoteAddr())

			// 为每个流启动处理协程
			process.SafeGo(func() {
				s.handleStream(ctx, conn, stream)
			})
		}
	}
}

// handleDatagrams 处理数据报
func (s *Server) handleDatagrams(ctx context.Context, conn quic.Connection) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-conn.Context().Done():
			return
		default:
			data, err := conn.ReceiveDatagram(ctx)
			if err != nil {
				if errors.Is(err, context.Canceled) ||
					errors.Is(err, io.EOF) {
					return
				}
				log.Warn(ctx, "receive datagram error: %s", err)
				continue
			}

			log.Debug(ctx, "received datagram from: %s, size: %d", conn.RemoteAddr(), len(data))

			// 异步处理数据报
			process.SafeGo(func() {
				s.handleDatagram(ctx, conn, data)
			})
		}
	}
}

// handleDatagram 处理单个数据报
func (s *Server) handleDatagram(ctx context.Context, conn quic.Connection, data []byte) {
	s.workerSem <- struct{}{}
	defer func() {
		<-s.workerSem
	}()

	// 解码数据包
	pack, err := s.packCodec.Decode(data)
	if err != nil {
		log.Warn(ctx, "decode datagram error: %s", err)
		return
	}

	// 获取上下文对象
	ctxObj := s.ctxPool.Get().(*Context)
	defer s.ctxPool.Put(ctxObj)

	ctxObj.Reset(conn, s.packCodec)
	ctxObj.SetData(pack)

	// 设置中间件和处理器
	ctxObj.handler = s.middlewares
	handler := s.handlers[ctxObj.OpCode]
	if handler == nil {
		ctxObj.handler = append(ctxObj.handler, func(ctx *Context) {
			_ = ctx.WriteNotFound()
		})
	} else {
		ctxObj.handler = append(ctxObj.handler, handler)
	}

	// 执行处理链
	ctxObj.Next()
	log.Debug(ctx, "processed datagram: %s", pack.Payload)
}

// handleStream 处理流
func (s *Server) handleStream(ctx context.Context, conn quic.Connection, stream quic.Stream) {
	defer stream.Close()

	s.workerSem <- struct{}{}
	defer func() {
		<-s.workerSem
	}()

	// 读取流数据
	buffer := make([]byte, defaultBufferSize)
	n, err := stream.Read(buffer)
	if err != nil && !errors.Is(err, io.EOF) {
		log.Warn(ctx, "read stream error: %s", err)
		return
	}

	if n == 0 {
		return
	}

	data := buffer[:n]
	log.Debug(ctx, "received stream data from: %s, size: %d", conn.RemoteAddr(), n)

	// 解码数据包
	pack, err := s.packCodec.Decode(data)
	if err != nil {
		log.Warn(ctx, "decode stream packet error: %s", err)
		return
	}

	// 获取流上下文对象
	streamCtx := s.streamCtxPool.Get().(*StreamContext)
	defer s.streamCtxPool.Put(streamCtx)

	streamCtx.Reset(conn, stream, s.packCodec)
	streamCtx.SetData(pack)

	// 查找流处理器
	handler := s.streamHandlers[streamCtx.OpCode]
	if handler == nil {
		_ = streamCtx.WriteNotFound()
		return
	}

	// 执行流处理器
	handler(streamCtx)
	log.Debug(ctx, "processed stream: %s", pack.Payload)
}
