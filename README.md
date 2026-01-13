# 青岛岗位匹配系统

基于Go开发的智能岗位匹配系统，兼容OpenAI `/v1/chat/completions` API接口。

## 配置文档

### 完整配置示例

编辑 `config.yaml`：

```yaml
# 城市配置
city:
  name: "石河子"                                  # 城市名称
  system_name: "石河子岗位匹配系统"                # 系统名称
  area_codes:                                   # 区域代码映射
    石河子市: "0"
    疏附县: "1"
    疏勒县: "2"
    英吉沙县: "3"
    岳普湖县: "4"
    伽师县: "5"
    莎车县: "6"
    泽普县: "7"
    叶城县: "8"
    麦盖提县: "9"
    巴楚县: "10"
    塔什库尔干塔吉克自治县: "11"
  landmarks:                                    # 地标示例（用于工具描述）
    - "石河子大学"
    - "军垦博物馆"
  abbreviations:                                # 简称映射（用于语义理解提示）
    石大: "石河子大学"

# 服务器配置
server:
  port: 8080                 # 服务端口
  host: "0.0.0.0"            # 监听地址
  read_timeout: 30s          # 读取请求超时
  write_timeout: 300s        # 写入响应超时（流式响应需要更长时间）

# LLM配置
llm:
  base_url: "https://apix.newapi.pro/v1"       # LLM API地址
  api_key: "sk-xxx"                            # LLM API密钥（必填）
  model: "gpt-4o"                              # 默认模型
  timeout: 120s                                # 请求超时
  max_retries: 3                               # 最大重试次数

# 高德地图配置
amap:
  api_key: "your-amap-api-key"                 # 高德地图API密钥（必填）
  base_url: "https://restapi.amap.com/v3"      # 高德API地址
  timeout: 10s                                 # 请求超时

# 岗位API配置
job_api:
  base_url: "http://127.0.0.1:9091/app/job/aiList"  # 岗位API地址
  timeout: 30s                                       # 请求超时

# OCR服务配置（文件解析）
ocr:
  base_url: "https://your-ocr-api.example.com"  # OCR服务地址（外网）
  # base_url: "http://127.0.0.1:9001"          # OCR服务地址（内网）
  timeout: 120s                                # 请求超时

# 政策咨询配置
policy:
  base_url: "http://your-policy-api:port"      # 政策API地址
  login_name: "your_login_name"                # 由平台提供的用户名
  user_key: "your_user_key"                    # 由平台提供的密码加密
  service_id: "your_service_id"                # 政策咨询服务ID
  timeout: 60s                                 # 请求超时

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
  gc_percent: 100               # GC触发百分比（默认100）
```

### 环境变量配置

环境变量会自动覆盖配置文件中的值（推荐生产环境使用）：

```bash
# 服务器
export SERVER_PORT="8080"
export SERVER_HOST="0.0.0.0"

# LLM
export LLM_API_KEY="sk-xxx"
export LLM_BASE_URL="https://apix.newapi.pro/v1"
export LLM_MODEL="gpt-4o"

# 高德地图
export AMAP_API_KEY="your-amap-key"

# OCR服务
export OCR_BASE_URL="https://your-ocr-api.example.com"

# 岗位API
export JOB_API_BASE_URL="http://127.0.0.1:9091/app/job/aiList"

# 政策API
export POLICY_BASE_URL="http://your-policy-api:port"
export POLICY_LOGIN_NAME="your_login_name"
export POLICY_USER_KEY="your_user_key"
export POLICY_SERVICE_ID="your_service_id"

# 城市配置
export CITY_NAME="石河子"
export CITY_SYSTEM_NAME="石河子岗位匹配系统"
```

## 运行

```bash
# 安装依赖
go mod download

# 直接运行
go run cmd/server/main.go

# 或编译后运行
make build
./qd-sc-server

# 使用自定义配置文件
./qd-sc-server -config=/path/to/config.yaml
```

服务启动在 `http://localhost:8080`

## API端点

系统提供以下端点：

| 端点 | 方法 | 说明 |
|------|------|------|
| `/` | GET | API信息和端点列表 |
| `/health` | GET | 健康检查 |
| `/metrics` | GET | 性能指标（需启用 `performance.enable_metrics`） |
| `/v1/chat/completions` | POST | OpenAI兼容的聊天接口（主要接口） |
| `/debug/pprof/*` | GET | 性能分析（pprof，需启用 `performance.enable_pprof`） |

## API调用

### 1. 普通对话

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

### 2. 带文件URL（Vision API 兼容格式）

通过 `image_url` 字段发送文件 URL，支持图片、PDF、Excel、PPT 等格式：

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "qd-job-turbo",
    "messages": [
      {
        "role": "user",
        "content": [
          {"type": "text", "text": "根据这份简历推荐岗位"},
          {"type": "image_url", "image_url": {"url": "https://example.com/resume.pdf"}}
        ]
      }
    ],
    "stream": true
  }'
```

> **说明**: `image_url` 是 OpenAI Vision API 的兼容字段名，实际支持图片、PDF、Excel、PPT 等多种文件格式。

### 3. 多轮对话

支持上下文对话，只需在 `messages` 中包含历史消息：

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "qd-job-turbo",
    "messages": [
      {"role": "user", "content": "我想找Java开发岗位"},
      {"role": "assistant", "content": "好的，请问您希望在哪个区域找工作？"},
      {"role": "user", "content": "城阳区"}
    ],
    "stream": true
  }'
```

### 4. 请求参数说明

```json
{
  "model": "qd-job-turbo",        // 必填：固定模型名称
  "messages": [                   // 必填：对话消息列表
    {
      "role": "user",             // 角色：user, assistant, system
      "content": "你的问题"        // 消息内容
    }
  ],
  "stream": true,                 // 可选：是否流式输出（推荐true）
  "temperature": 0.7,             // 可选：温度参数（0-2）
  "max_tokens": 2000,             // 可选：最大生成token数
  "top_p": 1.0,                   // 可选：nucleus采样参数
  "presence_penalty": 0.0,        // 可选：存在惩罚
  "frequency_penalty": 0.0        // 可选：频率惩罚
}
```

### 5. 响应格式

**流式响应**（`stream: true`）：

```
data: {"id":"chatcmpl-xxx","object":"chat.completion.chunk","created":1234567890,"model":"qd-job-turbo","choices":[{"index":0,"delta":{"role":"assistant","content":"您好"},"finish_reason":null}]}

data: {"id":"chatcmpl-xxx","object":"chat.completion.chunk","created":1234567890,"model":"qd-job-turbo","choices":[{"index":0,"delta":{"content":"，"},"finish_reason":null}]}

data:  [DONE]
```

**非流式响应**（`stream: false`）：

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

## 内置功能

系统会根据对话自动调用以下工具：

1. **queryLocation** - 查询地点坐标（高德地图）
2. **queryJobsByArea** - 按区域查询岗位
3. **queryJobsByLocation** - 按坐标查询岗位
4. **queryPolicy** - 政策咨询
5. **parsePDF** - PDF解析（OCR服务）
6. **parseImage** - 图片识别（OCR服务）

### 工具参数说明

#### queryJobsByArea（按区域查询岗位）

```json
{
  "area": 5,              // 区域代码（0-9）
  "keyword": "Java",      // 可选：关键词
  "education": 4,         // 可选：学历代码
  "experience": 5,        // 可选：经验代码
  "page": 1,              // 可选：页码
  "pageSize": 20          // 可选：每页数量
}
```

#### queryJobsByLocation（按坐标查询岗位）

```json
{
  "latitude": "36.307527",   // 纬度
  "longitude": "120.467121", // 经度
  "keyword": "Java",         // 可选：关键词
  "radius": 5000,            // 可选：搜索半径（米）
  "education": 4,            // 可选：学历代码
  "experience": 5            // 可选：经验代码
}
```

#### queryPolicy（政策咨询）

```json
{
  "message": "咨询问题",           // 必填：咨询内容
  "chatId": "xxx",                // 可选：会话ID（多轮对话）
  "conversationId": "xxx",        // 可选：对话ID（多轮对话）
  "realName": false,              // 可选：是否实名咨询
  "aac001": "个人编号",           // 实名时必填
  "aac147": "身份证号",           // 实名时必填
  "aac003": "姓名"                // 实名时必填
}
```

### 代码对照表

#### 区域代码

- 0:市南区, 1:市北区, 2:李沧区, 3:崂山区, 4:黄岛区
- 5:城阳区, 6:即墨区, 7:胶州市, 8:平度市, 9:莱西市

#### 学历代码

- -1:不限, 0:初中及以下, 1:中专/中技, 2:高中, 3:大专
- 4:本科, 5:硕士, 6:博士, 7:MBA/EMBA, 8-10:留学

#### 经验代码

- 0:不限, 1:实习生, 2:应届, 3:1年以下
- 4:1-3年, 5:3-5年, 6:5-10年, 7:10年以上

## 其他端点

### 健康检查

```bash
curl http://localhost:8080/health
```

响应：

```json
{
  "status": "ok",
  "timestamp": "2024-01-01T12:00:00Z"
}
```

### 性能指标

```bash
curl http://localhost:8080/metrics
```

响应：

```json
{
  "requests_total": 1234,
  "requests_success": 1200,
  "requests_failed": 34,
  "avg_response_time_ms": 150.5,
  "goroutines": 42,
  "memory_alloc_mb": 45.6
}
```

### 性能分析（pprof）

```bash
# CPU性能分析（采集30秒）
curl http://localhost:8080/debug/pprof/profile?seconds=30 -o cpu.prof

# 内存分析
curl http://localhost:8080/debug/pprof/heap -o heap.prof

# Goroutine分析
curl http://localhost:8080/debug/pprof/goroutine -o goroutine.prof

# 查看分析结果
go tool pprof cpu.prof
```

## 性能配置

### 限流

系统默认配置：
- **桶容量**: 200（突发请求）
- **补充速率**: 50/秒（持续QPS）

超过限流会返回 `429 Too Many Requests`。

### 连接池

- LLM API: 100最大连接，20/host
- 其他API: 50最大连接，10/host

### 超时配置

- 读取请求: 30秒
- 写入响应: 300秒（流式响应）
- LLM请求: 120秒
- 其他API: 10-60秒

## Docker部署

### 架构检测和选择

在部署前，请先检测您的系统架构：

```bash
# Linux/Mac 检测架构
uname -m
# 输出说明：
# x86_64    -> 使用 AMD64 配置
# aarch64   -> 使用 ARM64 配置
# arm64     -> 使用 ARM64 配置

# Windows PowerShell 检测架构
$env:PROCESSOR_ARCHITECTURE
# 输出说明：
# AMD64     -> 使用 AMD64 配置
# ARM64     -> 使用 ARM64 配置
```

### 使用docker-compose部署（推荐）

根据您的系统架构选择对应的配置文件：

#### AMD64 架构系统（Intel/AMD处理器）
```bash
# 使用AMD64专用配置
docker-compose -f docker-compose-amd64.yml up -d

# 查看服务状态
docker-compose -f docker-compose-amd64.yml ps

# 查看日志
docker-compose -f docker-compose-amd64.yml logs -f

# 停止服务
docker-compose -f docker-compose-amd64.yml down
```

#### ARM64 架构系统（Apple Silicon Mac、ARM服务器）
```bash
# 使用ARM64专用配置
docker-compose -f docker-compose-arm64.yml up -d

# 查看服务状态
docker-compose -f docker-compose-arm64.yml ps

# 查看日志
docker-compose -f docker-compose-arm64.yml logs -f

# 停止服务
docker-compose -f docker-compose-arm64.yml down
```

#### 通用配置（自动检测架构）
如果您不确定架构或希望使用通用配置：
```bash
# 使用默认配置（会自动选择合适的架构）
docker-compose up -d
```

### 环境变量配置

在启动前，请确保配置环境变量：

```bash
# 方法1：创建 .env 文件（推荐）
cp .env.example .env
# 编辑 .env 文件，填入真实的API密钥

# 方法2：直接设置环境变量
export LLM_API_KEY="sk-your-actual-api-key"
export AMAP_API_KEY="your-amap-api-key"
export OCR_BASE_URL="https://your-ocr-api.example.com"
```

### 完整部署流程

```bash
# 1. 检测系统架构
uname -m

# 2. 配置环境变量
cp .env.example .env
# 编辑 .env 文件

# 3. 根据架构启动服务
# AMD64 系统：
docker-compose -f docker-compose-amd64.yml up -d

# ARM64 系统：
docker-compose -f docker-compose-arm64.yml up -d

# 4. 验证服务
curl http://localhost:8080/health

# 5. 测试API
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "qd-job-turbo",
    "messages": [{"role": "user", "content": "你好"}],
    "stream": false
  }'
```

### 手动构建和运行

#### AMD64 架构
```bash
# 构建镜像
docker build --platform linux/amd64 -t qd-sc-server:amd64 .

# 运行容器
docker run -d -p 8080:8080 \
  --platform linux/amd64 \
  -e LLM_API_KEY="sk-xxx" \
  -e LLM_BASE_URL="https://apix.newapi.pro/v1" \
  -e AMAP_API_KEY="xxx" \
  -e OCR_BASE_URL="https://your-ocr-api.example.com" \
  -e JOB_API_BASE_URL="http://127.0.0.1:9091/app/job/aiList" \
  -v $(pwd)/config.yaml:/app/config.yaml:ro \
  --name qd-sc-server-amd64 \
  qd-sc-server:amd64
```

#### ARM64 架构
```bash
# 构建镜像
docker build --platform linux/arm64 -t qd-sc-server:arm64 .

# 运行容器
docker run -d -p 8080:8080 \
  --platform linux/arm64 \
  -e LLM_API_KEY="sk-xxx" \
  -e LLM_BASE_URL="https://apix.newapi.pro/v1" \
  -e AMAP_API_KEY="xxx" \
  -e OCR_BASE_URL="https://your-ocr-api.example.com" \
  -e JOB_API_BASE_URL="http://127.0.0.1:9091/app/job/aiList" \
  -e GOMAXPROCS=4 \
  -e GOMEMLIMIT=2GiB \
  -v $(pwd)/config.yaml:/app/config.yaml:ro \
  --name qd-sc-server-arm64 \
  qd-sc-server:arm64
```

### 多架构镜像构建和推送

> 说明：仓库内 `Dockerfile` 已支持 buildx 的 `TARGETARCH/TARGETOS`，可直接构建 `linux/amd64` + `linux/arm64` 的同 tag 多架构镜像（manifest）。

#### 使用buildx构建多架构镜像

```bash
# 创建buildx构建器
docker buildx create --name multiarch-builder --use

# 构建并推送多架构镜像（推荐）
docker buildx build --platform linux/amd64,linux/arm64 -t t0ng7u/qd-sc:latest --push .

# 或者分别构建单架构镜像
docker buildx build --platform linux/amd64 -t t0ng7u/qd-sc:amd64 --push .
docker buildx build --platform linux/arm64 -t t0ng7u/qd-sc:arm64 --push .
```

#### 本地构建多架构镜像

```bash
# 使用提供的构建脚本
# Linux/Mac:
chmod +x build-multiarch.sh
./build-multiarch.sh

# Windows:
build-multiarch.bat
```

### 架构特定优化

#### AMD64 优化配置
- 使用所有可用CPU核心：`GOMAXPROCS=0`
- 内存限制：`GOMEMLIMIT=4GiB`
- 启动时间：30-40秒
- 适合高性能服务器环境

#### ARM64 优化配置
- 限制CPU核心数：`GOMAXPROCS=4`
- 内存限制：`GOMEMLIMIT=2GiB`
- GC优化：`GOGC=200`
- 启动时间：60-90秒
- 适合ARM服务器和Mac M1/M2

### 服务验证

```bash
# 健康检查
curl http://localhost:8080/health

# 查看服务信息
curl http://localhost:8080/

# 测试聊天接口
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "qd-job-turbo",
    "messages": [
      {"role": "user", "content": "帮我推荐石河子市的Java开发岗位"}
    ],
    "stream": false
  }'
```

### 查看日志和故障排除

#### 查看服务日志
```bash
# AMD64 系统
docker-compose -f docker-compose-amd64.yml logs -f

# ARM64 系统
docker-compose -f docker-compose-arm64.yml logs -f

# 查看特定服务的日志
docker logs -f qd-sc-server-amd64
docker logs -f qd-sc-server-arm64
```

#### 常见问题排查

**1. 架构不匹配错误**
```bash
# 错误信息：exec format error
# 解决方案：检查系统架构并使用对应的配置文件

# 检查系统架构
uname -m

# 检查镜像架构
docker inspect qd-sc:v1.0-amd64 | grep Architecture
docker inspect qd-sc:v1.0-arm64 | grep Architecture
```

**2. 服务启动失败**
```bash
# 检查配置文件
docker run --rm -v $(pwd)/config.yaml:/app/config.yaml qd-sc:v1.0-amd64 cat /app/config.yaml

# 检查环境变量
docker-compose -f docker-compose-amd64.yml config
```

**3. API调用失败**
```bash
# 检查服务状态
curl http://localhost:8080/health

# 检查配置
curl http://localhost:8080/

# 查看详细错误日志
docker-compose -f docker-compose-amd64.yml logs --tail=100
```
