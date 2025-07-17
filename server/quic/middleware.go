package quic

import (
	"github.com/ilaziness/gokit/log"
)

// Ping ping中间件
func Ping(ctx *Context) {
	if ctx.OpCode != OpCodePing {
		return
	}
	if err := ctx.WriteWithOpCode(OpCodePong, nil); err != nil {
		log.Error(ctx, "ping response write error: %s", err)
	}
	ctx.Abort()
}

// RateLimiter 速率限制中间件
func RateLimiter(maxRequests int, windowSize int64) Handler {
	// 简单的内存速率限制器实现
	// 生产环境建议使用更复杂的限流算法，如令牌桶或滑动窗口
	requestCounts := make(map[string]int)

	return func(ctx *Context) {
		clientAddr := ctx.GetRemoteAddr()

		// 简单计数，实际应该考虑时间窗口
		if requestCounts[clientAddr] >= maxRequests {
			log.Warn(ctx, "rate limit exceeded for client: %s", clientAddr)
			_ = ctx.ServerErr()
			ctx.Abort()
			return
		}

		requestCounts[clientAddr]++
		ctx.Next()
	}
}

// Logger 日志中间件
func Logger() Handler {
	return func(ctx *Context) {
		log.Info(ctx, "QUIC request from %s, OpCode: %d, SQID: %d",
			ctx.GetRemoteAddr(), ctx.OpCode, ctx.SQID)
		ctx.Next()
	}
}

// Recovery 恢复中间件
func Recovery() Handler {
	return func(ctx *Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Error(ctx, "panic recovered: %v", err)
				_ = ctx.ServerErr()
			}
		}()
		ctx.Next()
	}
}
