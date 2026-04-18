# ============================================================
# CodeMind Backend — Multi-stage Docker Build
# ============================================================

# ──────────────────────────────────
# Stage 1: Build
# ──────────────────────────────────
FROM golang:1.25-alpine AS builder

# 安装构建依赖
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# 先复制依赖文件，利用 Docker 缓存
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建二进制
ARG VERSION=dev
ARG BUILD_TIME=unknown
ARG GIT_COMMIT=unknown

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
    -o /build/bin/codemind \
    cmd/server/main.go

# ──────────────────────────────────
# Stage 2: Runtime
# ──────────────────────────────────
FROM alpine:3.20

# 安装运行时依赖（wget 用于健康检查）
RUN apk add --no-cache ca-certificates tzdata wget

# 设置时区
ENV TZ=Asia/Shanghai

# 创建非 root 用户
RUN adduser -D -g '' appuser

WORKDIR /app

# 从构建阶段复制二进制
COPY --from=builder /build/bin/codemind .

# 创建配置目录
RUN mkdir -p /app/config /app/logs && \
    chown -R appuser:appuser /app

USER appuser

EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=5s --retries=3 \
    CMD wget -qO- http://localhost:8080/health || exit 1

CMD ["./codemind"]
