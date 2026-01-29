package handler

import (
	"context"
	"net/http"
	"qd-sc/internal/service"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// PolicyHandler 政策处理器
type PolicyHandler struct {
	policyService *service.PolicyService
	response      *Response
}

// NewPolicyHandler 创建政策处理器
func NewPolicyHandler(policyService *service.PolicyService) *PolicyHandler {
	return &PolicyHandler{
		policyService: policyService,
		response:      NewResponse(),
	}
}

// UpdatePolicies 更新政策到向量数据库
// @Summary 更新政策
// @Description 从政策API获取最新政策并更新到向量数据库
// @Tags 政策
// @Accept json
// @Produce json
// @Success 200 {object} Response
// @Failure 500 {object} Response
// @Router /api/policy/update [post]
func (h *PolicyHandler) UpdatePolicies(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if err := h.policyService.UpdatePolicies(ctx); err != nil {
		h.response.Error(c, http.StatusInternalServerError, "update_failed", err.Error())
		return
	}

	h.response.Success(c, gin.H{"message": "政策更新成功"})
}

// SearchPolicies 搜索政策
// @Summary 搜索政策
// @Description 根据查询文本搜索相关政策
// @Tags 政策
// @Accept json
// @Produce json
// @Param query query string true "查询文本"
// @Param topK query int false "返回结果数量" default(5)
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 500 {object} Response
// @Router /api/policy/search [get]
func (h *PolicyHandler) SearchPolicies(c *gin.Context) {
	query := c.Query("query")
	if query == "" {
		h.response.Error(c, http.StatusBadRequest, "invalid_request", "查询文本不能为空")
		return
	}

	topK := 5
	if topKStr := c.Query("topK"); topKStr != "" {
		if k, err := strconv.Atoi(topKStr); err == nil && k > 0 {
			topK = k
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	results, err := h.policyService.SearchPolicies(ctx, query, topK)
	if err != nil {
		h.response.Error(c, http.StatusInternalServerError, "search_failed", err.Error())
		return
	}

	h.response.Success(c, gin.H{
		"query":   query,
		"results": results,
	})
}
