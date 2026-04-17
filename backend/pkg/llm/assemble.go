package llm

import (
	"encoding/json"
	"time"
)

// AssembleChatResponse constructs equivalent non-streaming Chat Completions response JSON from stream result.
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

// AssembleCompletionResponse constructs equivalent non-streaming Completions response JSON from stream result.
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
