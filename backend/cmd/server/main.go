// Package main 是 CodeMind 服务端入口。
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
	"codemind/internal/pkg/crypto"
	"codemind/internal/repository"
	"codemind/internal/router"
	"codemind/internal/service"
	"codemind/pkg/llm"

	jwtPkg "codemind/internal/pkg/jwt"

	mcpPkg "codemind/pkg/mcp"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

// Build-time injected version info.
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

func main() {
	// 1. Load configuration
	cfg, err := config.Load("")
	if err != nil {
		fmt.Printf("failed to load config: %v\n", err)
		os.Exit(1)
	}

	// 2. Initialize logger
	logger, err := config.InitLogger(&cfg.Log)
	if err != nil {
		fmt.Printf("failed to init logger: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = logger.Sync() }()

	logger.Info("CodeMind starting",
		zap.String("version", Version),
		zap.String("build_time", BuildTime),
		zap.String("git_commit", GitCommit),
	)

	// 3. Connect to database
	db, err := initDatabase(cfg, logger)
	if err != nil {
		logger.Fatal("database connection failed", zap.Error(err))
	}
	logger.Info("database connected")

	// Auto-migrate: add new columns/tables without dropping existing data
	if err = db.AutoMigrate(
		&model.APIKey{},
		&model.LLMBackend{},
		&model.RateLimit{},
		&monitor.SystemMetric{},
		&monitor.LLMNodeMetric{},
		&model.Document{},
		&model.LLMTrainingData{},
		&model.TokenUsage{},
		&model.TokenUsageDaily{},
		&model.TokenUsageDailyKey{},
		&model.ThirdPartyProviderTemplate{},
		&model.UserThirdPartyProvider{},
		&model.ThirdPartyTokenUsage{},
	); err != nil {
		logger.Warn("AutoMigrate failed", zap.Error(err))
	}
	// Fix legacy data: populate period_hours for records with period_hours=0
	fixPeriodHours(db, logger)

	// 4. Connect to Redis
	rdb, err := initRedis(cfg, logger)
	if err != nil {
		logger.Fatal("Redis connection failed", zap.Error(err))
	}
	logger.Info("Redis connected")

	// 5. Initialize infrastructure
	jwtManager, err := jwtPkg.NewManager(cfg.JWT.Secret, cfg.JWT.ExpireHours, rdb)
	if err != nil {
		logger.Fatal("JWT manager init failed", zap.Error(err))
	}

	// LLM Provider manager with multi-provider and model routing support
	providerManager := initProviderManager(cfg, logger)

	// 6. Initialize repositories
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
	thirdPartyRepo := repository.NewThirdPartyProviderRepository(db)

	// 7. Initialize load balancer
	loadBalancer := llm.NewLoadBalancer(rdb, logger)

	// 8. Initialize services
	authService := service.NewAuthService(userRepo, auditRepo, jwtManager, logger)
	userService := service.NewUserService(userRepo, deptRepo, auditRepo, logger)
	deptService := service.NewDepartmentService(deptRepo, userRepo, auditRepo, logger)
	encryptor := crypto.NewEncryptor(cfg.JWT.Secret)
	apiKeyService := service.NewAPIKeyService(apiKeyRepo, auditRepo, rdb, logger, encryptor)
	limitService := service.NewLimitService(limitRepo, usageRepo, auditRepo, rdb, logger)
	trainingDataBuffer := service.NewTrainingDataBuffer(trainingDataRepo, logger)
	trainingDataArchiver := service.NewTrainingDataArchiver(trainingDataRepo, logger, "")
	dataRetentionCleaner := service.NewDataRetentionCleaner(usageRepo, logger, 0)
	llmProxyService := service.NewLLMProxyService(providerManager, loadBalancer, usageRepo, limitRepo, apiKeyRepo, trainingDataBuffer, sysConfigRepo, limitService, rdb, logger)
	statsService := service.NewStatsService(usageRepo, userRepo, deptRepo, apiKeyRepo, logger)
	systemService := service.NewSystemService(sysConfigRepo, auditRepo, annRepo, logger)
	mcpProxy := mcpPkg.NewProxy(logger)
	mcpService := service.NewMCPService(mcpRepo, mcpProxy, logger)
	llmBackendService := service.NewLLMBackendService(backendRepo, auditRepo, loadBalancer, logger)
	monitorService := service.NewMonitorService(monitorRepo, usageRepo, backendRepo, rdb, logger)
	docService := service.NewDocumentService(docRepo, logger)
	uploadService := service.NewUploadService(cfg.Upload.Dir, cfg.Upload.MaxSizeMB, cfg.Upload.URLPrefix, logger)
	trainingDataService := service.NewTrainingDataService(trainingDataRepo, logger)

	// Third-party provider service (uses JWT secret derived AES key for encryption)
	thirdPartyService := service.NewThirdPartyProviderService(thirdPartyRepo, backendRepo, encryptor, rdb, logger)
	llmProxyService.SetThirdPartyService(thirdPartyService)

	// Load LLM backend nodes from database into load balancer
	llmBackendService.RefreshLoadBalancer()

	// 9. Initialize handlers
	handlers := &router.Handlers{
		Auth:               handler.NewAuthHandler(authService),
		User:               handler.NewUserHandler(userService),
		Department:         handler.NewDepartmentHandler(deptService),
		APIKey:             handler.NewAPIKeyHandler(apiKeyService),
		LLMProxy:           handler.NewLLMProxyHandler(llmProxyService, logger),
		Stats:              handler.NewStatsHandler(statsService),
		Limit:              handler.NewLimitHandler(limitService),
		System:             handler.NewSystemHandler(systemService),
		MCPAdmin:           handler.NewMCPAdminHandler(mcpService, logger),
		MCPGateway:         handler.NewMCPGatewayHandler(mcpService, logger),
		LLMBackend:         handler.NewLLMBackendHandler(llmBackendService),
		Monitor:            handler.NewMonitorHandler(monitorService, logger),
		Document:           handler.NewDocumentHandler(docService),
		Upload:             handler.NewUploadHandler(uploadService),
		TrainingData:       handler.NewTrainingDataHandler(trainingDataService, logger),
		ThirdPartyProvider: handler.NewThirdPartyProviderHandler(thirdPartyService),
	}

	// 10. Initialize HTTP engine
	switch cfg.Server.Mode {
	case "release", "production":
		gin.SetMode(gin.ReleaseMode)
	case "test":
		gin.SetMode(gin.TestMode)
	}

	engine := gin.New()
	router.Setup(engine, handlers, jwtManager, db, rdb, logger, cfg.Server.CORSOrigins, cfg.Upload.Dir)

	// 11. Start HTTP server with graceful shutdown
	const llmStreamWriteTimeout = 600
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      engine,
		ReadTimeout:  30 * time.Second,                    //nolint:mnd // intentional constant.
		WriteTimeout: llmStreamWriteTimeout * time.Second, // LLM 流式请求需要较长超时
		IdleTimeout:  120 * time.Second,                   //nolint:mnd // intentional constant.
	}

	go func() {
		logger.Info("HTTP server starting", zap.String("addr", addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("HTTP server error", zap.Error(err))
		}
	}()

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutdown signal received, gracefully stopping...")

	// Allow 10 seconds for in-flight requests
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) //nolint:mnd // intentional constant.
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("server shutdown failed", zap.Error(err))
	}

	// Flush buffers to ensure all training data is persisted
	trainingDataBuffer.Close()
	trainingDataArchiver.Close()
	dataRetentionCleaner.Close()

	logger.Info("server stopped")
}

// initDatabase initializes database connection.
func initDatabase(cfg *config.Config, _ *zap.Logger) (*gorm.DB, error) {
	var logLevel gormLogger.LogLevel
	switch cfg.Server.Mode {
	case "debug":
		logLevel = gormLogger.Info
	case "release", "production":
		logLevel = gormLogger.Error
	default:
		logLevel = gormLogger.Warn
	}

	db, err := gorm.Open(postgres.Open(cfg.Database.DSN()), &gorm.Config{
		Logger: gormLogger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.Database.ConnMaxLifetimeMin) * time.Minute)

	return db, nil
}

// initProviderManager initializes multi-provider manager.
func initProviderManager(cfg *config.Config, logger *zap.Logger) *llm.ProviderManager {
	providers := cfg.LLM.GetEffectiveProviders()
	defaultName := cfg.LLM.GetDefaultProviderName()
	manager := llm.NewProviderManager(defaultName)

	for _, pc := range providers {
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
			logger.Info("registered Anthropic provider", zap.String("name", pc.Name), zap.String("base_url", pc.BaseURL))
		default:
			client := llm.NewClient(pc.BaseURL, pc.APIKey, pc.TimeoutSeconds, pc.StreamTimeoutSeconds)
			provider := llm.NewOpenAIProvider(pc.Name, client)
			manager.Register(provider)
			logger.Info("registered OpenAI provider", zap.String("name", pc.Name), zap.String("base_url", pc.BaseURL))
		}
	}

	if cfg.LLM.ModelRouting != nil {
		manager.SetModelRoutes(cfg.LLM.ModelRouting)
		logger.Info("model routing rules loaded", zap.Int("rules", len(cfg.LLM.ModelRouting)))
	}

	logger.Debug(manager.DebugRoutes())
	return manager
}

// fixPeriodHours migrates rate_limits table: populates period_hours and updates unique index.
func fixPeriodHours(db *gorm.DB, logger *zap.Logger) {
	// Populate period_hours for legacy records
	periodMap := map[string]int{
		"daily": 24, "weekly": 168, "monthly": 720, //nolint:mnd // intentional constant.
	}
	for period, hours := range periodMap {
		result := db.Exec(
			"UPDATE rate_limits SET period_hours = ? WHERE period = ? AND (period_hours = 0 OR period_hours IS NULL)",
			hours, period,
		)
		if result.RowsAffected > 0 {
			logger.Info("fixed period_hours for legacy records",
				zap.String("period", period),
				zap.Int("hours", hours),
				zap.Int64("rows", result.RowsAffected),
			)
		}
	}

	// Drop old unique index if it exists (based on period column instead of period_hours)
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

	// Create new unique index (idempotent)
	if err := db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_rate_limits_target
		ON rate_limits(target_type, target_id, period_hours);
	`).Error; err != nil {
		logger.Warn("failed to create idx_rate_limits_target index", zap.Error(err))
	}
}

// initRedis initializes Redis connection.
func initRedis(cfg *config.Config, _ *zap.Logger) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		PoolSize: cfg.Redis.PoolSize,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) //nolint:mnd // intentional constant.
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("Redis ping failed: %w", err)
	}

	return rdb, nil
}
