package router

import (
	"codemind/internal/handler"
	"codemind/internal/middleware"
	"codemind/internal/model"
	jwtPkg "codemind/internal/pkg/jwt"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Handlers 所有 Handler 集合
type Handlers struct {
	Auth       *handler.AuthHandler
	User       *handler.UserHandler
	Department *handler.DepartmentHandler
	APIKey     *handler.APIKeyHandler
	LLMProxy   *handler.LLMProxyHandler
	Stats      *handler.StatsHandler
	Limit      *handler.LimitHandler
	System     *handler.SystemHandler
	MCPAdmin   *handler.MCPAdminHandler
	MCPGateway *handler.MCPGatewayHandler
}

// Setup 初始化路由
func Setup(
	engine *gin.Engine,
	handlers *Handlers,
	jwtManager *jwtPkg.Manager,
	db *gorm.DB,
	rdb *redis.Client,
	logger *zap.Logger,
) {
	// ──────────────────────────────────
	// 全局中间件
	// ──────────────────────────────────
	engine.Use(middleware.Recovery(logger))
	engine.Use(middleware.CORS())
	engine.Use(middleware.Logger(logger))

	// ──────────────────────────────────
	// 健康检查（无需认证）
	// ──────────────────────────────────
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// ──────────────────────────────────
	// 管理平台 API (/api/v1)
	// ──────────────────────────────────
	apiV1 := engine.Group("/api/v1")

	// 认证接口（无需 JWT）
	auth := apiV1.Group("/auth")
	{
		auth.POST("/login", handlers.Auth.Login)
	}

	// 需要 JWT 认证的接口
	authenticated := apiV1.Group("")
	authenticated.Use(middleware.JWTAuth(jwtManager))
	{
		// 认证相关
		authProtected := authenticated.Group("/auth")
		{
			authProtected.POST("/logout", handlers.Auth.Logout)
			authProtected.GET("/profile", handlers.Auth.GetProfile)
			authProtected.PUT("/profile", handlers.Auth.UpdateProfile)
			authProtected.PUT("/password", handlers.Auth.ChangePassword)
		}

		// API Key 管理（所有已登录用户可用）
		keys := authenticated.Group("/keys")
		{
			keys.GET("", handlers.APIKey.List)
			keys.POST("", handlers.APIKey.Create)
			keys.PUT("/:id/status", handlers.APIKey.UpdateStatus)
			keys.DELETE("/:id", handlers.APIKey.Delete)
		}

		// 用量统计（所有已登录用户可用，权限在 Service 层控制）
		stats := authenticated.Group("/stats")
		{
			stats.GET("/overview", handlers.Stats.Overview)
			stats.GET("/usage", handlers.Stats.Usage)
			stats.GET("/ranking", handlers.Stats.Ranking)
		}

		// 限额查询（所有用户可查看自己的限额）
		limits := authenticated.Group("/limits")
		{
			limits.GET("/my", handlers.Limit.GetMyLimits)
		}

		// 公告查询（所有用户可查看已发布公告）
		announcements := authenticated.Group("/announcements")
		{
			announcements.GET("", handlers.System.ListAnnouncements)
		}

		// 用户管理（管理员 + 部门经理）
		users := authenticated.Group("/users")
		users.Use(middleware.RequireRole(model.RoleSuperAdmin, model.RoleDeptManager))
		{
			users.GET("", handlers.User.List)
			users.POST("", handlers.User.Create)
			users.GET("/:id", handlers.User.GetDetail)
			users.PUT("/:id", handlers.User.Update)
			users.PUT("/:id/status", handlers.User.UpdateStatus)
			users.PUT("/:id/reset-password", handlers.User.ResetPassword)
			users.DELETE("/:id", handlers.User.Delete)
		}

		// 部门管理（管理员 + 部门经理）
		departments := authenticated.Group("/departments")
		departments.Use(middleware.RequireRole(model.RoleSuperAdmin, model.RoleDeptManager))
		{
			departments.GET("", handlers.Department.List)
			departments.GET("/:id", handlers.Department.GetDetail)
			departments.POST("", handlers.Department.Create)
			departments.PUT("/:id", handlers.Department.Update)
			departments.DELETE("/:id", handlers.Department.Delete)
		}

		// 限额管理（管理员 + 部门经理）
		limitsAdmin := authenticated.Group("/limits")
		limitsAdmin.Use(middleware.RequireRole(model.RoleSuperAdmin, model.RoleDeptManager))
		{
			limitsAdmin.GET("", handlers.Limit.List)
			limitsAdmin.PUT("", handlers.Limit.Upsert)
			limitsAdmin.DELETE("/:id", handlers.Limit.Delete)
		}

		// 系统管理（仅超级管理员）
		system := authenticated.Group("/system")
		system.Use(middleware.RequireAdmin())
		{
			// 系统配置
			system.GET("/configs", handlers.System.GetConfigs)
			system.PUT("/configs", handlers.System.UpdateConfigs)

			// 公告管理（增删改）
			system.POST("/announcements", handlers.System.CreateAnnouncement)
			system.PUT("/announcements/:id", handlers.System.UpdateAnnouncement)
			system.DELETE("/announcements/:id", handlers.System.DeleteAnnouncement)

			// 审计日志
			system.GET("/audit-logs", handlers.System.ListAuditLogs)
		}

		// MCP 服务管理（仅超级管理员）
		mcpAdmin := authenticated.Group("/mcp")
		mcpAdmin.Use(middleware.RequireAdmin())
		{
			mcpAdmin.GET("/services", handlers.MCPAdmin.ListServices)
			mcpAdmin.POST("/services", handlers.MCPAdmin.CreateService)
			mcpAdmin.PUT("/services/:id", handlers.MCPAdmin.UpdateService)
			mcpAdmin.DELETE("/services/:id", handlers.MCPAdmin.DeleteService)
			mcpAdmin.POST("/services/:id/sync", handlers.MCPAdmin.SyncTools)
			mcpAdmin.GET("/services/:id/tools", handlers.MCPAdmin.GetServiceTools)
			mcpAdmin.GET("/access-rules", handlers.MCPAdmin.ListAccessRules)
			mcpAdmin.POST("/access-rules", handlers.MCPAdmin.SetAccessRule)
			mcpAdmin.DELETE("/access-rules/:id", handlers.MCPAdmin.DeleteAccessRule)
		}
	}

	// ──────────────────────────────────
	// MCP 网关协议接口 (/mcp)
	// 使用 API Key 认证
	// ──────────────────────────────────
	mcpGateway := engine.Group("/mcp")
	mcpGateway.Use(middleware.APIKeyAuth(db, rdb))
	{
		mcpGateway.GET("/sse", handlers.MCPGateway.SSEConnect)
		mcpGateway.POST("/message", handlers.MCPGateway.HandleMessage)
		mcpGateway.POST("/", handlers.MCPGateway.HandleStreamableHTTP)
	}

	// ──────────────────────────────────
	// LLM 代理接口 (/v1)
	// 使用 API Key 认证
	// ──────────────────────────────────
	llmV1 := engine.Group("/v1")
	llmV1.Use(middleware.APIKeyAuth(db, rdb))
	{
		// OpenAI 兼容接口
		llmV1.POST("/chat/completions", handlers.LLMProxy.ChatCompletions)
		llmV1.POST("/completions", handlers.LLMProxy.Completions)
		llmV1.GET("/models", handlers.LLMProxy.ListModels)

		// Anthropic 原生接口
		llmV1.POST("/messages", handlers.LLMProxy.AnthropicMessages)
	}
}
