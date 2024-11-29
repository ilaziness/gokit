package response

import (
	"net/http"

	"github.com/ilaziness/gokit/base/errcode"

	"github.com/gin-gonic/gin"
)

const successCode = 0

type Format struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

func Success(ctx *gin.Context, data any) {
	respData := Format{
		Code:    successCode,
		Message: "",
		Data:    data,
	}
	ctx.JSON(http.StatusOK, respData)
}

func Error(ctx *gin.Context, err *errcode.Code) {
	respData := Format{
		Code:    err.Code,
		Message: err.Message,
		Data:    err.Data,
	}
	ctx.JSON(http.StatusOK, respData)
}
