#!/usr/bin/env bash
# ============================================================
# CodeMind 服务状态查看脚本
# ============================================================
# 显示所有容器状态、资源占用和磁盘使用
#
# 用法: bash scripts/status.sh
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
SERVER_IP=$(get_server_ip)
FRONTEND_PORT=$(get_env_value "FRONTEND_PORT" "$INSTALL_DIR/.env" "18080")

echo ""
echo "╔══════════════════════════════════════════════════╗"
echo "║     CodeMind 服务状态  v${VERSION}                     ║"
echo "╚══════════════════════════════════════════════════╝"
echo ""

cd "$INSTALL_DIR"

# ── 容器状态 ──
echo "── 容器状态 ───────────────────────────────────────"
docker compose ps
echo ""

# ── 资源占用 ──
RUNNING=$(docker compose ps --status running -q 2>/dev/null | wc -l | tr -d ' ')
if [ "$RUNNING" -gt 0 ]; then
    echo "── 资源占用 ───────────────────────────────────────"
    docker stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.NetIO}}" \
        $(docker compose ps -q 2>/dev/null) 2>/dev/null || true
    echo ""
fi

# ── 磁盘使用 ──
echo "── 磁盘使用 ───────────────────────────────────────"

PG_VOL_SIZE=$(docker volume inspect codemind_codemind_postgres_data --format '{{.Mountpoint}}' 2>/dev/null | xargs du -sh 2>/dev/null | cut -f1 || echo "N/A")
REDIS_VOL_SIZE=$(docker volume inspect codemind_codemind_redis_data --format '{{.Mountpoint}}' 2>/dev/null | xargs du -sh 2>/dev/null | cut -f1 || echo "N/A")
BACKUP_SIZE=$(du -sh "$INSTALL_DIR/backups" 2>/dev/null | cut -f1 || echo "0")
BACKUP_COUNT=$(find "$INSTALL_DIR/backups" -name "*.tar.gz" 2>/dev/null | wc -l | tr -d ' ')

echo "  PostgreSQL 数据:  ${PG_VOL_SIZE}"
echo "  Redis 数据:       ${REDIS_VOL_SIZE}"
echo "  备份文件:         ${BACKUP_SIZE} (${BACKUP_COUNT} 个)"
echo ""

# ── 连通性检查 ──
echo "── 连通性检查 ─────────────────────────────────────"

if docker compose exec -T postgres pg_isready -U codemind -d codemind &>/dev/null; then
    echo -e "  PostgreSQL:  ${GREEN}● 正常${NC}"
else
    echo -e "  PostgreSQL:  ${RED}● 异常${NC}"
fi

if docker compose exec -T redis redis-cli --no-auth-warning ping &>/dev/null; then
    echo -e "  Redis:       ${GREEN}● 正常${NC}"
else
    echo -e "  Redis:       ${YELLOW}● 需认证或异常${NC}"
fi

if docker compose exec -T backend wget -qO- http://localhost:8080/health &>/dev/null; then
    echo -e "  后端 API:    ${GREEN}● 正常${NC}"
else
    echo -e "  后端 API:    ${RED}● 异常${NC}"
fi

if docker compose exec -T frontend wget -qO- http://localhost:80/nginx-health &>/dev/null; then
    echo -e "  前端 Nginx:  ${GREEN}● 正常${NC}"
else
    echo -e "  前端 Nginx:  ${RED}● 异常${NC}"
fi

echo ""
echo "── 访问信息 ───────────────────────────────────────"
echo "  地址: http://${SERVER_IP}:${FRONTEND_PORT}"
echo ""
