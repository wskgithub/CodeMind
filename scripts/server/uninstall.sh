#!/usr/bin/env bash
# ============================================================
# CodeMind 卸载脚本
# ============================================================
# 完全卸载 CodeMind：停止服务、移除容器/镜像/数据卷
#
# 用法:
#   sudo bash scripts/uninstall.sh                   # 交互式卸载
#   sudo bash scripts/uninstall.sh --keep-data       # 保留数据库数据
#   sudo bash scripts/uninstall.sh --keep-backups    # 保留备份文件
# ============================================================
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/utils.sh"

INSTALL_DIR=$(get_install_dir)

# ── 解析参数 ──
KEEP_DATA=false
KEEP_BACKUPS=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --keep-data)    KEEP_DATA=true; shift ;;
        --keep-backups) KEEP_BACKUPS=true; shift ;;
        --help|-h)
            echo "用法: sudo bash $0 [--keep-data] [--keep-backups]"
            echo ""
            echo "选项:"
            echo "  --keep-data      保留数据库和 Redis 数据卷"
            echo "  --keep-backups   保留备份目录"
            exit 0 ;;
        *) log_error "未知参数: $1"; exit 1 ;;
    esac
done

if [ ! -f "$INSTALL_DIR/docker-compose.yml" ]; then
    log_error "未找到 CodeMind 安装: $INSTALL_DIR"
    exit 1
fi

VERSION=$(tr -d '[:space:]' < "$INSTALL_DIR/VERSION")

echo ""
echo "╔══════════════════════════════════════════════════╗"
echo "║     CodeMind — 卸载                               ║"
echo "║     版本: v${VERSION}                                 ║"
echo "╚══════════════════════════════════════════════════╝"
echo ""

echo "  安装目录:  $INSTALL_DIR"
echo ""

# 显示将要执行的操作
echo "  即将执行:"
echo "    ✓ 停止所有容器"
echo "    ✓ 移除容器和网络"
if [ "$KEEP_DATA" = true ]; then
    echo "    ○ 保留数据卷（--keep-data）"
else
    echo "    ✓ 移除数据卷（数据库、Redis 数据将永久删除）"
fi
echo "    ✓ 移除 Docker 镜像"
if [ "$KEEP_BACKUPS" = true ]; then
    echo "    ○ 保留备份目录（--keep-backups）"
    echo "    ✓ 移除安装目录（备份除外）"
else
    echo "    ✓ 移除安装目录（含备份）"
fi
echo ""

log_warn "⚠  此操作不可逆！"
echo ""
read -rp "确认卸载？(输入 UNINSTALL 继续) " confirm
if [ "$confirm" != "UNINSTALL" ]; then
    echo "已取消"
    exit 0
fi

echo ""

# ============================================================
# Step 1: 建议先备份
# ============================================================
RUNNING=$(cd "$INSTALL_DIR" && docker compose ps --status running -q 2>/dev/null | wc -l | tr -d ' ')
if [ "$RUNNING" -gt 0 ] && [ "$KEEP_DATA" = false ]; then
    echo ""
    read -rp "是否在卸载前备份数据库？(Y/n) " backup_confirm
    if [[ ! "$backup_confirm" =~ ^[Nn]$ ]]; then
        if [ -f "$INSTALL_DIR/scripts/backup.sh" ]; then
            bash "$INSTALL_DIR/scripts/backup.sh" || log_warn "备份失败，继续卸载..."
        fi
    fi
fi

# ============================================================
# Step 2: 停止并移除容器
# ============================================================
log_step "停止并移除容器..."

cd "$INSTALL_DIR"

if [ "$KEEP_DATA" = true ]; then
    docker compose down --rmi local 2>/dev/null || true
else
    docker compose down --volumes --rmi local 2>/dev/null || true
fi

log_ok "容器和网络已移除"

# ============================================================
# Step 3: 清理 Docker 镜像
# ============================================================
log_step "清理 Docker 镜像..."

docker rmi codemind-frontend codemind-backend 2>/dev/null || true
docker rmi "codemind-frontend:${VERSION}" "codemind-backend:${VERSION}" 2>/dev/null || true
docker image prune -f &>/dev/null || true

log_ok "Docker 镜像已清理"

# ============================================================
# Step 4: 移除安装目录
# ============================================================
log_step "移除安装文件..."

if [ "$KEEP_BACKUPS" = true ]; then
    BACKUP_DIR="$INSTALL_DIR/backups"
    TEMP_BACKUP="/tmp/codemind-backups-$$"

    if [ -d "$BACKUP_DIR" ] && [ "$(ls -A "$BACKUP_DIR" 2>/dev/null)" ]; then
        mv "$BACKUP_DIR" "$TEMP_BACKUP"
        log_info "备份已临时转移"
    fi

    rm -rf "$INSTALL_DIR"

    if [ -d "$TEMP_BACKUP" ]; then
        mkdir -p "$BACKUP_DIR"
        mv "$TEMP_BACKUP"/* "$BACKUP_DIR/" 2>/dev/null || true
        rmdir "$TEMP_BACKUP" 2>/dev/null || true
        log_ok "备份已保留: $BACKUP_DIR"
    fi
else
    rm -rf "$INSTALL_DIR"
fi

log_ok "安装文件已移除"

# ============================================================
# 完成
# ============================================================
echo ""
echo "╔══════════════════════════════════════════════════╗"
echo "║               卸载完成                            ║"
echo "╚══════════════════════════════════════════════════╝"
echo ""

if [ "$KEEP_DATA" = true ]; then
    echo "  数据卷已保留，查看: docker volume ls | grep codemind"
fi

if [ "$KEEP_BACKUPS" = true ] && [ -d "$INSTALL_DIR/backups" ]; then
    echo "  备份已保留: $INSTALL_DIR/backups/"
fi

echo ""
echo "  CodeMind v${VERSION} 已完全卸载"
echo ""
