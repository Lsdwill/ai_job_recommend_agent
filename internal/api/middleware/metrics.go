package middleware

import (
	"qd-sc/pkg/metrics"
	"time"

	"github.com/gin-gonic/gin"
)

// Metrics 指标收集中间件
func Metrics() gin.HandlerFunc {
	m := metrics.GetGlobalMetrics()

	return func(c *gin.Context) {
		// 记录请求开始时间
		start := time.Now()

		// 增加总请求数和活跃请求数
		m.IncTotalRequests()
		m.IncActiveRequests()
		defer m.DecActiveRequests()

		// 检查是否是流式请求
		if c.GetHeader("Accept") == "text/event-stream" || c.Query("stream") == "true" {
			m.IncStreamRequests()
		}

		// 处理请求
		c.Next()

		// 记录延迟
		duration := time.Since(start)
		endpoint := c.Request.Method + " " + c.FullPath()
		m.RecordLatency(endpoint, duration)

		// 如果请求失败，增加失败计数
		if c.Writer.Status() >= 400 {
			m.IncFailedRequests()
		}
	}
}
