#!/usr/bin/env bash
# ============================================================
# CodeMind 部署工具 — 共享工具函数
# ============================================================

# ── 颜色定义 ──
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

log_info()  { echo -e "${BLUE}[INFO]${NC}  $1"; }
log_ok()    { echo -e "${GREEN}[ OK ]${NC}  $1"; }
log_warn()  { echo -e "${YELLOW}[WARN]${NC}  $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_step()  { echo -e "${CYAN}[STEP]${NC}  $1"; }

# ── 获取安装目录 ──
get_install_dir() {
    echo "${CODEMIND_HOME:-/opt/codemind}"
}

# ── 检查 Docker 环境 ──
check_docker() {
    if ! command -v docker &>/dev/null; then
        log_error "Docker 未安装"
        log_info "安装指南: https://docs.docker.com/engine/install/ubuntu/"
        return 1
    fi

    if ! docker compose version &>/dev/null; then
        log_error "Docker Compose V2 未安装"
        log_info "Docker Engine 20.10+ 自带 Compose V2 插件"
        return 1
    fi

    if ! docker info &>/dev/null 2>&1; then
        log_error "Docker 服务未运行，或当前用户无权限"
        log_info "尝试: sudo systemctl start docker"
        return 1
    fi

    local docker_version
    docker_version=$(docker version --format '{{.Server.Version}}' 2>/dev/null || echo "unknown")
    log_ok "Docker ${docker_version}"
    return 0
}

# ── 检查端口是否可用 ──
check_port_available() {
    local port=$1
    local name=${2:-""}

    if ss -tlnp 2>/dev/null | grep -q ":${port} " || \
       netstat -tlnp 2>/dev/null | grep -q ":${port} "; then
        if [ -n "$name" ]; then
            log_error "端口 ${port} (${name}) 已被占用"
        else
            log_error "端口 ${port} 已被占用"
        fi
        return 1
    fi
    return 0
}

# ── 生成随机安全密码 ──
generate_password() {
    local length=${1:-24}
    if command -v openssl &>/dev/null; then
        openssl rand -base64 48 | tr -d '/+=' | head -c "$length"
    else
        tr -dc 'A-Za-z0-9' < /dev/urandom | head -c "$length"
    fi
}

# ── 等待 Docker 服务就绪 ──
wait_for_service() {
    local service=$1
    local check_cmd=$2
    local max_wait=${3:-60}
    local elapsed=0

    log_info "等待 ${service} 就绪 (最长 ${max_wait}s)..."

    while [ $elapsed -lt $max_wait ]; do
        if eval "$check_cmd" &>/dev/null; then
            log_ok "${service} 已就绪 (${elapsed}s)"
            return 0
        fi
        sleep 2
        elapsed=$((elapsed + 2))
    done

    log_error "${service} 在 ${max_wait}s 内未就绪"
    return 1
}

# ── 读取 .env 中的配置值 ──
get_env_value() {
    local key=$1
    local env_file=$2
    local default=${3:-""}

    local value
    value=$(grep -E "^${key}=" "$env_file" 2>/dev/null | head -1 | cut -d= -f2-)

    if [ -n "$value" ]; then
        echo "$value"
    else
        echo "$default"
    fi
}

# ── 获取服务器 IP ──
get_server_ip() {
    hostname -I 2>/dev/null | awk '{print $1}' || \
    ip route get 1.1.1.1 2>/dev/null | awk '{print $7}' || \
    echo "your-server-ip"
}

# ── 打印分隔线 ──
print_separator() {
    echo "────────────────────────────────────────────────"
}
