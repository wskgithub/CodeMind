# 功能优化修复说明

本次修复解决了以下 6 个问题：

## 1. 修复 API Key 创建失败问题

**问题描述：** 创建 API Key 时提示数据库错误 `value too long for type character varying(10)`

**原因：** `api_keys` 表的 `key_prefix` 字段定义为 `VARCHAR(10)`，但实际生成的前缀为 `cm-48cf4808`（12 个字符）

**修复内容：**
- 修改数据库 schema：`deploy/docker/postgres/init.sql` 中将 `key_prefix` 字段长度从 `VARCHAR(10)` 改为 `VARCHAR(20)`
- 修改 Go 模型：`backend/internal/model/apikey.go` 中将 GORM tag 从 `size:10` 改为 `size:20`
- 创建迁移脚本：`deploy/docker/postgres/migrate_key_prefix.sql`

**应用迁移：**

如果是全新部署，无需额外操作，直接启动即可。

如果是已有数据库，需要手动执行迁移：

```bash
# 方式 1：使用 psql 命令行
docker exec -i codemind-postgres psql -U codemind -d codemind < deploy/docker/postgres/migrate_key_prefix.sql

# 方式 2：进入容器执行
docker exec -it codemind-postgres bash
psql -U codemind -d codemind -f /docker-entrypoint-initdb.d/migrate_key_prefix.sql
```

## 2. 优化部门管理 - 添加用户选择下拉框

**问题描述：** 创建部门时需要输入"部门经理用户 ID"，但没有说明从哪里获取 ID

**修复内容：**
- 创建部门服务：`frontend/src/services/departmentService.ts`
- 优化部门管理页面：`frontend/src/pages/admin/departments/DepartmentsPage.tsx`
  - 添加用户列表加载功能（分批加载，避免超过后端 page_size=100 限制）
  - 将"部门经理用户 ID"输入框改为下拉选择器，显示"姓名 (用户名)"格式
  - 将"上级部门 ID"输入框改为树形选择器，支持层级展示
  - 编辑部门时也支持修改部门经理

**效果：**
- 创建部门时可以从下拉列表中选择部门经理，无需手动输入 ID
- 上级部门以树形结构展示，更加直观
- 支持搜索过滤用户
- 自动处理后端分页限制，最多加载 1000 个用户

## 3. 优化用户管理 - 添加部门选择下拉框

**问题描述：** 配置用户时需要输入"部门 ID"，但没有说明从哪里获取 ID

**修复内容：**
- 优化用户管理页面：`frontend/src/pages/admin/users/UsersPage.tsx`
  - 添加部门列表加载功能
  - 将"部门 ID"输入框改为下拉选择器
  - 支持层级部门展示（如"技术部 / 研发组"）
  - 支持搜索过滤部门

**效果：**
- 创建/编辑用户时可以从下拉列表中选择部门，无需手动输入 ID
- 部门以层级路径展示（父部门 / 子部门），更加清晰
- 支持搜索过滤部门

## 4. 优化系统配置 - 添加中文标题映射

**问题描述：** 系统配置页面的输入栏标题显示的是英文变量名（如 `llm.api_key`），不够友好

**修复内容：**
- 优化系统配置页面：`frontend/src/pages/admin/system/SystemPage.tsx`
  - 添加配置项中文标题映射
  - 为每个配置项添加说明文字
  - 针对不同类型的配置项使用不同的输入控件：
    - `system.force_change_password`：使用下拉选择器（是/否）
    - `llm.models`：使用多行文本框
    - 其他：使用普通输入框
- 更新类型定义：`frontend/src/types/index.ts` 添加 `description` 字段

**效果：**
- 配置项标题显示为友好的中文名称
- 每个配置项下方显示说明文字，帮助用户理解配置含义
- 不同类型的配置使用更合适的输入控件

## 配置项中文映射表

| 配置键 | 中文名称 | 说明 |
|--------|---------|------|
| `llm.api_key` | LLM API 密钥 | 用于访问 LLM 服务的 API 密钥 |
| `llm.base_url` | LLM 服务地址 | LLM 服务的基础 URL |
| `llm.default_model` | 默认模型 | 系统默认使用的 LLM 模型名称 |
| `llm.models` | 可用模型列表 | 支持的模型列表（JSON 数组格式）|
| `system.default_concurrency` | 默认并发数 | 用户默认的最大并发请求数 |
| `system.force_change_password` | 强制修改密码 | 用户首次登录是否强制修改密码 |
| `system.max_keys_per_user` | 每用户最大密钥数 | 每个用户可创建的最大 API Key 数量 |
| `system.site_name` | 站点名称 | 系统显示的站点名称 |
| `system.site_logo` | 站点 Logo | 站点 Logo 的 URL |
| `system.contact_email` | 联系邮箱 | 系统管理员联系邮箱 |

## 测试建议

1. **API Key 创建测试**
   - 应用数据库迁移后，尝试创建新的 API Key
   - 验证创建成功且前缀正常显示

2. **部门管理测试**
   - 创建新部门，选择部门经理和上级部门
   - 编辑现有部门，修改部门经理
   - 验证下拉列表正常显示用户和部门信息

3. **用户管理测试**
   - 创建新用户，从下拉列表选择部门
   - 编辑现有用户，修改所属部门
   - 验证层级部门路径正确显示

4. **系统配置测试**
   - 访问系统配置页面
   - 验证所有配置项标题显示为中文
   - 验证说明文字正确显示
   - 修改配置并保存，验证功能正常

## 5. 修复用户退出登录 Panic 错误

**问题描述：** 用户点击退出登录时，后端出现 `runtime error: invalid memory address or nil pointer dereference` 错误

**原因：** `auth.go` 中的 `Logout` 方法传递了 `nil` 作为 context 参数给 Redis 操作，导致空指针错误

**修复内容：**
- 修改后端服务：`backend/internal/service/auth.go`
  - 添加 `context` 包导入
  - 将 `Logout` 方法中的 `nil` 改为 `context.Background()`

**效果：**
- 用户可以正常退出登录
- JWT Token 正确加入 Redis 黑名单
- 不再出现服务 Panic 错误

## 6. 修复登录失败时前端无错误提示

**问题描述：** 用户输入错误的用户名或密码时，后端返回 401 错误，但前端没有显示任何错误提示

**原因：** `request.ts` 的 401 错误处理逻辑有问题，在登录页面时不显示错误消息

**修复内容：**
- 修改前端服务：`frontend/src/services/request.ts`
  - 区分登录失败（登录页面）和 Token 过期（其他页面）两种 401 场景
  - 登录页面显示具体错误消息，其他页面跳转到登录页
- 优化前端 Store：`frontend/src/store/authStore.ts`
  - 保留原始错误信息，不覆盖错误消息

**效果：**
- 登录失败时显示明确的错误提示（如"用户名或密码错误"）
- Token 过期时正常跳转到登录页
- 用户体验显著提升

## 技术说明

### 为什么用户列表要分批加载？

后端对 `page_size` 有最大值限制（100），这是合理的安全措施。前端需要加载所有用户用于下拉选择时，采用分批加载策略：
- 每次请求 100 条数据
- 循环请求直到返回数据少于 100 条（最后一页）
- 设置安全上限为 10 页（1000 个用户）
- 避免一次性请求大量数据导致后端返回 400 错误

### 后端参数验证

```go
type UserListQuery struct {
    Page     int `form:"page" binding:"omitempty,min=1"`
    PageSize int `form:"page_size" binding:"omitempty,min=1,max=100"`
    // ...
}
```

## 部署步骤

### 开发环境

```bash
# 1. 应用数据库迁移（如果是已有数据库）
docker exec -i codemind-postgres psql -U codemind -d codemind < deploy/docker/postgres/migrate_key_prefix.sql

# 2. 重启后端服务（如果需要）
cd backend
go run cmd/server/main.go

# 3. 重新构建前端（如果需要）
cd frontend
npm run build
```

### 生产环境

```bash
# 1. 应用数据库迁移
docker exec -i codemind-postgres psql -U codemind -d codemind < deploy/docker/postgres/migrate_key_prefix.sql

# 2. 重新构建并启动服务
docker compose down
docker compose up -d --build
```

## 注意事项

1. 数据库迁移是向后兼容的，不会影响现有数据
2. 前端页面优化不影响后端 API，可以独立部署
3. 如果添加新的系统配置项，记得在 `SystemPage.tsx` 中添加对应的中文映射
4. 部门和用户的下拉列表会加载所有数据，如果数据量很大，建议后续优化为分页加载或搜索接口
