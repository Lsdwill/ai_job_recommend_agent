@echo off
setlocal enabledelayedexpansion

REM 配置
set IMAGE_NAME=qd-sc
set VERSION=v1.0
set REGISTRY=t0ng7u

echo === 青岛岗位匹配系统 Docker 多架构构建 ===

REM 创建buildx构建器
echo 1. 创建buildx构建器...
docker buildx create --name multiarch-builder --use 2>nul
docker buildx inspect --bootstrap

REM 构建AMD64镜像
echo 2. 构建AMD64镜像...
docker buildx build --platform linux/amd64 -t %IMAGE_NAME%:%VERSION%-amd64 --load .

REM 构建ARM64镜像
echo 3. 构建ARM64镜像...
docker buildx build --platform linux/arm64 -t %IMAGE_NAME%:%VERSION%-arm64 --load .

REM 验证镜像
echo 4. 验证镜像架构...
echo AMD64 架构:
docker inspect %IMAGE_NAME%:%VERSION%-amd64 | findstr Architecture
echo ARM64 架构:
docker inspect %IMAGE_NAME%:%VERSION%-arm64 | findstr Architecture

REM 保存镜像
echo 5. 保存镜像文件...
docker save %IMAGE_NAME%:%VERSION%-amd64 | gzip > %IMAGE_NAME%-%VERSION%-amd64.tar.gz
docker save %IMAGE_NAME%:%VERSION%-arm64 | gzip > %IMAGE_NAME%-%VERSION%-arm64.tar.gz

REM 显示文件大小
echo 6. 镜像文件大小:
dir /s %IMAGE_NAME%-%VERSION%-*.tar.gz

echo === 构建完成 ===
echo AMD64 镜像: %IMAGE_NAME%-%VERSION%-amd64.tar.gz
echo ARM64 镜像: %IMAGE_NAME%-%VERSION%-arm64.tar.gz

REM 可选：推送到仓库
set /p PUSH="是否推送到Docker仓库? (y/N): "
if /i "%PUSH%"=="y" (
    echo 7. 推送多架构镜像到仓库...
    docker buildx build --platform linux/amd64,linux/arm64 -t %REGISTRY%/%IMAGE_NAME%:%VERSION% --push .
    docker buildx build --platform linux/amd64,linux/arm64 -t %REGISTRY%/%IMAGE_NAME%:latest --push .
    echo 推送完成!
)

pause