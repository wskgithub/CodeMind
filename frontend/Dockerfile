# ============================================================
# CodeMind Frontend — Multi-stage Docker Build
# ============================================================

# ──────────────────────────────────
# Stage 1: Build
# ──────────────────────────────────
# 构建上下文为仓库根目录，见 docker-compose.yml
FROM node:20-alpine AS builder

WORKDIR /build

# 复制根目录 VERSION 文件（vite.config.ts 通过 ../VERSION 读取，统一版本号来源）
COPY VERSION /VERSION

# 先复制依赖文件，利用 Docker 缓存
COPY frontend/package.json frontend/package-lock.json* ./
RUN npm ci --no-audit --no-fund

# 复制源代码并构建
COPY frontend/ ./
RUN npm run build

# ──────────────────────────────────
# Stage 2: Serve with Nginx
# ──────────────────────────────────
FROM nginx:1.27-alpine

# Nginx 配置通过 docker-compose volumes 注入，不在镜像中硬编码
COPY --from=builder /build/dist /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
