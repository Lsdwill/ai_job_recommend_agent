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
	"sync"
	"time"
)

// PolicyClient 政策大模型客户端
type PolicyClient struct {
	baseURL    string
	loginName  string
	userKey    string
	serviceID  string
	httpClient *http.Client

	// ticket管理
	mu              sync.RWMutex
	currentTicket   *model.PolicyTicketData
	ticketExpiresAt time.Time
}

// NewPolicyClient 创建政策大模型客户端
func NewPolicyClient(cfg *config.Config) *PolicyClient {
	return &PolicyClient{
		baseURL:   cfg.Policy.BaseURL,
		loginName: "", // 旧版API字段，已废弃
		userKey:   "", // 旧版API字段，已废弃
		serviceID: "", // 旧版API字段，已废弃
		httpClient: NewHTTPClient(HTTPClientConfig{
			Timeout:             cfg.Policy.Timeout,
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 50,
			MaxConnsPerHost:     0,
		}),
	}
}

// GetTicket 获取ticket（会自动缓存和刷新）
func (c *PolicyClient) GetTicket() (*model.PolicyTicketData, error) {
	c.mu.RLock()
	// 如果ticket存在且未过期（提前5分钟刷新）
	if c.currentTicket != nil && time.Now().Before(c.ticketExpiresAt.Add(-5*time.Minute)) {
		ticket := c.currentTicket
		c.mu.RUnlock()
		return ticket, nil
	}
	c.mu.RUnlock()

	// 需要获取新ticket
	c.mu.Lock()
	defer c.mu.Unlock()

	// 双重检查，防止并发重复请求
	if c.currentTicket != nil && time.Now().Before(c.ticketExpiresAt.Add(-5*time.Minute)) {
		return c.currentTicket, nil
	}

	// 发起请求获取ticket
	req := &model.PolicyTicketRequest{
		LoginName: c.loginName,
		UserKey:   c.userKey,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	url := fmt.Sprintf("%s/api/aiServer/getAccessUserInfo", c.baseURL)
	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API返回错误状态码 %d: %s", resp.StatusCode, string(body))
	}

	var result model.PolicyTicketResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if result.Code != 200 {
		return nil, fmt.Errorf("获取ticket失败: %s", result.Message)
	}

	if result.Data == nil {
		return nil, fmt.Errorf("返回数据为空")
	}

	// 缓存ticket，设置过期时间为1小时后
	c.currentTicket = result.Data
	c.ticketExpiresAt = time.Now().Add(1 * time.Hour)

	return c.currentTicket, nil
}

// Chat 发起政策咨询对话（非流式）
func (c *PolicyClient) Chat(chatReq *model.PolicyChatData) (*model.PolicyChatResponse, error) {
	// 获取ticket
	ticketData, err := c.GetTicket()
	if err != nil {
		return nil, fmt.Errorf("获取ticket失败: %w", err)
	}

	// 构造请求
	req := &model.PolicyChatRequest{
		AppID:  ticketData.AppID,
		Ticket: ticketData.Ticket,
		Data:   chatReq,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	url := fmt.Sprintf("%s/api/aiServer/aichat/stream-ai/%s", c.baseURL, c.serviceID)
	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API返回错误状态码 %d: %s", resp.StatusCode, string(body))
	}

	var result model.PolicyChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if result.Code != 200 {
		return nil, fmt.Errorf("对话请求失败: %s", result.Message)
	}

	return &result, nil
}

// ChatStream 发起政策咨询对话（流式）
func (c *PolicyClient) ChatStream(chatReq *model.PolicyChatData) (chan string, chan error, error) {
	// 获取ticket
	ticketData, err := c.GetTicket()
	if err != nil {
		return nil, nil, fmt.Errorf("获取ticket失败: %w", err)
	}

	// 设置流式模式
	chatReq.Stream = true

	// 构造请求
	req := &model.PolicyChatRequest{
		AppID:  ticketData.AppID,
		Ticket: ticketData.Ticket,
		Data:   chatReq,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	url := fmt.Sprintf("%s/api/aiServer/aichat/stream-ai/%s", c.baseURL, c.serviceID)
	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, nil, fmt.Errorf("创建HTTP请求失败: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
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

	contentChan := make(chan string, 100)
	errChan := make(chan error, 1)

	go func() {
		defer resp.Body.Close()
		defer close(contentChan)
		defer close(errChan)

		scanner := bufio.NewScanner(resp.Body)
		scanner.Buffer(make([]byte, 64*1024), 1024*1024)

		for scanner.Scan() {
			line := scanner.Text()

			// 跳过空行
			if strings.TrimSpace(line) == "" {
				continue
			}

			// 尝试解析为JSON响应
			var chunkResp model.PolicyChatResponse
			if err := json.Unmarshal([]byte(line), &chunkResp); err != nil {
				// 如果不是JSON格式，可能是纯文本流，直接发送
				contentChan <- line
				continue
			}

			// 检查响应码
			if chunkResp.Code != 200 {
				errChan <- fmt.Errorf("对话请求失败: %s", chunkResp.Message)
				return
			}

			// 发送消息内容
			if chunkResp.Data != nil && chunkResp.Data.Message != "" {
				contentChan <- chunkResp.Data.Message
			}
		}

		if err := scanner.Err(); err != nil {
			errChan <- fmt.Errorf("读取流失败: %w", err)
		}
	}()

	return contentChan, errChan, nil
}
