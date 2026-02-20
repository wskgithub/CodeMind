#!/usr/bin/env bash
# ============================================================
# CodeMind 打包脚本
# ============================================================
# 在开发机器上运行，将前后端编译打包为可部署的压缩包
# 后端交叉编译为 linux/amd64 平台二进制
# 前端构建为静态文件
#
# 用法: bash scripts/package.sh
# ============================================================
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

VERSION=$(tr -d '[:space:]' < "$PROJECT_DIR/VERSION")
PACKAGE_NAME="codemind-v${VERSION}"
BUILD_DIR="/tmp/${PACKAGE_NAME}-build-$$"
OUTPUT_DIR="$PROJECT_DIR/dist"

# ── 颜色输出 ──
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info()  { echo -e "${BLUE}[INFO]${NC} $1"; }
log_ok()    { echo -e "${GREEN}[OK]${NC} $1"; }
log_warn()  { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# ── 清理函数 ──
cleanup() {
    if [ -d "$BUILD_DIR" ]; then
        rm -rf "$BUILD_DIR"
    fi
}
trap cleanup EXIT

echo ""
echo "╔══════════════════════════════════════════════════╗"
echo "║     CodeMind 度影智能编码服务 — 打包工具          ║"
echo "║     版本: v${VERSION}                                 ║"
echo "╚══════════════════════════════════════════════════╝"
echo ""

# ── 环境检查 ──
log_info "检查构建环境..."

if ! command -v go &>/dev/null; then
    log_error "Go 未安装，请安装 Go 1.23+"
    exit 1
fi

if ! command -v node &>/dev/null || ! command -v npm &>/dev/null; then
    log_error "Node.js/npm 未安装，请安装 Node.js 20+"
    exit 1
fi

GO_VERSION=$(go version | sed 's/.*go\([0-9]*\.[0-9]*\).*/\1/')
NODE_VERSION=$(node -v)
log_ok "Go ${GO_VERSION}, Node ${NODE_VERSION}"

# ── 初始化构建目录 ──
rm -rf "$BUILD_DIR"
PKG="$BUILD_DIR/$PACKAGE_NAME"
mkdir -p "$PKG"/{frontend,backend,config,docker/{nginx,postgres},migrations,scripts}

# ============================================================
# Step 1: 构建前端
# ============================================================
echo ""
log_info "━━━ [1/5] 构建前端 ━━━"

cd "$PROJECT_DIR/frontend"

log_info "安装依赖..."
npm ci --no-audit --no-fund --silent 2>&1 | tail -1

log_info "执行构建..."
npm run build

cp -r dist "$PKG/frontend/"
log_ok "前端构建完成 ($(du -sh dist | cut -f1))"

# ============================================================
# Step 2: 交叉编译后端
# ============================================================
echo ""
log_info "━━━ [2/5] 交叉编译后端 (linux/amd64) ━━━"

cd "$PROJECT_DIR/backend"

GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(date -u '+%Y-%m-%dT%H:%M:%SZ')

log_info "目标平台: linux/amd64"
log_info "版本信息: ${VERSION} / ${GIT_COMMIT} / ${BUILD_TIME}"

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags "-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
    -o "$PKG/backend/codemind" \
    cmd/server/main.go

BINARY_SIZE=$(du -sh "$PKG/backend/codemind" | cut -f1)
log_ok "后端编译完成 (${BINARY_SIZE})"

# ============================================================
# Step 3: 收集部署文件
# ============================================================
echo ""
log_info "━━━ [3/5] 收集部署文件 ━━━"

# 生产 Dockerfiles
cp "$PROJECT_DIR/deploy/production/docker/Dockerfile.frontend" "$PKG/frontend/Dockerfile"
cp "$PROJECT_DIR/deploy/production/docker/Dockerfile.backend" "$PKG/backend/Dockerfile"

# Docker Compose
cp "$PROJECT_DIR/deploy/production/docker-compose.prod.yml" "$PKG/docker-compose.yml"

# 配置模板
cp "$PROJECT_DIR/deploy/production/.env.production" "$PKG/.env.template"
cp "$PROJECT_DIR/deploy/production/config/app.yaml.production" "$PKG/config/app.yaml.template"

# Nginx 配置
cp "$PROJECT_DIR/deploy/production/docker/nginx/nginx.prod.conf" "$PKG/docker/nginx/nginx.conf"

# 数据库初始化脚本
cp "$PROJECT_DIR/deploy/docker/postgres/init.sql" "$PKG/docker/postgres/init.sql"
cp "$PROJECT_DIR/deploy/docker/postgres/seed.sql" "$PKG/docker/postgres/seed.sql"

# 数据库迁移脚本
if ls "$PROJECT_DIR/backend/migrations/"*.sql &>/dev/null; then
    cp "$PROJECT_DIR/backend/migrations/"*.sql "$PKG/migrations/"
    MIGRATION_COUNT=$(ls "$PKG/migrations/"*.sql | wc -l | tr -d ' ')
    log_info "包含 ${MIGRATION_COUNT} 个数据库迁移文件"
fi

# 服务器端脚本
cp "$PROJECT_DIR/scripts/server/"*.sh "$PKG/scripts/"
chmod +x "$PKG/scripts/"*.sh

# 版本文件
cp "$PROJECT_DIR/VERSION" "$PKG/VERSION"

log_ok "部署文件收集完成"

# ============================================================
# Step 4: 打包压缩
# ============================================================
echo ""
log_info "━━━ [4/5] 创建压缩包 ━━━"

mkdir -p "$OUTPUT_DIR"
cd "$BUILD_DIR"
tar -czf "$OUTPUT_DIR/${PACKAGE_NAME}.tar.gz" "$PACKAGE_NAME"

PACKAGE_SIZE=$(du -sh "$OUTPUT_DIR/${PACKAGE_NAME}.tar.gz" | cut -f1)
log_ok "压缩包: dist/${PACKAGE_NAME}.tar.gz (${PACKAGE_SIZE})"

# ============================================================
# Step 5: 生成校验和
# ============================================================
echo ""
log_info "━━━ [5/5] 生成校验和 ━━━"

cd "$OUTPUT_DIR"
if command -v sha256sum &>/dev/null; then
    sha256sum "${PACKAGE_NAME}.tar.gz" > "${PACKAGE_NAME}.tar.gz.sha256"
elif command -v shasum &>/dev/null; then
    shasum -a 256 "${PACKAGE_NAME}.tar.gz" > "${PACKAGE_NAME}.tar.gz.sha256"
fi
log_ok "SHA256 校验文件已生成"

# ============================================================
# 完成
# ============================================================
echo ""
echo "╔══════════════════════════════════════════════════╗"
echo "║                    打包成功！                     ║"
echo "╚══════════════════════════════════════════════════╝"
echo ""
echo "  版本:    v${VERSION}"
echo "  平台:    linux/amd64"
echo "  文件:    dist/${PACKAGE_NAME}.tar.gz"
echo "  大小:    ${PACKAGE_SIZE}"
echo "  校验和:  dist/${PACKAGE_NAME}.tar.gz.sha256"
echo ""
echo "  ┌─ 部署步骤 ────────────────────────────────────┐"
echo "  │                                                │"
echo "  │  1. 上传到服务器:                               │"
echo "  │     scp dist/${PACKAGE_NAME}.tar.gz user@host:/tmp/"
echo "  │                                                │"
echo "  │  2. 在服务器上解压:                             │"
echo "  │     tar -xzf ${PACKAGE_NAME}.tar.gz"
echo "  │                                                │"
echo "  │  3. 执行部署:                                   │"
echo "  │     cd ${PACKAGE_NAME}                "
echo "  │     sudo bash scripts/deploy.sh                │"
echo "  │                                                │"
echo "  └────────────────────────────────────────────────┘"
echo ""
