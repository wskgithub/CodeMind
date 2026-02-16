# Bug 修复：登录失败时前端无错误提示

## 问题描述

用户在登录页面输入错误的用户名或密码时，后端返回 401 错误，但前端没有显示任何错误提示，用户体验不佳。

## 错误日志

**后端日志：**
```
POST /api/v1/auth/login
Status: 401 Unauthorized
Body: {"code": 40101, "message": "用户名或密码错误"}
```

**前端表现：**
- ❌ 没有错误提示
- ❌ 用户不知道为什么登录失败
- ❌ 登录按钮只是停止 loading 状态

## 根本原因

在 `request.ts` 的响应拦截器中，401 错误的处理逻辑有问题：

```typescript
case 401:
  // Token 无效或过期，清除登录状态并跳转
  localStorage.removeItem('token');
  localStorage.removeItem('user');
  // 避免重复跳转
  if (window.location.pathname !== '/login') {
    message.error('登录已过期，请重新登录');
    window.location.href = '/login';
  }
  break;
```

**问题分析：**
1. 代码检查了 `window.location.pathname !== '/login'`，意图是避免在登录页重复跳转
2. 但这导致在登录页面时，401 错误不会显示任何消息
3. 用户输入错误密码时，看不到任何错误提示

## 解决方案

修改 `frontend/src/services/request.ts` 和 `frontend/src/store/authStore.ts`：

### 1. 修复 request.ts 的 401 错误处理

**修改前（错误）：**
```typescript
case 401:
  localStorage.removeItem('token');
  localStorage.removeItem('user');
  if (window.location.pathname !== '/login') {
    message.error('登录已过期，请重新登录');
    window.location.href = '/login';
  }
  break;
```

**修改后（正确）：**
```typescript
case 401:
  // 区分登录失败和 Token 过期
  if (window.location.pathname === '/login') {
    // 登录页面的 401 错误（用户名或密码错误）
    message.error(data?.message || '用户名或密码错误');
  } else {
    // 其他页面的 401 错误（Token 无效或过期）
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    message.error('登录已过期，请重新登录');
    window.location.href = '/login';
  }
  break;
```

### 2. 优化 authStore.ts 的错误处理

**修改前：**
```typescript
} catch {
  set({ loading: false });
  throw new Error('登录失败');  // ❌ 丢失了原始错误信息
}
```

**修改后：**
```typescript
} catch (error) {
  set({ loading: false });
  // 向上抛出错误，让调用方处理（错误消息已在 request 拦截器中显示）
  throw error;  // ✅ 保留原始错误
}
```

## 修改的文件

- `frontend/src/services/request.ts` - 修复 401 错误处理逻辑
- `frontend/src/store/authStore.ts` - 优化错误抛出

## 改进点

### 1. 区分两种 401 场景

| 场景 | 位置 | 原因 | 处理方式 |
|------|------|------|----------|
| 登录失败 | 登录页面 | 用户名或密码错误 | 显示错误消息，不跳转 |
| Token 过期 | 其他页面 | Token 无效或过期 | 清除状态，跳转登录页 |

### 2. 显示后端返回的具体错误消息

- 使用 `data?.message` 获取后端返回的具体错误信息
- 提供默认消息作为降级方案

### 3. 保留原始错误信息

- 不在 authStore 中覆盖错误消息
- 让 request 拦截器统一处理错误显示

## 测试验证

### 1. 测试登录失败

```
步骤：
1. 打开登录页面
2. 输入错误的用户名或密码
3. 点击登录

预期结果：
✅ 显示错误提示："用户名或密码错误"
✅ 不清除 localStorage
✅ 不跳转页面
✅ 用户可以重新输入
```

### 2. 测试 Token 过期

```
步骤：
1. 登录成功后，手动修改 localStorage 中的 token 为无效值
2. 刷新页面或访问任意需要认证的页面

预期结果：
✅ 显示错误提示："登录已过期，请重新登录"
✅ 清除 localStorage
✅ 跳转到登录页
```

### 3. 测试正常登录

```
步骤：
1. 输入正确的用户名和密码（admin / Admin@123456）
2. 点击登录

预期结果：
✅ 显示成功提示："登录成功"
✅ 跳转到 Dashboard
✅ 不显示任何错误消息
```

## 用户体验改进

### 修复前
```
用户输入错误密码
↓
点击登录按钮
↓
按钮显示 loading
↓
loading 消失
↓
❌ 没有任何提示
↓
用户困惑：为什么登录不了？
```

### 修复后
```
用户输入错误密码
↓
点击登录按钮
↓
按钮显示 loading
↓
loading 消失
↓
✅ 显示错误提示："用户名或密码错误"
↓
用户知道原因，重新输入
```

## 相关代码

### 后端返回的错误格式

```go
// backend/internal/pkg/errcode/codes.go
ErrInvalidCredentials = &Error{
    Code:    40101,
    Message: "用户名或密码错误",
}
```

### 前端 API 响应类型

```typescript
interface ApiResponse<T = any> {
  code: number;
  message: string;
  data?: T;
}
```

## 后续优化建议

### 1. 添加登录失败次数限制

```typescript
// 记录登录失败次数
let loginFailCount = 0;

if (status === 401 && window.location.pathname === '/login') {
  loginFailCount++;
  
  if (loginFailCount >= 3) {
    message.error('登录失败次数过多，请稍后再试');
    // 可以添加验证码或锁定机制
  } else {
    message.error(data?.message || '用户名或密码错误');
  }
}
```

### 2. 添加密码强度提示

在登录页面添加密码输入提示，帮助用户记忆密码格式。

### 3. 添加"忘记密码"功能

提供密码重置功能，提升用户体验。

---

**修复时间**：2026-02-15  
**Bug 发现**：用户测试  
**修复状态**：✅ 已完成  
**需要重启**：否（前端自动热重载）
