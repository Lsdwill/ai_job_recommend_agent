package service

import (
	"fmt"
	"qd-sc/internal/client"
	"qd-sc/internal/config"
)

// LocationService 地理位置服务
type LocationService struct {
	cfg        *config.Config
	amapClient *client.AmapClient
}

// NewLocationService 创建位置服务
func NewLocationService(cfg *config.Config, amapClient *client.AmapClient) *LocationService {
	return &LocationService{
		cfg:        cfg,
		amapClient: amapClient,
	}
}

// QueryLocation 查询地点经纬度
func (s *LocationService) QueryLocation(keywords string) (latitude, longitude string, err error) {
	lat, lng, err := s.amapClient.GetLocationCoordinates(keywords)
	if err != nil {
		return "", "", fmt.Errorf("查询地点失败: %w", err)
	}
	return lat, lng, nil
}
