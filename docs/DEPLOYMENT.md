# CodeMind — 云服务器部署手册

> **目标平台**: Ubuntu 22.04 LTS (x86_64)
> **部署方式**: Docker 容器化部署
> **适用版本**: v0.3.0+

---

## 目录

1. [部署架构概览](#1-部署架构概览)
2. [服务器环境准备](#2-服务器环境准备)
3. [本地打包](#3-本地打包)
4. [首次部署](#4-首次部署)
5. [部署验证](#5-部署验证)
6. [配置说明](#6-配置说明)
7. [日常运维](#7-日常运维)
8. [版本升级](#8-版本升级)
9. [数据备份与恢复](#9-数据备份与恢复)
10. [故障排查](#10-故障排查)
11. [安全加固](#11-安全加固)
12. [附录](#附录)

---

## 1. 部署架构概览

### 1.1 架构图

```
┌─────────────────────────────────────────────────────────┐
│                    Ubuntu 22.04 Server                   │
│                                                         │
│  ┌─────────────────── Docker Network ─────────────────┐ │
│  │                                                     │ │
│  │  ┌──────────┐   ┌──────────┐   ┌──────────────┐   │ │
│  │  │ Frontend │──▶│ Backend  │──▶│  PostgreSQL   │   │ │
│  │  │ (Nginx)  │   │  (Go)    │   │    :5432      │   │ │
│  │  │  :80     │   │  :8080   │   └──────────────┘   │ │
│  │  └──────────┘   └────┬─────┘   ┌──────────────┐   │ │
│  │                      └────────▶│    Redis      │   │ │
│  │                                │    :6379      │   │ │
│  │                                └──────────────┘   │ │
│  └─────────────────────────────────────────────────────┘ │
│                         │                                │
│              :18080 (对外端口)                             │
└─────────────────────────────────────────────────────────┘
          ▲
          │  http://服务器IP:18080
     浏览器访问
```

### 1.2 端口规划

所有端口使用非常用端口，避免与服务器上其他项目冲突：

| 服务 | 容器内端口 | 对外映射端口 | 用途 |
|------|-----------|-------------|------|
| Frontend (Nginx) | 80 | **18080** | 浏览器访问入口 |
| PostgreSQL | 5432 | 15432 | 数据库远程管理（可选） |
| Redis | 6379 | 16379 | Redis 远程访问（可选） |
| Backend | 8080 | 不对外暴露 | 通过 Nginx 反向代理访问 |

> 端口可在 `.env` 文件中自由修改。

### 1.3 数据持久化

| Docker Volume | 用途 | 说明 |
|---------------|------|------|
| `codemind_postgres_data` | PostgreSQL 数据 | 所有业务数据 |
| `codemind_redis_data` | Redis 持久化数据 | 缓存与会话 |
| `codemind_logs` | 应用日志 | 后端运行日志 |

---

## 2. 服务器环境准备

### 2.1 最低配置要求

| 项目 | 最低要求 | 推荐配置 |
|------|---------|---------|
| CPU | 2 核 | 4 核+ |
| 内存 | 4 GB | 8 GB+ |
| 磁盘 | 20 GB | 50 GB+ (SSD) |
| 系统 | Ubuntu 22.04 LTS | Ubuntu 22.04 LTS |
| 架构 | x86_64 (amd64) | x86_64 (amd64) |

### 2.2 安装 Docker

如果服务器尚未安装 Docker，执行以下命令：

```bash
# 更新包索引
sudo apt-get update

# 安装依赖
sudo apt-get install -y ca-certificates curl gnupg

# 添加 Docker GPG 密钥
sudo install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
sudo chmod a+r /etc/apt/keyrings/docker.gpg

# 添加 Docker 软件源
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
  $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
  sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

# 安装 Docker Engine
sudo apt-get update
sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

# 启动并设置开机自启
sudo systemctl start docker
sudo systemctl enable docker

# 验证安装
docker --version
docker compose version
```

### 2.3 配置 Docker 用户权限（可选）

```bash
# 将当前用户加入 docker 组（免 sudo）
sudo usermod -aG docker $USER

# 需要重新登录生效
```

### 2.4 开放防火墙端口

```bash
# 如果使用 ufw 防火墙
sudo ufw allow 18080/tcp comment 'CodeMind Frontend'

# 如果需要远程连接数据库（可选）
sudo ufw allow 15432/tcp comment 'CodeMind PostgreSQL'
```

> 如果服务器使用云服务商的安全组，需要在安全组规则中放行 **18080** 端口。

---

## 3. 本地打包

在 **开发机器** 上执行，将项目编译并打包为可部署的压缩包。

### 3.1 打包环境要求

| 工具 | 版本要求 |
|------|---------|
| Go | 1.23+ |
| Node.js | 20+ |
| npm | 10+ |

### 3.2 执行打包

```bash
# 在项目根目录执行
bash scripts/package.sh
```

打包脚本会自动完成：
1. 安装前端依赖并构建生产版本
2. 交叉编译后端为 **linux/amd64** 平台二进制
3. 收集配置模板、Dockerfile、数据库脚本、部署脚本
4. 生成压缩包和 SHA256 校验文件

### 3.3 打包产物

```
dist/
├── codemind-v0.3.0.tar.gz          # 部署包
└── codemind-v0.3.0.tar.gz.sha256   # 校验文件
```

### 3.4 压缩包内容

```
codemind-v0.3.0/
├── frontend/                # 前端 (Dockerfile + 静态文件)
│   ├── Dockerfile
│   └── dist/
├── backend/                 # 后端 (Dockerfile + 二进制)
│   ├── Dockerfile
│   └── codemind
├── docker-compose.yml       # 生产环境编排文件
├── .env.template            # 环境变量模板
├── config/
│   └── app.yaml.template   # 应用配置模板
├── docker/
│   ├── nginx/
│   │   └── nginx.conf      # Nginx 配置
│   └── postgres/
│       ├── init.sql         # 数据库初始化
│       └── seed.sql         # 种子数据
├── migrations/              # 数据库迁移脚本
├── scripts/
│   ├── deploy.sh            # 一键部署脚本
│   ├── upgrade.sh           # 版本升级脚本
│   ├── backup.sh            # 数据备份脚本
│   └── utils.sh             # 工具函数
└── VERSION                  # 版本号
```

---

## 4. 首次部署

### 4.1 上传压缩包到服务器

```bash
# 从本地上传（在开发机器上执行）
scp dist/codemind-v0.3.0.tar.gz user@your-server-ip:/tmp/
```

### 4.2 在服务器上解压

```bash
# SSH 登录服务器后执行
cd /tmp
tar -xzf codemind-v0.3.0.tar.gz
cd codemind-v0.3.0
```

### 4.3 执行一键部署

```bash
sudo bash scripts/deploy.sh
```

部署脚本会依次执行：

1. **环境检查** — 验证 Docker 是否就绪，端口是否可用
2. **创建目录** — 在 `/opt/codemind` 创建安装目录
3. **安装文件** — 复制所有必要文件到安装目录
4. **生成配置** — 自动生成安全随机密码，交互式配置 LLM 地址
5. **构建启动** — 构建 Docker 镜像并启动所有服务
6. **健康检查** — 验证所有服务正常运行

#### 部署过程中的交互

脚本会在 **Step 4** 时询问 LLM 服务配置：

```
请输入 LLM 服务地址 (例: http://192.168.1.100:11434): http://your-llm-server:11434
LLM API Key (无则直接回车跳过):
```

> 如果暂时没有 LLM 服务地址，可以直接回车跳过，后续在 `.env` 文件中配置。

### 4.4 自定义安装目录（可选）

默认安装到 `/opt/codemind`，可通过参数或环境变量修改：

```bash
# 方式一：命令参数
sudo bash scripts/deploy.sh --install-dir /data/codemind

# 方式二：环境变量
export CODEMIND_HOME=/data/codemind
sudo -E bash scripts/deploy.sh
```

### 4.5 部署完成

部署成功后会显示：

```
╔══════════════════════════════════════════════════╗
║               部署完成！                          ║
╚══════════════════════════════════════════════════╝

  访问地址:  http://your-server-ip:18080
  管理账号:  admin
  初始密码:  Admin@123456
```

---

## 5. 部署验证

### 5.1 检查容器状态

```bash
cd /opt/codemind
docker compose ps
```

正常输出应为：

```
NAME                IMAGE                          STATUS                  PORTS
codemind-backend    codemind-backend:0.3.0        Up (healthy)
codemind-frontend   codemind-frontend:0.3.0       Up           0.0.0.0:18080->80/tcp
codemind-postgres   postgres:16-alpine             Up (healthy) 0.0.0.0:15432->5432/tcp
codemind-redis      redis:7-alpine                 Up (healthy) 0.0.0.0:16379->6379/tcp
```

所有服务状态应为 `Up`，且 `postgres` 和 `backend` 显示 `(healthy)`。

### 5.2 浏览器访问

打开浏览器，访问：

```
http://服务器IP:18080
```

应该能看到 CodeMind 登录页面。

### 5.3 登录验证

使用默认管理员账号登录：

- **用户名**: `admin`
- **密码**: `Admin@123456`

> 首次登录后系统会强制要求修改密码。

### 5.4 API 健康检查

```bash
# 在服务器上执行
curl http://localhost:18080/api/health
```

应返回健康状态响应。

### 5.5 查看服务日志

```bash
cd /opt/codemind

# 查看所有服务日志
docker compose logs -f

# 查看特定服务日志
docker compose logs -f backend
docker compose logs -f postgres
```

---

## 6. 配置说明

### 6.1 配置文件位置

| 文件 | 路径 | 说明 |
|------|------|------|
| 环境变量 | `/opt/codemind/.env` | 密码、密钥、端口等 |
| 应用配置 | `/opt/codemind/config/app.yaml` | 应用运行参数 |
| Nginx 配置 | `/opt/codemind/docker/nginx/nginx.conf` | 前端反向代理 |
| Docker 编排 | `/opt/codemind/docker-compose.yml` | 容器编排（一般无需修改） |

### 6.2 环境变量 (.env)

`.env` 是最核心的配置文件，包含所有敏感信息。部署脚本已自动生成安全随机密码。

```bash
# 查看当前配置
sudo cat /opt/codemind/.env

# 编辑配置
sudo nano /opt/codemind/.env
```

**关键配置项**：

| 变量 | 说明 | 修改后操作 |
|------|------|-----------|
| `FRONTEND_PORT` | 前端访问端口 | 重启全部服务 |
| `DB_PASSWORD` | 数据库密码 | ⚠️ 已运行后勿修改 |
| `REDIS_PASSWORD` | Redis 密码 | ⚠️ 已运行后勿修改 |
| `JWT_SECRET` | JWT 签名密钥 | ⚠️ 修改后所有用户需重新登录 |
| `LLM_BASE_URL` | LLM 服务地址 | 重启后端服务 |
| `LLM_API_KEY` | LLM API 密钥 | 重启后端服务 |

### 6.3 修改端口

如果需要修改访问端口：

```bash
# 1. 编辑 .env
sudo nano /opt/codemind/.env
# 修改 FRONTEND_PORT=新端口号

# 2. 重启服务
cd /opt/codemind
sudo docker compose down
sudo docker compose up -d
```

### 6.4 修改 LLM 服务地址

```bash
# 1. 编辑 .env
sudo nano /opt/codemind/.env
# 修改 LLM_BASE_URL=http://新地址:端口

# 2. 仅重启后端
cd /opt/codemind
sudo docker compose restart backend
```

---

## 7. 日常运维

### 7.1 服务管理命令

```bash
cd /opt/codemind

# 查看服务状态
docker compose ps

# 启动所有服务
docker compose up -d

# 停止所有服务
docker compose down

# 重启所有服务
docker compose restart

# 重启单个服务
docker compose restart backend
docker compose restart frontend

# 查看实时日志
docker compose logs -f

# 查看指定服务最近 100 行日志
docker compose logs --tail 100 backend
```

### 7.2 磁盘空间管理

```bash
# 查看 Docker 磁盘占用
docker system df

# 清理未使用的镜像和缓存
docker system prune -f

# 清理所有未使用的资源（谨慎使用）
docker system prune -a -f
```

### 7.3 数据库管理

```bash
cd /opt/codemind

# 进入 PostgreSQL 命令行
docker compose exec postgres psql -U codemind -d codemind

# 查看数据库大小
docker compose exec postgres psql -U codemind -d codemind \
    -c "SELECT pg_size_pretty(pg_database_size('codemind'));"

# 查看各表大小
docker compose exec postgres psql -U codemind -d codemind \
    -c "SELECT tablename, pg_size_pretty(pg_total_relation_size(tablename::text)) 
        FROM pg_tables WHERE schemaname='public' ORDER BY pg_total_relation_size(tablename::text) DESC;"
```

### 7.4 配置定时备份（推荐）

```bash
# 编辑 crontab
sudo crontab -e

# 每天凌晨 3 点自动备份
0 3 * * * /usr/bin/bash /opt/codemind/scripts/backup.sh --quiet >> /opt/codemind/logs/backup.log 2>&1
```

---

## 8. 版本升级

### 8.1 升级流程

```
本地打包 → 上传到服务器 → 运行升级脚本
```

### 8.2 准备新版本包

在开发机器上：

```bash
# 确保 VERSION 文件已更新为新版本号
# 执行打包
bash scripts/package.sh
```

### 8.3 上传并升级

```bash
# 上传新版本包
scp dist/codemind-v新版本.tar.gz user@server:/tmp/

# SSH 到服务器
ssh user@server

# 解压
cd /tmp
tar -xzf codemind-v新版本.tar.gz
cd codemind-v新版本

# 执行升级
sudo bash scripts/upgrade.sh
```

### 8.4 升级脚本自动执行

升级脚本会自动完成以下步骤：

1. **备份** — 自动备份当前数据库和配置
2. **停止** — 停止所有服务
3. **更新** — 替换前端、后端、配置文件（**保留 .env 和 app.yaml**）
4. **重建** — 构建新版本 Docker 镜像并启动
5. **迁移** — 应用新增的数据库迁移脚本

### 8.5 升级注意事项

- 升级前会自动备份数据库，备份文件保存在 `/opt/codemind/backups/`
- `.env` 和 `app.yaml` 中的用户配置**不会被覆盖**
- 如需跳过备份：`sudo bash scripts/upgrade.sh --skip-backup`
- 升级失败可从备份恢复（见第 9 节）

---

## 9. 数据备份与恢复

### 9.1 手动备份

```bash
sudo bash /opt/codemind/scripts/backup.sh
```

备份内容包括：
- PostgreSQL 数据库完整转储
- `.env` 配置文件
- `app.yaml` 应用配置
- `nginx.conf` Nginx 配置
- `docker-compose.yml`
- 版本号和迁移记录

备份文件保存在 `/opt/codemind/backups/`，默认保留 **30 天**。

### 9.2 仅备份数据库

```bash
sudo bash /opt/codemind/scripts/backup.sh --db-only
```

### 9.3 修改备份保留天数

```bash
# 设置保留 60 天
sudo BACKUP_RETENTION_DAYS=60 bash /opt/codemind/scripts/backup.sh
```

### 9.4 数据恢复

#### 恢复数据库

```bash
cd /opt/codemind

# 1. 解压备份
cd backups
tar -xzf codemind-backup-v0.3.0-20260220_030000.tar.gz
cd codemind-backup-v0.3.0-20260220_030000

# 2. 确保 PostgreSQL 正在运行
cd /opt/codemind
docker compose up -d postgres

# 3. 等待 PostgreSQL 就绪
sleep 10

# 4. 恢复数据库
docker compose exec -T postgres pg_restore \
    -U codemind -d codemind \
    --clean --if-exists \
    < /opt/codemind/backups/codemind-backup-v0.3.0-20260220_030000/database.dump

# 5. 重启所有服务
docker compose restart
```

#### 恢复配置

```bash
# 从备份目录恢复配置
cp backups/xxx/env.bak /opt/codemind/.env
cp backups/xxx/app.yaml.bak /opt/codemind/config/app.yaml
cp backups/xxx/nginx.conf.bak /opt/codemind/docker/nginx/nginx.conf

# 重启服务
cd /opt/codemind
docker compose down
docker compose up -d
```

---

## 10. 故障排查

### 10.1 常见问题

#### 容器启动失败

```bash
# 查看容器状态
docker compose ps -a

# 查看失败容器的日志
docker compose logs backend
docker compose logs postgres
```

#### 后端连接数据库失败

```bash
# 检查 PostgreSQL 是否正常
docker compose exec postgres pg_isready -U codemind

# 检查数据库密码是否匹配
grep DB_PASSWORD /opt/codemind/.env
```

#### 前端显示空白页

```bash
# 检查 Nginx 日志
docker compose logs frontend

# 检查 Nginx 配置
docker compose exec frontend nginx -t

# 检查前端文件是否正确挂载
docker compose exec frontend ls /usr/share/nginx/html/
```

#### API 请求 502/504

```bash
# 检查后端是否健康
docker compose exec backend wget -qO- http://localhost:8080/health

# 检查后端日志
docker compose logs --tail 50 backend
```

#### 端口被占用

```bash
# 查看端口占用
sudo ss -tlnp | grep 18080

# 修改端口后重启
sudo nano /opt/codemind/.env
cd /opt/codemind && docker compose down && docker compose up -d
```

### 10.2 重置部署

如果需要完全重新部署（**会删除所有数据**）：

```bash
cd /opt/codemind

# 停止并删除所有容器和数据卷
docker compose down -v

# 删除安装目录
sudo rm -rf /opt/codemind

# 重新执行部署
cd /tmp/codemind-v0.3.0
sudo bash scripts/deploy.sh
```

### 10.3 查看资源占用

```bash
# 容器资源使用情况
docker stats --no-stream

# 系统整体资源
free -h && df -h
```

---

## 11. 安全加固

### 11.1 首次部署后必做

- [ ] **修改管理员密码**: 登录后立即修改 `admin` 的默认密码
- [ ] **确认防火墙**: 仅开放必要端口 (18080)
- [ ] **关闭数据库外部端口**: 如无远程数据库管理需求，注释掉 `docker-compose.yml` 中 postgres 和 redis 的 `ports` 映射

### 11.2 关闭不必要的外部端口

编辑 `/opt/codemind/docker-compose.yml`，注释掉数据库和 Redis 的端口映射：

```yaml
  postgres:
    # ports:
    #   - "${DB_EXTERNAL_PORT:-15432}:5432"

  redis:
    # ports:
    #   - "${REDIS_EXTERNAL_PORT:-16379}:6379"
```

然后重启：

```bash
cd /opt/codemind
docker compose down
docker compose up -d
```

### 11.3 配置定时备份

参见 [7.4 配置定时备份](#74-配置定时备份推荐)。

### 11.4 配置 HTTPS（可选）

如果需要 HTTPS 访问，可通过外部反向代理（如宿主机上的 Nginx）实现：

```nginx
# /etc/nginx/sites-available/codemind
server {
    listen 443 ssl;
    server_name your-domain.com;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    location / {
        proxy_pass http://127.0.0.1:18080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # SSE 支持
        proxy_buffering off;
        proxy_cache off;
        proxy_read_timeout 600s;
    }
}
```

---

## 附录

### A. 目录结构

部署完成后的服务器目录结构：

```
/opt/codemind/
├── frontend/                # 前端 Dockerfile + 静态文件
│   ├── Dockerfile
│   └── dist/
├── backend/                 # 后端 Dockerfile + 二进制
│   ├── Dockerfile
│   └── codemind
├── config/                  # 应用配置
│   └── app.yaml
├── docker/                  # Docker 相关配置
│   ├── nginx/
│   │   └── nginx.conf
│   └── postgres/
│       ├── init.sql
│       └── seed.sql
├── migrations/              # 数据库迁移
│   ├── *.sql
│   └── .applied
├── scripts/                 # 管理脚本
│   ├── deploy.sh
│   ├── upgrade.sh
│   ├── backup.sh
│   └── utils.sh
├── backups/                 # 备份文件
├── logs/                    # 日志
├── docker-compose.yml       # 容器编排
├── .env                     # 环境变量（敏感信息）
└── VERSION                  # 当前版本号
```

### B. Docker 命令速查

```bash
cd /opt/codemind

# ── 服务管理 ──
docker compose up -d              # 启动全部
docker compose down               # 停止全部
docker compose restart            # 重启全部
docker compose restart backend    # 重启后端

# ── 日志查看 ──
docker compose logs -f            # 实时日志
docker compose logs --tail 100    # 最近 100 行

# ── 容器操作 ──
docker compose exec postgres psql -U codemind -d codemind   # 进入数据库
docker compose exec backend sh                               # 进入后端容器
docker compose exec frontend sh                              # 进入前端容器

# ── 状态检查 ──
docker compose ps                 # 容器状态
docker stats --no-stream          # 资源使用
```

### C. 完整部署 Checklist

- [ ] 服务器满足最低配置要求
- [ ] Docker 和 Docker Compose 已安装
- [ ] 防火墙已开放 18080 端口
- [ ] 本地执行 `bash scripts/package.sh` 打包成功
- [ ] 压缩包已上传到服务器
- [ ] 执行 `sudo bash scripts/deploy.sh` 部署成功
- [ ] 所有容器状态正常 (`docker compose ps`)
- [ ] 浏览器可以访问登录页面
- [ ] 使用 admin 账号成功登录
- [ ] 已修改管理员默认密码
- [ ] LLM 服务地址已正确配置
- [ ] 已配置定时备份（推荐）
- [ ] 已关闭不必要的外部端口（推荐）
