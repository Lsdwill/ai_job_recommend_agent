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