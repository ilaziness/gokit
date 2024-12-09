package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/ilaziness/gokit/config"
	"github.com/ilaziness/gokit/hook"
	redisLib "github.com/redis/go-redis/v9"
)

var Client *redisLib.Client
var redLock *redsync.Redsync

// Init 初始化redis连接对象，并且可选初始化redsync分布式锁
func Init(cfg *config.Redis, distLock bool) {
	if cfg.Port == 0 {
		cfg.Port = 6379
	}
	Client = redisLib.NewClient(&redisLib.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Pass,
		DB:       int(cfg.Db),
	})
	status := Client.Ping(context.Background())
	if status.Err() != nil {
		panic(status.Err())
	}
	if distLock {
		pool := goredis.NewPool(Client)
		redLock = redsync.New(pool)
	}
	hook.Exit.Register(func() {
		_ = Client.Close()
	})
}

// Lock 获取redsync锁
func Lock(name string, expiry int) (*redsync.Mutex, error) {
	mutx := redLock.NewMutex(name, redsync.WithExpiry(time.Duration(expiry)*time.Second))
	if err := mutx.Lock(); err != nil {
		return nil, err
	}
	return mutx, nil
}
