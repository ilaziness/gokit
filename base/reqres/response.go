// Package reqres 提供了响应数据的一些功能

package reqres

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ilaziness/gokit/base/errcode"
)

const successCode = 0
const failCode = 1

type Format struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

func Success(ctx *gin.Context, data any) {
	if data == nil {
		data = gin.H{}
	}
	respData := Format{
		Code:    successCode,
		Message: "",
		Data:    data,
	}
	ctx.JSON(http.StatusOK, respData)
}

func Error(ctx *gin.Context, err error) {
	var ec *errcode.Code
	if errors.As(err, &ec) {
		respData := Format{
			Code:    ec.Code,
			Message: ec.Message,
			Data:    ec.Data,
		}
		ctx.JSON(http.StatusOK, respData)
		return
	}
	respData := Format{
		Code:    failCode,
		Message: err.Error(),
		Data:    gin.H{},
	}
	ctx.JSON(http.StatusOK, respData)
}
