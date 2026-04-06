package service

import (
	"encoding/json"
	"time"

	"codemind/internal/model"
	"codemind/internal/repository"

	"go.uber.org/zap"
)

// TrainingDataQualityScorer 质量评分器
type TrainingDataQualityScorer struct {
	sysConfigRepo *repository.SystemRepository
	logger        *zap.Logger

	// 配置缓存
	enabled     bool
	lastRefresh time.Time
}

// NewTrainingDataQualityScorer 创建评分器
func NewTrainingDataQualityScorer(
	sysConfigRepo *repository.SystemRepository,
	logger *zap.Logger,
) *TrainingDataQualityScorer {
	return &TrainingDataQualityScorer{
		sysConfigRepo: sysConfigRepo,
		logger:        logger,
		enabled:       true,
	}
}

// IsEnabled 检查是否启用质量评分
func (s *TrainingDataQualityScorer) IsEnabled() bool {
	s.refreshConfigIfNeeded()
	return s.enabled
}

// Score 计算质量分数 (0-100)
// 评分维度：
// - 响应长度适当 (0-25分)
// - Token 使用效率 (0-25分)
// - 请求成功 (0-20分)
// - 响应时间合理 (0-15分)
// - 内容多样性 (0-15分)
func (s *TrainingDataQualityScorer) Score(
	requestBody, responseBody json.RawMessage,
	promptTokens, completionTokens int,
	statusCode int,
	durationMs *int,
) *int {
	if !s.IsEnabled() {
		return nil
	}

	score := 0

	// 1. 响应长度评分 (25分)
	score += s.scoreResponseLength(completionTokens)

	// 2. Token 效率评分 (25分)
	score += s.scoreTokenEfficiency(promptTokens, completionTokens)

	// 3. 状态码评分 (20分)
	score += s.scoreStatusCode(statusCode)

	// 4. 响应时间评分 (15分)
	score += s.scoreResponseTime(durationMs)

	// 5. 内容多样性评分 (15分)
	score += s.scoreContentDiversity(requestBody)

	// 限制最大值
	if score > 100 {
		score = 100
	}

	return &score
}

// scoreResponseLength 响应长度评分
// 最佳范围：50-2000 tokens
func (s *TrainingDataQualityScorer) scoreResponseLength(tokens int) int {
	switch {
	case tokens < 10:
		return 5 // 太短，可能无意义
	case tokens < 50:
		return 15
	case tokens <= 500:
		return 25 // 最佳范围
	case tokens <= 2000:
		return 20
	case tokens <= 4000:
		return 10
	default:
		return 5 // 太长
	}
}

// scoreTokenEfficiency Token 效率评分
// completion/prompt 比例在 0.5-2.0 之间最佳
func (s *TrainingDataQualityScorer) scoreTokenEfficiency(prompt, completion int) int {
	if prompt == 0 || completion == 0 {
		return 10
	}

	ratio := float64(completion) / float64(prompt)

	switch {
	case ratio >= 0.5 && ratio <= 2.0:
		return 25 // 最佳比例
	case ratio >= 0.3 && ratio <= 3.0:
		return 18
	case ratio >= 0.1 && ratio <= 5.0:
		return 10
	default:
		return 5
	}
}

// scoreStatusCode 状态码评分
func (s *TrainingDataQualityScorer) scoreStatusCode(statusCode int) int {
	switch {
	case statusCode == 200:
		return 20
	case statusCode >= 200 && statusCode < 300:
		return 15
	case statusCode >= 400 && statusCode < 500:
		return 5 // 客户端错误
	case statusCode >= 500:
		return 0 // 服务端错误
	default:
		return 10
	}
}

// scoreResponseTime 响应时间评分
func (s *TrainingDataQualityScorer) scoreResponseTime(durationMs *int) int {
	if durationMs == nil {
		return 10 // 无数据给中等分
	}

	d := *durationMs

	switch {
	case d < 1000:
		return 15 // 快速响应
	case d < 3000:
		return 12 // 正常
	case d < 10000:
		return 8 // 较慢
	default:
		return 3 // 很慢
	}
}

// scoreContentDiversity 内容多样性评分
func (s *TrainingDataQualityScorer) scoreContentDiversity(body json.RawMessage) int {
	if len(body) == 0 {
		return 5
	}

	var req struct {
		Messages []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"messages"`
	}

	if err := json.Unmarshal(body, &req); err != nil {
		return 8
	}

	msgCount := len(req.Messages)

	switch {
	case msgCount >= 6:
		return 15 // 多轮对话，质量较高
	case msgCount >= 4:
		return 12
	case msgCount >= 2:
		return 10
	default:
		return 5 // 单轮对话
	}
}

// refreshConfigIfNeeded 按需刷新配置
func (s *TrainingDataQualityScorer) refreshConfigIfNeeded() {
	if time.Since(s.lastRefresh) < 60*time.Second {
		return
	}

	if s.sysConfigRepo != nil {
		if cfg, err := s.sysConfigRepo.GetByKey(model.ConfigTrainingQualityScoringEnabled); err == nil {
			s.enabled = cfg.ConfigValue == "true"
		}
	}
	s.lastRefresh = time.Now()
}
