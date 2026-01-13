package middleware

import (
	"log"
	"net/http"
	"qd-sc/internal/model"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

// Recovery Panic恢复中间件
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 打印堆栈信息
				stack := debug.Stack()
				log.Printf("[PANIC] %v\n%s", err, string(stack))

				// 返回500错误
				c.JSON(http.StatusInternalServerError, model.ErrorResponse{
					Error: model.ErrorDetail{
						Message: "服务器内部错误",
						Type:    "internal_server_error",
					},
				})

				c.Abort()
			}
		}()

		c.Next()
	}
}
