package reqp

import (
	"context"

	"github.com/gin-gonic/gin"
)

// 提供快捷调用service的方法

// ServiceMethod service方法，方法签名入参ctx，req，出参时返回值P和错误error
type ServiceMethod[R any, P any] func(ctx context.Context, req R) (P, error)

// ServiceMethodWithoutReq service方法，方法签名入参ctx，出参时返回值P和错误error
type ServiceMethodWithoutReq[P any] func(ctx context.Context) (P, error)

// ServiceMethodWithoutRes service方法，方法签名入参ctx，req，出参时返回值error
type ServiceMethodWithoutRes[R any] func(ctx context.Context, req R) error

// ServiceMethodWithoutReqAndRes service方法，方法签名入参ctx，出参时返回值error
type ServiceMethodWithoutReqAndRes func(ctx context.Context) error

// CallService 是一个用于调用需要请求体的服务方法的函数。
func CallService[R any, P any](g *gin.Context, req R, sh ServiceMethod[*R, *P]) {
	if err := g.ShouldBind(&req); err != nil {
		Error(g, err)
		return
	}
	res, err := sh(g, &req)
	if err != nil {
		Error(g, err)
		return
	}
	Success(g, res)
}

// CallServiceWithoutReq 是一个用于调用不需要请求体的服务方法的函数。
func CallServiceWithoutReq[P any](g *gin.Context, sh ServiceMethodWithoutReq[P]) {
	res, err := sh(g)
	if err != nil {
		Error(g, err)
		return
	}
	Success(g, res)
}

// CallServiceWithoutRes 是一个用于调用不需要响应体的服务方法的函数。
func CallServiceWithoutRes[R any](g *gin.Context, req R, sh ServiceMethodWithoutRes[*R]) {
	if err := g.ShouldBind(&req); err != nil {
		Error(g, err)
		return
	}
	err := sh(g, &req)
	if err != nil {
		Error(g, err)
		return
	}
	Success(g, nil)
}

// CallServiceWithoutReqAndRes 是一个用于调用不需要请求体和响应体的服务方法的函数。
func CallServiceWithoutReqAndRes(g *gin.Context, sh ServiceMethodWithoutReqAndRes) {
	err := sh(g)
	if err != nil {
		Error(g, err)
		return
	}
	Success(g, nil)
}
