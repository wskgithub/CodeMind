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
	"codemind/internal/model"
	"codemind/internal/model/monitor"
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

	// 自动迁移：确保新增字段和表结构存在，对已有表只做增量变更不删除数据
	if err := db.AutoMigrate(&model.LLMBackend{}, &model.RateLimit{}, &monitor.SystemMetric{}, &monitor.LLMNodeMetric{}, &model.Document{}, &model.LLMTrainingData{}); err != nil {
		logger.Warn("AutoMigrate 失败", zap.Error(err))
	}
	// 修复旧数据：为 period_hours=0 的限额记录补充正确的周期小时数
	fixPeriodHours(db, logger)

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
	backendRepo := repository.NewLLMBackendRepository(db)
	monitorRepo := repository.NewMonitorRepository(db)
	docRepo := repository.NewDocumentRepository(db)
	trainingDataRepo := repository.NewTrainingDataRepository(db)

	// ──────────────────────────────────
	// 7. 初始化负载均衡器
	// ──────────────────────────────────
	loadBalancer := llm.NewLoadBalancer(rdb, logger)

	// ──────────────────────────────────
	// 8. 初始化 Service 层
	// ──────────────────────────────────
	authService := service.NewAuthService(userRepo, auditRepo, jwtManager, logger)
	userService := service.NewUserService(userRepo, deptRepo, auditRepo, logger)
	deptService := service.NewDepartmentService(deptRepo, userRepo, auditRepo, logger)
	apiKeyService := service.NewAPIKeyService(apiKeyRepo, auditRepo, logger)
	limitService := service.NewLimitService(limitRepo, usageRepo, auditRepo, rdb, logger)
	llmProxyService := service.NewLLMProxyService(providerManager, loadBalancer, usageRepo, limitRepo, apiKeyRepo, trainingDataRepo, sysConfigRepo, limitService, rdb, logger)
	statsService := service.NewStatsService(usageRepo, userRepo, deptRepo, apiKeyRepo, logger)
	systemService := service.NewSystemService(sysConfigRepo, auditRepo, annRepo, logger)
	mcpProxy := mcpPkg.NewProxy(logger)
	mcpService := service.NewMCPService(mcpRepo, mcpProxy, logger)
	llmBackendService := service.NewLLMBackendService(backendRepo, auditRepo, loadBalancer, logger)
	monitorService := service.NewMonitorService(monitorRepo, usageRepo, rdb, logger)
	docService := service.NewDocumentService(docRepo, logger)
	trainingDataService := service.NewTrainingDataService(trainingDataRepo, logger)

	// 从数据库加载 LLM 后端节点到负载均衡器
	llmBackendService.RefreshLoadBalancer()

	// ──────────────────────────────────
	// 9. 初始化 Handler 层
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
		LLMBackend: handler.NewLLMBackendHandler(llmBackendService),
		Monitor:      handler.NewMonitorHandler(monitorService, logger),
		Document:     handler.NewDocumentHandler(docService),
		TrainingData: handler.NewTrainingDataHandler(trainingDataService, logger),
	}

	// ──────────────────────────────────
	// 10. 初始化 HTTP 引擎
	// ──────────────────────────────────
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	router.Setup(engine, handlers, jwtManager, db, rdb, logger)

	// ──────────────────────────────────
	// 11. 启动 HTTP 服务（优雅关停）
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

// fixPeriodHours 修复 rate_limits 表的 period_hours 字段及唯一索引
// 步骤：
//  1. 为 period_hours=0 的旧记录从 period 标签填充正确的小时数
//  2. 若旧的 (target_type, target_id, period) 唯一索引存在则删除
//  3. 确保新的 (target_type, target_id, period_hours) 唯一索引存在
func fixPeriodHours(db *gorm.DB, logger *zap.Logger) {
	// 1. 补充旧数据的 period_hours
	periodMap := map[string]int{
		"daily": 24, "weekly": 168, "monthly": 720,
	}
	for period, hours := range periodMap {
		result := db.Exec(
			"UPDATE rate_limits SET period_hours = ? WHERE period = ? AND (period_hours = 0 OR period_hours IS NULL)",
			hours, period,
		)
		if result.RowsAffected > 0 {
			logger.Info("修复 period_hours 旧数据",
				zap.String("period", period),
				zap.Int("hours", hours),
				zap.Int64("rows", result.RowsAffected),
			)
		}
	}

	// 2. 删除以 period 列结尾的旧唯一索引（判断 indexdef 包含 "period)" 而非 "period_hours)"）
	db.Exec(`
		DO $$ BEGIN
			IF EXISTS (
				SELECT 1 FROM pg_indexes
				WHERE tablename  = 'rate_limits'
				  AND indexname  = 'idx_rate_limits_target'
				  AND indexdef   NOT LIKE '%period_hours%'
			) THEN
				DROP INDEX idx_rate_limits_target;
			END IF;
		END $$;
	`)

	// 3. 创建新的唯一索引（幂等，已存在则跳过）
	if err := db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_rate_limits_target
		ON rate_limits(target_type, target_id, period_hours);
	`).Error; err != nil {
		logger.Warn("创建 idx_rate_limits_target 索引失败", zap.Error(err))
	}
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
