# CodeMind 部署与运维手册

> 本文档适用于将 CodeMind 部署到 Ubuntu 22.04 云服务器的完整流程。  
> 涵盖首次部署、日常运维、版本升级、备份恢复和故障排查。

---

## 目录

- [1. 部署架构](#1-部署架构)
- [2. 环境准备](#2-环境准备)
- [3. 首次部署](#3-首次部署)
- [4. 日常运维](#4-日常运维)
- [5. 版本升级](#5-版本升级)
- [6. 备份与恢复](#6-备份与恢复)
- [7. 配置说明](#7-配置说明)
- [8. 故障排查](#8-故障排查)
- [9. 安全加固](#9-安全加固)
- [10. 完全卸载](#10-完全卸载)

---

## 1. 部署架构

```
                        ┌─────────────────────────────────────────────┐
                        │              Ubuntu 22.04 Server            │
    用户浏览器           │                                             │
   ┌──────────┐         │  ┌──────────┐    ┌───────────┐              │
   │ Browser  │────────►│  │ Frontend │───►│  Backend  │              │
   │          │  :18080 │  │ (Nginx)  │    │  (Go API) │              │
   └──────────┘         │  └──────────┘    └─────┬─────┘              │
                        │                    ┌───┴───┐                │
    IDE / SDK           │               ┌────┴──┐ ┌──┴────┐          │
   ┌──────────┐         │               │ PgSQL │ │ Redis │          │
   │ Cursor   │────────►│               │  :15432│ │ :16379│          │
   │ VS Code  │  :18080 │               └───────┘ └───────┘          │
   │ OpenAI   │ (/api/openai/v1/* 等) │                              │
   └──────────┘         │       LLM 服务（独立部署）                    │
                        │            ┌──────────┐                     │
                        │    Backend─┤  Ollama / │                    │
                        │            │  vLLM 等  │                    │
                        │            └──────────┘                     │
                        └─────────────────────────────────────────────┘
```

**端口规划**（使用非常用端口，避免与其他服务冲突）：

| 服务 | 容器内端口 | 宿主机端口 | 用途 |
|------|-----------|-----------|------|
| Frontend (Nginx) | 80 | **18080** | 浏览器访问 + LLM API 代理 |
| PostgreSQL | 5432 | **15432** | 数据库远程管理（可选） |
| Redis | 6379 | **16379** | 缓存远程管理（可选） |
| Backend | 8080 | 不暴露 | 仅通过 Nginx 内部代理 |

---

## 2. 环境准备

### 2.1 服务器要求

| 项目 | 最低要求 | 推荐配置 |
|------|---------|---------|
| CPU | 2 核 | 4 核+ |
| 内存 | 4 GB | 8 GB+ |
| 硬盘 | 40 GB | 100 GB+ SSD |
| 系统 | Ubuntu 22.04 LTS | Ubuntu 22.04 LTS |
| 网络 | 可达 LLM 服务 | 内网通信 |

### 2.2 安装 Docker（服务器上执行）

```bash
# 安装 Docker Engine
curl -fsSL https://get.docker.com | sh

# 启动 Docker 并设置开机自启
sudo systemctl enable --now docker

# 将当前用户加入 docker 组（可选，免 sudo）
sudo usermod -aG docker $USER

# 验证安装
docker --version
docker compose version
```

> 要求 Docker Engine 20.10+ 及 Docker Compose V2（随 Docker Engine 自带）。

### 2.3 开放防火墙端口

```bash
# UFW 防火墙
sudo ufw allow 18080/tcp   # CodeMind Web 访问

# 如需远程数据库管理（不推荐对公网开放）
# sudo ufw allow 15432/tcp
# sudo ufw allow 16379/tcp
```

### 2.4 LLM 服务

CodeMind 需要一个 LLM 推理服务作为后端（如 Ollama、vLLM、OpenAI 兼容服务）。  
请确保服务器能够访问 LLM 服务地址。

---

## 3. 首次部署

### 3.1 在开发机上打包

```bash
# 在项目根目录执行
bash scripts/package.sh
```

打包脚本自动完成：
1. 构建前端静态文件（`npm run build`）
2. 交叉编译后端二进制（`linux/amd64`）
3. 收集部署配置和脚本
4. 创建 `dist/codemind-v{version}.tar.gz`
5. 生成 SHA256 校验文件

### 3.2 上传到服务器

```bash
# 上传部署包
scp dist/codemind-v0.4.0.tar.gz user@server:/tmp/

# 上传校验文件（可选，用于验证完整性）
scp dist/codemind-v0.4.0.tar.gz.sha256 user@server:/tmp/
```

### 3.3 在服务器上部署

```bash
# 登录服务器
ssh user@server

# 验证文件完整性（可选）
cd /tmp
sha256sum -c codemind-v0.4.0.tar.gz.sha256

# 解压部署包
tar -xzf codemind-v0.4.0.tar.gz

# 执行一键部署
cd codemind-v0.4.0
sudo bash scripts/deploy.sh
```

部署脚本自动完成：
1. **环境检查** — 验证 Docker、检查端口冲突
2. **创建目录** — 安装到 `/opt/codemind`
3. **安装文件** — 复制前后端、配置、脚本
4. **生成配置** — 自动生成 DB/Redis/JWT 安全密码
5. **交互配置** — 提示输入 LLM 服务地址
6. **构建启动** — 构建 Docker 镜像并启动
7. **健康检查** — 等待所有服务就绪
8. **数据库迁移** — 自动应用 SQL 迁移

部署完成后输出：

```
╔══════════════════════════════════════════════════╗
║               部署完成！                          ║
╚══════════════════════════════════════════════════╝

  访问地址:  http://服务器IP:18080
  管理账号:  admin
  初始密码:  Admin@123456
```

> **安全提醒：** 部署后请立即登录修改管理员默认密码！

---

## 4. 日常运维

所有运维脚本位于 `/opt/codemind/scripts/`，需要 `sudo` 执行。

### 4.1 脚本速查表

| 操作 | 命令 |
|------|------|
| 查看状态 | `sudo bash /opt/codemind/scripts/status.sh` |
| 启动服务 | `sudo bash /opt/codemind/scripts/start.sh` |
| 停止服务 | `sudo bash /opt/codemind/scripts/stop.sh` |
| 重启服务 | `sudo bash /opt/codemind/scripts/restart.sh` |
| 重启单个服务 | `sudo bash /opt/codemind/scripts/restart.sh backend` |
| 完全重建 | `sudo bash /opt/codemind/scripts/restart.sh --full` |
| 查看日志 | `sudo bash /opt/codemind/scripts/logs.sh` |
| 实时日志 | `sudo bash /opt/codemind/scripts/logs.sh -f` |
| 后端日志 | `sudo bash /opt/codemind/scripts/logs.sh backend -f` |
| 备份数据 | `sudo bash /opt/codemind/scripts/backup.sh` |
| 恢复数据 | `sudo bash /opt/codemind/scripts/restore.sh <file>` |
| 查看备份 | `sudo bash /opt/codemind/scripts/restore.sh --list` |
| 版本升级 | `sudo bash /opt/codemind/scripts/upgrade.sh` |
| 完全卸载 | `sudo bash /opt/codemind/scripts/uninstall.sh` |

### 4.2 查看服务状态

```bash
sudo bash /opt/codemind/scripts/status.sh
```

输出包含：
- 容器运行状态（名称、状态、端口映射）
- 资源占用（CPU、内存、网络 I/O）
- 磁盘使用（数据库大小、备份大小）
- 连通性检查（PostgreSQL / Redis / 后端 / 前端）

### 4.3 查看和跟踪日志

```bash
# 查看所有服务最近 100 行日志
sudo bash /opt/codemind/scripts/logs.sh

# 实时跟踪后端日志（Ctrl+C 退出）
sudo bash /opt/codemind/scripts/logs.sh backend -f

# 查看数据库最近 200 行日志
sudo bash /opt/codemind/scripts/logs.sh postgres -n 200
```

### 4.4 配置定时备份（推荐）

```bash
# 编辑 crontab
sudo crontab -e

# 添加以下行：每天凌晨 3:00 自动备份
0 3 * * * /bin/bash /opt/codemind/scripts/backup.sh --quiet >> /opt/codemind/logs/backup.log 2>&1
```

备份自动保留最近 30 天，过期备份会被清理。  
修改保留天数：设置环境变量 `BACKUP_RETENTION_DAYS=60`。

### 4.5 修改配置

```bash
# 编辑环境变量
sudo vi /opt/codemind/.env

# 编辑应用配置
sudo vi /opt/codemind/config/app.yaml

# 修改后重启服务生效
sudo bash /opt/codemind/scripts/restart.sh --full
```

---

## 5. 版本升级

### 5.1 在开发机上打包新版本

```bash
# 更新 VERSION 文件
echo "0.4.0" > VERSION

# 打包
bash scripts/package.sh
```

### 5.2 上传并升级

```bash
# 上传
scp dist/codemind-v0.4.0.tar.gz user@server:/tmp/

# 登录服务器，解压
ssh user@server
cd /tmp
tar -xzf codemind-v0.4.0.tar.gz
cd codemind-v0.4.0

# 执行升级（自动备份 + 保留配置）
sudo bash scripts/upgrade.sh
```

升级脚本自动完成：
1. 版本对比确认
2. **自动备份**当前数据库
3. 停止旧版服务
4. 更新文件（**保留 `.env` 和 `app.yaml`**）
5. 重建 Docker 镜像并启动
6. 应用新的数据库迁移
7. 清理旧版镜像

### 5.3 回滚

如果升级后发现问题，使用升级前的自动备份回滚：

```bash
# 查看可用备份
sudo bash /opt/codemind/scripts/restore.sh --list

# 恢复到升级前状态
sudo bash /opt/codemind/scripts/restore.sh /opt/codemind/backups/pre-upgrade-v0.3.0-to-v0.4.0-*.dump
```

---

## 6. 备份与恢复

### 6.1 手动备份

```bash
# 完整备份（数据库 + 配置文件）
sudo bash /opt/codemind/scripts/backup.sh

# 仅备份数据库
sudo bash /opt/codemind/scripts/backup.sh --db-only
```

备份保存在 `/opt/codemind/backups/`，格式：

```
codemind-backup-v0.4.0-20260403_030000.tar.gz
├── database.dump          # PostgreSQL 数据库转储
├── env.bak                # .env 配置
├── app.yaml.bak           # 应用配置
├── nginx.conf.bak         # Nginx 配置
├── docker-compose.yml.bak # Compose 配置
├── migrations_applied.bak # 迁移记录
└── VERSION                # 版本号
```

### 6.2 恢复数据

```bash
# 列出可用备份
sudo bash /opt/codemind/scripts/restore.sh --list

# 恢复数据库（不影响配置）
sudo bash /opt/codemind/scripts/restore.sh /opt/codemind/backups/codemind-backup-v0.4.0-20260403.tar.gz

# 同时恢复数据库和配置
sudo bash /opt/codemind/scripts/restore.sh /opt/codemind/backups/xxx.tar.gz --with-config
```

恢复脚本的安全机制：
- 恢复前自动创建当前数据快照（`pre-restore-*.dump`）
- 需要输入 `YES` 确认才执行
- 支持从快照回退

### 6.3 异地备份

建议将备份定期复制到其他存储：

```bash
# 复制最新备份到本地
scp user@server:/opt/codemind/backups/codemind-backup-*.tar.gz ./backups/

# 或同步到对象存储（以阿里云 OSS 为例）
ossutil cp /opt/codemind/backups/ oss://bucket/codemind-backups/ --recursive
```

---

## 7. 配置说明

### 7.1 环境变量 (.env)

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `FRONTEND_PORT` | 18080 | 前端访问端口 |
| `DB_EXTERNAL_PORT` | 15432 | 数据库外部端口 |
| `REDIS_EXTERNAL_PORT` | 16379 | Redis 外部端口 |
| `DB_NAME` | codemind | 数据库名称 |
| `DB_USER` | codemind | 数据库用户名 |
| `DB_PASSWORD` | (自动生成) | 数据库密码 |
| `REDIS_PASSWORD` | (自动生成) | Redis 密码 |
| `JWT_SECRET` | (自动生成) | JWT 签名密钥 |
| `LLM_BASE_URL` | (部署时配置) | LLM 服务地址 |
| `LLM_API_KEY` | 空 | LLM 服务 API Key |

### 7.2 应用配置 (config/app.yaml)

敏感配置由环境变量覆盖，`app.yaml` 主要控制非敏感参数：

```yaml
system:
  max_keys_per_user: 10          # 每用户最大 API Key 数
  default_concurrency: 5         # 默认并发请求上限
  default_daily_tokens: 1000000  # 默认每日 Token 限额
  default_monthly_tokens: 20000000
  force_change_password: true    # 强制首次修改密码

llm:
  timeout_seconds: 300           # 非流式请求超时
  stream_timeout_seconds: 600    # 流式请求超时

log:
  level: "info"                  # debug | info | warn | error
  format: "json"                 # json | console
```

### 7.3 安装目录结构

```
/opt/codemind/
├── frontend/           # 前端 Dockerfile + dist 静态文件
├── backend/            # 后端 Dockerfile + 二进制文件
├── config/
│   └── app.yaml        # 应用配置
├── docker/
│   ├── nginx/
│   │   └── nginx.conf  # Nginx 配置
│   └── postgres/
│       ├── init.sql    # 数据库建表
│       └── seed.sql    # 初始数据
├── migrations/         # 数据库迁移 SQL
├── scripts/            # 运维脚本
├── backups/            # 备份文件
├── logs/               # 日志目录
├── docker-compose.yml  # 容器编排
├── .env                # 环境变量（权限 600）
└── VERSION             # 当前版本号
```

---

## 8. 故障排查

### 8.1 常用诊断命令

```bash
# 查看所有容器状态
cd /opt/codemind && docker compose ps

# 查看特定服务日志
docker compose logs --tail 200 backend
docker compose logs --tail 200 postgres

# 进入容器排查
docker compose exec backend sh
docker compose exec postgres psql -U codemind -d codemind

# 查看容器资源占用
docker stats --no-stream

# 检查端口监听
ss -tlnp | grep -E '18080|15432|16379'
```

### 8.2 常见问题

#### 前端访问报 502 Bad Gateway

**原因：** 后端服务未就绪或已崩溃。

```bash
# 检查后端状态
docker compose logs backend --tail 50

# 重启后端
sudo bash /opt/codemind/scripts/restart.sh backend
```

#### 数据库连接失败

**原因：** PostgreSQL 未启动或密码不匹配。

```bash
# 检查数据库状态
docker compose exec postgres pg_isready -U codemind

# 查看数据库日志
docker compose logs postgres --tail 50

# 验证密码是否正确
docker compose exec postgres psql -U codemind -d codemind -c "SELECT 1;"
```

#### LLM 请求超时

**原因：** LLM 服务不可达或响应过慢。

```bash
# 从后端容器测试 LLM 连通性
LLM_URL=$(grep LLM_BASE_URL /opt/codemind/.env | cut -d= -f2-)
docker compose exec backend wget -qO- --timeout=5 "${LLM_URL}/v1/models" || echo "LLM 不可达"
```

#### 磁盘空间不足

```bash
# 查看磁盘使用
df -h

# 清理 Docker 缓存
docker system prune -af

# 清理旧备份（保留最近 7 天）
find /opt/codemind/backups -name "*.tar.gz" -mtime +7 -delete
```

#### 端口被占用

```bash
# 查看端口占用
ss -tlnp | grep 18080

# 修改端口
sudo vi /opt/codemind/.env
# 修改 FRONTEND_PORT=新端口号

# 重启服务
sudo bash /opt/codemind/scripts/restart.sh --full
```

---

## 9. 安全加固

### 9.1 基础安全措施

部署脚本已自动完成的安全配置：
- ✅ DB/Redis/JWT 密码自动生成（24-48 位随机字符串）
- ✅ `.env` 文件权限限制为 600
- ✅ 后端容器使用非 root 用户运行
- ✅ Nginx 添加安全响应头（X-Frame-Options, CSP 等）
- ✅ 后端 API 不直接暴露端口

### 9.2 建议额外配置

```bash
# 1. 关闭不需要的外部端口（数据库、Redis 端口一般不需要外部访问）
#    编辑 docker-compose.yml，注释掉 postgres 和 redis 的 ports 映射

# 2. 配置 HTTPS（推荐使用外部负载均衡器或 Nginx 反向代理加证书）

# 3. 限制 SSH 访问
sudo ufw default deny incoming
sudo ufw allow ssh
sudo ufw allow 18080/tcp
sudo ufw enable

# 4. 定期更新
sudo apt update && sudo apt upgrade -y
```

### 9.3 默认账号安全

| 项目 | 值 | 操作 |
|------|------|------|
| 管理员账号 | admin | 部署后立即修改密码 |
| 初始密码 | Admin@123456 | **必须修改** |

---

## 10. 完全卸载

```bash
# 交互式卸载（会提示确认，可选择先备份）
sudo bash /opt/codemind/scripts/uninstall.sh

# 保留数据卷卸载（可重新部署后恢复）
sudo bash /opt/codemind/scripts/uninstall.sh --keep-data

# 保留备份文件
sudo bash /opt/codemind/scripts/uninstall.sh --keep-backups
```

卸载流程：
1. 询问是否在卸载前备份数据
2. 停止并移除所有容器
3. 清理 Docker 镜像
4. 移除数据卷（除非 `--keep-data`）
5. 移除安装目录（备份可选保留）

---

## 附录：完整部署流程速查

```bash
# ──── 开发机（macOS）──────────────────────
bash scripts/package.sh                                  # 打包
scp dist/codemind-v0.4.0.tar.gz user@server:/tmp/       # 上传

# ──── 服务器（Ubuntu 22.04）────────────────
cd /tmp && tar -xzf codemind-v0.4.0.tar.gz             # 解压
cd codemind-v0.4.0 && sudo bash scripts/deploy.sh      # 部署

# ──── 日常运维 ─────────────────────────────
sudo bash /opt/codemind/scripts/status.sh               # 查看状态
sudo bash /opt/codemind/scripts/logs.sh backend -f      # 查看日志
sudo bash /opt/codemind/scripts/backup.sh               # 手动备份
sudo bash /opt/codemind/scripts/restart.sh              # 重启服务

# ──── 版本升级 ─────────────────────────────
# 开发机打包新版本后上传到 /tmp
cd /tmp/codemind-v0.4.0
sudo bash scripts/upgrade.sh                            # 自动升级
```
