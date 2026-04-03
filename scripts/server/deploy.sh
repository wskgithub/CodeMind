#!/usr/bin/env bash
# ============================================================
# CodeMind 一键部署脚本
# ============================================================
# 在服务器上运行，完成首次部署或全新安装
# 用法: sudo bash scripts/deploy.sh [--install-dir /opt/codemind]
#
# 部署流程:
#   1. 环境检查（Docker、端口）
#   2. 创建安装目录
#   3. 安装文件
#   4. 生成配置（自动生成安全密码）
#   5. 构建镜像并启动服务
#   6. 健康检查与迁移
# ============================================================
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PACKAGE_DIR="$(dirname "$SCRIPT_DIR")"
source "$SCRIPT_DIR/utils.sh"

VERSION=$(tr -d '[:space:]' < "$PACKAGE_DIR/VERSION")

# ── 解析参数 ──
INSTALL_DIR="${CODEMIND_HOME:-/opt/codemind}"
while [[ $# -gt 0 ]]; do
    case $1 in
        --install-dir) INSTALL_DIR="$2"; shift 2 ;;
        --help|-h)
            echo "用法: sudo bash $0 [--install-dir /opt/codemind]"
            exit 0 ;;
        *) log_error "未知参数: $1"; exit 1 ;;
    esac
done

echo ""
echo "╔══════════════════════════════════════════════════╗"
echo "║     CodeMind — 一键部署                           ║"
echo "║     版本: v${VERSION}                                 ║"
echo "╚══════════════════════════════════════════════════╝"
echo ""

# ── 权限检查 ──
if [ "$EUID" -ne 0 ]; then
    log_error "请使用 sudo 运行此脚本"
    echo "  用法: sudo bash $0"
    exit 1
fi

# ── 检测已有安装 ──
if [ -f "$INSTALL_DIR/VERSION" ]; then
    INSTALLED_VERSION=$(tr -d '[:space:]' < "$INSTALL_DIR/VERSION")
    log_warn "检测到已安装的 CodeMind v${INSTALLED_VERSION}"
    log_warn "如需升级请使用: sudo bash ${INSTALL_DIR}/scripts/upgrade.sh"
    echo ""
    read -rp "是否覆盖安装？这将重建所有容器（数据库数据不受影响）(y/N) " confirm
    if [[ ! "$confirm" =~ ^[Yy]$ ]]; then
        echo "已取消"
        exit 0
    fi
    echo ""
fi

# ============================================================
# Step 1: 环境检查
# ============================================================
log_step "━━━ Step 1/6: 环境检查 ━━━"

check_docker || exit 1

# 读取端口配置（优先使用已有 .env，否则使用模板默认值）
if [ -f "$INSTALL_DIR/.env" ]; then
    FRONTEND_PORT=$(get_env_value "FRONTEND_PORT" "$INSTALL_DIR/.env" "18080")
    DB_EXTERNAL_PORT=$(get_env_value "DB_EXTERNAL_PORT" "$INSTALL_DIR/.env" "15432")
    REDIS_EXTERNAL_PORT=$(get_env_value "REDIS_EXTERNAL_PORT" "$INSTALL_DIR/.env" "16379")
else
    FRONTEND_PORT=$(get_env_value "FRONTEND_PORT" "$PACKAGE_DIR/.env.template" "18080")
    DB_EXTERNAL_PORT=$(get_env_value "DB_EXTERNAL_PORT" "$PACKAGE_DIR/.env.template" "15432")
    REDIS_EXTERNAL_PORT=$(get_env_value "REDIS_EXTERNAL_PORT" "$PACKAGE_DIR/.env.template" "16379")
fi

PORT_OK=true
check_port_available "$FRONTEND_PORT" "前端" || PORT_OK=false
check_port_available "$DB_EXTERNAL_PORT" "PostgreSQL" || PORT_OK=false
check_port_available "$REDIS_EXTERNAL_PORT" "Redis" || PORT_OK=false

if [ "$PORT_OK" = false ]; then
    # 如果是覆盖安装，端口被自己占用是正常的
    if [ -f "$INSTALL_DIR/VERSION" ]; then
        log_warn "端口被占用可能是已有 CodeMind 实例，将先停止旧服务"
        cd "$INSTALL_DIR" && docker compose down 2>/dev/null || true
    else
        log_error "请修改 .env.template 中的端口配置后重试"
        exit 1
    fi
fi

log_ok "环境检查通过"

# ============================================================
# Step 2: 创建安装目录
# ============================================================
echo ""
log_step "━━━ Step 2/6: 创建安装目录 ━━━"

mkdir -p "$INSTALL_DIR"/{config,docker/{nginx,postgres},migrations,scripts,backups,logs}
log_ok "安装目录: $INSTALL_DIR"

# ============================================================
# Step 3: 安装文件
# ============================================================
echo ""
log_step "━━━ Step 3/6: 安装文件 ━━━"

# 前端构建产物 + Dockerfile
cp -r "$PACKAGE_DIR/frontend" "$INSTALL_DIR/"
log_info "  前端文件已安装"

# 后端二进制 + Dockerfile
cp -r "$PACKAGE_DIR/backend" "$INSTALL_DIR/"
log_info "  后端文件已安装"

# Docker Compose
cp "$PACKAGE_DIR/docker-compose.yml" "$INSTALL_DIR/"

# Nginx 配置
cp "$PACKAGE_DIR/docker/nginx/nginx.conf" "$INSTALL_DIR/docker/nginx/"

# 数据库初始化
cp "$PACKAGE_DIR/docker/postgres/"*.sql "$INSTALL_DIR/docker/postgres/"

# 迁移文件
if ls "$PACKAGE_DIR/migrations/"*.sql &>/dev/null; then
    cp "$PACKAGE_DIR/migrations/"*.sql "$INSTALL_DIR/migrations/"
fi

# 服务器脚本
cp "$PACKAGE_DIR/scripts/"*.sh "$INSTALL_DIR/scripts/"
chmod +x "$INSTALL_DIR/scripts/"*.sh

# 版本文件
cp "$PACKAGE_DIR/VERSION" "$INSTALL_DIR/"

log_ok "文件安装完成"

# ============================================================
# Step 4: 配置
# ============================================================
echo ""
log_step "━━━ Step 4/6: 生成配置 ━━━"

if [ ! -f "$INSTALL_DIR/.env" ]; then
    log_info "首次部署，生成安全配置..."

    # 自动生成强密码
    DB_PASSWORD=$(generate_password 24)
    REDIS_PASSWORD=$(generate_password 24)
    JWT_SECRET=$(generate_password 48)

    # 从模板生成 .env
    cp "$PACKAGE_DIR/.env.template" "$INSTALL_DIR/.env"
    sed -i "s|^APP_VERSION=.*|APP_VERSION=${VERSION}|" "$INSTALL_DIR/.env"
    sed -i "s|^DB_PASSWORD=.*|DB_PASSWORD=${DB_PASSWORD}|" "$INSTALL_DIR/.env"
    sed -i "s|^REDIS_PASSWORD=.*|REDIS_PASSWORD=${REDIS_PASSWORD}|" "$INSTALL_DIR/.env"
    sed -i "s|^JWT_SECRET=.*|JWT_SECRET=${JWT_SECRET}|" "$INSTALL_DIR/.env"

    # 限制 .env 权限
    chmod 600 "$INSTALL_DIR/.env"

    log_ok ".env 已生成（密码已自动生成安全随机值）"

    # 生成 app.yaml
    cp "$PACKAGE_DIR/config/app.yaml.template" "$INSTALL_DIR/config/app.yaml"
    log_ok "app.yaml 已生成"

    # 交互配置 LLM 服务
    echo ""
    print_separator
    log_warn "需要配置 LLM 服务地址才能使用 AI 编码功能"
    print_separator
    echo ""

    read -rp "请输入 LLM 服务地址 (例: http://192.168.1.100:11434): " llm_url
    if [ -n "$llm_url" ]; then
        sed -i "s|^LLM_BASE_URL=.*|LLM_BASE_URL=${llm_url}|" "$INSTALL_DIR/.env"
        log_ok "LLM 服务地址: $llm_url"
    else
        log_warn "LLM 服务地址未配置，请稍后编辑 $INSTALL_DIR/.env"
    fi

    read -rp "LLM API Key (无则直接回车跳过): " llm_key
    if [ -n "$llm_key" ]; then
        sed -i "s|^LLM_API_KEY=.*|LLM_API_KEY=${llm_key}|" "$INSTALL_DIR/.env"
        log_ok "LLM API Key 已配置"
    fi
else
    # 保留现有配置，仅更新版本号
    sed -i "s|^APP_VERSION=.*|APP_VERSION=${VERSION}|" "$INSTALL_DIR/.env"
    log_ok "保留现有配置，已更新版本号"
fi

# ============================================================
# Step 5: 构建镜像并启动
# ============================================================
echo ""
log_step "━━━ Step 5/6: 构建镜像并启动服务 ━━━"

cd "$INSTALL_DIR"

log_info "构建 Docker 镜像..."
docker compose build --no-cache 2>&1 | while IFS= read -r line; do
    echo "  $line"
done

log_info "启动服务..."
docker compose up -d

log_ok "服务启动指令已发送"

# ============================================================
# Step 6: 健康检查
# ============================================================
echo ""
log_step "━━━ Step 6/6: 健康检查 ━━━"

cd "$INSTALL_DIR"

# 等待 PostgreSQL
wait_for_service "PostgreSQL" \
    "docker compose exec -T postgres pg_isready -U codemind -d codemind" 60 || {
    log_error "PostgreSQL 启动失败，查看日志: docker compose logs postgres"
    exit 1
}

# 等待 Redis
wait_for_service "Redis" \
    "docker compose exec -T redis redis-cli --no-auth-warning -a \"\$(grep '^REDIS_PASSWORD=' .env | cut -d= -f2-)\" ping" 30 || {
    log_error "Redis 启动失败，查看日志: docker compose logs redis"
    exit 1
}

# 等待后端
wait_for_service "后端服务" \
    "docker compose exec -T backend wget -qO- http://localhost:8080/health" 60 || {
    log_error "后端启动失败，查看日志: docker compose logs backend"
    exit 1
}

# 应用数据库迁移
MIGRATIONS_APPLIED="$INSTALL_DIR/migrations/.applied"
touch "$MIGRATIONS_APPLIED"

if ls "$INSTALL_DIR/migrations/"*.sql &>/dev/null; then
    log_info "应用数据库迁移..."
    MIGRATION_COUNT=0

    for sql_file in $(ls "$INSTALL_DIR/migrations/"*.sql | sort); do
        filename=$(basename "$sql_file")
        if ! grep -qF "$filename" "$MIGRATIONS_APPLIED"; then
            log_info "  执行: $filename"
            if docker compose exec -T postgres \
                psql -U codemind -d codemind < "$sql_file" 2>&1 | tail -5; then
                echo "$filename" >> "$MIGRATIONS_APPLIED"
                MIGRATION_COUNT=$((MIGRATION_COUNT + 1))
            else
                log_warn "  $filename 可能已部分应用（已标记为完成）"
                echo "$filename" >> "$MIGRATIONS_APPLIED"
            fi
        fi
    done

    if [ $MIGRATION_COUNT -gt 0 ]; then
        log_ok "已应用 $MIGRATION_COUNT 个迁移"
    else
        log_ok "无需执行新迁移"
    fi
fi

# 等待前端
wait_for_service "前端服务" \
    "docker compose exec -T frontend wget -qO- http://localhost:80/nginx-health" 30 || {
    log_warn "前端健康检查未响应，但服务可能仍在启动中"
}

# ============================================================
# 部署完成
# ============================================================
SERVER_IP=$(get_server_ip)
FRONTEND_PORT=$(get_env_value "FRONTEND_PORT" "$INSTALL_DIR/.env" "18080")

echo ""
echo "╔══════════════════════════════════════════════════╗"
echo "║               部署完成！                          ║"
echo "╚══════════════════════════════════════════════════╝"
echo ""
echo "  访问地址:  http://${SERVER_IP}:${FRONTEND_PORT}"
echo "  管理账号:  admin"
echo "  初始密码:  Admin@123456"
echo ""
echo "  安装目录:  $INSTALL_DIR"
echo "  配置文件:  $INSTALL_DIR/.env"
echo ""
echo "  ┌─ 常用命令 ─────────────────────────────────────┐"
echo "  │  查看状态:  cd $INSTALL_DIR && docker compose ps"
echo "  │  查看日志:  cd $INSTALL_DIR && docker compose logs -f"
echo "  │  备份数据:  sudo bash $INSTALL_DIR/scripts/backup.sh"
echo "  │  停止服务:  cd $INSTALL_DIR && docker compose down"
echo "  │  重启服务:  cd $INSTALL_DIR && docker compose restart"
echo "  └────────────────────────────────────────────────┘"
echo ""
log_warn "安全提醒: 请立即登录并修改管理员默认密码！"
echo ""
