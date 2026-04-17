# ============================================================
# CodeMind Frontend — Multi-stage Docker Build
# ============================================================

# ──────────────────────────────────
# Stage 1: Build
# ──────────────────────────────────
FROM node:20-alpine AS builder

WORKDIR /build

# 先复制依赖文件，利用 Docker 缓存
COPY package.json package-lock.json* ./
RUN npm ci --no-audit --no-fund

# 复制源代码并构建
COPY . .
RUN npm run build

# ──────────────────────────────────
# Stage 2: Serve with Nginx
# ──────────────────────────────────
FROM nginx:1.27-alpine

# Nginx 配置通过 docker-compose volumes 注入，不在镜像中硬编码
COPY --from=builder /build/dist /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
