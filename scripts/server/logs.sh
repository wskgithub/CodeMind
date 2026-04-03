#!/usr/bin/env bash
# ============================================================
# CodeMind 日志查看脚本
# ============================================================
# 查看 Docker 容器日志，支持指定服务和实时跟踪
#
# 用法:
#   bash scripts/logs.sh                    # 查看所有服务最近 100 行
#   bash scripts/logs.sh -f                 # 实时跟踪所有日志
#   bash scripts/logs.sh backend            # 查看后端日志
#   bash scripts/logs.sh backend -f         # 实时跟踪后端日志
#   bash scripts/logs.sh postgres -n 200    # 查看 PostgreSQL 最近 200 行
# ============================================================
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/utils.sh"

INSTALL_DIR=$(get_install_dir)

if [ ! -f "$INSTALL_DIR/docker-compose.yml" ]; then
    log_error "未找到 CodeMind 安装: $INSTALL_DIR"
    exit 1
fi

SERVICE=""
FOLLOW=false
LINES=100
EXTRA_ARGS=()

while [[ $# -gt 0 ]]; do
    case $1 in
        -f|--follow)   FOLLOW=true; shift ;;
        -n|--lines)    LINES="$2"; shift 2 ;;
        --help|-h)
            echo "用法: bash $0 [service] [-f] [-n lines]"
            echo ""
            echo "服务: frontend | backend | postgres | redis"
            echo ""
            echo "选项:"
            echo "  -f, --follow   实时跟踪日志输出"
            echo "  -n, --lines    显示最近 N 行（默认 100）"
            echo ""
            echo "示例:"
            echo "  bash $0 backend -f       # 实时查看后端日志"
            echo "  bash $0 postgres -n 50   # 查看数据库最近 50 行"
            exit 0 ;;
        -*) EXTRA_ARGS+=("$1"); shift ;;
        *)  SERVICE="$1"; shift ;;
    esac
done

cd "$INSTALL_DIR"

CMD_ARGS=(--tail "$LINES" --timestamps)

if [ "$FOLLOW" = true ]; then
    CMD_ARGS+=(-f)
fi

if [ ${#EXTRA_ARGS[@]} -gt 0 ]; then
    CMD_ARGS+=("${EXTRA_ARGS[@]}")
fi

if [ -n "$SERVICE" ]; then
    docker compose logs "${CMD_ARGS[@]}" "$SERVICE"
else
    docker compose logs "${CMD_ARGS[@]}"
fi
