# CodeMind 快速启动指南

## ✅ 当前运行状态

### 基础设施
- ✅ **PostgreSQL** - 运行在 `localhost:5434`
- ✅ **Redis** - 运行在 `localhost:6379`

### 应用服务
- ✅ **后端 API** - 运行在 `http://localhost:8080`
- ✅ **前端界面** - 运行在 `http://localhost:3000`

## 🚀 访问应用

打开浏览器访问：**http://localhost:3000**

### 默认管理员账号
```
用户名：admin
密码：Admin@123456
```

## 📋 常用操作

### 查看服务状态

```bash
# 查看基础设施服务
docker compose ps

# 查看后端日志（在运行后端的终端窗口）
# 或者查看日志文件
tail -f /Users/wangsk/.cursor/projects/Users-wangsk-workspace-projects-myproject-CodeMind/terminals/419659.txt

# 查看前端日志（在运行前端的终端窗口）
# 或者查看日志文件
tail -f /Users/wangsk/.cursor/projects/Users-wangsk-workspace-projects-myproject-CodeMind/terminals/579093.txt
```

### 停止服务

```bash
# 停止后端：在运行后端的终端按 Ctrl+C

# 停止前端：在运行前端的终端按 Ctrl+C

# 停止基础设施
cd /Users/wangsk/workspace/projects/myproject/CodeMind
docker compose down
```

### 重启服务

```bash
# 1. 重启基础设施（如果已停止）
cd /Users/wangsk/workspace/projects/myproject/CodeMind
docker compose up -d postgres redis

# 2. 重启后端（新终端窗口）
cd /Users/wangsk/workspace/projects/myproject/CodeMind/backend
go run cmd/server/main.go

# 3. 重启前端（新终端窗口）
cd /Users/wangsk/workspace/projects/myproject/CodeMind/frontend
npm run dev
```

## 🔧 配置文件位置

- **环境变量**: `.env`
- **后端配置**: `deploy/config/app.yaml`
- **前端配置**: `frontend/vite.config.ts`

## 📚 API 文档

- **健康检查**: http://localhost:8080/health
- **管理 API**: http://localhost:8080/api/v1/*
- **LLM 代理**: http://localhost:8080/v1/*

## 🧪 测试 API

```bash
# 健康检查
curl http://localhost:8080/health

# 登录获取 Token
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"Admin@123456"}'

# 获取用户信息（需要替换 YOUR_TOKEN）
curl http://localhost:8080/api/v1/auth/profile \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## 🐛 常见问题

### Q: 后端连接数据库失败
**A**: 检查 PostgreSQL 是否运行：
```bash
docker compose ps postgres
```
如果未运行，执行：
```bash
docker compose up -d postgres
```

### Q: 前端请求 API 失败（CORS 错误）
**A**: 检查后端是否正常运行在 8080 端口：
```bash
curl http://localhost:8080/health
```

### Q: 端口被占用
**A**: 
- PostgreSQL (5434): 修改 `.env` 中的 `DB_EXTERNAL_PORT`
- Redis (6379): 修改 `.env` 中的 `REDIS_EXTERNAL_PORT`
- 后端 (8080): 修改 `deploy/config/app.yaml` 中的 `server.port`
- 前端 (3000): 修改 `frontend/vite.config.ts` 中的 `server.port`

### Q: Go 依赖下载慢
**A**: 已配置国内镜像 `https://goproxy.cn`，如需切换：
```bash
go env -w GOPROXY=https://goproxy.io,direct
```

## 📖 更多文档

- **项目架构**: 查看 `README.md`
- **开发规范**: 查看 `CLAUDE.md`
- **更新日志**: 查看 `CHANGELOG.md`
- **详细搭建**: 查看 `SETUP-DEV.md`

## 🎯 下一步

1. **修改默认密码**: 登录后在"个人中心"修改管理员密码
2. **配置 LLM 服务**: 在 `deploy/config/app.yaml` 中配置你的 LLM 服务地址
3. **创建部门和用户**: 在"部门管理"和"用户管理"中添加组织结构
4. **生成 API Key**: 在"API 密钥"页面生成密钥用于 LLM 调用

---

**祝你使用愉快！** 🚀
