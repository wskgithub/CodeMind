#!/usr/bin/env bash
# ============================================================
# CodeMind 版本升级脚本
# ============================================================
# 在服务器上运行，将已部署的 CodeMind 升级到新版本
# 升级前自动备份数据库，确保数据安全
#
# 用法: sudo bash scripts/upgrade.sh [--install-dir /opt/codemind]
#
# 升级流程:
#   1. 版本检查
#   2. 升级前自动备份
#   3. 停止服务
#   4. 更新文件（保留 .env 和 app.yaml）
#   5. 重建镜像并启动
#   6. 应用新数据库迁移 + 健康检查
# ============================================================
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PACKAGE_DIR="$(dirname "$SCRIPT_DIR")"
source "$SCRIPT_DIR/utils.sh"

NEW_VERSION=$(tr -d '[:space:]' < "$PACKAGE_DIR/VERSION")

# ── 解析参数 ──
INSTALL_DIR="${CODEMIND_HOME:-/opt/codemind}"
SKIP_BACKUP=false
while [[ $# -gt 0 ]]; do
    case $1 in
        --install-dir) INSTALL_DIR="$2"; shift 2 ;;
        --skip-backup) SKIP_BACKUP=true; shift ;;
        --help|-h)
            echo "用法: sudo bash $0 [--install-dir /opt/codemind] [--skip-backup]"
            echo ""
            echo "选项:"
            echo "  --install-dir  指定安装目录（默认: /opt/codemind）"
            echo "  --skip-backup  跳过升级前备份（不推荐）"
            exit 0 ;;
        *) log_error "未知参数: $1"; exit 1 ;;
    esac
done

echo ""
echo "╔══════════════════════════════════════════════════╗"
echo "║     CodeMind — 版本升级                           ║"
echo "╚══════════════════════════════════════════════════╝"
echo ""

# ── 权限检查 ──
if [ "$EUID" -ne 0 ]; then
    log_error "请使用 sudo 运行此脚本"
    exit 1
fi

# ── 检查现有安装 ──
if [ ! -f "$INSTALL_DIR/VERSION" ]; then
    log_error "未找到 CodeMind 安装: $INSTALL_DIR"
    log_info "请先使用 deploy.sh 进行首次部署"
    exit 1
fi

if [ ! -f "$INSTALL_DIR/.env" ]; then
    log_error "未找到配置文件: $INSTALL_DIR/.env"
    exit 1
fi

CURRENT_VERSION=$(tr -d '[:space:]' < "$INSTALL_DIR/VERSION")

log_info "当前版本: v${CURRENT_VERSION}"
log_info "目标版本: v${NEW_VERSION}"
echo ""

if [ "$CURRENT_VERSION" = "$NEW_VERSION" ]; then
    log_warn "当前已是 v${NEW_VERSION}"
    read -rp "是否强制重新部署？(y/N) " confirm
    [[ ! "$confirm" =~ ^[Yy]$ ]] && exit 0
    echo ""
fi

read -rp "确认升级 v${CURRENT_VERSION} → v${NEW_VERSION}？(Y/n) " confirm
[[ "$confirm" =~ ^[Nn]$ ]] && { echo "已取消"; exit 0; }
echo ""

# ============================================================
# Step 1: 升级前备份
# ============================================================
log_step "━━━ Step 1/5: 升级前备份 ━━━"

if [ "$SKIP_BACKUP" = true ]; then
    log_warn "已跳过备份（--skip-backup）"
else
    if [ -f "$INSTALL_DIR/scripts/backup.sh" ]; then
        bash "$INSTALL_DIR/scripts/backup.sh" || {
            log_error "备份失败，升级已中止"
            log_info "使用 --skip-backup 可跳过备份（不推荐）"
            exit 1
        }
    else
        log_warn "未找到备份脚本，执行简易数据库备份..."
        BACKUP_DIR="$INSTALL_DIR/backups"
        mkdir -p "$BACKUP_DIR"
        cd "$INSTALL_DIR"

        DB_USER=$(get_env_value "DB_USER" "$INSTALL_DIR/.env" "codemind")
        DB_NAME=$(get_env_value "DB_NAME" "$INSTALL_DIR/.env" "codemind")
        DUMP_FILE="$BACKUP_DIR/pre-upgrade-v${CURRENT_VERSION}-to-v${NEW_VERSION}-$(date +%Y%m%d_%H%M%S).dump"

        docker compose exec -T postgres \
            pg_dump -U "$DB_USER" -d "$DB_NAME" --format=custom > "$DUMP_FILE"
        log_ok "数据库备份: $DUMP_FILE"
    fi
fi

# ============================================================
# Step 2: 停止服务
# ============================================================
echo ""
log_step "━━━ Step 2/5: 停止服务 ━━━"

cd "$INSTALL_DIR"
docker compose down
log_ok "服务已停止"

# ============================================================
# Step 3: 更新文件
# ============================================================
echo ""
log_step "━━━ Step 3/5: 更新文件 ━━━"

# 更新前端
cp -r "$PACKAGE_DIR/frontend" "$INSTALL_DIR/"
log_info "  前端已更新"

# 更新后端
cp -r "$PACKAGE_DIR/backend" "$INSTALL_DIR/"
log_info "  后端已更新"

# 更新 docker-compose
cp "$PACKAGE_DIR/docker-compose.yml" "$INSTALL_DIR/"
log_info "  docker-compose.yml 已更新"

# 更新 Nginx 配置
cp "$PACKAGE_DIR/docker/nginx/nginx.conf" "$INSTALL_DIR/docker/nginx/"
log_info "  Nginx 配置已更新"

# 更新数据库初始化脚本（不影响已有数据）
cp "$PACKAGE_DIR/docker/postgres/"*.sql "$INSTALL_DIR/docker/postgres/"

# 更新迁移文件
if ls "$PACKAGE_DIR/migrations/"*.sql &>/dev/null; then
    cp "$PACKAGE_DIR/migrations/"*.sql "$INSTALL_DIR/migrations/"
    log_info "  迁移文件已更新"
fi

# 更新脚本
cp "$PACKAGE_DIR/scripts/"*.sh "$INSTALL_DIR/scripts/"
chmod +x "$INSTALL_DIR/scripts/"*.sh
log_info "  管理脚本已更新"

# 更新版本号
cp "$PACKAGE_DIR/VERSION" "$INSTALL_DIR/"

# 更新 .env 中的版本号（保留其他所有配置）
sed -i "s|^APP_VERSION=.*|APP_VERSION=${NEW_VERSION}|" "$INSTALL_DIR/.env"

log_ok "文件更新完成"
log_ok ".env 和 app.yaml 已保留（未修改用户配置）"

# ============================================================
# Step 4: 重建镜像并启动
# ============================================================
echo ""
log_step "━━━ Step 4/5: 重建镜像并启动 ━━━"

cd "$INSTALL_DIR"

log_info "构建新版本 Docker 镜像..."
docker compose build --no-cache 2>&1 | while IFS= read -r line; do
    echo "  $line"
done

log_info "启动服务..."
docker compose up -d

log_ok "服务启动指令已发送"

# ============================================================
# Step 5: 迁移 + 健康检查
# ============================================================
echo ""
log_step "━━━ Step 5/5: 迁移与健康检查 ━━━"

cd "$INSTALL_DIR"

# 从配置读取数据库连接信息
DB_USER=$(get_env_value "DB_USER" "$INSTALL_DIR/.env" "codemind")
DB_NAME=$(get_env_value "DB_NAME" "$INSTALL_DIR/.env" "codemind")

# 等待 PostgreSQL
wait_for_service "PostgreSQL" \
    "docker compose exec -T postgres pg_isready -U $DB_USER -d $DB_NAME" 60 || {
    log_error "PostgreSQL 启动失败"
    log_warn "回滚方法: 从备份恢复后重启旧版本"
    exit 1
}

# 应用新的数据库迁移
MIGRATIONS_APPLIED="$INSTALL_DIR/migrations/.applied"
touch "$MIGRATIONS_APPLIED"

if ls "$INSTALL_DIR/migrations/"*.sql &>/dev/null; then
    log_info "检查数据库迁移..."
    MIGRATION_COUNT=0

    for sql_file in $(ls "$INSTALL_DIR/migrations/"*.sql | sort); do
        filename=$(basename "$sql_file")
        if ! grep -qF "$filename" "$MIGRATIONS_APPLIED"; then
            log_info "  应用新迁移: $filename"
            if docker compose exec -T postgres \
                psql -U "$DB_USER" -d "$DB_NAME" < "$sql_file" 2>&1; then
                echo "$filename" >> "$MIGRATIONS_APPLIED"
                MIGRATION_COUNT=$((MIGRATION_COUNT + 1))
            else
                log_warn "  $filename 执行出现警告（已标记为完成）"
                echo "$filename" >> "$MIGRATIONS_APPLIED"
            fi
        fi
    done

    if [ $MIGRATION_COUNT -gt 0 ]; then
        log_ok "已应用 $MIGRATION_COUNT 个新迁移"
    else
        log_ok "无需执行新迁移"
    fi
fi

# 等待后端
wait_for_service "后端服务" \
    "docker compose exec -T backend wget -qO- http://localhost:8080/health" 60 || {
    log_error "后端启动失败，查看日志: docker compose logs backend"
    exit 1
}

# 等待前端
wait_for_service "前端服务" \
    "docker compose exec -T frontend wget -qO- http://localhost:80/nginx-health" 30 || {
    log_warn "前端健康检查未响应，请手动验证"
}

# 清理旧镜像
log_info "清理旧版本镜像..."
docker image prune -f &>/dev/null || true

# ============================================================
# 升级完成
# ============================================================
SERVER_IP=$(get_server_ip)
FRONTEND_PORT=$(get_env_value "FRONTEND_PORT" "$INSTALL_DIR/.env" "18080")

echo ""
echo "╔══════════════════════════════════════════════════╗"
echo "║               升级完成！                          ║"
echo "╚══════════════════════════════════════════════════╝"
echo ""
echo "  版本:      v${CURRENT_VERSION} → v${NEW_VERSION}"
echo "  访问地址:  http://${SERVER_IP}:${FRONTEND_PORT}"
echo ""
echo "  查看状态:  cd $INSTALL_DIR && docker compose ps"
echo "  查看日志:  cd $INSTALL_DIR && docker compose logs -f"
echo ""
echo "  如需回滚，请使用备份恢复:"
echo "    1. docker compose down"
echo "    2. 解压备份中的 database.dump"
echo "    3. docker compose up -d postgres"
echo "    4. docker compose exec -T postgres pg_restore \\"
echo "         -U codemind -d codemind --clean --if-exists < database.dump"
echo "    5. docker compose up -d"
echo ""
