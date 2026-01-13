package client

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"qd-sc/internal/config"
	"qd-sc/internal/model"
	"strings"
	"time"
)

// LLMClient LLM客户端
type LLMClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	maxRetries int
}

// NewLLMClient 创建LLM客户端
func NewLLMClient(cfg *config.Config) *LLMClient {
	return &LLMClient{
		baseURL:    cfg.LLM.BaseURL,
		apiKey:     cfg.LLM.APIKey,
		httpClient: NewLLMHTTPClient(cfg.LLM.Timeout),
		maxRetries: cfg.LLM.MaxRetries,
	}
}

// ChatCompletion 发起聊天补全请求（非流式）
func (c *LLMClient) ChatCompletion(req *model.ChatCompletionRequest) (*model.ChatCompletionResponse, error) {
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	var resp *http.Response
	var lastErr error

	// 重试机制
	for i := 0; i < c.maxRetries; i++ {
		// 注意：http.Request 的 Body 在 Do() 后会被读取消耗，重试必须重建 request/body
		httpReq, err := http.NewRequest("POST", c.baseURL+"/chat/completions", bytes.NewReader(reqBody))
		if err != nil {
			return nil, fmt.Errorf("创建HTTP请求失败: %w", err)
		}

		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

		resp, lastErr = c.httpClient.Do(httpReq)

		// 请求未发出/网络错误：可重试
		if lastErr != nil {
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}

		// 根据状态码决定是否重试：5xx 和 429 重试；其他 4xx 直接返回
		if resp.StatusCode < 500 && resp.StatusCode != http.StatusTooManyRequests {
			break
		}

		resp.Body.Close()
		time.Sleep(time.Duration(i+1) * time.Second)
	}

	if lastErr != nil {
		return nil, fmt.Errorf("HTTP请求失败: %w", lastErr)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API返回错误状态码 %d: %s", resp.StatusCode, string(body))
	}

	var result model.ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &result, nil
}

// ChatCompletionStream 发起聊天补全请求（流式）
func (c *LLMClient) ChatCompletionStream(req *model.ChatCompletionRequest) (chan *model.ChatCompletionChunk, chan error, error) {
	req.Stream = true
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.baseURL+"/chat/completions", bytes.NewReader(reqBody))
	if err != nil {
		return nil, nil, fmt.Errorf("创建HTTP请求失败: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, nil, fmt.Errorf("HTTP请求失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, nil, fmt.Errorf("API返回错误状态码 %d: %s", resp.StatusCode, string(body))
	}

	chunkChan := make(chan *model.ChatCompletionChunk, 100)
	errChan := make(chan error, 1)

	go func() {
		defer resp.Body.Close()
		defer close(chunkChan)
		defer close(errChan)

		scanner := bufio.NewScanner(resp.Body)
		// 设置最大buffer大小，防止被超大响应攻击（默认64KB，增加到1MB）
		scanner.Buffer(make([]byte, 64*1024), 1024*1024)

		for scanner.Scan() {
			line := scanner.Text()
			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				return
			}

			var chunk model.ChatCompletionChunk
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				errChan <- fmt.Errorf("解析流式响应失败: %w", err)
				return
			}

			chunkChan <- &chunk
		}

		if err := scanner.Err(); err != nil {
			errChan <- fmt.Errorf("读取流失败: %w", err)
		}
	}()

	return chunkChan, errChan, nil
}
