package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/ilaziness/gokit/log"
)

func Test() gin.HandlerFunc {
	return func(_ *gin.Context) {
		log.Logger.Info("test middleware")
	}
}
