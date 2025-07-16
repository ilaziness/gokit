package udp

// Ping ping中间件
func Ping(ctx *Context) {
	if ctx.OpCode == OpCodePing {
		_ = ctx.WriteWithOpCode(OpCodePong, ctx.Payload)
		ctx.Abort()
		return
	}
	ctx.Next()
}
