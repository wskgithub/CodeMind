package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"codemind/internal/config"
	"codemind/internal/handler"
	jwtPkg "codemind/internal/pkg/jwt"
	"codemind/internal/repository"
	"codemind/internal/router"
	"codemind/internal/service"
	"codemind/pkg/llm"
	mcpPkg "codemind/pkg/mcp"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

// 构建时注入的版本信息
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

func main() {
	// ──────────────────────────────────
	// 1. 加载配置
	// ──────────────────────────────────
	cfg, err := config.Load("")
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// ──────────────────────────────────
	// 2. 初始化日志
	// ──────────────────────────────────
	logger, err := config.InitLogger(&cfg.Log)
	if err != nil {
		fmt.Printf("初始化日志失败: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("CodeMind 服务启动中",
		zap.String("version", Version),
		zap.String("build_time", BuildTime),
		zap.String("git_commit", GitCommit),
	)

	// ──────────────────────────────────
	// 3. 连接数据库
	// ──────────────────────────────────
	db, err := initDatabase(cfg, logger)
	if err != nil {
		logger.Fatal("数据库连接失败", zap.Error(err))
	}
	logger.Info("数据库连接成功")

	// ──────────────────────────────────
	// 4. 连接 Redis
	// ──────────────────────────────────
	rdb, err := initRedis(cfg, logger)
	if err != nil {
		logger.Fatal("Redis 连接失败", zap.Error(err))
	}
	logger.Info("Redis 连接成功")

	// ──────────────────────────────────
	// 5. 初始化基础设施
	// ──────────────────────────────────
	jwtManager := jwtPkg.NewManager(cfg.JWT.Secret, cfg.JWT.ExpireHours, rdb)

	// LLM Provider 管理器（支持多 Provider 和模型路由）
	providerManager := initProviderManager(cfg, logger)

	// ──────────────────────────────────
	// 6. 初始化 Repository 层
	// ──────────────────────────────────
	userRepo := repository.NewUserRepository(db)
	deptRepo := repository.NewDepartmentRepository(db)
	apiKeyRepo := repository.NewAPIKeyRepository(db)
	auditRepo := repository.NewAuditRepository(db)
	usageRepo := repository.NewUsageRepository(db)
	limitRepo := repository.NewRateLimitRepository(db)
	sysConfigRepo := repository.NewSystemRepository(db)
	annRepo := repository.NewAnnouncementRepository(db)
	mcpRepo := repository.NewMCPRepository(db)

	// ──────────────────────────────────
	// 7. 初始化 Service 层
	// ──────────────────────────────────
	authService := service.NewAuthService(userRepo, auditRepo, jwtManager, logger)
	userService := service.NewUserService(userRepo, deptRepo, auditRepo, logger)
	deptService := service.NewDepartmentService(deptRepo, userRepo, auditRepo, logger)
	apiKeyService := service.NewAPIKeyService(apiKeyRepo, auditRepo, logger)
	llmProxyService := service.NewLLMProxyService(providerManager, usageRepo, limitRepo, apiKeyRepo, rdb, logger)
	statsService := service.NewStatsService(usageRepo, userRepo, deptRepo, apiKeyRepo, logger)
	limitService := service.NewLimitService(limitRepo, usageRepo, auditRepo, rdb, logger)
	systemService := service.NewSystemService(sysConfigRepo, auditRepo, annRepo, logger)
	mcpProxy := mcpPkg.NewProxy(logger)
	mcpService := service.NewMCPService(mcpRepo, mcpProxy, logger)

	// ──────────────────────────────────
	// 8. 初始化 Handler 层
	// ──────────────────────────────────
	handlers := &router.Handlers{
		Auth:       handler.NewAuthHandler(authService),
		User:       handler.NewUserHandler(userService),
		Department: handler.NewDepartmentHandler(deptService),
		APIKey:     handler.NewAPIKeyHandler(apiKeyService),
		LLMProxy:   handler.NewLLMProxyHandler(llmProxyService, logger),
		Stats:      handler.NewStatsHandler(statsService),
		Limit:      handler.NewLimitHandler(limitService),
		System:     handler.NewSystemHandler(systemService),
		MCPAdmin:   handler.NewMCPAdminHandler(mcpService, logger),
		MCPGateway: handler.NewMCPGatewayHandler(mcpService, logger),
	}

	// ──────────────────────────────────
	// 9. 初始化 HTTP 引擎
	// ──────────────────────────────────
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	router.Setup(engine, handlers, jwtManager, db, rdb, logger)

	// ──────────────────────────────────
	// 10. 启动 HTTP 服务（优雅关停）
	// ──────────────────────────────────
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      engine,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 600 * time.Second, // LLM 流式请求需要较长超时
		IdleTimeout:  120 * time.Second,
	}

	// 异步启动服务
	go func() {
		logger.Info("HTTP 服务启动", zap.String("addr", addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("HTTP 服务异常退出", zap.Error(err))
		}
	}()

	// 等待终止信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("收到关停信号，开始优雅关停...")

	// 给予 10 秒超时来处理正在进行的请求
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("服务关停失败", zap.Error(err))
	}

	logger.Info("服务已停止")
}

// initDatabase 初始化数据库连接
func initDatabase(cfg *config.Config, logger *zap.Logger) (*gorm.DB, error) {
	// 配置 GORM 日志级别
	var logLevel gormLogger.LogLevel
	switch cfg.Server.Mode {
	case "debug":
		logLevel = gormLogger.Info
	default:
		logLevel = gormLogger.Warn
	}

	db, err := gorm.Open(postgres.Open(cfg.Database.DSN()), &gorm.Config{
		Logger: gormLogger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("打开数据库连接失败: %w", err)
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取底层 DB 实例失败: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.Database.ConnMaxLifetimeMin) * time.Minute)

	return db, nil
}

// initProviderManager 初始化多 Provider 管理器
func initProviderManager(cfg *config.Config, logger *zap.Logger) *llm.ProviderManager {
	providers := cfg.LLM.GetEffectiveProviders()
	defaultName := cfg.LLM.GetDefaultProviderName()
	manager := llm.NewProviderManager(defaultName)

	for _, pc := range providers {
		// 设置默认超时
		if pc.TimeoutSeconds == 0 {
			pc.TimeoutSeconds = 300
		}
		if pc.StreamTimeoutSeconds == 0 {
			pc.StreamTimeoutSeconds = 600
		}

		switch pc.Format {
		case "anthropic":
			client := llm.NewAnthropicClient(pc.BaseURL, pc.APIKey, pc.TimeoutSeconds, pc.StreamTimeoutSeconds)
			provider := llm.NewAnthropicProvider(pc.Name, client)
			manager.Register(provider)
			logger.Info("注册 Anthropic Provider", zap.String("name", pc.Name), zap.String("base_url", pc.BaseURL))
		default: // "openai" 或未指定
			client := llm.NewClient(pc.BaseURL, pc.APIKey, pc.TimeoutSeconds, pc.StreamTimeoutSeconds)
			provider := llm.NewOpenAIProvider(pc.Name, client)
			manager.Register(provider)
			logger.Info("注册 OpenAI Provider", zap.String("name", pc.Name), zap.String("base_url", pc.BaseURL))
		}
	}

	// 设置模型路由规则
	if cfg.LLM.ModelRouting != nil {
		manager.SetModelRoutes(cfg.LLM.ModelRouting)
		logger.Info("模型路由规则已加载", zap.Int("rules", len(cfg.LLM.ModelRouting)))
	}

	logger.Info(manager.DebugRoutes())
	return manager
}

// initRedis 初始化 Redis 连接
func initRedis(cfg *config.Config, logger *zap.Logger) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		PoolSize: cfg.Redis.PoolSize,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("Redis 连接测试失败: %w", err)
	}

	return rdb, nil
}
