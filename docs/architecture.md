# 系统架构概览

> 本文档描述 CodeMind 的整体架构、技术栈、项目结构和请求流程。

## 项目简介

CodeMind 是企业级 AI 编码服务管理平台。作为中间件层，代理请求到本地部署的 LLM 模型和第三方 AI 服务商，提供用户管理、资源控制和用量统计功能。

## 技术栈

| 层级 | 技术 | 版本 |
|------|------|------|
| 前端 | React + TypeScript + Vite + Ant Design + TailwindCSS | 18 / 5 / 6 |
| 后端 | Go + Gin + GORM | 1.24 |
| 数据库 | PostgreSQL | 16 |
| 缓存 | Redis | 7 |
| 部署 | Docker + Docker Compose + Nginx | - |

## 后端结构（分层架构）

```
backend/
├── cmd/server/          # 应用入口
├── internal/
│   ├── config/          # 配置管理（Viper: YAML + 环境变量）
│   ├── middleware/      # HTTP 中间件（认证、CORS、限流等）
│   ├── handler/         # HTTP 处理器（控制器）
│   ├── service/         # 业务逻辑层
│   ├── repository/      # 数据访问层（GORM）
│   ├── model/           # 数据模型与 DTO
│   ├── router/          # 路由定义（Gin）
│   └── pkg/             # 内部工具（JWT、加密、响应、错误码）
├── pkg/                 # 外部共享包（LLM 客户端、Token 计数器）
└── migrations/          # SQL 迁移文件（编号前缀: 001_, 002_, ...）
```

> 测试文件与源文件同目录，遵循 Go 惯例（如 `service/user.go` + `service/user_test.go`）。

**依赖注入**：应用在 `main.go` 中自底向上初始化：Config → DB/Redis → Repositories → Services → Handlers → Router，所有依赖通过构造函数注入。

## 前端结构

```
frontend/src/
├── pages/               # 页面组件（路由驱动，懒加载）
│   ├── admin/           # 管理页面（用户、部门、限额、系统等）
│   │   ├── templates/   # 第三方服务商模板管理
│   │   └── platform/    # 平台设置
│   └── models/          # 模型服务页面（用户的第三方服务商）
├── components/
│   ├── common/          # 可复用组件（UsageProgressCards 等）
│   └── layout/          # 布局组件（DashboardLayout）
├── services/            # API 客户端层（Axios，基础路径: /api/v1）
├── store/               # Zustand 状态管理
│   ├── authStore.ts     # 认证状态（token/用户信息存于 localStorage）
│   └── appStore.ts      # UI 状态（主题、侧边栏）
├── router/              # React Router，含 AuthGuard（角色控制）和 GuestGuard
└── types/               # TypeScript 类型定义
```

**前端样式**：Ant Design + Tailwind 工具类 + CSS 变量内联样式。Preflight 已禁用。Vite 构建将 React、Ant Design、ECharts 拆分为独立 vendor chunk。

## 请求流程

### 管理请求

```
Router → 中间件链 → Handler → Service → Repository → PostgreSQL
```

### LLM 代理请求

```
API Key 验证 → 限流检查 → 并发检查 → 负载均衡选择服务商
→ 转发到 LLM → 流式响应 → 记录用量
```

### 第三方代理请求

```
API Key 验证 → 解析模型对应的用户第三方服务商 → 透明代理（无配额/并发控制）→ 记录用量
```

## 相关文档

- [API 路由与认证](api-routes.md) — 各路由组的认证方式和访问权限
- [LLM 代理架构](llm-proxy.md) — 代理层详细设计
- [核心设计模式](design-patterns.md) — RBAC、API Key、限流等设计
- [配置管理](configuration.md) — 配置加载与环境变量
- [后端开发规范](backend-standards.md) — 后端代码规范
- [前端开发规范](frontend-standards.md) — 前端代码规范
