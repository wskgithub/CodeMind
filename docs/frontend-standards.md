# CodeMind 前端开发规范

本文档定义了 CodeMind 前端的开发规范和最佳实践。

## 目录

- [技术栈](#技术栈)
- [项目结构](#项目结构)
- [组件规范](#组件规范)
- [状态管理](#状态管理)
- [API 请求](#api-请求)
- [样式规范](#样式规范)
- [测试规范](#测试规范)

---

## 技术栈

| 技术 | 版本 | 说明 |
|------|------|------|
| React | 18.x | UI 框架 |
| TypeScript | 5.x | 类型安全 |
| Vite | 6.x | 构建工具 |
| Ant Design | 5.x | UI 组件库 |
| TailwindCSS | 3.x | 原子化 CSS |
| React Router | 7.x | 路由管理 |
| Zustand | 5.x | 状态管理 |
| Axios | 1.x | HTTP 客户端 |
| ECharts | 5.x | 图表库 |
| Vitest | 2.x | 单元测试 |

---

## 项目结构

```
frontend/src/
├── main.tsx              # 应用入口
├── App.tsx               # 根组件
├── assets/               # 静态资源
│   ├── images/           # 图片资源
│   └── styles/           # 全局样式
├── components/           # 组件
│   ├── common/           # 通用组件
│   ├── layout/           # 布局组件
│   └── charts/           # 图表组件
├── pages/                # 页面组件
│   ├── home/             # 首页
│   ├── login/            # 登录页
│   ├── dashboard/        # 仪表盘
│   ├── profile/          # 个人中心
│   ├── keys/             # API Key 管理
│   ├── usage/            # 用量统计
│   └── admin/            # 管理后台
│       ├── users/        # 用户管理
│       ├── departments/  # 部门管理
│       └── ...
├── hooks/                # 自定义 Hooks
├── services/             # API 服务
├── store/                # 状态管理
├── types/                # 类型定义
├── utils/                # 工具函数
└── router/               # 路由配置
```

---

## 组件规范

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

### 组件命名规范

- **组件文件**: PascalCase (如 `UserCard.tsx`)
- **样式文件**: 与组件同名 (如 `UserCard.module.css`)
- **Hook 文件**: camelCase 以 `use` 开头 (如 `useUserData.ts`)
- **工具文件**: camelCase (如 `formatDate.ts`)

### 组件设计原则

1. **单一职责**: 每个组件只做一件事
2. **可复用性**: 通过 props 控制行为
3. **类型安全**: 所有 props 必须定义类型
4. **性能优化**: 合理使用 useMemo、useCallback

```typescript
// ✅ 好的组件设计
interface ButtonProps {
  type?: 'primary' | 'default' | 'danger';
  size?: 'small' | 'medium' | 'large';
  loading?: boolean;
  disabled?: boolean;
  onClick?: () => void;
  children: React.ReactNode;
}

export function Button({ type = 'default', size = 'medium', children, ...props }: ButtonProps) {
  return <button className={`btn btn-${type} btn-${size}`} {...props}>{children}</button>;
}

// ❌ 避免的组件设计
// 过多职责、耦合业务逻辑
```

---

## 状态管理

### Zustand Store 规范

```typescript
// store/authStore.ts
import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { User } from '@/types';

interface AuthState {
  // 状态
  token: string | null;
  user: User | null;
  isAuthenticated: boolean;

  // Actions
  login: (credentials: LoginParams) => Promise<void>;
  logout: () => void;
  updateUser: (user: User) => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      // 初始状态
      token: null,
      user: null,
      isAuthenticated: false,

      // Actions
      login: async (credentials) => {
        const response = await authService.login(credentials);
        set({ token: response.token, user: response.user, isAuthenticated: true });
      },

      logout: () => {
        set({ token: null, user: null, isAuthenticated: false });
      },

      updateUser: (user) => {
        set({ user });
      },
    }),
    {
      name: 'codemind-auth', // localStorage key
      partialize: (state) => ({ token: state.token, user: state.user }),
    }
  )
);
```

### Store 使用规范

```typescript
// ✅ 在组件中使用
import { useAuthStore } from '@/store/authStore';

function UserProfile() {
  const { user, logout } = useAuthStore();

  return <div>{user?.displayName} <button onClick={logout}>退出</button></div>;
}

// ✅ 选择性订阅（避免不必要的重渲染）
function UserAvatar() {
  const user = useAuthStore((state) => state.user);

  return <Avatar src={user?.avatarUrl} />;
}
```

---

## API 请求

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
  // 获取用户列表
  list: (params: UserListParams) =>
    request.get<PaginatedResponse<User>>('/api/v1/users', { params }),

  // 创建用户
  create: (data: CreateUserRequest) =>
    request.post<User>('/api/v1/users', data),

  // 更新用户
  update: (id: number, data: UpdateUserRequest) =>
    request.put<User>(`/api/v1/users/${id}`, data),

  // 删除用户
  delete: (id: number) =>
    request.delete(`/api/v1/users/${id}`),
};
```

### Request 配置

```typescript
// services/request.ts
import axios from 'axios';
import type { AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios';
import { message } from 'antd';
import { useAuthStore } from '@/store/authStore';

// 创建 axios 实例
const instance: AxiosInstance = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// 请求拦截器
instance.interceptors.request.use(
  (config) => {
    // 添加认证 token
    const { token } = useAuthStore.getState();
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// 响应拦截器
instance.interceptors.response.use(
  (response: AxiosResponse) => {
    const { code, data, message: msg } = response.data;

    if (code === 0) {
      return data;
    }

    // 业务错误处理
    message.error(msg || '请求失败');
    return Promise.reject(new Error(msg));
  },
  (error) => {
    // HTTP 错误处理
    if (error.response) {
      const { status } = error.response;

      switch (status) {
        case 401:
          // Token 过期，跳转登录
          useAuthStore.getState().logout();
          window.location.href = '/login';
          break;
        case 403:
          message.error('没有权限访问');
          break;
        case 404:
          message.error('请求的资源不存在');
          break;
        case 500:
          message.error('服务器错误');
          break;
        default:
          message.error(error.response.data?.message || '请求失败');
      }
    } else {
      message.error('网络错误');
    }

    return Promise.reject(error);
  }
);

export default instance;
```

---

## 样式规范

### CSS Modules

```css
/* components/common/UserCard.module.css */
.card {
  display: flex;
  align-items: center;
  padding: 16px;
  border-radius: 8px;
  background-color: var(--color-white);
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  transition: box-shadow 0.2s ease;
}

.card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.avatar {
  width: 48px;
  height: 48px;
  margin-right: 16px;
}

.info {
  flex: 1;
}

.name {
  font-size: 16px;
  font-weight: 500;
  color: var(--color-text);
}

.email {
  font-size: 14px;
  color: var(--color-text-secondary);
  margin-top: 4px;
}
```

### TailwindCSS 使用

```typescript
// 使用 TailwindCSS 的场景
<div className="flex items-center gap-4 px-4 py-3 bg-white rounded-lg shadow-sm hover:shadow-md transition-shadow">
  {/* ... */}
</div>

// 响应式设计
<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
  {/* ... */}
</div>

// 动态类名
<div className={cn(
  'base-class',
  isActive && 'active-class',
  'other-class'
)}>
  {/* ... */}
</div>
```

### 主题变量

```css
/* assets/styles/variables.css */
:root {
  /* 品牌色 */
  --color-primary: #2b7cb3;
  --color-secondary: #4ba3d4;
  --color-accent: #6bc5e8;

  /* 中性色 */
  --color-bg-layout: #f0f5fa;
  --color-white: #ffffff;
  --color-text: #1a3a5c;
  --color-text-secondary: #2e5a7e;

  /* 状态色 */
  --color-success: #52c41a;
  --color-warning: #faad14;
  --color-error: #ff4d4f;

  /* 间距 */
  --spacing-xs: 4px;
  --spacing-sm: 8px;
  --spacing-md: 16px;
  --spacing-lg: 24px;
  --spacing-xl: 32px;

  /* 圆角 */
  --radius-sm: 4px;
  --radius-md: 8px;
  --radius-lg: 12px;

  /* 阴影 */
  --shadow-sm: 0 1px 3px rgba(0, 0, 0, 0.1);
  --shadow-md: 0 4px 12px rgba(0, 0, 0, 0.15);
}
```

---

## 测试规范

### 组件测试

```typescript
// components/common/UserCard.test.tsx
import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { UserCard } from './UserCard';

describe('UserCard', () => {
  const mockUser = {
    id: 1,
    username: 'testuser',
    display_name: '测试用户',
    email: 'test@example.com',
  };

  it('should render user info correctly', () => {
    render(<UserCard user={mockUser} />);

    expect(screen.getByText('测试用户')).toBeInTheDocument();
    expect(screen.getByText('test@example.com')).toBeInTheDocument();
  });

  it('should call onEdit when edit button clicked', () => {
    const onEdit = vi.fn();

    render(<UserCard user={mockUser} onEdit={onEdit} showActions />);

    fireEvent.click(screen.getByRole('button', { name: /编辑/i }));

    expect(onEdit).toHaveBeenCalledWith(mockUser);
  });
});
```

### Hook 测试

```typescript
// hooks/useUserData.test.ts
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { useUserData } from './useUserData';
import { userService } from '@/services/userService';

vi.mock('@/services/userService');

describe('useUserData', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should fetch user data on mount', async () => {
    const mockUser = { id: 1, username: 'test' };
    vi.mocked(userService.get).mockResolvedValue(mockUser);

    const { result } = renderHook(() => useUserData(1));

    expect(result.current.loading).toBe(true);

    await waitFor(() => {
      expect(result.current.user).toEqual(mockUser);
      expect(result.current.loading).toBe(false);
    });
  });
});
```

---

## 路由规范

### 路由守卫

```typescript
// router/guards.tsx
import { Navigate } from 'react-router-dom';
import { useAuthStore } from '@/store/authStore';

interface GuardProps {
  children: React.ReactNode;
  requireAuth?: boolean;
  requireAdmin?: boolean;
}

export functionAuthGuard({ children, requireAuth = true, requireAdmin = false }: GuardProps) {
  const { isAuthenticated, user } = useAuthStore();

  if (requireAuth && !isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  if (requireAdmin && user?.role !== 'super_admin') {
    return <Navigate to="/dashboard" replace />;
  }

  return <>{children}</>;
}
```

### 路由配置

```typescript
// router/index.tsx
import { createBrowserRouter } from 'react-router-dom';
import {AuthGuard} from './guards';
import {Login} from '@/pages/login';
import {Dashboard} from '@/pages/dashboard';
import {AdminUsers} from '@/pages/admin/users';

export const router = createBrowserRouter([
  {
    path: '/login',
    element: <Login />,
  },
  {
    path: '/',
    element: <AuthGuard><DashboardLayout /></AuthGuard>,
    children: [
      { index: true, element: <Navigate to="/dashboard" replace /> },
      { path: 'dashboard', element: <Dashboard /> },
      {
        path: 'admin',
        element: <AuthGuard requireAdmin><AdminLayout /></AuthGuard>,
        children: [
          { path: 'users', element: <AdminUsers /> },
          { path: 'departments', element: <AdminDepartments /> },
        ],
      },
    ],
  },
]);
```

---

## 性能优化

### 组件优化

```typescript
// ✅ 使用 React.memo 避免不必要的重渲染
export const UserCard = React.memo(function UserCard({ user }: UserCardProps) {
  // ...
}, (prev, next) => {
  return prev.user.id === next.user.id;
});

// ✅ 使用 useCallback 缓存事件处理
const handleClick = useCallback(() => {
  // ...
}, [dependency]);

// ✅ 使用 useMemo 缓存计算结果
const fullName = useMemo(() => {
  return `${user.firstName} ${user.lastName}`;
}, [user.firstName, user.lastName]);
```

### 代码分割

```typescript
// ✅ 路由级别的代码分割
const AdminUsers = lazy(() => import('@/pages/admin/users'));

// ✅ 组件级别的代码分割
const HeavyChart = lazy(() => import('@/components/charts/HeavyChart'));
```

---

## 开发工具配置

### VSCode 设置

```json
// .vscode/settings.json
{
  "typescript.tsdk": "node_modules/typescript/lib",
  "typescript.enablePromptUseWorkspaceTsdk": true,
  "editor.formatOnSave": true,
  "editor.codeActionsOnSave": {
    "source.fixAll.eslint": "explicit"
  },
  "eslint.validate": ["javascript", "javascriptreact", "typescript", "typescriptreact"],
  "files.associations": {
    "*.css": "tailwindcss"
  }
}
```

### 推荐插件

- ESLint
- Prettier
- Tailwind CSS IntelliSense
- TypeScript Vue Plugin (Volar)
- Vitest
