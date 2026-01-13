# 青岛岗位匹配系统 - 代码架构文档

## 项目概述

这是一个基于 Go 语言开发的智能岗位匹配系统，提供兼容 OpenAI `/v1/chat/completions` API 的接口。系统集成了 LLM 大模型、高德地图、OCR 文件解析、政策咨询等多种服务。

## 目录结构

```
qd-sc/
├── cmd/server/main.go          # 应用入口
├── internal/                   # 内部包（不对外暴露）
│   ├── api/                    # API 层
│   │   ├── handler/            # HTTP 请求处理器
│   │   └── middleware/         # 中间件
│   ├── client/                 # 外部服务客户端
│   ├── config/                 # 配置管理
│   ├── model/                  # 数据模型定义
│   └── service/                # 业务逻辑层
├── pkg/                        # 公共包（可对外暴露）
│   ├── metrics/                # 性能指标收集
│   └── utils/                  # 工具函数
└── config.yaml                 # 配置文件
```

---

## 模块详解

### 1. 入口模块 (`cmd/server/main.go`)

**职责**：应用启动、依赖注入、路由配置、优雅关闭

**核心流程**：
1. 加载配置文件（支持 `-config` 参数指定路径）
2. 初始化各类客户端（LLM、高德、岗位、OCR）
3. 初始化服务层（位置、岗位、政策、聊天）
4. 配置 Gin 路由和中间件
5. 启动 HTTP 服务器
6. 监听系统信号，实现优雅关闭

**路由配置**：
| 路径 | 方法 | 说明 |
|------|------|------|
| `/` | GET | API 信息 |
| `/health` | GET | 健康检查 |
| `/metrics` | GET | 性能指标 |
| `/v1/chat/completions` | POST | 聊天接口（主接口） |
| `/debug/pprof/*` | GET | 性能分析 |

---

### 2. 配置模块 (`internal/config/`)

**文件**：`config.go`

**职责**：配置加载、环境变量覆盖、默认值设置

**配置结构**：
```go
type Config struct {
    Server      ServerConfig      // 服务器配置（端口、超时）
    LLM         LLMConfig         // LLM 配置（API地址、密钥、模型）
    Amap        AmapConfig        // 高德地图配置
    JobAPI      JobAPIConfig      // 岗位 API 配置
    OCR         OCRConfig         // OCR 服务配置
    Policy      PolicyConfig      // 政策咨询配置
    Logging     LoggingConfig     // 日志配置
    Performance PerformanceConfig // 性能配置
}
```

**环境变量支持**：
- `LLM_API_KEY` / `LLM_BASE_URL` / `LLM_MODEL`
- `AMAP_API_KEY`
- `OCR_BASE_URL`
- `POLICY_LOGIN_NAME` / `POLICY_USER_KEY` / `POLICY_SERVICE_ID`
- `SERVER_PORT`

---

### 3. 数据模型 (`internal/model/`)

#### 3.1 `openai.go` - OpenAI 兼容模型

定义与 OpenAI API 兼容的请求/响应结构：

- `ChatCompletionRequest` - 聊天请求
- `ChatCompletionResponse` - 聊天响应
- `ChatCompletionChunk` - 流式响应块
- `Message` - 消息结构（支持文本和多模态）
- `Tool` / `ToolCall` - 工具调用相关
- `ErrorResponse` - 错误响应

#### 3.2 `job.go` - 岗位相关模型

- `JobQueryRequest` - 岗位查询请求参数
- `JobAPIResponse` - 岗位 API 原始响应
- `JobListing` - 单个岗位信息
- `FormattedJob` - 格式化后的岗位信息
- `EducationMap` / `ExperienceMap` / `LocationMap` - 代码映射表

#### 3.3 `policy.go` - 政策咨询模型

- `PolicyTicketRequest/Response` - Ticket 获取
- `PolicyChatRequest/Response` - 政策对话

#### 3.4 `tool.go` - 工具定义

**`GetAvailableTools()`** - 返回所有可用工具：

| 工具名 | 功能 |
|--------|------|
| `queryLocation` | 查询地点经纬度（高德地图） |
| `queryJobsByArea` | 按区域代码查询岗位 |
| `queryJobsByLocation` | 按经纬度查询附近岗位 |
| `parsePDF` | 解析 PDF 文件 |
| `parseImage` | 解析图片文件 |
| `queryPolicy` | 政策咨询 |

**`SystemPrompt`** - 系统提示词，定义 AI 助手的行为规范：
- 强制调用工具获取岗位数据，禁止编造
- 工具链思维模式
- 输出格式要求

---

### 4. 客户端层 (`internal/client/`)

#### 4.1 `http_client.go` - HTTP 客户端工厂

提供优化的 HTTP 客户端配置：
- 连接池管理
- HTTP/2 支持
- 超时配置
- 缓冲区优化

```go
NewHTTPClient(config)     // 通用客户端
NewLLMHTTPClient(timeout) // LLM 专用（更大连接池）
```

#### 4.2 `llm_client.go` - LLM 客户端

**方法**：
- `ChatCompletion()` - 非流式请求（带重试机制）
- `ChatCompletionStream()` - 流式请求（返回 channel）

**特性**：
- 自动重试（5xx 和 429 错误）
- SSE 流式解析
- 大缓冲区防止超大响应

#### 4.3 `amap_client.go` - 高德地图客户端

**方法**：
- `SearchPlace(keywords)` - 搜索地点
- `GetLocationCoordinates(keywords)` - 获取经纬度

#### 4.4 `job_client.go` - 岗位 API 客户端

**方法**：
- `QueryJobs(req)` - 查询岗位列表
- `FormatJobResponse(apiResp)` - 格式化响应（代码转文字）

#### 4.5 `ocr_client.go` - OCR 服务客户端

**方法**：
- `ParseURL(fileURL)` - 解析远程文件（图片/PDF/Excel/PPT）

#### 4.6 `policy_client.go` - 政策咨询客户端

**方法**：
- `GetTicket()` - 获取访问票据（自动缓存，1小时有效）
- `Chat(chatReq)` - 非流式对话
- `ChatStream(chatReq)` - 流式对话

---

### 5. 服务层 (`internal/service/`)

#### 5.1 `chat_service.go` - 核心聊天服务

**职责**：消息处理、工具调用编排、流式输出

**核心方法**：

```go
ProcessChatRequest(req)       // 非流式处理
ProcessChatRequestStream(ctx, req) // 流式处理
```

**关键特性**：

1. **岗位意图检测** (`isJobQueryIntent`)
   - 关键词匹配检测用户是否在查询岗位
   - 检测到岗位意图时强制调用工具

2. **幻觉检测** (`containsJobHallucination`)
   - 正则匹配检测 AI 是否自行编造岗位信息
   - 拦截幻觉输出，强制重新调用工具

3. **简历检测** (`isResumeContent`)
   - 关键词匹配判断 OCR 内容是否为简历
   - 非简历内容会提示先询问用户意图

4. **工具调用循环**
   - 最多 10 轮工具调用
   - 自动合并流式响应中的分块工具调用
   - 岗位结果分块输出（每个岗位间隔 1 秒）

5. **消息预处理** (`prepareMessages`)
   - 注入系统提示词
   - 处理 Vision API 格式的文件 URL
   - 自动调用 OCR 解析文件

#### 5.2 `job_service.go` - 岗位服务

**方法**：
- `QueryJobsByArea(params)` - 按区域查询
- `QueryJobsByLocation(params)` - 按位置查询

#### 5.3 `location_service.go` - 位置服务

**方法**：
- `QueryLocation(keywords)` - 查询地点坐标

#### 5.4 `policy_service.go` - 政策服务

**方法**：
- `QueryPolicy(...)` - 政策咨询（支持多轮对话和实名咨询）

---

### 6. API 处理器 (`internal/api/handler/`)

#### 6.1 `chat.go` - 聊天处理器

**方法**：
- `ChatCompletions(c)` - 主接口处理
  - 验证请求参数
  - 根据 `stream` 参数分发到流式/非流式处理

**流式响应**：
- SSE 格式输出
- 支持客户端断开检测
- 错误时发送错误 chunk

#### 6.2 `health.go` - 健康检查

返回服务状态信息。

#### 6.3 `metrics.go` - 指标处理器

返回性能统计数据。

#### 6.4 `response.go` - 统一响应

提供 `Error()` 和 `Success()` 方法统一响应格式。

---

### 7. 中间件 (`internal/api/middleware/`)

#### 7.1 `cors.go` - 跨域处理

- 支持携带凭证的请求
- 预检请求处理
- 24 小时缓存

#### 7.2 `ratelimit.go` - 限流

**算法**：令牌桶（Token Bucket）

**配置**：
- 桶容量：200（最大突发）
- 补充速率：50/秒（持续 QPS）

**实现**：CAS 无锁算法，高并发友好

#### 7.3 `recovery.go` - Panic 恢复

捕获 panic，打印堆栈，返回 500 错误。

#### 7.4 `metrics.go` - 指标收集

记录请求数、延迟、失败数等指标。

---

### 8. 工具包 (`pkg/`)

#### 8.1 `metrics/metrics.go` - 性能指标

**收集指标**：
- 请求计数（总数、活跃、失败、流式）
- 延迟统计（平均、最小、最大）
- 系统指标（goroutine、内存、GC）

**实现**：原子操作 + sync.Map，无锁高性能

#### 8.2 `utils/json.go` - JSON 工具

`ToJSONStringPretty()` - 格式化 JSON 输出

---

## 数据流图

```
用户请求
    │
    ▼
┌─────────────────┐
│   Gin Router    │
│   + 中间件      │
│  (CORS/限流/恢复)│
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  ChatHandler    │
│  (请求验证)      │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  ChatService    │◄──────────────────────┐
│  (核心编排)      │                       │
└────────┬────────┘                       │
         │                                │
    ┌────┴────┬────────┬────────┐        │
    ▼         ▼        ▼        ▼        │
┌───────┐ ┌───────┐ ┌───────┐ ┌───────┐  │
│LLM    │ │Job    │ │Location│ │Policy │  │
│Client │ │Service│ │Service │ │Service│  │
└───┬───┘ └───┬───┘ └───┬───┘ └───┬───┘  │
    │         │        │         │       │
    ▼         ▼        ▼         ▼       │
┌───────┐ ┌───────┐ ┌───────┐ ┌───────┐  │
│OpenAI │ │岗位API │ │高德API │ │政策API │  │
│  API  │ │       │ │       │ │       │  │
└───────┘ └───────┘ └───────┘ └───────┘  │
    │                                     │
    └─────────────────────────────────────┘
              (工具调用循环)
```

---

## 关键设计

### 1. 工具调用机制

系统通过 LLM 的 Function Calling 能力实现智能工具调用：

1. 用户发送消息
2. LLM 分析意图，决定调用哪些工具
3. 服务端执行工具调用，获取真实数据
4. 将工具结果返回给 LLM
5. LLM 基于真实数据生成回复

### 2. 幻觉防护

为防止 LLM 编造岗位信息：

1. **意图检测**：检测到岗位查询时设置 `tool_choice=required`
2. **输出检测**：正则匹配检测疑似编造的岗位信息
3. **强制重试**：检测到幻觉时强制重新调用工具

### 3. 流式输出优化

- 岗位信息分块输出，每个岗位间隔 1 秒
- 使用 `job-json` 代码块格式，便于前端渲染
- 过滤 tool_calls 相关 chunk，对客户端透明

### 4. 连接池优化

- LLM 客户端：200 最大连接，100/host
- 其他客户端：100 最大连接，50/host
- 启用 HTTP/2
- 32KB 读写缓冲

---

## 扩展指南

### 添加新工具

1. 在 `model/tool.go` 的 `GetAvailableTools()` 中添加工具定义
2. 在 `service/chat_service.go` 的 `executeToolCall()` 中添加处理分支
3. 实现对应的处理函数

### 添加新的外部服务

1. 在 `config/config.go` 中添加配置结构
2. 在 `client/` 下创建新的客户端
3. 在 `service/` 下创建对应的服务层
4. 在 `main.go` 中初始化并注入
