package handler

import (
	"qd-sc/internal/model"

	"github.com/gin-gonic/gin"
)

// Response 统一响应处理器
type Response struct{}

// Error 发送错误响应
func (r *Response) Error(c *gin.Context, statusCode int, errorType, message string) {
	c.JSON(statusCode, model.ErrorResponse{
		Error: model.ErrorDetail{
			Message: message,
			Type:    errorType,
		},
	})
}

// Success 发送成功响应
func (r *Response) Success(c *gin.Context, data interface{}) {
	c.JSON(200, data)
}

// NewResponse 创建响应处理器
func NewResponse() *Response {
	return &Response{}
}

// 全局响应处理器实例
var DefaultResponse = NewResponse()
