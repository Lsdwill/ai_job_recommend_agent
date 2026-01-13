@echo off
setlocal enabledelayedexpansion

REM 检测系统架构（Windows上通常是AMD64）
set PLATFORM=amd64
set COMPOSE_FILE=docker-compose-amd64.yml

echo === 青岛岗位匹配系统部署 ===
echo 使用 %PLATFORM% 镜像

REM 检查必要文件
if not exist "%COMPOSE_FILE%" (
    echo 错误: 找不到 %COMPOSE_FILE%
    pause
    exit /b 1
)

if not exist "config.yaml" (
    echo 错误: 找不到 config.yaml 配置文件
    pause
    exit /b 1
)

if not exist ".env" (
    echo 警告: 找不到 .env 文件，请确保环境变量已正确配置
)

REM 停止现有服务
echo 1. 停止现有服务...
docker-compose -f %COMPOSE_FILE% down 2>nul

REM 启动服务
echo 2. 启动服务...
docker-compose -f %COMPOSE_FILE% up -d

REM 等待服务启动
echo 3. 等待服务启动...
timeout /t 10 /nobreak >nul

REM 健康检查
echo 4. 健康检查...
for /l %%i in (1,1,30) do (
    curl -s http://localhost:8080/health >nul 2>&1
    if !errorlevel! equ 0 (
        echo ✓ 服务启动成功!
        goto :health_ok
    )
    echo 等待服务启动... (%%i/30)
    timeout /t 2 /nobreak >nul
)

:health_ok
REM 显示服务状态
echo 5. 服务状态:
docker-compose -f %COMPOSE_FILE% ps

echo === 部署完成 ===
echo 服务地址: http://localhost:8080
echo 健康检查: http://localhost:8080/health
echo API文档: http://localhost:8080/

pause