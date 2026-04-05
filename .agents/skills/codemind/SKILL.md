---
name: codemind
description: CodeMind 企业级 AI 编码服务管理平台开发规范。在处理 CodeMind 项目时始终加载，包含代码规范（Go/TypeScript）、UI/UX 指南、Git 工作流、安全要求、项目启动命令等。当用户涉及 CodeMind 项目开发、代码审查、功能实现时使用此 skill。
---

# CodeMind 开发规范

> 企业级 AI 编码服务管理平台（LLM 代理 + 用户管理 + 资源控制）
> 
> **技术栈**: React 18 + TypeScript + Vite | Go 1.23 + Gin + GORM | PostgreSQL 16 | Redis 7

---

## 快速参考

### 项目启动
```bash
# 基础设施
docker compose up -d postgres redis

# 后端
cd backend && go mod download && go run cmd/server/main.go

# 前端
cd frontend && npm install && npm run dev
```

### 默认账号
- 用户名: `admin`
- 密码: `Admin@123456`

### 关键端点
- 前端: `http://localhost:3000`
- 后端: `http://localhost:8080`
- 健康检查: `http://localhost:8080/health`

### Git 工作流速查

```bash
# 创建功能分支
git checkout develop && git pull
git checkout -b feature/xxx

# 开发完成后 - rebase 并合并
git checkout develop && git pull
git checkout feature/xxx
git rebase develop  # 解决冲突后强制推送
git push --force-with-lease

git checkout develop
git merge --no-ff feature/xxx
git push

git branch -d feature/xxx && git push origin --delete feature/xxx

# ⚠️ master 合并需用户确认
# 1. develop rebase master
# 2. master squash merge develop
# 3. 打标签
```

---

## 代码规范要点

### 1. 注释必须使用中文
```go
// UserService 用户服务层
type UserService struct { }

// CreateUser 创建新用户
func (s *UserService) CreateUser(ctx context.Context, req *dto.CreateUserRequest) (*model.User, error) { }
```

### 2. 分层架构（后端）
```
Router → Middleware → Handler → Service → Repository → Model
```

### 3. 组件结构（前端）
```typescript
// 1. 导入 2. 样式 3. 常量 4. 类型 5. 组件 6. Hooks 7. 事件 8. 渲染
```

### 4. Git 提交格式
```
<type>(<scope>): <subject>

feat(user): add user creation API
fix(auth): resolve token expiration issue
```

---

## 安全要求

| 数据类型 | 处理方式 |
|---------|---------|
| 密码 | bcrypt (cost=12) |
| API Key | SHA-256 哈希，格式 `cm-{32-char-hex}` |
| JWT | HS256, 24小时有效期 |
| 数据库查询 | 必须使用参数化查询 |

---

## 版本号管理

格式: `MAJOR.MINOR.PATCH`
- MAJOR: 不兼容 API 变更
- MINOR: 向下兼容功能新增
- PATCH: 向下兼容问题修复

更新步骤:
1. 更新 `VERSION` 文件
2. 更新 `CHANGELOG.md`
3. 创建 Git Tag: `git tag -a v0.6.0 -m "Release version 0.6.0"`

---

## 详细参考文档

| 文档 | 说明 | 何时阅读 |
|------|------|---------|
| [references/code-standards.md](references/code-standards.md) | Go + TypeScript 详细代码规范 | 编写代码时 |
| [references/ui-ux-guide.md](references/ui-ux-guide.md) | 设计系统、配色、组件规范 | UI 开发时 |
| [references/git-workflow.md](references/git-workflow.md) | Git 分支、提交、PR 规范 | 提交代码时 |
| [references/code-review.md](references/code-review.md) | 代码审查清单、覆盖率要求 | 审查代码时 |
| [references/security.md](references/security.md) | 安全规范、加密、防护 | 涉及安全功能时 |
| [references/versioning.md](references/versioning.md) | 版本号管理、发布流程 | 发布版本时 |
| `docs/architecture.md` | 系统架构、请求流程 | 了解架构时 |
| `docs/backend-standards.md` | 后端完整规范 | 后端开发时 |
| `docs/frontend-standards.md` | 前端完整规范 | 前端开发时 |
| `docs/security.md` | 安全规范详情 | 涉及安全时 |

---

## 开发前检查清单

- [ ] 阅读相关参考文档
- [ ] 确认技术方案
- [ ] 代码注释使用中文
- [ ] 敏感操作需审计日志
- [ ] 数据库操作使用参数化查询
