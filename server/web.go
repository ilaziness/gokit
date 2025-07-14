// Package server provide web engine
package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/ilaziness/gokit/config"
	"github.com/ilaziness/gokit/hook"
	"github.com/ilaziness/gokit/log"
	"github.com/ilaziness/gokit/middleware"
	"github.com/ilaziness/gokit/timer"
)

type WebApp struct {
	Gin    *gin.Engine
	config *config.App
}

func init() {
	hook.Start.Register(timer.Run)
}

// NewWeb 创建一个web app
func NewWeb(appCfg *config.App) *WebApp {
	if appCfg.Mode == gin.ReleaseMode {
		gin.SetMode(gin.ReleaseMode)
	}
	a := &WebApp{
		Gin:    NewGin(),
		config: appCfg,
	}
	a.setDefaultMiddleware()
	return a
}

// NewGin gin engine
func NewGin() *gin.Engine {
	e := gin.New()
	// 没有这个设置gin context和原生Request的content会不兼容
	e.ContextWithFallback = true
	return e
}

// Run 运行应用
func (a *WebApp) Run() {
	a.starup()
	srv := http.Server{
		Addr:    fmt.Sprintf(":%d", a.config.Port),
		Handler: a.Gin,
	}
	go func() {
		log.Logger.Infof("app [%s] started on %s", a.config.Name, srv.Addr)
		err := srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Logger.Fatal("Start Server error:", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Logger.Infoln("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Logger.Fatal("Server Shutdown error:", err)
	}
	a.destroy()
	log.Logger.Info("Server exiting")
}

func (a *WebApp) setDefaultMiddleware() {
	a.Gin.Use(gin.CustomRecoveryWithWriter(nil, middleware.RecoveryHandle))
	if a.config.LogReq {
		a.Gin.Use(middleware.LogReq())
	}
	corsCfg := cors.DefaultConfig()
	corsCfg.AllowOrigins = []string{"*"}
	if a.config.Cors != nil {
		if len(a.config.Cors.AllowOrigin) > 0 {
			corsCfg.AllowOrigins = a.config.Cors.AllowOrigin
		}
		if len(a.config.Cors.AllowMethods) > 0 {
			corsCfg.AllowMethods = a.config.Cors.AllowMethods
		}
		if len(a.config.Cors.AllowHeaders) > 0 {
			corsCfg.AllowHeaders = a.config.Cors.AllowHeaders
		}
		corsCfg.AllowCredentials = a.config.Cors.AllowCredentials
	}
	a.Gin.Use(cors.New(corsCfg))
	a.Gin.Use(middleware.Otel(a.config.Name))
}

func (a *WebApp) starup() {
	hook.Start.Trigger()
	a.initPprof()
}

func (a *WebApp) destroy() {
	hook.Exit.Trigger()
}

// initPprof 初始化pprof功能
// 需要在main包里面导入pprof包`_ "net/http/pprof"`
func (a *WebApp) initPprof() {
	if !a.config.Pprof {
		return
	}
	a.Gin.GET("/debug/pprof/*any", gin.WrapH(http.DefaultServeMux))
}
