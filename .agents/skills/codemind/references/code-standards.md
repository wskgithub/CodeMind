# 代码规范详情

## 通用原则

- **KISS 原则**: 保持简单，避免过度设计
- **DRY 原则**: 不要重复代码，提取公共逻辑
- **SOLID 原则**: 遵循单一职责、开闭原则等
- **可读性优先**: 代码是写给人看的，其次是机器

---

## Go 后端规范

### 命名规范

```go
// 包名：小写单词，不使用下划线
package service

// 接口：以行为命名，通常以 -er 结尾
type UserRepository interface {
    Create(ctx context.Context, user *model.User) error
    FindByID(ctx context.Context, id int64) (*model.User, error)
}

// 结构体：驼峰命名
type UserService struct {
    repo repository.UserRepository
    log  *zap.Logger
}

// 常量：驼峰命名或大写下划线
const (
    DefaultPageSize = 20
    MaxPageSize     = 100
)

// 函数：驼峰命名，导出函数首字母大写
func CreateUser(ctx context.Context, req *dto.CreateUserRequest) (*model.User, error)

// 私有函数：小写开头
func validateUser(req *dto.CreateUserRequest) error { }
```

### 注释规范（强制中文注释）

```go
// Package service 提供业务逻辑层的实现
//
// 核心职责：
//   - 实现业务规则和验证
//   - 协调多个 Repository 完成复杂业务
//   - 处理事务边界
package service

// UserService 用户服务层
//
// 负责用户相关的业务逻辑处理，包括用户创建、更新、删除、
// 状态管理等。所有操作都经过权限验证和数据校验。
type UserService struct {
    repo repository.UserRepository
    log  *zap.Logger
}

// CreateUser 创建新用户
//
// 此方法执行以下操作：
//   1. 验证用户名唯一性
//   2. 加密用户密码
//   3. 保存用户到数据库
//   4. 记录审计日志
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

### 分层架构

```
Router → Middleware → Handler → Service → Repository → Model
```

| 层级 | 职责 | 规范 |
|------|------|------|
| Handler | HTTP 请求处理 | 参数解析、调用 Service、返回响应 |
| Service | 业务逻辑 | 业务规则验证、协调多个 Repository |
| Repository | 数据访问 | 数据库操作、事务管理 |
| Model | 数据模型 | 结构体定义、表名方法 |

### 错误处理规范

```go
// 定义错误码
package errcode

const (
    Success = 0
    ErrCodeUserNotFound = 40301
    ErrCodeUserExists   = 40302
)

// 定义业务错误
var (
    ErrUserNotFound = errors.New("user not found")
    ErrUserExists   = errors.New("username already exists")
)

// Service 层返回错误
func (s *UserService) GetUser(ctx context.Context, id int64) (*model.User, error) {
    user, err := s.repo.FindByID(ctx, id)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, fmt.Errorf("%w: %d", ErrUserNotFound, id)
        }
        return nil, fmt.Errorf("failed to find user: %w", err)
    }
    return user, nil
}

// Handler 层转换错误
func (h *UserHandler) GetUser(c *gin.Context) {
    id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
    user, err := h.service.GetUser(c.Request.Context(), id)
    if err != nil {
        if errors.Is(err, service.ErrUserNotFound) {
            response.Error(c, errcode.ErrCodeUserNotFound, "用户不存在")
            return
        }
        response.Error(c, errcode.ErrCodeInternal, "获取用户信息失败")
        return
    }
    response.Success(c, user)
}
```

### 项目结构

```
backend/
├── cmd/server/main.go          # 应用入口
├── internal/
│   ├── config/                 # 配置管理
│   ├── middleware/             # 中间件
│   ├── handler/                # HTTP 处理器
│   ├── service/                # 业务逻辑
│   ├── repository/             # 数据访问
│   ├── model/                  # 数据模型
│   │   └── dto/                # 数据传输对象
│   ├── router/                 # 路由定义
│   └── pkg/                    # 内部工具包
├── pkg/                        # 外部共享包
├── migrations/                 # 数据库迁移
└── tests/                      # 测试
```

---

## TypeScript 前端规范

### 组件文件结构

```typescript
// components/common/UserCard.tsx

// 1. 导入
import { useState, useEffect } from 'react';
import { Card, Avatar } from 'antd';
import type { User } from '@/types';

// 2. 样式 (如果有)
import styles from './UserCard.module.css';

// 3. 常量定义
const DEFAULT_AVATAR = '/images/default-avatar.png';

// 4. 类型定义
interface UserCardProps {
  user: User;
  onEdit?: (user: User) => void;
  showActions?: boolean;
}

// 5. 组件定义
export function UserCard({ user, onEdit, showActions = false }: UserCardProps) {
  // 6. Hooks
  const [isHovered, setIsHovered] = useState(false);

  // 7. 副作用
  useEffect(() => {
    // ...
  }, []);

  // 8. 事件处理
  const handleEdit = () => {
    onEdit?.(user);
  };

  // 9. 渲染
  return (
    <div className={styles.card}>
      {/* ... */}
    </div>
  );
}
```

### 命名规范

- **组件文件**: PascalCase (如 `UserCard.tsx`)
- **样式文件**: 与组件同名 (如 `UserCard.module.css`)
- **Hook 文件**: camelCase 以 `use` 开头 (如 `useUserData.ts`)
- **工具文件**: camelCase (如 `formatDate.ts`)

### 项目结构

```
frontend/src/
├── main.tsx              # 应用入口
├── App.tsx               # 根组件
├── assets/               # 静态资源
├── components/           # 组件
│   ├── common/           # 通用组件
│   ├── layout/           # 布局组件
│   └── charts/           # 图表组件
├── pages/                # 页面组件
├── hooks/                # 自定义 Hooks
├── services/             # API 服务
├── store/                # 状态管理
├── types/                # 类型定义
├── utils/                # 工具函数
└── router/               # 路由配置
```

### 组件设计原则

1. **单一职责**: 每个组件只做一件事
2. **可复用性**: 通过 props 控制行为
3. **类型安全**: 所有 props 必须定义类型
4. **性能优化**: 合理使用 useMemo、useCallback

### Service 层设计

```typescript
// services/userService.ts
import request from './request';
import type { User, PaginationParams, PaginatedResponse } from '@/types';

export interface UserListParams extends PaginationParams {
  keyword?: string;
  department_id?: number;
  role?: string;
  status?: number;
}

export const userService = {
  list: (params: UserListParams) =>
    request.get<PaginatedResponse<User>>('/api/v1/users', { params }),
  
  create: (data: CreateUserRequest) =>
    request.post<User>('/api/v1/users', data),
  
  update: (id: number, data: UpdateUserRequest) =>
    request.put<User>(`/api/v1/users/${id}`, data),
  
  delete: (id: number) =>
    request.delete(`/api/v1/users/${id}`),
};
```
