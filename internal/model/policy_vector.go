package model

// PolicyInfo 政策信息
type PolicyInfo struct {
	ID                string `json:"id"`
	Zcmc              string `json:"zcmc"`              // 政策名称
	Type              string `json:"type"`              // 类型
	ZcLevel           string `json:"zcLevel"`           // 政策级别
	SourceUnit        string `json:"sourceUnit"`        // 来源单位
	PublishTime       string `json:"publishTime"`       // 发布时间
	PolicyExplanation string `json:"policyExplanation"` // 政策说明
	ApplicableObjects string `json:"applicableObjects"` // 适用对象
	Zclx              string `json:"zclx"`              // 政策类型
	ApplyCondition    string `json:"applyCondition"`    // 申请条件
	Zczc              string `json:"zczc"`              // 政策支持
	Phone             string `json:"phone"`             // 联系电话
	Remarks           string `json:"remarks"`           // 备注
	Btbz              string `json:"btbz"`              // 补贴标准
	Sqcl              string `json:"sqcl"`              // 申请材料
	Jbqd              string `json:"jbqd"`              // 经办渠道
	Zcsylx            string `json:"zcsylx"`            // 政策所属类型
	Jyzcbq            string `json:"jyzcbq"`            // 就业政策标签
	Gjcbq             string `json:"gjcbq"`             // 关键词标签
}

// PolicyResponse 政策API响应
type PolicyResponse struct {
	Total int          `json:"total"`
	Rows  []PolicyInfo `json:"rows"`
	Code  int          `json:"code"`
	Msg   string       `json:"msg"`
}

// PolicyVector 政策向量存储结构
type PolicyVector struct {
	ID      string    `json:"id"`      // 政策ID
	Content string    `json:"content"` // 政策文本内容
	Vector  []float32 `json:"vector"`  // 向量
}

// EmbeddingRequest Embedding请求
type EmbeddingRequest struct {
	Inputs string `json:"inputs"`
}

// EmbeddingResponse Embedding响应（嵌套数组格式）
type EmbeddingResponse [][]float32
