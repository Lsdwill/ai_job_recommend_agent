package model

// ChatCompletionRequest OpenAI Chat Completion请求
type ChatCompletionRequest struct {
	Model            string             `json:"model"`
	Messages         []Message          `json:"messages"`
	Temperature      *float64           `json:"temperature,omitempty"`
	TopP             *float64           `json:"top_p,omitempty"`
	N                *int               `json:"n,omitempty"`
	Stream           bool               `json:"stream,omitempty"`
	Stop             interface{}        `json:"stop,omitempty"`
	MaxTokens        *int               `json:"max_tokens,omitempty"`
	PresencePenalty  *float64           `json:"presence_penalty,omitempty"`
	FrequencyPenalty *float64           `json:"frequency_penalty,omitempty"`
	LogitBias        map[string]float64 `json:"logit_bias,omitempty"`
	User             string             `json:"user,omitempty"`
	Tools            []Tool             `json:"tools,omitempty"`
	ToolChoice       interface{}        `json:"tool_choice,omitempty"`
}

// Message 消息结构
type Message struct {
	Role       string      `json:"role,omitempty"` // system, user, assistant, tool - omitempty让空字符串不被序列化
	Content    interface{} `json:"content,omitempty"`
	Name       string      `json:"name,omitempty"`
	ToolCalls  []ToolCall  `json:"tool_calls,omitempty"`
	ToolCallID string      `json:"tool_call_id,omitempty"`
}

// MessageContent 消息内容（支持文本和图片）
type MessageContent struct {
	Type     string        `json:"type"` // text, image_url
	Text     string        `json:"text,omitempty"`
	ImageURL *ImageURLInfo `json:"image_url,omitempty"`
}

// ImageURLInfo 图片URL信息
type ImageURLInfo struct {
	URL    string `json:"url"`
	Detail string `json:"detail,omitempty"` // auto, low, high
}

// Tool 工具定义
type Tool struct {
	Type     string      `json:"type"` // function
	Function FunctionDef `json:"function"`
}

// FunctionDef 函数定义
type FunctionDef struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	Parameters  interface{} `json:"parameters,omitempty"`
}

// ToolCall 工具调用
type ToolCall struct {
	Index    *int         `json:"index,omitempty"` // 流式响应中的索引
	ID       string       `json:"id,omitempty"`
	Type     string       `json:"type,omitempty"` // function
	Function FunctionCall `json:"function"`
}

// FunctionCall 函数调用
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ChatCompletionResponse OpenAI Chat Completion响应
type ChatCompletionResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   *Usage   `json:"usage,omitempty"`
}

// Choice 选择项
type Choice struct {
	Index        int      `json:"index"`
	Message      Message  `json:"message,omitempty"`
	Delta        *Message `json:"delta,omitempty"`
	FinishReason string   `json:"finish_reason,omitempty"`
}

// Usage token使用情况
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ChatCompletionChunk 流式响应chunk
type ChatCompletionChunk struct {
	ID      string        `json:"id"`
	Object  string        `json:"object"`
	Created int64         `json:"created"`
	Model   string        `json:"model"`
	Choices []ChunkChoice `json:"choices"`
}

// ChunkChoice 流式选择项
type ChunkChoice struct {
	Index        int     `json:"index"`
	Delta        Message `json:"delta"`
	FinishReason string  `json:"finish_reason,omitempty"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail 错误详情
type ErrorDetail struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code,omitempty"`
}
