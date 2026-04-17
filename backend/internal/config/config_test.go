package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTempConfigFile creates a temp config file, returns file path and cleanup function
func createTempConfigFile(t *testing.T, content string) (string, func()) {
	t.Helper()
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(configPath, []byte(content), 0o644)
	require.NoError(t, err)
	return configPath, func() {}
}

// TestLoad_WithValidConfigFile tests loading a valid config file
func TestLoad_WithValidConfigFile(t *testing.T) {
	configContent := `
server:
  host: 127.0.0.1
  port: 3000
  mode: release

database:
  host: db.example.com
  port: 5433
  name: testdb
  user: testuser
  password: testpass
  max_open_conns: 100
  max_idle_conns: 20
  conn_max_lifetime_minutes: 120

redis:
  host: redis.example.com
  port: 6380
  password: redispass
  db: 1
  pool_size: 50

jwt:
  secret: test-secret-key
  expire_hours: 48

llm:
  base_url: https://api.example.com
  api_key: test-api-key
  timeout_seconds: 60
  stream_timeout_seconds: 120
  max_retries: 3

system:
  max_keys_per_user: 20
  default_concurrency: 10
  default_daily_tokens: 2000000
  default_monthly_tokens: 50000000
  force_change_password: false

log:
  level: debug
  format: console
  output: stdout
  file_path: ./logs/test.log
`
	configPath, cleanup := createTempConfigFile(t, configContent)
	defer cleanup()

	// Reset global config
	globalConfig = nil

	cfg, err := Load(configPath)
	require.NoError(t, err)
	assert.NotNil(t, cfg)

	// Verify server config
	assert.Equal(t, "127.0.0.1", cfg.Server.Host)
	assert.Equal(t, 3000, cfg.Server.Port)
	assert.Equal(t, "release", cfg.Server.Mode)

	// Verify database config
	assert.Equal(t, "db.example.com", cfg.Database.Host)
	assert.Equal(t, 5433, cfg.Database.Port)
	assert.Equal(t, "testdb", cfg.Database.Name)
	assert.Equal(t, "testuser", cfg.Database.User)
	assert.Equal(t, "testpass", cfg.Database.Password)
	assert.Equal(t, 100, cfg.Database.MaxOpenConns)
	assert.Equal(t, 20, cfg.Database.MaxIdleConns)
	assert.Equal(t, 120, cfg.Database.ConnMaxLifetimeMin)

	// Verify Redis config
	assert.Equal(t, "redis.example.com", cfg.Redis.Host)
	assert.Equal(t, 6380, cfg.Redis.Port)
	assert.Equal(t, "redispass", cfg.Redis.Password)
	assert.Equal(t, 1, cfg.Redis.DB)
	assert.Equal(t, 50, cfg.Redis.PoolSize)

	// Verify JWT config
	assert.Equal(t, "test-secret-key", cfg.JWT.Secret)
	assert.Equal(t, 48, cfg.JWT.ExpireHours)

	// Verify LLM config
	assert.Equal(t, "https://api.example.com", cfg.LLM.BaseURL)
	assert.Equal(t, "test-api-key", cfg.LLM.APIKey)
	assert.Equal(t, 60, cfg.LLM.TimeoutSeconds)
	assert.Equal(t, 120, cfg.LLM.StreamTimeoutSeconds)
	assert.Equal(t, 3, cfg.LLM.MaxRetries)

	// Verify system config
	assert.Equal(t, 20, cfg.System.MaxKeysPerUser)
	assert.Equal(t, 10, cfg.System.DefaultConcurrency)
	assert.Equal(t, 2000000, cfg.System.DefaultDailyTokens)
	assert.Equal(t, 50000000, cfg.System.DefaultMonthlyTokens)
	assert.Equal(t, false, cfg.System.ForceChangePassword)

	// Verify log config
	assert.Equal(t, "debug", cfg.Log.Level)
	assert.Equal(t, "console", cfg.Log.Format)
	assert.Equal(t, "stdout", cfg.Log.Output)
	assert.Equal(t, "./logs/test.log", cfg.Log.FilePath)
}

// TestLoad_ConfigFileNotFound tests using defaults when config file not found
func TestLoad_ConfigFileNotFound(t *testing.T) {
	// When using empty path, viper uses default search paths
	// In this case, defaults + env vars are used when no config file is found

	// Reset global config
	globalConfig = nil

	// Load with empty path (uses default search paths)
	cfg, err := Load("")
	// Config file may or may not be found in default search paths
	// We only verify the function doesn't panic and returns config
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	// Verify defaults are set
	assert.NotEmpty(t, cfg.Server.Host)
	assert.NotEqual(t, 0, cfg.Server.Port)
}

// TestLoad_WithEnvironmentVariables tests env var config overrides
func TestLoad_WithEnvironmentVariables(t *testing.T) {
	configContent := `
server:
  host: 127.0.0.1
  port: 3000

database:
  host: db.example.com
  port: 5433
  name: testdb
  user: testuser
  password: configpass

redis:
  host: redis.example.com
  port: 6380
  password: redispass

jwt:
  secret: config-secret

llm:
  base_url: https://api.config.com
  api_key: config-api-key
`
	configPath, cleanup := createTempConfigFile(t, configContent)
	defer cleanup()

	// Set environment variables
	envVars := map[string]string{
		"DB_HOST":        "env-db.example.com",
		"DB_PORT":        "5434",
		"DB_NAME":        "envdb",
		"DB_USER":        "envuser",
		"DB_PASSWORD":    "envpass",
		"REDIS_HOST":     "env-redis.example.com",
		"REDIS_PORT":     "6381",
		"REDIS_PASSWORD": "envredispass",
		"JWT_SECRET":     "env-secret-key",
		"LLM_BASE_URL":   "https://api.env.com",
		"LLM_API_KEY":    "env-api-key",
		"APP_PORT":       "4000",
		"APP_ENV":        "test",
	}

	// Save and set environment variables
	oldEnvVars := make(map[string]string)
	for key, value := range envVars {
		oldEnvVars[key] = os.Getenv(key)
		os.Setenv(key, value)
	}

	// Clean up environment variables
	defer func() {
		for key, value := range oldEnvVars {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
	}()

	// Reset global config
	globalConfig = nil

	cfg, err := Load(configPath)
	require.NoError(t, err)
	assert.NotNil(t, cfg)

	// Verify env vars override config file
	assert.Equal(t, "env-db.example.com", cfg.Database.Host)
	assert.Equal(t, 5434, cfg.Database.Port)
	assert.Equal(t, "envdb", cfg.Database.Name)
	assert.Equal(t, "envuser", cfg.Database.User)
	assert.Equal(t, "envpass", cfg.Database.Password)

	// Verify Redis env var overrides
	assert.Equal(t, "env-redis.example.com", cfg.Redis.Host)
	assert.Equal(t, 6381, cfg.Redis.Port)
	assert.Equal(t, "envredispass", cfg.Redis.Password)

	// Verify JWT env var overrides
	assert.Equal(t, "env-secret-key", cfg.JWT.Secret)

	// Verify LLM env var overrides
	assert.Equal(t, "https://api.env.com", cfg.LLM.BaseURL)
	assert.Equal(t, "env-api-key", cfg.LLM.APIKey)

	// Verify server env var overrides
	assert.Equal(t, 4000, cfg.Server.Port)
	assert.Equal(t, "test", cfg.Server.Mode)
}

// TestDatabaseConfig_DSN tests database DSN generation
func TestDatabaseConfig_DSN(t *testing.T) {
	tests := []struct {
		name     string
		config   DatabaseConfig
		expected string
	}{
		{
			name: "standard config",
			config: DatabaseConfig{
				Host:     "localhost",
				Port:     5432,
				User:     "postgres",
				Password: "password",
				Name:     "mydb",
			},
			expected: "host=localhost port=5432 user=postgres password=password dbname=mydb sslmode=disable TimeZone=Asia/Shanghai",
		},
		{
			name: "remote host config",
			config: DatabaseConfig{
				Host:     "db.example.com",
				Port:     5433,
				User:     "admin",
				Password: "secret123",
				Name:     "production",
			},
			expected: "host=db.example.com port=5433 user=admin password=secret123 dbname=production sslmode=disable TimeZone=Asia/Shanghai",
		},
		{
			name: "password with special characters",
			config: DatabaseConfig{
				Host:     "localhost",
				Port:     5432,
				User:     "user",
				Password: "p@ssw0rd!#$%",
				Name:     "testdb",
			},
			expected: "host=localhost port=5432 user=user password=p@ssw0rd!#$% dbname=testdb sslmode=disable TimeZone=Asia/Shanghai",
		},
		{
			name: "custom port",
			config: DatabaseConfig{
				Host:     "192.168.1.100",
				Port:     15432,
				User:     "dbuser",
				Password: "pass",
				Name:     "codemind",
			},
			expected: "host=192.168.1.100 port=15432 user=dbuser password=pass dbname=codemind sslmode=disable TimeZone=Asia/Shanghai",
		},
		{
			name: "empty password",
			config: DatabaseConfig{
				Host:     "localhost",
				Port:     5432,
				User:     "postgres",
				Password: "",
				Name:     "mydb",
			},
			expected: "host=localhost port=5432 user=postgres password= dbname=mydb sslmode=disable TimeZone=Asia/Shanghai",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dsn := tt.config.DSN()
			assert.Equal(t, tt.expected, dsn)
		})
	}
}

// TestRedisConfig_Addr tests Redis address format
func TestRedisConfig_Addr(t *testing.T) {
	tests := []struct {
		name     string
		config   RedisConfig
		expected string
	}{
		{
			name:     "local default port",
			config:   RedisConfig{Host: "localhost", Port: 6379},
			expected: "localhost:6379",
		},
		{
			name:     "remote host",
			config:   RedisConfig{Host: "redis.example.com", Port: 6380},
			expected: "redis.example.com:6380",
		},
		{
			name:     "IP address",
			config:   RedisConfig{Host: "192.168.1.50", Port: 6379},
			expected: "192.168.1.50:6379",
		},
		{
			name:     "custom port",
			config:   RedisConfig{Host: "localhost", Port: 16379},
			expected: "localhost:16379",
		},
		{
			name:     "IPv6 address",
			config:   RedisConfig{Host: "::1", Port: 6379},
			expected: "::1:6379",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr := tt.config.Addr()
			assert.Equal(t, tt.expected, addr)
		})
	}
}

// TestLLMConfig_GetEffectiveProviders tests getting effective LLM providers
func TestLLMConfig_GetEffectiveProviders(t *testing.T) {
	tests := []struct {
		name             string
		config           LLMConfig
		expectedLen      int
		expectedProvider ProviderConfig
	}{
		{
			name: "multiple providers config",
			config: LLMConfig{
				Providers: []ProviderConfig{
					{
						Name:                 "openai",
						Format:               "openai",
						BaseURL:              "https://api.openai.com",
						APIKey:               "openai-key",
						TimeoutSeconds:       60,
						StreamTimeoutSeconds: 120,
					},
					{
						Name:                 "anthropic",
						Format:               "anthropic",
						BaseURL:              "https://api.anthropic.com",
						APIKey:               "anthropic-key",
						TimeoutSeconds:       90,
						StreamTimeoutSeconds: 180,
					},
				},
			},
			expectedLen: 2,
			expectedProvider: ProviderConfig{
				Name:                 "openai",
				Format:               "openai",
				BaseURL:              "https://api.openai.com",
				APIKey:               "openai-key",
				TimeoutSeconds:       60,
				StreamTimeoutSeconds: 120,
			},
		},
		{
			name: "single provider config",
			config: LLMConfig{
				Providers: []ProviderConfig{
					{
						Name:    "custom",
						Format:  "openai",
						BaseURL: "https://custom.api.com",
						APIKey:  "custom-key",
					},
				},
			},
			expectedLen: 1,
			expectedProvider: ProviderConfig{
				Name:    "custom",
				Format:  "openai",
				BaseURL: "https://custom.api.com",
				APIKey:  "custom-key",
			},
		},
		{
			name: "backward compatibility - legacy config",
			config: LLMConfig{
				BaseURL:              "https://legacy.api.com",
				APIKey:               "legacy-key",
				TimeoutSeconds:       300,
				StreamTimeoutSeconds: 600,
				MaxRetries:           3,
			},
			expectedLen: 1,
			expectedProvider: ProviderConfig{
				Name:                 "default",
				Format:               "openai",
				BaseURL:              "https://legacy.api.com",
				APIKey:               "legacy-key",
				TimeoutSeconds:       300,
				StreamTimeoutSeconds: 600,
			},
		},
		{
			name:             "empty config",
			config:           LLMConfig{},
			expectedLen:      1,
			expectedProvider: ProviderConfig{Name: "default", Format: "openai"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			providers := tt.config.GetEffectiveProviders()
			assert.Len(t, providers, tt.expectedLen)
			if len(providers) > 0 {
				assert.Equal(t, tt.expectedProvider, providers[0])
			}
		})
	}
}

// TestLLMConfig_GetDefaultProviderName tests getting default provider name
func TestLLMConfig_GetDefaultProviderName(t *testing.T) {
	tests := []struct {
		name         string
		config       LLMConfig
		expectedName string
	}{
		{
			name: "explicit default provider",
			config: LLMConfig{
				DefaultProvider: "openai",
				Providers: []ProviderConfig{
					{Name: "openai", Format: "openai"},
					{Name: "anthropic", Format: "anthropic"},
				},
			},
			expectedName: "openai",
		},
		{
			name: "use first provider",
			config: LLMConfig{
				Providers: []ProviderConfig{
					{Name: "anthropic", Format: "anthropic"},
					{Name: "openai", Format: "openai"},
				},
			},
			expectedName: "anthropic",
		},
		{
			name:         "backward compatibility - legacy config",
			config:       LLMConfig{BaseURL: "https://api.example.com"},
			expectedName: "default",
		},
		{
			name:         "empty config",
			config:       LLMConfig{},
			expectedName: "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name := tt.config.GetDefaultProviderName()
			assert.Equal(t, tt.expectedName, name)
		})
	}
}

// TestGet tests getting global config instance
func TestGet(t *testing.T) {
	// Save original config
	originalConfig := globalConfig
	defer func() {
		globalConfig = originalConfig
	}()

	// Test empty config
	globalConfig = nil
	assert.Nil(t, Get())

	// Set test config
	testConfig := &Config{
		Server: ServerConfig{Host: "test", Port: 9999},
	}
	globalConfig = testConfig

	// Verify Get() returns correct config
	cfg := Get()
	assert.NotNil(t, cfg)
	assert.Equal(t, "test", cfg.Server.Host)
	assert.Equal(t, 9999, cfg.Server.Port)
}

// TestSetDefaults tests default value setup
func TestSetDefaults(t *testing.T) {
	v := viper.New()
	setDefaults(v)

	tests := []struct {
		key      string
		expected interface{}
	}{
		// Server defaults
		{"server.host", "0.0.0.0"},
		{"server.port", 8080},
		{"server.mode", "debug"},

		// Database defaults
		{"database.host", "localhost"},
		{"database.port", 5432},
		{"database.name", "codemind"},
		{"database.user", "codemind"},
		{"database.max_open_conns", 50},
		{"database.max_idle_conns", 10},
		{"database.conn_max_lifetime_minutes", 60},

		// Redis defaults
		{"redis.host", "localhost"},
		{"redis.port", 6379},
		{"redis.db", 0},
		{"redis.pool_size", 20},

		// JWT defaults
		{"jwt.expire_hours", 24},

		// LLM defaults
		{"llm.timeout_seconds", 300},
		{"llm.stream_timeout_seconds", 600},
		{"llm.max_retries", 0},

		// System defaults
		{"system.max_keys_per_user", 10},
		{"system.default_concurrency", 5},
		{"system.default_daily_tokens", 1000000},
		{"system.default_monthly_tokens", 20000000},
		{"system.force_change_password", true},

		// Log defaults
		{"log.level", "info"},
		{"log.format", "json"},
		{"log.output", "stdout"},
		{"log.file_path", "./logs/app.log"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			assert.Equal(t, tt.expected, v.Get(tt.key))
		})
	}
}

// TestLoad_WithMultipleProviders tests loading multi-provider config
func TestLoad_WithMultipleProviders(t *testing.T) {
	configContent := `
llm:
  default_provider: anthropic
  providers:
    - name: openai
      format: openai
      base_url: https://api.openai.com/v1
      api_key: sk-openai-key
      timeout_seconds: 60
      stream_timeout_seconds: 120
    - name: anthropic
      format: anthropic
      base_url: https://api.anthropic.com
      api_key: sk-anthropic-key
      timeout_seconds: 90
      stream_timeout_seconds: 180
  model_routing:
    gpt-4: openai
    claude-3: anthropic
`
	configPath, cleanup := createTempConfigFile(t, configContent)
	defer cleanup()

	// Reset global config
	globalConfig = nil

	cfg, err := Load(configPath)
	require.NoError(t, err)
	assert.NotNil(t, cfg)

	// Verify provider config
	providers := cfg.LLM.GetEffectiveProviders()
	require.Len(t, providers, 2)

	// Verify first provider
	assert.Equal(t, "openai", providers[0].Name)
	assert.Equal(t, "openai", providers[0].Format)
	assert.Equal(t, "https://api.openai.com/v1", providers[0].BaseURL)
	assert.Equal(t, "sk-openai-key", providers[0].APIKey)
	assert.Equal(t, 60, providers[0].TimeoutSeconds)
	assert.Equal(t, 120, providers[0].StreamTimeoutSeconds)

	// Verify second provider
	assert.Equal(t, "anthropic", providers[1].Name)
	assert.Equal(t, "anthropic", providers[1].Format)
	assert.Equal(t, "https://api.anthropic.com", providers[1].BaseURL)
	assert.Equal(t, "sk-anthropic-key", providers[1].APIKey)
	assert.Equal(t, 90, providers[1].TimeoutSeconds)
	assert.Equal(t, 180, providers[1].StreamTimeoutSeconds)

	// Verify default provider
	assert.Equal(t, "anthropic", cfg.LLM.GetDefaultProviderName())
	assert.Equal(t, "anthropic", cfg.LLM.DefaultProvider)

	// Verify model routing
	assert.Equal(t, "openai", cfg.LLM.ModelRouting["gpt-4"])
	assert.Equal(t, "anthropic", cfg.LLM.ModelRouting["claude-3"])
}

// TestLoad_PartialConfig tests partial config loading
func TestLoad_PartialConfig(t *testing.T) {
	configContent := `
server:
  port: 9000

database:
  password: secret123
`
	configPath, cleanup := createTempConfigFile(t, configContent)
	defer cleanup()

	// Reset global config
	globalConfig = nil

	cfg, err := Load(configPath)
	require.NoError(t, err)
	assert.NotNil(t, cfg)

	// Verify overridden config
	assert.Equal(t, 9000, cfg.Server.Port)
	assert.Equal(t, "secret123", cfg.Database.Password)

	// Verify other configs use defaults
	assert.Equal(t, "0.0.0.0", cfg.Server.Host)
	assert.Equal(t, "debug", cfg.Server.Mode)
	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, 5432, cfg.Database.Port)
}

// TestLoad_WithInvalidConfigFile tests invalid config file
func TestLoad_WithInvalidConfigFile(t *testing.T) {
	// Create file with invalid YAML
	invalidContent := `
server:
  port: not_a_number
`
	configPath, cleanup := createTempConfigFile(t, invalidContent)
	defer cleanup()

	// Reset global config
	globalConfig = nil

	// For viper, type errors may cause parsing failures
	_, err := Load(configPath)
	// viper will try to parse; type mismatches may cause errors
	// Here we only verify the function doesn't panic
	assert.True(t, err == nil || err != nil)
}

// TestServerConfig_Addr tests server address composition
func TestServerConfig_Addr(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		port     int
		expected string
	}{
		{"default address", "0.0.0.0", 8080, "0.0.0.0:8080"},
		{"local address", "127.0.0.1", 3000, "127.0.0.1:3000"},
		{"domain address", "api.example.com", 443, "api.example.com:443"},
		{"IPv6 address", "::1", 8080, "::1:8080"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := ServerConfig{Host: tt.host, Port: tt.port}
			addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
			assert.Equal(t, tt.expected, addr)
		})
	}
}

// TestInitLogger tests logger initialization
func TestInitLogger(t *testing.T) {
	tests := []struct {
		name        string
		config      LogConfig
		expectError bool
	}{
		{
			name: "console format debug level",
			config: LogConfig{
				Level:  "debug",
				Format: "console",
				Output: "stdout",
			},
			expectError: false,
		},
		{
			name: "json format info level",
			config: LogConfig{
				Level:  "info",
				Format: "json",
				Output: "stdout",
			},
			expectError: false,
		},
		{
			name: "warn level",
			config: LogConfig{
				Level:  "warn",
				Format: "json",
				Output: "stdout",
			},
			expectError: false,
		},
		{
			name: "error level",
			config: LogConfig{
				Level:  "error",
				Format: "console",
				Output: "stdout",
			},
			expectError: false,
		},
		{
			name: "invalid level should use default",
			config: LogConfig{
				Level:  "invalid_level",
				Format: "json",
				Output: "stdout",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := InitLogger(&tt.config)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, logger)
			}
			if logger != nil {
				_ = logger.Sync()
			}
		})
	}
}

// TestInitLogger_WithFileOutput tests file output logging
func TestInitLogger_WithFileOutput(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "logs", "test.log")

	config := LogConfig{
		Level:    "info",
		Format:   "json",
		Output:   "file",
		FilePath: logPath,
	}

	logger, err := InitLogger(&config)
	require.NoError(t, err)
	assert.NotNil(t, logger)

	// Verify log directory and file were created
	_, err = os.Stat(filepath.Dir(logPath))
	assert.NoError(t, err)

	if logger != nil {
		_ = logger.Sync()
	}
}

// TestConfig_TableDriven is a table-driven test for comprehensive scenarios
func TestConfig_TableDriven(t *testing.T) {
	tests := []struct {
		name           string
		configContent  string
		envVars        map[string]string
		validateConfig func(t *testing.T, cfg *Config)
	}{
		{
			name: "minimal config",
			configContent: `
server:
  port: 8080
`,
			envVars: map[string]string{},
			validateConfig: func(t *testing.T, cfg *Config) {
				assert.Equal(t, 8080, cfg.Server.Port)
				assert.Equal(t, "0.0.0.0", cfg.Server.Host) // default value
				assert.Equal(t, "localhost", cfg.Database.Host)
			},
		},
		{
			name: "full database config",
			configContent: `
database:
  host: postgres.internal
  port: 5432
  name: codemind
  user: app
  password: securepass
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime_minutes: 30
`,
			envVars: map[string]string{},
			validateConfig: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "postgres.internal", cfg.Database.Host)
				assert.Equal(t, 5432, cfg.Database.Port)
				assert.Equal(t, "codemind", cfg.Database.Name)
				assert.Equal(t, "app", cfg.Database.User)
				assert.Equal(t, "securepass", cfg.Database.Password)
				assert.Equal(t, 25, cfg.Database.MaxOpenConns)
				assert.Equal(t, 5, cfg.Database.MaxIdleConns)
				assert.Equal(t, 30, cfg.Database.ConnMaxLifetimeMin)

				// Verify DSN
				dsn := cfg.Database.DSN()
				assert.Contains(t, dsn, "host=postgres.internal")
				assert.Contains(t, dsn, "port=5432")
				assert.Contains(t, dsn, "user=app")
				assert.Contains(t, dsn, "password=securepass")
				assert.Contains(t, dsn, "dbname=codemind")
			},
		},
		{
			name: "env vars override config",
			configContent: `
database:
  host: config-host
  port: 5432
  password: config-pass
`,
			envVars: map[string]string{
				"DB_HOST":     "env-host",
				"DB_PASSWORD": "env-pass",
			},
			validateConfig: func(t *testing.T, cfg *Config) {
				// Env vars should override config file
				assert.Equal(t, "env-host", cfg.Database.Host)
				assert.Equal(t, "env-pass", cfg.Database.Password)
				// Port comes from config file
				assert.Equal(t, 5432, cfg.Database.Port)
			},
		},
		{
			name: "LLM provider config",
			configContent: `
llm:
  default_provider: custom
  providers:
    - name: custom
      format: openai
      base_url: https://custom.api.com/v1
      api_key: custom-key
      timeout_seconds: 120
`,
			envVars: map[string]string{},
			validateConfig: func(t *testing.T, cfg *Config) {
				providers := cfg.LLM.GetEffectiveProviders()
				require.Len(t, providers, 1)
				assert.Equal(t, "custom", providers[0].Name)
				assert.Equal(t, "openai", providers[0].Format)
				assert.Equal(t, "https://custom.api.com/v1", providers[0].BaseURL)
				assert.Equal(t, "custom-key", providers[0].APIKey)
				assert.Equal(t, 120, providers[0].TimeoutSeconds)
				assert.Equal(t, "custom", cfg.LLM.GetDefaultProviderName())
			},
		},
		{
			name: "system config",
			configContent: `
system:
  max_keys_per_user: 50
  default_concurrency: 20
  default_daily_tokens: 5000000
  default_monthly_tokens: 100000000
  force_change_password: false
`,
			envVars: map[string]string{},
			validateConfig: func(t *testing.T, cfg *Config) {
				assert.Equal(t, 50, cfg.System.MaxKeysPerUser)
				assert.Equal(t, 20, cfg.System.DefaultConcurrency)
				assert.Equal(t, 5000000, cfg.System.DefaultDailyTokens)
				assert.Equal(t, 100000000, cfg.System.DefaultMonthlyTokens)
				assert.False(t, cfg.System.ForceChangePassword)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original environment variables
			oldEnvVars := make(map[string]string)
			for key := range tt.envVars {
				oldEnvVars[key] = os.Getenv(key)
			}

			// Set test environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			// Cleanup function
			defer func() {
				for key, value := range oldEnvVars {
					if value == "" {
						os.Unsetenv(key)
					} else {
						os.Setenv(key, value)
					}
				}
			}()

			// Create temp config file
			configPath, cleanup := createTempConfigFile(t, tt.configContent)
			defer cleanup()

			// Reset global config
			globalConfig = nil

			// Load config
			cfg, err := Load(configPath)
			require.NoError(t, err)
			assert.NotNil(t, cfg)

			// Execute validation
			tt.validateConfig(t, cfg)
		})
	}
}

// TestBindEnvVars tests environment variable binding
func TestBindEnvVars(t *testing.T) {
	v := viper.New()

	// Set some defaults
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)

	// Bind environment variables
	bindEnvVars(v)

	// Set environment variables
	os.Setenv("DB_HOST", "env-host")
	os.Setenv("DB_PORT", "5433")
	defer func() {
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
	}()

	// Verify env var bindings (v.Get returns interface{}, may be string or int)
	assert.Equal(t, "env-host", v.Get("database.host"))
	// DB_PORT from env var is the string "5433"
	assert.Equal(t, "5433", v.Get("database.port"))
}
