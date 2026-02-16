# CodeMind 开发环境快速搭建指南

## 当前状态

✅ PostgreSQL (端口 5434) - 已启动  
✅ Redis (端口 6379) - 已启动  
⏳ 后端 - 需要 Go 环境  
⏳ 前端 - 需要 Node.js 环境

## 第一步：安装 Go 环境

### macOS (推荐使用 Homebrew)

```bash
# 安装 Homebrew (如果还没有)
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# 安装 Go 1.23
brew install go@1.23

# 验证安装
go version  # 应该显示 go1.23.x
```

### 或者手动下载安装

1. 访问 https://go.dev/dl/
2. 下载 `go1.23.x.darwin-amd64.pkg` (Intel Mac) 或 `go1.23.x.darwin-arm64.pkg` (M1/M2 Mac)
3. 双击安装包，按提示完成安装
4. 打开新终端窗口，运行 `go version` 验证

## 第二步：启动后端服务

```bash
cd /Users/wangsk/workspace/projects/myproject/CodeMind/backend

# 下载 Go 依赖（首次运行）
go mod download

# 启动后端服务
go run cmd/server/main.go
```

后端将在 `http://localhost:8080` 启动。

## 第三步：启动前端服务

```bash
cd /Users/wangsk/workspace/projects/myproject/CodeMind/frontend

# 安装依赖（首次运行）
npm install

# 启动开发服务器
npm run dev
```

前端将在 `http://localhost:5173` 启动。

## 第四步：访问应用

打开浏览器访问 http://localhost:5173

**默认管理员账号：**
- 用户名：`admin`
- 密码：`Admin@123456`

## 常见问题

### Q: Go 安装后提示 "command not found"

A: 需要配置环境变量。编辑 `~/.zshrc` 或 `~/.bash_profile`，添加：

```bash
export PATH=$PATH:/usr/local/go/bin
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
```

然后运行 `source ~/.zshrc` 重新加载配置。

### Q: 后端连接数据库失败

A: 检查 `deploy/config/app.yaml` 中的数据库配置：
- `host: "localhost"`
- `port: 5434`
- `password: "codemind_dev_2026"`

### Q: 前端请求后端 API 失败

A: 检查 `frontend/vite.config.ts` 中的代理配置是否正确指向 `http://localhost:8080`。

### Q: 如何停止服务

- 后端：在终端按 `Ctrl+C`
- 前端：在终端按 `Ctrl+C`
- 数据库/Redis：`docker compose down`

## 完整重启流程

```bash
# 1. 停止所有服务
docker compose down

# 2. 启动基础设施
docker compose up -d postgres redis

# 3. 等待数据库就绪（约 10 秒）
sleep 10

# 4. 启动后端（新终端窗口）
cd backend && go run cmd/server/main.go

# 5. 启动前端（新终端窗口）
cd frontend && npm run dev
```

## 下一步

开发环境搭建完成后，可以：

1. 查看 `README.md` 了解项目架构
2. 查看 `CLAUDE.md` 了解开发规范
3. 查看 `CHANGELOG.md` 了解功能列表
4. 开始开发新功能！
