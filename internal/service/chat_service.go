package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"qd-sc/internal/client"
	"qd-sc/internal/config"
	"qd-sc/internal/model"
	contentutils "qd-sc/internal/pkg/utils"
	"qd-sc/pkg/utils"
	"regexp"
	"strings"
	"time"
)

// ExposedModelName 对外暴露的固定模型名称
const ExposedModelName = "qd-job-turbo"

// 岗位查询意图关键词
var jobIntentKeywords = []string{
	"岗位", "工作", "招聘", "职位", "就业", "求职", "找工作", "应聘",
	"薪资", "薪酬", "工资", "待遇", "月薪", "年薪",
	"推荐岗位", "推荐工作", "推荐职位",
	"附近的工作", "附近的岗位", "附近招聘",
	"适合我的", "匹配的岗位", "匹配的工作",
	"开发工程师", "产品经理", "设计师", "运营", "销售", "会计", "财务",
	"前端", "后端", "全栈", "Java", "Python", "测试", "运维",
}

// 简历内容特征关键词（用于判断OCR内容是否是简历）
var resumeKeywords = []string{
	// 个人信息相关
	"姓名", "性别", "年龄", "出生", "籍贯", "民族", "身份证",
	"电话", "手机", "邮箱", "邮件", "地址", "住址",
	// 教育背景
	"学历", "学位", "毕业", "本科", "硕士", "博士", "大专", "高中",
	"专业", "院校", "大学", "学院", "在读", "应届",
	// 工作经历
	"工作经验", "工作经历", "任职", "就职", "离职", "在职",
	"公司", "企业", "单位", "部门", "岗位", "职位", "职务",
	// 技能相关
	"技能", "特长", "证书", "资格", "熟练", "精通", "掌握",
	// 自我评价
	"自我评价", "个人简介", "自我介绍", "个人总结", "求职意向",
	// 简历特有标记
	"简历", "履历", "个人资料", "基本信息", "联系方式",
}

// resumeKeywordThreshold 简历关键词匹配阈值（需要匹配到的最小数量）
const resumeKeywordThreshold = 3

// isResumeContent 检测OCR内容是否是简历
func isResumeContent(content string) bool {
	if content == "" {
		return false
	}

	matchCount := 0
	contentLower := strings.ToLower(content)

	for _, keyword := range resumeKeywords {
		if strings.Contains(contentLower, strings.ToLower(keyword)) {
			matchCount++
			if matchCount >= resumeKeywordThreshold {
				log.Printf("检测到简历内容，匹配关键词数量: %d", matchCount)
				return true
			}
		}
	}

	log.Printf("内容不符合简历特征，仅匹配到 %d 个关键词（阈值: %d）", matchCount, resumeKeywordThreshold)
	return false
}

// 岗位信息输出的特征模式（用于检测AI幻觉）
var jobHallucinationPatterns = []*regexp.Regexp{
	regexp.MustCompile(`岗位名称[：:]\s*\S+`),
	regexp.MustCompile(`公司名称[：:]\s*\S+`),
	regexp.MustCompile(`薪资范围[：:]\s*\d+`),
	regexp.MustCompile(`工作地点[：:]\s*\S+`),
	regexp.MustCompile(`学历要求[：:]\s*\S+`),
	regexp.MustCompile(`经验要求[：:]\s*\S+`),
	regexp.MustCompile(`\d+[-~到至]\d+元[/／每]月`),
	regexp.MustCompile(`\d+[kK][-~到至]\d+[kK]`),
	regexp.MustCompile(`(?:推荐|适合)[^。]*(?:岗位|职位|工作)[：:]\s*\d+[.、]`),
	regexp.MustCompile(`以下是[^。]*(?:岗位|职位|工作)`),
	regexp.MustCompile(`为您(?:推荐|找到)[^。]*(?:岗位|职位|工作)`),
}

// containsNonResumeImageHint 检测消息中是否包含非简历图片的提示
func containsNonResumeImageHint(content string) bool {
	return strings.Contains(content, "[用户上传的图片内容（非简历格式）]")
}

// isJobQueryIntent 检测用户输入是否是岗位查询意图
func (s *ChatService) isJobQueryIntent(messages []model.Message) bool {
	// 获取最后一条用户消息
	var lastUserMessage string
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "user" {
			if content, ok := messages[i].Content.(string); ok {
				lastUserMessage = content
				break
			}
		}
	}

	if lastUserMessage == "" {
		return false
	}

	// 如果消息中包含非简历图片的提示，不视为岗位查询意图
	// 因为需要先询问用户意图
	if containsNonResumeImageHint(lastUserMessage) {
		log.Printf("检测到非简历图片，不强制岗位查询意图，需先询问用户")
		return false
	}

	// 转换为小写进行匹配
	lowerMsg := strings.ToLower(lastUserMessage)

	// 检查是否包含岗位相关关键词
	for _, keyword := range jobIntentKeywords {
		if strings.Contains(lowerMsg, strings.ToLower(keyword)) {
			log.Printf("检测到岗位查询意图，匹配关键词: %s", keyword)
			return true
		}
	}

	return false
}

// containsJobHallucination 检测AI回复是否包含岗位幻觉（在没有调用工具的情况下自行输出岗位信息）
func (s *ChatService) containsJobHallucination(content string) bool {
	if content == "" {
		return false
	}

	matchCount := 0
	for _, pattern := range jobHallucinationPatterns {
		if pattern.MatchString(content) {
			matchCount++
			if matchCount >= 2 {
				log.Printf("检测到岗位幻觉输出，匹配模式数量: %d", matchCount)
				return true
			}
		}
	}

	return false
}

// getJobToolChoice 获取强制调用岗位工具的 tool_choice 配置
func (s *ChatService) getJobToolChoice() interface{} {
	// 返回 "required" 强制模型必须调用某个工具
	// 或者返回特定工具的配置强制调用该工具
	return map[string]interface{}{
		"type": "function",
		"function": map[string]string{
			"name": "queryJobsByArea",
		},
	}
}

// getHallucinationWarningMessage 获取幻觉拦截后的警告消息
func (s *ChatService) getHallucinationWarningMessage() string {
	return "抱歉，我需要先查询实际的岗位数据才能为您推荐。请稍等，我正在为您搜索符合条件的岗位..."
}

// rebuildChunksWithFilteredContent 重新构建过滤思维链后的chunks
func (s *ChatService) rebuildChunksWithFilteredContent(originalChunks []*model.ChatCompletionChunk, filteredContent string) []*model.ChatCompletionChunk {
	if len(originalChunks) == 0 || filteredContent == "" {
		return []*model.ChatCompletionChunk{}
	}

	// 创建一个新的chunk包含过滤后的完整内容
	filteredChunk := &model.ChatCompletionChunk{
		ID:      originalChunks[0].ID,
		Object:  originalChunks[0].Object,
		Created: originalChunks[0].Created,
		Model:   originalChunks[0].Model,
		Choices: []model.ChunkChoice{
			{
				Index: 0,
				Delta: model.Message{
					Role:    "assistant",
					Content: filteredContent,
				},
				FinishReason: "stop",
			},
		},
	}

	return []*model.ChatCompletionChunk{filteredChunk}
}

// thinkingTagState 思维链标签状态
type thinkingTagState struct {
	insideThinking bool
	buffer         strings.Builder
}

var globalThinkingState = &thinkingTagState{}

// filterThinkingTagsRealtime 实时过滤思维链标签（处理跨chunk的情况）
func (s *ChatService) filterThinkingTagsRealtime(content string) string {
	if content == "" {
		return content
	}

	var result strings.Builder
	i := 0

	for i < len(content) {
		if globalThinkingState.insideThinking {
			// 当前在思维链内部，查找结束标签
			endIndex := strings.Index(content[i:], "</think>")
			if endIndex != -1 {
				// 找到结束标签，跳过思维链内容和结束标签
				i += endIndex + len("</think>")
				globalThinkingState.insideThinking = false
				globalThinkingState.buffer.Reset()
			} else {
				// 没找到结束标签，整个chunk都在思维链内部
				return ""
			}
		} else {
			// 当前在思维链外部，查找开始标签
			startIndex := strings.Index(content[i:], "<think>")
			if startIndex != -1 {
				// 找到开始标签，保留之前的内容
				result.WriteString(content[i : i+startIndex])
				i += startIndex + len("<think>")
				globalThinkingState.insideThinking = true
			} else {
				// 没找到开始标签，保留剩余内容
				result.WriteString(content[i:])
				break
			}
		}
	}

	return result.String()
}

// ChatService 对话服务
type ChatService struct {
	cfg             *config.Config
	llmClient       *client.LLMClient
	ocrClient       *client.OCRClient
	locationService *LocationService
	jobService      *JobService
	policyService   *PolicyService
}

// NewChatService 创建对话服务
func NewChatService(
	cfg *config.Config,
	llmClient *client.LLMClient,
	ocrClient *client.OCRClient,
	locationService *LocationService,
	jobService *JobService,
	policyService *PolicyService,
) *ChatService {
	return &ChatService{
		cfg:             cfg,
		llmClient:       llmClient,
		ocrClient:       ocrClient,
		locationService: locationService,
		jobService:      jobService,
		policyService:   policyService,
	}
}

// ProcessChatRequest 处理聊天请求
// ProcessChatRequest 处理聊天请求
func (s *ChatService) ProcessChatRequest(req *model.ChatCompletionRequest) (*model.ChatCompletionResponse, error) {
	// 重置思维链状态
	globalThinkingState.insideThinking = false
	globalThinkingState.buffer.Reset()

	// 准备消息
	messages := s.prepareMessages(req.Messages)

	// 添加工具定义
	tools := model.GetAvailableTools()

	// 检测是否是岗位查询意图
	isJobIntent := s.isJobQueryIntent(req.Messages)
	var toolChoice interface{} = "auto"
	if isJobIntent {
		// 岗位场景：使用 auto 模式让模型自动决定是否调用工具
		toolChoice = "auto"
		log.Printf("检测到岗位查询意图，设置 tool_choice=auto")
	}

	// 构建请求（使用配置文件中的实际模型名称）
	llmReq := &model.ChatCompletionRequest{
		Model:       s.cfg.LLM.Model,
		Messages:    messages,
		Tools:       tools,
		ToolChoice:  toolChoice,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		MaxTokens:   req.MaxTokens,
	}

	// 追踪是否已调用过岗位工具
	jobToolCalled := false

	// 开始对话循环（支持多轮工具调用）
	maxIterations := 10
	for i := 0; i < maxIterations; i++ {
		resp, err := s.llmClient.ChatCompletion(llmReq)
		if err != nil {
			return nil, fmt.Errorf("LLM请求失败: %w", err)
		}

		if len(resp.Choices) == 0 {
			return nil, fmt.Errorf("LLM返回空结果")
		}

		choice := resp.Choices[0]

		// 检查finish_reason，如果是"stop"表示对话已完成
		if choice.FinishReason == "stop" && len(choice.Message.ToolCalls) == 0 {
			// 岗位意图场景下的幻觉检测
			if isJobIntent && !jobToolCalled {
				if content, ok := choice.Message.Content.(string); ok {
					// 过滤思维链标签
					filteredContent := contentutils.FilterThinkingTags(content)
					choice.Message.Content = filteredContent

					if s.containsJobHallucination(filteredContent) {
						log.Printf("拦截岗位幻觉输出，强制重新调用工具")
						// 修改消息，添加系统提示强制调用工具
						llmReq.Messages = append(llmReq.Messages, model.Message{
							Role:    "user",
							Content: "请调用岗位查询工具获取真实数据，不要自行编造岗位信息。",
						})
						llmReq.ToolChoice = s.getJobToolChoice()
						continue
					}
				}
			} else {
				// 非岗位意图场景也需要过滤思维链
				if content, ok := choice.Message.Content.(string); ok {
					filteredContent := contentutils.FilterThinkingTags(content)
					choice.Message.Content = filteredContent
				}
			}
			log.Printf("模型返回finish_reason=stop，对话结束")
			return resp, nil
		}

		// 如果没有工具调用，返回最终结果
		if len(choice.Message.ToolCalls) == 0 {
			// 岗位意图场景下的幻觉检测
			if isJobIntent && !jobToolCalled {
				if content, ok := choice.Message.Content.(string); ok {
					// 过滤思维链标签
					filteredContent := contentutils.FilterThinkingTags(content)
					choice.Message.Content = filteredContent

					if s.containsJobHallucination(filteredContent) {
						log.Printf("拦截岗位幻觉输出，强制重新调用工具")
						llmReq.Messages = append(llmReq.Messages, model.Message{
							Role:    "user",
							Content: "请调用岗位查询工具获取真实数据，不要自行编造岗位信息。",
						})
						llmReq.ToolChoice = s.getJobToolChoice()
						continue
					}
				}
			} else {
				// 非岗位意图场景也需要过滤思维链
				if content, ok := choice.Message.Content.(string); ok {
					filteredContent := contentutils.FilterThinkingTags(content)
					choice.Message.Content = filteredContent
				}
			}
			return resp, nil
		}

		// 处理工具调用
		log.Printf("检测到工具调用，finish_reason: %s", choice.FinishReason)
		llmReq.Messages = append(llmReq.Messages, choice.Message)

		for _, toolCall := range choice.Message.ToolCalls {
			// 检查是否是岗位工具调用
			if toolCall.Function.Name == "queryJobsByArea" || toolCall.Function.Name == "queryJobsByLocation" {
				jobToolCalled = true
			}

			result, err := s.executeToolCall(&toolCall)
			if err != nil {
				result = fmt.Sprintf("工具调用失败: %s", err.Error())
				log.Printf("工具调用失败 [%s]: %v", toolCall.Function.Name, err)
			}

			// 添加工具响应
			llmReq.Messages = append(llmReq.Messages, model.Message{
				Role:       "tool",
				Content:    result,
				ToolCallID: toolCall.ID,
			})
		}

		// 岗位工具调用成功后，恢复为 auto 模式
		if jobToolCalled {
			llmReq.ToolChoice = "auto"
		}
	}

	return nil, fmt.Errorf("超过最大工具调用次数")
}

// ProcessChatRequestStream 处理聊天请求（流式）
func (s *ChatService) ProcessChatRequestStream(ctx context.Context, req *model.ChatCompletionRequest) (chan *model.ChatCompletionChunk, chan error) {
	chunkChan := make(chan *model.ChatCompletionChunk, 100)
	errChan := make(chan error, 1)

	go func() {
		defer close(chunkChan)
		defer close(errChan)

		// 重置思维链状态
		globalThinkingState.insideThinking = false
		globalThinkingState.buffer.Reset()

		// 准备消息
		messages := s.prepareMessages(req.Messages)

		// 添加工具定义
		tools := model.GetAvailableTools()

		// 检测是否是岗位查询意图
		isJobIntent := s.isJobQueryIntent(req.Messages)
		var toolChoice interface{} = "auto"
		if isJobIntent {
			// 岗位场景：使用 auto 模式让模型自动决定是否调用工具
			toolChoice = "auto"
			log.Printf("【流式】检测到岗位查询意图，设置 tool_choice=auto")
		}

		// 构建请求（使用配置文件中的实际模型名称）
		llmReq := &model.ChatCompletionRequest{
			Model:       s.cfg.LLM.Model,
			Messages:    messages,
			Tools:       tools,
			ToolChoice:  toolChoice,
			Temperature: req.Temperature,
			TopP:        req.TopP,
			MaxTokens:   req.MaxTokens,
			Stream:      true,
		}

		// 追踪是否已发送第一个chunk（用于正确设置role）
		firstChunkSent := false

		// 追踪是否已调用过岗位工具
		jobToolCalled := false

		// 追踪是否已经发送过幻觉拦截消息（避免重复发送）
		hallucinationIntercepted := false

		// 开始对话循环
		maxIterations := 10
		for iteration := 0; iteration < maxIterations; iteration++ {
			// 检查context是否已取消
			select {
			case <-ctx.Done():
				log.Printf("请求被取消: %v", ctx.Err())
				return
			default:
			}
			responseChan, respErrChan, err := s.llmClient.ChatCompletionStream(llmReq)
			if err != nil {
				errChan <- fmt.Errorf("LLM流式请求失败: %w", err)
				return
			}

			var currentMessage model.Message
			var toolCalls []model.ToolCall
			var finishReason string
			currentMessage.Role = "assistant"

			// 用于合并重复日志的计数器
			filteredToolCallsCount := 0
			filteredFinishReasonCount := 0

			// 用于岗位幻觉检测的内容缓冲（岗位场景下先缓冲，检测后再决定是否转发）
			var contentBuffer strings.Builder
			var pendingChunks []*model.ChatCompletionChunk

			// 收集流式响应
			for {
				select {
				case <-ctx.Done():
					// context被取消，停止处理
					log.Printf("流式处理被取消: %v", ctx.Err())
					return
				case chunk, ok := <-responseChan:
					if !ok {
						responseChan = nil
						break
					}

					// 处理chunk的role字段：只有第一个chunk保留role，后续chunk清除role
					if len(chunk.Choices) > 0 {
						if !firstChunkSent {
							// 第一个chunk，确保有role="assistant"
							if chunk.Choices[0].Delta.Role == "" {
								chunk.Choices[0].Delta.Role = "assistant"
							}
							firstChunkSent = true
						} else {
							// 后续chunks，清除role让omitempty生效
							chunk.Choices[0].Delta.Role = ""
						}
					}

					// 收集消息内容和finish_reason（在转发之前）
					if len(chunk.Choices) > 0 {
						delta := chunk.Choices[0].Delta

						if content, ok := delta.Content.(string); ok && content != "" {
							if currentMessage.Content == nil {
								currentMessage.Content = ""
							}
							currentMessage.Content = currentMessage.Content.(string) + content
							// 同时写入缓冲区
							contentBuffer.WriteString(content)
						}

						// 收集工具调用（注意：流式响应中工具调用可能分块到达）
						if len(delta.ToolCalls) > 0 {
							toolCalls = append(toolCalls, delta.ToolCalls...)
						}

						// 收集finish_reason
						if chunk.Choices[0].FinishReason != "" {
							finishReason = chunk.Choices[0].FinishReason
							log.Printf("收到finish_reason: %s", finishReason)
						}
					}

					// 只转发内容chunk，不转发包含tool_calls或finish_reason=tool_calls的chunk
					// 因为我们的工具调用是在服务端自动处理的，不需要客户端参与
					shouldForward := true
					if len(chunk.Choices) > 0 {
						delta := chunk.Choices[0].Delta
						// 如果chunk包含tool_calls，不转发
						if len(delta.ToolCalls) > 0 {
							shouldForward = false
							filteredToolCallsCount++
						}
						// 如果finish_reason是tool_calls，不转发
						if chunk.Choices[0].FinishReason == "tool_calls" {
							shouldForward = false
							filteredFinishReasonCount++
						}
					}

					// 岗位意图场景下的特殊处理：先缓冲内容，检测幻觉后再决定转发
					if shouldForward && isJobIntent && !jobToolCalled && len(toolCalls) == 0 {
						// 岗位场景且未调用工具：缓冲chunk，稍后检测
						chunkCopy := *chunk
						pendingChunks = append(pendingChunks, &chunkCopy)
					} else if shouldForward {
						// 非岗位场景或已调用工具：实时过滤思维链后转发
						if len(chunk.Choices) > 0 {
							if content, ok := chunk.Choices[0].Delta.Content.(string); ok && content != "" {
								// 实时过滤思维链标签 - 处理跨chunk的情况
								filteredContent := s.filterThinkingTagsRealtime(content)
								if filteredContent != "" {
									// 创建过滤后的chunk
									filteredChunk := *chunk
									filteredChunk.Choices[0].Delta.Content = filteredContent
									chunkChan <- &filteredChunk
								}
								// 如果过滤后内容为空，则不转发这个chunk
							} else {
								// 非内容chunk，直接转发
								chunkChan <- chunk
							}
						} else {
							// 没有choices，直接转发
							chunkChan <- chunk
						}
					}

				case err, ok := <-respErrChan:
					if ok && err != nil {
						errChan <- err
						return
					}
				}

				if responseChan == nil {
					break
				}
			}

			// 输出合并后的过滤日志
			if filteredToolCallsCount > 0 {
				log.Printf("过滤tool_calls chunk x%d，不转发给客户端", filteredToolCallsCount)
			}
			if filteredFinishReasonCount > 0 {
				log.Printf("过滤finish_reason=tool_calls chunk x%d，不转发给客户端", filteredFinishReasonCount)
			}

			// 岗位意图场景下的幻觉检测
			if isJobIntent && !jobToolCalled && len(toolCalls) == 0 {
				bufferedContent := contentBuffer.String()

				// 过滤思维链标签
				filteredContent := contentutils.FilterThinkingTags(bufferedContent)

				if s.containsJobHallucination(filteredContent) {
					log.Printf("【流式】拦截岗位幻觉输出，内容长度: %d，丢弃缓冲的 %d 个chunks", len(filteredContent), len(pendingChunks))

					// 不转发幻觉内容，发送警告消息并强制重新调用工具
					if !hallucinationIntercepted {
						hallucinationIntercepted = true
						// 发送警告消息给用户
						warningChunk := &model.ChatCompletionChunk{
							ID:      fmt.Sprintf("chatcmpl-%d", time.Now().Unix()),
							Object:  "chat.completion.chunk",
							Created: time.Now().Unix(),
							Model:   ExposedModelName,
							Choices: []model.ChunkChoice{
								{
									Index: 0,
									Delta: model.Message{
										Content: s.getHallucinationWarningMessage(),
									},
								},
							},
						}
						chunkChan <- warningChunk
					}

					// 修改消息，强制调用工具
					llmReq.Messages = append(llmReq.Messages, model.Message{
						Role:    "user",
						Content: "请调用岗位查询工具获取真实数据，不要自行编造岗位信息。",
					})
					llmReq.ToolChoice = s.getJobToolChoice()
					continue
				} else {
					// 没有幻觉，但需要过滤思维链后转发缓冲的chunks
					if contentutils.ContainsThinkingTags(bufferedContent) {
						log.Printf("【流式】检测到思维链标签，过滤后转发")
						// 重新构建过滤后的chunks
						filteredChunks := s.rebuildChunksWithFilteredContent(pendingChunks, filteredContent)
						for _, filteredChunk := range filteredChunks {
							chunkChan <- filteredChunk
						}
					} else {
						// 没有思维链，直接转发原始chunks
						for _, pendingChunk := range pendingChunks {
							chunkChan <- pendingChunk
						}
					}
				}
			}

			// 根据finish_reason决定是否继续
			// 如果finish_reason是"stop"，表示模型认为对话已完成，应该结束
			if finishReason == "stop" {
				log.Printf("模型返回finish_reason=stop，对话结束")
				return
			}

			// 如果没有工具调用，结束
			if len(toolCalls) == 0 {
				return
			}

			// 合并工具调用
			currentMessage.ToolCalls = s.mergeToolCalls(toolCalls)
			llmReq.Messages = append(llmReq.Messages, currentMessage)

			// 执行工具调用并继续对话
			for _, toolCall := range currentMessage.ToolCalls {
				// 检查是否是岗位工具调用
				if toolCall.Function.Name == "queryJobsByArea" || toolCall.Function.Name == "queryJobsByLocation" {
					jobToolCalled = true
				}

				result, err := s.executeToolCall(&toolCall)
				var callSuccess bool

				if err != nil {
					result = fmt.Sprintf("工具调用失败: %s", err.Error())
					log.Printf("工具调用失败 [%s]: %v", toolCall.Function.Name, err)
					callSuccess = false
				} else {
					callSuccess = true
				}

				// 检查是否是岗位查询工具，且调用成功，需要分块输出
				if callSuccess && (toolCall.Function.Name == "queryJobsByArea" || toolCall.Function.Name == "queryJobsByLocation") {
					// 分块输出岗位信息
					if err := s.streamJobResults(chunkChan, result, ExposedModelName); err != nil {
						log.Printf("流式输出岗位失败: %v", err)
					}

					// 岗位展示完成后，直接发送一个空的final chunk结束对话
					// 这样客户端会正确识别对话已完成，不会再发起后续请求
					finalChunk := &model.ChatCompletionChunk{
						ID:      fmt.Sprintf("chatcmpl-%d", time.Now().Unix()),
						Object:  "chat.completion.chunk",
						Created: time.Now().Unix(),
						Model:   ExposedModelName,
						Choices: []model.ChunkChoice{
							{
								Index:        0,
								Delta:        model.Message{},
								FinishReason: "stop",
							},
						},
					}
					chunkChan <- finalChunk
					log.Printf("岗位推荐完成，发送finish_reason=stop并结束")
					return
				}

				// 确保result不为空
				if result == "" {
					result = "工具执行完成"
				}

				llmReq.Messages = append(llmReq.Messages, model.Message{
					Role:       "tool",
					Content:    result,
					ToolCallID: toolCall.ID,
				})
			}

			// 岗位工具调用成功后，恢复为 auto 模式
			if jobToolCalled {
				llmReq.ToolChoice = "auto"
			}

			// 发送一个提示chunk，表示正在处理工具调用
			// role留空，因为这不是第一个chunk
			chunkChan <- &model.ChatCompletionChunk{
				ID:      fmt.Sprintf("chatcmpl-%d", time.Now().Unix()),
				Object:  "chat.completion.chunk",
				Created: time.Now().Unix(),
				Model:   ExposedModelName,
				Choices: []model.ChunkChoice{
					{
						Index: 0,
						Delta: model.Message{
							Content: "\n\n",
						},
					},
				},
			}
		}

		errChan <- fmt.Errorf("超过最大工具调用次数")
	}()

	return chunkChan, errChan
}

// prepareMessages 准备消息列表
func (s *ChatService) prepareMessages(userMessages []model.Message) []model.Message {
	messages := []model.Message{
		{
			Role:    "system",
			Content: model.GetSystemPrompt(),
		},
	}

	// 处理用户消息，支持 OpenAI Vision API 格式的文件URL消息
	for _, msg := range userMessages {
		processedMsg := s.processMessageWithFileURLs(msg)
		messages = append(messages, processedMsg)
	}

	return messages
}

// processMessageWithFileURLs 处理消息中的文件URL，使用OCR服务解析
// 支持 OpenAI Vision API 兼容格式（image_url 字段），可解析图片、PDF、Excel、PPT 等文件
func (s *ChatService) processMessageWithFileURLs(msg model.Message) model.Message {
	// 检查 Content 是否是数组类型（OpenAI Vision API 格式）
	contentArray, ok := msg.Content.([]interface{})
	if !ok {
		// 不是数组，直接返回原消息
		return msg
	}

	var textParts []string
	var imageContents []string

	for _, item := range contentArray {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		contentType, _ := itemMap["type"].(string)

		switch contentType {
		case "text":
			if text, ok := itemMap["text"].(string); ok {
				textParts = append(textParts, text)
			}
		case "image_url":
			// 处理图片 URL
			imageURLData, ok := itemMap["image_url"].(map[string]interface{})
			if !ok {
				continue
			}
			imageURL, ok := imageURLData["url"].(string)
			if !ok || imageURL == "" {
				continue
			}

			// 使用 OCR 服务解析图片
			log.Printf("检测到图片URL，使用OCR服务解析: %s", imageURL)
			ocrContent, err := s.ocrClient.ParseURL(imageURL)
			if err != nil {
				log.Printf("OCR解析图片失败: %v", err)
				imageContents = append(imageContents, fmt.Sprintf("[图片解析失败: %s]", err.Error()))
			} else {
				log.Printf("OCR解析图片成功，内容长度: %d", len(ocrContent))

				// 检测OCR内容是否是简历
				if isResumeContent(ocrContent) {
					// 是简历，正常处理
					imageContents = append(imageContents, fmt.Sprintf("[用户上传的简历内容]:\n%s", ocrContent))
				} else {
					// 不是简历，添加提示让模型先询问用户意图
					imageContents = append(imageContents, fmt.Sprintf("[用户上传的图片内容（非简历格式）]:\n%s\n\n[重要提示]: 该图片内容不像标准简历，请先询问用户上传这张图片的意图是什么，确认用户需求后再提供相应帮助。不要直接假设用户想找工作。", ocrContent))
					log.Printf("OCR内容不是简历格式，已添加询问用户意图的提示")
				}
			}
		}
	}

	// 合并文本和图片内容
	var finalContent string
	if len(textParts) > 0 {
		finalContent = strings.Join(textParts, "\n")
	}
	if len(imageContents) > 0 {
		if finalContent != "" {
			finalContent += "\n\n"
		}
		finalContent += strings.Join(imageContents, "\n\n")
	}

	return model.Message{
		Role:    msg.Role,
		Content: finalContent,
		Name:    msg.Name,
	}
}

// executeToolCall 执行工具调用
func (s *ChatService) executeToolCall(toolCall *model.ToolCall) (string, error) {
	funcName := toolCall.Function.Name
	arguments := toolCall.Function.Arguments

	log.Printf("执行工具调用: %s", funcName)
	log.Printf("工具参数: %s", arguments)

	// 验证参数不为空
	if arguments == "" {
		return "", fmt.Errorf("工具参数为空")
	}

	// 解析参数
	var params map[string]interface{}
	if err := json.Unmarshal([]byte(arguments), &params); err != nil {
		log.Printf("参数解析失败，原始参数: [%s]", arguments)
		return "", fmt.Errorf("解析工具参数失败: %w", err)
	}

	log.Printf("解析后的参数: %+v", params)

	// 根据工具名称执行相应操作
	switch funcName {
	case "queryLocation":
		return s.handleQueryLocation(params)
	case "queryJobsByArea":
		return s.handleQueryJobsByArea(params)
	case "queryJobsByLocation":
		return s.handleQueryJobsByLocation(params)
	case "parsePDF":
		return s.handleParsePDF(params)
	case "parseImage":
		return s.handleParseImage(params)
	case "queryPolicy":
		return s.handleQueryPolicy(params)
	default:
		return "", fmt.Errorf("未知的工具: %s", funcName)
	}
}

// handleQueryLocation 处理地理位置查询
func (s *ChatService) handleQueryLocation(params map[string]interface{}) (string, error) {
	keywords, ok := params["keywords"].(string)
	if !ok {
		return "", fmt.Errorf("缺少keywords参数")
	}

	lat, lng, err := s.locationService.QueryLocation(keywords)
	if err != nil {
		return "", err
	}

	result := map[string]string{
		"keywords":  keywords,
		"latitude":  lat,
		"longitude": lng,
		"message":   fmt.Sprintf("成功获取地点 %s 的坐标", keywords),
	}

	return utils.ToJSONStringPretty(result)
}

// handleQueryJobsByArea 处理按区域查询岗位
func (s *ChatService) handleQueryJobsByArea(params map[string]interface{}) (string, error) {
	return s.jobService.QueryJobsByArea(params)
}

// handleQueryJobsByLocation 处理按位置查询岗位
func (s *ChatService) handleQueryJobsByLocation(params map[string]interface{}) (string, error) {
	return s.jobService.QueryJobsByLocation(params)
}

// handleParsePDF 处理PDF解析
func (s *ChatService) handleParsePDF(params map[string]interface{}) (string, error) {
	_, ok := params["fileUrl"].(string)
	if !ok {
		return "", fmt.Errorf("缺少fileUrl参数")
	}

	// 这里应该调用OCR服务解析，但由于URL可能是本地路径，需要特殊处理
	// 实际使用时，文件已在上传阶段被OCR服务解析，此工具仅作为备用
	return "", fmt.Errorf("PDF解析功能需要配合文件上传使用")
}

// handleParseImage 处理图片解析
func (s *ChatService) handleParseImage(params map[string]interface{}) (string, error) {
	_, ok := params["imageUrl"].(string)
	if !ok {
		return "", fmt.Errorf("缺少imageUrl参数")
	}

	// 这里应该调用视觉模型解析，但由于URL可能是本地路径，需要特殊处理
	// 实际使用时，文件已在上传阶段被解析，此工具仅作为备用
	return "", fmt.Errorf("图片解析功能需要配合文件上传使用")
}

// handleQueryPolicy 处理政策咨询
func (s *ChatService) handleQueryPolicy(params map[string]interface{}) (string, error) {
	query, ok := params["query"].(string)
	if !ok || query == "" {
		return "", fmt.Errorf("缺少query参数")
	}

	// 获取topK参数，默认为3
	topK := 3
	if v, ok := params["topK"].(float64); ok {
		topK = int(v)
	}

	// 搜索相关政策
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	results, err := s.policyService.SearchPolicies(ctx, query, topK)
	if err != nil {
		return "", fmt.Errorf("搜索政策失败: %w", err)
	}

	if len(results) == 0 {
		return "未找到相关政策信息，建议您换个关键词重新搜索或联系相关部门咨询。", nil
	}

	// 格式化返回结果
	var resultBuilder strings.Builder
	resultBuilder.WriteString(fmt.Sprintf("为您找到 %d 条相关政策：\n\n", len(results)))

	for i, result := range results {
		resultBuilder.WriteString(fmt.Sprintf("【政策 %d】\n", i+1))
		resultBuilder.WriteString(result.Content)
		resultBuilder.WriteString("\n")
		resultBuilder.WriteString(fmt.Sprintf("相似度评分: %.2f\n", 1.0-result.Distance))
		resultBuilder.WriteString("\n---\n\n")
	}

	return resultBuilder.String(), nil
}

// mergeToolCalls 合并流式响应中的工具调用
// 流式响应中，每个工具调用会分成多个chunk，每个chunk可能只包含几个字符
// 需要按照每个chunk中delta.tool_calls的index字段来正确合并
func (s *ChatService) mergeToolCalls(toolCalls []model.ToolCall) []model.ToolCall {
	if len(toolCalls) == 0 {
		return nil
	}

	// 使用map按index合并
	mergedMap := make(map[int]*model.ToolCall)
	maxIndex := -1

	for _, tc := range toolCalls {
		// 获取index
		idx := 0
		if tc.Index != nil {
			idx = *tc.Index
		}

		if idx > maxIndex {
			maxIndex = idx
		}

		// 如果该index已存在，合并数据
		if existing, ok := mergedMap[idx]; ok {
			// 合并ID（只在第一次出现时设置）
			if tc.ID != "" {
				existing.ID = tc.ID
			}
			// 合并Type
			if tc.Type != "" {
				existing.Type = tc.Type
			}
			// 累加函数名称
			if tc.Function.Name != "" {
				existing.Function.Name += tc.Function.Name
			}
			// 累加参数
			if tc.Function.Arguments != "" {
				existing.Function.Arguments += tc.Function.Arguments
			}
		} else {
			// 新的工具调用
			tcCopy := tc
			mergedMap[idx] = &tcCopy
		}
	}

	// 按index顺序转换为数组
	result := make([]model.ToolCall, 0, len(mergedMap))
	for i := 0; i <= maxIndex; i++ {
		if tc, ok := mergedMap[i]; ok {
			// 验证工具调用是否完整
			if tc.Function.Name != "" && tc.Function.Arguments != "" {
				result = append(result, *tc)
				log.Printf("合并后的工具调用 [%d]: %s, 参数长度: %d", i, tc.Function.Name, len(tc.Function.Arguments))
			} else {
				log.Printf("警告：工具调用 [%d] 不完整 - Name: '%s', Args length: %d",
					i, tc.Function.Name, len(tc.Function.Arguments))
			}
		}
	}

	return result
}

// streamJobResults 流式输出岗位结果（分块，每个岗位间隔1秒）
func (s *ChatService) streamJobResults(chunkChan chan *model.ChatCompletionChunk, jobsJSON string, modelName string) error {
	// 解析岗位列表
	var jobResp model.JobResponse
	if err := json.Unmarshal([]byte(jobsJSON), &jobResp); err != nil {
		return fmt.Errorf("解析岗位数据失败: %w", err)
	}

	if len(jobResp.JobListings) == 0 {
		// 没有岗位，发送提示信息（role留空，不是第一个chunk）
		chunk := &model.ChatCompletionChunk{
			ID:      fmt.Sprintf("chatcmpl-%d", time.Now().Unix()),
			Object:  "chat.completion.chunk",
			Created: time.Now().Unix(),
			Model:   modelName,
			Choices: []model.ChunkChoice{
				{
					Index: 0,
					Delta: model.Message{
						Content: "\n\n未找到符合条件的岗位。\n",
					},
				},
			},
		}
		chunkChan <- chunk
		return nil
	}

	// 发送提示信息（role留空，不是第一个chunk）
	introChunk := &model.ChatCompletionChunk{
		ID:      fmt.Sprintf("chatcmpl-%d", time.Now().Unix()),
		Object:  "chat.completion.chunk",
		Created: time.Now().Unix(),
		Model:   modelName,
		Choices: []model.ChunkChoice{
			{
				Index: 0,
				Delta: model.Message{
					Content: fmt.Sprintf("\n\n为您找到 %d 个相关岗位：\n\n", len(jobResp.JobListings)),
				},
			},
		},
	}
	chunkChan <- introChunk

	// 逐个输出岗位，每个间隔1秒
	for i, job := range jobResp.JobListings {
		// 如果是最后一个岗位且有data字段，添加data
		if i == len(jobResp.JobListings)-1 && jobResp.Data != nil {
			job.Data = jobResp.Data
		}

		// 将岗位格式化为JSON
		jobJSON, err := json.MarshalIndent(job, "", "  ")
		if err != nil {
			log.Printf("格式化岗位失败: %v", err)
			continue
		}

		// 使用 ``` job-json 包裹
		jobContent := fmt.Sprintf("``` job-json\n%s\n```\n\n", string(jobJSON))

		// 发送岗位chunk（role留空，不是第一个chunk）
		jobChunk := &model.ChatCompletionChunk{
			ID:      fmt.Sprintf("chatcmpl-%d-%d", time.Now().Unix(), i),
			Object:  "chat.completion.chunk",
			Created: time.Now().Unix(),
			Model:   modelName,
			Choices: []model.ChunkChoice{
				{
					Index: 0,
					Delta: model.Message{
						Content: jobContent,
					},
				},
			},
		}
		chunkChan <- jobChunk

		// 间隔1秒（除了最后一个）
		if i < len(jobResp.JobListings)-1 {
			time.Sleep(1 * time.Second)
		}
	}

	return nil
}
