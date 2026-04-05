# UI/UX 指南

## 设计系统

### 品牌配色

```css
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
}
```

### 间距规范

| Token | 值 | 用途 |
|-------|-----|------|
| `--spacing-xs` | 4px | 紧凑间距 |
| `--spacing-sm` | 8px | 小间距 |
| `--spacing-md` | 16px | 标准间距 |
| `--spacing-lg` | 24px | 大间距 |
| `--spacing-xl` | 32px | 超大间距 |

### 圆角规范

| Token | 值 | 用途 |
|-------|-----|------|
| `--radius-sm` | 4px | 小元素 |
| `--radius-md` | 8px | 标准元素 |
| `--radius-lg` | 12px | 大卡片 |

### 阴影规范

```css
--shadow-sm: 0 1px 3px rgba(0, 0, 0, 0.1);
--shadow-md: 0 4px 12px rgba(0, 0, 0, 0.15);
```

---

## 组件设计原则

1. **单一职责**: 每个组件只做一件事
2. **可复用性**: 通过 props 控制行为
3. **类型安全**: 所有 props 必须定义类型
4. **性能优化**: 合理使用 useMemo、useCallback

---

## 样式使用规范

| 场景 | 推荐使用 |
|------|---------|
| 简单布局/间距 | TailwindCSS 原子类 |
| 复杂组件样式 | CSS Modules |
| 全局主题变量 | CSS 自定义属性 |

### TailwindCSS 示例

```typescript
// 基础布局
<div className="flex items-center gap-4 px-4 py-3 bg-white rounded-lg shadow-sm">

// 响应式设计
<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">

// 动态类名
<div className={cn(
  'base-class',
  isActive && 'active-class',
  'other-class'
)}>
```

### CSS Modules 示例

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
```

---

## 响应式设计

### 断点

| 断点 | 前缀 | 范围 |
|------|------|------|
| sm | `sm:` | ≥640px |
| md | `md:` | ≥768px |
| lg | `lg:` | ≥1024px |
| xl | `xl:` | ≥1280px |

### 响应式模式

```typescript
// 网格布局
<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">

// 隐藏/显示
<div className="hidden md:block">桌面端显示</div>
<div className="md:hidden">移动端显示</div>

// 字体大小
<h1 className="text-xl md:text-2xl lg:text-3xl">
```

---

## 性能优化

### 组件优化

```typescript
// 使用 React.memo 避免不必要的重渲染
export const UserCard = React.memo(function UserCard({ user }: UserCardProps) {
  // ...
}, (prev, next) => {
  return prev.user.id === next.user.id;
});

// 使用 useCallback 缓存事件处理
const handleClick = useCallback(() => {
  // ...
}, [dependency]);

// 使用 useMemo 缓存计算结果
const fullName = useMemo(() => {
  return `${user.firstName} ${user.lastName}`;
}, [user.firstName, user.lastName]);
```

### 代码分割

```typescript
// 路由级别的代码分割
const AdminUsers = lazy(() => import('@/pages/admin/users'));

// 组件级别的代码分割
const HeavyChart = lazy(() => import('@/components/charts/HeavyChart'));
```

---

## 状态管理 (Zustand)

### Store 规范

```typescript
// store/authStore.ts
import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { User } from '@/types';

interface AuthState {
  token: string | null;
  user: User | null;
  isAuthenticated: boolean;
  login: (credentials: LoginParams) => Promise<void>;
  logout: () => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      token: null,
      user: null,
      isAuthenticated: false,
      login: async (credentials) => {
        const response = await authService.login(credentials);
        set({ token: response.token, user: response.user, isAuthenticated: true });
      },
      logout: () => {
        set({ token: null, user: null, isAuthenticated: false });
      },
    }),
    {
      name: 'codemind-auth',
      partialize: (state) => ({ token: state.token, user: state.user }),
    }
  )
);
```

### 选择性订阅

```typescript
// ✅ 选择性订阅避免不必要的重渲染
function UserAvatar() {
  const user = useAuthStore((state) => state.user);
  return <Avatar src={user?.avatarUrl} />;
}
```
