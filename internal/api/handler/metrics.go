package handler

import (
	"net/http"
	"qd-sc/pkg/metrics"

	"github.com/gin-gonic/gin"
)

// MetricsHandler 指标处理器
type MetricsHandler struct {
	metrics *metrics.Metrics
}

// NewMetricsHandler 创建指标处理器
func NewMetricsHandler() *MetricsHandler {
	return &MetricsHandler{
		metrics: metrics.GetGlobalMetrics(),
	}
}

// GetMetrics 获取性能指标
func (h *MetricsHandler) GetMetrics(c *gin.Context) {
	stats := h.metrics.GetStats()
	c.JSON(http.StatusOK, stats)
}
