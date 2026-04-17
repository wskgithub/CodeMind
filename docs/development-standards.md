# CodeMind 开发规范综述

本文档是 CodeMind 项目开发规范的总体指南，包含前后端开发的共同规范。

## 📋 目录

1. [项目信息](#项目信息)
2. [开发工作流](#开发工作流)
3. [代码质量标准](#代码质量标准)
4. [安全规范](#安全规范)
5. [文档要求](#文档要求)

---

## 项目信息

### 基本信息

| 项目 | 说明 |
|------|------|
| 项目名称 | CodeMind |
| 项目代号 | CodeMind |
| 仓库地址 | https://github.com/wskgithub/CodeMind |
| 文档地址 | `docs/` |

### 技术栈概览

**前端**
- React 18 + TypeScript 5
- Vite 6 + Ant Design 5
- Zustand 状态管理
- ECharts 图表

**后端**
- Go 1.23 + Gin 1.10
- GORM 1.25 + PostgreSQL 16
- Redis 7 + Zap 日志

**部署**
- Docker 27 + Docker Compose
- Nginx 反向代理

---

## 开发工作流

### 1. 需求确认

在开始开发前，确保：
- [ ] 已阅读相关 Issue 或需求文档
- [ ] 了解相关功能模块的上下文
- [ ] 确认技术方案和实现方式
- [ ] 与团队沟通确认无歧义

### 2. 分支管理

```
main (生产)
  ↑
  develop (开发)
    ↑
    feature/* (功能)
    fix/* (修复)
```

**创建功能分支**
```bash
git checkout develop
git pull origin develop
git checkout -b feature/user-management
```

### 3. 开发规范

**代码提交前检查**
```bash
# 后端
cd backend
go test ./...
make lint

# 前端
cd frontend
npm test
npm run lint
```

**提交代码**
```bash
git add .
git commit -m "feat(user): add user creation API"
git push origin feature/user-management
```

### 4. 代码审查

- 创建 Pull Request 到 `develop` 分支
- 填写 PR 描述模板
- 至少一人审查通过后合并
- 解决所有审查意见

### 5. 测试要求

| 测试类型 | 后端覆盖率 | 前端覆盖率 |
|---------|-----------|-----------|
| 单元测试 | > 80% | > 70% |
| 集成测试 | 核心流程 | 核心流程 |
| E2E 测试 | - | 关键路径 |

---

## 代码质量标准

### 通用原则

1. **KISS 原则**: 保持简单，避免过度设计
2. **DRY 原则**: 不要重复代码，提取公共逻辑
3. **SOLID 原则**: 遵循单一职责、开闭原则等
4. **可读性优先**: 代码是写给人看的，其次是机器

### 后端 Go 规范

**文件组织**
```go
// 1. 包声明和导入
package service

import (
    "context"
    "fmt"

    "codemind/internal/model"
    "codemind/internal/repository"
    "go.uber.org/zap"
)

// 2. 常量定义
const (
    DefaultPageSize = 20
    MaxPageSize     = 100
)

// 3. 类型定义
type UserService struct {
    repo repository.UserRepository
    log  *zap.Logger
}

// 4. 接口实现
func NewUserService(repo repository.UserRepository, log *zap.Logger) *UserService {
    return &UserService{repo: repo, log: log}
}

// 5. 公共方法
func (s *UserService) CreateUser(ctx context.Context, req *dto.CreateUserRequest) (*model.User, error) {
    // ...
}
```

**错误处理**
```go
// ✅ 好的错误处理
user, err := s.repo.FindByID(ctx, userID)
if err != nil {
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, fmt.Errorf("%w: %d", ErrCodeUserNotFound, userID)
    }
    return nil, fmt.Errorf("failed to find user: %w", err)
}

// ❌ 不好的错误处理
user, err := s.repo.FindByID(ctx, userID)
if err != nil {
    return nil, err // 丢失了错误上下文
}
```

### 前端 TypeScript 规范

**文件组织**
```typescript
// 1. 导入
import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import type { User } from '@/types';
import { userService } from '@/services';

// 2. 类型定义
interface UserListProps {
  users: User[];
  onEdit: (user: User) => void;
}

// 3. 组件定义
export function UserList({ users, onEdit }: UserListProps) {
  // 4. Hooks
  const navigate = useNavigate();
  const [filter, setFilter] = useState('');

  // 5. 副作用
  useEffect(() => {
    // ...
  }, []);

  // 6. 事件处理
  const handleEdit = (user: User) => {
    onEdit(user);
  };

  // 7. 渲染
  return <div>{/* ... */}</div>;
}
```

---

## 安全规范

### 敏感数据处理

| 数据类型 | 处理方式 | 说明 |
|---------|---------|------|
| 密码 | bcrypt (cost=12) | 不记录日志 |
| API Key | SHA-256 哈希存储 | 仅创建时显示 |
| JWT Secret | 环境变量 | 最少 32 字符 |
| 数据库密码 | 环境变量 | 不提交到代码库 |

### 输入验证

```go
// 后端验证
func validateUserRequest(req *dto.CreateUserRequest) error {
    if req.Username == "" {
        return errors.New("username is required")
    }
    if len(req.Username) < 3 || len(req.Username) > 50 {
        return errors.New("username length must be between 3 and 50")
    }
    if req.Password == "" {
        return errors.New("password is required")
    }
    // ...
}
```

```typescript
// 前端验证
const validatePassword = (password: string): string | null => {
  if (password.length < 8) {
    return '密码至少 8 位';
  }
  if (!/[A-Z]/.test(password)) {
    return '密码必须包含大写字母';
  }
  if (!/[a-z]/.test(password)) {
    return '密码必须包含小写字母';
  }
  if (!/[0-9]/.test(password)) {
    return '密码必须包含数字';
  }
  return null;
};
```

### SQL 注入防护

```go
// ✅ 使用参数化查询
db.Where("username = ?", username).First(&user)

// ❌ 禁止字符串拼接
db.Where(fmt.Sprintf("username = '%s'", username)).First(&user)
```

### XSS 防护

```typescript
// React 自动转义，但要注意：
// ✅ 安全：自动转义
<div>{userInput}</div>

// ❌ 危险：直接渲染 HTML
<div dangerouslySetInnerHTML={{ __html: userInput }} />

// ✅ 如需渲染 HTML，使用 DOMPurify 清理
import DOMPurify from 'dompurify';
const clean = DOMPurify.sanitize(userInput);
<div dangerouslySetInnerHTML={{ __html: clean }} />
```

---

## 文档要求

### 代码文档

**必须添加文档注释的位置**:
- 所有导出的函数和方法
- 复杂的业务逻辑
- 公共 API 接口
- 配置结构和选项

```go
// CreateUser 创建新用户
//
// 此方法执行以下操作：
// 1. 验证用户名唯一性
// 2. 加密用户密码
// 3. 保存用户到数据库
// 4. 记录审计日志
//
// 参数:
//   ctx - 请求上下文
//   req - 创建用户请求
//
// 返回:
//   创建的用户信息
//   错误信息（用户名重复、部门不存在等）
func (s *UserService) CreateUser(ctx context.Context, req *dto.CreateUserRequest) (*model.User, error) {
    // ...
}
```

### API 文档

使用 Swagger 注释标注 API 接口：

```go
// CreateUser godoc
// @Summary 创建用户
// @Description 创建新用户账号（需要管理员权限）
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param request body dto.CreateUserRequest true "创建用户请求"
// @Success 200 {object} response.Response{data=model.User}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/users [post]
// @Security Bearer
```

### 更新文档

以下情况需要更新文档：
- [ ] 新增或修改 API 接口
- [ ] 变更数据库结构
- [ ] 修改配置项
- [ ] 更新部署流程

---

## 相关文档

- [开发计划](./development-plan.md) - 详细的功能设计和实现计划
- [前端规范](./frontend-standards.md) - 前端开发详细规范
- [后端规范](./backend-standards.md) - 后端开发详细规范
- [测试指南](./testing-guide.md) - 测试编写指南
- [贡献指南](../.github/CONTRIBUTING.md) - Git 工作流和 PR 规范

---

## 联系方式

- 问题反馈: [GitHub Issues](https://github.com/wskgithub/CodeMind/issues)
- 讨论区: [GitHub Discussions](https://github.com/wskgithub/CodeMind/discussions)
