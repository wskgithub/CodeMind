# CodeMind

> 企业级 AI 编码服务管理平台，基于本地部署的 LLM 模型为团队提供安全、可控、高效的 AI 编码辅助服务。

## Overview

CodeMind 是一个 AI 编码服务上层管理平台，类似于 Kimi、GLM 等提供的开发者控制台。它作为中间层，接入本地服务器部署的 LLM 编码模型，为公司内部员工提供统一的 AI 编码服务入口，并实现用户管理、资源管控和用量统计等功能。

## Features

- **用户分级管理** — 超级管理员 / 部门经理 / 普通用户三级角色体系
- **API Key 管理** — 用户通过 API Key 接入服务，支持创建、禁用、过期设置
- **LLM 请求代理** — 兼容 OpenAI API 标准格式，支持流式传输（SSE）
- **Token 用量统计** — 按日 / 周 / 月维度统计，可视化图表展示
- **灵活的限额管控** — 全局 / 部门 / 用户三级限额配置，并发请求数控制
- **安全审计** — 完整的操作审计日志，敏感数据加密存储

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Frontend | React 18 + TypeScript + Vite + Ant Design 5 + TailwindCSS |
| Backend | Go 1.23 + Gin + GORM |
| Database | PostgreSQL 16 |
| Cache | Redis 7 |
| Deployment | Docker + Docker Compose + Nginx |

## Project Structure

```
CodeMind/
├── frontend/          # React frontend application
├── backend/           # Go backend API server
├── deploy/            # Docker & deployment configs
├── docs/              # Project documentation
├── scripts/           # Development & build scripts
├── docker-compose.yml # Container orchestration
├── CHANGELOG.md       # Version changelog
└── VERSION            # Current version
```

## Quick Start

### Prerequisites

- Docker >= 27.x
- Docker Compose >= 2.x
- Node.js >= 20.x (for frontend development)
- Go >= 1.23 (for backend development)

### Development Setup

```bash
# Clone the repository
git clone <repository-url>
cd CodeMind

# Copy configuration template
cp deploy/config/app.yaml.example deploy/config/app.yaml

# Start infrastructure services (PostgreSQL + Redis)
docker compose up -d postgres redis

# Start backend (in a new terminal)
cd backend
go run cmd/server/main.go

# Start frontend (in a new terminal)
cd frontend
npm install
npm run dev
```

### Production Deployment

```bash
# Configure environment
cp deploy/config/app.yaml.example deploy/config/app.yaml
# Edit deploy/config/app.yaml with production values

# Build and start all services
docker compose up -d --build
```

## Default Credentials

| Username | Password | Role |
|----------|----------|------|
| admin | Admin@123456 | Super Admin |

> Please change the default password immediately after first login.

## Documentation

- [Development Setup](docs/dev-setup.md) — 开发环境搭建与启动指南
- [Development Plan](docs/development-plan.md) — Detailed development plan and system design
- [Development Standards](docs/development-standards.md) — 开发规范
- [Backend Standards](docs/backend-standards.md) — 后端编码规范
- [Frontend Standards](docs/frontend-standards.md) — 前端编码规范
- [Testing Guide](docs/testing-guide.md) — 测试指南
- [Monitoring](docs/monitoring.md) — 监控说明

## Version

Current version: see [VERSION](VERSION)

Version history: see [CHANGELOG](CHANGELOG.md)

## License

This project is proprietary software for internal use only.

Copyright © 2026 CodeMind. All rights reserved.
