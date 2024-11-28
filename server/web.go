// Package server provide web engine
package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
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
	Gin *gin.Engine
	cfg *config.App
}

// NewWeb 创建一个web app
func NewWeb(appCfg *config.App) *WebApp {
	if appCfg.Mode == gin.ReleaseMode {
		gin.SetMode(gin.ReleaseMode)
	}
	a := &WebApp{
		Gin: NewGin(),
		cfg: appCfg,
	}
	a.setDefaultMiddleware()
	return a
}

// NewGin gin engine
func NewGin() *gin.Engine {
	e := gin.New()
	// 没有这个设置gin context和原生content会不兼容
	e.ContextWithFallback = true
	return e
}

// Run 运行应用
func (a *WebApp) Run() {
	a.starup()
	srv := http.Server{
		Addr:    fmt.Sprintf(":%d", a.cfg.Port),
		Handler: a.Gin,
	}
	go func() {
		log.Logger.Infof("app [%s] started on %s", a.cfg.ID, srv.Addr)
		err := srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Logger.Fatal("Start Server error:", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
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
	a.Gin.Use(middleware.LogReq(), gin.CustomRecoveryWithWriter(nil, middleware.RecoveryHandle))
	corsCfg := cors.DefaultConfig()
	corsCfg.AllowOrigins = []string{"*"}
	if a.cfg.Cors != nil {
		if len(a.cfg.Cors.AllowOrigin) > 0 {
			corsCfg.AllowOrigins = a.cfg.Cors.AllowOrigin
		}
		if len(a.cfg.Cors.AllowMethods) > 0 {
			corsCfg.AllowMethods = a.cfg.Cors.AllowMethods
		}
		if len(a.cfg.Cors.AllowHeaders) > 0 {
			corsCfg.AllowHeaders = a.cfg.Cors.AllowHeaders
		}
		corsCfg.AllowCredentials = a.cfg.Cors.AllowCredentials
	}
	a.Gin.Use(cors.New(corsCfg))
	a.Gin.Use(middleware.Otel(a.cfg.ID))
}

func (a *WebApp) starup() {
	timer.Run()
}

func (a *WebApp) destroy() {
	hook.Exit.Trigger()
}
