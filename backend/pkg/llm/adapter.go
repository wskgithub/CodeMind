package llm

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ──────────────────────────────────
// OpenAI → Anthropic 请求转换
// ──────────────────────────────────

// OpenAIToAnthropic 将 OpenAI ChatCompletion 请求转换为 Anthropic Messages 请求
func OpenAIToAnthropic(req *ChatCompletionRequest) *AnthropicMessagesRequest {
	anthropicReq := &AnthropicMessagesRequest{
		Model:       req.Model,
		Stream:      req.Stream,
		Temperature: req.Temperature,
		TopP:        req.TopP,
	}

	// 转换 max_tokens（Anthropic 必填，默认给一个合理值）
	if req.MaxTokens != nil {
		anthropicReq.MaxTokens = *req.MaxTokens
	} else {
		anthropicReq.MaxTokens = 4096
	}

	// 转换 stop 序列
	if req.Stop != nil {
		switch v := req.Stop.(type) {
		case string:
			anthropicReq.StopSequences = []string{v}
		case []interface{}:
			for _, s := range v {
				if str, ok := s.(string); ok {
					anthropicReq.StopSequences = append(anthropicReq.StopSequences, str)
				}
			}
		}
	}

	// 提取 system 消息并转换对话消息
	var systemParts []string
	var messages []AnthropicMessage

	for _, msg := range req.Messages {
		switch msg.Role {
		case "system":
			systemParts = append(systemParts, msg.Content)
		case "user":
			messages = append(messages, AnthropicMessage{
				Role:    "user",
				Content: msg.Content,
			})
		case "assistant":
			messages = append(messages, AnthropicMessage{
				Role:    "assistant",
				Content: msg.Content,
			})
		default:
			// 其他角色（如 function/tool）映射为 user 消息
			messages = append(messages, AnthropicMessage{
				Role:    "user",
				Content: fmt.Sprintf("[%s]: %s", msg.Role, msg.Content),
			})
		}
	}

	// 合并 system 消息
	if len(systemParts) > 0 {
		anthropicReq.System = strings.Join(systemParts, "\n\n")
	}

	// Anthropic 要求消息列表不能为空
	if len(messages) == 0 {
		messages = append(messages, AnthropicMessage{
			Role:    "user",
			Content: "Hello",
		})
	}

	// Anthropic 要求第一条消息必须是 user 角色
	if messages[0].Role != "user" {
		messages = append([]AnthropicMessage{{
			Role:    "user",
			Content: "Continue",
		}}, messages...)
	}

	anthropicReq.Messages = messages
	return anthropicReq
}

// ──────────────────────────────────
// Anthropic → OpenAI 请求转换
// ──────────────────────────────────

// AnthropicToOpenAI 将 Anthropic Messages 请求转换为 OpenAI ChatCompletion 请求
func AnthropicToOpenAI(req *AnthropicMessagesRequest) *ChatCompletionRequest {
	openaiReq := &ChatCompletionRequest{
		Model:       req.Model,
		Stream:      req.Stream,
		Temperature: req.Temperature,
		TopP:        req.TopP,
	}

	// 转换 max_tokens
	if req.MaxTokens > 0 {
		openaiReq.MaxTokens = &req.MaxTokens
	}

	// 转换 stop_sequences
	if len(req.StopSequences) > 0 {
		openaiReq.Stop = req.StopSequences
	}

	var messages []ChatMessage

	// 将 system 转换为 system 消息
	if req.System != nil {
		switch v := req.System.(type) {
		case string:
			if v != "" {
				messages = append(messages, ChatMessage{Role: "system", Content: v})
			}
		case []interface{}:
			// 处理 system block 数组
			var systemText []string
			for _, block := range v {
				if blockMap, ok := block.(map[string]interface{}); ok {
					if text, ok := blockMap["text"].(string); ok {
						systemText = append(systemText, text)
					}
				}
			}
			if len(systemText) > 0 {
				messages = append(messages, ChatMessage{Role: "system", Content: strings.Join(systemText, "\n\n")})
			}
		}
	}

	// 转换对话消息
	for _, msg := range req.Messages {
		content := extractAnthropicMessageContent(msg.Content)
		messages = append(messages, ChatMessage{
			Role:    msg.Role,
			Content: content,
		})
	}

	openaiReq.Messages = messages
	return openaiReq
}

// extractAnthropicMessageContent 从 Anthropic 消息中提取文本内容
// Anthropic 的 content 可以是 string 或 []ContentBlock
func extractAnthropicMessageContent(content interface{}) string {
	switch v := content.(type) {
	case string:
		return v
	case []interface{}:
		var parts []string
		for _, block := range v {
			if blockMap, ok := block.(map[string]interface{}); ok {
				if blockMap["type"] == "text" {
					if text, ok := blockMap["text"].(string); ok {
						parts = append(parts, text)
					}
				}
			}
		}
		return strings.Join(parts, "")
	default:
		// 尝试 JSON 序列化
		if data, err := json.Marshal(v); err == nil {
			return string(data)
		}
		return fmt.Sprintf("%v", v)
	}
}

// ──────────────────────────────────
// OpenAI → Anthropic 响应转换
// ──────────────────────────────────

// OpenAIResponseToAnthropic 将 OpenAI Chat 响应转换为 Anthropic Messages 响应
func OpenAIResponseToAnthropic(resp *ChatCompletionResponse) *AnthropicMessagesResponse {
	var content []AnthropicContentBlock

	for _, choice := range resp.Choices {
		if choice.Message != nil && choice.Message.Content != "" {
			content = append(content, AnthropicContentBlock{
				Type: "text",
				Text: choice.Message.Content,
			})
		}
	}

	// 转换停止原因
	var stopReason *string
	if len(resp.Choices) > 0 && resp.Choices[0].FinishReason != nil {
		reason := mapOpenAIStopReasonToAnthropic(*resp.Choices[0].FinishReason)
		stopReason = &reason
	}

	// 转换用量
	var usage *AnthropicUsage
	if resp.Usage != nil {
		usage = &AnthropicUsage{
			InputTokens:  resp.Usage.PromptTokens,
			OutputTokens: resp.Usage.CompletionTokens,
		}
	}

	return &AnthropicMessagesResponse{
		ID:         resp.ID,
		Type:       "message",
		Role:       "assistant",
		Content:    content,
		Model:      resp.Model,
		StopReason: stopReason,
		Usage:      usage,
	}
}

// ──────────────────────────────────
// Anthropic → OpenAI 响应转换
// ──────────────────────────────────

// AnthropicResponseToOpenAI 将 Anthropic Messages 响应转换为 OpenAI Chat 响应
func AnthropicResponseToOpenAI(resp *AnthropicMessagesResponse) *ChatCompletionResponse {
	// 从内容块中提取文本
	var textParts []string
	for _, block := range resp.Content {
		if block.Type == "text" {
			textParts = append(textParts, block.Text)
		}
	}
	content := strings.Join(textParts, "")

	// 转换停止原因
	var finishReason *string
	if resp.StopReason != nil {
		reason := mapAnthropicStopReasonToOpenAI(*resp.StopReason)
		finishReason = &reason
	}

	// 转换用量
	var usage *Usage
	if resp.Usage != nil {
		usage = resp.Usage.ToUsage()
	}

	return &ChatCompletionResponse{
		ID:      resp.ID,
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   resp.Model,
		Choices: []ChatChoice{
			{
				Index: 0,
				Message: &ChatMessage{
					Role:    "assistant",
					Content: content,
				},
				FinishReason: finishReason,
			},
		},
		Usage: usage,
	}
}

// ──────────────────────────────────
// 流式格式转换辅助
// ──────────────────────────────────

// OpenAIChunkToAnthropicEvents 将 OpenAI 流式 chunk 转换为 Anthropic 格式的 SSE 文本
// 返回需要写入客户端的原始 SSE 行
func OpenAIChunkToAnthropicEvents(chunk *ChatCompletionChunk, isFirst bool) string {
	var sb strings.Builder

	// 首个 chunk：发送 message_start 事件
	if isFirst {
		msgStart := fmt.Sprintf(
			`{"id":"%s","type":"message","role":"assistant","content":[],"model":"%s","stop_reason":null,"stop_sequence":null,"usage":{"input_tokens":0,"output_tokens":0}}`,
			chunk.ID, chunk.Model,
		)
		sb.WriteString(fmt.Sprintf("event: message_start\ndata: {\"type\":\"message_start\",\"message\":%s}\n\n", msgStart))
		sb.WriteString("event: content_block_start\ndata: {\"type\":\"content_block_start\",\"index\":0,\"content_block\":{\"type\":\"text\",\"text\":\"\"}}\n\n")
		sb.WriteString("event: ping\ndata: {\"type\":\"ping\"}\n\n")
	}

	// 提取文本增量
	if len(chunk.Choices) > 0 {
		delta := chunk.Choices[0].Delta
		if delta != nil && delta.Content != "" {
			// 对 JSON 进行转义
			textJSON, _ := json.Marshal(delta.Content)
			sb.WriteString(fmt.Sprintf(
				"event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":%s}}\n\n",
				string(textJSON),
			))
		}

		// 检查是否结束
		if chunk.Choices[0].FinishReason != nil {
			stopReason := mapOpenAIStopReasonToAnthropic(*chunk.Choices[0].FinishReason)
			sb.WriteString("event: content_block_stop\ndata: {\"type\":\"content_block_stop\",\"index\":0}\n\n")

			// 构建最终用量
			outputTokens := 0
			if chunk.Usage != nil {
				outputTokens = chunk.Usage.CompletionTokens
			}
			sb.WriteString(fmt.Sprintf(
				"event: message_delta\ndata: {\"type\":\"message_delta\",\"delta\":{\"stop_reason\":\"%s\",\"stop_sequence\":null},\"usage\":{\"output_tokens\":%d}}\n\n",
				stopReason, outputTokens,
			))
			sb.WriteString("event: message_stop\ndata: {\"type\":\"message_stop\"}\n\n")
		}
	}

	return sb.String()
}

// AnthropicEventToOpenAIChunk 将 Anthropic 流式事件转换为 OpenAI 格式的 SSE 文本
func AnthropicEventToOpenAIChunk(eventType string, event *AnthropicStreamEvent, model string) string {
	chunkID := "chatcmpl-" + uuid.New().String()[:8]

	switch eventType {
	case AnthropicEventContentBlockDelta:
		if event.Delta != nil && event.Delta.Text != "" {
			chunk := ChatCompletionChunk{
				ID:      chunkID,
				Object:  "chat.completion.chunk",
				Created: time.Now().Unix(),
				Model:   model,
				Choices: []ChatChoice{
					{
						Index: 0,
						Delta: &ChatMessage{
							Content: event.Delta.Text,
						},
					},
				},
			}
			data, _ := json.Marshal(chunk)
			return fmt.Sprintf("data: %s\n\n", string(data))
		}

	case AnthropicEventMessageDelta:
		if event.Delta != nil && event.Delta.StopReason != nil {
			reason := mapAnthropicStopReasonToOpenAI(*event.Delta.StopReason)
			chunk := ChatCompletionChunk{
				ID:      chunkID,
				Object:  "chat.completion.chunk",
				Created: time.Now().Unix(),
				Model:   model,
				Choices: []ChatChoice{
					{
						Index:        0,
						Delta:        &ChatMessage{},
						FinishReason: &reason,
					},
				},
			}
			// 附加 usage
			if event.Usage != nil {
				chunk.Usage = event.Usage.ToUsage()
			}
			data, _ := json.Marshal(chunk)
			return fmt.Sprintf("data: %s\n\n", string(data))
		}

	case AnthropicEventMessageStart:
		// 发送首个角色信息 chunk
		chunk := ChatCompletionChunk{
			ID:      chunkID,
			Object:  "chat.completion.chunk",
			Created: time.Now().Unix(),
			Model:   model,
			Choices: []ChatChoice{
				{
					Index: 0,
					Delta: &ChatMessage{
						Role: "assistant",
					},
				},
			},
		}
		data, _ := json.Marshal(chunk)
		return fmt.Sprintf("data: %s\n\n", string(data))

	case AnthropicEventMessageStop:
		return "data: [DONE]\n\n"
	}

	return ""
}

// ──────────────────────────────────
// 停止原因映射
// ──────────────────────────────────

// mapOpenAIStopReasonToAnthropic 将 OpenAI 停止原因映射为 Anthropic 格式
func mapOpenAIStopReasonToAnthropic(reason string) string {
	switch reason {
	case "stop":
		return "end_turn"
	case "length":
		return "max_tokens"
	case "tool_calls", "function_call":
		return "tool_use"
	default:
		return "end_turn"
	}
}

// mapAnthropicStopReasonToOpenAI 将 Anthropic 停止原因映射为 OpenAI 格式
func mapAnthropicStopReasonToOpenAI(reason string) string {
	switch reason {
	case "end_turn":
		return "stop"
	case "max_tokens":
		return "length"
	case "tool_use":
		return "tool_calls"
	case "stop_sequence":
		return "stop"
	default:
		return "stop"
	}
}
