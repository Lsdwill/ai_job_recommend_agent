package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"qd-sc/internal/config"
	"qd-sc/internal/model"
	"time"
)

// EmbeddingClient Embedding客户端
type EmbeddingClient struct {
	baseURL string
	client  *http.Client
}

// NewEmbeddingClient 创建Embedding客户端
func NewEmbeddingClient(cfg *config.EmbeddingConfig) *EmbeddingClient {
	return &EmbeddingClient{
		baseURL: cfg.BaseURL,
		client: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

// GetEmbedding 获取文本的向量表示
func (c *EmbeddingClient) GetEmbedding(text string) ([]float32, error) {
	reqBody := model.EmbeddingRequest{
		Inputs: text,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	var embResp model.EmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if len(embResp) == 0 || len(embResp[0]) == 0 {
		return nil, fmt.Errorf("返回的向量为空")
	}

	return embResp[0], nil
}

// GetEmbeddingWithRetry 带重试的获取向量
func (c *EmbeddingClient) GetEmbeddingWithRetry(text string, maxRetries int) ([]float32, error) {
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		embedding, err := c.GetEmbedding(text)
		if err == nil {
			return embedding, nil
		}
		lastErr = err
		if i < maxRetries-1 {
			time.Sleep(time.Second * time.Duration(i+1))
		}
	}
	return nil, fmt.Errorf("重试%d次后仍失败: %w", maxRetries, lastErr)
}
