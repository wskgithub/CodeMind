# CodeMind 贡献指南

本文档定义了 CodeMind 项目的开发规范和最佳实践。

## 目录

- [代码规范](#代码规范)
- [Git 工作流](#git-工作流)
- [Commit 规范](#commit-规范)
- [代码审查](#代码审查)
- [测试要求](#测试要求)

---

## 代码规范

### Go 后端规范

**代码风格**
- 遵循 [Effective Go](https://go.dev/doc/effective_go) 指南
- 使用 `gofmt` 格式化代码
- 使用 `golangci-lint` 进行静态检查

**命名规范**
- 包名：小写单词，不使用下划线或驼峰
- 接口名：以 `-er` 结尾的方法命名接口
- 常量：驼峰命名，导出常量首字母大写

**注释规范**
- 所有导出函数必须有文档注释
- 包注释放在 `doc.go` 文件中
- 使用 godoc 格式

```go
// UserService 用户服务层
// 负责用户相关的业务逻辑处理
type UserService struct {
    repo repository.UserRepository
}

// CreateUser 创建新用户
// 参数: ctx - 请求上下文, req - 创建请求
// 返回: 创建的用户信息和错误
func (s *UserService) CreateUser(ctx context.Context, req *dto.CreateUserRequest) (*model.User, error) {
    // ...
}
```

**错误处理**
- 不要忽略错误
- 使用 `errors.Wrap` 包装错误上下文
- 定义明确的错误码和错误消息

### TypeScript 前端规范

**代码风格**
- 遵循 [Airbnb TypeScript Style Guide](https://github.com/airbnb/typescript)
- 使用 ESLint + Prettier 格式化
- 组件使用函数式组件 + Hooks

**命名规范**
- 组件：PascalCase (如 `UserList.tsx`)
- 工具函数：camelCase (如 `formatDate.ts`)
- 常量：UPPER_SNAKE_CASE
- 类型/接口：PascalCase (如 `UserData`)

**组件规范**
```typescript
// ✅ 组件文件结构
// 1. 导入
import { useState } from 'react';
import type { User } from '@/types';

// 2. 类型定义
interface UserListProps {
  users: User[];
  onEdit: (user: User) => void;
}

// 3. 组件定义
export function UserList({ users, onEdit }: UserListProps) {
  // 4. Hooks
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
  return (
    <div className="user-list">
      {/* ... */}
    </div>
  );
}
```

---

## Git 工作流

### 分支模型

```
main (生产分支)
  ↑
  ├── develop (开发分支)
       ↑
       ├── feature/* (功能分支)
       ├── fix/* (修复分支)
       └── release/* (发布分支)
```

### 分支命名规范

| 分支类型 | 命名格式 | 示例 |
|---------|----------|------|
| 功能开发 | `feature/<模块>-<功能>` | `feature/user-management` |
| Bug 修复 | `fix/<模块>-<问题>` | `fix/auth-login-validation` |
| 发布准备 | `release/<版本号>` | `release/0.1.0` |
| 热修复 | `hotfix/<版本号>-<问题>` | `hotfix/1.0.0-security-fix` |

### 分支操作流程

1. **从 develop 创建功能分支**
   ```bash
   git checkout develop
   git pull origin develop
   git checkout -b feature/user-management
   ```

2. **开发并提交**
   ```bash
   git add .
   git commit -m "feat(user): add user creation API"
   ```

3. **推送到远程**
   ```bash
   git push origin feature/user-management
   ```

4. **创建 Pull Request**
   - 目标分支: `develop`
   - 填写 PR 描述模板
   - 请求代码审查

5. **合并后清理分支**
   ```bash
   git checkout develop
   git pull origin develop
   git branch -d feature/user-management
   ```

---

## Commit 规范

### Conventional Commits

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Type 类型

| Type | 说明 | 示例 |
|------|------|------|
| `feat` | 新功能 | `feat(user): add user creation API` |
| `fix` | Bug 修复 | `fix(auth): resolve token expiration issue` |
| `docs` | 文档变更 | `docs(readme): update deployment instructions` |
| `style` | 代码格式（不影响功能） | `style(backend): format code with prettier` |
| `refactor` | 代码重构 | `refactor(service): extract common validation logic` |
| `perf` | 性能优化 | `perf(cache): add redis caching for user data` |
| `test` | 测试相关 | `test(auth): add login unit tests` |
| `chore` | 构建/工具变更 | `chore(deps): upgrade gin framework to v1.10.0` |

### Scope 范围

**后端**: `auth`, `user`, `department`, `apikey`, `stats`, `limit`, `system`, `llm`, `db`
**前端**: `login`, `dashboard`, `admin`, `components`, `api`, `style`
**部署**: `docker`, `deploy`, `ci`

### Commit 示例

```bash
# 功能
git commit -m "feat(user): add batch import users from CSV

- Parse CSV file with validation
- Create users in batch with transaction
- Return import summary with success/failure count

Closes #123"

# Bug 修复
git commit -m "fix(auth): fix JWT blacklist not working

Use JTI instead of full token for blacklist key to avoid
URL encoding issues.

Fixes #145"
```

---

## 代码审查

### PR 审查清单

**功能完整性**
- [ ] 功能符合需求描述
- [ ] 边界情况已处理
- [ ] 错误处理完善

**代码质量**
- [ ] 代码风格符合规范
- [ ] 命名清晰易懂
- [ ] 没有重复代码
- [ ] 注释恰当且准确

**测试覆盖**
- [ ] 单元测试覆盖率 > 80%
- [ ] 关键路径有集成测试
- [ ] 测试用例完整

**安全性**
- [ ] 输入验证完整
- [ ] 敏感数据加密存储
- [ ] SQL/命令注入防护
- [ ] 权限检查正确

### 审查反馈规范

**建设性反馈**
- 指出问题时同时说明原因
- 提供改进建议
- 对于代码风格问题，直接修改或指明规范位置

**标签使用**
- `LGTM` (Looks Good To Me): 审查通过
- `Request Changes`: 需要修改后再次审查
- `Concept ACK`: 设计方向认可，细节待定

---

## 测试要求

### 后端测试

**单元测试**
- Service 层覆盖率 > 80%
- Repository 层使用 mock 数据
- 每个函数至少一个正常路径 + 一个异常路径

```go
func TestUserService_CreateUser(t *testing.T) {
    tests := []struct {
        name    string
        req     *dto.CreateUserRequest
        want    *model.User
        wantErr bool
    }{
        {
            name: "正常创建用户",
            req:  &dto.CreateUserRequest{Username: "test", ...},
            want: &model.User{ID: 1, Username: "test", ...},
        },
        {
            name:    "用户名重复应报错",
            req:     &dto.CreateUserRequest{Username: "existing", ...},
            wantErr: true,
        },
    }
    // ...
}
```

### 前端测试

**组件测试**
- 使用 React Testing Library
- 测试用户交互而非实现细节
- 模拟 API 调用

```typescript
describe('LoginForm', () => {
  it('should show error on failed login', async () => {
    const mockLogin = vi.fn().mockRejectedValue(new Error('Invalid credentials'));
    render(<LoginForm onLogin={mockLogin} />);

    await userEvent.click(screen.getByRole('button', { name: /登录/i }));

    expect(await screen.findByText('用户名或密码错误')).toBeInTheDocument();
  });
});
```

---

## 安全规范

### 密码安全
- 使用 bcrypt 加密，cost factor = 12
- 密码最少 8 位，包含大小写字母和数字
- 禁止在日志中记录密码

### API Key 安全
- Key 仅在创建时显示完整值
- 数据库存储 SHA-256 哈希值
- 日志中只记录 key_prefix

### JWT 安全
- 使用 HS256 算法签名
- Token 有效期 24 小时
- 登出时加入黑名单

### 输入验证
- 所有用户输入必须验证
- 使用白名单而非黑名单
- SQL 查询使用参数化

---

## 资源链接

- [项目开发计划](../docs/development-plan.md)
- [API 文档](../docs/api/openapi.yaml)
- [CLAUDE.md](../CLAUDE.md)
