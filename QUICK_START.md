# 快速开始指南

## 1. 准备工作

### 检查系统架构
```bash
# Linux/Mac
uname -m

# Windows PowerShell
$env:PROCESSOR_ARCHITECTURE
```

### 安装Docker和Docker Compose
确保已安装Docker和Docker Compose，并启用buildx功能。

## 2. 配置环境

### 复制配置文件
```bash
# 复制环境变量配置
cp .env.example .env

# 编辑配置文件，填入真实的API密钥
# 必填项：LLM_API_KEY, AMAP_API_KEY
```

### 编辑config.yaml
确保config.yaml中的配置正确，特别是API地址和密钥。

## 3. 构建镜像

### Linux/Mac系统
```bash
# 给脚本添加执行权限
chmod +x build-multiarch.sh deploy.sh

# 构建多架构镜像
./build-multiarch.sh
```

### Windows系统
```cmd
# 运行构建脚本
build-multiarch.bat
```

## 4. 部署服务

### 自动部署（推荐）
```bash
# Linux/Mac
./deploy.sh

# Windows
deploy.bat
```

### 手动部署
```bash
# 检查系统架构选择对应的compose文件
# AMD64系统
docker-compose -f docker-compose-amd64.yml up -d

# ARM64系统
docker-compose -f docker-compose-arm64.yml up -d
```

## 5. 验证服务

```bash
# 健康检查
curl http://localhost:8080/health

# 测试API
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "qd-job-turbo",
    "messages": [
      {"role": "user", "content": "你好"}
    ],
    "stream": false
  }'
```

## 6. 常用命令

```bash
# 查看日志
docker-compose -f docker-compose-amd64.yml logs -f

# 重启服务
docker-compose -f docker-compose-amd64.yml restart

# 停止服务
docker-compose -f docker-compose-amd64.yml down

# 查看服务状态
docker-compose -f docker-compose-amd64.yml ps
```

## 故障排除

### 服务无法启动
1. 检查配置文件是否正确
2. 检查环境变量是否设置
3. 查看容器日志排查错误

### 架构不匹配
1. 确认系统架构
2. 使用对应的docker-compose文件
3. 重新构建正确架构的镜像

### API调用失败
1. 检查LLM_API_KEY是否正确
2. 检查网络连接
3. 查看服务日志确认错误信息