package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"qd-sc/internal/config"
	"qd-sc/internal/model"
	"strings"
)

// AmapClient 高德地图客户端
type AmapClient struct {
	baseURL    string
	apiKey     string
	cityName   string
	httpClient *http.Client
}

// NewAmapClient 创建高德地图客户端
func NewAmapClient(cfg *config.Config) *AmapClient {
	return &AmapClient{
		baseURL:    cfg.Amap.BaseURL,
		apiKey:     cfg.Amap.APIKey,
		cityName:   cfg.City.Name,
		httpClient: NewHTTPClient(HTTPClientConfig{Timeout: cfg.Amap.Timeout, MaxIdleConns: 100, MaxIdleConnsPerHost: 50, MaxConnsPerHost: 0}),
	}
}

// SearchPlace 搜索地点，返回经纬度
func (c *AmapClient) SearchPlace(keywords string) (*model.AmapPlaceResponse, error) {
	// 构建请求URL
	params := url.Values{}
	params.Set("key", c.apiKey)
	params.Set("keywords", keywords)
	params.Set("types", "190000") // 地名地址信息类型
	params.Set("city", c.cityName)
	params.Set("output", "JSON")

	reqURL := fmt.Sprintf("%s/place/text?%s", c.baseURL, params.Encode())

	resp, err := c.httpClient.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API返回错误状态码 %d: %s", resp.StatusCode, string(body))
	}

	var result model.AmapPlaceResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if result.Status != "1" {
		return nil, fmt.Errorf("高德API返回错误: %s", result.Info)
	}

	return &result, nil
}

// GetLocationCoordinates 获取地点的经纬度坐标
func (c *AmapClient) GetLocationCoordinates(keywords string) (latitude, longitude string, err error) {
	result, err := c.SearchPlace(keywords)
	if err != nil {
		return "", "", err
	}

	if len(result.Pois) == 0 {
		return "", "", fmt.Errorf("未找到地点: %s", keywords)
	}

	// 取第一个结果
	location := result.Pois[0].Location
	parts := strings.Split(location, ",")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("解析坐标失败: %s", location)
	}

	// 高德返回的格式是"经度,纬度"
	longitude = parts[0]
	latitude = parts[1]

	return latitude, longitude, nil
}
