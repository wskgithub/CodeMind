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

// TrainingDataService 训练数据管理业务逻辑
type TrainingDataService struct {
	repo   *repository.TrainingDataRepository
	logger *zap.Logger
}

// NewTrainingDataService 创建训练数据管理服务
func NewTrainingDataService(repo *repository.TrainingDataRepository, logger *zap.Logger) *TrainingDataService {
	return &TrainingDataService{repo: repo, logger: logger}
}

// List 分页查询训练数据
func (s *TrainingDataService) List(filter repository.TrainingDataFilter) ([]model.LLMTrainingDataListItem, int64, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 || filter.PageSize > 100 {
		filter.PageSize = 20
	}
	return s.repo.List(filter)
}

// GetByID 获取单条训练数据详情
func (s *TrainingDataService) GetByID(id int64) (*model.LLMTrainingData, error) {
	return s.repo.GetByID(id)
}

// UpdateExcluded 更新训练数据排除状态
func (s *TrainingDataService) UpdateExcluded(id int64, excluded bool) error {
	return s.repo.UpdateExcluded(id, excluded)
}

// GetStats 获取训练数据统计
func (s *TrainingDataService) GetStats() (*model.TrainingDataStats, error) {
	return s.repo.GetStats()
}

// ExportJSONL 导出为 OpenAI fine-tuning JSONL 格式
// 通过流式写入 writer，避免一次性加载所有数据到内存
func (s *TrainingDataService) ExportJSONL(filter repository.TrainingDataFilter, w io.Writer) (int, error) {
	exported := 0

	err := s.repo.BatchIterator(filter, 500, func(batch []model.LLMTrainingData) error {
		for _, record := range batch {
			line, err := s.convertToTrainingFormat(record)
			if err != nil {
				s.logger.Warn("跳过无法转换的训练数据",
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

// convertToTrainingFormat 将单条记录转换为 OpenAI fine-tuning 格式
func (s *TrainingDataService) convertToTrainingFormat(record model.LLMTrainingData) ([]byte, error) {
	switch record.RequestType {
	case "chat_completion", "anthropic_messages":
		return s.convertChatToTraining(record)
	case "completion":
		return s.convertCompletionToTraining(record)
	default:
		return nil, fmt.Errorf("不支持的请求类型: %s", record.RequestType)
	}
}

// convertChatToTraining 将 chat completion 类型转换为训练格式
// 格式: {"messages": [{"role": "system", "content": "..."}, {"role": "user", "content": "..."}, {"role": "assistant", "content": "..."}]}
func (s *TrainingDataService) convertChatToTraining(record model.LLMTrainingData) ([]byte, error) {
	// 从请求体提取 messages
	var reqBody struct {
		Messages []json.RawMessage `json:"messages"`
	}
	if err := json.Unmarshal(record.RequestBody, &reqBody); err != nil {
		return nil, fmt.Errorf("解析请求体 messages 失败: %w", err)
	}

	if record.ResponseBody == nil {
		return nil, fmt.Errorf("响应体为空")
	}

	// 从响应体提取 assistant 回复
	var respBody struct {
		Choices []struct {
			Message struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		// Anthropic 格式兼容
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		Role string `json:"role"`
	}
	if err := json.Unmarshal(record.ResponseBody, &respBody); err != nil {
		return nil, fmt.Errorf("解析响应体失败: %w", err)
	}

	// 获取 assistant 回复内容
	var assistantContent string
	if len(respBody.Choices) > 0 {
		assistantContent = respBody.Choices[0].Message.Content
	} else if respBody.Role == "assistant" && len(respBody.Content) > 0 {
		// Anthropic 格式：从 content 块中拼接文本
		for _, block := range respBody.Content {
			if block.Type == "text" {
				assistantContent += block.Text
			}
		}
	}
	if assistantContent == "" {
		return nil, fmt.Errorf("无法提取 assistant 回复内容")
	}

	// 构造训练数据：原始 messages + assistant 回复
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

// convertCompletionToTraining 将 completion 类型转换为训练格式
func (s *TrainingDataService) convertCompletionToTraining(record model.LLMTrainingData) ([]byte, error) {
	var reqBody struct {
		Prompt interface{} `json:"prompt"`
	}
	if err := json.Unmarshal(record.RequestBody, &reqBody); err != nil {
		return nil, fmt.Errorf("解析请求体 prompt 失败: %w", err)
	}

	if record.ResponseBody == nil {
		return nil, fmt.Errorf("响应体为空")
	}

	var respBody struct {
		Choices []struct {
			Text string `json:"text"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(record.ResponseBody, &respBody); err != nil {
		return nil, fmt.Errorf("解析响应体失败: %w", err)
	}

	if len(respBody.Choices) == 0 {
		return nil, fmt.Errorf("响应体 choices 为空")
	}

	// 将 prompt 转为字符串
	promptStr := ""
	switch v := reqBody.Prompt.(type) {
	case string:
		promptStr = v
	default:
		promptBytes, _ := json.Marshal(v)
		promptStr = string(promptBytes)
	}

	// 转换为 chat 格式
	return json.Marshal(map[string]interface{}{
		"messages": []map[string]string{
			{"role": "user", "content": promptStr},
			{"role": "assistant", "content": respBody.Choices[0].Text},
		},
	})
}

// ExportFilename 生成导出文件名
func (s *TrainingDataService) ExportFilename() string {
	return fmt.Sprintf("training_data_%s.jsonl", time.Now().Format("20060102_150405"))
}
