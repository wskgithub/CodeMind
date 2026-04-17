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

// Handlers aggregates all HTTP handlers.
type Handlers struct {
	Auth               *handler.AuthHandler
	User               *handler.UserHandler
	Department         *handler.DepartmentHandler
	APIKey             *handler.APIKeyHandler
	LLMProxy           *handler.LLMProxyHandler
	Stats              *handler.StatsHandler
	Limit              *handler.LimitHandler
	System             *handler.SystemHandler
	MCPAdmin           *handler.MCPAdminHandler
	MCPGateway         *handler.MCPGatewayHandler
	LLMBackend         *handler.LLMBackendHandler
	Monitor            *handler.MonitorHandler
	Document           *handler.DocumentHandler
	Upload             *handler.UploadHandler
	TrainingData       *handler.TrainingDataHandler
	ThirdPartyProvider *handler.ThirdPartyProviderHandler
}

// Setup initializes all routes.
func Setup(
	engine *gin.Engine,
	handlers *Handlers,
	jwtManager *jwtPkg.Manager,
	db *gorm.DB,
	rdb *redis.Client,
	logger *zap.Logger,
	corsOrigins []string,
	uploadDir string,
) {
	// Global middleware
	engine.Use(middleware.Recovery(logger))
	engine.Use(middleware.CORS(corsOrigins))
	engine.Use(middleware.Logger(logger))

	if handlers.Monitor != nil {
		engine.Use(middleware.RequestMonitor(handlers.Monitor))
	}

	// Health check (no auth)
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"}) //nolint:mnd // intentional constant.
	})

	// 静态文件服务（上传的图片等资源）
	engine.Static("/uploads", uploadDir)

	// Management API (/api/v1)
	apiV1 := engine.Group("/api/v1")

	// Auth endpoints (no JWT)
	auth := apiV1.Group("/auth")
	{
		auth.POST("/login", handlers.Auth.Login)
	}

	// JWT-protected endpoints
	authenticated := apiV1.Group("")
	authenticated.Use(middleware.JWTAuth(jwtManager))
	{
		authProtected := authenticated.Group("/auth")
		{
			authProtected.POST("/logout", handlers.Auth.Logout)
			authProtected.GET("/profile", handlers.Auth.GetProfile)
			authProtected.PUT("/profile", handlers.Auth.UpdateProfile)
			authProtected.PUT("/password", handlers.Auth.ChangePassword)
		}

		// API Key management (all authenticated users)
		keys := authenticated.Group("/keys")
		{
			keys.GET("", handlers.APIKey.List)
			keys.POST("", handlers.APIKey.Create)
			keys.POST("/:id/copy", handlers.APIKey.Copy)
			keys.PUT("/:id/status", handlers.APIKey.UpdateStatus)
			keys.DELETE("/:id", handlers.APIKey.Delete)
		}

		// Usage stats (all authenticated users, permission checked in service)
		stats := authenticated.Group("/stats")
		{
			stats.GET("/overview", handlers.Stats.Overview)
			stats.GET("/usage", handlers.Stats.Usage)
			stats.GET("/ranking", handlers.Stats.Ranking)
			stats.GET("/key-usage", handlers.Stats.KeyUsageSummary)
			stats.GET("/export/csv", handlers.Stats.ExportCSV).Use(middleware.RequireRole(model.RoleSuperAdmin, model.RoleDeptManager))
		}

		// Quota queries (users can view their own limits)
		limits := authenticated.Group("/limits")
		{
			limits.GET("/my", handlers.Limit.GetMyLimits)
			limits.GET("/my/progress", handlers.Limit.GetMyProgress)
		}

		// Platform settings (all users)
		settings := authenticated.Group("/settings")
		{
			settings.GET("/platform", handlers.System.GetPlatformServiceURL)
		}

		// Announcements (all users can view published)
		announcements := authenticated.Group("/announcements")
		{
			announcements.GET("", handlers.System.ListAnnouncements)
		}

		// Documents (all users can view published)
		docs := authenticated.Group("/docs")
		{
			docs.GET("", handlers.Document.ListDocuments)
			docs.GET("/:slug", handlers.Document.GetDocument)
		}

		// Model services (all authenticated users)
		models := authenticated.Group("/models")
		{
			models.GET("/platform", handlers.ThirdPartyProvider.ListPlatformModels)
			models.GET("/templates", handlers.ThirdPartyProvider.ListTemplatesForUser)
			models.GET("/third-party", handlers.ThirdPartyProvider.ListProviders)
			models.POST("/third-party", handlers.ThirdPartyProvider.CreateProvider)
			models.PUT("/third-party/:id", handlers.ThirdPartyProvider.UpdateProvider)
			models.PUT("/third-party/:id/status", handlers.ThirdPartyProvider.UpdateProviderStatus)
			models.DELETE("/third-party/:id", handlers.ThirdPartyProvider.DeleteProvider)
		}

		// User management (admin + dept manager)
		users := authenticated.Group("/users")
		users.Use(middleware.RequireRole(model.RoleSuperAdmin, model.RoleDeptManager))
		{
			users.GET("", handlers.User.List)
			users.POST("", handlers.User.Create)
			users.GET("/:id", handlers.User.GetDetail)
			users.PUT("/:id", handlers.User.Update)
			users.PUT("/:id/status", handlers.User.UpdateStatus)
			users.PUT("/:id/reset-password", handlers.User.ResetPassword)
			users.PUT("/:id/unlock", handlers.User.UnlockUser)
			users.DELETE("/:id", handlers.User.Delete)
		}

		// Department management (admin + dept manager)
		departments := authenticated.Group("/departments")
		departments.Use(middleware.RequireRole(model.RoleSuperAdmin, model.RoleDeptManager))
		{
			departments.GET("", handlers.Department.List)
			departments.GET("/:id", handlers.Department.GetDetail)
			departments.POST("", handlers.Department.Create)
			departments.PUT("/:id", handlers.Department.Update)
			departments.DELETE("/:id", handlers.Department.Delete)
		}

		// Quota management (admin + dept manager)
		limitsAdmin := authenticated.Group("/limits")
		limitsAdmin.Use(middleware.RequireRole(model.RoleSuperAdmin, model.RoleDeptManager))
		{
			limitsAdmin.GET("", handlers.Limit.List)
			limitsAdmin.PUT("", handlers.Limit.Upsert)
			limitsAdmin.DELETE("/:id", handlers.Limit.Delete)
		}

		// System management (super admin only)
		system := authenticated.Group("/system")
		system.Use(middleware.RequireAdmin())
		{
			system.GET("/configs", handlers.System.GetConfigs)
			system.PUT("/configs", handlers.System.UpdateConfigs)
			system.POST("/announcements", handlers.System.CreateAnnouncement)
			system.PUT("/announcements/:id", handlers.System.UpdateAnnouncement)
			system.DELETE("/announcements/:id", handlers.System.DeleteAnnouncement)
			system.GET("/audit-logs", handlers.System.ListAuditLogs)
			system.GET("/llm-backends", handlers.LLMBackend.List)
			system.POST("/llm-backends", handlers.LLMBackend.Create)
			system.PUT("/llm-backends/:id", handlers.LLMBackend.Update)
			system.DELETE("/llm-backends/:id", handlers.LLMBackend.Delete)
		}

		// Third-party provider templates (super admin only)
		providerTemplates := authenticated.Group("/system/provider-templates")
		providerTemplates.Use(middleware.RequireAdmin())
		{
			providerTemplates.GET("", handlers.ThirdPartyProvider.ListTemplatesAdmin)
			providerTemplates.POST("", handlers.ThirdPartyProvider.CreateTemplate)
			providerTemplates.PUT("/:id", handlers.ThirdPartyProvider.UpdateTemplate)
			providerTemplates.DELETE("/:id", handlers.ThirdPartyProvider.DeleteTemplate)
		}

		// MCP service management (super admin only)
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

		// System monitoring (super admin only)
		monitorGroup := authenticated.Group("/monitor")
		monitorGroup.Use(middleware.RequireAdmin())
		{
			monitorGroup.GET("/dashboard", handlers.Monitor.DashboardSummary)
			monitorGroup.GET("/system", handlers.Monitor.SystemMetrics)
			monitorGroup.GET("/requests", handlers.Monitor.RequestMetrics)
			monitorGroup.GET("/llm-nodes", handlers.Monitor.LLMNodeMetrics)
			monitorGroup.GET("/health", handlers.Monitor.HealthCheck)
		}

		// Document management (super admin only)
		docsAdmin := authenticated.Group("/docs/admin")
		docsAdmin.Use(middleware.RequireAdmin())
		{
			docsAdmin.GET("", handlers.Document.ListAllDocuments)
			docsAdmin.GET("/:id", handlers.Document.GetDocumentByID)
			docsAdmin.POST("", handlers.Document.CreateDocument)
			docsAdmin.PUT("/:id", handlers.Document.UpdateDocument)
			docsAdmin.DELETE("/:id", handlers.Document.DeleteDocument)
			docsAdmin.POST("/upload/image", handlers.Upload.UploadImage)
		}

		// Training data management (super admin only)
		trainingData := authenticated.Group("/training-data")
		trainingData.Use(middleware.RequireAdmin())
		{
			trainingData.GET("", handlers.TrainingData.List)
			trainingData.GET("/stats", handlers.TrainingData.GetStats)
			trainingData.GET("/:id", handlers.TrainingData.GetDetail)
			trainingData.PUT("/:id/exclude", handlers.TrainingData.UpdateExcluded)
			trainingData.POST("/export", handlers.TrainingData.Export)
		}
	}

	// MCP Gateway (API Key auth)
	mcpGateway := engine.Group("/mcp")
	mcpGateway.Use(middleware.APIKeyAuth(db, rdb, logger))
	{
		mcpGateway.GET("/sse", handlers.MCPGateway.SSEConnect)
		mcpGateway.POST("/message", handlers.MCPGateway.HandleMessage)
		mcpGateway.POST("/", handlers.MCPGateway.HandleStreamableHTTP)
	}

	// LLM Proxy - OpenAI protocol (API Key auth)
	openaiLLM := engine.Group("/api/openai/v1")
	openaiLLM.Use(middleware.SetLLMProtocol("openai"))
	openaiLLM.Use(middleware.APIKeyAuth(db, rdb, logger))
	{
		openaiLLM.POST("/chat/completions", handlers.LLMProxy.ChatCompletions)
		openaiLLM.POST("/completions", handlers.LLMProxy.Completions)
		openaiLLM.GET("/models", handlers.LLMProxy.ListModels)
		openaiLLM.GET("/models/:model", handlers.LLMProxy.RetrieveModel)
		openaiLLM.POST("/embeddings", handlers.LLMProxy.Embeddings)
		openaiLLM.POST("/responses", handlers.LLMProxy.Responses)
	}

	// LLM Proxy - Anthropic protocol (API Key auth)
	// Anthropic SDK appends /v1/messages to base URL automatically
	anthropicLLM := engine.Group("/api/anthropic")
	anthropicLLM.Use(middleware.SetLLMProtocol("anthropic"))
	anthropicLLM.Use(middleware.APIKeyAuth(db, rdb, logger))
	{
		anthropicLLM.POST("/v1/messages", handlers.LLMProxy.AnthropicMessages)
	}
}
