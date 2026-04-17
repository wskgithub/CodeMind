#!/usr/bin/env bash
# ============================================================
# CodeMind 重启服务脚本
# ============================================================
# 重启所有或指定的 Docker 容器
#
# 用法:
#   bash scripts/restart.sh              # 重启所有服务
#   bash scripts/restart.sh backend      # 仅重启后端
#   bash scripts/restart.sh frontend     # 仅重启前端
#   bash scripts/restart.sh --full       # 完全重建（down + up）
# ============================================================
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/utils.sh"

INSTALL_DIR=$(get_install_dir)

if [ ! -f "$INSTALL_DIR/docker-compose.yml" ]; then
    log_error "未找到 CodeMind 安装: $INSTALL_DIR"
    exit 1
fi

FULL_RESTART=false
SERVICE=""

while [[ $# -gt 0 ]]; do
    case $1 in
        --full)     FULL_RESTART=true; shift ;;
        --help|-h)
            echo "用法: bash $0 [service] [--full]"
            echo ""
            echo "参数:"
            echo "  service       指定服务名（frontend/backend/postgres/redis）"
            echo ""
            echo "选项:"
            echo "  --full        完全重建（docker compose down + up）"
            exit 0 ;;
        -*) log_error "未知选项: $1"; exit 1 ;;
        *)  SERVICE="$1"; shift ;;
    esac
done

cd "$INSTALL_DIR"

if [ "$FULL_RESTART" = true ]; then
    log_info "完全重启：停止所有服务..."
    docker compose down
    log_info "重新启动所有服务..."
    docker compose up -d
    log_ok "完全重启完成"
elif [ -n "$SERVICE" ]; then
    log_info "重启服务: ${SERVICE}"
    docker compose restart "$SERVICE"
    log_ok "${SERVICE} 已重启"
else
    log_info "重启所有服务..."
    docker compose restart
    log_ok "所有服务已重启"
fi

echo ""
log_info "当前状态:"
docker compose ps
echo ""
