package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	"codemind/internal/model"
	"codemind/internal/repository"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	dedupKeyPrefix  = "codemind:training:hash:"
	defaultDedupTTL = 7 * 24 * time.Hour
)

// TrainingDataDeduplicator deduplicates training data using Redis.
type TrainingDataDeduplicator struct {
	lastRefresh   time.Time
	rdb           *redis.Client
	sysConfigRepo *repository.SystemRepository
	logger        *zap.Logger
	ttl           time.Duration
	enabled       bool
}

// NewTrainingDataDeduplicator creates a new deduplicator.
func NewTrainingDataDeduplicator(
	rdb *redis.Client,
	sysConfigRepo *repository.SystemRepository,
	logger *zap.Logger,
) *TrainingDataDeduplicator {
	return &TrainingDataDeduplicator{
		rdb:           rdb,
		sysConfigRepo: sysConfigRepo,
		logger:        logger,
		ttl:           defaultDedupTTL,
		enabled:       true,
	}
}

// IsEnabled returns whether deduplication is enabled.
func (d *TrainingDataDeduplicator) IsEnabled() bool {
	d.refreshConfigIfNeeded()
	return d.enabled
}

// ComputeContentHash computes a hash based on prompt and response.
func (d *TrainingDataDeduplicator) ComputeContentHash(requestBody, responseBody json.RawMessage) string {
	prompt := d.extractPrompt(requestBody)
	response := d.extractResponse(responseBody)

	combined := prompt + "|" + response
	if combined == "|" {
		return ""
	}

	hash := sha256.Sum256([]byte(combined))
	return hex.EncodeToString(hash[:16])
}

// IsDuplicate checks if content hash exists in Redis.
func (d *TrainingDataDeduplicator) IsDuplicate(contentHash string) bool {
	if !d.IsEnabled() || contentHash == "" || d.rdb == nil {
		return false
	}

	key := dedupKeyPrefix + contentHash
	exists, err := d.rdb.Exists(context.Background(), key).Result()
	if err != nil {
		d.logger.Warn("dedup check failed", zap.String("hash", contentHash), zap.Error(err))
		return false
	}
	return exists > 0
}

// MarkAsSeen marks a content hash as seen in Redis.
func (d *TrainingDataDeduplicator) MarkAsSeen(contentHash string) {
	if !d.IsEnabled() || contentHash == "" || d.rdb == nil {
		return
	}

	key := dedupKeyPrefix + contentHash
	if err := d.rdb.Set(context.Background(), key, "1", d.ttl).Err(); err != nil {
		d.logger.Warn("dedup mark failed", zap.String("hash", contentHash), zap.Error(err))
	}
}

// CheckAndMark atomically checks and marks content, returns true if duplicate.
func (d *TrainingDataDeduplicator) CheckAndMark(contentHash string) bool {
	if !d.IsEnabled() || contentHash == "" || d.rdb == nil {
		return false
	}

	key := dedupKeyPrefix + contentHash

	created, err := d.rdb.SetNX(context.Background(), key, "1", d.ttl).Result()
	if err != nil {
		d.logger.Warn("atomic dedup check failed", zap.String("hash", contentHash), zap.Error(err))
		return false
	}

	return !created
}

func (d *TrainingDataDeduplicator) extractPrompt(body json.RawMessage) string {
	if len(body) == 0 {
		return ""
	}

	var chatReq struct {
		Messages []struct {
			Content interface{} `json:"content"`
			Role    string      `json:"role"`
		} `json:"messages"`
	}
	if err := json.Unmarshal(body, &chatReq); err == nil && len(chatReq.Messages) > 0 {
		var result string
		for _, m := range chatReq.Messages {
			if m.Role == messageRoleUser {
				result += contentToString(m.Content)
			}
		}
		if result != "" {
			return result
		}
	}

	var compReq struct {
		Prompt interface{} `json:"prompt"`
	}
	if err := json.Unmarshal(body, &compReq); err == nil && compReq.Prompt != nil {
		return contentToString(compReq.Prompt)
	}

	return ""
}

func (d *TrainingDataDeduplicator) extractResponse(body json.RawMessage) string {
	if len(body) == 0 {
		return ""
	}

	var openaiResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
			Text string `json:"text"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(body, &openaiResp); err == nil && len(openaiResp.Choices) > 0 {
		if openaiResp.Choices[0].Message.Content != "" {
			return openaiResp.Choices[0].Message.Content
		}
		return openaiResp.Choices[0].Text
	}

	var anthropicResp struct {
		Role    string `json:"role"`
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(body, &anthropicResp); err == nil && anthropicResp.Role == "assistant" {
		var result string
		for _, b := range anthropicResp.Content {
			if b.Type == contentTypeText {
				result += b.Text
			}
		}
		if result != "" {
			return result
		}
	}

	return ""
}

func contentToString(content interface{}) string {
	switch v := content.(type) {
	case string:
		return v
	case []interface{}:
		for _, item := range v {
			if itemMap, ok := item.(map[string]interface{}); ok {
				if itemMap["type"] == contentTypeText {
					if text, ok := itemMap["text"].(string); ok {
						return text
					}
				}
			}
		}
	}
	return ""
}

func (d *TrainingDataDeduplicator) refreshConfigIfNeeded() {
	if time.Since(d.lastRefresh) < 60*time.Second {
		return
	}

	if d.sysConfigRepo != nil {
		if cfg, err := d.sysConfigRepo.GetByKey(model.ConfigTrainingDedupEnabled); err == nil {
			d.enabled = cfg.ConfigValue == configValueTrue
		}
	}
	d.lastRefresh = time.Now()
}
