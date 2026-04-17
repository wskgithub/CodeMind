# CodeMind 故障排查指南

> 开发环境常见问题诊断与解决方案。

---

## 目录

- [后端连接数据库失败](#后端连接数据库失败)
- [前端 CORS 错误](#前端-cors-错误)
- [端口冲突](#端口冲突)
- [Go 依赖下载慢](#go-依赖下载慢)
- [前端依赖安装失败](#前端依赖安装失败)

---

## 后端连接数据库失败

**症状**：后端启动时报数据库连接错误

```
dial tcp localhost:5434: connect: connection refused
```

**排查步骤**：

```bash
# 1. 检查 PostgreSQL 容器状态
docker compose ps postgres

# 2. 查看数据库日志
docker compose logs postgres

# 3. 确认数据库已就绪
docker compose exec postgres pg_isready -U codemind

# 4. 验证连接信息
cat deploy/config/app.yaml | grep -A5 database
```

**解决方案**：
- 确保 PostgreSQL 容器已启动：`docker compose up -d postgres`
- 等待数据库完全初始化（约 10 秒）
- 检查 `deploy/config/app.yaml` 中数据库配置（host、port、password）
- 确认 `.env` 中的密码与配置文件中一致

---

## 前端 CORS 错误

**症状**：浏览器控制台显示跨域错误

```
Access to XMLHttpRequest at 'http://localhost:8080/api/v1/auth/login' 
from origin 'http://localhost:3000' has been blocked by CORS policy
```

**排查步骤**：

```bash
# 1. 确认后端在 8080 端口运行
curl http://localhost:8080/health

# 2. 检查前端代理配置
cat frontend/vite.config.ts | grep -A5 proxy
```

**解决方案**：
- 确保后端服务已启动：`cd backend && go run cmd/server/main.go`
- 检查 `frontend/vite.config.ts` 代理配置指向 `http://localhost:8080`
- 清除浏览器缓存后重试

---

## 端口冲突

**症状**：启动服务时报端口被占用错误

```
bind: address already in use
```

| 服务 | 默认端口 | 修改位置 |
|------|---------|---------|
| PostgreSQL | 5432/5434 | `.env` 中的 `DB_EXTERNAL_PORT` |
| Redis | 6379 | `.env` 中的 `REDIS_EXTERNAL_PORT` |
| 后端 | 8080 | `deploy/config/app.yaml` 中的 `server.port` |
| 前端 | 3000 | `frontend/vite.config.ts` 中的 `server.port` |

**查找占用端口的进程**：

```bash
# macOS
lsof -i :8080

# Linux
ss -tlnp | grep 8080
```

---

## Go 依赖下载慢

**症状**：`go mod download` 执行缓慢或超时

**解决方案**：

已配置国内镜像，如需切换：

```bash
# 设置国内代理
go env -w GOPROXY=https://goproxy.cn,direct

# 验证设置
go env GOPROXY
```

---

## 前端依赖安装失败

**症状**：`npm install` 报错或卡住

**解决方案**：

```bash
# 1. 清除 npm 缓存
npm cache clean --force

# 2. 删除 node_modules 重新安装
rm -rf node_modules package-lock.json
npm install

# 3. 或使用国内镜像
npm config set registry https://registry.npmmirror.com
```

---

## 其他问题

如以上方案无法解决问题，请：

1. 查看相关服务的详细日志
2. 参考 [部署与运维手册](./deployment-guide.md) 的故障排查章节
3. 查阅 [fixes.md](./fixes.md) 已修复问题的记录
