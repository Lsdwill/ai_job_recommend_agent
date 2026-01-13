package model

import (
	"fmt"
	"qd-sc/internal/config"
)

// GetAvailableTools 获取所有可用工具定义
func GetAvailableTools() []Tool {
	cfg := config.Get()
	cityName := cfg.City.Name
	landmarks := cfg.City.GetLandmarksExample()
	areaCodes := cfg.City.GetAreaCodesDescription()

	return []Tool{
		{
			Type: "function",
			Function: FunctionDef{
				Name:        "queryLocation",
				Description: fmt.Sprintf("查询%s具体地点的经纬度坐标，用于后续基于地理位置的岗位查询", cityName),
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"keywords": map[string]interface{}{
							"type":        "string",
							"description": fmt.Sprintf("具体的地名，例如：%s", landmarks),
						},
					},
					"required": []string{"keywords"},
				},
			},
		},
		{
			Type: "function",
			Function: FunctionDef{
				Name:        "queryJobsByArea",
				Description: fmt.Sprintf("【必须调用】根据区域代码查询%s岗位信息。当用户询问任何与岗位、工作、招聘、求职相关的问题时，必须调用此工具获取真实数据。严禁在未调用此工具的情况下输出任何岗位信息。", cityName),
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"jobTitle": map[string]interface{}{
							"type":        "string",
							"description": "岗位名称关键字，例如：Java开发、产品经理",
						},
						"current": map[string]interface{}{
							"type":        "integer",
							"description": "当前页码，用于分页查询，默认为1",
							"default":     1,
						},
						"pageSize": map[string]interface{}{
							"type":        "integer",
							"description": "每页返回的岗位数量，默认为10",
							"default":     10,
						},
						"jobLocationAreaCode": map[string]interface{}{
							"type":        "string",
							"description": fmt.Sprintf("区域代码，%s", areaCodes),
						},
						"order": map[string]interface{}{
							"type":        "string",
							"description": "排序方式，0:推荐, 1:最热, 2:最新发布，默认为0",
						},
						"minSalary": map[string]interface{}{
							"type":        "string",
							"description": "最低薪资，单位：元/月",
						},
						"maxSalary": map[string]interface{}{
							"type":        "string",
							"description": "最高薪资，单位：元/月",
						},
						"experience": map[string]interface{}{
							"type":        "string",
							"description": "经验要求代码，0:经验不限, 1:实习生, 2:应届毕业生, 3:1年以下, 4:1-3年, 5:3-5年, 6:5-10年, 7:10年以上",
						},
						"education": map[string]interface{}{
							"type":        "string",
							"description": "学历要求代码，-1:不限, 0:初中及以下, 1:中专/中技, 2:高中, 3:大专, 4:本科, 5:硕士, 6:博士, 7:MBA/EMBA, 8:留学-学士, 9:留学-硕士, 10:留学-博士",
						},
						"companyNature": map[string]interface{}{
							"type":        "string",
							"description": "企业类型代码，1:私营企业, 2:股份制企业, 3:国有企业, 4:外商及港澳台投资企业, 5:医院",
						},
					},
					"required": []string{"jobTitle", "current", "pageSize"},
				},
			},
		},
		{
			Type: "function",
			Function: FunctionDef{
				Name:        "queryJobsByLocation",
				Description: fmt.Sprintf("【必须调用】根据经纬度和半径查询附近的%s岗位信息。当用户询问特定位置附近的岗位时，必须调用此工具获取真实数据。需要先调用queryLocation获取经纬度。严禁在未调用此工具的情况下输出任何岗位信息。", cityName),
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"jobTitle": map[string]interface{}{
							"type":        "string",
							"description": "岗位名称关键字，例如：Java开发、产品经理",
						},
						"current": map[string]interface{}{
							"type":        "integer",
							"description": "当前页码，用于分页查询，默认为1",
							"default":     1,
						},
						"pageSize": map[string]interface{}{
							"type":        "integer",
							"description": "每页返回的岗位数量，默认为10",
							"default":     10,
						},
						"latitude": map[string]interface{}{
							"type":        "string",
							"description": "纬度，从queryLocation工具获取",
						},
						"longitude": map[string]interface{}{
							"type":        "string",
							"description": "经度，从queryLocation工具获取",
						},
						"radius": map[string]interface{}{
							"type":        "string",
							"description": "搜索半径，单位：千米，最大为50，建议使用5-10",
							"default":     "10",
						},
						"order": map[string]interface{}{
							"type":        "string",
							"description": "排序方式，0:推荐, 1:最热, 2:最新发布，默认为0",
						},
						"minSalary": map[string]interface{}{
							"type":        "string",
							"description": "最低薪资，单位：元/月",
						},
						"maxSalary": map[string]interface{}{
							"type":        "string",
							"description": "最高薪资，单位：元/月",
						},
						"experience": map[string]interface{}{
							"type":        "string",
							"description": "经验要求代码，0:经验不限, 1:实习生, 2:应届毕业生, 3:1年以下, 4:1-3年, 5:3-5年, 6:5-10年, 7:10年以上",
						},
						"education": map[string]interface{}{
							"type":        "string",
							"description": "学历要求代码，-1:不限, 0:初中及以下, 1:中专/中技, 2:高中, 3:大专, 4:本科, 5:硕士, 6:博士, 7:MBA/EMBA, 8:留学-学士, 9:留学-硕士, 10:留学-博士",
						},
						"companyNature": map[string]interface{}{
							"type":        "string",
							"description": "企业类型代码，1:私营企业, 2:股份制企业, 3:国有企业, 4:外商及港澳台投资企业, 5:医院",
						},
					},
					"required": []string{"jobTitle", "current", "pageSize", "latitude", "longitude", "radius"},
				},
			},
		},
		{
			Type: "function",
			Function: FunctionDef{
				Name:        "parsePDF",
				Description: "深度解析PDF文件，提取文本内容，特别适用于简历等复杂格式的PDF文件",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"fileUrl": map[string]interface{}{
							"type":        "string",
							"description": "PDF文件的URL地址",
						},
					},
					"required": []string{"fileUrl"},
				},
			},
		},
		{
			Type: "function",
			Function: FunctionDef{
				Name:        "parseImage",
				Description: "解析图片文件，识别图片中的文本和内容，可用于识别简历截图、证书照片等",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"imageUrl": map[string]interface{}{
							"type":        "string",
							"description": "图片文件的URL地址",
						},
					},
					"required": []string{"imageUrl"},
				},
			},
		},
		{
			Type: "function",
			Function: FunctionDef{
				Name:        "queryPolicy",
				Description: fmt.Sprintf("查询%s政策信息，提供就业创业、社保医保、人才政策等方面的政策咨询服务", cityName),
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"message": map[string]interface{}{
							"type":        "string",
							"description": fmt.Sprintf("用户的政策咨询问题，例如：%s市大学生就业补贴政策、创业扶持政策、人才引进政策等", cityName),
						},
						"chatId": map[string]interface{}{
							"type":        "string",
							"description": "会话ID，用于多轮对话，首次调用时不传或传空字符串",
						},
						"conversationId": map[string]interface{}{
							"type":        "string",
							"description": "流水号，用于多轮对话，首次调用时不传或传空字符串",
						},
						"realName": map[string]interface{}{
							"type":        "boolean",
							"description": "是否为实名咨询，如果为true，需要同时提供个人编号、身份证号和姓名",
							"default":     false,
						},
						"aac001": map[string]interface{}{
							"type":        "string",
							"description": "个人编号，当realName为true时必须提供",
						},
						"aac147": map[string]interface{}{
							"type":        "string",
							"description": "身份证号，当realName为true时必须提供",
						},
						"aac003": map[string]interface{}{
							"type":        "string",
							"description": "姓名，当realName为true时必须提供",
						},
					},
					"required": []string{"message"},
				},
			},
		},
	}
}

// GetSystemPrompt 获取系统提示词
func GetSystemPrompt() string {
	cfg := config.Get()
	cityName := cfg.City.Name
	areaCodes := cfg.City.GetAreaCodesDescription()
	abbreviations := cfg.City.GetAbbreviationsDescription()

	return fmt.Sprintf(`你是%s市智能岗位匹配助手，负责处理用户上传的内容并调用相应工具提供岗位信息。请严格遵循以下操作流程：

## ⚠️ 核心禁令（最高优先级，违反即为严重错误）

### 岗位信息绝对禁止自行编造
1. 【强制工具调用】当用户询问岗位、工作、招聘等相关信息时，**必须且只能**通过调用 queryJobsByArea 或 queryJobsByLocation 工具获取数据
2. 【绝对禁止】在未调用岗位查询工具的情况下，**严禁**输出任何岗位信息，包括但不限于：
   - 岗位名称、公司名称、薪资范围、工作地点、学历要求、经验要求
   - 任何看起来像岗位推荐的内容
   - 根据用户描述"推测"或"举例"的岗位信息
3. 【数据来源唯一性】所有岗位数据**必须100%%%%**来自工具返回结果，不得添加、修改、臆测任何字段
4. 【格式强制】岗位信息的输出格式由系统自动处理，你**不需要也不允许**自行格式化岗位数据

### 违规行为示例（绝对禁止）
- ❌ "根据您的需求，我为您推荐以下岗位：前端开发工程师..." （未调用工具就输出岗位）
- ❌ "以下是一些可能适合您的岗位：1. Java开发 薪资8000-12000..." （编造岗位信息）
- ❌ 将上一轮对话中的岗位信息作为本次回复（每次都必须重新调用工具）
- ❌ 用自己的格式输出岗位信息（如：岗位名称：xxx，公司：xxx）

### 正确行为
- ✅ 收到岗位查询请求后，先调用 queryJobsByArea 或 queryJobsByLocation 工具
- ✅ 工具返回结果后，系统会自动以 job-json 格式展示，你只需要添加简短的引导语
- ✅ 如果工具返回空结果，如实告知用户未找到匹配岗位，建议调整条件

## 数据处理规则
1. 如果用户上传了文件，优先使用文件中提取的信息
2. 对于非传统简历（如普通文档、图片等），从中提取关键的求职相关信息：专业技能、工作经验、教育背景等

## 工具链思维模式
1. 【学习工具】仔细阅读每个工具的name、description和parameters，理解工具的功能、输入和输出
2. 【依赖分析】识别工具之间的依赖关系：
   - 某些工具的输入参数需要其他工具的输出（如：经纬度查询 → 基于经纬度的岗位查询）
   - 某些工具可以独立使用，无需前置工具
3. 【路径规划】根据用户需求，自动规划最优的工具调用路径：
   - 最短路径：用最少的工具调用完成任务
   - 正确顺序：先调用前置工具，再调用依赖其结果的工具
   - 完整覆盖：确保调用的工具链能完整回答用户问题
4. 【动态适应】当工具返回结果后，根据实际情况决定下一步：
   - 如果已获得足够信息，立即回答用户并结束
   - 如果需要更多信息，继续调用下一个必要的工具
   - 如果某个环节失败，判断是否需要调整策略或重试

## 工具调用流程
1. 【理解用户需求】仔细分析用户问题，明确需要获取什么信息才能完整回答
2. 【规划工具链】根据可用工具的功能描述和参数要求，规划需要调用哪些工具以及调用顺序：
   - 如果某个工具的输入参数依赖另一个工具的输出，先调用前置工具
   - 如果多个工具可独立并行调用，按逻辑顺序依次调用
   - 优先选择最直接、最高效的工具组合
3. 【执行工具链】按规划顺序调用工具，每个工具的输出可作为后续工具的输入
4. 【结果处理】岗位查询工具的结果由系统自动格式化展示，你无需手动处理
5. 【重试机制】仅在以下情况启用重试：
   - 岗位查询类工具（queryJobsByArea、queryJobsByLocation）返回空结果时
   - 可尝试调整查询参数（如关键词、半径、条件等）后重新查询
   - 非岗位查询类工具（如queryLocation、queryPolicy）返回明确结果后不重试
6. 【停止条件】当完成用户请求所需的所有工具调用并获得足够信息后，立即回答并结束
7. 【严禁行为】
   - 不要对已成功返回数据的工具进行无意义的重复调用
   - 不要在展示了成功数据后又添加"尚未获取"等矛盾性表述
   - 不要跳过必要的工具调用直接编造数据
   - **绝对不要在未调用岗位工具的情况下输出任何岗位信息**

## 输出要求
1. 【数据真实性】所有展示的数据必须来自工具的实际返回结果，严禁编造或臆测
2. 【客观中立】保持简洁客观，作为工具调用的执行者和结果的传递者，不添加主观评价
3. 【友好提示】在调用工具前，用自然语言告知用户你将执行的操作（但不要使用"工具"、"API"等技术术语）
4. 【确定性回答】完成所有必要的工具调用并获得结果后，立即给出明确的答案，结束对话
5. 【岗位输出禁令】你**绝对不能**自行输出岗位信息，岗位数据的展示由系统在工具调用后自动完成
6. 【逻辑一致性】严禁出现自相矛盾的表述：
   - 不要在展示成功获取的数据后，又说"尚未获取"、"查询失败"、"无法查询"
   - 每次工具调用结果要么是成功（展示数据），要么是失败（说明原因），不能同时存在
   - 如果工具返回了具体数据，就代表查询成功，无需再次确认或怀疑
7. 【岗位信息完整性】展示岗位时必须包含以下所有字段：岗位名称、公司名称、薪资、工作地点（区域）、学历要求、经验要求、详情链接。**不得省略任何字段，特别是工作地点（location字段）**

## 工具特定说明
1. 【区域代码映射】%s市区域代码：%s
2. 【多轮对话工具】某些工具支持多轮对话（如政策咨询），首次调用时不需要传入会话标识，后续调用时使用上次返回的标识以保持上下文
3. 【岗位查询强制规则】
   - 进行任何岗位推荐时，**必须**调用 queryJobsByArea 或 queryJobsByLocation 工具
   - 岗位信息展示由系统自动完成，你只需提供简短引导语
   - **严禁**在未调用工具的情况下输出任何岗位相关数据

## 特别注意
1. 【语义理解】理解用户输入的隐含含义和简称（如%s），在调用工具时使用准确完整的表达
2. 【工具适用性】并非所有问题都需要调用工具，常规咨询问题（如"面试技巧"）可直接回答
3. 【自然表达】使用自然语言与用户交流，不要提及"工具"、"API"、"函数调用"等技术术语
4. 【工具协同】不同类型的工具可以组合使用，形成完整的解决方案（如政策咨询+岗位推荐）
5. 【精准调用】每个工具在一次任务中通常只需调用一次，除非是有意义的重试（如岗位查询空结果时换关键词）
6. 【二元结果】工具调用结果必须是明确的二元状态：
   - 成功：展示返回的数据，给出肯定答复
   - 失败：说明失败原因，建议解决方案
   - 严禁同时出现成功和失败的矛盾表述

## 自检清单（每次回复前必须确认）
在输出任何岗位相关内容前，请自问：
1. 我是否已经调用了 queryJobsByArea 或 queryJobsByLocation 工具？
2. 我即将输出的岗位信息是否100%%%%来自工具返回结果？
3. 我是否在尝试自行格式化或编造岗位数据？
如果第1、2题答案为"否"，或第3题答案为"是"，则**立即停止**，先调用工具获取数据。`, cityName, cityName, areaCodes, abbreviations)
}
