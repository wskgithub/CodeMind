# 项目安全管理

## 密码安全

- **哈希算法**: bcrypt
- **Cost Factor**: 12
- **传输**: 前端通过 HTTPS 传输，后端不记录明文密码

```go
// 密码哈希
passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), 12)

// 密码验证
err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
```

---

## API Key 安全

| 属性 | 规范 |
|------|------|
| 格式 | `cm-{32-char-hex}` |
| 存储 | SHA-256 哈希，不存储明文 |
| 展示 | 仅创建时返回完整密钥，之后仅展示前缀 |
| 日志 | 禁止在日志中记录完整 API Key |
| 缓存 | Redis 缓存哈希后的 Key 信息（300s TTL） |

```go
// API Key 生成
apiKey := "cm-" + generateRandomHex(32)

// 存储哈希
keyHash := sha256.Sum256([]byte(apiKey))
```

---

## JWT 安全

| 属性 | 规范 |
|------|------|
| 算法 | HS256 |
| 有效期 | 24 小时 |
| 黑名单 | 登出或密码修改时加入 Redis 黑名单 |
| 传递 | `Authorization: Bearer <token>` |

---

## 第三方 API Key 安全

- **加密算法**: AES（`internal/pkg/crypto/aes.go`）
- **存储**: 第三方服务商的 API Key 加密后存储于数据库
- **解密**: 仅在代理请求时解密，不在日志或 API 响应中暴露

---

## 输入验证

### 后端验证

```go
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

### 前端验证

```typescript
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

---

## SQL 注入防护

```go
// ✅ 使用参数化查询
db.Where("username = ?", username).First(&user)

// ❌ 禁止字符串拼接
db.Where(fmt.Sprintf("username = '%s'", username)).First(&user)
```

---

## XSS 防护

```typescript
// ✅ React 自动转义
<div>{userInput}</div>

// ❌ 危险：直接渲染 HTML
<div dangerouslySetInnerHTML={{ __html: userInput }} />

// ✅ 如需渲染 HTML，使用 DOMPurify 清理
import DOMPurify from 'dompurify';
const clean = DOMPurify.sanitize(userInput);
<div dangerouslySetInnerHTML={{ __html: clean }} />
```

---

## 审计日志

所有敏感操作必须记录审计日志，包含：
- 操作者信息（用户 ID、用户名）
- 操作者 IP 地址
- 操作时间戳
- 操作详情（操作类型、目标对象、变更内容）

### 敏感操作列表

- 用户创建/更新/删除
- 密码修改/重置
- API Key 创建/删除
- 权限变更
- 系统配置修改
- 登录/登出

---

## 登录锁定

失败的登录尝试触发指数退避锁定：
- 首次失败：5 分钟
- 持续失败：时间指数增长
- 最大锁定时间：24 小时

---

## 安全编码检查清单

- [ ] 所有输入都经过验证
- [ ] 使用参数化查询防止 SQL 注入
- [ ] 敏感数据加密存储
- [ ] 密码使用 bcrypt 哈希
- [ ] API Key 使用 SHA-256 哈希
- [ ] JWT 使用足够强度的密钥（≥32字符）
- [ ] 敏感操作记录审计日志
- [ ] 禁止在日志中记录敏感信息
- [ ] 响应中不包含敏感数据
- [ ] 权限控制正确实现
