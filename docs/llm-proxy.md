# LLM 代理与多服务商架构

> 本文档描述 LLM 代理层的路由模式、多服务商管理和第三方模型代理机制。

核心代码位于 `backend/pkg/llm/` 和 `backend/internal/handler/`。

## 内置 LLM 服务商

### HTTP 端点

| 端点 | 协议 | 认证方式 |
|------|------|---------|
| `/api/openai/v1/*` | OpenAI | API Key |
| `/api/anthropic/*` | Anthropic | API Key |

### 核心组件

- **Provider Manager** (`manager.go`)：管理多个 LLM 服务商（OpenAI/Anthropic），支持模型路由规则（如 `claude-*` → anthropic 服务商）
- **Load Balancer** (`balancer.go`)：加权轮询，基于 Redis 的用户亲和性实现粘性会话
- **SSE 流式传输**：所有 LLM 响应通过 Server-Sent Events 流式返回
- **用量记录**：从响应中解析 Token 用量并异步记录

## 第三方模型服务商

第三方代理允许用户绑定自己的 AI 服务账号（如用户自己的 OpenAI/Anthropic 账户），系统透明转发请求。

### 关键特性

| 特性 | 说明 |
|------|------|
| 透明代理 | `handler/third_party_proxy.go` 直接转发请求到用户配置的服务 |
| 无配额控制 | 第三方请求绕过限流和并发检查，用户自行承担费用 |
| 双协议支持 | 每个服务商同时支持 OpenAI 和 Anthropic 协议格式，各有独立 base URL |
| API Key 加密 | 第三方 API Key 使用 AES 加密存储（`internal/pkg/crypto/aes.go`） |

### 核心概念

- **服务商模板（ThirdPartyProviderTemplate）**：管理员预定义的服务入口，供用户选择绑定
- **用户服务商（UserThirdPartyProvider）**：用户绑定的第三方服务配置，含加密 API Key
- **模型解析**：代理处理器中的 `checkAndHandleThirdParty` 先检查请求模型是否属于用户的第三方服务商，再回退到内置路由

### 请求流程

```
请求 → API Key 验证 → 解析模型归属
  ├─ 属于第三方服务商 → 透明代理 → 记录用量
  └─ 不属于 → 内置 LLM 路由（限流 → 并发 → 负载均衡 → 转发 → 流式响应 → 记录用量）
```

## MCP 网关

- **端点**：`/mcp/*`，使用 API Key 认证
- **功能**：Model Context Protocol 集成网关
- **支持传输**：SSE、HTTP streamable、消息端点

## 相关文档

- [系统架构概览](architecture.md) — 整体架构与请求流程
- [API 路由与认证](api-routes.md) — 各端点认证方式
- [核心设计模式](design-patterns.md) — API Key 格式、限流优先级
- [安全要求](security.md) — API Key 加密与存储安全
