# 权限控制优化

## 问题描述

测试中发现以下权限问题：

1. **部门经理和普通用户可以看到"系统管理"入口** - 只有超级管理员才应该访问
2. **部门经理可以访问"部门管理"** - 部门经理不应该管理部门，只能通过用户管理创建自己部门的用户
3. **"限额管理"对非管理员开放** - 只有超级管理员才应该访问

## 权限设计

### 角色定义

| 角色 | 英文标识 | 权限范围 |
|------|---------|---------|
| 超级管理员 | `super_admin` | 全部功能 |
| 部门经理 | `dept_manager` | 管理本部门用户 |
| 普通用户 | `user` | 个人功能 |

### 功能权限矩阵

| 功能模块 | 超级管理员 | 部门经理 | 普通用户 |
|---------|-----------|---------|---------|
| 总览 | ✅ | ✅ | ✅ |
| API Key 管理 | ✅ | ✅ | ✅ |
| 用量统计 | ✅ | ✅ | ✅ |
| 个人中心 | ✅ | ✅ | ✅ |
| **用户管理** | ✅ 全部用户 | ✅ 本部门用户 | ❌ |
| **部门管理** | ✅ | ❌ | ❌ |
| **限额管理** | ✅ | ❌ | ❌ |
| **系统管理** | ✅ | ❌ | ❌ |

## 修复内容

### 1. 优化菜单显示权限

**文件：** `frontend/src/components/layout/DashboardLayout.tsx`

**修改前：**
```typescript
const isAdmin = user?.role === 'super_admin' || user?.role === 'dept_manager';

const menuItems: MenuProps['items'] = [
  // ... 基础菜单
  ...(isAdmin
    ? [
        { type: 'divider' as const },
        { key: '/admin/users', icon: <UserOutlined />, label: '用户管理' },
        { key: '/admin/departments', icon: <TeamOutlined />, label: '部门管理' },
        { key: '/admin/limits', icon: <SafetyOutlined />, label: '限额管理' },
        { key: '/admin/system', icon: <SettingOutlined />, label: '系统管理' },
      ]
    : []),
];
```

**修改后：**
```typescript
// 权限判断
const isSuperAdmin = user?.role === 'super_admin';
const isDeptManager = user?.role === 'dept_manager';
const isAdmin = isSuperAdmin || isDeptManager;

const menuItems: MenuProps['items'] = [
  // ... 基础菜单
  // 管理员和部门经理都可以看到用户管理
  ...(isAdmin
    ? [
        { type: 'divider' as const },
        { key: '/admin/users', icon: <UserOutlined />, label: '用户管理' },
      ]
    : []),
  // 只有超级管理员可以看到部门管理、限额管理、系统管理
  ...(isSuperAdmin
    ? [
        { key: '/admin/departments', icon: <TeamOutlined />, label: '部门管理' },
        { key: '/admin/limits', icon: <SafetyOutlined />, label: '限额管理' },
        { key: '/admin/system', icon: <SettingOutlined />, label: '系统管理' },
      ]
    : []),
];
```

### 2. 增强路由守卫

**文件：** `frontend/src/router/AuthGuard.tsx`

**修改前：**
```typescript
interface AuthGuardProps {
  children: React.ReactNode;
  requireAdmin?: boolean;
}

const AuthGuard: React.FC<AuthGuardProps> = ({ children, requireAdmin = false }) => {
  // ...
  // 需要管理员权限但不是管理员或部门经理
  if (requireAdmin && user?.role === 'user') {
    return <Navigate to="/dashboard" replace />;
  }
  // ...
};
```

**修改后：**
```typescript
interface AuthGuardProps {
  children: React.ReactNode;
  requireAdmin?: boolean;
  requireSuperAdmin?: boolean;  // ✅ 新增超级管理员权限检查
}

const AuthGuard: React.FC<AuthGuardProps> = ({ 
  children, 
  requireAdmin = false,
  requireSuperAdmin = false,
}) => {
  // ...
  // 需要超级管理员权限但不是超级管理员
  if (requireSuperAdmin && user?.role !== 'super_admin') {
    return <Navigate to="/dashboard" replace />;
  }

  // 需要管理员权限但不是管理员或部门经理
  if (requireAdmin && user?.role === 'user') {
    return <Navigate to="/dashboard" replace />;
  }
  // ...
};
```

### 3. 细化路由权限配置

**文件：** `frontend/src/router/index.tsx`

**修改前：**
```typescript
{
  path: '/admin',
  element: (
    <AuthGuard requireAdmin>
      <DashboardLayout />
    </AuthGuard>
  ),
  children: [
    { index: true, element: <Navigate to="/admin/users" replace /> },
    { path: 'users', element: <UsersPage /> },
    { path: 'departments', element: <DepartmentsPage /> },
    { path: 'limits', element: <LimitsPage /> },
    { path: 'system', element: <SystemPage /> },
  ],
},
```

**修改后：**
```typescript
// 管理员和部门经理可访问的页面
{
  path: '/admin',
  element: (
    <AuthGuard requireAdmin>
      <DashboardLayout />
    </AuthGuard>
  ),
  children: [
    { index: true, element: <Navigate to="/admin/users" replace /> },
    { path: 'users', element: <UsersPage /> },
  ],
},
// 只有超级管理员可访问的页面
{
  path: '/admin',
  element: (
    <AuthGuard requireSuperAdmin>
      <DashboardLayout />
    </AuthGuard>
  ),
  children: [
    { path: 'departments', element: <DepartmentsPage /> },
    { path: 'limits', element: <LimitsPage /> },
    { path: 'system', element: <SystemPage /> },
  ],
},
```

### 4. 限制部门经理只能管理本部门用户

**文件：** `frontend/src/pages/admin/users/UsersPage.tsx`

#### 4.1 初始化时过滤部门

```typescript
const isSuperAdmin = currentUser?.role === 'super_admin';
const isDeptManager = currentUser?.role === 'dept_manager';

const [params, setParams] = useState<UserListParams>(() => {
  // 部门经理默认只查看自己部门的用户
  const initialParams: UserListParams = { page: 1, page_size: 20 };
  if (isDeptManager && currentUser?.department?.id) {
    initialParams.department_id = currentUser.department.id;
  }
  return initialParams;
});
```

#### 4.2 创建用户时锁定部门

```typescript
const handleCreate = () => {
  setEditingUser(null);
  form.resetFields();
  // 部门经理创建用户时，默认设置为自己的部门
  if (isDeptManager && currentUser?.department?.id) {
    form.setFieldsValue({ department_id: currentUser.department.id });
  }
  setModalOpen(true);
};
```

#### 4.3 表单中禁用部门选择

```typescript
<Form.Item 
  name="department_id" 
  label="所属部门"
  rules={isDeptManager ? [{ required: true, message: '请选择所属部门' }] : undefined}
>
  <Select
    placeholder={isDeptManager ? "所属部门" : "选择所属部门（可选）"}
    allowClear={!isDeptManager}
    disabled={isDeptManager}  // ✅ 部门经理不能修改部门
    showSearch
    filterOption={(input, option) =>
      (option?.label?.toString() ?? '').toLowerCase().includes(input.toLowerCase())
    }
    options={flattenDepartments(departments)}
  />
</Form.Item>
```

#### 4.4 限制角色选择

```typescript
{!editingUser && (
  <Form.Item name="role" label="角色" rules={[{ required: true, message: '请选择角色' }]}>
    <Select>
      {isSuperAdmin && <Select.Option value="super_admin">超级管理员</Select.Option>}
      {isSuperAdmin && <Select.Option value="dept_manager">部门经理</Select.Option>}
      <Select.Option value="user">普通用户</Select.Option>
    </Select>
  </Form.Item>
)}
```

#### 4.5 隐藏删除按钮

```typescript
{isSuperAdmin && (
  <Button type="link" size="small" danger icon={<DeleteOutlined />} onClick={() => handleDelete(record)}>
    删除
  </Button>
)}
```

## 修改的文件

1. `frontend/src/components/layout/DashboardLayout.tsx` - 优化菜单显示权限
2. `frontend/src/router/AuthGuard.tsx` - 增强路由守卫
3. `frontend/src/router/index.tsx` - 细化路由权限配置
4. `frontend/src/pages/admin/users/UsersPage.tsx` - 限制部门经理权限

## 安全增强

### 前端权限控制

| 层级 | 位置 | 作用 |
|------|------|------|
| **菜单层** | DashboardLayout | 隐藏无权访问的菜单项 |
| **路由层** | AuthGuard | 拦截未授权的路由访问 |
| **页面层** | UsersPage | 限制数据范围和操作权限 |
| **表单层** | UsersPage | 禁用/隐藏无权操作的字段 |

### 后端权限验证（建议）

虽然前端已做权限控制，但**后端仍需验证权限**，防止绕过前端直接调用 API：

```go
// 建议在后端添加权限中间件
func RequireSuperAdmin() gin.HandlerFunc {
    return func(c *gin.Context) {
        claims := c.MustGet("claims").(*jwt.Claims)
        if claims.Role != "super_admin" {
            c.JSON(403, gin.H{"code": 40300, "message": "需要超级管理员权限"})
            c.Abort()
            return
        }
        c.Next()
    }
}

// 应用到路由
adminGroup := router.Group("/admin")
adminGroup.Use(middleware.RequireSuperAdmin())
{
    adminGroup.GET("/departments", handler.ListDepartments)
    adminGroup.GET("/system/configs", handler.GetConfigs)
    // ...
}
```

## 测试验证

### 1. 超级管理员测试

```
登录账号：admin / Admin@123456
预期结果：
✅ 可以看到所有菜单（用户管理、部门管理、限额管理、系统管理）
✅ 可以访问所有管理页面
✅ 可以管理所有部门的用户
✅ 可以创建任意角色的用户
✅ 可以删除用户
```

### 2. 部门经理测试

```
准备工作：
1. 以超级管理员身份创建一个部门经理账号
2. 设置部门为"技术部"

登录测试：
预期结果：
✅ 只能看到"用户管理"菜单
❌ 看不到"部门管理"、"限额管理"、"系统管理"菜单
✅ 用户列表只显示"技术部"的用户
✅ 创建用户时，部门字段被锁定为"技术部"
✅ 只能创建"普通用户"角色
❌ 没有删除用户按钮
❌ 尝试访问 /admin/departments 会被重定向到 /dashboard
```

### 3. 普通用户测试

```
预期结果：
❌ 看不到任何管理菜单
❌ 尝试访问 /admin/* 会被重定向到 /dashboard
✅ 只能访问个人功能（总览、API Key、用量统计、个人中心）
```

## 用户体验优化

### 1. 部门经理创建用户流程

**优化前：**
```
1. 点击"创建用户"
2. 手动选择部门（可能选错）
3. 手动选择角色（可能越权）
4. 提交
```

**优化后：**
```
1. 点击"创建用户"
2. 部门自动设置为本部门（不可修改）
3. 角色只能选择"普通用户"
4. 提交
```

### 2. 权限不足时的提示

当用户尝试访问无权限的页面时：
- ✅ 自动重定向到 Dashboard
- ✅ 不显示错误提示（避免暴露系统结构）
- ✅ 菜单中不显示无权访问的入口

## 后续优化建议

### 1. 添加权限说明页面

在个人中心添加"权限说明"，让用户了解自己的权限范围。

### 2. 操作审计

记录所有管理操作，包括：
- 谁在什么时间
- 对哪个资源
- 执行了什么操作
- 操作结果如何

### 3. 权限变更通知

当用户权限发生变化时（如从普通用户升级为部门经理），发送通知。

### 4. 细粒度权限控制

未来可以考虑实现基于 RBAC 的更细粒度权限控制，如：
- 只读权限
- 审批权限
- 导出权限
- 等等

---

**优化时间**：2026-02-15  
**优化人员**：AI Assistant  
**测试状态**：待用户验证  
**安全等级**：高
