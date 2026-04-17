// Package config 管理应用配置的加载和验证。
package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Config holds the application configuration.
type Config struct {
	LLM      LLMConfig      `mapstructure:"llm"`
	Log      LogConfig      `mapstructure:"log"`
	Upload   UploadConfig   `mapstructure:"upload"`
	Server   ServerConfig   `mapstructure:"server"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Database DatabaseConfig `mapstructure:"database"`
	System   SystemConfig   `mapstructure:"system"`
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Host        string   `mapstructure:"host"`
	Mode        string   `mapstructure:"mode"`
	CORSOrigins []string `mapstructure:"cors_origins"`
	Port        int      `mapstructure:"port"`
}

// DatabaseConfig holds database connection settings.
type DatabaseConfig struct {
	Host               string `mapstructure:"host"`
	Name               string `mapstructure:"name"`
	User               string `mapstructure:"user"`
	Password           string `mapstructure:"password"`
	SSLMode            string `mapstructure:"sslmode"`
	Port               int    `mapstructure:"port"`
	MaxOpenConns       int    `mapstructure:"max_open_conns"`
	MaxIdleConns       int    `mapstructure:"max_idle_conns"`
	ConnMaxLifetimeMin int    `mapstructure:"conn_max_lifetime_minutes"`
}

// DSN returns the PostgreSQL connection string.
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

// RedisConfig holds Redis connection settings.
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Password string `mapstructure:"password"`
	Port     int    `mapstructure:"port"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

// Addr returns the Redis address.
func (r *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

// JWTConfig holds JWT authentication settings.
type JWTConfig struct {
	Secret      string `mapstructure:"secret"`
	ExpireHours int    `mapstructure:"expire_hours"`
}

// LLMConfig holds LLM service settings with multi-provider support.
type LLMConfig struct {
	ModelRouting         map[string]string `mapstructure:"model_routing"`
	BaseURL              string            `mapstructure:"base_url"`
	APIKey               string            `mapstructure:"api_key"`
	DefaultProvider      string            `mapstructure:"default_provider"`
	Providers            []ProviderConfig  `mapstructure:"providers"`
	TimeoutSeconds       int               `mapstructure:"timeout_seconds"`
	StreamTimeoutSeconds int               `mapstructure:"stream_timeout_seconds"`
	MaxRetries           int               `mapstructure:"max_retries"`
}

// ProviderConfig holds settings for a single LLM provider.
type ProviderConfig struct {
	Name                 string `mapstructure:"name"`
	Format               string `mapstructure:"format"`
	BaseURL              string `mapstructure:"base_url"`
	APIKey               string `mapstructure:"api_key"`
	TimeoutSeconds       int    `mapstructure:"timeout_seconds"`
	StreamTimeoutSeconds int    `mapstructure:"stream_timeout_seconds"`
}

// GetEffectiveProviders returns the provider list, falling back to legacy config if needed.
func (c *LLMConfig) GetEffectiveProviders() []ProviderConfig {
	if len(c.Providers) > 0 {
		return c.Providers
	}
	// Backward compatibility: create default provider from legacy config
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

// GetDefaultProviderName returns the default provider name.
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

// SystemConfig holds system-level business settings.
type SystemConfig struct {
	MaxKeysPerUser       int  `mapstructure:"max_keys_per_user"`
	DefaultConcurrency   int  `mapstructure:"default_concurrency"`
	DefaultDailyTokens   int  `mapstructure:"default_daily_tokens"`
	DefaultMonthlyTokens int  `mapstructure:"default_monthly_tokens"`
	ForceChangePassword  bool `mapstructure:"force_change_password"`
}

// UploadConfig holds file upload settings.
type UploadConfig struct {
	Dir       string `mapstructure:"dir"`
	URLPrefix string `mapstructure:"url_prefix"`
	MaxSizeMB int    `mapstructure:"max_size_mb"`
}

// LogConfig holds logging settings.
type LogConfig struct {
	Level    string `mapstructure:"level"`  // debug | info | warn | error
	Format   string `mapstructure:"format"` // json | console
	Output   string `mapstructure:"output"` // stdout | file
	FilePath string `mapstructure:"file_path"`
}

var globalConfig *Config

// Load loads configuration from file and environment variables.
func Load(configPath string) (*Config, error) {
	v := viper.New()

	setDefaults(v)

	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("app")
		v.SetConfigType("yaml")
		v.AddConfigPath("./config")
		v.AddConfigPath("/app/config")
		v.AddConfigPath(".")
	}

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
	}

	// Environment variables take precedence
	bindEnvVars(v)

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	globalConfig = cfg
	return cfg, nil
}

// Get returns the global config instance.
func Get() *Config {
	return globalConfig
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080) //nolint:mnd // intentional constant.
	v.SetDefault("server.mode", "debug")

	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432) //nolint:mnd // intentional constant.
	v.SetDefault("database.name", "codemind")
	v.SetDefault("database.user", "codemind")
	v.SetDefault("database.max_open_conns", 50)            //nolint:mnd // intentional constant.
	v.SetDefault("database.max_idle_conns", 10)            //nolint:mnd // intentional constant.
	v.SetDefault("database.conn_max_lifetime_minutes", 60) //nolint:mnd // intentional constant.

	v.SetDefault("redis.host", "localhost")
	v.SetDefault("redis.port", 6379) //nolint:mnd // intentional constant.
	v.SetDefault("redis.db", 0)
	v.SetDefault("redis.pool_size", 20) //nolint:mnd // intentional constant.

	v.SetDefault("jwt.expire_hours", 24) //nolint:mnd // intentional constant.

	v.SetDefault("llm.timeout_seconds", 300)        //nolint:mnd // intentional constant.
	v.SetDefault("llm.stream_timeout_seconds", 600) //nolint:mnd // intentional constant.
	v.SetDefault("llm.max_retries", 0)

	v.SetDefault("system.max_keys_per_user", 10)            //nolint:mnd // intentional constant.
	v.SetDefault("system.default_concurrency", 5)           //nolint:mnd // intentional constant.
	v.SetDefault("system.default_daily_tokens", 1000000)    //nolint:mnd // intentional constant.
	v.SetDefault("system.default_monthly_tokens", 20000000) //nolint:mnd // intentional constant.
	v.SetDefault("system.force_change_password", true)

	v.SetDefault("upload.dir", "./uploads")
	v.SetDefault("upload.max_size_mb", 5) //nolint:mnd // intentional constant.
	v.SetDefault("upload.url_prefix", "/uploads")

	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "json")
	v.SetDefault("log.output", "stdout")
	v.SetDefault("log.file_path", "./logs/app.log")
}

func bindEnvVars(v *viper.Viper) {
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

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

// InitLogger initializes the Zap logger.
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

	if cfg.Output == "file" && cfg.FilePath != "" {
		zapCfg.OutputPaths = []string{cfg.FilePath}
		zapCfg.ErrorOutputPaths = []string{cfg.FilePath}
		//nolint:mnd // magic number for configuration/defaults.
		if err := os.MkdirAll(cfg.FilePath[:strings.LastIndex(cfg.FilePath, "/")], 0o755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}
	}

	logger, err := zapCfg.Build(zap.AddCallerSkip(1))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	return logger, nil
}
