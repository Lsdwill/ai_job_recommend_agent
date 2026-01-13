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