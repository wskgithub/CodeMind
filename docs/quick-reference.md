# CodeMind 快速参考

> 常用命令和代码片段速查手册。

---

## 目录

- [命令速查](#命令速查)
- [常用代码片段](#常用代码片段)
- [项目结构速览](#项目结构速览)
- [版本管理](#版本管理)

---

## 命令速查

### 后端命令（在 `backend/` 目录执行）

```bash
# 开发运行
go run cmd/server/main.go

# 构建二进制
make build                    # 输出到 bin/codemind

# 测试
make test                     # 运行所有测试
go test ./... -cover          # 查看覆盖率
make test-coverage            # 生成 HTML 覆盖率报告

# 代码检查
make lint                     # 运行 golangci-lint

# 其他
make clean                    # 清理构建产物
make swagger                  # 生成 Swagger 文档
make docker-build             # 构建 Docker 镜像
```

### 前端命令（在 `frontend/` 目录执行）

```bash
# 安装依赖
npm install

# 开发服务器（端口 3000）
npm run dev

# 构建（输出到 dist/）
npm run build

# 测试
npm run test                  # 运行 Vitest

# 代码检查
npm run lint
```

### Docker 命令（在项目根目录执行）

```bash
# 开发环境：仅启动基础设施（PostgreSQL + Redis）
docker compose up -d postgres redis

# 生产环境：构建并启动所有服务
docker compose up -d --build

# 查看日志
docker compose logs -f backend
docker compose logs -f postgres

# 停止所有服务
docker compose down
```

---

## 常用代码片段

### 后端：添加新的 API 端点

```go
// 1. internal/model/dto/request.go - 定义请求结构
type CreateXxxRequest struct {
    Name string `json:"name" binding:"required"`
}

// 2. internal/service/xxx.go - 实现业务逻辑
func (s *XxxService) Create(ctx context.Context, req *dto.CreateXxxRequest) error {
    // 业务验证 → 调用 Repository → 记录审计日志
}

// 3. internal/handler/xxx.go - 实现 Handler
func (h *XxxHandler) Create(c *gin.Context) {
    var req dto.CreateXxxRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, errcode.ErrCodeInvalidParam, err.Error())
        return
    }
    // 调用 Service → 返回响应
}

// 4. internal/router/router.go - 注册路由
xxxGroup := apiV1.Group("/xxx")
{
    xxxGroup.POST("", handlers.Xxx.Create)
}

// 5. main.go - 初始化（如需新增依赖）
```

### 前端：添加新的页面

```typescript
// 1. src/pages/xxx/XxxPage.tsx - 创建页面组件
export function XxxPage() {
  // 使用 hooks 获取数据
  // 渲染 Ant Design 组件
}

// 2. src/services/xxxService.ts - 创建服务
export const xxxService = {
  list: () => request.get('/api/v1/xxx'),
  create: (data: CreateXxxRequest) => request.post('/api/v1/xxx', data),
};

// 3. src/router/index.tsx - 添加路由
{ path: 'xxx', element: <XxxPage /> }

// 4. 在导航菜单中添加入口（如需要）
```

### 日志规范（Zap）

```go
log.Info("用户创建成功",
    zap.Int64("user_id", user.ID),
    zap.String("username", user.Username),
)
```

---

## 项目结构速览

```
CodeMind/
├── frontend/              # React 前端应用
│   ├── src/
│   │   ├── pages/         # 页面组件（路由驱动）
│   │   ├── components/    # 可复用组件
│   │   ├── services/      # API 客户端（Axios）
│   │   ├── store/         # Zustand 状态管理
│   │   ├── router/        # React Router 配置
│   │   └── types/         # TypeScript 类型定义
│   ├── package.json       # NPM 配置
│   └── vite.config.ts     # Vite 构建配置
│
├── backend/               # Go 后端 API 服务
│   ├── cmd/server/
│   │   └── main.go        # 应用入口（依赖注入初始化）
│   ├── internal/
│   │   ├── config/        # 配置管理（Viper）
│   │   ├── handler/       # HTTP 处理器（Controller）
│   │   ├── service/       # 业务逻辑层
│   │   ├── repository/    # 数据访问层（GORM）
│   │   ├── model/         # 数据模型与 DTO
│   │   ├── middleware/    # HTTP 中间件
│   │   ├── router/        # 路由定义（Gin）
│   │   └── pkg/           # 内部工具（JWT、加密、响应、错误码）
│   ├── pkg/               # 外部共享包
│   │   ├── llm/           # LLM 客户端（OpenAI/Anthropic）
│   │   ├── mcp/           # MCP 协议网关
│   │   └── token/         # Token 计数器
│   ├── migrations/        # SQL 迁移文件（编号前缀 001_, 002_...）
│   ├── go.mod             # Go 模块定义
│   └── Makefile           # 构建脚本
│
├── deploy/                # Docker & 部署配置
│   ├── config/            # 应用配置（app.yaml）
│   └── docker/            # Nginx、PostgreSQL 配置
│
├── docs/                  # 项目文档
├── scripts/               # 开发与部署脚本
├── docker-compose.yml     # 容器编排
├── CHANGELOG.md           # 版本变更日志
└── VERSION                # 当前版本号（单行）
```

---

## 版本管理

当前版本：见 [VERSION](../VERSION) 文件（单行版本号，如 `0.4.0`）

版本历史：见 [CHANGELOG.md](../CHANGELOG.md)

升级版本时需：
1. 更新 `VERSION` 文件
2. 在 `CHANGELOG.md` 添加变更记录
3. 前端会自动从根目录 VERSION 文件读取版本号
4. 使用 `scripts/package.sh` 打包部署

---

## 相关文档

- [开发环境搭建](./dev-setup.md) - 完整的开发环境配置步骤
- [部署与运维](./deployment-guide.md) - 生产环境部署指南
- [后端规范](./backend-standards.md) - 后端代码规范
- [前端规范](./frontend-standards.md) - 前端代码规范
- [测试指南](./testing-guide.md) - 测试编写指南
