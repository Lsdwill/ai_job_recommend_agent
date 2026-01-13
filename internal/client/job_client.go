package client

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"qd-sc/internal/config"
	"qd-sc/internal/model"
	"strconv"
)

// JobClient 岗位API客户端
type JobClient struct {
	baseURL     string
	httpClient  *http.Client
	logLevel    string
	locationMap map[string]string // 区域代码到名称的映射
}

// NewJobClient 创建岗位API客户端
func NewJobClient(cfg *config.Config) *JobClient {
	// 构建区域代码到名称的映射（反转配置中的映射）
	locationMap := make(map[string]string)
	for name, code := range cfg.City.AreaCodes {
		locationMap[code] = name
	}

	return &JobClient{
		baseURL:     cfg.JobAPI.BaseURL,
		httpClient:  NewHTTPClient(HTTPClientConfig{Timeout: cfg.JobAPI.Timeout, MaxIdleConns: 100, MaxIdleConnsPerHost: 50, MaxConnsPerHost: 0}),
		logLevel:    cfg.Logging.Level,
		locationMap: locationMap,
	}
}

// QueryJobs 查询岗位
func (c *JobClient) QueryJobs(req *model.JobQueryRequest) (*model.JobAPIResponse, error) {
	// 构建请求URL
	params := url.Values{}
	params.Set("current", strconv.Itoa(req.Current))
	params.Set("pageSize", strconv.Itoa(req.PageSize))

	if req.JobTitle != "" {
		params.Set("jobTitle", req.JobTitle)
	}
	if req.Latitude != "" {
		params.Set("latitude", req.Latitude)
	}
	if req.Longitude != "" {
		params.Set("longitude", req.Longitude)
	}
	if req.Radius != "" {
		params.Set("radius", req.Radius)
	}
	if req.Order != "" {
		params.Set("order", req.Order)
	}
	if req.MinSalary != "" {
		params.Set("minSalary", req.MinSalary)
	}
	if req.MaxSalary != "" {
		params.Set("maxSalary", req.MaxSalary)
	}
	if req.Experience != "" {
		params.Set("experience", req.Experience)
	}
	if req.Education != "" {
		params.Set("education", req.Education)
	}
	if req.CompanyNature != "" {
		params.Set("companyNature", req.CompanyNature)
	}
	if req.JobLocationAreaCode != "" {
		params.Set("jobLocationAreaCode", req.JobLocationAreaCode)
	}

	reqURL := fmt.Sprintf("%s?%s", c.baseURL, params.Encode())

	httpReq, err := http.NewRequest("GET", reqURL, nil)
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

	// 读取响应体用于日志和解析
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 打印原始响应（仅在debug级别时显示完整响应，否则显示摘要）
	if c.logLevel == "debug" {
		log.Printf("岗位API原始响应: %s", string(body))
	} else {
		// 非debug模式下显示响应摘要
		if len(body) > 200 {
			log.Printf("岗位API响应摘要: %s... (共%d字节)", string(body[:200]), len(body))
		} else {
			log.Printf("岗位API响应: %s", string(body))
		}
	}

	var result model.JobAPIResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w, 原始响应: %s", err, string(body))
	}

	log.Printf("岗位API解析结果: Code=%d, Msg=%s, Rows数量=%d", result.Code, result.Msg, len(result.Rows))

	return &result, nil
}

// FormatJobResponse 格式化岗位响应
func (c *JobClient) FormatJobResponse(apiResp *model.JobAPIResponse) *model.JobResponse {
	formattedJobs := make([]model.FormattedJob, 0, len(apiResp.Rows))

	for _, job := range apiResp.Rows {
		// 格式化薪资
		salary := "薪资面议"
		if job.MinSalary > 0 || job.MaxSalary > 0 {
			salary = fmt.Sprintf("%d-%d元/月", job.MinSalary, job.MaxSalary)
		}

		// 转换学历代码
		education := model.EducationMap[job.Education]
		if education == "" {
			education = "学历不限"
		}

		// 转换经验代码
		experience := model.ExperienceMap[job.Experience]
		if experience == "" {
			experience = "经验不限"
		}

		// 转换区域代码
		location := c.locationMap[strconv.Itoa(job.JobLocationAreaCode)]
		if location == "" {
			location = "未知地区"
		}

		formattedJobs = append(formattedJobs, model.FormattedJob{
			JobTitle:    job.JobTitle,
			CompanyName: job.CompanyName,
			Salary:      salary,
			Location:    location,
			Education:   education,
			Experience:  experience,
			AppJobURL:   job.AppJobURL,
		})
	}

	// 如果有data字段，在最后一条job中添加
	if len(formattedJobs) > 0 && apiResp.Data != nil {
		formattedJobs[len(formattedJobs)-1].Data = apiResp.Data
	}

	return &model.JobResponse{
		JobListings: formattedJobs,
		Data:        apiResp.Data,
	}
}
