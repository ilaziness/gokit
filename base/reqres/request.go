// Package reqres 包含了用于请求处理调用服务的工具函数。

package reqres

import (
	"context"

	"github.com/gin-gonic/gin"
)

// 提供快捷调用service的方法

// ServiceMethod service方法，方法签名入参ctx，req，出参时返回值P和错误error
type ServiceMethod[R any, P any] func(ctx context.Context, req R) (P, error)

// ServiceMethodNoReq service方法，方法签名入参ctx，出参时返回值P和错误error
type ServiceMethodNoReq[P any] func(ctx context.Context) (P, error)

// ServiceMethodNoRes service方法，方法签名入参ctx，req，出参时返回值error
type ServiceMethodNoRes[R any] func(ctx context.Context, req R) error

// ServiceMethodNoReqRes service方法，方法签名入参ctx，出参时返回值error
type ServiceMethodNoReqRes func(ctx context.Context) error

type bindType string

var BindTypeDefault bindType = "default"
var BindTypeURI bindType = "uri"
var BindTypeQuery bindType = "query"

// ginBind 解析请求参数
func ginBind[R any](g *gin.Context, req R, bt ...bindType) (err error) {
	btl := len(bt)
	var cbt bindType
	if btl == 1 {
		cbt = bt[0]
	}
	if btl > 1 {
		for k, t := range bt {
			if (btl - 1) == k {
				cbt = t
				break
			}
			switch t {
			case BindTypeURI:
				err = g.ShouldBindUri(&req)
			case BindTypeQuery:
				err = g.ShouldBindQuery(&req)
			default:
				err = g.ShouldBind(&req)
			}
		}
	}
	if err != nil {
		return
	}
	switch cbt {
	case BindTypeURI:
		err = g.ShouldBindUri(&req)
	case BindTypeQuery:
		err = g.ShouldBindQuery(&req)
	default:
		err = g.ShouldBind(&req)
	}
	return
}

// CallService 是一个用于调用需要请求体的服务方法的函数。
func CallService[R any, P any](g *gin.Context, req R, sh ServiceMethod[*R, *P], bt ...bindType) {
	if err := ginBind(g, &req, bt...); err != nil {
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

// CallServiceNoReq 是一个用于调用不需要请求体的服务方法的函数。
func CallServiceNoReq[P any](g *gin.Context, sh ServiceMethodNoReq[P]) {
	res, err := sh(g)
	if err != nil {
		Error(g, err)
		return
	}
	Success(g, res)
}

// CallServiceNoRes 是一个用于调用不需要响应体的服务方法的函数。
func CallServiceNoRes[R any](g *gin.Context, req R, sh ServiceMethodNoRes[*R], bt ...bindType) {
	if err := ginBind(g, &req, bt...); err != nil {
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

// CallServiceNoReqRes 是一个用于调用不需要请求体和响应体的服务方法的函数。
func CallServiceNoReqRes(g *gin.Context, sh ServiceMethodNoReqRes) {
	err := sh(g)
	if err != nil {
		Error(g, err)
		return
	}
	Success(g, nil)
}
