package llm

import (
	"encoding/json"
	"time"
)

// AssembleChatResponse 从流式聚合结果构造等效的非流式 Chat Completions 响应 JSON
// 用于训练数据记录，使流式和非流式请求的存储格式统一
func AssembleChatResponse(result *StreamResult) json.RawMessage {
	if result == nil {
		return nil
	}

	finishReason := result.FinishReason
	if finishReason == "" {
		finishReason = "stop"
	}

	resp := ChatCompletionResponse{
		ID:      result.ResponseID,
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   result.Model,
		Choices: []ChatChoice{
			{
				Index: 0,
				Message: &ChatMessage{
					Role:    "assistant",
					Content: result.Content,
				},
				FinishReason: &finishReason,
			},
		},
		Usage: result.Usage,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		return nil
	}
	return data
}

// AssembleCompletionResponse 从流式聚合结果构造等效的非流式 Completions 响应 JSON
func AssembleCompletionResponse(result *StreamResult) json.RawMessage {
	if result == nil {
		return nil
	}

	finishReason := result.FinishReason
	if finishReason == "" {
		finishReason = "stop"
	}

	resp := CompletionResponse{
		ID:      result.ResponseID,
		Object:  "text_completion",
		Created: time.Now().Unix(),
		Model:   result.Model,
		Choices: []CompletionChoice{
			{
				Index:        0,
				Text:         result.Content,
				FinishReason: &finishReason,
			},
		},
		Usage: result.Usage,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		return nil
	}
	return data
}
