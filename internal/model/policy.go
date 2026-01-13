package model

// PolicyTicketRequest 获取ticket的请求
type PolicyTicketRequest struct {
	LoginName string `json:"loginname"` // 用户名
	UserKey   string `json:"userkey"`   // 密码加密
}

// PolicyTicketResponse 获取ticket的响应
type PolicyTicketResponse struct {
	Code    int               `json:"code"`    // 响应编码，成功：200
	Message string            `json:"message"` // 响应信息
	Data    *PolicyTicketData `json:"data"`    // 响应数据
}

// PolicyTicketData ticket响应数据
type PolicyTicketData struct {
	AppID      string `json:"appid"`      // appid
	PrivateKey string `json:"privateKey"` // 私钥
	SM4Key     string `json:"sm4Key"`     // SM4加密key
	Ticket     string `json:"ticket"`     // Ticket，有效时间1小时
}

// PolicyChatRequest 政策咨询对话请求
type PolicyChatRequest struct {
	AppID  string          `json:"appid"`  // 用户唯一标识
	Ticket string          `json:"ticket"` // 请求票据号
	Data   *PolicyChatData `json:"data"`   // 接口入参信息
}

// PolicyChatData 对话请求数据
type PolicyChatData struct {
	ChatID         string `json:"chatId,omitempty"`         // 会话ID，首次调用为空
	ConversationID string `json:"conversationId,omitempty"` // 流水号，首次调用为空
	Stream         bool   `json:"stream"`                   // 流式访问类型
	RealName       bool   `json:"realName"`                 // 是否实名
	Message        string `json:"message"`                  // 消息内容
	MegType        string `json:"megType"`                  // 消息类型
	AAC001         string `json:"aac001,omitempty"`         // 个人编号，realName为true时必输
	AAC147         string `json:"aac147,omitempty"`         // 身份证号，realName为true时必输
	AAC003         string `json:"aac003,omitempty"`         // 姓名，realName为true时必输
	ReqType        string `json:"reqtype"`                  // 请求类型，值：1、政策咨询
}

// PolicyChatResponse 政策咨询对话响应
type PolicyChatResponse struct {
	Code    int                `json:"code"`    // 响应编码，成功：200
	Message string             `json:"message"` // 响应信息
	Data    *PolicyChatResData `json:"data"`    // 响应数据
}

// PolicyChatResData 对话响应数据
type PolicyChatResData struct {
	ChatID         string      `json:"chatId"`                   // 会话ID
	Message        string      `json:"message"`                  // 消息内容
	ConversationID string      `json:"conversationId,omitempty"` // 流水号
	MegType        string      `json:"megType"`                  // 消息类型
	Data           interface{} `json:"data,omitempty"`           // 查询数据结果
}
