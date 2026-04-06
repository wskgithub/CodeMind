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
	// 去重 Redis 键前缀
	dedupKeyPrefix = "codemind:training:hash:"
	// 默认去重窗口期（7天）
	defaultDedupTTL = 7 * 24 * time.Hour
)

// TrainingDataDeduplicator 训练数据去重器
type TrainingDataDeduplicator struct {
	rdb           *redis.Client
	sysConfigRepo *repository.SystemRepository
	logger        *zap.Logger
	ttl           time.Duration

	// 配置缓存
	enabled     bool
	lastRefresh time.Time
}

// NewTrainingDataDeduplicator 创建去重器
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

// IsEnabled 检查是否启用去重
func (d *TrainingDataDeduplicator) IsEnabled() bool {
	d.refreshConfigIfNeeded()
	return d.enabled
}

// ComputeContentHash 计算内容哈希
// 基于 prompt + response 组合生成唯一标识
func (d *TrainingDataDeduplicator) ComputeContentHash(requestBody, responseBody json.RawMessage) string {
	prompt := d.extractPrompt(requestBody)
	response := d.extractResponse(responseBody)

	// 组合计算哈希
	combined := prompt + "|" + response
	if combined == "|" {
		return "" // 空内容
	}

	hash := sha256.Sum256([]byte(combined))
	return hex.EncodeToString(hash[:16]) // 使用前 16 字节（32 字符）
}

// IsDuplicate 检查是否重复（使用 Redis）
func (d *TrainingDataDeduplicator) IsDuplicate(contentHash string) bool {
	if !d.IsEnabled() || contentHash == "" || d.rdb == nil {
		return false
	}

	key := dedupKeyPrefix + contentHash
	exists, err := d.rdb.Exists(context.Background(), key).Result()
	if err != nil {
		d.logger.Warn("检查去重失败", zap.String("hash", contentHash), zap.Error(err))
		return false
	}
	return exists > 0
}

// MarkAsSeen 标记为已见
func (d *TrainingDataDeduplicator) MarkAsSeen(contentHash string) {
	if !d.IsEnabled() || contentHash == "" || d.rdb == nil {
		return
	}

	key := dedupKeyPrefix + contentHash
	if err := d.rdb.Set(context.Background(), key, "1", d.ttl).Err(); err != nil {
		d.logger.Warn("标记去重失败", zap.String("hash", contentHash), zap.Error(err))
	}
}

// CheckAndMark 检查并标记（原子操作，返回是否为重复）
func (d *TrainingDataDeduplicator) CheckAndMark(contentHash string) bool {
	if !d.IsEnabled() || contentHash == "" || d.rdb == nil {
		return false // 未启用或无 Redis，不视为重复
	}

	key := dedupKeyPrefix + contentHash

	// 使用 SetNX 实现原子检查和设置
	created, err := d.rdb.SetNX(context.Background(), key, "1", d.ttl).Result()
	if err != nil {
		d.logger.Warn("原子去重检查失败", zap.String("hash", contentHash), zap.Error(err))
		return false
	}

	// created=true 表示新创建（非重复），created=false 表示已存在（重复）
	return !created
}

// extractPrompt 提取 prompt 内容
func (d *TrainingDataDeduplicator) extractPrompt(body json.RawMessage) string {
	if len(body) == 0 {
		return ""
	}

	// 尝试解析 Chat Completion 格式
	var chatReq struct {
		Messages []struct {
			Role    string `json:"role"`
			Content interface{} `json:"content"`
		} `json:"messages"`
	}
	if err := json.Unmarshal(body, &chatReq); err == nil && len(chatReq.Messages) > 0 {
		var result string
		for _, m := range chatReq.Messages {
			if m.Role == "user" {
				result += contentToString(m.Content)
			}
		}
		if result != "" {
			return result
		}
	}

	// 尝试解析 Completion 格式
	var compReq struct {
		Prompt interface{} `json:"prompt"`
	}
	if err := json.Unmarshal(body, &compReq); err == nil && compReq.Prompt != nil {
		return contentToString(compReq.Prompt)
	}

	return ""
}

// extractResponse 提取响应内容
func (d *TrainingDataDeduplicator) extractResponse(body json.RawMessage) string {
	if len(body) == 0 {
		return ""
	}

	// 尝试解析 OpenAI 格式
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

	// 尝试解析 Anthropic 格式
	var anthropicResp struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		Role string `json:"role"`
	}
	if err := json.Unmarshal(body, &anthropicResp); err == nil && anthropicResp.Role == "assistant" {
		var result string
		for _, b := range anthropicResp.Content {
			if b.Type == "text" {
				result += b.Text
			}
		}
		if result != "" {
			return result
		}
	}

	return ""
}

// contentToString 将 content 字段转换为字符串
func contentToString(content interface{}) string {
	switch v := content.(type) {
	case string:
		return v
	case []interface{}:
		// 多模态内容，提取文本
		for _, item := range v {
			if itemMap, ok := item.(map[string]interface{}); ok {
				if itemMap["type"] == "text" {
					if text, ok := itemMap["text"].(string); ok {
						return text
					}
				}
			}
		}
	}
	return ""
}

// refreshConfigIfNeeded 按需刷新配置
func (d *TrainingDataDeduplicator) refreshConfigIfNeeded() {
	if time.Since(d.lastRefresh) < 60*time.Second {
		return
	}

	if d.sysConfigRepo != nil {
		if cfg, err := d.sysConfigRepo.GetByKey(model.ConfigTrainingDedupEnabled); err == nil {
			d.enabled = cfg.ConfigValue == "true"
		}
	}
	d.lastRefresh = time.Now()
}
