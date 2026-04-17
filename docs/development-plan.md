# CodeMind — 开发计划文档

# CodeMind — Development Plan

---

> **文档版本**: v1.0.0
> **创建日期**: 2026-02-15
> **项目代号**: CodeMind
> **文档状态**: 初稿

---

## 目录

- [1. 项目概述](#1-项目概述)
- [2. 需求分析](#2-需求分析)
- [3. 技术栈选型](#3-技术栈选型)
- [4. 系统架构设计](#4-系统架构设计)
- [5. 数据库设计](#5-数据库设计)
- [6. API 接口设计](#6-api-接口设计)
- [7. 前端模块设计](#7-前端模块设计)
- [8. 后端模块设计](#8-后端模块设计)
- [9. LLM 代理层设计](#9-llm-代理层设计)
- [10. 安全设计](#10-安全设计)
- [11. 部署方案](#11-部署方案)
- [12. 测试方案](#12-测试方案)
- [13. 版本管理规范](#13-版本管理规范)
- [14. 开发阶段规划](#14-开发阶段规划)
- [15. UI/UX 设计规范](#15-uiux-设计规范)
- [16. 附录](#16-附录)

---

## 1. 项目概述

### 1.1 项目背景

Organizations need to provide AI coding assistance services to their developers, based on locally deployed LLM coding models. To effectively manage users, control resource usage, and ensure service quality, an upper-level management platform is needed, similar to developer consoles provided by AI service providers.

### 1.2 项目目标

- Provide a unified AI coding service entry point with user access management
- Implement a multi-level user management system (Super Admin / Department Manager / User)
- Proxy all LLM requests with Token usage statistics and quota control
- Provide API Key management for service access
- Provide visual usage statistics dashboards

### 1.3 项目命名

| 项目 | 名称 |
|------|------|
| 中文名 | CodeMind |
| 英文名 | CodeMind |
| 项目代号 | CodeMind |
| 仓库名 | CodeMind |

---

## 2. 需求分析

### 2.1 用户角色定义

系统定义三种用户角色，权限由高到低：

| 角色 | 英文标识 | 权限说明 |
|------|----------|----------|
| 超级管理员 | `super_admin` | 全系统最高权限，管理所有用户、部门、系统配置 |
| 部门经理 | `dept_manager` | 管理所属部门的用户，查看部门统计数据 |
| 普通用户 | `user` | 使用 AI 编码服务，管理自己的 API Key，查看个人用量 |

### 2.2 功能需求矩阵

#### 2.2.1 认证模块

| 功能 | 说明 | super_admin | dept_manager | user |
|------|------|:-----------:|:------------:|:----:|
| 用户登录 | 用户名 + 密码登录 | ✅ | ✅ | ✅ |
| 修改密码 | 修改自己的登录密码 | ✅ | ✅ | ✅ |
| 退出登录 | 清除会话，退出系统 | ✅ | ✅ | ✅ |

#### 2.2.2 用户管理模块

| 功能 | 说明 | super_admin | dept_manager | user |
|------|------|:-----------:|:------------:|:----:|
| 创建用户 | 创建新用户账号 | ✅（全部） | ✅（本部门） | ❌ |
| 编辑用户 | 修改用户信息 | ✅（全部） | ✅（本部门） | ❌ |
| 禁用/启用用户 | 切换用户账号状态 | ✅（全部） | ✅（本部门） | ❌ |
| 删除用户 | 删除用户账号（软删除） | ✅（全部） | ❌ | ❌ |
| 重置密码 | 重置用户登录密码 | ✅（全部） | ✅（本部门） | ❌ |
| 查看用户列表 | 列表展示用户信息 | ✅（全部） | ✅（本部门） | ❌ |
| 批量导入用户 | 通过 CSV 批量创建用户 | ✅ | ❌ | ❌ |
| 分配角色 | 给用户分配角色 | ✅ | ❌ | ❌ |
| 查看个人信息 | 查看自己的账号信息 | ✅ | ✅ | ✅ |
| 编辑个人信息 | 修改自己的昵称、头像等 | ✅ | ✅ | ✅ |

#### 2.2.3 部门管理模块

| 功能 | 说明 | super_admin | dept_manager | user |
|------|------|:-----------:|:------------:|:----:|
| 创建部门 | 创建新部门 | ✅ | ❌ | ❌ |
| 编辑部门 | 修改部门信息 | ✅ | ❌ | ❌ |
| 删除部门 | 删除部门（需先转移人员） | ✅ | ❌ | ❌ |
| 查看部门列表 | 列表展示所有部门 | ✅ | ✅（本部门） | ❌ |
| 设置部门经理 | 指定部门经理 | ✅ | ❌ | ❌ |
| 部门调动 | 将用户在部门间调动 | ✅ | ❌ | ❌ |

#### 2.2.4 API Key 管理模块

| 功能 | 说明 | super_admin | dept_manager | user |
|------|------|:-----------:|:------------:|:----:|
| 创建 API Key | 生成新的 API Key | ✅ | ✅ | ✅ |
| 查看 API Key 列表 | 列出自己的 Key | ✅ | ✅ | ✅ |
| 禁用/启用 Key | 切换 Key 状态 | ✅（全部） | ✅（本部门） | ✅（自己） |
| 删除 Key | 删除指定 Key | ✅（全部） | ✅（本部门） | ✅（自己） |
| 设置 Key 备注 | 给 Key 添加备注信息 | ✅ | ✅ | ✅ |
| 查看 Key 用量 | 查看某个 Key 的使用统计 | ✅ | ✅ | ✅ |

#### 2.2.5 用量统计模块

| 功能 | 说明 | super_admin | dept_manager | user |
|------|------|:-----------:|:------------:|:----:|
| 总览面板 | 系统整体用量概览 | ✅ | ✅（本部门） | ✅（个人） |
| 每日统计 | 按天查看 Token 用量 | ✅（全部） | ✅（本部门） | ✅（个人） |
| 每周统计 | 按周查看 Token 用量 | ✅（全部） | ✅（本部门） | ✅（个人） |
| 每月统计 | 按月查看 Token 用量 | ✅（全部） | ✅（本部门） | ✅（个人） |
| 用户用量排行 | 用量排行榜 | ✅ | ✅（本部门） | ❌ |
| 部门用量排行 | 部门用量对比 | ✅ | ❌ | ❌ |
| 导出报表 | 导出用量数据为 CSV/Excel | ✅ | ✅（本部门） | ✅（个人） |

#### 2.2.6 限额管理模块

| 功能 | 说明 | super_admin | dept_manager | user |
|------|------|:-----------:|:------------:|:----:|
| 设置全局默认限额 | 系统级默认 Token 限额 | ✅ | ❌ | ❌ |
| 设置部门限额 | 部门级 Token 限额 | ✅ | ❌ | ❌ |
| 设置用户限额 | 用户级 Token 限额 | ✅ | ✅（本部门） | ❌ |
| 设置并发限制 | 单用户最大并发请求数 | ✅ | ❌ | ❌ |
| 查看限额配置 | 查看当前限额设置 | ✅ | ✅（本部门） | ✅（个人） |
| 限额告警配置 | 设置用量告警阈值 | ✅ | ✅（本部门） | ❌ |

#### 2.2.7 系统管理模块

| 功能 | 说明 | super_admin | dept_manager | user |
|------|------|:-----------:|:------------:|:----:|
| LLM 服务配置 | 配置后端 LLM 服务地址 | ✅ | ❌ | ❌ |
| 模型管理 | 管理可用的 LLM 模型列表 | ✅ | ❌ | ❌ |
| 系统日志 | 查看系统操作日志 | ✅ | ❌ | ❌ |
| 公告管理 | 发布/编辑系统公告 | ✅ | ❌ | ❌ |

### 2.3 非功能需求

| 需求类型 | 描述 |
|----------|------|
| 性能 | API 响应时间 < 200ms（不含 LLM 请求），LLM 代理请求使用流式传输 |
| 并发 | 支持至少 500 个并发用户的 LLM 请求代理 |
| 可用性 | 系统可用性 > 99.5%，支持优雅重启 |
| 安全性 | 密码加密存储，JWT Token 鉴权，API Key 加密存储，支持 HTTPS |
| 兼容性 | LLM 代理接口兼容 OpenAI API 标准格式 |
| 可维护性 | 清晰的代码结构，完善的单元测试，覆盖率 > 80% |
| 可部署性 | Docker 容器化部署，支持 docker-compose 一键启动 |

---

## 3. 技术栈选型

### 3.1 前端技术栈

| 技术 | 版本 | 用途 |
|------|------|------|
| React | 18.x | UI 框架 |
| TypeScript | 5.x | 类型安全 |
| Vite | 6.x | 构建工具 |
| Ant Design | 5.x | UI 组件库（适合管理后台） |
| TailwindCSS | 3.x | 原子化 CSS（首页等定制页面） |
| React Router | 7.x | 路由管理 |
| Zustand | 5.x | 轻量级状态管理 |
| Axios | 1.x | HTTP 客户端 |
| ECharts | 5.x | 图表库（用量统计可视化） |
| dayjs | 1.x | 日期处理 |
| Vitest | 2.x | 单元测试 |
| React Testing Library | 16.x | 组件测试 |

**选型理由**:
- **React + TypeScript**: 主流前端方案，类型安全，生态成熟
- **Ant Design**: 专为企业级管理后台设计，组件丰富，开箱即用
- **TailwindCSS**: 首页等需要定制化设计的页面使用，灵活高效
- **Zustand**: 相比 Redux 更加轻量，API 简洁，适合中等复杂度状态管理
- **ECharts**: 功能强大的图表库，适合复杂的用量统计可视化

### 3.2 后端技术栈

| 技术 | 版本 | 用途 |
|------|------|------|
| Go | 1.23.x | 后端语言 |
| Gin | 1.10.x | Web 框架 |
| GORM | 1.25.x | ORM 框架 |
| golang-jwt | 5.x | JWT 认证 |
| go-redis | 9.x | Redis 客户端 |
| viper | 1.19.x | 配置管理 |
| zap | 1.27.x | 结构化日志 |
| swaggo/swag | 1.16.x | Swagger 文档生成 |
| testify | 1.9.x | 测试断言库 |
| golang-migrate | 4.x | 数据库迁移 |
| uuid | 1.6.x | UUID 生成 |
| crypto/bcrypt | stdlib | 密码加密 |

**选型理由**:
- **Go**: 高性能、原生并发支持、编译型语言。作为 LLM 请求代理层，Go 的 goroutine 模型天然适合处理大量并发的流式请求转发
- **Gin**: Go 生态最成熟的 Web 框架，性能优异，中间件丰富
- **GORM**: Go 最流行的 ORM，支持 PostgreSQL，自动迁移方便开发
- **Redis**: 用于 JWT 黑名单、限流计数器、并发控制、缓存

### 3.3 基础设施

| 技术 | 版本 | 用途 |
|------|------|------|
| PostgreSQL | 16.x | 主数据库 |
| Redis | 7.x | 缓存、限流、会话 |
| Nginx | 1.27.x | 前端静态文件服务 & 反向代理 |
| Docker | 27.x | 容器化 |
| Docker Compose | 2.x | 容器编排 |

### 3.4 技术架构图

```
┌─────────────────────────────────────────────────────────────┐
│                        Nginx (反向代理)                       │
│               前端静态资源 + API 请求转发                       │
└──────────────┬─────────────────────────┬────────────────────┘
               │                         │
               ▼                         ▼
┌──────────────────────┐   ┌──────────────────────────────────┐
│   Frontend (React)   │   │        Backend (Go/Gin)          │
│                      │   │                                  │
│  • 首页/登录          │   │  ┌─────────┐  ┌──────────────┐  │
│  • 管理后台           │   │  │  Auth   │  │  User Mgmt   │  │
│  • 用量统计           │   │  │ Module  │  │   Module     │  │
│  • API Key 管理      │   │  └─────────┘  └──────────────┘  │
│  • 个人中心           │   │  ┌─────────┐  ┌──────────────┐  │
│                      │   │  │  Stats  │  │  Rate Limit  │  │
│                      │   │  │ Module  │  │   Module     │  │
│                      │   │  └─────────┘  └──────────────┘  │
│                      │   │  ┌──────────────────────────┐   │
│                      │   │  │     LLM Proxy Module     │   │
│                      │   │  │  (Request/Response 转发)  │   │
│                      │   │  └────────────┬─────────────┘   │
└──────────────────────┘   └───────┬───────┼─────────────────┘
                                   │       │
                          ┌────────┴──┐    │
                          │           │    │
                     ┌────▼────┐ ┌────▼──┐ │
                     │PostgreSQL│ │ Redis │ │
                     │  (数据)  │ │(缓存) │ │
                     └─────────┘ └───────┘ │
                                           │
                                    ┌──────▼──────┐
                                    │  LLM Server │
                                    │ (本地部署)   │
                                    └─────────────┘
```

---

## 4. 系统架构设计

### 4.1 后端分层架构

采用经典的分层架构，各层职责清晰：

```
┌─────────────────────────────────┐
│         Router Layer            │  路由层：路由定义、中间件挂载
├─────────────────────────────────┤
│       Middleware Layer           │  中间件层：认证、限流、日志、CORS
├─────────────────────────────────┤
│        Handler Layer            │  控制器层：请求参数校验、响应封装
├─────────────────────────────────┤
│        Service Layer            │  业务层：核心业务逻辑
├─────────────────────────────────┤
│       Repository Layer          │  数据访问层：数据库 CRUD 操作
├─────────────────────────────────┤
│         Model Layer             │  模型层：数据结构定义
└─────────────────────────────────┘
```

**各层职责说明**：

- **Router**: 定义 URL 路由映射，组织中间件链
- **Middleware**: 横切关注点（认证、限流、日志、CORS、错误恢复）
- **Handler**: 接收 HTTP 请求，校验参数，调用 Service，封装响应
- **Service**: 实现业务逻辑，调用 Repository，处理业务规则
- **Repository**: 封装数据库操作，提供数据访问接口
- **Model**: 定义数据库模型与数据传输对象（DTO）

### 4.2 前端架构

```
┌──────────────────────────────────────────────────┐
│                   Pages Layer                    │  页面组件
├──────────────────────────────────────────────────┤
│                Components Layer                  │  通用组件、布局组件
├──────────────────────────────────────────────────┤
│    Hooks Layer    │     Store Layer              │  自定义 Hooks、状态管理
├──────────────────────────────────────────────────┤
│               Services Layer                     │  API 请求封装
├──────────────────────────────────────────────────┤
│     Utils Layer   │     Types Layer              │  工具函数、类型定义
└──────────────────────────────────────────────────┘
```

### 4.3 请求流转流程

#### 4.3.1 普通 API 请求（如用户管理）

```
Client → Nginx → Backend API → Middleware Chain → Handler → Service → Repository → PostgreSQL
```

#### 4.3.2 LLM 代理请求

```
Client (IDE/Editor)
    │
    ▼ (携带 API Key)
  Nginx
    │
    ▼
  Backend API
    │
    ├── 1. API Key 验证 (Middleware)
    ├── 2. 用户状态检查 (Middleware)
    ├── 3. 限额检查 - Token 余额 (Middleware)
    ├── 4. 并发数检查 (Middleware, Redis)
    │
    ▼ (通过所有检查)
  LLM Proxy Handler
    │
    ├── 5. 记录并发计数 +1 (Redis INCR)
    ├── 6. 转发请求至 LLM Server
    ├── 7. 流式传输响应给 Client (SSE)
    ├── 8. 统计 Token 用量 (从 LLM 响应中解析)
    ├── 9. 记录用量到数据库
    └── 10. 并发计数 -1 (Redis DECR)
```

### 4.4 LLM 代理层详细设计

LLM 代理层是本系统的核心模块，负责所有 AI 编码请求的转发与管控。

#### 4.4.1 兼容 OpenAI API 格式

代理层对外暴露的接口完全兼容 OpenAI API 标准格式，便于各类 IDE 插件（如 Continue、Cline、Cursor 等）直接接入：

| 代理接口 | 对应 OpenAI 接口 | 说明 |
|----------|-----------------|------|
| `POST /api/openai/v1/chat/completions` | Chat Completions | 对话补全（主要接口） |
| `POST /api/openai/v1/completions` | Completions | 文本补全 |
| `GET /api/openai/v1/models` | List Models | 获取可用模型列表 |

#### 4.4.2 请求转发流程

```go
// Pseudocode for LLM proxy flow
func ProxyLLMRequest(ctx) {
    // 1. Extract API Key from Authorization header
    apiKey := extractAPIKey(ctx)

    // 2. Validate API Key and get user info (cached in Redis)
    user, err := validateAPIKey(apiKey)

    // 3. Check user rate limits
    if !checkTokenQuota(user) {
        return Error(429, "Token quota exceeded")
    }

    // 4. Check concurrent request limit
    if !acquireConcurrencySlot(user) {
        return Error(429, "Too many concurrent requests")
    }
    defer releaseConcurrencySlot(user)

    // 5. Forward request to LLM server
    // 6. Stream response back to client
    // 7. Count tokens from response
    // 8. Record usage
}
```

#### 4.4.3 Token 计量方式

- **输入 Token**: 从请求 body 中解析 messages，使用 tiktoken 或从 LLM 响应 usage 字段获取
- **输出 Token**: 从 LLM 响应的 `usage.completion_tokens` 字段获取
- **总 Token**: `prompt_tokens + completion_tokens`
- 若 LLM 响应不包含 usage 信息（如流式传输中），在流结束后从最后一个 chunk 中获取，或根据内容估算

#### 4.4.4 并发控制

使用 Redis 实现并发控制：

```
Key: codemind:concurrency:{user_id}
Type: String (Counter)
TTL: 300s (安全超时，防止异常情况下计数器不释放)
```

- 请求开始：`INCR key`，若值 > 限制，`DECR key` 并拒绝
- 请求结束：`DECR key`
- 使用 `EXPIRE` 设置 TTL 防止计数泄漏

---

## 5. 数据库设计

### 5.1 ER 关系图

```
┌──────────────┐     ┌───────────────┐     ┌──────────────┐
│  departments │     │     users     │     │   api_keys   │
├──────────────┤     ├───────────────┤     ├──────────────┤
│ id           │◄────│ department_id │     │ id           │
│ name         │     │ id            │◄────│ user_id      │
│ manager_id   │─────│ username      │     │ key_hash     │
│ ...          │     │ role          │     │ key_prefix   │
└──────────────┘     │ ...           │     │ ...          │
                     └───────┬───────┘     └──────┬───────┘
                             │                    │
                     ┌───────▼───────┐     ┌──────▼───────┐
                     │ token_usage   │     │ request_logs │
                     ├───────────────┤     ├──────────────┤
                     │ id            │     │ id           │
                     │ user_id       │     │ api_key_id   │
                     │ api_key_id    │     │ user_id      │
                     │ prompt_tokens │     │ model        │
                     │ compl_tokens  │     │ status_code  │
                     │ ...           │     │ ...          │
                     └───────────────┘     └──────────────┘

┌──────────────┐     ┌───────────────┐     ┌──────────────┐
│ rate_limits  │     │ announcements │     │system_configs│
├──────────────┤     ├───────────────┤     ├──────────────┤
│ id           │     │ id            │     │ id           │
│ target_type  │     │ title         │     │ key          │
│ target_id    │     │ content       │     │ value        │
│ ...          │     │ ...           │     │ ...          │
└──────────────┘     └───────────────┘     └──────────────┘
```

### 5.2 数据表详细定义

#### 5.2.1 departments（部门表）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGSERIAL | PRIMARY KEY | 自增主键 |
| name | VARCHAR(100) | NOT NULL, UNIQUE | 部门名称 |
| description | TEXT | | 部门描述 |
| manager_id | BIGINT | REFERENCES users(id) | 部门经理 ID |
| parent_id | BIGINT | REFERENCES departments(id) | 上级部门 ID（支持树形结构） |
| status | SMALLINT | NOT NULL DEFAULT 1 | 状态：1-启用 0-禁用 |
| created_at | TIMESTAMPTZ | NOT NULL DEFAULT NOW() | 创建时间 |
| updated_at | TIMESTAMPTZ | NOT NULL DEFAULT NOW() | 更新时间 |

#### 5.2.2 users（用户表）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGSERIAL | PRIMARY KEY | 自增主键 |
| username | VARCHAR(50) | NOT NULL, UNIQUE | 登录用户名 |
| password_hash | VARCHAR(255) | NOT NULL | bcrypt 加密后的密码 |
| display_name | VARCHAR(100) | NOT NULL | 显示名称 |
| email | VARCHAR(255) | UNIQUE | 邮箱 |
| phone | VARCHAR(20) | | 手机号 |
| avatar_url | VARCHAR(500) | | 头像 URL |
| role | VARCHAR(20) | NOT NULL DEFAULT 'user' | 角色：super_admin / dept_manager / user |
| department_id | BIGINT | REFERENCES departments(id) | 所属部门 |
| status | SMALLINT | NOT NULL DEFAULT 1 | 状态：1-启用 0-禁用 |
| last_login_at | TIMESTAMPTZ | | 最后登录时间 |
| last_login_ip | VARCHAR(45) | | 最后登录 IP |
| created_at | TIMESTAMPTZ | NOT NULL DEFAULT NOW() | 创建时间 |
| updated_at | TIMESTAMPTZ | NOT NULL DEFAULT NOW() | 更新时间 |
| deleted_at | TIMESTAMPTZ | | 软删除时间 |

**索引**：
- `idx_users_username` ON (username)
- `idx_users_department_id` ON (department_id)
- `idx_users_role` ON (role)
- `idx_users_status` ON (status)

#### 5.2.3 api_keys（API 密钥表）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGSERIAL | PRIMARY KEY | 自增主键 |
| user_id | BIGINT | NOT NULL, REFERENCES users(id) | 所属用户 |
| name | VARCHAR(100) | NOT NULL | Key 名称/备注 |
| key_prefix | VARCHAR(10) | NOT NULL | Key 前缀（如 `cm-abc`），用于展示 |
| key_hash | VARCHAR(255) | NOT NULL, UNIQUE | Key 的 SHA-256 哈希值 |
| status | SMALLINT | NOT NULL DEFAULT 1 | 状态：1-启用 0-禁用 |
| last_used_at | TIMESTAMPTZ | | 最后使用时间 |
| expires_at | TIMESTAMPTZ | | 过期时间（NULL 表示永不过期） |
| created_at | TIMESTAMPTZ | NOT NULL DEFAULT NOW() | 创建时间 |

**索引**：
- `idx_api_keys_key_hash` UNIQUE ON (key_hash)
- `idx_api_keys_user_id` ON (user_id)
- `idx_api_keys_key_prefix` ON (key_prefix)

**设计说明**：
- API Key 生成格式：`cm-{32位随机字符}`，如 `cm-a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6`
- Key 仅在创建时返回一次完整值，之后仅存储哈希值
- `key_prefix` 存储前 8 位，便于用户识别

#### 5.2.4 token_usage（Token 用量记录表）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGSERIAL | PRIMARY KEY | 自增主键 |
| user_id | BIGINT | NOT NULL, REFERENCES users(id) | 用户 ID |
| api_key_id | BIGINT | NOT NULL, REFERENCES api_keys(id) | 使用的 API Key |
| model | VARCHAR(100) | NOT NULL | 使用的模型名称 |
| prompt_tokens | INTEGER | NOT NULL DEFAULT 0 | 输入 Token 数 |
| completion_tokens | INTEGER | NOT NULL DEFAULT 0 | 输出 Token 数 |
| total_tokens | INTEGER | NOT NULL DEFAULT 0 | 总 Token 数 |
| request_type | VARCHAR(30) | NOT NULL | 请求类型：chat_completion / completion |
| duration_ms | INTEGER | | 请求耗时（毫秒） |
| created_at | TIMESTAMPTZ | NOT NULL DEFAULT NOW() | 记录时间 |

**索引**：
- `idx_token_usage_user_id_created_at` ON (user_id, created_at)
- `idx_token_usage_api_key_id` ON (api_key_id)
- `idx_token_usage_created_at` ON (created_at)
- `idx_token_usage_model` ON (model)

**设计说明**：
- 每次 LLM 请求完成后写入一条记录
- 统计查询通过 `GROUP BY` + `DATE_TRUNC` 实现日/周/月聚合
- 考虑数据量增长，后续可引入按月分区

#### 5.2.5 token_usage_daily（Token 用量日汇总表）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGSERIAL | PRIMARY KEY | 自增主键 |
| user_id | BIGINT | NOT NULL, REFERENCES users(id) | 用户 ID |
| usage_date | DATE | NOT NULL | 统计日期 |
| prompt_tokens | BIGINT | NOT NULL DEFAULT 0 | 当日输入 Token 总量 |
| completion_tokens | BIGINT | NOT NULL DEFAULT 0 | 当日输出 Token 总量 |
| total_tokens | BIGINT | NOT NULL DEFAULT 0 | 当日 Token 总量 |
| request_count | INTEGER | NOT NULL DEFAULT 0 | 当日请求次数 |
| created_at | TIMESTAMPTZ | NOT NULL DEFAULT NOW() | 创建时间 |
| updated_at | TIMESTAMPTZ | NOT NULL DEFAULT NOW() | 更新时间 |

**索引**：
- `idx_token_usage_daily_user_date` UNIQUE ON (user_id, usage_date)
- `idx_token_usage_daily_date` ON (usage_date)

**设计说明**：
- 每次请求完成后同步更新（`ON CONFLICT DO UPDATE`）
- 用于加速日/周/月统计查询，避免从明细表实时聚合

#### 5.2.6 rate_limits（限额配置表）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGSERIAL | PRIMARY KEY | 自增主键 |
| target_type | VARCHAR(20) | NOT NULL | 限额目标类型：global / department / user |
| target_id | BIGINT | NOT NULL DEFAULT 0 | 目标 ID（global 时为 0） |
| period | VARCHAR(20) | NOT NULL | 限额周期：daily / weekly / monthly |
| max_tokens | BIGINT | NOT NULL | 最大 Token 数 |
| max_requests | INTEGER | NOT NULL DEFAULT 0 | 最大请求数（0 表示不限制） |
| max_concurrency | INTEGER | NOT NULL DEFAULT 5 | 最大并发请求数 |
| alert_threshold | SMALLINT | NOT NULL DEFAULT 80 | 告警阈值百分比（达到该比例时告警） |
| status | SMALLINT | NOT NULL DEFAULT 1 | 状态：1-启用 0-禁用 |
| created_at | TIMESTAMPTZ | NOT NULL DEFAULT NOW() | 创建时间 |
| updated_at | TIMESTAMPTZ | NOT NULL DEFAULT NOW() | 更新时间 |

**索引**：
- `idx_rate_limits_target` UNIQUE ON (target_type, target_id, period)

**限额优先级**（从高到低）：
1. 用户级限额
2. 部门级限额
3. 全局默认限额

#### 5.2.7 request_logs（请求日志表）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGSERIAL | PRIMARY KEY | 自增主键 |
| user_id | BIGINT | NOT NULL | 用户 ID |
| api_key_id | BIGINT | NOT NULL | API Key ID |
| request_type | VARCHAR(30) | NOT NULL | 请求类型 |
| model | VARCHAR(100) | | 请求模型 |
| status_code | INTEGER | NOT NULL | HTTP 状态码 |
| error_message | TEXT | | 错误信息 |
| client_ip | VARCHAR(45) | | 客户端 IP |
| user_agent | VARCHAR(500) | | User Agent |
| duration_ms | INTEGER | | 耗时（毫秒） |
| created_at | TIMESTAMPTZ | NOT NULL DEFAULT NOW() | 请求时间 |

**索引**：
- `idx_request_logs_user_id_created_at` ON (user_id, created_at)
- `idx_request_logs_created_at` ON (created_at)

#### 5.2.8 announcements（公告表）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGSERIAL | PRIMARY KEY | 自增主键 |
| title | VARCHAR(200) | NOT NULL | 公告标题 |
| content | TEXT | NOT NULL | 公告内容（Markdown） |
| author_id | BIGINT | NOT NULL, REFERENCES users(id) | 发布者 |
| status | SMALLINT | NOT NULL DEFAULT 1 | 状态：1-已发布 0-草稿 |
| pinned | BOOLEAN | NOT NULL DEFAULT FALSE | 是否置顶 |
| created_at | TIMESTAMPTZ | NOT NULL DEFAULT NOW() | 创建时间 |
| updated_at | TIMESTAMPTZ | NOT NULL DEFAULT NOW() | 更新时间 |

#### 5.2.9 system_configs（系统配置表）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGSERIAL | PRIMARY KEY | 自增主键 |
| config_key | VARCHAR(100) | NOT NULL, UNIQUE | 配置键 |
| config_value | TEXT | NOT NULL | 配置值（JSON 格式） |
| description | VARCHAR(500) | | 配置说明 |
| updated_at | TIMESTAMPTZ | NOT NULL DEFAULT NOW() | 更新时间 |

**预置配置项**：
- `llm.base_url`: LLM 服务基础 URL
- `llm.api_key`: LLM 服务 API Key（如果需要）
- `llm.models`: 可用模型列表（JSON 数组）
- `llm.default_model`: 默认模型
- `system.max_keys_per_user`: 每用户最大 Key 数量
- `system.default_concurrency`: 默认并发限制

#### 5.2.10 audit_logs（审计日志表）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGSERIAL | PRIMARY KEY | 自增主键 |
| operator_id | BIGINT | NOT NULL | 操作者 ID |
| action | VARCHAR(50) | NOT NULL | 操作类型：create_user / delete_user / update_limit 等 |
| target_type | VARCHAR(50) | NOT NULL | 操作目标类型：user / department / api_key 等 |
| target_id | BIGINT | | 操作目标 ID |
| detail | JSONB | | 操作详情（变更前后的值） |
| client_ip | VARCHAR(45) | | 操作者 IP |
| created_at | TIMESTAMPTZ | NOT NULL DEFAULT NOW() | 操作时间 |

**索引**：
- `idx_audit_logs_operator_id` ON (operator_id)
- `idx_audit_logs_action` ON (action)
- `idx_audit_logs_created_at` ON (created_at)

---

## 6. API 接口设计

### 6.1 接口规范

#### 6.1.1 基础 URL

```
管理平台 API：  /api/v1/*
LLM 代理 API：  OpenAI 兼容 `/api/openai/v1/*`；Anthropic `/api/anthropic/*`
```

#### 6.1.2 认证方式

| 接口类型 | 认证方式 | Header |
|----------|----------|--------|
| 管理平台 API | JWT Bearer Token | `Authorization: Bearer <jwt_token>` |
| LLM 代理 API | API Key | `Authorization: Bearer <api_key>` |

#### 6.1.3 统一响应格式

**成功响应**：
```json
{
  "code": 0,
  "message": "success",
  "data": { ... }
}
```

**分页响应**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [ ... ],
    "pagination": {
      "page": 1,
      "page_size": 20,
      "total": 100,
      "total_pages": 5
    }
  }
}
```

**错误响应**：
```json
{
  "code": 40001,
  "message": "Invalid username or password",
  "data": null
}
```

#### 6.1.4 错误码定义

| 错误码区间 | 说明 |
|-----------|------|
| 0 | 成功 |
| 40001-40099 | 认证相关错误 |
| 40100-40199 | 权限相关错误 |
| 40200-40299 | 参数校验错误 |
| 40300-40399 | 业务逻辑错误 |
| 42900-42999 | 限流/限额相关错误 |
| 50000-50099 | 系统内部错误 |

**详细错误码**：

| 错误码 | 说明 |
|--------|------|
| 40001 | 用户名或密码错误 |
| 40002 | Token 已过期 |
| 40003 | Token 无效 |
| 40004 | 账号已被禁用 |
| 40005 | API Key 无效 |
| 40006 | API Key 已过期 |
| 40007 | API Key 已被禁用 |
| 40101 | 无权访问该资源 |
| 40102 | 无权操作该用户 |
| 40103 | 无权操作该部门 |
| 40201 | 请求参数错误 |
| 40202 | 必填参数缺失 |
| 40301 | 用户名已存在 |
| 40302 | 邮箱已被使用 |
| 40303 | 部门不存在 |
| 40304 | API Key 数量已达上限 |
| 40305 | 部门下还有用户，无法删除 |
| 42901 | Token 用量已达限额 |
| 42902 | 并发请求数已达上限 |
| 42903 | 请求频率过快 |
| 50001 | 系统内部错误 |
| 50002 | LLM 服务不可用 |
| 50003 | 数据库错误 |

### 6.2 认证接口

#### POST /api/v1/auth/login

用户登录。

**请求体**：
```json
{
  "username": "zhangsan",
  "password": "password123"
}
```

**响应**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "expires_at": "2026-02-16T14:30:00Z",
    "user": {
      "id": 1,
      "username": "zhangsan",
      "display_name": "张三",
      "role": "user",
      "department": {
        "id": 1,
        "name": "研发部"
      }
    }
  }
}
```

#### POST /api/v1/auth/logout

退出登录（将当前 JWT 加入黑名单）。

**请求头**：`Authorization: Bearer <jwt_token>`

**响应**：
```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

#### GET /api/v1/auth/profile

获取当前登录用户信息。

**请求头**：`Authorization: Bearer <jwt_token>`

**响应**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "username": "zhangsan",
    "display_name": "张三",
    "email": "zhangsan@example.com",
    "phone": "13800138000",
    "avatar_url": "/avatars/1.jpg",
    "role": "user",
    "department": {
      "id": 1,
      "name": "研发部"
    },
    "status": 1,
    "last_login_at": "2026-02-15T10:00:00Z",
    "created_at": "2026-01-01T00:00:00Z"
  }
}
```

#### PUT /api/v1/auth/profile

更新当前用户个人信息。

**请求体**：
```json
{
  "display_name": "张三丰",
  "email": "zsf@example.com",
  "phone": "13900139000"
}
```

#### PUT /api/v1/auth/password

修改密码。

**请求体**：
```json
{
  "old_password": "oldpass123",
  "new_password": "newpass456"
}
```

### 6.3 用户管理接口

#### GET /api/v1/users

获取用户列表（管理员、部门经理可用）。

**查询参数**：
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认 1 |
| page_size | int | 否 | 每页数量，默认 20，最大 100 |
| keyword | string | 否 | 搜索关键词（用户名/姓名/邮箱） |
| department_id | int | 否 | 按部门筛选 |
| role | string | 否 | 按角色筛选 |
| status | int | 否 | 按状态筛选 |

#### POST /api/v1/users

创建用户。

**请求体**：
```json
{
  "username": "lisi",
  "password": "initialPass123",
  "display_name": "李四",
  "email": "lisi@example.com",
  "phone": "13700137000",
  "role": "user",
  "department_id": 2
}
```

#### GET /api/v1/users/:id

获取用户详情。

#### PUT /api/v1/users/:id

更新用户信息。

**请求体**：
```json
{
  "display_name": "李四",
  "email": "lisi@example.com",
  "phone": "13700137000",
  "role": "dept_manager",
  "department_id": 2,
  "status": 1
}
```

#### DELETE /api/v1/users/:id

删除用户（软删除，仅超级管理员可用）。

#### PUT /api/v1/users/:id/status

切换用户状态（启用/禁用）。

**请求体**：
```json
{
  "status": 0
}
```

#### PUT /api/v1/users/:id/reset-password

重置用户密码。

**请求体**：
```json
{
  "new_password": "resetPass123"
}
```

#### POST /api/v1/users/import

批量导入用户（CSV 文件上传，仅超级管理员）。

**请求**：`Content-Type: multipart/form-data`
- `file`: CSV 文件

**CSV 格式**：
```csv
username,display_name,email,phone,department_id,role
zhangsan,张三,zhangsan@example.com,13800138000,1,user
lisi,李四,lisi@example.com,13700137000,2,user
```

### 6.4 部门管理接口

#### GET /api/v1/departments

获取部门列表（树形结构）。

**响应**：
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "name": "研发部",
      "description": "产品研发部门",
      "manager": {
        "id": 10,
        "display_name": "王经理"
      },
      "user_count": 15,
      "children": [
        {
          "id": 3,
          "name": "前端组",
          "manager": null,
          "user_count": 5,
          "children": []
        }
      ]
    }
  ]
}
```

#### POST /api/v1/departments

创建部门。

**请求体**：
```json
{
  "name": "AI研究院",
  "description": "AI技术研究部门",
  "parent_id": null,
  "manager_id": 5
}
```

#### PUT /api/v1/departments/:id

更新部门信息。

#### DELETE /api/v1/departments/:id

删除部门（需确保部门下无用户）。

#### GET /api/v1/departments/:id/users

获取部门下的用户列表。

### 6.5 API Key 管理接口

#### GET /api/v1/keys

获取当前用户的 API Key 列表。

**响应**：
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "name": "IDE 插件用",
      "key_prefix": "cm-a1b2c3",
      "status": 1,
      "last_used_at": "2026-02-15T10:30:00Z",
      "expires_at": null,
      "created_at": "2026-02-01T00:00:00Z"
    }
  ]
}
```

#### POST /api/v1/keys

创建新的 API Key。

**请求体**：
```json
{
  "name": "VSCode Cline 插件",
  "expires_at": "2026-12-31T23:59:59Z"
}
```

**响应**（仅此次返回完整 Key）：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 2,
    "name": "VSCode Cline 插件",
    "key": "cm-a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6",
    "key_prefix": "cm-a1b2c3",
    "expires_at": "2026-12-31T23:59:59Z",
    "created_at": "2026-02-15T14:00:00Z"
  }
}
```

#### DELETE /api/v1/keys/:id

删除 API Key。

#### PUT /api/v1/keys/:id/status

切换 Key 状态。

**请求体**：
```json
{
  "status": 0
}
```

#### GET /api/v1/keys/:id/usage

获取某个 Key 的用量统计。

**查询参数**：
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| period | string | 否 | 统计周期：daily / weekly / monthly，默认 daily |
| start_date | string | 否 | 开始日期，默认近 30 天 |
| end_date | string | 否 | 结束日期 |

### 6.6 用量统计接口

#### GET /api/v1/stats/overview

获取用量总览。

**响应**（管理员视角）：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "today": {
      "total_tokens": 1500000,
      "total_requests": 320,
      "active_users": 45
    },
    "this_month": {
      "total_tokens": 35000000,
      "total_requests": 8500,
      "active_users": 120
    },
    "total_users": 200,
    "total_departments": 10,
    "total_api_keys": 350,
    "system_status": "healthy"
  }
}
```

#### GET /api/v1/stats/usage

获取用量统计数据。

**查询参数**：
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| period | string | 是 | 统计周期：daily / weekly / monthly |
| start_date | string | 否 | 开始日期 |
| end_date | string | 否 | 结束日期 |
| user_id | int | 否 | 用户 ID（管理员可查其他用户） |
| department_id | int | 否 | 部门 ID（管理员可查部门数据） |

**响应**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "period": "daily",
    "items": [
      {
        "date": "2026-02-15",
        "prompt_tokens": 500000,
        "completion_tokens": 200000,
        "total_tokens": 700000,
        "request_count": 150
      },
      {
        "date": "2026-02-14",
        "prompt_tokens": 450000,
        "completion_tokens": 180000,
        "total_tokens": 630000,
        "request_count": 130
      }
    ]
  }
}
```

#### GET /api/v1/stats/ranking

获取用量排行榜。

**查询参数**：
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| type | string | 是 | 排行类型：user / department |
| period | string | 是 | 统计周期：daily / weekly / monthly |
| limit | int | 否 | 返回条数，默认 10 |

### 6.7 限额管理接口

#### GET /api/v1/limits

获取限额配置列表（管理员）。

**查询参数**：
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| target_type | string | 否 | 目标类型：global / department / user |
| target_id | int | 否 | 目标 ID |

#### PUT /api/v1/limits

创建或更新限额配置。

**请求体**：
```json
{
  "target_type": "user",
  "target_id": 5,
  "period": "monthly",
  "max_tokens": 5000000,
  "max_requests": 1000,
  "max_concurrency": 3,
  "alert_threshold": 80
}
```

#### GET /api/v1/limits/my

获取当前用户的限额信息及用量。

**响应**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "limits": {
      "daily": {
        "max_tokens": 500000,
        "used_tokens": 150000,
        "remaining_tokens": 350000,
        "usage_percent": 30
      },
      "monthly": {
        "max_tokens": 5000000,
        "used_tokens": 1200000,
        "remaining_tokens": 3800000,
        "usage_percent": 24
      }
    },
    "concurrency": {
      "max": 5,
      "current": 1
    }
  }
}
```

#### DELETE /api/v1/limits/:id

删除限额配置（恢复为上级默认值）。

### 6.8 系统管理接口

#### GET /api/v1/system/configs

获取系统配置（管理员）。

#### PUT /api/v1/system/configs

更新系统配置。

**请求体**：
```json
{
  "configs": [
    {
      "key": "llm.base_url",
      "value": "http://llm-server:8080"
    },
    {
      "key": "llm.models",
      "value": "[\"deepseek-coder-v2\", \"codellama-34b\"]"
    }
  ]
}
```

#### GET /api/v1/system/models

获取可用模型列表。

#### GET /api/v1/system/audit-logs

获取审计日志（管理员）。

**查询参数**：
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码 |
| page_size | int | 否 | 每页数量 |
| action | string | 否 | 操作类型筛选 |
| operator_id | int | 否 | 操作者 ID |
| start_date | string | 否 | 开始日期 |
| end_date | string | 否 | 结束日期 |

#### GET /api/v1/system/announcements

获取公告列表。

#### POST /api/v1/system/announcements

创建公告（管理员）。

#### PUT /api/v1/system/announcements/:id

更新公告（管理员）。

#### DELETE /api/v1/system/announcements/:id

删除公告（管理员）。

### 6.9 LLM 代理接口

以下接口兼容 OpenAI API 标准格式，客户端使用 API Key 进行认证。

#### POST /api/openai/v1/chat/completions

对话补全接口（主要的 AI 编码接口）。

**请求头**：
```
Authorization: Bearer cm-a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6
Content-Type: application/json
```

**请求体**（兼容 OpenAI 格式）：
```json
{
  "model": "deepseek-coder-v2",
  "messages": [
    {
      "role": "system",
      "content": "You are a helpful coding assistant."
    },
    {
      "role": "user",
      "content": "Write a Python function to sort a list."
    }
  ],
  "stream": true,
  "temperature": 0.7,
  "max_tokens": 2048
}
```

**响应**（非流式）：
```json
{
  "id": "chatcmpl-xxx",
  "object": "chat.completion",
  "created": 1708000000,
  "model": "deepseek-coder-v2",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "Here's a Python function..."
      },
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 50,
    "completion_tokens": 150,
    "total_tokens": 200
  }
}
```

**响应**（流式 SSE）：
```
data: {"id":"chatcmpl-xxx","object":"chat.completion.chunk","choices":[{"delta":{"content":"Here"},"index":0}]}

data: {"id":"chatcmpl-xxx","object":"chat.completion.chunk","choices":[{"delta":{"content":"'s"},"index":0}]}

...

data: {"id":"chatcmpl-xxx","object":"chat.completion.chunk","choices":[{"delta":{},"index":0,"finish_reason":"stop"}],"usage":{"prompt_tokens":50,"completion_tokens":150,"total_tokens":200}}

data: [DONE]
```

#### POST /api/openai/v1/completions

文本补全接口。

**请求体**：
```json
{
  "model": "deepseek-coder-v2",
  "prompt": "def fibonacci(n):",
  "max_tokens": 256,
  "stream": true
}
```

#### GET /api/openai/v1/models

获取可用模型列表。

**响应**：
```json
{
  "object": "list",
  "data": [
    {
      "id": "deepseek-coder-v2",
      "object": "model",
      "created": 1708000000,
      "owned_by": "local"
    }
  ]
}
```

---

## 7. 前端模块设计

### 7.1 页面结构

```
/                           # 首页（Landing Page）
/login                      # 登录页
/dashboard                  # 仪表盘（登录后首页）
/dashboard/usage            # 用量统计
/dashboard/keys             # API Key 管理
/dashboard/profile          # 个人中心
/admin                      # 管理后台（需管理员权限）
/admin/users                # 用户管理
/admin/departments          # 部门管理
/admin/limits               # 限额管理
/admin/models               # 模型管理
/admin/announcements        # 公告管理
/admin/audit-logs           # 审计日志
/admin/settings             # 系统设置
```

### 7.2 组件设计

#### 7.2.1 布局组件

| 组件 | 文件路径 | 说明 |
|------|----------|------|
| LandingLayout | `components/layout/LandingLayout.tsx` | 首页布局（全屏、沉浸式） |
| DashboardLayout | `components/layout/DashboardLayout.tsx` | 仪表盘布局（侧边栏 + 顶栏 + 内容区） |
| AdminLayout | `components/layout/AdminLayout.tsx` | 管理后台布局（继承 DashboardLayout） |
| Sidebar | `components/layout/Sidebar.tsx` | 侧边导航栏 |
| Header | `components/layout/Header.tsx` | 顶部栏（用户信息、通知） |

#### 7.2.2 通用组件

| 组件 | 文件路径 | 说明 |
|------|----------|------|
| Logo | `components/common/Logo.tsx` | Logo 组件 |
| LoadingSpinner | `components/common/LoadingSpinner.tsx` | 加载动画 |
| ErrorBoundary | `components/common/ErrorBoundary.tsx` | 错误边界 |
| PermissionGuard | `components/common/PermissionGuard.tsx` | 权限守卫组件 |
| EmptyState | `components/common/EmptyState.tsx` | 空状态占位 |
| PageHeader | `components/common/PageHeader.tsx` | 页面标题栏 |
| CopyButton | `components/common/CopyButton.tsx` | 复制到剪贴板按钮 |
| ConfirmModal | `components/common/ConfirmModal.tsx` | 确认弹窗 |

#### 7.2.3 图表组件

| 组件 | 文件路径 | 说明 |
|------|----------|------|
| TokenUsageChart | `components/charts/TokenUsageChart.tsx` | Token 用量折线图 |
| UsageBarChart | `components/charts/UsageBarChart.tsx` | 用量柱状图 |
| UsagePieChart | `components/charts/UsagePieChart.tsx` | 用量占比饼图 |
| RankingList | `components/charts/RankingList.tsx` | 用量排行榜 |
| QuotaGauge | `components/charts/QuotaGauge.tsx` | 配额仪表盘 |

### 7.3 状态管理设计

使用 Zustand 管理全局状态：

```typescript
// store/authStore.ts - Authentication state
interface AuthState {
  token: string | null;
  user: User | null;
  isAuthenticated: boolean;
  login: (credentials: LoginParams) => Promise<void>;
  logout: () => void;
  fetchProfile: () => Promise<void>;
}

// store/appStore.ts - Application global state
interface AppState {
  sidebarCollapsed: boolean;
  announcements: Announcement[];
  toggleSidebar: () => void;
  fetchAnnouncements: () => Promise<void>;
}
```

### 7.4 API Service 层设计

```typescript
// services/request.ts - Axios instance with interceptors
// services/authService.ts - Authentication APIs
// services/userService.ts - User management APIs
// services/departmentService.ts - Department management APIs
// services/keyService.ts - API Key management APIs
// services/statsService.ts - Statistics APIs
// services/limitService.ts - Rate limit APIs
// services/systemService.ts - System configuration APIs
```

### 7.5 路由守卫

```typescript
// router/guards.ts
// - AuthGuard: Redirect to login if not authenticated
// - AdminGuard: Redirect to dashboard if not admin/manager
// - GuestGuard: Redirect to dashboard if already authenticated
```

---

## 8. 后端模块设计

### 8.1 项目结构

```
backend/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go            # Configuration management
│   ├── middleware/
│   │   ├── auth.go              # JWT authentication middleware
│   │   ├── cors.go              # CORS middleware
│   │   ├── ratelimit.go         # Rate limiting middleware
│   │   ├── logger.go            # Request logging middleware
│   │   ├── recovery.go          # Panic recovery middleware
│   │   └── apikey.go            # API Key authentication middleware
│   ├── handler/
│   │   ├── auth.go              # Authentication handlers
│   │   ├── user.go              # User management handlers
│   │   ├── department.go        # Department management handlers
│   │   ├── apikey.go            # API Key management handlers
│   │   ├── stats.go             # Statistics handlers
│   │   ├── limit.go             # Rate limit handlers
│   │   ├── llm_proxy.go         # LLM proxy handlers
│   │   └── system.go            # System management handlers
│   ├── service/
│   │   ├── auth.go              # Authentication business logic
│   │   ├── user.go              # User business logic
│   │   ├── department.go        # Department business logic
│   │   ├── apikey.go            # API Key business logic
│   │   ├── stats.go             # Statistics business logic
│   │   ├── limit.go             # Rate limit business logic
│   │   ├── llm_proxy.go         # LLM proxy business logic
│   │   └── system.go            # System management business logic
│   ├── repository/
│   │   ├── user.go              # User data access
│   │   ├── department.go        # Department data access
│   │   ├── apikey.go            # API Key data access
│   │   ├── usage.go             # Token usage data access
│   │   ├── ratelimit.go         # Rate limit config data access
│   │   ├── audit.go             # Audit log data access
│   │   ├── announcement.go      # Announcement data access
│   │   └── system.go            # System config data access
│   ├── model/
│   │   ├── user.go              # User model
│   │   ├── department.go        # Department model
│   │   ├── apikey.go            # API Key model
│   │   ├── usage.go             # Token usage model
│   │   ├── ratelimit.go         # Rate limit model
│   │   ├── audit.go             # Audit log model
│   │   ├── announcement.go      # Announcement model
│   │   ├── system.go            # System config model
│   │   └── dto/                 # Data Transfer Objects
│   │       ├── request.go       # Request DTOs
│   │       └── response.go      # Response DTOs
│   ├── router/
│   │   └── router.go            # Route definitions
│   └── pkg/
│       ├── jwt/
│       │   └── jwt.go           # JWT utility
│       ├── crypto/
│       │   └── crypto.go        # Encryption/hashing utility
│       ├── response/
│       │   └── response.go      # Unified response helper
│       ├── validator/
│       │   └── validator.go     # Custom validators
│       └── errcode/
│           └── errcode.go       # Error code definitions
├── pkg/
│   ├── llm/
│   │   ├── client.go            # LLM HTTP client
│   │   ├── types.go             # LLM request/response types
│   │   └── stream.go            # SSE stream handler
│   └── token/
│       └── counter.go           # Token counting utility
└── migrations/
    ├── 000001_init_schema.up.sql
    └── 000001_init_schema.down.sql
```

### 8.2 核心模块详细设计

#### 8.2.1 认证模块 (Auth)

**JWT Token 设计**：
- 算法：HS256
- 有效期：24 小时（可配置）
- Payload 包含：`user_id`, `username`, `role`, `department_id`, `exp`, `iat`
- 登出通过 Redis 黑名单实现（将 token 的 JTI 加入黑名单，TTL = 剩余有效期）

**认证流程**：
```
1. 用户提交用户名密码
2. 查询数据库验证用户
3. 验证密码（bcrypt compare）
4. 检查用户状态（是否禁用）
5. 生成 JWT Token
6. 更新最后登录时间和 IP
7. 返回 Token 和用户信息
```

#### 8.2.2 API Key 认证模块 (APIKey Auth)

**API Key 格式**：`cm-{32位随机hex字符}`

**验证流程**：
```
1. 从 Authorization Header 提取 Key
2. 计算 Key 的 SHA-256 哈希
3. 查询数据库（优先查 Redis 缓存）
4. 验证 Key 状态和过期时间
5. 获取关联用户信息
6. 检查用户状态
7. 将用户信息注入请求上下文
```

**缓存策略**：
```
Redis Key: codemind:apikey:{sha256_hash}
Value: JSON { user_id, username, role, department_id, key_id, status }
TTL: 300s (5 分钟)
Invalidation: Key 状态变更时删除缓存
```

#### 8.2.3 限流模块 (Rate Limit)

**Token 限额检查**：

```
1. 获取用户的限额配置（优先级：用户级 > 部门级 > 全局）
2. 从 Redis 获取当前周期已使用量
   Key: codemind:usage:{user_id}:{period}:{period_key}
   例: codemind:usage:5:daily:2026-02-15
3. 判断剩余量是否足够
4. 请求完成后更新 Redis 计数（INCRBY）
5. 异步同步到数据库
```

**并发控制**：

```
Redis Key: codemind:concurrency:{user_id}
操作: INCR / DECR
检查: GET value <= max_concurrency
TTL: 300s (防止泄漏)
```

#### 8.2.4 LLM 代理模块 (LLM Proxy)

**核心职责**：
1. 请求转发：将客户端请求转发至后端 LLM 服务
2. 流式传输：支持 SSE（Server-Sent Events）流式响应
3. Token 计量：从 LLM 响应中提取 token 用量
4. 用量记录：将每次请求的用量写入数据库

**流式代理实现思路**：
```go
func (h *LLMProxyHandler) ChatCompletions(c *gin.Context) {
    // 1. Parse and validate request
    var req openai.ChatCompletionRequest
    c.ShouldBindJSON(&req)

    // 2. Override/validate model
    req.Model = validateModel(req.Model)

    // 3. Forward to LLM server
    llmResp, err := h.llmClient.CreateChatCompletion(req)

    if req.Stream {
        // 4a. Stream response
        c.Writer.Header().Set("Content-Type", "text/event-stream")
        c.Writer.Header().Set("Cache-Control", "no-cache")
        c.Writer.Header().Set("Connection", "keep-alive")

        var totalUsage Usage
        for chunk := range llmResp.StreamChannel() {
            // Forward each chunk to client
            c.SSEvent("", chunk)
            c.Writer.Flush()
            // Accumulate usage from final chunk
            if chunk.Usage != nil {
                totalUsage = *chunk.Usage
            }
        }

        // 5. Record usage
        h.statsService.RecordUsage(userId, keyId, req.Model, totalUsage)
    } else {
        // 4b. Non-stream response
        c.JSON(200, llmResp)
        h.statsService.RecordUsage(userId, keyId, req.Model, llmResp.Usage)
    }
}
```

---

## 9. LLM 代理层设计

### 9.1 客户端连接设计

LLM 代理层需要处理的客户端类型：
- IDE 插件（Continue、Cline、Cursor、GitHub Copilot 等）
- 自定义脚本
- 其他兼容 OpenAI API 的客户端

### 9.2 LLM Client 设计

```go
// pkg/llm/client.go
type LLMClient struct {
    baseURL    string
    apiKey     string
    httpClient *http.Client
}

// Methods
func (c *LLMClient) ChatCompletion(req ChatCompletionRequest) (*ChatCompletionResponse, error)
func (c *LLMClient) ChatCompletionStream(req ChatCompletionRequest) (<-chan ChatCompletionChunk, error)
func (c *LLMClient) Completion(req CompletionRequest) (*CompletionResponse, error)
func (c *LLMClient) CompletionStream(req CompletionRequest) (<-chan CompletionChunk, error)
func (c *LLMClient) ListModels() (*ModelListResponse, error)
```

### 9.3 超时与重试策略

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| 连接超时 | 10s | 建立连接的超时时间 |
| 请求超时 | 300s | 非流式请求的最大等待时间 |
| 流式超时 | 600s | 流式请求的最大持续时间 |
| 重试次数 | 0 | LLM 请求不重试（避免重复计费） |

### 9.4 错误处理

| LLM 返回状态 | 平台处理方式 |
|-------------|-------------|
| 200 | 正常转发 |
| 400 | 转发错误给客户端 |
| 429 | 返回 503（LLM 服务繁忙） |
| 500 | 返回 502（LLM 服务错误） |
| 超时 | 返回 504（LLM 服务超时） |
| 连接失败 | 返回 503（LLM 服务不可用） |

---

## 10. 安全设计

### 10.1 密码安全

- 使用 bcrypt 加密存储，cost factor = 12
- 密码强度要求：最少 8 位，包含大小写字母和数字
- 首次登录强制修改初始密码（可配置）

### 10.2 API Key 安全

- Key 仅在创建时显示完整值，之后不可再查看
- 数据库中仅存储 SHA-256 哈希值
- Key 前缀 `cm-` 用于识别来源
- 支持设置过期时间

### 10.3 JWT 安全

- 使用 HS256 算法签名
- Token 有效期 24 小时
- 支持通过 Redis 黑名单使 Token 失效
- Refresh Token 机制（可选，后续迭代）

### 10.4 接口安全

- 所有管理接口需要 JWT 认证
- 所有 LLM 接口需要 API Key 认证
- 基于角色的访问控制（RBAC）
- 请求参数严格校验
- SQL 注入防护（GORM 参数化查询）
- XSS 防护（前端输入转义）
- CORS 白名单配置

### 10.5 审计日志

- 所有敏感操作记录审计日志
- 包含：操作者、操作类型、操作目标、变更详情、IP 地址、时间
- 审计日志不可修改和删除

---

## 11. 部署方案

### 11.1 Docker 容器设计

| 容器 | 基础镜像 | 端口 | 说明 |
|------|----------|------|------|
| codemind-frontend | nginx:1.27-alpine | 80 | 前端静态文件 + 反向代理 |
| codemind-backend | golang:1.23-alpine (build) / alpine:3.20 (run) | 8080 | 后端 API 服务 |
| codemind-postgres | postgres:16-alpine | 5432 | PostgreSQL 数据库 |
| codemind-redis | redis:7-alpine | 6379 | Redis 缓存 |

### 11.2 docker-compose.yml 设计

```yaml
services:
  frontend:
    build: ./frontend
    ports:
      - "80:80"
      - "443:443"
    depends_on:
      - backend
    networks:
      - codemind-network

  backend:
    build: ./backend
    ports:
      - "8080:8080"
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    environment:
      - DB_HOST=postgres
      - DB_PORT=5432
      - REDIS_HOST=redis
      - REDIS_PORT=6379
    networks:
      - codemind-network

  postgres:
    image: postgres:16-alpine
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./deploy/docker/postgres/init.sql:/docker-entrypoint-initdb.d/init.sql
    environment:
      POSTGRES_DB: codemind
      POSTGRES_USER: codemind
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U codemind"]
      interval: 5s
      timeout: 5s
      retries: 5
    networks:
      - codemind-network

  redis:
    image: redis:7-alpine
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 5s
      retries: 5
    networks:
      - codemind-network

volumes:
  postgres_data:
  redis_data:

networks:
  codemind-network:
    driver: bridge
```

### 11.3 Nginx 反向代理配置

```nginx
# Frontend static files
location / {
    root /usr/share/nginx/html;
    try_files $uri $uri/ /index.html;
}

# Management API proxy
location /api/ {
    proxy_pass http://backend:8080;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
}

# OpenAI 兼容 LLM 代理（含 SSE）
location /api/openai/ {
    proxy_pass http://backend:8080;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_buffering off;           # SSE 所需
    proxy_cache off;
    proxy_read_timeout 600s;       # LLM 长连接超时
    chunked_transfer_encoding on;
}

# Anthropic LLM 代理（含 SSE）
location /api/anthropic/ {
    proxy_pass http://backend:8080;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_buffering off;
    proxy_cache off;
    proxy_read_timeout 600s;
    chunked_transfer_encoding on;
}
```

### 11.4 环境配置

```yaml
# deploy/config/app.yaml.example
server:
  host: "0.0.0.0"
  port: 8080
  mode: "release"  # debug / release

database:
  host: "postgres"
  port: 5432
  name: "codemind"
  user: "codemind"
  password: "${DB_PASSWORD}"
  max_open_conns: 50
  max_idle_conns: 10

redis:
  host: "redis"
  port: 6379
  password: ""
  db: 0

jwt:
  secret: "${JWT_SECRET}"
  expire_hours: 24

llm:
  base_url: "http://llm-server:8080"
  api_key: ""
  timeout_seconds: 300
  stream_timeout_seconds: 600

system:
  max_keys_per_user: 10
  default_concurrency: 5
  default_daily_tokens: 1000000
  default_monthly_tokens: 20000000
```

---

## 12. 测试方案

### 12.1 测试策略

| 测试类型 | 覆盖范围 | 工具 | 目标覆盖率 |
|----------|----------|------|-----------|
| 单元测试 | Service / Repository / Utils | Go testing + testify | > 80% |
| 集成测试 | Handler（含数据库） | Go testing + httptest | > 70% |
| 前端单元测试 | 组件 / Hooks / Utils | Vitest + React Testing Library | > 70% |
| E2E 测试 | 关键业务流程 | Playwright（后续迭代） | 核心流程 |

### 12.2 后端测试规范

#### 12.2.1 单元测试

每个 Service 和 Repository 方法都需要对应的测试：

```go
// service/user_test.go
func TestUserService_CreateUser(t *testing.T) {
    // Setup mock repository
    // Test normal case
    // Test duplicate username
    // Test invalid department
    // Test permission check
}
```

#### 12.2.2 集成测试

使用 Docker 启动测试数据库，测试完整请求链路：

```go
// tests/integration/user_test.go
func TestCreateUser_Integration(t *testing.T) {
    // Setup test server with test database
    // Login as admin
    // Create user via API
    // Verify user in database
    // Cleanup
}
```

### 12.3 前端测试规范

```typescript
// tests/components/LoginForm.test.tsx
describe('LoginForm', () => {
  it('should render login form', () => { ... });
  it('should validate required fields', () => { ... });
  it('should call login API on submit', () => { ... });
  it('should show error on invalid credentials', () => { ... });
});
```

### 12.4 测试数据

- 提供测试数据种子脚本（`deploy/docker/postgres/seed.sql`）
- 默认创建超级管理员账号：`admin / Admin@123456`
- 创建测试部门和测试用户

---

## 13. 版本管理规范

### 13.1 版本号规范

采用 [语义化版本 (SemVer)](https://semver.org/lang/zh-CN/) 规范：

```
MAJOR.MINOR.PATCH
```

| 版本位 | 变更条件 | 示例 |
|--------|----------|------|
| MAJOR | 不兼容的 API 变更 | 1.0.0 → 2.0.0 |
| MINOR | 向后兼容的功能新增 | 1.0.0 → 1.1.0 |
| PATCH | 向后兼容的问题修复 | 1.0.0 → 1.0.1 |

**初始版本**：`0.1.0`（开发阶段以 0.x.x 标识）

**版本迭代计划**：
- `0.1.0` - 项目框架搭建、基础认证
- `0.2.0` - 用户管理、部门管理
- `0.3.0` - API Key 管理
- `0.4.0` - LLM 代理层
- `0.5.0` - 用量统计
- `0.6.0` - 限额管理
- `0.7.0` - 系统管理
- `0.8.0` - 首页 & UI 完善
- `0.9.0` - 测试完善 & Bug 修复
- `1.0.0` - 首个正式发布版本

### 13.2 Git 分支规范

| 分支 | 用途 | 命名规范 |
|------|------|----------|
| main | 生产分支 | - |
| develop | 开发分支 | - |
| feature/* | 功能分支 | feature/user-management |
| fix/* | 修复分支 | fix/login-validation |
| release/* | 发布分支 | release/0.1.0 |

### 13.3 Commit Message 规范

采用 [Conventional Commits](https://www.conventionalcommits.org/) 规范：

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

**Type 类型**：
| Type | 说明 |
|------|------|
| feat | 新功能 |
| fix | Bug 修复 |
| docs | 文档变更 |
| style | 代码格式（不影响功能） |
| refactor | 代码重构 |
| perf | 性能优化 |
| test | 测试相关 |
| chore | 构建/工具变更 |

**示例**：
```
feat(user): add user creation API

- Add POST /api/v1/users endpoint
- Add input validation for user fields
- Add duplicate username check

Closes #12
```

### 13.4 CHANGELOG 规范

使用 [Keep a Changelog](https://keepachangelog.com/zh-CN/) 格式：

```markdown
## [0.2.0] - 2026-03-01

### Added
- 用户管理：创建、编辑、禁用、删除用户
- 部门管理：创建、编辑、删除部门

### Changed
- 优化登录接口响应速度

### Fixed
- 修复密码校验规则不生效的问题
```

---

## 14. 开发阶段规划

### Phase 1: 项目初始化与基础框架（预计 3 天）

#### 目标
搭建前后端项目框架，实现基础配置和开发环境。

#### 任务清单

| # | 任务 | 具体内容 | 产出 |
|---|------|----------|------|
| 1.1 | 初始化后端项目 | 创建 Go 项目，配置 Gin、GORM、Viper、Zap | go.mod, main.go, config |
| 1.2 | 初始化前端项目 | 创建 React + Vite 项目，配置 Ant Design、TailwindCSS | package.json, vite.config |
| 1.3 | 数据库初始化 | 编写数据库迁移脚本，创建所有表 | SQL migration files |
| 1.4 | Docker 环境 | 编写 Dockerfile 和 docker-compose.yml | Docker 配置文件 |
| 1.5 | 统一响应封装 | 后端统一响应格式、错误码定义 | response.go, errcode.go |
| 1.6 | 前端请求封装 | Axios 实例配置、拦截器、API 封装 | request.ts |
| 1.7 | 路由框架 | 前后端路由框架搭建 | router 配置 |
| 1.8 | 测试框架 | 前后端测试环境配置 | 测试配置文件 |

#### 验收标准
- [x] docker-compose up 可以正常启动所有服务
- [x] 数据库表结构正确创建
- [x] 前后端项目可以正常启动
- [x] 基础测试可以运行

---

### Phase 2: 认证模块（预计 3 天）

#### 目标
实现用户登录、登出、JWT 认证中间件。

#### 任务清单

| # | 任务 | 具体内容 | 产出 |
|---|------|----------|------|
| 2.1 | JWT 工具 | 实现 JWT 生成、解析、黑名单管理 | jwt.go |
| 2.2 | 密码工具 | bcrypt 加密、验证 | crypto.go |
| 2.3 | 认证中间件 | JWT 验证中间件，从 Header 提取并验证 Token | auth.go middleware |
| 2.4 | 登录接口 | 实现 POST /api/v1/auth/login | handler + service + repo |
| 2.5 | 登出接口 | 实现 POST /api/v1/auth/logout | handler + service |
| 2.6 | 个人信息接口 | 实现 GET/PUT /api/v1/auth/profile | handler + service |
| 2.7 | 修改密码接口 | 实现 PUT /api/v1/auth/password | handler + service |
| 2.8 | 登录页面 | 前端登录页面和表单 | LoginPage.tsx |
| 2.9 | 认证状态管理 | 前端 auth store，token 持久化 | authStore.ts |
| 2.10 | 路由守卫 | 前端路由鉴权守卫 | guards.tsx |
| 2.11 | 单元测试 | 认证模块全部单元测试 | *_test.go, *.test.tsx |

#### 验收标准
- [x] 可以使用用户名密码登录
- [x] JWT Token 正确生成和验证
- [x] 登出后 Token 失效
- [x] 未认证请求被正确拒绝
- [x] 单元测试覆盖率 > 80%

---

### Phase 3: 用户与部门管理（预计 5 天）

#### 目标
实现完整的用户管理和部门管理功能。

#### 任务清单

| # | 任务 | 具体内容 | 产出 |
|---|------|----------|------|
| 3.1 | 权限中间件 | 基于角色的权限校验中间件 | permission middleware |
| 3.2 | 用户 CRUD 接口 | 完整的用户增删改查 API | handler + service + repo |
| 3.3 | 用户状态管理 | 启用/禁用用户 | handler + service |
| 3.4 | 密码重置 | 管理员重置用户密码 | handler + service |
| 3.5 | 批量导入 | CSV 批量导入用户 | handler + service |
| 3.6 | 部门 CRUD 接口 | 完整的部门增删改查 API | handler + service + repo |
| 3.7 | 部门树形结构 | 部门层级关系查询 | service + repo |
| 3.8 | 审计日志 | 敏感操作审计日志记录 | audit service + repo |
| 3.9 | 用户管理页面 | 前端用户列表、创建/编辑弹窗 | admin/users pages |
| 3.10 | 部门管理页面 | 前端部门列表、创建/编辑弹窗 | admin/departments pages |
| 3.11 | 个人中心页面 | 个人信息查看和编辑 | profile page |
| 3.12 | 管理后台布局 | 侧边栏、顶栏、内容区布局组件 | layout components |
| 3.13 | 单元测试 | 用户和部门模块全部测试 | test files |

#### 验收标准
- [x] 超级管理员可以管理所有用户和部门
- [x] 部门经理只能管理本部门用户
- [x] 普通用户无法访问管理功能
- [x] 部门树形结构正确展示
- [x] 批量导入功能正常工作
- [x] 所有敏感操作有审计日志
- [x] 单元测试覆盖率 > 80%

---

### Phase 4: API Key 管理（预计 3 天）

#### 目标
实现 API Key 的创建、管理和认证功能。

#### 任务清单

| # | 任务 | 具体内容 | 产出 |
|---|------|----------|------|
| 4.1 | Key 生成工具 | API Key 生成和哈希工具 | crypto.go |
| 4.2 | Key CRUD 接口 | API Key 的增删改查 API | handler + service + repo |
| 4.3 | Key 认证中间件 | API Key 验证中间件 | apikey.go middleware |
| 4.4 | Key 缓存 | Redis 缓存 Key 信息 | service |
| 4.5 | Key 管理页面 | 前端 Key 列表、创建弹窗、复制功能 | keys page |
| 4.6 | Key 用量查看 | Key 维度的用量查看 | service + page |
| 4.7 | 单元测试 | API Key 模块全部测试 | test files |

#### 验收标准
- [x] 用户可以创建、查看、禁用、删除 API Key
- [x] Key 仅在创建时显示一次完整值
- [x] 使用 Key 可以正确认证 LLM 请求
- [x] Key 禁用后无法认证
- [x] Key 数量限制生效
- [x] 单元测试覆盖率 > 80%

---

### Phase 5: LLM 代理层（预计 5 天）

#### 目标
实现 LLM 请求的代理转发，支持流式和非流式响应。

#### 任务清单

| # | 任务 | 具体内容 | 产出 |
|---|------|----------|------|
| 5.1 | LLM Client | HTTP 客户端，支持流式和非流式请求 | pkg/llm/client.go |
| 5.2 | SSE 流式处理 | Server-Sent Events 流式转发 | pkg/llm/stream.go |
| 5.3 | Chat Completions | POST /api/openai/v1/chat/completions 代理 | handler |
| 5.4 | Completions | POST /api/openai/v1/completions 代理 | handler |
| 5.5 | Models List | GET /api/openai/v1/models 代理 | handler |
| 5.6 | Token 计量 | 从响应中解析 token 用量 | pkg/token/counter.go |
| 5.7 | 用量记录 | 请求完成后写入用量数据 | service + repo |
| 5.8 | 并发控制 | Redis 实现的并发请求限制 | middleware |
| 5.9 | 错误处理 | LLM 服务异常的错误处理和转换 | handler |
| 5.10 | 请求日志 | 记录所有 LLM 请求日志 | service + repo |
| 5.11 | 单元测试 | LLM 代理模块全部测试（Mock LLM 服务） | test files |

#### 验收标准
- [x] 兼容 OpenAI API 格式的请求和响应
- [x] 流式请求正常工作
- [x] Token 用量正确统计
- [x] 并发限制生效
- [x] 各种异常情况正确处理
- [x] 单元测试覆盖率 > 80%

---

### Phase 6: 用量统计（预计 4 天）

#### 目标
实现用量数据的聚合统计和可视化展示。

#### 任务清单

| # | 任务 | 具体内容 | 产出 |
|---|------|----------|------|
| 6.1 | 日汇总服务 | Token 用量日汇总逻辑 | service + repo |
| 6.2 | 统计查询接口 | 日/周/月统计查询 API | handler + service |
| 6.3 | 总览接口 | 系统总览数据 API | handler + service |
| 6.4 | 排行榜接口 | 用户/部门用量排行 API | handler + service |
| 6.5 | 数据导出 | CSV/Excel 导出功能 | handler + service |
| 6.6 | Dashboard 页面 | 仪表盘总览页面 | dashboard page |
| 6.7 | 用量统计页面 | 详细的用量统计页面（图表） | usage page |
| 6.8 | 图表组件 | ECharts 图表组件封装 | chart components |
| 6.9 | 管理员统计视图 | 管理员查看全局统计 | admin stats page |
| 6.10 | 单元测试 | 统计模块全部测试 | test files |

#### 验收标准
- [x] 日/周/月统计数据正确
- [x] 图表正确展示用量趋势
- [x] 管理员可查看全局和用户级统计
- [x] 部门经理可查看本部门统计
- [x] 普通用户只能看个人统计
- [x] 数据导出功能正常
- [x] 单元测试覆盖率 > 80%

---

### Phase 7: 限额管理（预计 3 天）

#### 目标
实现完整的限额配置和限额检查功能。

#### 任务清单

| # | 任务 | 具体内容 | 产出 |
|---|------|----------|------|
| 7.1 | 限额配置接口 | 限额的增删改查 API | handler + service + repo |
| 7.2 | 限额检查中间件 | 在 LLM 请求前检查 Token 限额 | middleware |
| 7.3 | 限额优先级逻辑 | 用户 > 部门 > 全局的优先级 | service |
| 7.4 | Redis 限额计数 | Redis 实现的周期限额计数 | service |
| 7.5 | 限额告警 | 达到阈值时的告警逻辑 | service |
| 7.6 | 限额管理页面 | 管理员限额配置页面 | admin limits page |
| 7.7 | 个人限额展示 | 用户查看自己的限额和用量 | dashboard widget |
| 7.8 | 单元测试 | 限额模块全部测试 | test files |

#### 验收标准
- [x] 全局/部门/用户三级限额配置
- [x] 限额优先级正确
- [x] 超限后请求被正确拒绝
- [x] 限额告警正常触发
- [x] 并发限制正常工作
- [x] 单元测试覆盖率 > 80%

---

### Phase 8: 系统管理（预计 3 天）

#### 目标
实现系统配置、公告管理和审计日志查看。

#### 任务清单

| # | 任务 | 具体内容 | 产出 |
|---|------|----------|------|
| 8.1 | 系统配置接口 | 系统配置的读写 API | handler + service + repo |
| 8.2 | LLM 服务配置 | LLM 连接配置、模型管理 | service |
| 8.3 | 公告管理接口 | 公告的增删改查 API | handler + service + repo |
| 8.4 | 审计日志查询 | 审计日志的查询 API | handler + service |
| 8.5 | 系统设置页面 | 管理员系统设置页面 | admin settings page |
| 8.6 | 模型管理页面 | 管理员模型配置页面 | admin models page |
| 8.7 | 公告管理页面 | 管理员公告管理页面 | admin announcements page |
| 8.8 | 审计日志页面 | 管理员审计日志页面 | admin audit-logs page |
| 8.9 | 公告展示 | 前台公告通知展示 | announcement component |
| 8.10 | 单元测试 | 系统管理模块全部测试 | test files |

#### 验收标准
- [x] LLM 服务配置可动态修改
- [x] 公告发布和展示正常
- [x] 审计日志完整记录和查询
- [x] 仅管理员可访问系统管理功能
- [x] 单元测试覆盖率 > 80%

---

### Phase 9: 首页 & UI 优化（预计 4 天）

#### 目标
设计开发酷炫的首页，优化整体 UI 体验。

#### 任务清单

| # | 任务 | 具体内容 | 产出 |
|---|------|----------|------|
| 9.1 | 首页 Hero 区域 | 全屏 Hero 区域，品牌展示、标语 | home page |
| 9.2 | 功能介绍区域 | 平台功能特点展示（卡片/图标） | home section |
| 9.3 | 动效设计 | 滚动动画、悬浮效果、粒子背景 | animations |
| 9.4 | 响应式适配 | 移动端/平板适配 | responsive styles |
| 9.5 | 品牌配色落地 | 全平台统一配色方案应用 | theme config |
| 9.6 | 暗色模式 | 支持暗色模式切换（可选） | dark mode |
| 9.7 | Loading 体验 | 页面加载、数据加载状态优化 | loading states |
| 9.8 | 空状态设计 | 各页面空状态占位设计 | empty states |
| 9.9 | 通知提示 | 操作反馈、消息通知优化 | notification |
| 9.10 | UI Review | 全平台 UI 走查和优化 | UI fixes |

#### 验收标准
- [x] 首页视觉效果酷炫、专业
- [x] 配色统一，符合品牌调性
- [x] 动效流畅，不卡顿
- [x] 所有页面体验一致
- [x] 响应式适配良好

---

### Phase 10: 测试完善 & 发布准备（预计 3 天）

#### 目标
补充测试用例，修复 Bug，准备首个正式版本发布。

#### 任务清单

| # | 任务 | 具体内容 | 产出 |
|---|------|----------|------|
| 10.1 | 补充单元测试 | 确保所有模块覆盖率 > 80% | test files |
| 10.2 | 集成测试 | 端到端关键流程测试 | integration tests |
| 10.3 | Bug 修复 | 修复测试发现的 Bug | bug fixes |
| 10.4 | 性能测试 | 基础负载测试 | test report |
| 10.5 | 安全检查 | 安全漏洞扫描和修复 | security report |
| 10.6 | 文档完善 | API 文档、部署文档、用户手册 | docs |
| 10.7 | 种子数据 | 初始化管理员账号和默认配置 | seed.sql |
| 10.8 | Release 准备 | 版本号、CHANGELOG、Tag | v1.0.0 release |

#### 验收标准
- [x] 所有测试通过
- [x] 单元测试覆盖率 > 80%
- [x] 无已知严重 Bug
- [x] 文档完善
- [x] docker-compose 可一键部署

---

## 15. UI/UX 设计规范

### 15.1 品牌配色方案

基于 Logo 提取的配色方案：

| 色彩角色 | 色值 | 用途 |
|----------|------|------|
| Primary Blue | `#2B7CB3` | 主色调，按钮、链接、高亮 |
| Secondary Blue | `#4BA3D4` | 辅助色，渐变、hover 状态 |
| Accent Cyan | `#6BC5E8` | 强调色，图标、装饰 |
| Light Cyan | `#8FD8F0` | 浅色装饰、背景渐变 |
| Pale Cyan | `#D4F1F9` | 极浅背景、卡片背景 |
| Dark Navy | `#1A3A5C` | 正文文字、标题 |
| Medium Navy | `#2E5A7E` | 二级文字 |
| Light Gray | `#F0F5FA` | 页面背景 |
| White | `#FFFFFF` | 卡片背景、内容区 |
| Success | `#52C41A` | 成功状态 |
| Warning | `#FAAD14` | 警告状态 |
| Error | `#FF4D4F` | 错误状态 |

### 15.2 Ant Design 主题定制

```typescript
// theme/themeConfig.ts
const themeConfig = {
  token: {
    colorPrimary: '#2B7CB3',
    colorInfo: '#4BA3D4',
    colorSuccess: '#52C41A',
    colorWarning: '#FAAD14',
    colorError: '#FF4D4F',
    colorBgLayout: '#F0F5FA',
    borderRadius: 8,
    fontFamily: "'Inter', 'PingFang SC', 'Microsoft YaHei', sans-serif",
  },
};
```

### 15.3 首页设计方案

#### Hero 区域
- 全屏渐变背景（从 `#1A3A5C` 到 `#2B7CB3`）
- 动态粒子/代码雨/网格线背景动效
- 居中显示 Logo、中英文名称、一句话标语
- 醒目的「登录」和「了解更多」按钮

#### 功能展示区域
- 4-6 个功能亮点卡片，每个配图标和简短描述
  - AI 智能编码：接入强大的本地 LLM 模型
  - 安全可控：企业级安全管控，数据不出服务器
  - 用量透明：实时 Token 统计，清晰的使用报表
  - 灵活配额：多层级配额管理，按需分配资源
  - API Key 管理：便捷的 Key 管理，轻松接入开发工具
  - 团队协作：部门级管理，促进团队高效协作

#### 底部区域
- 平台版本号
- 版权信息
- 技术支持联系方式

### 15.4 管理后台设计方案

#### 布局结构
```
┌─────────────────────────────────────────────┐
│  Logo    | Search        | User Avatar  ▼   │  ← Header
├──────┬──────────────────────────────────────┤
│      │                                      │
│ Nav  │          Content Area                │
│ Menu │                                      │
│      │                                      │
│      │                                      │
│      │                                      │
│      │                                      │
│      │                                      │
├──────┴──────────────────────────────────────┤
│                   Footer                    │
└─────────────────────────────────────────────┘
```

#### 侧边导航菜单
- **普通用户可见**：
  - 总览 Dashboard
  - 用量统计
  - API Key 管理
  - 个人中心
- **管理员额外可见**：
  - 用户管理
  - 部门管理
  - 限额管理
  - 模型管理
  - 公告管理
  - 审计日志
  - 系统设置

---

## 16. 附录

### 16.1 Redis Key 设计

| Key Pattern | Type | TTL | 用途 |
|-------------|------|-----|------|
| `codemind:jwt:blacklist:{jti}` | String | JWT 剩余有效期 | JWT 黑名单 |
| `codemind:apikey:{hash}` | String(JSON) | 300s | API Key 信息缓存 |
| `codemind:concurrency:{user_id}` | String(Counter) | 300s | 并发请求计数 |
| `codemind:usage:{user_id}:daily:{date}` | String(Counter) | 48h | 每日 Token 用量计数 |
| `codemind:usage:{user_id}:monthly:{month}` | String(Counter) | 35d | 每月 Token 用量计数 |
| `codemind:ratelimit:{user_id}` | String(JSON) | 600s | 用户限额配置缓存 |

### 16.2 初始化数据

系统启动时自动初始化：

1. **默认超级管理员**：
   - 用户名：`admin`
   - 密码：`Admin@123456`（首次登录强制修改）
   - 角色：`super_admin`

2. **默认全局限额**：
   - 每日 Token 上限：1,000,000
   - 每月 Token 上限：20,000,000
   - 默认并发数：5

3. **默认系统配置**：
   - 每用户最大 Key 数：10
   - JWT 过期时间：24 小时

### 16.3 环境变量列表

| 环境变量 | 说明 | 默认值 |
|----------|------|--------|
| `APP_ENV` | 运行环境 | development |
| `APP_PORT` | 服务端口 | 8080 |
| `DB_HOST` | 数据库地址 | localhost |
| `DB_PORT` | 数据库端口 | 5432 |
| `DB_NAME` | 数据库名 | codemind |
| `DB_USER` | 数据库用户 | codemind |
| `DB_PASSWORD` | 数据库密码 | - |
| `REDIS_HOST` | Redis 地址 | localhost |
| `REDIS_PORT` | Redis 端口 | 6379 |
| `REDIS_PASSWORD` | Redis 密码 | - |
| `JWT_SECRET` | JWT 签名密钥 | - |
| `LLM_BASE_URL` | LLM 服务地址 | - |
| `LLM_API_KEY` | LLM 服务密钥 | - |

### 16.4 项目目录结构总览

```
CodeMind/
├── README.md                          # Project README
├── CHANGELOG.md                       # Changelog
├── VERSION                            # Version file
├── .gitignore                         # Git ignore rules
├── docker-compose.yml                 # Docker Compose configuration
├── docs/                              # Documentation
│   ├── development-plan.md            # This document
│   └── api/                           # API documentation
│       └── openapi.yaml               # OpenAPI specification
├── frontend/                          # Frontend project
│   ├── Dockerfile                     # Frontend Docker build
│   ├── nginx.conf                     # Nginx configuration
│   ├── package.json                   # Node.js dependencies
│   ├── tsconfig.json                  # TypeScript configuration
│   ├── vite.config.ts                 # Vite configuration
│   ├── tailwind.config.js             # TailwindCSS configuration
│   ├── public/                        # Public static assets
│   │   └── logo.svg                   # Logo file
│   ├── src/                           # Source code
│   │   ├── main.tsx                   # Application entry
│   │   ├── App.tsx                    # Root component
│   │   ├── assets/                    # Assets
│   │   │   ├── images/                # Image files
│   │   │   └── styles/                # Global styles
│   │   ├── components/                # Shared components
│   │   │   ├── common/                # Common components
│   │   │   ├── layout/                # Layout components
│   │   │   └── charts/                # Chart components
│   │   ├── pages/                     # Page components
│   │   │   ├── home/                  # Landing page
│   │   │   ├── login/                 # Login page
│   │   │   ├── dashboard/             # Dashboard
│   │   │   ├── profile/               # Profile page
│   │   │   ├── keys/                  # API Key management
│   │   │   ├── usage/                 # Usage statistics
│   │   │   └── admin/                 # Admin pages
│   │   │       ├── users/             # User management
│   │   │       ├── departments/       # Department management
│   │   │       ├── limits/            # Rate limit management
│   │   │       ├── models/            # Model management
│   │   │       ├── announcements/     # Announcement management
│   │   │       ├── audit-logs/        # Audit logs
│   │   │       └── settings/          # System settings
│   │   ├── hooks/                     # Custom React hooks
│   │   ├── services/                  # API service layer
│   │   ├── store/                     # Zustand state stores
│   │   ├── types/                     # TypeScript type definitions
│   │   ├── utils/                     # Utility functions
│   │   └── router/                    # React Router configuration
│   └── tests/                         # Frontend tests
│       ├── components/                # Component tests
│       ├── pages/                     # Page tests
│       └── utils/                     # Utility tests
├── backend/                           # Backend project
│   ├── Dockerfile                     # Backend Docker build
│   ├── Makefile                       # Build & dev commands
│   ├── go.mod                         # Go module definition
│   ├── cmd/                           # Entry points
│   │   └── server/                    # Main server
│   │       └── main.go               # Application entry
│   ├── internal/                      # Internal packages
│   │   ├── config/                    # Configuration
│   │   ├── middleware/                # HTTP middleware
│   │   ├── handler/                   # HTTP handlers
│   │   ├── service/                   # Business logic
│   │   ├── repository/               # Data access layer
│   │   ├── model/                     # Data models
│   │   │   └── dto/                   # Data Transfer Objects
│   │   ├── router/                    # Route definitions
│   │   └── pkg/                       # Internal shared packages
│   │       ├── jwt/                   # JWT utilities
│   │       ├── crypto/                # Encryption utilities
│   │       ├── response/              # Response helpers
│   │       ├── validator/             # Custom validators
│   │       └── errcode/               # Error code definitions
│   ├── pkg/                           # External shared packages
│   │   ├── llm/                       # LLM client library
│   │   └── token/                     # Token counter
│   ├── migrations/                    # Database migrations
│   └── tests/                         # Backend tests
│       ├── handler/                   # Handler tests
│       ├── service/                   # Service tests
│       ├── repository/                # Repository tests
│       └── integration/               # Integration tests
├── deploy/                            # Deployment configurations
│   ├── docker/                        # Docker-related configs
│   │   ├── postgres/                  # PostgreSQL initialization
│   │   │   ├── init.sql               # Schema initialization
│   │   │   └── seed.sql               # Seed data
│   │   ├── redis/                     # Redis configuration
│   │   │   └── redis.conf             # Redis config file
│   │   └── nginx/                     # Nginx configuration
│   │       └── nginx.conf             # Nginx config file
│   ├── config/                        # Application config
│   │   └── app.yaml.example           # Config template
│   └── scripts/                       # Deployment scripts
│       ├── setup.sh                   # Initial setup script
│       └── migrate.sh                 # Database migration script
└── scripts/                           # Development scripts
    ├── dev.sh                         # Start development environment
    ├── build.sh                       # Build all services
    └── test.sh                        # Run all tests
```

---

> **文档维护说明**: 本文档随项目开发持续更新。每个阶段完成后，需要回顾和更新对应章节的内容，确保文档与实际实现保持一致。
