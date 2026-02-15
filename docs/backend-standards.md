# CodeMind 后端开发规范

本文档定义了 CodeMind 后端的开发规范和最佳实践。

## 目录

- [技术栈](#技术栈)
- [项目结构](#项目结构)
- [代码规范](#代码规范)
- [分层架构](#分层架构)
- [错误处理](#错误处理)
- [数据库操作](#数据库操作)
- [中间件开发](#中间件开发)
- [测试规范](#测试规范)

---

## 技术栈

| 技术 | 版本 | 用途 |
|------|------|------|
| Go | 1.23.x | 后端语言 |
| Gin | 1.10.x | Web 框架 |
| GORM | 1.25.x | ORM 框架 |
| Viper | 1.19.x | 配置管理 |
| Zap | 1.27.x | 结构化日志 |
| golang-jwt | 5.x | JWT 认证 |

---

## 项目结构

```
backend/
├── cmd/
│   └── server/
│       └── main.go              # 应用入口
├── internal/
│   ├── config/                  # 配置管理
│   │   └── config.go
│   ├── middleware/              # 中间件
│   │   ├── auth.go              # JWT 认证
│   │   ├── cors.go              # 跨域处理
│   │   ├── logger.go            # 请求日志
│   │   ├── recovery.go          # 恢复中间件
│   │   ├── ratelimit.go         # 限流
│   │   └── apikey.go            # API Key 认证
│   ├── handler/                 # HTTP 处理器
│   │   ├── auth.go
│   │   ├── user.go
│   │   ├── department.go
│   │   ├── apikey.go
│   │   ├── stats.go
│   │   ├── limit.go
│   │   ├── llm_proxy.go
│   │   └── system.go
│   ├── service/                 # 业务逻辑
│   │   ├── auth.go
│   │   ├── user.go
│   │   ├── department.go
│   │   ├── apikey.go
│   │   ├── stats.go
│   │   ├── limit.go
│   │   ├── llm_proxy.go
│   │   └── system.go
│   ├── repository/              # 数据访问
│   │   ├── user.go
│   │   ├── department.go
│   │   ├── apikey.go
│   │   ├── usage.go
│   │   ├── ratelimit.go
│   │   ├── audit.go
│   │   ├── announcement.go
│   │   └── system.go
│   ├── model/                   # 数据模型
│   │   ├── user.go
│   │   ├── department.go
│   │   ├── apikey.go
│   │   ├── usage.go
│   │   ├── ratelimit.go
│   │   ├── audit.go
│   │   ├── announcement.go
│   │   ├── system.go
│   │   └── dto/                  # 数据传输对象
│   │       ├── request.go
│   │       └── response.go
│   ├── router/                  # 路由定义
│   │   └── router.go
│   └── pkg/                     # 内部工具包
│       ├── jwt/
│       ├── crypto/
│       ├── response/
│       ├── validator/
│       └── errcode/
├── pkg/                         # 外部共享包
│   ├── llm/                     # LLM 客户端
│   └── token/                   # Token 计数器
├── migrations/                  # 数据库迁移
├── tests/                        # 测试
│   ├── integration/
│   └── mocks/
└── go.mod
```

---

## 代码规范

### 命名规范

```go
// 包名：小写单词，不使用下划线
package service

// 接口：以行为命名，通常以 -er 结尾
type UserRepository interface {
    Create(ctx context.Context, user *model.User) error
    FindByID(ctx context.Context, id int64) (*model.User, error)
}

// 结构体：驼峰命名
type UserService struct {
    repo repository.UserRepository
    log  *zap.Logger
}

// 常量：驼峰命名或大写下划线
const (
    DefaultPageSize = 20
    MaxPageSize     = 100
)

// 变量：驼峰命名
var currentUser *model.User

// 函数：驼峰命名，导出函数首字母大写
func CreateUser(ctx context.Context, req *dto.CreateUserRequest) (*model.User, error)

// 私有函数：小写开头
func validateUser(req *dto.CreateUserRequest) error {
    // ...
}

// 接口实现：通常以接口名去掉 'I' 前缀命名
type userRepository struct {
    db *gorm.DB
}

func NewUserRepository(db *gorm.DB) repository.UserRepository {
    return &userRepository{db: db}
}
```

### 注释规范

```go
// Package service 提供业务逻辑层的实现
//
// 核心职责：
//   - 实现业务规则和验证
//   - 协调多个 Repository 完成复杂业务
//   - 处理事务边界
package service

// UserService 用户服务层
//
// 负责用户相关的业务逻辑处理，包括用户创建、更新、删除、
// 状态管理等。所有操作都经过权限验证和数据校验。
type UserService struct {
    repo repository.UserRepository
    log  *zap.Logger
}

// NewUserService 创建新的用户服务实例
//
// 参数:
//   repo - 用户数据访问层
//   log  - 日志记录器
//
// 返回:
//   用户服务实例
func NewUserService(repo repository.UserRepository, log *zap.Logger) *UserService {
    return &UserService{
        repo: repo,
        log:  log,
    }
}

// CreateUser 创建新用户
//
// 此方法执行以下操作：
//   1. 验证用户名唯一性
//   2. 加密用户密码
//   3. 保存用户到数据库
//   4. 记录审计日志
//
// 参数:
//   ctx - 请求上下文
//   req - 创建用户请求
//
// 返回:
//   创建的用户信息
//   错误信息（如用户名重复、部门不存在等）
func (s *UserService) CreateUser(ctx context.Context, req *dto.CreateUserRequest) (*model.User, error) {
    // ...
}
```

### 错误处理规范

```go
// 定义错误码
package errcode

const (
    // Success 成功
    Success = 0

    // ErrCodeUserNotFound 用户不存在
    ErrCodeUserNotFound = 40301

    // ErrCodeUserExists 用户名已存在
    ErrCodeUserExists = 40302
)

// 定义业务错误
var (
    ErrUserNotFound = errors.New("user not found")
    ErrUserExists   = errors.New("username already exists")
)

// Service 层返回错误
func (s *UserService) GetUser(ctx context.Context, id int64) (*model.User, error) {
    user, err := s.repo.FindByID(ctx, id)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, fmt.Errorf("%w: %d", ErrUserNotFound, id)
        }
        return nil, fmt.Errorf("failed to find user: %w", err)
    }
    return user, nil
}

// Handler 层转换错误
func (h *UserHandler) GetUser(c *gin.Context) {
    id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

    user, err := h.service.GetUser(c.Request.Context(), id)
    if err != nil {
        if errors.Is(err, service.ErrUserNotFound) {
            response.Error(c, errcode.ErrCodeUserNotFound, "用户不存在")
            return
        }
        response.Error(c, errcode.ErrCodeInternal, "获取用户信息失败")
        return
    }

    response.Success(c, user)
}
```

---

## 分层架构

### Handler 层规范

```go
// handler/user.go
package handler

import (
    "github.com/gin-gonic/gin"
    "codemind/internal/model/dto"
    "codemind/internal/service"
    "codemind/internal/pkg/response"
)

type UserHandler struct {
    service *service.UserService
    log     *zap.Logger
}

func NewUserHandler(service *service.UserService, log *zap.Logger) *UserHandler {
    return &UserHandler{
        service: service,
        log:     log,
    }
}

// RegisterRoutes 注册路由
func (h *UserHandler) RegisterRoutes(r *gin.RouterGroup) {
    users := r.Group("/users")
    {
        users.GET("", h.ListUsers)           // 获取用户列表
        users.POST("", h.CreateUser)         // 创建用户
        users.GET("/:id", h.GetUser)         // 获取用户详情
        users.PUT("/:id", h.UpdateUser)      // 更新用户
        users.DELETE("/:id", h.DeleteUser)   // 删除用户
    }
}

// ListUsers 获取用户列表
// @Summary 获取用户列表
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Success 200 {object} response.Response{data=response.PaginatedResponse}
// @Router /api/v1/users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
    // 解析查询参数
    var req dto.ListUsersRequest
    if err := c.ShouldBindQuery(&req); err != nil {
        response.Error(c, errcode.ErrCodeInvalidParam, err.Error())
        return
    }

    // 验证参数
    if req.Page < 1 {
        req.Page = 1
    }
    if req.PageSize < 1 || req.PageSize > 100 {
        req.PageSize = 20
    }

    // 调用 Service
    result, err := h.service.ListUsers(c.Request.Context(), &req)
    if err != nil {
        response.Error(c, errcode.ErrCodeInternal, "获取用户列表失败")
        return
    }

    response.SuccessWithPagination(c, result.List, result.Pagination)
}
```

### Service 层规范

```go
// service/user.go
package service

import (
    "context"
    "fmt"

    "go.uber.org/zap"
    "gorm.io/gorm"

    "codemind/internal/model"
    "codemind/internal/model/dto"
    "codemind/internal/repository"
)

type UserService struct {
    repo   repository.UserRepository
    deptRepo repository.DepartmentRepository
    log    *zap.Logger
}

func NewUserService(
    userRepo repository.UserRepository,
    deptRepo repository.DepartmentRepository,
    log *zap.Logger,
) *UserService {
    return &UserService{
        repo:     userRepo,
        deptRepo: deptRepo,
        log:      log,
    }
}

// CreateUser 创建用户
func (s *UserService) CreateUser(ctx context.Context, req *dto.CreateUserRequest) (*model.User, error) {
    // 1. 业务验证
    if req.Username == "" {
        return nil, fmt.Errorf("username is required")
    }

    // 2. 检查用户名唯一性
    existing, _ := s.repo.FindByUsername(ctx, req.Username)
    if existing != nil {
        s.log.Warn("username already exists", zap.String("username", req.Username))
        return nil, fmt.Errorf("%w: username %s", ErrUserExists, req.Username)
    }

    // 3. 验证部门是否存在
    if req.DepartmentID > 0 {
        dept, err := s.deptRepo.FindByID(ctx, req.DepartmentID)
        if err != nil {
            return nil, fmt.Errorf("department not found: %d", req.DepartmentID)
        }
        if dept == nil || dept.Status != 1 {
            return nil, fmt.Errorf("department is not available: %d", req.DepartmentID)
        }
    }

    // 4. 加密密码
    passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        return nil, fmt.Errorf("failed to hash password: %w", err)
    }

    // 5. 创建用户对象
    user := &model.User{
        Username:     req.Username,
        PasswordHash: string(passwordHash),
        DisplayName:  req.DisplayName,
        Email:        req.Email,
        Phone:        req.Phone,
        Role:         req.Role,
        DepartmentID: req.DepartmentID,
        Status:       1,
    }

    // 6. 保存到数据库
    if err := s.repo.Create(ctx, user); err != nil {
        s.log.Error("failed to create user",
            zap.String("username", req.Username),
            zap.Error(err))
        return nil, fmt.Errorf("failed to create user: %w", err)
    }

    s.log.Info("user created successfully",
        zap.Int64("user_id", user.ID),
        zap.String("username", user.Username))

    return user, nil
}
```

### Repository 层规范

```go
// repository/user.go
package repository

import (
    "context"
    "errors"

    "gorm.io/gorm"

    "codemind/internal/model"
)

type UserRepository interface {
    Create(ctx context.Context, user *model.User) error
    FindByID(ctx context.Context, id int64) (*model.User, error)
    FindByUsername(ctx context.Context, username string) (*model.User, error)
    Update(ctx context.Context, user *model.User) error
    Delete(ctx context.Context, id int64) error
    List(ctx context.Context, opts *ListOptions) ([]*model.User, int64, error)
}

type userRepository struct {
    db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
    return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
    return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) FindByID(ctx context.Context, id int64) (*model.User, error) {
    var user model.User
    err := r.db.WithContext(ctx).
        Where("id = ? AND deleted_at IS NULL", id).
        First(&user).Error

    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, nil
        }
        return nil, err
    }
    return &user, nil
}

func (r *userRepository) List(ctx context.Context, opts *ListOptions) ([]*model.User, int64, error) {
    var users []*model.User
    var total int64

    query := r.db.WithContext(ctx).Model(&model.User{})

    // 应用过滤条件
    if opts != nil {
        if opts.Keyword != "" {
            keyword := "%" + opts.Keyword + "%"
            query = query.Where("username LIKE ? OR display_name LIKE ?", keyword, keyword)
        }
        if opts.DepartmentID > 0 {
            query = query.Where("department_id = ?", opts.DepartmentID)
        }
        if opts.Role != "" {
            query = query.Where("role = ?", opts.Role)
        }
        if opts.Status > 0 {
            query = query.Where("status = ?", opts.Status)
        }
    }

    // 计算总数
    if err := query.Count(&total).Error; err != nil {
        return nil, 0, err
    }

    // 分页查询
    if opts != nil && opts.PageSize > 0 {
        offset := (opts.Page - 1) * opts.PageSize
        query = query.Offset(offset).Limit(opts.PageSize)
    }

    // 预加载关联数据
    err := query.
        Preload("Department").
        Order("created_at DESC").
        Find(&users).Error

    return users, total, err
}
```

---

## 数据库操作

### Model 定义规范

```go
// model/user.go
package model

import (
    "time"

    "gorm.io/gorm"
)

// User 用户模型
type User struct {
    // 主键
    ID int64 `gorm:"primaryKey;autoIncrement;comment:用户ID"`

    // 基本信息
    Username     string    `gorm:"size:50;not null;uniqueIndex:idx_users_username;comment:登录用户名"`
    PasswordHash string    `gorm:"size:255;not null;comment:密码哈希"`
    DisplayName  string    `gorm:"size:100;not null;comment:显示名称"`
    Email        string    `gorm:"size:255;uniqueIndex;comment:邮箱"`
    Phone        string    `gorm:"size:20;comment:手机号"`
    AvatarURL    string    `gorm:"size:500;comment:头像URL"`

    // 组织信息
    Role         string    `gorm:"size:20;not null;default:'user';index:idx_users_role;comment:角色"`
    DepartmentID *int64    `gorm:"index:idx_users_department_id;comment:所属部门"`
    Department   *Department `gorm:"foreignKey:DepartmentID"`

    // 状态
    Status        int8      `gorm:"not null;default:1;index:idx_users_status;comment:状态1启用0禁用"`
    LastLoginAt  *time.Time `gorm:"comment:最后登录时间"`
    LastLoginIP  string    `gorm:"size:45;comment:最后登录IP"`

    // 审计字段
    CreatedAt time.Time `gorm:"not null;comment:创建时间"`
    UpdatedAt time.Time `gorm:"not null;comment:更新时间"`
    DeletedAt gorm.DeletedAt `gorm:"index;comment:软删除时间"`
}

// TableName 指定表名
func (User) TableName() string {
    return "users"
}
```

### 事务处理规范

```go
// Service 层事务处理
func (s *UserService) TransferUser(ctx context.Context, userID int64, newDeptID int64) error {
    // 使用 Repository 的事务方法
    err := s.repo.Transaction(ctx, func(txRepo repository.Repository) error {
        // 1. 获取用户
        user, err := txRepo.UserRepo().FindByID(ctx, userID)
        if err != nil {
            return err
        }

        // 2. 验证新部门
        dept, err := txRepo.DepartmentRepo().FindByID(ctx, newDeptID)
        if err != nil || dept == nil {
            return fmt.Errorf("department not found")
        }

        // 3. 更新用户部门
        user.DepartmentID = &newDeptID
        if err := txRepo.UserRepo().Update(ctx, user); err != nil {
            return err
        }

        // 4. 记录审计日志
        audit := &model.AuditLog{
            OperatorID:  getOperatorID(ctx),
            Action:     "transfer_user",
            TargetType: "user",
            TargetID:   userID,
            Detail:     fmt.Sprintf("transfer to department %d", newDeptID),
        }
        if err := txRepo.AuditRepo().Create(ctx, audit); err != nil {
            return err
        }

        return nil
    })

    return err
}
```

---

## 中间件开发

### 认证中间件

```go
// middleware/auth.go
package middleware

import (
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"

    "codemind/internal/pkg/jwt"
    "codemind/internal/pkg/response"
)

type AuthMiddleware struct {
    jwtManager *jwt.Manager
}

func NewAuthMiddleware(jwtManager *jwt.Manager) *AuthMiddleware {
    return &AuthMiddleware{jwtManager: jwtManager}
}

// RequireAuth JWT 认证中间件
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 获取 token
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            response.Error(c, 40001, "missing authorization header")
            c.Abort()
            return
        }

        // 解析 Bearer token
        parts := strings.SplitN(authHeader, " ", 2)
        if len(parts) != 2 || parts[0] != "Bearer" {
            response.Error(c, 40001, "invalid authorization format")
            c.Abort()
            return
        }

        token := parts[1]

        // 验证 token
        claims, err := m.jwtManager.Verify(token)
        if err != nil {
            response.Error(c, 40003, "invalid or expired token")
            c.Abort()
            return
        }

        // 将用户信息存入上下文
        c.Set("user_id", claims.UserID)
        c.Set("username", claims.Username)
        c.Set("role", claims.Role)

        c.Next()
    }
}

// RequireRole 角色验证中间件
func (m *AuthMiddleware) RequireRole(roles ...string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userRole, exists := c.Get("role")
        if !exists {
            response.Error(c, 40101, "unauthorized")
            c.Abort()
            return
        }

        roleStr := userRole.(string)
        for _, role := range roles {
            if roleStr == role {
                c.Next()
                return
            }
        }

        response.Error(c, 40101, "insufficient permissions")
        c.Abort()
    }
}
```

### API Key 认证中间件

```go
// middleware/apikey.go
package middleware

import (
    "fmt"

    "github.com/gin-gonic/gin"

    "codemind/internal/pkg/response"
)

type APIKeyMiddleware struct {
    cache *Cache
}

func NewAPIKeyMiddleware(cache *Cache) *APIKeyMiddleware {
    return &APIKeyMiddleware{cache: cache}
}

// RequireAPIKey API Key 认证中间件
func (m *APIKeyMiddleware) RequireAPIKey() gin.HandlerFunc {
    return func(c *gin.Context) {
        apiKey := c.GetHeader("Authorization")
        if apiKey == "" {
            response.Error(c, 40005, "missing API key")
            c.Abort()
            return
        }

        // 移除 Bearer 前缀
        if len(apiKey) > 7 && apiKey[:7] == "Bearer " {
            apiKey = apiKey[7:]
        }

        // 验证 API Key
        userInfo, err := m.cache.GetAPIKey(c.Request.Context(), apiKey)
        if err != nil || userInfo == nil {
            response.Error(c, 40005, "invalid API key")
            c.Abort()
            return
        }

        // 检查用户状态
        if userInfo.Status != 1 {
            response.Error(c, 40004, "user account is disabled")
            c.Abort()
            return
        }

        // 将用户信息存入上下文
        c.Set("user_id", userInfo.ID)
        c.Set("username", userInfo.Username)
        c.Set("role", userInfo.Role)
        c.Set("api_key_id", userInfo.KeyID)

        c.Next()
    }
}
```

---

## 配置管理

```go
// config/config.go
package config

import (
    "time"

    "github.com/spf13/viper"
)

type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    Redis    RedisConfig
    JWT      JWTConfig
    LLM      LLMConfig
    System   SystemConfig
    Log      LogConfig
}

type ServerConfig struct {
    Host string
    Port int
    Mode string // debug | release
}

type DatabaseConfig struct {
    Host            string
    Port            int
    Name            string
    User            string
    Password        string
    MaxOpenConns    int
    MaxIdleConns    int
    ConnMaxLifetime int // minutes
}

type JWTConfig struct {
    Secret     string
    ExpireHours int
}

// Load 加载配置
func Load(path string) (*Config, error) {
    viper.SetConfigFile(path)
    viper.SetConfigType("yaml")

    // 环境变量覆盖
    viper.AutomaticEnv()
    viper.SetEnvPrefix("CODEMIND")

    if err := viper.ReadInConfig(); err != nil {
        return nil, fmt.Errorf("failed to read config: %w", err)
    }

    var cfg Config
    if err := viper.Unmarshal(&cfg); err != nil {
        return nil, fmt.Errorf("failed to unmarshal config: %w", err)
    }

    // 验证配置
    if err := cfg.Validate(); err != nil {
        return nil, err
    }

    return &cfg, nil
}

func (c *Config) Validate() error {
    if c.Database.Password == "" {
        return fmt.Errorf("database password is required")
    }
    if c.JWT.Secret == "" {
        return fmt.Errorf("JWT secret is required")
    }
    if len(c.JWT.Secret) < 32 {
        return fmt.Errorf("JWT secret must be at least 32 characters")
    }
    return nil
}
```

---

## 日志规范

```go
// 使用结构化日志
import "go.uber.org/zap"

func (s *UserService) CreateUser(ctx context.Context, req *dto.CreateUserRequest) (*model.User, error) {
    s.log.Info("Creating user",
        zap.String("username", req.Username),
        zap.String("display_name", req.DisplayName),
        zap.String("operator", getOperator(ctx)),
    )

    user, err := s.repo.Create(ctx, req)
    if err != nil {
        s.log.Error("Failed to create user",
            zap.String("username", req.Username),
            zap.Error(err),
        )
        return nil, err
    }

    s.log.Info("User created successfully",
        zap.Int64("user_id", user.ID),
        zap.String("username", user.Username),
    )

    return user, nil
}

// 日志级别使用规范
// Debug: 详细的调试信息，开发环境使用
// Info: 重要的业务流程节点
// Warn: 警告信息，如重试操作、降级处理
// Error: 错误信息，但程序可以继续运行
// Fatal: 致命错误，程序无法继续运行
```

---

## 部署规范

### Dockerfile

```dockerfile
# 多阶段构建
FROM golang:1.23-alpine AS builder

WORKDIR /build

# 安装依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制源码
COPY . .

# 构建
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o codemind cmd/server/main.go

# 运行镜像
FROM alpine:3.20

WORKDIR /app

# 安装 ca-certificates (用于 HTTPS 请求)
RUN apk --no-cache add ca-certificates tzdata

# 从 builder 复制编译好的二进制文件
COPY --from=builder /build/codemind .

# 复制配置文件（可选，通常通过 volume 挂载）
COPY --from=builder /build/deploy/config ./config

# 设置时区
ENV TZ=Asia/Shanghai

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# 运行
ENTRYPOINT ["./codemind"]
```
