# 配置管理

> 本文档描述 CodeMind 的配置加载机制、环境变量和基础设施配置。

## 配置加载顺序

```
默认值 → YAML 配置文件 → 环境变量覆盖
```

环境变量使用 `_` 替代 `.` 作为层级分隔符，例如：
- `database.host` → `DB_HOST`
- `database.port` → `DB_PORT`

## 环境搭建

### 步骤

1. 复制 `.env.example` 到 `.env` 并填写必要值
2. 复制 `deploy/config/app.yaml.example` 到 `deploy/config/app.yaml`
3. 最少必要环境变量：`DB_PASSWORD`、`JWT_SECRET`、`LLM_BASE_URL`

### 默认凭据

- 用户名：`admin`
- 密码：`Admin@123456`（首次登录后请修改）

## Redis Key 模式

| Key 模式 | 类型 | TTL | 用途 |
|----------|------|-----|------|
| `codemind:jwt:blacklist:{jti}` | String | JWT 剩余时间 | JWT 黑名单 |
| `codemind:apikey:{hash}` | String(JSON) | 300s | API Key 缓存 |
| `codemind:concurrency:{user_id}` | Counter | 300s | 并发请求计数 |
| `codemind:usage:{user_id}:daily:{date}` | Counter | 48h | 每日 Token 用量 |

## 常用开发命令

### 基础设施（Docker）

```bash
# 仅启动基础设施（PostgreSQL + Redis）
docker compose up -d postgres redis

# 启动所有服务（生产）
docker compose up -d --build

# 停止所有服务
docker compose down

# 查看日志
docker compose logs -f [service]
```

### 后端

```bash
cd backend

# 启动开发服务器（需要基础设施运行中）
go run cmd/server/main.go

# 构建
make build

# 运行所有测试（含覆盖率）
make test

# 运行单个测试
go test ./internal/service/ -run TestUserService -v

# 生成覆盖率报告（HTML）
make test-coverage

# 代码检查
make lint

# 生成 Swagger 文档
make swagger

# 数据库迁移
make migrate
```

**后端入口**：`backend/cmd/server/main.go`

### 前端

```bash
cd frontend

# 安装依赖
npm install

# 启动开发服务器（通过 Vite 代理 API 到 localhost:8080）
npm run dev

# 生产构建
npm run build

# 运行测试
npm test -- --run

# 代码检查
npm run lint
```

### 项目级脚本

```bash
./scripts/dev.sh      # 启动开发基础设施（postgres + redis）
./scripts/build.sh    # 构建所有 Docker 镜像
./scripts/test.sh     # 运行所有测试（后端 + 前端）
```

## 相关文档

- [开发环境搭建](dev-setup.md) — 详细的开发环境配置指南
- [部署指南](deployment-guide.md) — 生产环境部署
- [系统架构概览](architecture.md) — 整体架构
- [安全要求](security.md) — JWT、加密等安全配置
