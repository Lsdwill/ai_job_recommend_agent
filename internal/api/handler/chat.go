package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"qd-sc/internal/model"
	"qd-sc/internal/service"
	"time"

	"github.com/gin-gonic/gin"
)

// 使用服务层定义的固定模型名称

// ChatHandler 聊天处理器
type ChatHandler struct {
	chatService *service.ChatService
	response    *Response
}

// NewChatHandler 创建聊天处理器
func NewChatHandler(chatService *service.ChatService) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
		response:    DefaultResponse,
	}
}

// ChatCompletions 处理聊天completions请求
func (h *ChatHandler) ChatCompletions(c *gin.Context) {
	var req model.ChatCompletionRequest

	// 只支持 JSON 请求（文件通过 image_url 字段以 URL 方式传递）
	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.Error(c, http.StatusBadRequest, "invalid_request", "无效的请求格式: "+err.Error())
		return
	}

	// 验证请求
	if req.Model == "" {
		h.response.Error(c, http.StatusBadRequest, "invalid_request", "缺少model参数")
		return
	}
	// 验证模型名称（只接受固定模型名）
	if req.Model != service.ExposedModelName {
		h.response.Error(c, http.StatusBadRequest, "invalid_request", fmt.Sprintf("不支持的模型，请使用: %s", service.ExposedModelName))
		return
	}
	if len(req.Messages) == 0 {
		h.response.Error(c, http.StatusBadRequest, "invalid_request", "messages不能为空")
		return
	}

	// 根据stream参数决定返回方式
	if req.Stream {
		h.handleStreamResponse(c, &req)
	} else {
		h.handleNonStreamResponse(c, &req)
	}
}

// handleNonStreamResponse 处理非流式响应
func (h *ChatHandler) handleNonStreamResponse(c *gin.Context, req *model.ChatCompletionRequest) {
	resp, err := h.chatService.ProcessChatRequest(req)
	if err != nil {
		log.Printf("处理聊天请求失败: %v", err)
		h.response.Error(c, http.StatusInternalServerError, "internal_error", "处理请求失败: "+err.Error())
		return
	}

	h.response.Success(c, resp)
}

// handleStreamResponse 处理流式响应
func (h *ChatHandler) handleStreamResponse(c *gin.Context, req *model.ChatCompletionRequest) {
	// 设置SSE响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Transfer-Encoding", "chunked")

	// 传递context以支持取消
	ctx := c.Request.Context()
	chunkChan, errChan := h.chatService.ProcessChatRequestStream(ctx, req)

	// 持续发送流式数据
	for {
		select {
		case chunk, ok := <-chunkChan:
			if !ok {
				// 通道已关闭，发送[DONE]标记（OpenAI标准格式）
				if _, err := fmt.Fprintf(c.Writer, "data: [DONE]\n\n"); err != nil {
					log.Printf("写入[DONE]标记失败: %v", err)
				}
				c.Writer.Flush()
				log.Printf("SSE流已结束，已发送[DONE]标记")
				return
			}

			// 发送chunk
			chunkJSON, err := json.Marshal(chunk)
			if err != nil {
				log.Printf("序列化chunk失败: %v", err)
				continue
			}

			// 写入SSE格式
			if _, err := fmt.Fprintf(c.Writer, "data: %s\n\n", string(chunkJSON)); err != nil {
				log.Printf("写入SSE数据失败: %v", err)
				return
			}
			c.Writer.Flush()

			// 如果这个chunk包含finish_reason，记录日志
			if len(chunk.Choices) > 0 && chunk.Choices[0].FinishReason != "" {
				log.Printf("已发送finish_reason=%s的chunk", chunk.Choices[0].FinishReason)
			}

		case err, ok := <-errChan:
			if ok && err != nil {
				log.Printf("流式处理错误: %v", err)
				// 发送错误信息
				errChunk := model.ChatCompletionChunk{
					ID:      fmt.Sprintf("chatcmpl-%d", time.Now().Unix()),
					Object:  "chat.completion.chunk",
					Created: time.Now().Unix(),
					Model:   service.ExposedModelName,
					Choices: []model.ChunkChoice{
						{
							Index: 0,
							Delta: model.Message{
								Role:    "assistant",
								Content: fmt.Sprintf("\n\n错误：%s", err.Error()),
							},
							FinishReason: "error",
						},
					},
				}
				chunkJSON, _ := json.Marshal(errChunk)
				fmt.Fprintf(c.Writer, "data: %s\n\n", string(chunkJSON))
				c.Writer.Flush()

				// 发送DONE
				fmt.Fprintf(c.Writer, "data: [DONE]\n\n")
				c.Writer.Flush()
				return
			}

		case <-c.Request.Context().Done():
			// 客户端断开连接
			log.Printf("客户端断开连接")
			return
		}
	}
}
