# Bug 修复：用户退出登录 Panic

## 问题描述

用户点击退出登录时，后端出现 Panic 错误：

```
runtime error: invalid memory address or nil pointer dereference
```

## 错误堆栈

```
codemind/internal/pkg/jwt.(*Manager).Blacklist
    /backend/internal/pkg/jwt/jwt.go:101
codemind/internal/service.(*AuthService).Logout
    /backend/internal/service/auth.go:99
codemind/internal/handler.(*AuthHandler).Logout
    /backend/internal/handler/auth.go:57
```

## 根本原因

在 `auth.go` 的 `Logout` 方法中，传递了 `nil` 作为 context 参数给 Redis 操作：

```go
// 错误代码
func (s *AuthService) Logout(claims *jwtPkg.Claims) error {
	return s.jwtManager.Blacklist(
		nil, // ❌ 传递 nil 导致空指针错误
		claims.ID,
		claims.ExpiresAt.Time,
	)
}
```

Redis 客户端在执行 `Set` 操作时需要一个有效的 context，传递 `nil` 会导致空指针解引用错误。

## 解决方案

修改 `backend/internal/service/auth.go`：

### 1. 添加 context 包导入

```go
import (
	"context"  // ✅ 添加这一行
	"encoding/json"
	"time"
	// ... 其他导入
)
```

### 2. 修改 Logout 方法

```go
// 修复后的代码
func (s *AuthService) Logout(claims *jwtPkg.Claims) error {
	return s.jwtManager.Blacklist(
		context.Background(), // ✅ 使用有效的 context
		claims.ID,
		claims.ExpiresAt.Time,
	)
}
```

## 修改的文件

- `backend/internal/service/auth.go`
  - 添加 `context` 包导入
  - 修改 `Logout` 方法，使用 `context.Background()` 替代 `nil`

## 技术说明

### 为什么使用 context.Background()？

1. **简单场景**：退出登录是一个简单的操作，不需要复杂的 context 传递
2. **独立操作**：不依赖于请求的 context，即使请求被取消也应该完成退出
3. **标准做法**：对于不需要取消或超时控制的后台操作，使用 `context.Background()` 是标准做法

### 更好的实践（可选优化）

如果需要更好的控制，可以使用带超时的 context：

```go
func (s *AuthService) Logout(claims *jwtPkg.Claims) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	return s.jwtManager.Blacklist(
		ctx,
		claims.ID,
		claims.ExpiresAt.Time,
	)
}
```

## 测试验证

### 1. 重启后端服务

```bash
# 停止当前运行的服务（Ctrl+C）
# 重新启动
cd backend
go run cmd/server/main.go
```

### 2. 测试退出登录

1. 登录系统
2. 点击用户头像 → 退出登录
3. 验证：
   - ✅ 退出成功，跳转到登录页
   - ✅ 后端日志无 Panic 错误
   - ✅ JWT Token 被加入 Redis 黑名单

### 3. 验证 Redis 黑名单

```bash
# 连接到 Redis
docker exec -it codemind-redis redis-cli

# 查看黑名单 key
KEYS codemind:jwt:blacklist:*

# 查看某个 key 的值和 TTL
GET codemind:jwt:blacklist:{jti}
TTL codemind:jwt:blacklist:{jti}
```

## 预期日志

**修复前（错误）：**
```
ERROR runtime/panic.go:860 服务器 Panic 恢复
error: runtime error: invalid memory address or nil pointer dereference
```

**修复后（正确）：**
```
INFO gin@v1.10.0/context.go:185 请求完成
method: POST
path: /api/v1/auth/logout
status: 200
```

## 相关代码

### JWT Blacklist 方法

```go
// backend/internal/pkg/jwt/jwt.go
func (m *Manager) Blacklist(ctx context.Context, jti string, expiration time.Time) error {
	ttl := time.Until(expiration)
	if ttl <= 0 {
		return nil // 已过期，无需加入黑名单
	}

	key := fmt.Sprintf("codemind:jwt:blacklist:%s", jti)
	return m.rdb.Set(ctx, key, "1", ttl).Err() // ⚠️ 这里需要有效的 ctx
}
```

## 影响范围

- **功能**：用户退出登录
- **严重程度**：高（功能完全不可用，导致服务 Panic）
- **影响用户**：所有尝试退出登录的用户
- **修复难度**：低（一行代码修改）

## 预防措施

### 代码审查检查项

1. ✅ 所有 Redis 操作必须传递有效的 context
2. ✅ 避免传递 `nil` 作为 context 参数
3. ✅ 使用 linter 检查空指针风险

### 建议的 Linter 规则

可以添加 golangci-lint 配置检查 nil context：

```yaml
# .golangci.yml
linters-settings:
  nilnil:
    checked-types:
      - ptr
      - func
      - iface
      - map
      - chan
```

---

**修复时间**：2026-02-15  
**Bug 发现**：用户测试  
**修复状态**：✅ 已完成  
**需要重启**：是（后端服务）
