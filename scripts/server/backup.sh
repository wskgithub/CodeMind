#!/usr/bin/env bash
# ============================================================
# CodeMind 数据备份脚本
# ============================================================
# 备份 PostgreSQL 数据库和配置文件
# 支持自动清理过期备份
#
# 用法:
#   sudo bash backup.sh              # 完整备份
#   sudo bash backup.sh --db-only    # 仅备份数据库
#   sudo bash backup.sh --quiet      # 静默模式（用于 cron）
# ============================================================
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/utils.sh"

INSTALL_DIR=$(get_install_dir)
BACKUP_DIR="$INSTALL_DIR/backups"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
RETENTION_DAYS=${BACKUP_RETENTION_DAYS:-30}

# ── 解析参数 ──
DB_ONLY=false
QUIET=false
while [[ $# -gt 0 ]]; do
    case $1 in
        --db-only)  DB_ONLY=true; shift ;;
        --quiet|-q) QUIET=true; shift ;;
        --help|-h)
            echo "用法: sudo bash $0 [--db-only] [--quiet]"
            echo ""
            echo "选项:"
            echo "  --db-only   仅备份数据库"
            echo "  --quiet     静默模式（适用于 cron 定时任务）"
            exit 0 ;;
        *) log_error "未知参数: $1"; exit 1 ;;
    esac
done

# 静默模式下重定向非错误输出
if [ "$QUIET" = true ]; then
    log_info()  { :; }
    log_ok()    { :; }
    log_step()  { :; }
fi

# ── 检查安装 ──
if [ ! -f "$INSTALL_DIR/docker-compose.yml" ]; then
    log_error "未找到 CodeMind 安装: $INSTALL_DIR"
    log_info "设置 CODEMIND_HOME 环境变量指定安装目录"
    exit 1
fi

VERSION=$(tr -d '[:space:]' < "$INSTALL_DIR/VERSION")
BACKUP_NAME="codemind-backup-v${VERSION}-${TIMESTAMP}"
BACKUP_PATH="$BACKUP_DIR/$BACKUP_NAME"

if [ "$QUIET" = false ]; then
    echo ""
    echo "═══════════════════════════════════════════"
    echo "  CodeMind 数据备份  v${VERSION}"
    echo "═══════════════════════════════════════════"
    echo ""
fi

cd "$INSTALL_DIR"

# ── 检查 PostgreSQL ──
if ! docker compose ps postgres 2>/dev/null | grep -q "running"; then
    log_error "PostgreSQL 未运行，无法备份"
    exit 1
fi

mkdir -p "$BACKUP_PATH"

# ============================================================
# 1. 备份数据库
# ============================================================
log_info "备份 PostgreSQL 数据库..."

DB_USER=$(get_env_value "DB_USER" "$INSTALL_DIR/.env" "codemind")
DB_NAME=$(get_env_value "DB_NAME" "$INSTALL_DIR/.env" "codemind")

docker compose exec -T postgres \
    pg_dump -U "$DB_USER" -d "$DB_NAME" \
    --format=custom --compress=6 \
    > "$BACKUP_PATH/database.dump"

DB_SIZE=$(du -sh "$BACKUP_PATH/database.dump" | cut -f1)
log_ok "数据库备份完成 ($DB_SIZE)"

# ============================================================
# 2. 备份配置文件
# ============================================================
if [ "$DB_ONLY" = false ]; then
    log_info "备份配置文件..."

    cp "$INSTALL_DIR/.env" "$BACKUP_PATH/env.bak"
    cp "$INSTALL_DIR/config/app.yaml" "$BACKUP_PATH/app.yaml.bak"
    cp "$INSTALL_DIR/docker/nginx/nginx.conf" "$BACKUP_PATH/nginx.conf.bak"
    cp "$INSTALL_DIR/docker-compose.yml" "$BACKUP_PATH/docker-compose.yml.bak"
    cp "$INSTALL_DIR/VERSION" "$BACKUP_PATH/VERSION"

    # 备份迁移记录
    if [ -f "$INSTALL_DIR/migrations/.applied" ]; then
        cp "$INSTALL_DIR/migrations/.applied" "$BACKUP_PATH/migrations_applied.bak"
    fi

    log_ok "配置备份完成"
fi

# ============================================================
# 3. 压缩
# ============================================================
log_info "压缩备份文件..."

cd "$BACKUP_DIR"
tar -czf "${BACKUP_NAME}.tar.gz" "$BACKUP_NAME"
rm -rf "$BACKUP_PATH"

BACKUP_SIZE=$(du -sh "$BACKUP_DIR/${BACKUP_NAME}.tar.gz" | cut -f1)
log_ok "压缩完成: ${BACKUP_NAME}.tar.gz ($BACKUP_SIZE)"

# ============================================================
# 4. 清理过期备份
# ============================================================
log_info "清理 ${RETENTION_DAYS} 天前的备份..."

DELETED_COUNT=0
while IFS= read -r old_backup; do
    rm -f "$old_backup"
    DELETED_COUNT=$((DELETED_COUNT + 1))
done < <(find "$BACKUP_DIR" -name "codemind-backup-*.tar.gz" -mtime +${RETENTION_DAYS} 2>/dev/null || true)

REMAINING_COUNT=$(find "$BACKUP_DIR" -name "codemind-backup-*.tar.gz" 2>/dev/null | wc -l)

if [ $DELETED_COUNT -gt 0 ]; then
    log_info "已清理 $DELETED_COUNT 个过期备份"
fi
log_ok "当前保留 $REMAINING_COUNT 个备份"

# ============================================================
# 完成
# ============================================================
if [ "$QUIET" = false ]; then
    echo ""
    echo "═══════════════════════════════════════════"
    echo "  备份完成"
    echo "═══════════════════════════════════════════"
    echo ""
    echo "  文件:      $BACKUP_DIR/${BACKUP_NAME}.tar.gz"
    echo "  大小:      $BACKUP_SIZE"
    echo "  保留策略:  ${RETENTION_DAYS} 天"
    echo ""
    echo "  恢复数据库:"
    echo "    docker compose exec -T postgres pg_restore \\"
    echo "      -U codemind -d codemind --clean --if-exists \\"
    echo "      < database.dump"
    echo ""
else
    # cron 模式仅输出结果摘要
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] 备份完成: ${BACKUP_NAME}.tar.gz ($BACKUP_SIZE)"
fi
