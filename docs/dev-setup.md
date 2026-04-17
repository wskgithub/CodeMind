# CodeMind 开发环境搭建与启动指南

## 环境要求

- Go 1.23+
- Node.js 18+
- Docker & Docker Compose
- macOS / Linux

## 第一步：安装 Go 环境

### macOS（推荐使用 Homebrew）

```bash
# 安装 Go 1.23
brew install go@1.23

# 验证安装
go version  # 应该显示 go1.23.x
```

### 手动安装

1. 访问 https://go.dev/dl/
2. 下载对应平台的安装包
3. 按提示完成安装，运行 `go version` 验证

> 如果提示 "command not found"，编辑 `~/.zshrc` 添加：
> ```bash
> export PATH=$PATH:/usr/local/go/bin
> export GOPATH=$HOME/go
> export PATH=$PATH:$GOPATH/bin
> ```
> 然后执行 `source ~/.zshrc`。

## 第二步：启动基础设施

```bash
cd /path/to/CodeMind

# 启动 PostgreSQL + Redis
docker compose up -d postgres redis

# 等待数据库就绪（约 10 秒）
sleep 10

# 查看状态
docker compose ps
```

默认端口：
- PostgreSQL: `5434`
- Redis: `6379`

## 第三步：启动后端服务

```bash
cd backend

# 下载依赖（首次运行）
go mod download

# 启动后端
go run cmd/server/main.go
```

后端运行在 `http://localhost:8080`。

## 第四步：启动前端服务

```bash
cd frontend

# 安装依赖（首次运行）
npm install

# 启动开发服务器
npm run dev
```

前端运行在 `http://localhost:3000`。

## 第五步：访问应用

打开浏览器访问 **http://localhost:3000**

默认管理员账号：
- 用户名：`admin`
- 密码：`Admin@123456`

## 配置文件

| 配置 | 路径 |
|------|------|
| 环境变量 | `.env` |
| 后端配置 | `deploy/config/app.yaml` |
| 前端配置 | `frontend/vite.config.ts` |

## API 端点

| 端点 | 说明 |
|------|------|
| `http://localhost:8080/health` | 健康检查 |
| `http://localhost:8080/api/v1/*` | 管理 API |
| `http://localhost:8080/v1/*` | LLM 代理 |

## 测试 API

```bash
# 健康检查
curl http://localhost:8080/health

# 登录获取 Token
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"Admin@123456"}'

# 获取用户信息（替换 YOUR_TOKEN）
curl http://localhost:8080/api/v1/auth/profile \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## 停止与重启

```bash
# 停止后端/前端：在对应终端按 Ctrl+C

# 停止基础设施
docker compose down

# 完整重启
docker compose up -d postgres redis
sleep 10
cd backend && go run cmd/server/main.go  # 新终端
cd frontend && npm run dev               # 新终端
```

## 常见问题

### Q: 后端连接数据库失败
检查 PostgreSQL 是否运行：
```bash
docker compose ps postgres
```
确认 `deploy/config/app.yaml` 中数据库配置（host、port、password）正确。

### Q: 前端请求 API 失败（CORS 错误）
确认后端在 8080 端口正常运行：
```bash
curl http://localhost:8080/health
```
检查 `frontend/vite.config.ts` 代理配置指向 `http://localhost:8080`。

### Q: 端口被占用
- PostgreSQL (5434): 修改 `.env` 中的 `DB_EXTERNAL_PORT`
- Redis (6379): 修改 `.env` 中的 `REDIS_EXTERNAL_PORT`
- 后端 (8080): 修改 `deploy/config/app.yaml` 中的 `server.port`
- 前端 (3000): 修改 `frontend/vite.config.ts` 中的 `server.port`

### Q: Go 依赖下载慢
已配置国内镜像，如需切换：
```bash
go env -w GOPROXY=https://goproxy.cn,direct
```

## 下一步

1. 修改默认密码：登录后在"个人中心"修改管理员密码
2. 配置 LLM 服务：在 `deploy/config/app.yaml` 中配置 LLM 服务地址
3. 创建部门和用户：在"部门管理"和"用户管理"中添加组织结构
4. 生成 API Key：在"API 密钥"页面生成密钥用于 LLM 调用
