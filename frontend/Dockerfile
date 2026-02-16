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

# 复制 Nginx 配置
COPY --from=builder /build/../deploy/docker/nginx/nginx.conf /etc/nginx/conf.d/default.conf

# 复制构建产物
COPY --from=builder /build/dist /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
