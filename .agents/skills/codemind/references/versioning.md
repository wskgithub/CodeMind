# 版本号管理

## 语义化版本规范

格式: `MAJOR.MINOR.PATCH`

| 版本位 | 递增时机 | 示例 |
|--------|---------|------|
| MAJOR | 不兼容的 API 变更 | 1.0.0 → 2.0.0 |
| MINOR | 向下兼容的功能新增 | 1.0.0 → 1.1.0 |
| PATCH | 向下兼容的问题修复 | 1.0.0 → 1.0.1 |

---

## 版本号规则

### MAJOR 版本（主版本）

- 不兼容的 API 变更
- 重大架构调整
- 破坏性改动

### MINOR 版本（次版本）

- 新功能添加
- 功能增强
- 向下兼容的改动

### PATCH 版本（修订版本）

- Bug 修复
- 安全补丁
- 性能优化

---

## 预发布版本

格式: `MAJOR.MINOR.PATCH-<prerelease>.<number>`

| 预发布类型 | 示例 | 说明 |
|-----------|------|------|
| alpha | 1.0.0-alpha.1 | 内部测试版本 |
| beta | 1.0.0-beta.1 | 公开测试版本 |
| rc | 1.0.0-rc.1 | 发布候选版本 |

---

## 版本更新流程

### 1. 更新 VERSION 文件

```bash
echo "0.6.0" > VERSION
```

### 2. 更新 CHANGELOG.md

```markdown
## [0.6.0] - 2026-04-05

### 新增
- 添加 xxx 功能
- 支持 yyy 特性

### 修复
- 修复 zzz 问题

### 变更
- 优化 aaa 性能

[0.6.0]: https://github.com/example/codemind/releases/tag/v0.6.0
```

### 3. 提交变更

```bash
git add VERSION CHANGELOG.md
git commit -m "chore(version): bump version to 0.6.0"
```

### 4. 创建 Git Tag

```bash
# 创建附注标签
git tag -a v0.6.0 -m "Release version 0.6.0"

# 推送标签到远程
git push origin v0.6.0
```

### 5. 创建 Release（可选）

在 GitHub 上基于 Tag 创建 Release，添加发布说明。

---

## 版本记录位置

| 文件 | 用途 | 格式 |
|------|------|------|
| `VERSION` | 当前版本号 | `0.5.0` |
| `CHANGELOG.md` | 版本变更历史 | Keep a Changelog |
| Git Tags | 发布标记 | `v0.5.0` |

---

## 发布分支流程

### 功能开发阶段

```
feature/* → develop
```

### 发布准备阶段

```
develop → release/v0.6.0
```

### 正式发布阶段

```
release/v0.6.0 → main (tag v0.6.0)
release/v0.6.0 → develop
```

### 热修复阶段

```
hotfix/* → main (tag v0.6.1)
hotfix/* → develop
```

---

## 版本变更日志格式

遵循 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.1.0/) 格式：

```markdown
# Changelog

本文件记录 CodeMind 各版本的变更内容。

格式遵循 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.1.0/)，版本号遵循 [语义化版本](https://semver.org/lang/zh-CN/)。

---

## [Unreleased]

### 新增
- 即将发布的功能

## [0.5.0] - 2026-04-05

### 新增
- 功能 A
- 功能 B

### 修复
- 问题 C

[0.5.0]: https://github.com/example/codemind/releases/tag/v0.5.0
```

### 变更类型

| 类型 | 说明 |
|------|------|
| 新增 | 新功能 |
| 变更 | 现有功能的变更 |
| 废弃 | 即将移除的功能 |
| 移除 | 已移除的功能 |
| 修复 | Bug 修复 |
| 安全 | 安全相关的修复 |
