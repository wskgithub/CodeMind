# API 路由与认证

> 本文档描述 CodeMind 的 API 路由分组、认证方式和访问权限。

## 路由分组

| 路由组 | 认证方式 | 访问权限 |
|--------|---------|---------|
| `/health` | 无 | 公开 |
| `/api/v1/auth/*` | 无 | 登录注册 |
| `/api/v1/*`（认证路由） | JWT | 所有已登录用户 |
| `/api/v1/admin/*` | JWT + 角色 | `super_admin` / `dept_manager` |
| `/api/v1/system/*` | JWT + 角色 | 仅 `super_admin` |
| `/api/v1/models/*` | JWT | 模型与第三方服务商 |
| `/api/v1/system/provider-templates/*` | JWT + 管理员 | 仅 `super_admin` |
| `/api/openai/v1/*` | API Key | OpenAI 协议 LLM 代理 |
| `/api/anthropic/*` | API Key | Anthropic 协议 LLM 代理 |
| `/mcp/*` | API Key | MCP 协议网关 |

## 认证机制

### JWT 认证

- **算法**：HS256
- **有效期**：24 小时
- **黑名单**：登出或密码修改时将 JWT 加入 Redis 黑名单
- **传递方式**：`Authorization: Bearer <token>` 请求头

### API Key 认证

- **格式**：`cm-{32-char-hex}`
- **传递方式**：`Authorization: Bearer <api-key>` 请求头（兼容 OpenAI 格式）
- **验证流程**：截取 API Key → SHA-256 哈希 → 查询数据库（优先从 Redis 缓存读取）
- **缓存**：Redis 缓存 300 秒

## Nginx 路由

生产环境中 Nginx 负责请求分发：

| 路径 | 目标 | 说明 |
|------|------|------|
| `/` | 前端静态文件 | SPA 回退 |
| `/api/` | 后端管理 API | - |
| `/api/openai/v1/` | LLM 代理（OpenAI 协议） | SSE 支持，600s 超时 |
| `/api/anthropic/` | LLM 代理（Anthropic 协议） | SSE 支持，600s 超时 |
| `/mcp/` | MCP 网关 | - |

## 相关文档

- [系统架构概览](architecture.md) — 整体架构与请求流程
- [LLM 代理架构](llm-proxy.md) — 代理层详细设计
- [安全要求](security.md) — JWT、API Key 安全规范
- [核心设计模式](design-patterns.md) — RBAC 角色体系
