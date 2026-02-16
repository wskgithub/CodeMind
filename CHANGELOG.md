# Changelog

本文件记录 CodeMind（度影智能编码服务）各版本的变更内容。

格式遵循 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.1.0/)，版本号遵循 [语义化版本](https://semver.org/lang/zh-CN/)。

---

## [0.1.0] - 2026-02-15

### 新增

#### 后端 (Go/Gin)
- **基础架构**：分层架构搭建（Router → Middleware → Handler → Service → Repository）
- **配置管理**：Viper 配置加载，支持 YAML 文件 + 环境变量覆盖
- **认证系统**：JWT 登录/登出、Token 黑名单（Redis）、bcrypt 密码哈希
- **用户管理**：CRUD、角色管理（super_admin / dept_manager / user）、状态切换、密码重置
- **部门管理**：CRUD、树形结构查询、级联校验
- **API Key 管理**：创建（cm- 前缀 + SHA-256 哈希存储）、启用/禁用、Redis 缓存验证
- **LLM 代理层**：
  - OpenAI 兼容 API（`/v1/chat/completions`、`/v1/completions`、`/v1/models`）
  - SSE 流式响应转发
  - Token 用量计量与异步记录
  - Redis 并发控制与 Token 配额检查
- **用量统计**：总览、按日/周/月聚合查询、用户/部门排行榜
- **限额管理**：全局/部门/用户三级限额配置，优先级级联
- **系统管理**：系统配置 CRUD、公告管理、审计日志查询
- **中间件**：JWT 认证、API Key 认证、CORS、请求日志（Zap）、Panic 恢复
- **安全**：bcrypt 密码存储、API Key SHA-256 哈希、JWT 黑名单、RBAC 权限控制
- **Docker**：多阶段构建 Dockerfile

#### 前端 (React 18 + TypeScript)
- **品牌主题**：RayShape 品牌配色方案落地，Ant Design 5 主题定制
- **暗色模式**：一键切换亮/暗色主题
- **首页**：Canvas 粒子连线动效、全屏 Hero 区域、功能展示卡片（滚动动画）、数字亮点
- **登录页**：Ant Design 表单，渐变背景
- **仪表盘**：ECharts 用量趋势图表、统计卡片、系统公告展示
- **API Key 管理页**：创建/删除/启用/禁用，新 Key 一次性展示
- **用量统计页**：ECharts 可视化（堆叠柱状图 + 折线图）、日期范围筛选、排行榜
- **用户管理页**：表格列表、搜索/筛选、创建/编辑/删除/状态切换/密码重置
- **部门管理页**：树形表格展示、创建/编辑/删除
- **限额管理页**：限额配置表格、创建/删除
- **系统管理页**：三标签页（系统配置、公告管理、审计日志）
- **个人中心**：信息编辑、密码修改
- **通用组件**：EmptyState 空状态、PageLoading 加载状态
- **路由守卫**：AuthGuard（认证）、GuestGuard（游客）
- **响应式**：全平台 clamp() 适配

#### 基础设施
- Docker Compose 编排（Frontend + Backend + PostgreSQL + Redis）
- Nginx 反向代理（SPA + API + LLM SSE）
- PostgreSQL 初始化脚本（10 张表 + 索引）
- 种子数据（默认管理员、系统配置、全局限额）

#### 测试
- 后端：crypto / validator / errcode / token / LLM client / SSE stream 单元测试
- 前端：AppStore / AuthStore 状态管理测试
- Vitest + React Testing Library 测试基础设施

---

[0.1.0]: https://github.com/example/codemind/releases/tag/v0.1.0
