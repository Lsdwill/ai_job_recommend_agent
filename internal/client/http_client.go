package client

import (
	"net/http"
	"time"
)

// HTTPClientConfig HTTP客户端配置
type HTTPClientConfig struct {
	Timeout             time.Duration
	MaxIdleConns        int
	MaxIdleConnsPerHost int
	MaxConnsPerHost     int
}

// DefaultHTTPClientConfig 默认HTTP客户端配置
var DefaultHTTPClientConfig = HTTPClientConfig{
	Timeout:             30 * time.Second,
	MaxIdleConns:        100,
	MaxIdleConnsPerHost: 50,
	MaxConnsPerHost:     0, // 0表示不限制
}

// NewHTTPClient 创建优化的HTTP客户端
func NewHTTPClient(config HTTPClientConfig) *http.Client {
	transport := &http.Transport{
		// 连接池配置
		MaxIdleConns:        config.MaxIdleConns,
		MaxIdleConnsPerHost: config.MaxIdleConnsPerHost,
		MaxConnsPerHost:     config.MaxConnsPerHost,
		IdleConnTimeout:     90 * time.Second,

		// 性能优化
		DisableCompression: false,
		DisableKeepAlives:  false,
		ForceAttemptHTTP2:  true, // 启用HTTP/2

		// 超时配置
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,

		// 缓冲大小
		WriteBufferSize: 32 * 1024, // 32KB写缓冲
		ReadBufferSize:  32 * 1024, // 32KB读缓冲
	}

	return &http.Client{
		Timeout:   config.Timeout,
		Transport: transport,
	}
}

// NewLLMHTTPClient 创建用于LLM的HTTP客户端（连接池更大）
func NewLLMHTTPClient(timeout time.Duration) *http.Client {
	config := HTTPClientConfig{
		Timeout:             timeout,
		MaxIdleConns:        200,
		MaxIdleConnsPerHost: 100,
		MaxConnsPerHost:     0,
	}
	return NewHTTPClient(config)
}
