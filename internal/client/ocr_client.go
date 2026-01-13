package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"qd-sc/internal/config"
)

// OCRClient OCR服务客户端
type OCRClient struct {
	baseURL    string
	httpClient *http.Client
	logLevel   string
}

// OCRResponse OCR服务响应结构
type OCRResponse struct {
	Code       int     `json:"code"`
	Data       string  `json:"data"`
	CostTimeMs float64 `json:"cost_time_ms"`
	Msg        string  `json:"msg,omitempty"`
}

// NewOCRClient 创建OCR客户端
func NewOCRClient(cfg *config.Config) *OCRClient {
	return &OCRClient{
		baseURL:    cfg.OCR.BaseURL,
		httpClient: NewHTTPClient(HTTPClientConfig{Timeout: cfg.OCR.Timeout, MaxIdleConns: 100, MaxIdleConnsPerHost: 50, MaxConnsPerHost: 0}),
		logLevel:   cfg.Logging.Level,
	}
}

// ParseURL 通过URL解析远程文件内容
// 支持图片、PDF、Excel、PPT等格式
func (c *OCRClient) ParseURL(fileURL string) (string, error) {
	// 构建请求体
	reqBody := map[string]string{
		"url": fileURL,
	}

	reqData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %w", err)
	}

	// 打印文件解析请求体
	log.Printf("OCR文件解析请求: URL=%s, 请求体=%s", c.baseURL+"/ocr/url", string(reqData))

	// 创建HTTP请求
	httpReq, err := http.NewRequest("POST", c.baseURL+"/ocr/url", bytes.NewReader(reqData))
	if err != nil {
		return "", fmt.Errorf("创建HTTP请求失败: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	// 打印原始响应（仅在debug级别时显示完整响应，否则显示摘要）
	if c.logLevel == "debug" {
		log.Printf("OCR原始响应: %s", string(respBody))
	} else {
		// 非debug模式下显示响应摘要
		if len(respBody) > 200 {
			log.Printf("OCR响应摘要: %s... (共%d字节)", string(respBody[:200]), len(respBody))
		} else {
			log.Printf("OCR响应: %s", string(respBody))
		}
	}

	// 解析响应
	var ocrResp OCRResponse
	if err := json.Unmarshal(respBody, &ocrResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	// 打印解析结果
	log.Printf("OCR解析结果: Code=%d, CostTimeMs=%.2f, 数据长度=%d字节", ocrResp.Code, ocrResp.CostTimeMs, len(ocrResp.Data))
	if c.logLevel == "debug" && ocrResp.Data != "" {
		// debug模式下打印解析出的数据内容
		if len(ocrResp.Data) > 500 {
			log.Printf("OCR解析数据(前500字符): %s...", ocrResp.Data[:500])
		} else {
			log.Printf("OCR解析数据: %s", ocrResp.Data)
		}
	}

	// 检查业务状态码
	if ocrResp.Code != 200 {
		errMsg := ocrResp.Msg
		if errMsg == "" {
			errMsg = "OCR解析失败"
		}
		return "", fmt.Errorf("OCR服务返回错误: %s", errMsg)
	}

	return ocrResp.Data, nil
}
