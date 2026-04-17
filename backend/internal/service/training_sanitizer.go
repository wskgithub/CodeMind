package service

import (
	"encoding/json"
	"regexp"
	"strings"
	"sync"
	"time"

	"codemind/internal/model"
	"codemind/internal/repository"

	"go.uber.org/zap"
)

// SensitivePattern defines a pattern for detecting sensitive information.
type SensitivePattern struct {
	Name    string
	Pattern *regexp.Regexp
	Replace string
}

// TrainingDataSanitizer redacts sensitive data from training records.
type TrainingDataSanitizer struct {
	sysConfigRepo *repository.SystemRepository
	logger        *zap.Logger

	mu          sync.RWMutex
	enabled     bool
	patterns    []string
	lastRefresh time.Time
}

var builtinPatterns = []SensitivePattern{
	// API Keys
	{Name: "openai_key", Pattern: regexp.MustCompile(`sk-[a-zA-Z0-9]{20,}`), Replace: "[API_KEY_REDACTED]"},
	{Name: "anthropic_key", Pattern: regexp.MustCompile(`sk-ant-[a-zA-Z0-9-]{20,}`), Replace: "[API_KEY_REDACTED]"},
	{Name: "generic_bearer", Pattern: regexp.MustCompile(`(?i)(api[_-]?key|apikey|access[_-]?key|secret[_-]?key)\s*[:=]\s*["']?[a-zA-Z0-9_-]{16,}["']?`), Replace: "$1=[REDACTED]"},

	// Tokens
	{Name: "bearer_token", Pattern: regexp.MustCompile(`(?i)Bearer\s+[a-zA-Z0-9._-]{20,}`), Replace: "Bearer [TOKEN_REDACTED]"},
	{Name: "jwt_token", Pattern: regexp.MustCompile(`eyJ[a-zA-Z0-9_-]*\.eyJ[a-zA-Z0-9_-]*\.[a-zA-Z0-9_-]*`), Replace: "[JWT_REDACTED]"},

	// Personal Info
	{Name: "email", Pattern: regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`), Replace: "[EMAIL_REDACTED]"},
	{Name: "phone_cn", Pattern: regexp.MustCompile(`1[3-9]\d{9}`), Replace: "[PHONE_REDACTED]"},
	{Name: "ip_v4", Pattern: regexp.MustCompile(`\b(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\b`), Replace: "[IP_REDACTED]"},

	// Passwords in text
	{Name: "password_assignment", Pattern: regexp.MustCompile(`(?i)(password|passwd|pwd)\s*[:=]\s*["']?[^\s"']{4,}["']?`), Replace: "$1=[REDACTED]"},
}

var defaultSensitiveKeys = []string{
	"password", "passwd", "pwd",
	"secret", "api_key", "apikey",
	"token", "authorization",
	"credential", "private_key",
	"access_key", "secret_key",
	"api_secret", "privatekey",
}

// NewTrainingDataSanitizer creates a new sanitizer instance.
func NewTrainingDataSanitizer(sysConfigRepo *repository.SystemRepository, logger *zap.Logger) *TrainingDataSanitizer {
	return &TrainingDataSanitizer{
		sysConfigRepo: sysConfigRepo,
		logger:        logger,
		enabled:       true,
		patterns:      defaultSensitiveKeys,
	}
}

// IsEnabled returns whether sanitization is enabled.
func (s *TrainingDataSanitizer) IsEnabled() bool {
	s.refreshConfigIfNeeded()
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.enabled
}

// SanitizeRequestBody redacts sensitive data from request body.
func (s *TrainingDataSanitizer) SanitizeRequestBody(body json.RawMessage) json.RawMessage {
	if !s.IsEnabled() || len(body) == 0 {
		return body
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return json.RawMessage(s.sanitizeString(string(body)))
	}

	s.sanitizeMap(data)
	result, _ := json.Marshal(data)
	return result
}

// SanitizeResponseBody redacts sensitive data from response body.
func (s *TrainingDataSanitizer) SanitizeResponseBody(body json.RawMessage) json.RawMessage {
	return s.SanitizeRequestBody(body)
}

func (s *TrainingDataSanitizer) sanitizeMap(m map[string]interface{}) {
	sensitiveKeys := s.getSensitiveKeys()

	for key, value := range m {
		lowerKey := strings.ToLower(key)

		isSensitive := false
		for _, sk := range sensitiveKeys {
			skLower := strings.ToLower(sk)
			if lowerKey == skLower ||
				strings.HasPrefix(lowerKey, skLower+"_") ||
				strings.HasPrefix(lowerKey, skLower+".") ||
				strings.HasSuffix(lowerKey, "_"+skLower) ||
				strings.HasSuffix(lowerKey, "."+skLower) ||
				strings.Contains(lowerKey, "_"+skLower+"_") ||
				strings.Contains(lowerKey, "."+skLower+"_") ||
				strings.Contains(lowerKey, "_"+skLower+".") ||
				strings.Contains(lowerKey, "."+skLower+".") {
				isSensitive = true
				break
			}
		}

		if isSensitive {
			m[key] = "[REDACTED]"
			continue
		}

		switch v := value.(type) {
		case map[string]interface{}:
			s.sanitizeMap(v)
		case []interface{}:
			for i, item := range v {
				if itemMap, ok := item.(map[string]interface{}); ok {
					s.sanitizeMap(itemMap)
				} else if str, ok := item.(string); ok {
					v[i] = s.sanitizeString(str)
				}
			}
		case string:
			m[key] = s.sanitizeString(v)
		}
	}
}

func (s *TrainingDataSanitizer) sanitizeString(str string) string {
	result := str
	for _, p := range builtinPatterns {
		result = p.Pattern.ReplaceAllString(result, p.Replace)
	}
	return result
}

func (s *TrainingDataSanitizer) getSensitiveKeys() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if len(s.patterns) > 0 {
		return s.patterns
	}
	return defaultSensitiveKeys
}

func (s *TrainingDataSanitizer) refreshConfigIfNeeded() {
	s.mu.RLock()
	if time.Since(s.lastRefresh) < 60*time.Second {
		s.mu.RUnlock()
		return
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()

	if time.Since(s.lastRefresh) < 60*time.Second {
		return
	}

	if s.sysConfigRepo != nil {
		if cfg, err := s.sysConfigRepo.GetByKey(model.ConfigTrainingSanitizeEnabled); err == nil {
			s.enabled = cfg.ConfigValue == "true"
		}
		if cfg, err := s.sysConfigRepo.GetByKey(model.ConfigTrainingSanitizePatterns); err == nil {
			var patterns []string
			if json.Unmarshal([]byte(cfg.ConfigValue), &patterns) == nil {
				s.patterns = patterns
			}
		}
	}
	s.lastRefresh = time.Now()
}
