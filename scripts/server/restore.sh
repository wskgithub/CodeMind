#!/usr/bin/env bash
# ============================================================
# CodeMind 数据恢复脚本
# ============================================================
# 从备份文件恢复 PostgreSQL 数据库和配置
#
# 用法:
#   bash scripts/restore.sh <backup_file>                    # 仅恢复数据库
#   bash scripts/restore.sh <backup_file> --with-config      # 同时恢复配置
#   bash scripts/restore.sh --list                           # 列出可用备份
#
# 支持的备份格式:
#   - codemind-backup-*.tar.gz （backup.sh 生成的完整备份）
#   - *.dump                   （pg_dump 格式的数据库备份）
# ============================================================
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/utils.sh"

INSTALL_DIR=$(get_install_dir)
BACKUP_DIR="$INSTALL_DIR/backups"

if [ ! -f "$INSTALL_DIR/docker-compose.yml" ]; then
    log_error "未找到 CodeMind 安装: $INSTALL_DIR"
    exit 1
fi

# ── 解析参数 ──
BACKUP_FILE=""
WITH_CONFIG=false
LIST_ONLY=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --with-config)  WITH_CONFIG=true; shift ;;
        --list|-l)      LIST_ONLY=true; shift ;;
        --help|-h)
            echo "用法: bash $0 <backup_file> [--with-config]"
            echo "      bash $0 --list"
            echo ""
            echo "选项:"
            echo "  --with-config  同时恢复配置文件（.env, app.yaml 等）"
            echo "  --list, -l     列出可用备份"
            exit 0 ;;
        -*) log_error "未知选项: $1"; exit 1 ;;
        *)  BACKUP_FILE="$1"; shift ;;
    esac
done

# ── 列出备份 ──
if [ "$LIST_ONLY" = true ]; then
    echo ""
    echo "── 可用备份 ───────────────────────────────────────"
    if ls "$BACKUP_DIR"/codemind-backup-*.tar.gz &>/dev/null; then
        echo ""
        for f in $(ls -t "$BACKUP_DIR"/codemind-backup-*.tar.gz); do
            SIZE=$(du -sh "$f" | cut -f1)
            DATE=$(stat -c '%y' "$f" 2>/dev/null || stat -f '%Sm' "$f" 2>/dev/null || echo "unknown")
            BASENAME=$(basename "$f")
            echo "  ${BASENAME}  (${SIZE}, ${DATE})"
        done
        echo ""
        echo "恢复示例: bash $0 $BACKUP_DIR/<filename>"
    else
        echo ""
        echo "  （无可用备份）"
    fi
    echo ""
    exit 0
fi

# ── 验证参数 ──
if [ -z "$BACKUP_FILE" ]; then
    log_error "请指定备份文件路径"
    echo ""
    echo "用法: bash $0 <backup_file> [--with-config]"
    echo "      bash $0 --list"
    exit 1
fi

if [ ! -f "$BACKUP_FILE" ]; then
    log_error "备份文件不存在: $BACKUP_FILE"
    exit 1
fi

echo ""
echo "╔══════════════════════════════════════════════════╗"
echo "║     CodeMind — 数据恢复                           ║"
echo "╚══════════════════════════════════════════════════╝"
echo ""

# ── 准备临时目录 ──
TEMP_DIR=$(mktemp -d)
cleanup() { rm -rf "$TEMP_DIR"; }
trap cleanup EXIT

# ── 解析备份文件 ──
DUMP_FILE=""

if [[ "$BACKUP_FILE" == *.tar.gz ]]; then
    log_info "解压备份: $(basename "$BACKUP_FILE")"
    tar -xzf "$BACKUP_FILE" -C "$TEMP_DIR"

    # 查找 database.dump
    DUMP_FILE=$(find "$TEMP_DIR" -name "database.dump" -type f | head -1)
    if [ -z "$DUMP_FILE" ]; then
        log_error "备份中未找到 database.dump"
        exit 1
    fi
    log_ok "找到数据库备份: database.dump"

    # 查找配置文件
    if [ "$WITH_CONFIG" = true ]; then
        BACKUP_CONTENT_DIR=$(dirname "$DUMP_FILE")
        if [ -f "$BACKUP_CONTENT_DIR/env.bak" ]; then
            log_info "找到配置备份: env.bak, app.yaml.bak 等"
        else
            log_warn "备份中未找到配置文件，将仅恢复数据库"
            WITH_CONFIG=false
        fi
    fi

elif [[ "$BACKUP_FILE" == *.dump ]]; then
    DUMP_FILE="$BACKUP_FILE"
    if [ "$WITH_CONFIG" = true ]; then
        log_warn ".dump 格式不包含配置文件，将仅恢复数据库"
        WITH_CONFIG=false
    fi
else
    log_error "不支持的文件格式（需要 .tar.gz 或 .dump）"
    exit 1
fi

# ── 确认恢复 ──
echo ""
log_warn "⚠  恢复操作将覆盖当前数据库中的所有数据！"
if [ "$WITH_CONFIG" = true ]; then
    log_warn "⚠  同时将覆盖当前配置文件！"
fi
echo ""
read -rp "确认恢复？(输入 YES 继续) " confirm
if [ "$confirm" != "YES" ]; then
    echo "已取消"
    exit 0
fi

cd "$INSTALL_DIR"

# ── 恢复前备份当前数据 ──
echo ""
log_info "恢复前自动备份当前数据..."
DB_USER=$(get_env_value "DB_USER" "$INSTALL_DIR/.env" "codemind")
DB_NAME=$(get_env_value "DB_NAME" "$INSTALL_DIR/.env" "codemind")

PRE_RESTORE_DUMP="$BACKUP_DIR/pre-restore-$(date +%Y%m%d_%H%M%S).dump"
mkdir -p "$BACKUP_DIR"

if docker compose ps postgres 2>/dev/null | grep -q "running"; then
    docker compose exec -T postgres \
        pg_dump -U "$DB_USER" -d "$DB_NAME" --format=custom > "$PRE_RESTORE_DUMP" 2>/dev/null || true
    log_ok "当前数据已备份: $(basename "$PRE_RESTORE_DUMP")"
fi

# ── 确保 PostgreSQL 运行 ──
if ! docker compose ps postgres 2>/dev/null | grep -q "running"; then
    log_info "启动 PostgreSQL..."
    docker compose up -d postgres
    wait_for_service "PostgreSQL" \
        "docker compose exec -T postgres pg_isready -U $DB_USER -d $DB_NAME" 60 || {
        log_error "PostgreSQL 启动失败"
        exit 1
    }
fi

# ── 恢复数据库 ──
echo ""
log_step "恢复数据库..."

docker compose exec -T postgres \
    pg_restore -U "$DB_USER" -d "$DB_NAME" \
    --clean --if-exists --no-owner --no-privileges \
    < "$DUMP_FILE" 2>&1 | tail -5 || true

log_ok "数据库恢复完成"

# ── 恢复配置 ──
if [ "$WITH_CONFIG" = true ]; then
    echo ""
    log_step "恢复配置文件..."

    BACKUP_CONTENT_DIR=$(dirname "$DUMP_FILE")

    if [ -f "$BACKUP_CONTENT_DIR/env.bak" ]; then
        cp "$BACKUP_CONTENT_DIR/env.bak" "$INSTALL_DIR/.env"
        chmod 600 "$INSTALL_DIR/.env"
        log_ok ".env 已恢复"
    fi

    if [ -f "$BACKUP_CONTENT_DIR/app.yaml.bak" ]; then
        cp "$BACKUP_CONTENT_DIR/app.yaml.bak" "$INSTALL_DIR/config/app.yaml"
        log_ok "app.yaml 已恢复"
    fi

    if [ -f "$BACKUP_CONTENT_DIR/nginx.conf.bak" ]; then
        cp "$BACKUP_CONTENT_DIR/nginx.conf.bak" "$INSTALL_DIR/docker/nginx/nginx.conf"
        log_ok "nginx.conf 已恢复"
    fi

    log_warn "配置已更新，建议重启服务: bash $INSTALL_DIR/scripts/restart.sh --full"
fi

echo ""
echo "╔══════════════════════════════════════════════════╗"
echo "║               恢复完成                            ║"
echo "╚══════════════════════════════════════════════════╝"
echo ""
echo "  恢复前备份: $PRE_RESTORE_DUMP"
echo ""
echo "  如恢复有误，可回退:"
echo "    bash $0 $PRE_RESTORE_DUMP"
echo ""
