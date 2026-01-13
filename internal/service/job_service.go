package service

import (
	"encoding/json"
	"fmt"
	"qd-sc/internal/client"
	"qd-sc/internal/config"
	"qd-sc/internal/model"
	"qd-sc/pkg/utils"
)

// JobService 岗位服务
type JobService struct {
	cfg       *config.Config
	jobClient *client.JobClient
}

// NewJobService 创建岗位服务
func NewJobService(cfg *config.Config, jobClient *client.JobClient) *JobService {
	return &JobService{
		cfg:       cfg,
		jobClient: jobClient,
	}
}

// QueryJobsByArea 根据区域代码查询岗位
func (s *JobService) QueryJobsByArea(params map[string]interface{}) (string, error) {
	return s.queryJobs(params)
}

// QueryJobsByLocation 根据经纬度查询岗位
func (s *JobService) QueryJobsByLocation(params map[string]interface{}) (string, error) {
	return s.queryJobs(params)
}

// queryJobs 通用岗位查询方法
func (s *JobService) queryJobs(params map[string]interface{}) (string, error) {
	req := s.buildJobQueryRequest(params)

	apiResp, err := s.jobClient.QueryJobs(req)
	if err != nil {
		return "", fmt.Errorf("查询岗位失败: %w", err)
	}

	if apiResp.Code != 200 {
		errMsg := apiResp.Msg
		if errMsg == "" {
			errMsg = fmt.Sprintf("API返回错误代码: %d", apiResp.Code)
		}
		return "", fmt.Errorf("岗位API返回错误: %s", errMsg)
	}

	if len(apiResp.Rows) == 0 {
		return s.formatEmptyResult(), nil
	}

	// 格式化响应
	formattedResp := s.jobClient.FormatJobResponse(apiResp)
	return s.formatJobResponse(formattedResp)
}

// buildJobQueryRequest 构建岗位查询请求
func (s *JobService) buildJobQueryRequest(params map[string]interface{}) *model.JobQueryRequest {
	req := &model.JobQueryRequest{
		Current:  1,
		PageSize: 10,
	}

	// 解析参数
	if v, ok := params["current"].(float64); ok {
		req.Current = int(v)
	}
	if v, ok := params["pageSize"].(float64); ok {
		req.PageSize = int(v)
	}
	if v, ok := params["jobTitle"].(string); ok {
		req.JobTitle = v
	}
	if v, ok := params["latitude"].(string); ok {
		req.Latitude = v
	}
	if v, ok := params["longitude"].(string); ok {
		req.Longitude = v
	}
	if v, ok := params["radius"].(string); ok {
		req.Radius = v
	}
	if v, ok := params["order"].(string); ok {
		req.Order = v
	}
	if v, ok := params["minSalary"].(string); ok {
		req.MinSalary = v
	}
	if v, ok := params["maxSalary"].(string); ok {
		req.MaxSalary = v
	}
	if v, ok := params["experience"].(string); ok {
		req.Experience = v
	}
	if v, ok := params["education"].(string); ok {
		req.Education = v
	}
	if v, ok := params["companyNature"].(string); ok {
		req.CompanyNature = v
	}
	if v, ok := params["jobLocationAreaCode"].(string); ok {
		req.JobLocationAreaCode = v
	}

	return req
}

// formatJobResponse 格式化岗位响应为JSON字符串
func (s *JobService) formatJobResponse(resp *model.JobResponse) (string, error) {
	// 将响应格式化为JSON
	jsonStr, err := utils.ToJSONStringPretty(resp)
	if err != nil {
		return "", fmt.Errorf("格式化岗位信息失败: %w", err)
	}
	return jsonStr, nil
}

// formatEmptyResult 格式化空结果
func (s *JobService) formatEmptyResult() string {
	emptyResp := &model.JobResponse{
		JobListings: []model.FormattedJob{},
		Data:        nil,
	}
	jsonStr, _ := json.MarshalIndent(emptyResp, "", "  ")
	return string(jsonStr)
}
