package service

import (
	"encoding/json"
	"time"

	"codemind/internal/model"
	"codemind/internal/repository"

	"go.uber.org/zap"
)

// TrainingDataQualityScorer scores training data quality.
type TrainingDataQualityScorer struct {
	sysConfigRepo *repository.SystemRepository
	logger        *zap.Logger

	enabled     bool
	lastRefresh time.Time
}

// NewTrainingDataQualityScorer creates a new quality scorer.
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

// IsEnabled returns whether quality scoring is enabled.
func (s *TrainingDataQualityScorer) IsEnabled() bool {
	s.refreshConfigIfNeeded()
	return s.enabled
}

// Score calculates quality score (0-100).
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

	score += s.scoreResponseLength(completionTokens)
	score += s.scoreTokenEfficiency(promptTokens, completionTokens)
	score += s.scoreStatusCode(statusCode)
	score += s.scoreResponseTime(durationMs)
	score += s.scoreContentDiversity(requestBody)

	if score > 100 {
		score = 100
	}

	return &score
}

func (s *TrainingDataQualityScorer) scoreResponseLength(tokens int) int {
	switch {
	case tokens < 10:
		return 5
	case tokens < 50:
		return 15
	case tokens <= 500:
		return 25
	case tokens <= 2000:
		return 20
	case tokens <= 4000:
		return 10
	default:
		return 5
	}
}

func (s *TrainingDataQualityScorer) scoreTokenEfficiency(prompt, completion int) int {
	if prompt == 0 || completion == 0 {
		return 10
	}

	ratio := float64(completion) / float64(prompt)

	switch {
	case ratio >= 0.5 && ratio <= 2.0:
		return 25
	case ratio >= 0.3 && ratio <= 3.0:
		return 18
	case ratio >= 0.1 && ratio <= 5.0:
		return 10
	default:
		return 5
	}
}

func (s *TrainingDataQualityScorer) scoreStatusCode(statusCode int) int {
	switch {
	case statusCode == 200:
		return 20
	case statusCode >= 200 && statusCode < 300:
		return 15
	case statusCode >= 400 && statusCode < 500:
		return 5
	case statusCode >= 500:
		return 0
	default:
		return 10
	}
}

func (s *TrainingDataQualityScorer) scoreResponseTime(durationMs *int) int {
	if durationMs == nil {
		return 10
	}

	d := *durationMs

	switch {
	case d < 1000:
		return 15
	case d < 3000:
		return 12
	case d < 10000:
		return 8
	default:
		return 3
	}
}

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
		return 15
	case msgCount >= 4:
		return 12
	case msgCount >= 2:
		return 10
	default:
		return 5
	}
}

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
