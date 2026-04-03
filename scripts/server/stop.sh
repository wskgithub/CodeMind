#!/usr/bin/env bash
# ============================================================
# CodeMind 停止服务脚本
# ============================================================
# 停止所有 Docker 容器（数据卷保留）
#
# 用法:
#   bash scripts/stop.sh              # 停止服务
#   bash scripts/stop.sh --force      # 强制停止（不确认）
# ============================================================
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/utils.sh"

INSTALL_DIR=$(get_install_dir)

if [ ! -f "$INSTALL_DIR/docker-compose.yml" ]; then
    log_error "未找到 CodeMind 安装: $INSTALL_DIR"
    exit 1
fi

FORCE=false
while [[ $# -gt 0 ]]; do
    case $1 in
        --force|-f) FORCE=true; shift ;;
        --help|-h)
            echo "用法: bash $0 [--force]"
            echo ""
            echo "选项:"
            echo "  --force, -f   跳过确认直接停止"
            exit 0 ;;
        *) log_error "未知参数: $1"; exit 1 ;;
    esac
done

cd "$INSTALL_DIR"

# 检查运行状态
RUNNING=$(docker compose ps --status running -q 2>/dev/null | wc -l | tr -d ' ')
if [ "$RUNNING" -eq 0 ]; then
    log_info "当前没有运行中的服务"
    exit 0
fi

if [ "$FORCE" = false ]; then
    log_warn "即将停止所有 CodeMind 服务（${RUNNING} 个容器）"
    read -rp "确认停止？(Y/n) " confirm
    [[ "$confirm" =~ ^[Nn]$ ]] && { echo "已取消"; exit 0; }
fi

log_info "停止服务..."
docker compose down

log_ok "所有服务已停止"
echo ""
echo "  数据卷已保留，重启: bash $INSTALL_DIR/scripts/start.sh"
echo ""
