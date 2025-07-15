package tcp

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/ilaziness/gokit/config"
	"github.com/ilaziness/gokit/hook"
	"github.com/ilaziness/gokit/log"
	"github.com/ilaziness/gokit/process"
)

const (
	defaultWorkerNum = 100000
)

type (
	// Handler 处理函数
	Handler func(ctx *Context)
	// OpCode 操作码, 业务使用>=1000，1000以下作为保留
	OpCode uint
)

type Server struct {
	config      *config.TCPServer
	workerSem   chan struct{} // 控制同时执行的请求数
	handlers    map[OpCode]Handler
	middlewares []Handler
	ctxPool     sync.Pool
	packCodec   Codec
}

// NewTCP 创建一个tcp服务，不含任何中间件
func NewTCP(config *config.TCPServer) *Server {
	return newBaseTCP(config)
}

// NewDefaultTCP 创建一个包含默认中间件的tcp服务
func NewDefaultTCP(config *config.TCPServer) *Server {
	srv := newBaseTCP(config)
	srv.setDefaultMiddleware()
	return srv
}

func newBaseTCP(config *config.TCPServer) *Server {
	workerNum := defaultWorkerNum
	if config.WorkerNum > 0 {
		workerNum = config.WorkerNum
	}
	return &Server{
		config:      config,
		workerSem:   make(chan struct{}, workerNum),
		handlers:    make(map[OpCode]Handler),
		packCodec:   NewPackCodec(),
		middlewares: []Handler{},
		ctxPool: sync.Pool{
			New: func() any {
				return &Context{
					index:   -1,
					Context: context.Background(),
				}
			},
		},
	}
}

// AddHandler 添加请求处理函数
func (t *Server) AddHandler(oc OpCode, h Handler) {
	t.handlers[oc] = h
}

// AddMiddleware 添加中间件
func (t *Server) AddMiddleware(ms ...Handler) {
	t.middlewares = append(t.middlewares, ms...)
}

func (t *Server) setDefaultMiddleware() {
	t.AddMiddleware(Ping)
}

func (t *Server) Start() {
	if t.config.Debug {
		log.SetLevel(log.ModeDebug)
	} else {
		log.SetLevel(log.ModeRelease)
	}
	// otel.InitTracer("tcp-server", &config.Otel{})

	var ln net.Listener
	var err error
	ctx, stop := context.WithCancel(context.Background())

	// 如果提供了TLS配置，则使用TLS监听
	tlsConfig := t.loadTLSConfig()
	if tlsConfig != nil {
		ln, err = tls.Listen("tcp", t.config.Address, tlsConfig)
		log.Info(ctx, "start tls server")
	} else {
		ln, err = net.Listen("tcp", t.config.Address)
	}

	if err != nil {
		panic(err)
	}
	process.SafeGo(func() {
		for {
			conn, err := ln.Accept()
			if err != nil && errors.Is(err, net.ErrClosed) {
				return
			}
			if err != nil {
				log.Error(ctx, "accept error: %s", err)
				continue
			}
			log.Debug(ctx, "accept new conn: %s", conn.RemoteAddr())
			process.SafeGo(func() {
				t.handleConn(conn)
			})
		}
	})
	log.Info(ctx, "tcp server start at: %s", ln.Addr())
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	stop()
	hook.Exit.Trigger()

	if err = ln.Close(); err != nil {
		log.Warn(ctx, "close listener error: %s", err)
	}

	log.Logger.Infoln("Shutdown Server ...")
	time.Sleep(time.Second * 2)
	log.Logger.Infoln("Server Shutdown")
}

// 加载TLS配置
func (t *Server) loadTLSConfig() *tls.Config {
	if t.config.CertFile == "" || t.config.KeyFile == "" {
		return nil
	}
	cert, err := tls.LoadX509KeyPair(t.config.CertFile, t.config.KeyFile)
	if err != nil {
		log.Error(context.Background(), "load tls config error: %s", err)
		return nil
	}
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		// 推荐的最低 TLS 版本
		MinVersion: tls.VersionTLS12,
		// 密钥交互算法列表
		CurvePreferences: []tls.CurveID{
			tls.X25519MLKEM768,
			tls.CurveP256,
			tls.X25519,
		},
		// 加密套件列表
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
		},
		// ClientAuth: tls.RequireAndVerifyClientCert,
		// ClientCAs: nil,
	}
	return tlsConfig
}

func (t *Server) handleConn(conn net.Conn) {
	defer func(conn net.Conn) {
		log.Debug(context.Background(), "connection closed: %s", conn.RemoteAddr())
		err := conn.Close()
		if err != nil {
			slog.Warn("conn close error", "err", err.Error())
		}
	}(conn)
	handleMessage := func() (ct bool) {
		// 设置读取超时
		_ = conn.SetDeadline(time.Now().Add(300 * time.Second))
		ctx := t.ctxPool.Get().(*Context)
		ctx.Reset(conn, t.packCodec)
		pack, err := t.packCodec.Decode(conn)
		if err != nil && errors.Is(err, io.EOF) {
			log.Debug(ctx, "connection closed")
			return
		}
		if err != nil {
			log.Warn(ctx, "decode error: %s", err)
			return
		}
		ctx.SetData(pack)
		t.workerSem <- struct{}{}

		// 处理数据包
		process.SafeGo(func() {
			defer func() {
				<-t.workerSem
			}()
			defer t.ctxPool.Put(ctx)

			ctx.handler = t.middlewares
			handler := t.handlers[ctx.OpCode]
			if handler == nil {
				ctx.handler = append(ctx.handler, func(ctx *Context) {
					_ = ctx.WriteNotFund()
				})
			} else {
				ctx.handler = append(ctx.handler, handler)
			}
			ctx.Next()
			log.Debug(ctx, "read data: %s", pack.Payload)
		})
		return true
	}

	for {
		if !handleMessage() {
			return
		}
	}
}

type Context struct {
	context.Context

	index     int
	isAbort   bool
	handler   []Handler
	packCodec Codec
	Conn      net.Conn
	Pack      *Pack
	SQID      uint32
	OpCode    OpCode
	Payload   []byte
}

func (c *Context) Reset(conn net.Conn, pc Codec) {
	c.index = -1
	c.isAbort = false
	c.handler = nil
	c.Conn = conn
	c.packCodec = pc
}

// Next 运行中间件
func (c *Context) Next() {
	c.index++
	l := len(c.handler)
	for ; c.index < l; c.index++ {
		if c.isAbort {
			// 中断执行
			return
		}
		c.handler[c.index](c)
	}
}

// SetData 设置数据
func (c *Context) SetData(pack *Pack) {
	c.SQID = pack.Head.SQID
	c.OpCode = OpCode(pack.Head.OpCode)
	c.Payload = pack.Payload

	c.Pack = pack
}

// Write 写入一般响应数据
func (c *Context) Write(data []byte) error {
	pack := &Pack{
		Head: PackHead{
			SQID:    c.SQID,
			Version: c.Pack.Head.Version,
		},
		Payload: data,
	}
	return c.packCodec.Encode(c.Conn, pack)
}

// WriteWithOpCode 写入指定操作码响应数据
func (c *Context) WriteWithOpCode(opcode OpCode, data []byte) error {
	pack := &Pack{
		Head: PackHead{
			OpCode:  uint16(opcode),
			SQID:    c.SQID,
			Version: c.Pack.Head.Version,
		},
		Payload: data,
	}
	return c.packCodec.Encode(c.Conn, pack)
}

// ServerErr 写入服务错误响应
func (c *Context) ServerErr() error {
	return c.WriteWithOpCode(OpCodeServerErr, nil)
}

// WriteNotFund 写入404错误响应
func (c *Context) WriteNotFund() error {
	return c.WriteWithOpCode(OpCodeNotFound, nil)
}

// Abort 中断处理
func (c *Context) Abort() {
	c.isAbort = true
}
