package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Config 应用全局配置结构
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	LLM      LLMConfig      `mapstructure:"llm"`
	System   SystemConfig   `mapstructure:"system"`
	Log      LogConfig      `mapstructure:"log"`
}

// ServerConfig HTTP 服务器配置
type ServerConfig struct {
	Host           string   `mapstructure:"host"`
	Port           int      `mapstructure:"port"`
	Mode           string   `mapstructure:"mode"` // debug | release | production | test
	CORSOrigins    []string `mapstructure:"cors_origins"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host               string `mapstructure:"host"`
	Port               int    `mapstructure:"port"`
	Name               string `mapstructure:"name"`
	User               string `mapstructure:"user"`
	Password           string `mapstructure:"password"`
	SSLMode            string `mapstructure:"sslmode"`
	MaxOpenConns       int    `mapstructure:"max_open_conns"`
	MaxIdleConns       int    `mapstructure:"max_idle_conns"`
	ConnMaxLifetimeMin int    `mapstructure:"conn_max_lifetime_minutes"`
}

// DSN 返回 PostgreSQL 连接字符串
func (d *DatabaseConfig) DSN() string {
	sslmode := d.SSLMode
	if sslmode == "" {
		sslmode = "disable"
	}
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=Asia/Shanghai",
		d.Host, d.Port, d.User, d.Password, d.Name, sslmode,
	)
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

// Addr 返回 Redis 连接地址
func (r *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

// JWTConfig JWT 认证配置
type JWTConfig struct {
	Secret      string `mapstructure:"secret"`
	ExpireHours int    `mapstructure:"expire_hours"`
}

// LLMConfig LLM 服务配置（支持多 Provider 模式）
type LLMConfig struct {
	// 旧版单 Provider 配置（向后兼容）
	BaseURL              string `mapstructure:"base_url"`
	APIKey               string `mapstructure:"api_key"`
	TimeoutSeconds       int    `mapstructure:"timeout_seconds"`
	StreamTimeoutSeconds int    `mapstructure:"stream_timeout_seconds"`
	MaxRetries           int    `mapstructure:"max_retries"`

	// 新版多 Provider 配置
	Providers       []ProviderConfig  `mapstructure:"providers"`
	DefaultProvider string            `mapstructure:"default_provider"`
	ModelRouting    map[string]string `mapstructure:"model_routing"`
}

// ProviderConfig 单个 LLM Provider 配置
type ProviderConfig struct {
	Name                 string `mapstructure:"name"`
	Format               string `mapstructure:"format"` // "openai" 或 "anthropic"
	BaseURL              string `mapstructure:"base_url"`
	APIKey               string `mapstructure:"api_key"`
	TimeoutSeconds       int    `mapstructure:"timeout_seconds"`
	StreamTimeoutSeconds int    `mapstructure:"stream_timeout_seconds"`
}

// GetEffectiveProviders 获取有效的 Provider 配置列表
// 如果配置了新版多 Provider，使用新版；否则从旧版配置生成一个默认 Provider
func (c *LLMConfig) GetEffectiveProviders() []ProviderConfig {
	if len(c.Providers) > 0 {
		return c.Providers
	}
	// 向后兼容：从旧版配置生成默认 Provider
	return []ProviderConfig{
		{
			Name:                 "default",
			Format:               "openai",
			BaseURL:              c.BaseURL,
			APIKey:               c.APIKey,
			TimeoutSeconds:       c.TimeoutSeconds,
			StreamTimeoutSeconds: c.StreamTimeoutSeconds,
		},
	}
}

// GetDefaultProviderName 获取默认 Provider 名称
func (c *LLMConfig) GetDefaultProviderName() string {
	if c.DefaultProvider != "" {
		return c.DefaultProvider
	}
	providers := c.GetEffectiveProviders()
	if len(providers) > 0 {
		return providers[0].Name
	}
	return "default"
}

// SystemConfig 系统业务配置
type SystemConfig struct {
	MaxKeysPerUser       int  `mapstructure:"max_keys_per_user"`
	DefaultConcurrency   int  `mapstructure:"default_concurrency"`
	DefaultDailyTokens   int  `mapstructure:"default_daily_tokens"`
	DefaultMonthlyTokens int  `mapstructure:"default_monthly_tokens"`
	ForceChangePassword  bool `mapstructure:"force_change_password"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level    string `mapstructure:"level"`    // debug | info | warn | error
	Format   string `mapstructure:"format"`   // json | console
	Output   string `mapstructure:"output"`   // stdout | file
	FilePath string `mapstructure:"file_path"`
}

// 全局配置实例
var globalConfig *Config

// Load 加载配置文件并合并环境变量
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// 设置默认值
	setDefaults(v)

	// 读取配置文件
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		// 默认搜索路径
		v.SetConfigName("app")
		v.SetConfigType("yaml")
		v.AddConfigPath("./config")
		v.AddConfigPath("/app/config")
		v.AddConfigPath(".")
	}

	if err := v.ReadInConfig(); err != nil {
		// 配置文件不存在时不报错，使用默认值 + 环境变量
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("读取配置文件失败: %w", err)
		}
	}

	// 绑定环境变量（环境变量优先级高于配置文件）
	bindEnvVars(v)

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	globalConfig = cfg
	return cfg, nil
}

// Get 获取全局配置实例
func Get() *Config {
	return globalConfig
}

// setDefaults 设置配置默认值
func setDefaults(v *viper.Viper) {
	// 服务器默认配置
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.mode", "debug")

	// 数据库默认配置
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.name", "codemind")
	v.SetDefault("database.user", "codemind")
	v.SetDefault("database.max_open_conns", 50)
	v.SetDefault("database.max_idle_conns", 10)
	v.SetDefault("database.conn_max_lifetime_minutes", 60)

	// Redis 默认配置
	v.SetDefault("redis.host", "localhost")
	v.SetDefault("redis.port", 6379)
	v.SetDefault("redis.db", 0)
	v.SetDefault("redis.pool_size", 20)

	// JWT 默认配置
	v.SetDefault("jwt.expire_hours", 24)

	// LLM 默认配置
	v.SetDefault("llm.timeout_seconds", 300)
	v.SetDefault("llm.stream_timeout_seconds", 600)
	v.SetDefault("llm.max_retries", 0)

	// 系统默认配置
	v.SetDefault("system.max_keys_per_user", 10)
	v.SetDefault("system.default_concurrency", 5)
	v.SetDefault("system.default_daily_tokens", 1000000)
	v.SetDefault("system.default_monthly_tokens", 20000000)
	v.SetDefault("system.force_change_password", true)

	// 日志默认配置
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "json")
	v.SetDefault("log.output", "stdout")
	v.SetDefault("log.file_path", "./logs/app.log")
}

// bindEnvVars 绑定环境变量，环境变量优先级最高
func bindEnvVars(v *viper.Viper) {
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 显式绑定关键环境变量
	envBindings := map[string]string{
		"database.host":     "DB_HOST",
		"database.port":     "DB_PORT",
		"database.name":     "DB_NAME",
		"database.user":     "DB_USER",
		"database.password": "DB_PASSWORD",
		"database.sslmode":  "DB_SSLMODE",
		"redis.host":        "REDIS_HOST",
		"redis.port":        "REDIS_PORT",
		"redis.password":    "REDIS_PASSWORD",
		"jwt.secret":        "JWT_SECRET",
		"llm.base_url":      "LLM_BASE_URL",
		"llm.api_key":       "LLM_API_KEY",
		"server.mode":       "APP_ENV",
		"server.port":       "APP_PORT",
	}

	for configKey, envKey := range envBindings {
		_ = v.BindEnv(configKey, envKey)
	}
}

// InitLogger 初始化 Zap 日志器
func InitLogger(cfg *LogConfig) (*zap.Logger, error) {
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		level = zapcore.InfoLevel
	}

	var zapCfg zap.Config
	if cfg.Format == "console" {
		zapCfg = zap.NewDevelopmentConfig()
	} else {
		zapCfg = zap.NewProductionConfig()
	}

	zapCfg.Level = zap.NewAtomicLevelAt(level)

	// 配置输出目标
	if cfg.Output == "file" && cfg.FilePath != "" {
		zapCfg.OutputPaths = []string{cfg.FilePath}
		zapCfg.ErrorOutputPaths = []string{cfg.FilePath}
		// 确保日志目录存在
		if err := os.MkdirAll(cfg.FilePath[:strings.LastIndex(cfg.FilePath, "/")], 0755); err != nil {
			return nil, fmt.Errorf("创建日志目录失败: %w", err)
		}
	}

	logger, err := zapCfg.Build(zap.AddCallerSkip(1))
	if err != nil {
		return nil, fmt.Errorf("初始化日志器失败: %w", err)
	}

	return logger, nil
}
