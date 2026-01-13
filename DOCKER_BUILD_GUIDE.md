# Docker 多架构构建和打包指南

## 概述

本指南专门说明如何为青岛岗位匹配系统（qd-sc）构建和保存 ARM64 和 AMD64 架构的 Docker 镜像。

## 项目信息

- **项目名称**: 青岛岗位匹配系统 (qd-sc)
- **Go版本**: 1.21
- **服务端口**: 8080
- **镜像名称**: t0ng7u/qd-sc

## 架构说明

| 架构名称 | Docker平台 | 适用设备 |
|----------|------------|----------|
| AMD64 | linux/amd64 | Intel/AMD处理器的服务器、PC |
| ARM64 | linux/arm64 | Apple Silicon Mac、ARM服务器、树莓派4+ |

## 构建和保存镜像

### 方法1：使用buildx构建多架构镜像（推荐）

#### 1. 准备buildx环境

```bash
# 创建并使用buildx构建器
docker buildx create --name multiarch-builder --use

# 验证构建器支持的平台
docker buildx inspect --bootstrap
```

#### 2. 构建单架构镜像

**AMD64 架构构建**
```bash
# 构建 AMD64 镜像
docker buildx build --platform linux/amd64 -t qd-sc:v1.0-amd64 --load .

# 验证构建结果
docker images | grep qd-sc
docker inspect qd-sc:v1.0-amd64 | grep Architecture
# 应该显示: "Architecture": "amd64"

# 保存为压缩文件
docker save qd-sc:v1.0-amd64 | gzip > qd-sc-v1.0-amd64.tar.gz

# 查看文件大小
ls -lh qd-sc-v1.0-amd64.tar.gz
```

**ARM64 架构构建**
```bash
# 构建 ARM64 镜像
docker buildx build --platform linux/arm64 -t qd-sc:v1.0-arm64 --load .

# 验证构建结果
docker images | grep qd-sc
docker inspect qd-sc:v1.0-arm64 | grep Architecture
# 应该显示: "Architecture": "arm64"

# 保存为压缩文件
docker save qd-sc:v1.0-arm64 | gzip > qd-sc-v1.0-arm64.tar.gz

# 查看文件大小
ls -lh qd-sc-v1.0-arm64.tar.gz
```

#### 3. 构建多架构镜像（推荐生产环境）

```bash
# 同时构建两个架构并推送到仓库
docker buildx build --platform linux/amd64,linux/arm64 -t t0ng7u/qd-sc:latest --push .

# 或者构建到本地（需要先创建manifest）
docker buildx build --platform linux/amd64,linux/arm64 -t qd-sc:multiarch .
```

### 方法2：传统构建方式

#### AMD64 架构
```bash
# 构建 AMD64 镜像
docker build --platform linux/amd64 -t qd-sc:v1.0-amd64 .

# 验证和保存
docker inspect qd-sc:v1.0-amd64 | grep Architecture
docker save qd-sc:v1.0-amd64 | gzip > qd-sc-v1.0-amd64.tar.gz
```

#### ARM64 架构
```bash
# 构建 ARM64 镜像
docker build --platform linux/arm64 -t qd-sc:v1.0-arm64 .

# 验证和保存
docker inspect qd-sc:v1.0-arm64 | grep Architecture
docker save qd-sc:v1.0-arm64 | gzip > qd-sc-v1.0-arm64.tar.gz
```

## 加载和部署镜像

### 1. 加载镜像

**AMD64 架构加载**
```bash
# 加载 AMD64 压缩镜像
gunzip -c qd-sc-v1.0-amd64.tar.gz | docker load

# 验证加载结果
docker images | grep qd-sc
docker inspect qd-sc:v1.0-amd64 | grep Architecture
```

**ARM64 架构加载**
```bash
# 加载 ARM64 压缩镜像
gunzip -c qd-sc-v1.0-arm64.tar.gz | docker load

# 验证加载结果
docker images | grep qd-sc
docker inspect qd-sc:v1.0-arm64 | grep Architecture
```

### 2. Docker Compose 配置

#### AMD64 架构配置 (docker-compose-amd64.yml)

```yaml
version: '3.8'

services:
  qd-sc-server:
    image: qd-sc:v1.0-amd64
    container_name: qd-sc-server-amd64
    restart: unless-stopped
    ports:
      - "8080:8080"
    volumes:
      - ./config.yaml:/app/config.yaml:ro
    environment:
      - GIN_MODE=release
      - TZ=Asia/Shanghai
      # 环境变量配置（覆盖config.yaml）
      - LLM_API_KEY=${LLM_API_KEY}
      - LLM_BASE_URL=${LLM_BASE_URL}
      - AMAP_API_KEY=${AMAP_API_KEY}
      - OCR_BASE_URL=${OCR_BASE_URL}
      - JOB_API_BASE_URL=${JOB_API_BASE_URL}
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
    logging:
      driver: "json-file"
      options:
        max-size: "50m"
        max-file: "3"
```

#### ARM64 架构配置 (docker-compose-arm64.yml)

```yaml
version: '3.8'

services:
  qd-sc-server:
    image: qd-sc:v1.0-arm64
    platform: linux/arm64  # 强制指定平台
    container_name: qd-sc-server-arm64
    restart: unless-stopped
    ports:
      - "8080:8080"
    volumes:
      - ./config.yaml:/app/config.yaml:ro
    environment:
      - GIN_MODE=release
      - TZ=Asia/Shanghai
      # ARM优化配置
      - GOMAXPROCS=4
      - GOMEMLIMIT=2GiB
      # 环境变量配置
      - LLM_API_KEY=${LLM_API_KEY}
      - LLM_BASE_URL=${LLM_BASE_URL}
      - AMAP_API_KEY=${AMAP_API_KEY}
      - OCR_BASE_URL=${OCR_BASE_URL}
      - JOB_API_BASE_URL=${JOB_API_BASE_URL}
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 60s  # ARM启动可能稍慢
    logging:
      driver: "json-file"
      options:
        max-size: "50m"
        max-file: "3"
    # ARM平台资源限制
    deploy:
      resources:
        limits:
          memory: 2G
          cpus: '4'
        reservations:
          memory: 1G
          cpus: '2'
```

### 3. 环境变量配置文件

创建 `.env` 文件：

```bash
# LLM配置
LLM_API_KEY=sk-your-api-key
LLM_BASE_URL=https://your-llm-api.com/v1

# 高德地图配置
AMAP_API_KEY=your-amap-api-key

# OCR服务配置
OCR_BASE_URL=https://your-ocr-api.example.com

# 岗位API配置
JOB_API_BASE_URL=http://127.0.0.1:9091/app/job/aiList

# 政策API配置
POLICY_BASE_URL=http://your-policy-api:port
POLICY_LOGIN_NAME=your_login_name
POLICY_USER_KEY=your_user_key
POLICY_SERVICE_ID=your_service_id
```

### 4. 启动服务

```bash
# 检查系统架构
uname -m
# x86_64    -> 使用 AMD64 镜像
# aarch64   -> 使用 ARM64 镜像
# arm64     -> 使用 ARM64 镜像

# AMD64 系统启动
docker-compose -f docker-compose-amd64.yml up -d

# ARM64 系统启动
docker-compose -f docker-compose-arm64.yml up -d

# 查看服务状态
docker-compose -f docker-compose-amd64.yml ps
docker-compose -f docker-compose-arm64.yml ps

# 查看日志
docker-compose -f docker-compose-amd64.yml logs -f
docker-compose -f docker-compose-arm64.yml logs -f
```

### 5. 服务验证

```bash
# 健康检查
curl http://localhost:8080/health

# API信息
curl http://localhost:8080/

# 性能指标（如果启用）
curl http://localhost:8080/metrics

# 测试聊天接口
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "qd-job-turbo",
    "messages": [
      {"role": "user", "content": "你好，帮我推荐城阳区的Java开发岗位"}
    ],
    "stream": false
  }'
```

## 自动化构建脚本

### 构建脚本 (build-multiarch.sh)

```bash
#!/bin/bash

set -e

# 配置
IMAGE_NAME="qd-sc"
VERSION="v1.0"
REGISTRY="t0ng7u"

echo "=== 青岛岗位匹配系统 Docker 多架构构建 ==="

# 创建buildx构建器
echo "1. 创建buildx构建器..."
docker buildx create --name multiarch-builder --use 2>/dev/null || true
docker buildx inspect --bootstrap

# 构建AMD64镜像
echo "2. 构建AMD64镜像..."
docker buildx build --platform linux/amd64 -t ${IMAGE_NAME}:${VERSION}-amd64 --load .

# 构建ARM64镜像
echo "3. 构建ARM64镜像..."
docker buildx build --platform linux/arm64 -t ${IMAGE_NAME}:${VERSION}-arm64 --load .

# 验证镜像
echo "4. 验证镜像架构..."
echo "AMD64 架构:"
docker inspect ${IMAGE_NAME}:${VERSION}-amd64 | grep Architecture
echo "ARM64 架构:"
docker inspect ${IMAGE_NAME}:${VERSION}-arm64 | grep Architecture

# 保存镜像
echo "5. 保存镜像文件..."
docker save ${IMAGE_NAME}:${VERSION}-amd64 | gzip > ${IMAGE_NAME}-${VERSION}-amd64.tar.gz
docker save ${IMAGE_NAME}:${VERSION}-arm64 | gzip > ${IMAGE_NAME}-${VERSION}-arm64.tar.gz

# 显示文件大小
echo "6. 镜像文件大小:"
ls -lh ${IMAGE_NAME}-${VERSION}-*.tar.gz

echo "=== 构建完成 ==="
echo "AMD64 镜像: ${IMAGE_NAME}-${VERSION}-amd64.tar.gz"
echo "ARM64 镜像: ${IMAGE_NAME}-${VERSION}-arm64.tar.gz"

# 可选：推送到仓库
read -p "是否推送到Docker仓库? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "7. 推送多架构镜像到仓库..."
    docker buildx build --platform linux/amd64,linux/arm64 -t ${REGISTRY}/${IMAGE_NAME}:${VERSION} --push .
    docker buildx build --platform linux/amd64,linux/arm64 -t ${REGISTRY}/${IMAGE_NAME}:latest --push .
    echo "推送完成!"
fi
```

### 部署脚本 (deploy.sh)

```bash
#!/bin/bash

set -e

# 检测系统架构
ARCH=$(uname -m)
case $ARCH in
    x86_64)
        PLATFORM="amd64"
        COMPOSE_FILE="docker-compose-amd64.yml"
        ;;
    aarch64|arm64)
        PLATFORM="arm64"
        COMPOSE_FILE="docker-compose-arm64.yml"
        ;;
    *)
        echo "不支持的架构: $ARCH"
        exit 1
        ;;
esac

echo "=== 青岛岗位匹配系统部署 ==="
echo "检测到架构: $ARCH -> 使用 $PLATFORM 镜像"

# 检查必要文件
if [ ! -f "$COMPOSE_FILE" ]; then
    echo "错误: 找不到 $COMPOSE_FILE"
    exit 1
fi

if [ ! -f "config.yaml" ]; then
    echo "错误: 找不到 config.yaml 配置文件"
    exit 1
fi

if [ ! -f ".env" ]; then
    echo "警告: 找不到 .env 文件，请确保环境变量已正确配置"
fi

# 停止现有服务
echo "1. 停止现有服务..."
docker-compose -f $COMPOSE_FILE down 2>/dev/null || true

# 启动服务
echo "2. 启动服务..."
docker-compose -f $COMPOSE_FILE up -d

# 等待服务启动
echo "3. 等待服务启动..."
sleep 10

# 健康检查
echo "4. 健康检查..."
for i in {1..30}; do
    if curl -s http://localhost:8080/health > /dev/null; then
        echo "✓ 服务启动成功!"
        break
    fi
    echo "等待服务启动... ($i/30)"
    sleep 2
done

# 显示服务状态
echo "5. 服务状态:"
docker-compose -f $COMPOSE_FILE ps

echo "=== 部署完成 ==="
echo "服务地址: http://localhost:8080"
echo "健康检查: http://localhost:8080/health"
echo "API文档: http://localhost:8080/"
```

## 故障排除

### 常见问题

#### 1. 架构不匹配错误

```bash
# 问题：exec format error
# 原因：在ARM64系统上运行AMD64镜像，或反之

# 解决方案：检查镜像实际架构
docker inspect qd-sc:v1.0-arm64 | grep Architecture

# 如果架构不匹配，重新构建正确架构的镜像
docker buildx build --platform linux/arm64 -t qd-sc:v1.0-arm64 --load .
```

#### 2. buildx构建失败

```bash
# 检查buildx是否可用
docker buildx version

# 重新创建构建器
docker buildx rm multiarch-builder
docker buildx create --name multiarch-builder --use

# 清理构建缓存
docker buildx prune -f
```

#### 3. Go编译错误

```bash
# 问题：CGO相关错误
# 解决：确保Dockerfile中设置了CGO_ENABLED=0

# 问题：依赖下载失败
# 解决：使用Go代理
docker build --build-arg GOPROXY=https://goproxy.cn,direct .
```

#### 4. 服务启动失败

```bash
# 检查配置文件
docker run --rm -v $(pwd)/config.yaml:/app/config.yaml qd-sc:v1.0-amd64 cat /app/config.yaml

# 检查环境变量
docker-compose -f docker-compose-amd64.yml config

# 查看详细日志
docker-compose -f docker-compose-amd64.yml logs --tail=100
```

### 性能优化建议

#### ARM64平台优化

1. **Go运行时优化**
   ```bash
   # 限制CPU核心数
   export GOMAXPROCS=4
   
   # 设置内存限制
   export GOMEMLIMIT=2GiB
   ```

2. **容器资源限制**
   ```yaml
   deploy:
     resources:
       limits:
         memory: 2G
         cpus: '4'
   ```

3. **启动时间优化**
   - 增加健康检查的 `start_period`
   - 预热关键服务连接
   - 使用更小的基础镜像

#### 构建优化

1. **多阶段构建**
   - 当前Dockerfile已使用多阶段构建
   - 构建阶段使用golang:1.21-alpine
   - 运行阶段使用alpine:3.19

2. **依赖缓存**
   ```dockerfile
   # 优先复制go.mod和go.sum
   COPY go.mod go.sum ./
   RUN go mod download
   ```

3. **编译优化**
   ```dockerfile
   # 减小二进制大小
   RUN CGO_ENABLED=0 go build -ldflags="-s -w" -trimpath
   ```

## 总结

本指南提供了青岛岗位匹配系统的完整Docker多架构构建和部署方案。主要特点：

- ✅ 支持AMD64和ARM64双架构
- ✅ 使用buildx进行跨平台构建
- ✅ 提供完整的docker-compose配置
- ✅ 包含自动化构建和部署脚本
- ✅ 针对不同架构进行性能优化
- ✅ 详细的故障排除指南

按照本指南操作，可以在不同架构的服务器上成功部署和运行青岛岗位匹配系统。