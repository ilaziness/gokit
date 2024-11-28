package redis

import (
	"context"
	"fmt"

	"github.com/ilaziness/gokit/config"
)

var Client *redis.Client

func Init(cfg *config.Redis) {
	if cfg.Port == 0 {
		cfg.Port = 6379
	}
	Client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Pass,
		DB:       int(cfg.Db),
	})
	status := Client.Ping(context.Background())
	if status.Err() != nil {
		panic(status.Err())
	}
}
