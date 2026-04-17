# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Coding Conventions

- **Code comments should be written in Chinese** (代码注释使用中文)
- When task is complete or clarification is needed, ask for further instructions

## Project

CodeMind — Enterprise AI Coding Service Management Platform. Proxies LLM requests and provides user management, resource control, and usage analytics.

**Tech Stack**: React 18 + TypeScript + Vite | Go 1.24 + Gin + GORM | PostgreSQL 16 | Redis 7 | Docker

## 文档索引

根据任务需要，在 `docs/` 中按需阅读对应文档。

### 核心架构

| 文档 | 说明 | 何时阅读 |
|------|------|---------|
| [docs/architecture.md](docs/architecture.md) | 系统架构、项目结构、请求流程 | 首次了解项目、修改架构相关代码时 |
| [docs/llm-proxy.md](docs/llm-proxy.md) | LLM 代理、多服务商路由、第三方代理、MCP 网关 | 修改代理逻辑、添加服务商、调试路由时 |
| [docs/api-routes.md](docs/api-routes.md) | API 路由分组、认证方式、Nginx 路由 | 添加/修改 API 端点、调试认证问题时 |
| [docs/design-patterns.md](docs/design-patterns.md) | RBAC 角色、API Key 格式、限流、软删除、登录锁定 | 涉及权限、认证、限流、删除逻辑时 |

### 配置与安全

| 文档 | 说明 | 何时阅读 |
|------|------|---------|
| [docs/configuration.md](docs/configuration.md) | 配置加载、环境变量、Redis Key 模式、开发命令 | 配置环境、调试配置问题、查看 Redis Key 时 |
| [docs/security.md](docs/security.md) | 密码、API Key、JWT、审计日志安全规范 | 涉及认证、加密、审计日志时 |

### 开发规范

| 文档 | 说明 | 何时阅读 |
|------|------|---------|
| [docs/development-standards.md](docs/development-standards.md) | 开发工作流、代码质量、安全规范 | 开始开发前 |
| [docs/backend-standards.md](docs/backend-standards.md) | 后端代码规范、分层架构、错误处理 | 编写后端代码时 |
| [docs/frontend-standards.md](docs/frontend-standards.md) | 前端代码规范、组件规范、样式规范 | 编写前端代码时 |
| [docs/testing-guide.md](docs/testing-guide.md) | 测试原则、测试模板、覆盖率要求 | 编写测试时 |

### 环境与部署

| 文档 | 说明 | 何时阅读 |
|------|------|---------|
| [docs/dev-setup.md](docs/dev-setup.md) | 开发环境搭建步骤 | 首次配置开发环境时 |
| [docs/deployment-guide.md](docs/deployment-guide.md) | 部署与运维手册、升级、备份、故障排查 | 部署或运维时 |
| [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md) | 云服务器部署手册（Docker 容器化） | 服务器首次部署时 |

### 项目管理

| 文档 | 说明 | 何时阅读 |
|------|------|---------|
| [docs/development-plan.md](docs/development-plan.md) | 完整开发计划、数据库设计、API 设计 | 了解整体规划或设计决策时 |
| [docs/permission-optimization.md](docs/permission-optimization.md) | 权限矩阵与优化记录 | 修改权限相关功能时 |
| [docs/monitoring.md](docs/monitoring.md) | 系统监控仪表盘功能说明 | 修改监控相关功能时 |
