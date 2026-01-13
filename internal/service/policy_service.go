package service

import (
	"fmt"
	"qd-sc/internal/client"
	"qd-sc/internal/config"
	"qd-sc/internal/model"
)

// PolicyService 政策咨询服务
type PolicyService struct {
	policyClient *client.PolicyClient
}

// NewPolicyService 创建政策咨询服务
func NewPolicyService(cfg *config.Config) *PolicyService {
	return &PolicyService{
		policyClient: client.NewPolicyClient(cfg),
	}
}

// QueryPolicy 查询政策信息
func (s *PolicyService) QueryPolicy(
	message string,
	chatID string,
	conversationID string,
	realName bool,
	aac001 string,
	aac147 string,
	aac003 string,
) (string, string, string, error) {
	// 构建请求
	chatReq := &model.PolicyChatData{
		ChatID:         chatID,
		ConversationID: conversationID,
		Stream:         false,
		RealName:       realName,
		Message:        message,
		MegType:        "MESSAGE",
		AAC001:         aac001,
		AAC147:         aac147,
		AAC003:         aac003,
		ReqType:        "1", // 1表示政策咨询
	}

	// 调用政策大模型接口
	resp, err := s.policyClient.Chat(chatReq)
	if err != nil {
		return "", "", "", fmt.Errorf("政策咨询失败: %w", err)
	}

	if resp.Data == nil {
		return "", "", "", fmt.Errorf("政策咨询返回数据为空")
	}

	// 返回消息内容、chatID和conversationID供下次调用使用
	return resp.Data.Message, resp.Data.ChatID, resp.Data.ConversationID, nil
}
