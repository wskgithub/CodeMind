# CodeMind —— 企业级 AI 编码服务管理平台

> 本文档为 AI 编码助手提供项目导航和文档索引。
> 
> **代码注释使用中文**（重要约定）

---

## 项目简介

CodeMind 是一个企业级 AI 编码服务上层管理平台，作为中间层接入本地部署的 LLM 编码模型，为公司内部员工提供统一的 AI 编码服务入口。

**核心功能**：用户分级管理、API Key 管理、LLM 请求代理、Token 用量统计、三级限额管控、安全审计

**技术栈**：React + TypeScript + Go + Gin + PostgreSQL + Redis + Docker

---

## 文档索引

根据任务需要，按需阅读对应文档：

### 核心架构

| 文档 | 说明 | 何时阅读 |
|------|------|---------|
| [docs/architecture.md](docs/architecture.md) | 系统架构、项目结构、请求流程 | 首次了解项目、修改架构代码 |
| [docs/llm-proxy.md](docs/llm-proxy.md) | LLM 代理、多服务商路由、MCP 网关 | 修改代理逻辑、添加服务商 |
| [docs/api-routes.md](docs/api-routes.md) | API 路由分组、认证方式 | 添加/修改 API 端点 |
| [docs/design-patterns.md](docs/design-patterns.md) | RBAC 角色、限流、软删除、登录锁定 | 涉及权限、认证、限流逻辑 |

### 开发规范

| 文档 | 说明 | 何时阅读 |
|------|------|---------|
| [docs/development-standards.md](docs/development-standards.md) | 开发工作流、代码质量、安全规范 | 开始开发前 |
| [docs/backend-standards.md](docs/backend-standards.md) | 后端代码规范、分层架构、错误处理 | 编写后端代码 |
| [docs/frontend-standards.md](docs/frontend-standards.md) | 前端代码规范、组件规范、样式规范 | 编写前端代码 |
| [docs/testing-guide.md](docs/testing-guide.md) | 测试原则、测试模板、覆盖率要求 | 编写测试 |

### 配置与部署

| 文档 | 说明 | 何时阅读 |
|------|------|---------|
| [docs/dev-setup.md](docs/dev-setup.md) | 开发环境搭建步骤 | 首次配置开发环境 |
| [docs/configuration.md](docs/configuration.md) | 配置加载、环境变量、Redis Key 模式 | 配置环境、调试配置问题 |
| [docs/deployment-guide.md](docs/deployment-guide.md) | 部署与运维手册、升级、备份、故障排查 | 部署或运维 |
| [docs/security.md](docs/security.md) | 密码、API Key、JWT、审计日志安全规范 | 涉及认证、加密、审计 |

### 快速参考

| 文档 | 说明 | 何时阅读 |
|------|------|---------|
| [docs/quick-reference.md](docs/quick-reference.md) | 命令速查、常用代码片段、项目结构 | 日常开发查阅 |
| [docs/troubleshooting.md](docs/troubleshooting.md) | 开发环境故障排查 | 遇到问题 |
| [docs/fixes.md](docs/fixes.md) | 已修复问题的记录 | 排查类似问题 |

### 项目管理

| 文档 | 说明 | 何时阅读 |
|------|------|---------|
| [docs/development-plan.md](docs/development-plan.md) | 完整开发计划、数据库设计、API 设计 | 了解整体设计决策 |
| [docs/status.md](docs/status.md) | 项目当前状态报告 | 了解项目进度 |
| [CHANGELOG.md](CHANGELOG.md) | 版本变更日志 | 了解版本历史 |

---

## 快速开始

```bash
# 1. 启动基础设施
docker compose up -d postgres redis

# 2. 启动后端（新终端）
cd backend && go run cmd/server/main.go

# 3. 启动前端（新终端）
cd frontend && npm install && npm run dev

# 4. 访问 http://localhost:3000
# 默认账号: admin / Admin@123456
```

详细步骤见：[docs/dev-setup.md](docs/dev-setup.md)

---

*本文档由 AI 编码助手维护，详细内容请参考 `docs/` 目录下的各专题文档。*
