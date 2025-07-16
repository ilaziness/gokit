package udp

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/ilaziness/gokit/config"
	"github.com/ilaziness/gokit/hook"
	"github.com/ilaziness/gokit/log"
	"github.com/ilaziness/gokit/process"
	"github.com/pion/dtls/v3"
)

const (
	defaultWorkerNum  = 100000
	defaultBufferSize = 65536
)

type (
	// Handler 处理函数
	Handler func(ctx *Context)
	// OpCode 操作码, 业务使用>=1000，1000以下作为保留
	OpCode uint
)

type Server struct {
	config       *config.UDPServer
	workerSem    chan struct{} // 控制同时执行的请求数
	handlers     map[OpCode]Handler
	middlewares  []Handler
	ctxPool      sync.Pool
	packCodec    Codec
	conn         net.PacketConn
	dtlsListener net.Listener
	dtlsConfig   *dtls.Config
	isDTLS       bool
}

// NewUDP 创建一个UDP服务，不含任何中间件
func NewUDP(config *config.UDPServer) *Server {
	return newBaseUDP(config)
}

// NewDefaultUDP 创建一个包含默认中间件的UDP服务
func NewDefaultUDP(config *config.UDPServer) *Server {
	srv := newBaseUDP(config)
	srv.setDefaultMiddleware()
	return srv
}

func newBaseUDP(config *config.UDPServer) *Server {
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
func (s *Server) AddHandler(oc OpCode, h Handler) {
	s.handlers[oc] = h
}

// AddMiddleware 添加中间件
func (s *Server) AddMiddleware(ms ...Handler) {
	s.middlewares = append(s.middlewares, ms...)
}

func (s *Server) setDefaultMiddleware() {
	s.AddMiddleware(Ping)
}

func (s *Server) Start() {
	if s.config.Debug {
		log.SetLevel(log.ModeDebug)
	} else {
		log.SetLevel(log.ModeRelease)
	}

	var err error
	ctx, stop := context.WithCancel(context.Background())

	// 创建DTLS配置
	s.dtlsConfig = s.createDTLSConfig()

	// 如果有DTLS配置，创建DTLS监听器
	if s.dtlsConfig != nil {
		s.isDTLS = true
		s.dtlsListener, err = s.createDTLSListener()
		if err != nil {
			panic(err)
		}
		log.Info(ctx, "start DTLS UDP server")

		// 启动DTLS连接处理协程
		process.SafeGo(func() {
			s.handleDTLSConnections(ctx)
		})

		log.Info(ctx, "DTLS UDP server start at: %s", s.dtlsListener.Addr())
	} else {
		// 创建普通UDP连接
		s.conn, err = net.ListenPacket("udp", s.config.Address)
		if err != nil {
			panic(err)
		}
		log.Info(ctx, "start UDP server")

		// 启动消息处理协程
		process.SafeGo(func() {
			s.handleMessages(ctx)
		})

		log.Info(ctx, "UDP server start at: %s", s.conn.LocalAddr())
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	stop()
	hook.Exit.Trigger()

	if s.isDTLS {
		if err = s.dtlsListener.Close(); err != nil {
			log.Warn(ctx, "close DTLS listener error: %s", err)
		}
	} else {
		if err = s.conn.Close(); err != nil {
			log.Warn(ctx, "close UDP connection error: %s", err)
		}
	}

	log.Logger.Infoln("Shutdown Server ...")
	time.Sleep(time.Second * 2)
	log.Logger.Infoln("Server Shutdown")
}

// createDTLSConfig 创建DTLS配置
func (s *Server) createDTLSConfig() *dtls.Config {
	if s.config.CertFile == "" || s.config.KeyFile == "" {
		return nil
	}

	cert, err := tls.LoadX509KeyPair(s.config.CertFile, s.config.KeyFile)
	if err != nil {
		panic(fmt.Sprintf("load DTLS config error: %s", err))
	}

	return &dtls.Config{
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: false,
		CipherSuites: []dtls.CipherSuiteID{
			dtls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			dtls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			dtls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
			dtls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
		},
		ClientAuth:           dtls.NoClientCert,
		ExtendedMasterSecret: dtls.RequireExtendedMasterSecret,
		FlightInterval:       time.Second,
	}
}

// createDTLSListener 创建DTLS监听器
func (s *Server) createDTLSListener() (net.Listener, error) {
	addr, err := net.ResolveUDPAddr("udp", s.config.Address)
	if err != nil {
		return nil, err
	}

	// 创建DTLS监听器
	listener, err := dtls.Listen("udp", addr, s.dtlsConfig)
	if err != nil {
		return nil, err
	}

	return listener, nil
}

// handleDTLSConnections 处理DTLS连接
func (s *Server) handleDTLSConnections(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// 接受新的DTLS连接
			conn, err := s.dtlsListener.Accept()
			if err != nil {
				if errors.Is(err, net.ErrClosed) {
					return
				}
				log.Warn(ctx, "accept DTLS connection error: %s", err)
				continue
			}

			log.Debug(ctx, "accepted DTLS connection from: %s", conn.RemoteAddr())

			// 为每个连接启动处理协程
			process.SafeGo(func() {
				s.handleDTLSConnection(ctx, conn)
			})
		}
	}
}

// handleDTLSConnection 处理单个DTLS连接
func (s *Server) handleDTLSConnection(ctx context.Context, conn net.Conn) {
	defer conn.Close()

	buffer := make([]byte, defaultBufferSize)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			// 设置读取超时
			_ = conn.SetDeadline(time.Now().Add(300 * time.Second))

			n, err := conn.Read(buffer)
			if err != nil {
				// 连接关闭或超时错误不打印日志，直接返回
				if errors.Is(err, io.EOF) ||
					strings.Contains(err.Error(), "timeout") {
					return
				}
				log.Warn(ctx, "read DTLS data error: %s", err)
				return
			}

			log.Debug(ctx, "received DTLS data from: %s, size: %d", conn.RemoteAddr(), n)

			// 复制数据以避免并发问题
			data := make([]byte, n)
			copy(data, buffer[:n])

			// 异步处理消息
			process.SafeGo(func() {
				s.handleDTLSPacket(ctx, data, conn)
			})
		}
	}
}

// handleDTLSPacket 处理DTLS数据包
func (s *Server) handleDTLSPacket(ctx context.Context, data []byte, conn net.Conn) {
	s.workerSem <- struct{}{}
	defer func() {
		<-s.workerSem
	}()

	// 解码数据包
	pack, err := s.packCodec.Decode(data)
	if err != nil {
		log.Warn(ctx, "decode DTLS packet error: %s", err)
		return
	}

	// 获取上下文对象
	ctxObj := s.ctxPool.Get().(*Context)
	defer s.ctxPool.Put(ctxObj)

	// 为DTLS连接重置上下文
	ctxObj.ResetForDTLS(conn, s.packCodec)
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
	log.Debug(ctx, "processed DTLS packet: %s", pack.Payload)
}

func (s *Server) handleMessages(ctx context.Context) {
	buffer := make([]byte, defaultBufferSize)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			// 设置读取超时
			_ = s.conn.SetDeadline(time.Now().Add(300 * time.Second))

			n, addr, err := s.conn.ReadFrom(buffer)
			if err != nil {
				// 连接关闭或超时错误不打印日志，直接返回或继续
				if errors.Is(err, io.EOF) ||
					strings.Contains(err.Error(), "timeout") {
					return
				}
				log.Warn(ctx, "read UDP packet error: %s", err)
				continue
			}

			log.Debug(ctx, "received UDP packet from: %s, size: %d", addr, n)

			// 复制数据以避免并发问题
			data := make([]byte, n)
			copy(data, buffer[:n])

			// 异步处理消息
			process.SafeGo(func() {
				s.handlePacket(ctx, data, addr)
			})
		}
	}
}

func (s *Server) handlePacket(ctx context.Context, data []byte, addr net.Addr) {
	s.workerSem <- struct{}{}
	defer func() {
		<-s.workerSem
	}()

	// 解码数据包
	pack, err := s.packCodec.Decode(data)
	if err != nil {
		log.Warn(ctx, "decode UDP packet error: %s", err)
		return
	}

	// 获取上下文对象
	ctxObj := s.ctxPool.Get().(*Context)
	defer s.ctxPool.Put(ctxObj)

	ctxObj.Reset(s.conn, addr, s.packCodec)
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
	log.Debug(ctx, "processed UDP packet: %s", pack.Payload)
}

type Context struct {
	context.Context

	index     int
	isAbort   bool
	handler   []Handler
	packCodec Codec
	conn      net.PacketConn
	dtlsConn  net.Conn // DTLS连接
	addr      net.Addr
	Pack      *Pack
	SQID      uint32
	OpCode    OpCode
	Payload   []byte
	isDTLS    bool
}

func (c *Context) Reset(conn net.PacketConn, addr net.Addr, pc Codec) {
	c.index = -1
	c.isAbort = false
	c.handler = nil
	c.conn = conn
	c.dtlsConn = nil
	c.addr = addr
	c.packCodec = pc
	c.isDTLS = false
}

// ResetForDTLS 为DTLS连接重置上下文
func (c *Context) ResetForDTLS(conn net.Conn, pc Codec) {
	c.index = -1
	c.isAbort = false
	c.handler = nil
	c.conn = nil
	c.dtlsConn = conn
	c.addr = conn.RemoteAddr()
	c.packCodec = pc
	c.isDTLS = true
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
	return c.sendPacket(pack)
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
	return c.sendPacket(pack)
}

// ServerErr 写入服务错误响应
func (c *Context) ServerErr() error {
	return c.WriteWithOpCode(OpCodeServerErr, nil)
}

// WriteNotFound 写入404错误响应
func (c *Context) WriteNotFound() error {
	return c.WriteWithOpCode(OpCodeNotFound, nil)
}

// Abort 中断处理
func (c *Context) Abort() {
	c.isAbort = true
}

// GetRemoteAddr 获取客户端地址
func (c *Context) GetRemoteAddr() net.Addr {
	return c.addr
}

func (c *Context) sendPacket(pack *Pack) error {
	data, err := c.packCodec.Encode(pack)
	if err != nil {
		return err
	}

	if c.isDTLS {
		// DTLS连接直接写入
		_, err = c.dtlsConn.Write(data)
		return err
	} else {
		// UDP连接使用WriteTo
		_, err = c.conn.WriteTo(data, c.addr)
		return err
	}
}
