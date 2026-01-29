package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"qd-sc/internal/client"
	"qd-sc/internal/config"
	"qd-sc/internal/model"
	"regexp"
	"strings"
	"time"
)

// PolicyService 政策服务
type PolicyService struct {
	policyClient    *http.Client
	embeddingClient *client.EmbeddingClient
	milvusClient    *client.MilvusClient
	policyURL       string
}

// NewPolicyService 创建政策服务
func NewPolicyService(cfg *config.Config) (*PolicyService, error) {
	embClient := client.NewEmbeddingClient(&cfg.Embedding)

	milvusClient, err := client.NewMilvusClient(&cfg.Milvus)
	if err != nil {
		return nil, fmt.Errorf("创建Milvus客户端失败: %w", err)
	}

	return &PolicyService{
		policyClient: &http.Client{
			Timeout: cfg.Policy.Timeout,
		},
		embeddingClient: embClient,
		milvusClient:    milvusClient,
		policyURL:       cfg.Policy.BaseURL,
	}, nil
}

// FetchPolicies 从API获取政策列表
func (s *PolicyService) FetchPolicies() ([]model.PolicyInfo, error) {
	resp, err := s.policyClient.Get(s.policyURL)
	if err != nil {
		return nil, fmt.Errorf("请求政策API失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	var policyResp model.PolicyResponse
	if err := json.NewDecoder(resp.Body).Decode(&policyResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if policyResp.Code != 200 {
		return nil, fmt.Errorf("API返回错误: %s", policyResp.Msg)
	}

	return policyResp.Rows, nil
}

// cleanHTML 清理HTML标签
func cleanHTML(html string) string {
	// 移除HTML标签
	re := regexp.MustCompile(`<[^>]*>`)
	text := re.ReplaceAllString(html, " ")

	// 移除多余空格
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")

	return strings.TrimSpace(text)
}

// buildPolicyContent 构建政策文本内容（精简版，控制在512 tokens以内）
func (s *PolicyService) buildPolicyContent(policy model.PolicyInfo) string {
	var builder strings.Builder

	// 核心信息（约100 tokens）
	builder.WriteString(fmt.Sprintf("政策名称：%s\n", policy.Zcmc))
	builder.WriteString(fmt.Sprintf("政策级别：%s\n", policy.ZcLevel))
	builder.WriteString(fmt.Sprintf("来源单位：%s\n", policy.SourceUnit))

	// 政策说明（限制长度，约150 tokens）
	if policy.PolicyExplanation != "" {
		cleaned := cleanHTML(policy.PolicyExplanation)
		if len(cleaned) > 200 {
			cleaned = cleaned[:200] + "..."
		}
		builder.WriteString(fmt.Sprintf("政策说明：%s\n", cleaned))
	}

	// 适用对象（限制长度，约100 tokens）
	if policy.ApplicableObjects != "" {
		cleaned := cleanHTML(policy.ApplicableObjects)
		if len(cleaned) > 150 {
			cleaned = cleaned[:150] + "..."
		}
		builder.WriteString(fmt.Sprintf("适用对象：%s\n", cleaned))
	}

	// 申请条件（限制长度，约100 tokens）
	if policy.ApplyCondition != "" {
		cleaned := cleanHTML(policy.ApplyCondition)
		if len(cleaned) > 150 {
			cleaned = cleaned[:150] + "..."
		}
		builder.WriteString(fmt.Sprintf("申请条件：%s\n", cleaned))
	}

	// 补贴标准（限制长度，约100 tokens）
	if policy.Btbz != "" {
		cleaned := cleanHTML(policy.Btbz)
		if len(cleaned) > 150 {
			cleaned = cleaned[:150] + "..."
		}
		builder.WriteString(fmt.Sprintf("补贴标准：%s\n", cleaned))
	}

	// 政策标签（约20 tokens）
	if policy.Jyzcbq != "" {
		builder.WriteString(fmt.Sprintf("政策标签：%s\n", policy.Jyzcbq))
	}

	content := builder.String()

	// 最终安全检查：如果内容仍然太长，进行截断（约1500字符 ≈ 500 tokens）
	if len(content) > 1500 {
		content = content[:1500] + "..."
	}

	return content
}

// buildFullPolicyContent 构建完整的政策文本内容（用于展示）
func (s *PolicyService) buildFullPolicyContent(policy model.PolicyInfo) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("政策名称：%s\n", policy.Zcmc))
	builder.WriteString(fmt.Sprintf("政策级别：%s\n", policy.ZcLevel))
	builder.WriteString(fmt.Sprintf("来源单位：%s\n", policy.SourceUnit))
	builder.WriteString(fmt.Sprintf("发布时间：%s\n", policy.PublishTime))

	if policy.PolicyExplanation != "" {
		builder.WriteString(fmt.Sprintf("政策说明：%s\n", cleanHTML(policy.PolicyExplanation)))
	}

	if policy.ApplicableObjects != "" {
		builder.WriteString(fmt.Sprintf("适用对象：%s\n", cleanHTML(policy.ApplicableObjects)))
	}

	if policy.ApplyCondition != "" {
		builder.WriteString(fmt.Sprintf("申请条件：%s\n", cleanHTML(policy.ApplyCondition)))
	}

	if policy.Btbz != "" {
		builder.WriteString(fmt.Sprintf("补贴标准：%s\n", cleanHTML(policy.Btbz)))
	}

	if policy.Sqcl != "" {
		builder.WriteString(fmt.Sprintf("申请材料：%s\n", cleanHTML(policy.Sqcl)))
	}

	if policy.Jbqd != "" {
		builder.WriteString(fmt.Sprintf("经办渠道：%s\n", cleanHTML(policy.Jbqd)))
	}

	if policy.Zczc != "" {
		builder.WriteString(fmt.Sprintf("政策支持：%s\n", cleanHTML(policy.Zczc)))
	}

	if policy.Jyzcbq != "" {
		builder.WriteString(fmt.Sprintf("政策标签：%s\n", policy.Jyzcbq))
	}

	if policy.Phone != "" {
		builder.WriteString(fmt.Sprintf("联系电话：%s\n", policy.Phone))
	}

	if policy.Remarks != "" {
		builder.WriteString(fmt.Sprintf("备注：%s\n", policy.Remarks))
	}

	return builder.String()
}

// UpdatePolicies 更新政策到向量数据库
func (s *PolicyService) UpdatePolicies(ctx context.Context) error {
	// 1. 获取政策列表
	policies, err := s.FetchPolicies()
	if err != nil {
		return fmt.Errorf("获取政策列表失败: %w", err)
	}

	if len(policies) == 0 {
		return fmt.Errorf("未获取到政策数据")
	}

	// 2. 批量处理政策
	ids := make([]string, 0, len(policies))
	contents := make([]string, 0, len(policies))
	vectors := make([][]float32, 0, len(policies))

	for _, policy := range policies {
		// 构建精简的政策文本用于向量化
		shortContent := s.buildPolicyContent(policy)

		// 获取向量
		vector, err := s.embeddingClient.GetEmbeddingWithRetry(shortContent, 3)
		if err != nil {
			fmt.Printf("警告：政策 %s 向量化失败: %v\n", policy.Zcmc, err)
			continue
		}

		// 构建完整的政策内容用于存储和展示
		fullContent := s.buildFullPolicyContent(policy)

		ids = append(ids, policy.ID)
		contents = append(contents, fullContent)
		vectors = append(vectors, vector)

		// 避免请求过快
		time.Sleep(100 * time.Millisecond)
	}

	if len(ids) == 0 {
		return fmt.Errorf("没有成功向量化的政策")
	}

	// 3. 插入到Milvus
	if err := s.milvusClient.Insert(ctx, ids, contents, vectors); err != nil {
		return fmt.Errorf("插入向量数据库失败: %w", err)
	}

	fmt.Printf("成功更新 %d 条政策到向量数据库\n", len(ids))
	return nil
}

// SearchPolicies 搜索相关政策
func (s *PolicyService) SearchPolicies(ctx context.Context, query string, topK int) ([]client.SearchResult, error) {
	// 1. 获取查询向量
	vector, err := s.embeddingClient.GetEmbeddingWithRetry(query, 3)
	if err != nil {
		return nil, fmt.Errorf("查询向量化失败: %w", err)
	}

	// 2. 搜索相似政策
	results, err := s.milvusClient.Search(ctx, vector, topK)
	if err != nil {
		return nil, fmt.Errorf("搜索失败: %w", err)
	}

	return results, nil
}

// Close 关闭服务
func (s *PolicyService) Close() error {
	if s.milvusClient != nil {
		return s.milvusClient.Close()
	}
	return nil
}
