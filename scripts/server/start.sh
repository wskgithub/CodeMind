#!/usr/bin/env bash
# ============================================================
# CodeMind 启动服务脚本
# ============================================================
# 启动所有 Docker 容器并执行健康检查
#
# 用法: bash scripts/start.sh
# ============================================================
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/utils.sh"

INSTALL_DIR=$(get_install_dir)

if [ ! -f "$INSTALL_DIR/docker-compose.yml" ]; then
    log_error "未找到 CodeMind 安装: $INSTALL_DIR"
    exit 1
fi

VERSION=$(tr -d '[:space:]' < "$INSTALL_DIR/VERSION")

echo ""
log_step "启动 CodeMind v${VERSION} ..."
echo ""

cd "$INSTALL_DIR"

# 检查是否已在运行
RUNNING=$(docker compose ps --status running -q 2>/dev/null | wc -l | tr -d ' ')
if [ "$RUNNING" -gt 0 ]; then
    log_warn "检测到 ${RUNNING} 个容器已在运行"
    read -rp "是否重新启动所有服务？(Y/n) " confirm
    if [[ "$confirm" =~ ^[Nn]$ ]]; then
        echo "已取消"
        exit 0
    fi
    docker compose down 2>/dev/null || true
    echo ""
fi

docker compose up -d

echo ""
log_info "等待服务就绪..."

# 健康检查
wait_for_service "PostgreSQL" \
    "docker compose exec -T postgres pg_isready -U codemind -d codemind" 60 || {
    log_error "PostgreSQL 启动失败"
    log_info "查看日志: cd $INSTALL_DIR && docker compose logs postgres"
    exit 1
}

wait_for_service "Redis" \
    "docker compose exec -T redis redis-cli --no-auth-warning ping" 30 || {
    log_warn "Redis 健康检查未通过，可能需要密码认证（服务本身可能已正常）"
}

wait_for_service "后端服务" \
    "docker compose exec -T backend wget -qO- http://localhost:8080/health" 60 || {
    log_error "后端启动失败"
    log_info "查看日志: cd $INSTALL_DIR && docker compose logs backend"
    exit 1
}

wait_for_service "前端服务" \
    "docker compose exec -T frontend wget -qO- http://localhost:80/nginx-health" 30 || {
    log_warn "前端健康检查未响应，请手动验证"
}

SERVER_IP=$(get_server_ip)
FRONTEND_PORT=$(get_env_value "FRONTEND_PORT" "$INSTALL_DIR/.env" "18080")

echo ""
echo "╔══════════════════════════════════════════════════╗"
echo "║               服务已启动                          ║"
echo "╚══════════════════════════════════════════════════╝"
echo ""
echo "  访问地址:  http://${SERVER_IP}:${FRONTEND_PORT}"
echo "  查看状态:  bash $INSTALL_DIR/scripts/status.sh"
echo "  查看日志:  bash $INSTALL_DIR/scripts/logs.sh -f"
echo ""
