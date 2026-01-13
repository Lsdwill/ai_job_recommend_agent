# syntax=docker/dockerfile:1.6

# 构建阶段（支持 buildx 多架构）
FROM --platform=$BUILDPLATFORM golang:1.21-alpine AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

# 复制依赖文件（优先复制，利用 Docker 缓存）
COPY go.mod go.sum ./
RUN go mod download

# 仅复制必要的源代码目录
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY pkg/ ./pkg/

# 编译 - 优化参数减小二进制大小和加快编译
RUN CGO_ENABLED=0 \
    GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-$(go env GOARCH)} \
    go build -ldflags="-s -w" -trimpath \
    -o qd-sc-server cmd/server/main.go

# 运行阶段 - 使用最小化的 alpine
FROM alpine:3.19

# 安装必要的工具（ca-certificates 用于 HTTPS，wget 用于 healthcheck）
RUN apk --no-cache add ca-certificates tzdata wget && \
    addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser && \
    mkdir -p /app && \
    chown -R appuser:appuser /app

WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder --chown=appuser:appuser /app/qd-sc-server .

# 切换到非 root 用户
USER appuser

# 暴露端口
EXPOSE 8080

# 运行
ENTRYPOINT ["./qd-sc-server"]
CMD ["-config", "/app/config.yaml"]

