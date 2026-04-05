# Git 工作流规范

> **重要**: 本项目采用基于 rebase 的 Git 工作流，确保提交历史线性清晰。

---

## 分支模型

```
master (生产环境 - 受保护)
  ↑
  develop (开发分支)
    ↑
    feature/* (功能分支)
    fix/* (修复分支)
    refactor/* (重构分支)
```

| 分支 | 用途 | 生命周期 |
|------|------|---------|
| `master` | 生产环境代码 | 长期存在，受保护 |
| `develop` | 开发集成分支 | 长期存在 |
| `feature/*` | 新功能开发 | 合并后删除 |
| `fix/*` | Bug 修复 | 合并后删除 |
| `refactor/*` | 代码重构 | 合并后删除 |
| `docs/*` | 文档更新 | 合并后删除 |

---

## 分支命名规范

| 类型 | 命名格式 | 示例 |
|------|---------|------|
| 功能分支 | `feature/功能描述` | `feature/user-management` |
| 修复分支 | `fix/问题描述` | `fix/login-timeout` |
| 重构分支 | `refactor/重构描述` | `refactor/user-service` |
| 文档分支 | `docs/文档描述` | `docs/api-documentation` |

**命名要求**:
- 使用小写字母
- 单词间使用连字符 `-`
- 简洁明了，见名知意

---

## 提交信息规范

### 格式

```
<type>(<scope>): <subject>

<body> (可选)

<footer> (可选)
```

### Type 类型

| Type | 说明 |
|------|------|
| `feat` | 新功能 |
| `fix` | 修复问题 |
| `docs` | 文档更新 |
| `style` | 代码格式调整（不影响功能） |
| `refactor` | 代码重构 |
| `perf` | 性能优化 |
| `test` | 测试相关 |
| `chore` | 构建/工具相关 |
| `ci` | CI/CD 相关 |
| `revert` | 回滚提交 |

### 示例

```bash
# 功能提交
git commit -m "feat(user): add user creation API"
git commit -m "feat(auth): implement JWT refresh token"

# 修复提交
git commit -m "fix(auth): resolve token expiration issue"
git commit -m "fix(db): fix connection pool leak"

# 文档提交
git commit -m "docs(api): update swagger documentation"

# 重构提交
git commit -m "refactor(user): extract validation logic"
```

---

## 开发工作流（Develop 分支）

### 1. 创建功能分支

从 `develop` 分支创建：

```bash
# 切换到 develop 分支并更新
git checkout develop
git pull origin develop

# 创建功能分支
git checkout -b feature/user-management
```

### 2. 开发并提交

```bash
# 开发代码，进行多次提交
git add .
git commit -m "feat(user): add user creation API"

# 继续开发
git add .
git commit -m "feat(user): add username validation"

# 推送到远程
git push origin feature/user-management
```

### 3. 开发完成，准备合并（**关键步骤**）

开发完成后，**必须经用户确认**，然后按以下流程操作：

#### 3.1 先将 develop 分支的更新 rebase 到功能分支

```bash
# 切换到 develop 并拉取最新代码
git checkout develop
git pull origin develop

# 切换回功能分支
git checkout feature/user-management

# 将 develop 的变更 rebase 到当前分支
git rebase develop

# 如有冲突，解决后
git add .
git rebase --continue

# 强制推送（因为 rebase 修改了提交历史）
git push --force-with-lease origin feature/user-management
```

#### 3.2 将功能分支合并到 develop

```bash
# 切换到 develop 分支
git checkout develop

# 合并功能分支（使用 no-ff 保留分支信息，或使用普通合并）
git merge --no-ff feature/user-management

# 推送到远程
git push origin develop
```

#### 3.3 删除功能分支

```bash
# 删除本地分支
git branch -d feature/user-management

# 删除远程分支
git push origin --delete feature/user-management
```

---

## 发布工作流（Master 分支）

> ⚠️ **警告**: Master 分支的合并需要**用户主动要求**，不要自行操作！

### 合并流程

当用户要求发布到 master 时：

#### 1. 先将 master 分支的更新 rebase 到 develop

```bash
# 切换到 master 并拉取最新代码
git checkout master
git pull origin master

# 切换到 develop 分支
git checkout develop

# 将 master 的变更 rebase 到 develop
git rebase master

# 如有冲突，解决后
git add .
git rebase --continue

# 强制推送（因为 rebase 修改了提交历史）
git push --force-with-lease origin develop
```

#### 2. 将 develop 分支压缩合并到 master

```bash
# 切换到 master 分支
git checkout master

# 使用 squash 合并 develop 分支（将所有提交压缩为一个）
git merge --squash develop

# 创建压缩后的提交
git commit -m "release: merge develop into master for v0.x.x"

# 推送到远程
git push origin master
```

#### 3. 创建版本标签

```bash
# 创建标签
git tag -a v0.6.0 -m "Release version 0.6.0"

# 推送标签到远程
git push origin v0.6.0
```

---

## 提交前检查清单

```bash
# 后端检查
cd backend
go test ./...
make lint

# 前端检查
cd frontend
npm test
npm run lint
```

---

## 快速参考

### 功能开发流程

```bash
# 1. 创建分支
git checkout develop && git pull
git checkout -b feature/xxx

# 2. 开发并提交（多次提交）
git add . && git commit -m "feat: xxx"

# 3. 准备合并 - rebase develop
git checkout develop && git pull
git checkout feature/xxx
git rebase develop
# 解决冲突...
git push --force-with-lease

# 4. 合并到 develop
git checkout develop
git merge --no-ff feature/xxx
git push

# 5. 删除分支
git branch -d feature/xxx
git push origin --delete feature/xxx
```

### 发布流程（需用户确认）

```bash
# 1. rebase master 到 develop
git checkout master && git pull
git checkout develop
git rebase master
# 解决冲突...
git push --force-with-lease

# 2. squash 合并到 master（需用户确认）
git checkout master
git merge --squash develop
git commit -m "release: v0.x.x"
git push

# 3. 打标签
git tag -a v0.x.x -m "Release v0.x.x"
git push origin v0.x.x
```

---

## 注意事项

1. **永远不要直接修改 master 分支**，只能通过 squash merge 从 develop 合并
2. **rebase 会修改提交历史**，强制推送前确认无误
3. **功能分支合并后必须删除**，保持仓库整洁
4. **master 合并需要用户主动要求**，不要自行操作
5. **解决冲突时**，仔细确认保留正确的代码版本
