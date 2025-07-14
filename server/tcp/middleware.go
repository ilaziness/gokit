package tcp

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
