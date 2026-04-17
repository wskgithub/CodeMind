package service

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"codemind/internal/model"
	"codemind/internal/repository"

	"go.uber.org/zap"
)

// TrainingDataService handles training data management.
type TrainingDataService struct {
	repo   *repository.TrainingDataRepository
	logger *zap.Logger
}

// NewTrainingDataService creates a new training data service.
func NewTrainingDataService(repo *repository.TrainingDataRepository, logger *zap.Logger) *TrainingDataService {
	return &TrainingDataService{repo: repo, logger: logger}
}

// List returns paginated training data.
func (s *TrainingDataService) List(filter repository.TrainingDataFilter) ([]model.LLMTrainingDataListItem, int64, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 || filter.PageSize > 100 {
		filter.PageSize = 20
	}
	return s.repo.List(filter)
}

// GetByID returns a single training data record.
func (s *TrainingDataService) GetByID(id int64) (*model.LLMTrainingData, error) {
	return s.repo.GetByID(id)
}

// UpdateExcluded updates the excluded status of a training data record.
func (s *TrainingDataService) UpdateExcluded(id int64, excluded bool) error {
	return s.repo.UpdateExcluded(id, excluded)
}

// GetStats returns training data statistics.
func (s *TrainingDataService) GetStats() (*model.TrainingDataStats, error) {
	return s.repo.GetStats()
}

// ExportJSONL exports data in OpenAI fine-tuning JSONL format.
func (s *TrainingDataService) ExportJSONL(filter repository.TrainingDataFilter, w io.Writer) (int, error) {
	exported := 0

	//nolint:mnd // magic number for configuration/defaults.
	err := s.repo.BatchIterator(filter, 500, func(batch []model.LLMTrainingData) error {
		for _, record := range batch {
			line, err := s.convertToTrainingFormat(record)
			if err != nil {
				s.logger.Warn("skipping unconvertible training data",
					zap.Int64("id", record.ID),
					zap.Error(err),
				)
				continue
			}
			if _, err := fmt.Fprintf(w, "%s\n", line); err != nil {
				return err
			}
			exported++
		}
		return nil
	})

	return exported, err
}

func (s *TrainingDataService) convertToTrainingFormat(record model.LLMTrainingData) ([]byte, error) {
	switch record.RequestType {
	case "chat_completion", "anthropic_messages":
		return s.convertChatToTraining(record)
	case "completion":
		return s.convertCompletionToTraining(record)
	default:
		return nil, fmt.Errorf("unsupported request type: %s", record.RequestType)
	}
}

func (s *TrainingDataService) convertChatToTraining(record model.LLMTrainingData) ([]byte, error) {
	var reqBody struct {
		Messages []json.RawMessage `json:"messages"`
	}
	if err := json.Unmarshal(record.RequestBody, &reqBody); err != nil {
		return nil, fmt.Errorf("failed to parse request messages: %w", err)
	}

	if record.ResponseBody == nil {
		return nil, fmt.Errorf("response body is nil")
	}

	var respBody struct {
		Role    string `json:"role"`
		Choices []struct {
			Message struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(record.ResponseBody, &respBody); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	var assistantContent string
	if len(respBody.Choices) > 0 {
		assistantContent = respBody.Choices[0].Message.Content
	} else if respBody.Role == "assistant" && len(respBody.Content) > 0 {
		for _, block := range respBody.Content {
			if block.Type == contentTypeText {
				assistantContent += block.Text
			}
		}
	}
	if assistantContent == "" {
		return nil, fmt.Errorf("failed to extract assistant content")
	}

	assistantMsg, _ := json.Marshal(map[string]string{
		"role":    "assistant",
		"content": assistantContent,
	})

	allMessages := make([]json.RawMessage, 0, len(reqBody.Messages)+1)
	allMessages = append(allMessages, reqBody.Messages...)
	allMessages = append(allMessages, assistantMsg)

	return json.Marshal(map[string]interface{}{
		"messages": allMessages,
	})
}

func (s *TrainingDataService) convertCompletionToTraining(record model.LLMTrainingData) ([]byte, error) {
	var reqBody struct {
		Prompt interface{} `json:"prompt"`
	}
	if err := json.Unmarshal(record.RequestBody, &reqBody); err != nil {
		return nil, fmt.Errorf("failed to parse request prompt: %w", err)
	}

	if record.ResponseBody == nil {
		return nil, fmt.Errorf("response body is nil")
	}

	var respBody struct {
		Choices []struct {
			Text string `json:"text"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(record.ResponseBody, &respBody); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(respBody.Choices) == 0 {
		return nil, fmt.Errorf("response choices is empty")
	}

	promptStr := ""
	switch v := reqBody.Prompt.(type) {
	case string:
		promptStr = v
	default:
		promptBytes, _ := json.Marshal(v)
		promptStr = string(promptBytes)
	}

	return json.Marshal(map[string]interface{}{
		"messages": []map[string]string{
			{"role": "user", "content": promptStr},
			{"role": "assistant", "content": respBody.Choices[0].Text},
		},
	})
}

// ExportFilename generates export filename.
func (s *TrainingDataService) ExportFilename() string {
	return fmt.Sprintf("training_data_%s.jsonl", time.Now().Format("20060102_150405"))
}
