# 紧急修复：用户列表 400 错误

## 问题描述

前端在加载用户列表时出现 400 错误：

```
GET /api/v1/users?page=1&page_size=1000
Status: 400 Bad Request
```

## 根本原因

后端对 `page_size` 参数有验证限制，最大值为 100：

```go
type UserListQuery struct {
    PageSize int `form:"page_size" binding:"omitempty,min=1,max=100"`
}
```

前端为了加载所有用户用于下拉选择，直接请求了 `page_size=1000`，触发了后端参数验证失败。

## 解决方案

修改前端代码，采用**分批加载**策略：

### 修改前（错误）

```typescript
const loadUsers = useCallback(async () => {
  try {
    const resp = await userService.list({ page: 1, page_size: 1000 });
    setUsers(resp.data.data.list || []);
  } catch {
    // 错误已在拦截器中处理
  }
}, []);
```

### 修改后（正确）

```typescript
const loadUsers = useCallback(async () => {
  try {
    // 后端限制 page_size 最大为 100，这里分批加载
    const allUsers: UserDetail[] = [];
    let page = 1;
    const pageSize = 100;
    
    while (true) {
      const resp = await userService.list({ page, page_size: pageSize });
      const users = resp.data.data.list || [];
      allUsers.push(...users);
      
      // 如果返回的数据少于 pageSize，说明已经是最后一页
      if (users.length < pageSize) {
        break;
      }
      page++;
      
      // 安全限制：最多加载 10 页（1000 个用户）
      if (page > 10) {
        break;
      }
    }
    
    setUsers(allUsers);
  } catch {
    // 错误已在拦截器中处理
  }
}, []);
```

## 修改的文件

- `frontend/src/pages/admin/departments/DepartmentsPage.tsx` - 修复用户列表加载逻辑

## 优点

1. **遵守后端限制**：不违反后端的参数验证规则
2. **性能优化**：分批加载，避免一次性加载大量数据
3. **安全限制**：设置最大页数限制，防止无限循环
4. **向后兼容**：不需要修改后端代码

## 测试验证

1. 打开部门管理页面
2. 点击"创建部门"按钮
3. 查看"部门经理"下拉框是否正常显示用户列表
4. 检查浏览器控制台，确认没有 400 错误
5. 检查后端日志，确认请求正常（多次 page_size=100 的请求）

## 预期日志

**修复前（错误）：**
```
GET /api/v1/users?page=1&page_size=1000 → 400 Bad Request
```

**修复后（正确）：**
```
GET /api/v1/users?page=1&page_size=100 → 200 OK (返回 100 条)
GET /api/v1/users?page=2&page_size=100 → 200 OK (返回 50 条，最后一页)
```

## 后续优化建议

如果用户数量持续增长（超过 1000），建议：

1. **创建专用 API**：提供简化的用户列表接口（只返回 id、username、display_name）
2. **实现远程搜索**：下拉框支持输入关键字后端搜索，而不是前端过滤
3. **添加缓存**：对用户列表进行前端缓存，避免重复加载

示例专用 API：

```go
// GET /api/v1/users/simple
type SimpleUser struct {
    ID          int64  `json:"id"`
    Username    string `json:"username"`
    DisplayName string `json:"display_name"`
}
```

---

**修复时间**：2026-02-15  
**影响范围**：部门管理页面  
**严重程度**：中等（功能不可用）  
**修复状态**：✅ 已完成
