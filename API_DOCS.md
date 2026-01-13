# 青岛岗位匹配系统 API 调用文档

> **版本**: 1.0.0  
> **更新日期**: 2025-11  
> **协议**: 兼容 OpenAI Chat Completions API

---

## 目录

- [1. 系统概述](#1-系统概述)
- [2. 系统架构](#2-系统架构)
- [3. 快速开始](#3-快速开始)
- [4. API 端点列表](#4-api-端点列表)
- [5. 核心接口详解](#5-核心接口详解)
  - [5.1 聊天补全接口](#51-聊天补全接口)
  - [5.2 健康检查接口](#52-健康检查接口)
  - [5.3 性能指标接口](#53-性能指标接口)
  - [5.4 性能分析接口](#54-性能分析接口)
- [6. 内置工具说明](#6-内置工具说明)
- [7. 代码对照表](#7-代码对照表)
- [8. SDK 与代码示例](#8-sdk-与代码示例)
- [9. 错误处理](#9-错误处理)
- [10. 配置参考](#10-配置参考)
- [11. 部署指南](#11-部署指南)
- [12. 常见问题](#12-常见问题)

---

## 1. 系统概述

### 1.1 系统简介

青岛岗位匹配系统是一个基于 Go 语言开发的智能岗位推荐服务，提供与 OpenAI `/v1/chat/completions` 完全兼容的 API 接口。系统集成了多种智能工具，能够自动理解用户意图并调用相应功能。

### 1.2 核心功能

| 功能模块 | 说明 |
|---------|------|
| 🎯 **岗位推荐** | 根据用户需求、简历内容智能推荐青岛市岗位 |
| 📍 **地理位置查询** | 集成高德地图，支持按位置、区域搜索岗位 |
| 📄 **文件解析** | 支持 PDF、图片、Excel、PPT、Word 等文件的 OCR 智能解析 |
| 🖼️ **Vision API** | 兼容 OpenAI Vision API 格式，支持通过 URL 发送图片/PDF/Excel/PPT 等文件进行 OCR 识别 |
| 📋 **政策咨询** | 提供就业创业、社保医保、人才政策等咨询服务 |
| 🔄 **多轮对话** | 支持上下文连续对话 |
| ⚡ **流式输出** | 支持 SSE 流式响应，提升用户体验 |

### 1.3 技术特性

- **兼容性**: 100% 兼容 OpenAI Chat API 格式
- **高性能**: 基于 Gin 框架，支持高并发
- **限流保护**: 内置令牌桶限流（200 容量/50 QPS）
- **优雅关闭**: 支持信号处理和优雅停机
- **可观测性**: 内置 JSON 指标端点和 pprof 性能分析（可通过配置开关关闭）

---

## 2. 系统架构

### 2.1 架构图

```
┌─────────────────────────────────────────────────────────────────┐
│                        客户端 (Client)                          │
│            (Web/App/第三方 OpenAI 兼容客户端)                    │
└─────────────────────────────┬───────────────────────────────────┘
                              │ HTTP/HTTPS
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                      API 网关层 (Gateway)                        │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌──────────┐  │
│  │  CORS   │ │ Recovery│ │ Metrics │ │  Rate   │ │  Logger  │  │
│  │Middleware│ │Middleware│ │Middleware│ │ Limit  │ │          │  │
│  └─────────┘ └─────────┘ └─────────┘ └─────────┘ └──────────┘  │
└─────────────────────────────┬───────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                      处理器层 (Handlers)                         │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │ ChatHandler  │  │HealthHandler │  │MetricsHandler│          │
│  └──────────────┘  └──────────────┘  └──────────────┘          │
└─────────────────────────────┬───────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                      服务层 (Services)                           │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐               │
│  │ ChatService │ │ JobService  │ │FileService  │               │
│  └─────────────┘ └─────────────┘ └─────────────┘               │
│  ┌─────────────┐ ┌─────────────┐                               │
│  │LocationServ │ │PolicyService│                               │
│  └─────────────┘ └─────────────┘                               │
└─────────────────────────────┬───────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                      客户端层 (Clients)                          │
│  ┌───────────┐ ┌───────────┐ ┌───────────┐ ┌───────────┐       │
│  │ LLMClient │ │ JobClient │ │AmapClient │ │ OCRClient │       │
│  └─────┬─────┘ └─────┬─────┘ └─────┬─────┘ └─────┬─────┘       │
│        │             │             │             │              │
└────────┼─────────────┼─────────────┼─────────────┼──────────────┘
         │             │             │             │
         ▼             ▼             ▼             ▼
┌──────────────┐ ┌──────────────┐ ┌──────────┐ ┌──────────────┐
│   LLM API    │ │  岗位 API    │ │ 高德地图  │ │  OCR 服务   │
│ (OpenAI兼容) │ │              │ │   API    │ │             │
└──────────────┘ └──────────────┘ └──────────┘ └──────────────┘
```

### 2.2 数据流

```
用户请求 → 中间件处理 → ChatHandler → ChatService 
                                          ↓
                               识别意图 & 调用工具
                                          ↓
                         ┌────────────────┼────────────────┐
                         ↓                ↓                ↓
                   LocationService   JobService    PolicyService
                         ↓                ↓                ↓
                    高德地图API       岗位API         政策API
                         ↓                ↓                ↓
                         └────────────────┼────────────────┘
                                          ↓
                               整合结果 & 生成回复
                                          ↓
                              流式/非流式响应 → 用户
```

---

## 3. 快速开始

### 3.1 环境要求

- Go 1.20+
- 或 Docker 20.10+

### 3.2 安装运行

#### 方式一：源码运行

```bash
# 克隆项目
git clone <repository-url>
cd qd-sc

# 安装依赖
go mod download

# 配置文件（编辑 config.yaml）
cp config.yaml config.local.yaml
vim config.local.yaml

# 运行
go run cmd/server/main.go

# 或编译后运行
go build -o qd-sc-server cmd/server/main.go
./qd-sc-server -config=config.yaml
```

#### 方式二：Docker 运行

```bash
# 使用 docker-compose
docker-compose up -d

# 或手动构建
docker build -t qd-sc-server .
docker run -d -p 8080:8080 \
  -e LLM_API_KEY="sk-xxx" \
  -e LLM_BASE_URL="https://your-api.com/v1" \
  --name qd-sc-server \
  qd-sc-server
```

### 3.3 验证服务

```bash
# 健康检查
curl http://localhost:8080/health

# 测试对话（流式）
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "qd-job-turbo",
    "messages": [{"role": "user", "content": "推荐城阳区的Java开发岗位"}],
    "stream": true
  }'
```

---

## 4. API 端点列表

| 端点 | 方法 | 说明 | 认证 |
|------|------|------|------|
| `/` | GET | API 信息和端点列表 | 无 |
| `/health` | GET | 健康检查 | 无 |
| `/metrics` | GET | 性能指标（JSON，需启用 `performance.enable_metrics`） | 无 |
| `/v1/chat/completions` | POST | **核心接口** - OpenAI 兼容的聊天接口 | 无 |
| `/debug/pprof/*` | GET | pprof 性能分析（需启用 `performance.enable_pprof`） | 无 |

---

## 5. 核心接口详解

### 5.1 聊天补全接口

#### 基本信息

| 属性 | 值 |
|------|-----|
| **端点** | `POST /v1/chat/completions` |
| **Content-Type** | `application/json` |
| **响应格式** | JSON 或 SSE (Server-Sent Events) |

#### 5.1.1 JSON 请求格式

```http
POST /v1/chat/completions HTTP/1.1
Host: localhost:8080
Content-Type: application/json

{
  "model": "qd-job-turbo",
  "messages": [
    {
      "role": "system",
      "content": "你是一个有帮助的助手"
    },
    {
      "role": "user",
      "content": "帮我找青岛城阳区的Java开发岗位"
    }
  ],
  "stream": true,
  "temperature": 0.7,
  "max_tokens": 2000
}
```

#### 5.1.2 请求参数说明

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| `model` | string | ✅ | - | 模型名称，固定为 `qd-job-turbo` |
| `messages` | array | ✅ | - | 消息数组，见下表 |
| `stream` | boolean | ❌ | `false` | 是否流式输出（**推荐 `true`**） |
| `temperature` | float | ❌ | `1.0` | 采样温度，范围 0-2 |
| `top_p` | float | ❌ | `1.0` | 核采样参数 |
| `max_tokens` | integer | ❌ | - | 最大生成 token 数 |
| `presence_penalty` | float | ❌ | `0.0` | 存在惩罚，范围 -2.0 到 2.0 |
| `frequency_penalty` | float | ❌ | `0.0` | 频率惩罚，范围 -2.0 到 2.0 |
| `user` | string | ❌ | - | 用户标识 |

#### 5.1.3 消息对象格式

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `role` | string | ✅ | 角色：`system`、`user`、`assistant` |
| `content` | string/array | ✅ | 消息内容（支持文本或多模态数组） |
| `name` | string | ❌ | 发送者名称 |

**role 说明**:

| 角色 | 说明 |
|------|------|
| `system` | 系统指令，设置 AI 行为 |
| `user` | 用户消息 |
| `assistant` | AI 回复 |

**content 格式说明**:

`content` 支持两种格式：

1. **字符串格式**（普通文本消息）：
```json
{
  "role": "user",
  "content": "帮我推荐Java开发岗位"
}
```

2. **数组格式**（多模态消息，支持图片 - OpenAI Vision API 兼容）：
```json
{
  "role": "user",
  "content": [
    {"type": "text", "text": "根据这份简历帮我推荐合适的岗位"},
    {"type": "image_url", "image_url": {"url": "https://example.com/resume.jpg"}}
  ]
}
```

**多模态内容类型**:

| type | 说明 | 字段 |
|------|------|------|
| `text` | 文本内容 | `text`: 文本字符串 |
| `image_url` | 文件URL（兼容字段） | `image_url.url`: 文件地址 |

> **重要说明**: 
> - `image_url` 是为兼容 OpenAI Vision API 而保留的字段名称
> - 实际上不仅支持图片，**还支持 PDF、Excel、PPT 等所有 OCR 服务支持的文件类型**
> - 系统会自动调用 OCR 服务进行识别解析
> - 支持的文件格式：JPG、PNG、GIF、PDF、XLS、XLSX、PPT、PPTX 等

#### 5.1.4 带文件URL的请求（Vision API 兼容格式）

完全兼容 OpenAI Vision API 格式，通过 `image_url` 字段发送文件 URL 进行 OCR 识别。

> **字段说明**: `image_url` 是 OpenAI Vision API 的标准字段名，为保持兼容性沿用此名称。但本系统通过 OCR 服务，**不仅支持图片，还支持 PDF、Excel、PPT 等多种文件格式**。

**示例 1：发送图片简历**

```http
POST /v1/chat/completions HTTP/1.1
Host: localhost:8080
Content-Type: application/json

{
  "model": "qd-job-turbo",
  "messages": [
    {
      "role": "user",
      "content": [
        {"type": "text", "text": "根据这份简历帮我推荐合适的岗位"},
        {"type": "image_url", "image_url": {"url": "https://example.com/resume.jpg"}}
      ]
    }
  ],
  "stream": true
}
```

**示例 2：发送 PDF 文件**

```json
{
  "role": "user",
  "content": [
    {"type": "text", "text": "分析这份PDF简历"},
    {"type": "image_url", "image_url": {"url": "https://example.com/resume.pdf"}}
  ]
}
```

**示例 3：发送 Excel 文件**

```json
{
  "role": "user",
  "content": [
    {"type": "text", "text": "帮我分析这个表格数据"},
    {"type": "image_url", "image_url": {"url": "https://example.com/data.xlsx"}}
  ]
}
```

**示例 4：发送多个文件**

```json
{
  "role": "user",
  "content": [
    {"type": "text", "text": "分析这些文档内容"},
    {"type": "image_url", "image_url": {"url": "https://example.com/resume.pdf"}},
    {"type": "image_url", "image_url": {"url": "https://example.com/certificate.jpg"}},
    {"type": "image_url", "image_url": {"url": "https://example.com/transcript.xlsx"}}
  ]
}
```

**支持的文件格式**:

| 类型 | 格式 |
|------|------|
| 图片 | JPG、JPEG、PNG、GIF |
| 文档 | PDF |
| 表格 | XLS、XLSX |
| 演示 | PPT、PPTX |

#### 5.1.6 流式响应格式 (SSE)

当 `stream: true` 时，响应为 Server-Sent Events 格式：

```
data: {"id":"chatcmpl-xxx","object":"chat.completion.chunk","created":1234567890,"model":"qd-job-turbo","choices":[{"index":0,"delta":{"role":"assistant","content":"您好"},"finish_reason":null}]}

data: {"id":"chatcmpl-xxx","object":"chat.completion.chunk","created":1234567890,"model":"qd-job-turbo","choices":[{"index":0,"delta":{"content":"，"},"finish_reason":null}]}

data: {"id":"chatcmpl-xxx","object":"chat.completion.chunk","created":1234567890,"model":"qd-job-turbo","choices":[{"index":0,"delta":{"content":"我"},"finish_reason":null}]}

data: {"id":"chatcmpl-xxx","object":"chat.completion.chunk","created":1234567890,"model":"qd-job-turbo","choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}

data: [DONE]
```

**Chunk 对象结构**:

```typescript
interface ChatCompletionChunk {
  id: string;                    // 响应ID
  object: "chat.completion.chunk";
  created: number;               // Unix 时间戳
  model: string;                 // 模型名称
  choices: Array<{
    index: number;
    delta: {
      role?: "assistant";        // 仅首个 chunk 包含
      content?: string;          // 内容片段
    };
    finish_reason: string | null; // "stop" 或 null
  }>;
}
```

#### 5.1.7 非流式响应格式

当 `stream: false` 时：

```json
{
  "id": "chatcmpl-xxx",
  "object": "chat.completion",
  "created": 1234567890,
  "model": "qd-job-turbo",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "您好，我可以帮您推荐岗位..."
      },
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 100,
    "completion_tokens": 50,
    "total_tokens": 150
  }
}
```

#### 5.1.8 岗位推荐特殊响应格式

当系统返回岗位推荐时，岗位信息会以特殊的 Markdown 代码块格式输出：

```
为您找到 3 个相关岗位：

``` job-json
{
  "jobTitle": "Java开发工程师",
  "companyName": "青岛XX科技有限公司",
  "salary": "15000-25000元/月",
  "location": "城阳区",
  "education": "本科",
  "experience": "3-5年",
  "appJobUrl": "https://..."
}
```

``` job-json
{
  "jobTitle": "高级Java工程师",
  "companyName": "青岛YY信息技术有限公司",
  "salary": "20000-35000元/月",
  "location": "城阳区",
  "education": "本科",
  "experience": "5-10年",
  "appJobUrl": "https://..."
}
```
```

**岗位对象结构**:

```typescript
interface FormattedJob {
  jobTitle: string;      // 职位名称
  companyName: string;   // 公司名称
  salary: string;        // 薪资范围
  location: string;      // 工作地点
  education: string;     // 学历要求
  experience: string;    // 经验要求
  appJobUrl: string;     // 职位详情链接
  data?: any;            // 额外数据（分页信息等）
}
```

---

### 5.2 健康检查接口

#### 请求

```http
GET /health HTTP/1.1
Host: localhost:8080
```

#### 响应

```json
{
  "status": "ok",
  "timestamp": "2024-01-01T12:00:00Z"
}
```

---

### 5.3 性能指标接口

> 该接口需要启用 `performance.enable_metrics`；关闭后将不会注册 `/metrics` 端点。

#### 请求

```http
GET /metrics HTTP/1.1
Host: localhost:8080
```

#### 响应

```json
{
  "requests_total": 12345,
  "requests_success": 12300,
  "requests_failed": 45,
  "avg_response_time_ms": 156.8,
  "goroutines": 42,
  "memory_alloc_mb": 45.6
}
```

| 指标 | 说明 |
|------|------|
| `requests_total` | 总请求数 |
| `requests_success` | 成功请求数 |
| `requests_failed` | 失败请求数 |
| `avg_response_time_ms` | 平均响应时间（毫秒） |
| `goroutines` | 当前 goroutine 数量 |
| `memory_alloc_mb` | 内存分配（MB） |

---

### 5.4 性能分析接口

> 该接口需要启用 `performance.enable_pprof`；关闭后将不会注册 `/debug/pprof/*` 端点。

支持 Go pprof 标准端点：

| 端点 | 说明 |
|------|------|
| `/debug/pprof/` | pprof 索引页面 |
| `/debug/pprof/profile?seconds=30` | CPU 性能分析 |
| `/debug/pprof/heap` | 堆内存分析 |
| `/debug/pprof/goroutine` | Goroutine 分析 |
| `/debug/pprof/allocs` | 内存分配分析 |

**使用示例**:

```bash
# 采集 30 秒 CPU 数据
curl http://localhost:8080/debug/pprof/profile?seconds=30 -o cpu.prof

# 分析
go tool pprof cpu.prof
```

---

## 6. 内置工具说明

系统内置了多种智能工具，会根据用户意图自动调用：

### 6.1 queryLocation - 地理位置查询

**功能**: 查询青岛具体地点的经纬度坐标

**触发场景**: 用户提到具体地点名称时（如"五四广场附近"、"青岛啤酒博物馆周边"）

**参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `keywords` | string | ✅ | 地点名称，如"五四广场" |

**返回示例**:

```json
{
  "keywords": "五四广场",
  "latitude": "36.061892",
  "longitude": "120.384428",
  "message": "成功获取地点 五四广场 的坐标"
}
```

---

### 6.2 queryJobsByArea - 按区域查询岗位

**功能**: 根据青岛市区域代码查询岗位信息

**触发场景**: 用户指定区域名称时（如"城阳区"、"市北区"）

**参数**:

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| `jobTitle` | string | ✅ | - | 岗位关键词 |
| `current` | integer | ✅ | 1 | 页码 |
| `pageSize` | integer | ✅ | 10 | 每页数量 |
| `jobLocationAreaCode` | string | ❌ | - | 区域代码（见代码表） |
| `order` | string | ❌ | "0" | 排序：0-推荐，1-最热，2-最新 |
| `minSalary` | string | ❌ | - | 最低薪资（元/月） |
| `maxSalary` | string | ❌ | - | 最高薪资（元/月） |
| `experience` | string | ❌ | - | 经验要求代码 |
| `education` | string | ❌ | - | 学历要求代码 |
| `companyNature` | string | ❌ | - | 企业类型代码 |

---

### 6.3 queryJobsByLocation - 按坐标查询岗位

**功能**: 根据经纬度和半径查询附近岗位

**触发场景**: 用户指定具体地点后，需要查询附近岗位

**参数**:

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| `jobTitle` | string | ✅ | - | 岗位关键词 |
| `current` | integer | ✅ | 1 | 页码 |
| `pageSize` | integer | ✅ | 10 | 每页数量 |
| `latitude` | string | ✅ | - | 纬度 |
| `longitude` | string | ✅ | - | 经度 |
| `radius` | string | ✅ | "10" | 搜索半径（千米，最大50） |
| `order` | string | ❌ | "0" | 排序方式 |
| `minSalary` | string | ❌ | - | 最低薪资 |
| `maxSalary` | string | ❌ | - | 最高薪资 |
| `experience` | string | ❌ | - | 经验要求 |
| `education` | string | ❌ | - | 学历要求 |
| `companyNature` | string | ❌ | - | 企业类型 |

---

### 6.4 queryPolicy - 政策咨询

**功能**: 查询青岛市就业创业、社保医保、人才政策等

**触发场景**: 用户咨询政策相关问题

**参数**:

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| `message` | string | ✅ | - | 咨询问题 |
| `chatId` | string | ❌ | - | 会话ID（多轮对话） |
| `conversationId` | string | ❌ | - | 流水号（多轮对话） |
| `realName` | boolean | ❌ | false | 是否实名咨询 |
| `aac001` | string | ❌* | - | 个人编号（实名时必填） |
| `aac147` | string | ❌* | - | 身份证号（实名时必填） |
| `aac003` | string | ❌* | - | 姓名（实名时必填） |

---

### 6.5 parsePDF - PDF 解析

**功能**: 使用 OCR 服务解析 PDF 文件内容（如简历）

**说明**: 通常在文件上传时自动触发，无需手动调用。文件会通过 OCR 服务进行识别解析。

**参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `fileUrl` | string | ✅ | PDF 文件 URL |

---

### 6.6 parseImage - 图片解析

**功能**: 使用 OCR 服务识别图片中的文本内容

**说明**: 通常在文件上传时自动触发，无需手动调用。文件会通过 OCR 服务进行识别解析。

**参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `imageUrl` | string | ✅ | 图片文件 URL |

---

## 7. 代码对照表

### 7.1 区域代码 (jobLocationAreaCode)

| 代码 | 区域 |
|------|------|
| 0 | 市南区 |
| 1 | 市北区 |
| 2 | 李沧区 |
| 3 | 崂山区 |
| 4 | 黄岛区 |
| 5 | 城阳区 |
| 6 | 即墨区 |
| 7 | 胶州市 |
| 8 | 平度市 |
| 9 | 莱西市 |

### 7.2 学历代码 (education)

| 代码 | 学历 |
|------|------|
| -1 | 学历不限 |
| 0 | 初中及以下 |
| 1 | 中专/中技 |
| 2 | 高中 |
| 3 | 大专 |
| 4 | 本科 |
| 5 | 硕士 |
| 6 | 博士 |
| 7 | MBA/EMBA |
| 8 | 留学-学士 |
| 9 | 留学-硕士 |
| 10 | 留学-博士 |

### 7.3 经验代码 (experience)

| 代码 | 经验 |
|------|------|
| 0 | 经验不限 |
| 1 | 实习生 |
| 2 | 应届毕业生 |
| 3 | 1年以下 |
| 4 | 1-3年 |
| 5 | 3-5年 |
| 6 | 5-10年 |
| 7 | 10年以上 |

### 7.4 企业类型代码 (companyNature)

| 代码 | 类型 |
|------|------|
| 1 | 私营企业 |
| 2 | 股份制企业 |
| 3 | 国有企业 |
| 4 | 外商及港澳台投资企业 |
| 5 | 医院 |

### 7.5 排序方式代码 (order)

| 代码 | 排序 |
|------|------|
| 0 | 推荐 |
| 1 | 最热 |
| 2 | 最新发布 |

---

## 8. SDK 与代码示例

### 8.1 cURL 示例

#### 基础对话

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "qd-job-turbo",
    "messages": [
      {"role": "user", "content": "帮我推荐城阳区的Java开发岗位"}
    ],
    "stream": true
  }'
```

#### 带文件URL（Vision API 兼容格式）

通过 `image_url` 字段发送文件 URL（支持图片、PDF、Excel、PPT 等）：

```bash
# 发送图片
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "qd-job-turbo",
    "messages": [
      {
        "role": "user",
        "content": [
          {"type": "text", "text": "根据这份简历帮我推荐合适的岗位"},
          {"type": "image_url", "image_url": {"url": "https://example.com/resume.jpg"}}
        ]
      }
    ],
    "stream": true
  }'

# 发送 PDF 文件
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "qd-job-turbo",
    "messages": [
      {
        "role": "user",
        "content": [
          {"type": "text", "text": "分析这份PDF简历"},
          {"type": "image_url", "image_url": {"url": "https://example.com/resume.pdf"}}
        ]
      }
    ],
    "stream": true
  }'
```

#### 多轮对话

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "qd-job-turbo",
    "messages": [
      {"role": "user", "content": "我想找Java开发岗位"},
      {"role": "assistant", "content": "好的，请问您希望在青岛哪个区域工作？"},
      {"role": "user", "content": "城阳区，要求薪资15k以上"}
    ],
    "stream": true
  }'
```

#### 政策咨询

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "qd-job-turbo",
    "messages": [
      {"role": "user", "content": "青岛市大学生就业补贴政策是怎样的？"}
    ],
    "stream": true
  }'
```

---

### 8.2 Python 示例

#### 使用 OpenAI SDK

```python
from openai import OpenAI

# 创建客户端，指向本地服务
client = OpenAI(
    base_url="http://localhost:8080/v1",
    api_key="not-needed"  # 本系统不需要 API key
)

# 流式对话
def chat_stream(message: str):
    stream = client.chat.completions.create(
        model="qd-job-turbo",
        messages=[{"role": "user", "content": message}],
        stream=True
    )
    
    for chunk in stream:
        if chunk.choices[0].delta.content:
            print(chunk.choices[0].delta.content, end="", flush=True)
    print()

# 示例
chat_stream("帮我推荐青岛城阳区的Java开发岗位")
```

#### 使用 Vision API 兼容格式发送文件 URL

通过 `image_url` 字段发送文件 URL，支持图片、PDF、Excel、PPT 等格式：

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:8080/v1",
    api_key="not-needed"
)

# 带文件URL的对话（Vision API 兼容格式）
# 注意：image_url 是兼容字段名，实际支持多种文件格式
def chat_with_file_url(text: str, file_url: str):
    stream = client.chat.completions.create(
        model="qd-job-turbo",
        messages=[
            {
                "role": "user",
                "content": [
                    {"type": "text", "text": text},
                    {"type": "image_url", "image_url": {"url": file_url}}
                ]
            }
        ],
        stream=True
    )
    
    for chunk in stream:
        if chunk.choices[0].delta.content:
            print(chunk.choices[0].delta.content, end="", flush=True)
    print()

# 示例1：发送图片简历
chat_with_file_url(
    "根据这份简历帮我推荐合适的岗位",
    "https://example.com/resume.jpg"
)

# 示例2：发送 PDF 文件
chat_with_file_url(
    "分析这份PDF文档",
    "https://example.com/document.pdf"
)

# 示例3：发送 Excel 文件
chat_with_file_url(
    "帮我分析这个表格",
    "https://example.com/data.xlsx"
)
```

#### 使用 Requests 库

```python
import requests
import json

def chat_with_file(message: str, file_path: str = None):
    url = "http://localhost:8080/v1/chat/completions"
    
    if file_path:
        # 带文件上传
        files = {
            'file': open(file_path, 'rb'),
            'request': (None, json.dumps({
                "model": "qd-job-turbo",
                "messages": [{"role": "user", "content": message}],
                "stream": True
            }))
        }
        response = requests.post(url, files=files, stream=True)
    else:
        # 普通请求
        response = requests.post(
            url,
            json={
                "model": "qd-job-turbo",
                "messages": [{"role": "user", "content": message}],
                "stream": True
            },
            stream=True
        )
    
    # 处理流式响应
    for line in response.iter_lines():
        if line:
            line = line.decode('utf-8')
            if line.startswith('data: '):
                data = line[6:]
                if data == '[DONE]':
                    break
                chunk = json.loads(data)
                if chunk['choices'][0]['delta'].get('content'):
                    print(chunk['choices'][0]['delta']['content'], end='', flush=True)
    print()

# 示例
chat_with_file("根据我的简历推荐岗位", "resume.pdf")
```

---

### 8.3 JavaScript/TypeScript 示例

#### 使用 Fetch API

```typescript
interface ChatMessage {
  role: 'system' | 'user' | 'assistant';
  content: string;
}

async function chat(messages: ChatMessage[]): Promise<void> {
  const response = await fetch('http://localhost:8080/v1/chat/completions', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      model: 'qd-job-turbo',
      messages,
      stream: true,
    }),
  });

  const reader = response.body?.getReader();
  const decoder = new TextDecoder();

  if (!reader) return;

  while (true) {
    const { done, value } = await reader.read();
    if (done) break;

    const text = decoder.decode(value);
    const lines = text.split('\n');

    for (const line of lines) {
      if (line.startsWith('data: ')) {
        const data = line.slice(6);
        if (data === '[DONE]') return;

        try {
          const chunk = JSON.parse(data);
          const content = chunk.choices[0]?.delta?.content;
          if (content) {
            process.stdout.write(content);
          }
        } catch (e) {
          // 忽略解析错误
        }
      }
    }
  }
}

// 使用示例
chat([
  { role: 'user', content: '帮我推荐城阳区的Java开发岗位' }
]);
```

#### 使用 OpenAI SDK (Node.js)

```typescript
import OpenAI from 'openai';

const client = new OpenAI({
  baseURL: 'http://localhost:8080/v1',
  apiKey: 'not-needed',
});

async function chatStream(message: string) {
  const stream = await client.chat.completions.create({
    model: 'qd-job-turbo',
    messages: [{ role: 'user', content: message }],
    stream: true,
  });

  for await (const chunk of stream) {
    process.stdout.write(chunk.choices[0]?.delta?.content || '');
  }
  console.log();
}

chatStream('帮我推荐青岛城阳区的Java开发岗位');
```

---

### 8.4 Go 示例

```go
package main

import (
    "bufio"
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "strings"
)

type ChatRequest struct {
    Model    string    `json:"model"`
    Messages []Message `json:"messages"`
    Stream   bool      `json:"stream"`
}

type Message struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

func chat(message string) error {
    req := ChatRequest{
        Model: "qd-job-turbo",
        Messages: []Message{
            {Role: "user", Content: message},
        },
        Stream: true,
    }

    body, _ := json.Marshal(req)
    resp, err := http.Post(
        "http://localhost:8080/v1/chat/completions",
        "application/json",
        bytes.NewBuffer(body),
    )
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    scanner := bufio.NewScanner(resp.Body)
    for scanner.Scan() {
        line := scanner.Text()
        if strings.HasPrefix(line, "data: ") {
            data := strings.TrimPrefix(line, "data: ")
            if data == "[DONE]" {
                break
            }

            var chunk map[string]interface{}
            if err := json.Unmarshal([]byte(data), &chunk); err != nil {
                continue
            }

            if choices, ok := chunk["choices"].([]interface{}); ok && len(choices) > 0 {
                if choice, ok := choices[0].(map[string]interface{}); ok {
                    if delta, ok := choice["delta"].(map[string]interface{}); ok {
                        if content, ok := delta["content"].(string); ok {
                            fmt.Print(content)
                        }
                    }
                }
            }
        }
    }
    fmt.Println()
    return nil
}

func main() {
    chat("帮我推荐城阳区的Java开发岗位")
}
```

---

### 8.5 Vue.js 前端示例

```vue
<template>
  <div class="chat-container">
    <div class="messages">
      <div v-for="msg in messages" :key="msg.id" :class="['message', msg.role]">
        <div class="content" v-html="formatContent(msg.content)"></div>
      </div>
      <div v-if="loading" class="message assistant">
        <div class="content">{{ currentResponse }}</div>
      </div>
    </div>
    
    <div class="input-area">
      <input 
        v-model="input" 
        @keyup.enter="sendMessage"
        placeholder="输入消息..."
        :disabled="loading"
      />
      <button @click="sendMessage" :disabled="loading">发送</button>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue';

const messages = ref([]);
const input = ref('');
const loading = ref(false);
const currentResponse = ref('');

async function sendMessage() {
  if (!input.value.trim() || loading.value) return;
  
  const userMessage = input.value;
  messages.value.push({
    id: Date.now(),
    role: 'user',
    content: userMessage
  });
  input.value = '';
  loading.value = true;
  currentResponse.value = '';
  
  try {
    const response = await fetch('http://localhost:8080/v1/chat/completions', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        model: 'qd-job-turbo',
        messages: messages.value.map(m => ({
          role: m.role,
          content: m.content
        })),
        stream: true
      })
    });
    
    const reader = response.body.getReader();
    const decoder = new TextDecoder();
    
    while (true) {
      const { done, value } = await reader.read();
      if (done) break;
      
      const text = decoder.decode(value);
      const lines = text.split('\n');
      
      for (const line of lines) {
        if (line.startsWith('data: ')) {
          const data = line.slice(6);
          if (data === '[DONE]') continue;
          
          try {
            const chunk = JSON.parse(data);
            const content = chunk.choices[0]?.delta?.content;
            if (content) {
              currentResponse.value += content;
            }
          } catch (e) {}
        }
      }
    }
    
    messages.value.push({
      id: Date.now(),
      role: 'assistant',
      content: currentResponse.value
    });
  } catch (error) {
    console.error('Chat error:', error);
  } finally {
    loading.value = false;
    currentResponse.value = '';
  }
}

function formatContent(content) {
  // 解析 job-json 代码块
  return content.replace(
    /```\s*job-json\n([\s\S]*?)```/g,
    (_, json) => {
      try {
        const job = JSON.parse(json);
        return `<div class="job-card">
          <h3>${job.jobTitle}</h3>
          <p class="company">${job.companyName}</p>
          <p class="salary">${job.salary}</p>
          <p class="info">${job.location} | ${job.education} | ${job.experience}</p>
          <a href="${job.appJobUrl}" target="_blank">查看详情</a>
        </div>`;
      } catch {
        return json;
      }
    }
  );
}
</script>

<style scoped>
.chat-container {
  max-width: 800px;
  margin: 0 auto;
  padding: 20px;
}

.messages {
  height: 500px;
  overflow-y: auto;
  border: 1px solid #ddd;
  padding: 10px;
  margin-bottom: 10px;
}

.message {
  margin-bottom: 10px;
  padding: 10px;
  border-radius: 8px;
}

.message.user {
  background: #e3f2fd;
  text-align: right;
}

.message.assistant {
  background: #f5f5f5;
}

.input-area {
  display: flex;
  gap: 10px;
}

.input-area input {
  flex: 1;
  padding: 10px;
  border: 1px solid #ddd;
  border-radius: 4px;
}

.input-area button {
  padding: 10px 20px;
  background: #1976d2;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.job-card {
  background: white;
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  padding: 15px;
  margin: 10px 0;
}

.job-card h3 {
  margin: 0 0 8px 0;
  color: #1976d2;
}

.job-card .company {
  font-weight: bold;
  color: #333;
}

.job-card .salary {
  color: #e53935;
  font-size: 1.1em;
}

.job-card .info {
  color: #666;
  font-size: 0.9em;
}

.job-card a {
  display: inline-block;
  margin-top: 10px;
  color: #1976d2;
}
</style>
```

---

## 9. 错误处理

### 9.1 错误响应格式

```json
{
  "error": {
    "message": "错误描述信息",
    "type": "错误类型",
    "code": "错误代码"
  }
}
```

### 9.2 常见错误码

| HTTP 状态码 | 错误类型 | 说明 |
|------------|---------|------|
| 400 | `invalid_request` | 请求格式错误 |
| 400 | `multipart_parse_error` | multipart 表单解析失败 |
| 400 | `file_processing_error` | 文件处理失败 |
| 429 | `rate_limit_exceeded` | 请求过于频繁 |
| 500 | `internal_error` | 服务器内部错误 |

### 9.3 错误处理建议

```python
import requests

def safe_chat(message: str) -> str:
    try:
        response = requests.post(
            "http://localhost:8080/v1/chat/completions",
            json={
                "model": "qd-job-turbo",
                "messages": [{"role": "user", "content": message}],
                "stream": False
            },
            timeout=120
        )
        
        if response.status_code == 429:
            # 限流，等待后重试
            import time
            time.sleep(2)
            return safe_chat(message)
        
        response.raise_for_status()
        return response.json()["choices"][0]["message"]["content"]
        
    except requests.exceptions.Timeout:
        return "请求超时，请稍后重试"
    except requests.exceptions.RequestException as e:
        return f"请求失败: {e}"
```

---

## 10. 配置参考

### 10.1 完整配置文件 (config.yaml)

```yaml
# 服务器配置
server:
  port: 8080                 # 服务端口
  host: "0.0.0.0"            # 监听地址
  read_timeout: 30s          # 读取请求超时
  write_timeout: 300s        # 写入响应超时（流式响应需要更长时间）

# LLM配置
llm:
  base_url: "https://api.openai.com/v1"  # LLM API地址
  api_key: "sk-xxx"                       # LLM API密钥
  model: "gpt-4o"                         # 默认模型
  timeout: 120s                           # 请求超时
  max_retries: 3                          # 最大重试次数

# 高德地图配置
amap:
  api_key: "your-amap-key"                # 高德地图API密钥
  base_url: "https://restapi.amap.com/v3" # 高德API地址
  timeout: 10s                            # 请求超时

# 岗位API配置
job_api:
  base_url: "https://job-api.example.com" # 岗位API地址
  timeout: 30s                            # 请求超时

# OCR服务配置（文件解析）
ocr:
  base_url: "https://your-ocr-api.example.com"  # OCR服务地址（外网）
  # base_url: "http://127.0.0.1:9001"     # OCR服务地址（内网）
  timeout: 120s                           # 请求超时

# 政策咨询配置
policy:
  base_url: "http://policy-api.example.com"  # 政策API地址
  login_name: "your_login_name"              # 登录用户名
  user_key: "your_user_key"                  # 用户密钥
  service_id: "your_service_id"              # 服务ID
  timeout: 60s                               # 请求超时

# 日志配置
logging:
  level: "info"       # 日志级别：debug, info, warn, error
  format: "json"      # 日志格式：json, text

# 性能配置
performance:
  max_goroutines: 10000         # 最大并发goroutine数
  goroutine_pool_size: 5000     # goroutine池大小
  task_queue_size: 10000        # 任务队列大小
  enable_pprof: true            # 启用pprof性能分析（设为 false 可关闭 /debug/pprof/*）
  enable_metrics: true          # 启用指标收集（设为 false 可关闭 /metrics 与指标中间件）
  gc_percent: 100               # GC触发百分比
```

### 10.2 环境变量覆盖

环境变量会自动覆盖配置文件中的值：

| 环境变量 | 配置项 |
|---------|--------|
| `SERVER_PORT` | server.port |
| `SERVER_HOST` | server.host |
| `LLM_API_KEY` | llm.api_key |
| `LLM_BASE_URL` | llm.base_url |
| `LLM_MODEL` | llm.model |
| `AMAP_API_KEY` | amap.api_key |
| `OCR_BASE_URL` | ocr.base_url |
| `JOB_API_BASE_URL` | job_api.base_url |
| `POLICY_BASE_URL` | policy.base_url |
| `POLICY_LOGIN_NAME` | policy.login_name |
| `POLICY_USER_KEY` | policy.user_key |
| `POLICY_SERVICE_ID` | policy.service_id |

---

## 11. 部署指南

### 11.1 Docker Compose 部署

```yaml
# docker-compose.yml
version: '3.8'

services:
  qd-sc-server:
    build: .
    ports:
      - "8080:8080"
    environment:
      - LLM_API_KEY=sk-xxx
      - LLM_BASE_URL=https://api.openai.com/v1
      - AMAP_API_KEY=your-amap-key
      - OCR_BASE_URL=https://your-ocr-api.example.com
    volumes:
      - ./config.yaml:/app/config.yaml:ro
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
```

### 11.2 Kubernetes 部署

```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: qd-sc-server
spec:
  replicas: 3
  selector:
    matchLabels:
      app: qd-sc-server
  template:
    metadata:
      labels:
        app: qd-sc-server
    spec:
      containers:
      - name: qd-sc-server
        image: qd-sc-server:latest
        ports:
        - containerPort: 8080
        env:
        - name: LLM_API_KEY
          valueFrom:
            secretKeyRef:
              name: qd-sc-secrets
              key: llm-api-key
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 30
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 10
---
apiVersion: v1
kind: Service
metadata:
  name: qd-sc-server
spec:
  selector:
    app: qd-sc-server
  ports:
  - port: 80
    targetPort: 8080
  type: LoadBalancer
```

### 11.3 Nginx 反向代理

```nginx
upstream qd_sc_backend {
    server 127.0.0.1:8080;
    keepalive 32;
}

server {
    listen 443 ssl http2;
    server_name api.example.com;
    
    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;
    
    location / {
        proxy_pass http://qd_sc_backend;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # SSE 支持
        proxy_set_header Connection '';
        proxy_buffering off;
        proxy_cache off;
        chunked_transfer_encoding off;
        
        # 超时配置
        proxy_connect_timeout 60s;
        proxy_send_timeout 300s;
        proxy_read_timeout 300s;
    }
    
    # 文件上传大小限制
    client_max_body_size 20M;
}
```

---

## 12. 常见问题

### Q1: 如何处理流式响应中的岗位数据？

岗位数据会以特殊的 ```` ``` job-json ```` 代码块格式返回。客户端需要解析这个格式：

```javascript
function parseJobCards(content) {
  const jobRegex = /```\s*job-json\n([\s\S]*?)```/g;
  const jobs = [];
  let match;
  
  while ((match = jobRegex.exec(content)) !== null) {
    try {
      jobs.push(JSON.parse(match[1]));
    } catch (e) {}
  }
  
  return jobs;
}
```

### Q2: 流式响应中途断开怎么办？

建议实现重连机制：

```python
import time

def chat_with_retry(message, max_retries=3):
    for i in range(max_retries):
        try:
            return chat_stream(message)
        except Exception as e:
            if i < max_retries - 1:
                time.sleep(2 ** i)  # 指数退避
            else:
                raise e
```

### Q3: 如何实现多轮对话？

保存历史消息并在每次请求中传递：

```python
conversation = []

def chat(user_input):
    conversation.append({"role": "user", "content": user_input})
    
    response = client.chat.completions.create(
        model="qd-job-turbo",
        messages=conversation,
        stream=True
    )
    
    assistant_message = ""
    for chunk in response:
        content = chunk.choices[0].delta.content or ""
        assistant_message += content
        print(content, end="", flush=True)
    
    conversation.append({"role": "assistant", "content": assistant_message})
    print()
```

### Q4: 429 错误（限流）如何处理？

系统默认限流为：桶容量 200，每秒补充 50 个令牌。建议：
1. 实现请求重试机制，遇到 429 后等待 1-2 秒重试
2. 控制客户端的请求频率
3. 如需更高 QPS，联系管理员调整配置

### Q5: 如何调试 API 问题？

1. 启用 debug 日志级别
2. （可选）使用 `/metrics` 端点查看性能指标（需启用 `performance.enable_metrics`）
3. （可选）使用 `/debug/pprof/` 进行性能分析（需启用 `performance.enable_pprof`）
4. 检查服务器日志输出

---

## 附录

### A. 响应时间参考

| 操作 | 预期响应时间 |
|------|-------------|
| 健康检查 | < 10ms |
| 简单对话 | 2-5s |
| 岗位查询 | 3-8s |
| 文件解析 | 5-30s |
| 政策咨询 | 3-10s |

### B. 并发能力

| 指标 | 参考值 |
|------|--------|
| 最大并发连接 | 10,000 |
| 建议 QPS | 50 |
| 最大文件上传 | 10MB |

### C. 更新日志

| 版本 | 日期 | 说明 |
|------|------|------|
| 1.0.0 | 2025-11 | 初始版本 |

