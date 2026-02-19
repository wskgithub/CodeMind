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

	// 转换 max_tokens（Anthropic 必填）
	// 优先使用 max_completion_tokens（较新），其次 max_tokens
	if req.MaxCompletionTokens != nil {
		anthropicReq.MaxTokens = *req.MaxCompletionTokens
	} else if req.MaxTokens != nil {
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

	// 转换 tools
	for _, tool := range req.Tools {
		if tool.Type == "function" {
			anthropicReq.Tools = append(anthropicReq.Tools, AnthropicTool{
				Name:        tool.Function.Name,
				Description: tool.Function.Description,
				InputSchema: tool.Function.Parameters,
			})
		}
	}

	// 转换 tool_choice
	if req.ToolChoice != nil && len(anthropicReq.Tools) > 0 {
		anthropicReq.ToolChoice = mapOpenAIToolChoiceToAnthropic(req.ToolChoice)
	}

	// 提取 system 消息并转换对话消息
	var systemParts []string
	var messages []AnthropicMessage

	for _, msg := range req.Messages {
		switch msg.Role {
		case "system", "developer":
			systemParts = append(systemParts, msg.ContentString())
		case "user":
			messages = append(messages, AnthropicMessage{
				Role:    "user",
				Content: msg.Content,
			})
		case "assistant":
			anthropicMsg := AnthropicMessage{
				Role: "assistant",
			}
			// 转换 tool_calls 为 Anthropic 的 tool_use 内容块
			if len(msg.ToolCalls) > 0 {
				var blocks []AnthropicContentBlock
				text := msg.ContentString()
				if text != "" {
					blocks = append(blocks, AnthropicContentBlock{Type: "text", Text: text})
				}
				for _, tc := range msg.ToolCalls {
					var input interface{}
					json.Unmarshal([]byte(tc.Function.Arguments), &input)
					blocks = append(blocks, AnthropicContentBlock{
						Type:  "tool_use",
						ID:    tc.ID,
						Name:  tc.Function.Name,
						Input: input,
					})
				}
				anthropicMsg.Content = blocks
			} else {
				anthropicMsg.Content = msg.Content
			}
			messages = append(messages, anthropicMsg)
		case "tool":
			// OpenAI tool 消息 → Anthropic tool_result 内容块
			messages = append(messages, AnthropicMessage{
				Role: "user",
				Content: []AnthropicContentBlock{{
					Type:      "tool_result",
					ToolUseID: msg.ToolCallID,
					Content:   msg.Content,
				}},
			})
		default:
			messages = append(messages, AnthropicMessage{
				Role:    "user",
				Content: fmt.Sprintf("[%s]: %s", msg.Role, msg.ContentString()),
			})
		}
	}

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

	if req.MaxTokens > 0 {
		openaiReq.MaxTokens = &req.MaxTokens
	}

	if len(req.StopSequences) > 0 {
		openaiReq.Stop = req.StopSequences
	}

	// 转换 tools
	for _, tool := range req.Tools {
		openaiReq.Tools = append(openaiReq.Tools, Tool{
			Type: "function",
			Function: ToolFunction{
				Name:        tool.Name,
				Description: tool.Description,
				Parameters:  tool.InputSchema,
			},
		})
	}

	// 转换 tool_choice
	if req.ToolChoice != nil && len(openaiReq.Tools) > 0 {
		openaiReq.ToolChoice = mapAnthropicToolChoiceToOpenAI(req.ToolChoice)
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

	// 转换对话消息，包括 tool_use 和 tool_result
	for _, msg := range req.Messages {
		openaiMsg := convertAnthropicMessageToOpenAI(msg)
		messages = append(messages, openaiMsg...)
	}

	openaiReq.Messages = messages
	return openaiReq
}

// convertAnthropicMessageToOpenAI 转换单条 Anthropic 消息为 OpenAI 格式
// 可能返回多条消息（如 tool_result 需要拆分为多条 tool 消息）
func convertAnthropicMessageToOpenAI(msg AnthropicMessage) []ChatMessage {
	// content 为纯文本的简单场景
	if textContent, ok := msg.Content.(string); ok {
		return []ChatMessage{{
			Role:    msg.Role,
			Content: textContent,
		}}
	}

	// content 为内容块数组
	blocks, ok := msg.Content.([]interface{})
	if !ok {
		return []ChatMessage{{
			Role:    msg.Role,
			Content: extractAnthropicMessageContent(msg.Content),
		}}
	}

	if msg.Role == "assistant" {
		// assistant 消息：提取文本、tool_use，跳过 thinking（OpenAI 无对应概念）
		var textParts []string
		var toolCalls []ToolCall
		tcIdx := 0
		for _, block := range blocks {
			blockMap, ok := block.(map[string]interface{})
			if !ok {
				continue
			}
			switch blockMap["type"] {
			case "text":
				if text, ok := blockMap["text"].(string); ok {
					textParts = append(textParts, text)
				}
			case "tool_use":
				tc := ToolCall{
					Type: "function",
					Function: ToolCallFunction{
						Name: fmt.Sprintf("%v", blockMap["name"]),
					},
				}
				if id, ok := blockMap["id"].(string); ok {
					tc.ID = id
				}
				if input, ok := blockMap["input"]; ok {
					if argBytes, err := json.Marshal(input); err == nil {
						tc.Function.Arguments = string(argBytes)
					}
				}
				idx := tcIdx
				tc.Index = &idx
				toolCalls = append(toolCalls, tc)
				tcIdx++
			case "thinking":
				// thinking 块无 OpenAI 对应物，静默跳过
			}
		}
		result := ChatMessage{
			Role:    "assistant",
			Content: strings.Join(textParts, ""),
		}
		if len(toolCalls) > 0 {
			result.ToolCalls = toolCalls
		}
		return []ChatMessage{result}
	}

	// user 消息：检查 tool_result、text、image 块
	var regularParts []string
	var imageParts []ContentPart
	var toolResults []ChatMessage
	for _, block := range blocks {
		blockMap, ok := block.(map[string]interface{})
		if !ok {
			continue
		}
		switch blockMap["type"] {
		case "tool_result":
			toolMsg := ChatMessage{
				Role: "tool",
			}
			if id, ok := blockMap["tool_use_id"].(string); ok {
				toolMsg.ToolCallID = id
			}
			if content, ok := blockMap["content"]; ok {
				toolMsg.Content = extractAnthropicMessageContent(content)
			}
			toolResults = append(toolResults, toolMsg)
		case "text":
			if text, ok := blockMap["text"].(string); ok {
				regularParts = append(regularParts, text)
			}
		case "image":
			// Anthropic base64 图片 → OpenAI image_url (data URI)
			if source, ok := blockMap["source"].(map[string]interface{}); ok {
				mediaType, _ := source["media_type"].(string)
				data, _ := source["data"].(string)
				if mediaType != "" && data != "" {
					imageParts = append(imageParts, ContentPart{
						Type: "image_url",
						ImageURL: &ImageURL{
							URL: fmt.Sprintf("data:%s;base64,%s", mediaType, data),
						},
					})
				}
			}
		}
	}

	var result []ChatMessage
	if len(toolResults) > 0 {
		result = append(result, toolResults...)
	}
	// 混合文本+图片时使用 multimodal content parts
	if len(imageParts) > 0 {
		var parts []ContentPart
		for _, text := range regularParts {
			parts = append(parts, ContentPart{Type: "text", Text: text})
		}
		parts = append(parts, imageParts...)
		result = append(result, ChatMessage{
			Role:    "user",
			Content: parts,
		})
	} else if len(regularParts) > 0 {
		result = append(result, ChatMessage{
			Role:    "user",
			Content: strings.Join(regularParts, ""),
		})
	}
	if len(result) == 0 {
		result = append(result, ChatMessage{
			Role:    msg.Role,
			Content: extractAnthropicMessageContent(msg.Content),
		})
	}
	return result
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
		if choice.Message == nil {
			continue
		}
		// 转换文本内容
		text := choice.Message.ContentString()
		if text != "" {
			content = append(content, AnthropicContentBlock{
				Type: "text",
				Text: text,
			})
		}
		// 转换 tool_calls 为 Anthropic 的 tool_use 内容块
		for _, tc := range choice.Message.ToolCalls {
			var input interface{}
			json.Unmarshal([]byte(tc.Function.Arguments), &input)
			content = append(content, AnthropicContentBlock{
				Type:  "tool_use",
				ID:    tc.ID,
				Name:  tc.Function.Name,
				Input: input,
			})
		}
	}

	var stopReason *string
	if len(resp.Choices) > 0 && resp.Choices[0].FinishReason != nil {
		reason := mapOpenAIStopReasonToAnthropic(*resp.Choices[0].FinishReason)
		stopReason = &reason
	}

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
	var textParts []string
	var toolCalls []ToolCall
	tcIndex := 0

	for _, block := range resp.Content {
		switch block.Type {
		case "text":
			textParts = append(textParts, block.Text)
		case "tool_use":
			var args string
			if block.Input != nil {
				if argBytes, err := json.Marshal(block.Input); err == nil {
					args = string(argBytes)
				}
			}
			idx := tcIndex
			toolCalls = append(toolCalls, ToolCall{
				ID:   block.ID,
				Type: "function",
				Function: ToolCallFunction{
					Name:      block.Name,
					Arguments: args,
				},
				Index: &idx,
			})
			tcIndex++
		case "thinking":
			// thinking 块无 OpenAI 对应物，跳过
		}
	}

	message := &ChatMessage{
		Role:    "assistant",
		Content: strings.Join(textParts, ""),
	}
	if len(toolCalls) > 0 {
		message.ToolCalls = toolCalls
	}

	var finishReason *string
	if resp.StopReason != nil {
		reason := mapAnthropicStopReasonToOpenAI(*resp.StopReason)
		finishReason = &reason
	}

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
				Index:        0,
				Message:      message,
				FinishReason: finishReason,
			},
		},
		Usage: usage,
	}
}

// ──────────────────────────────────
// 流式格式转换辅助
// ──────────────────────────────────

// OpenAIToAnthropicState 跟踪 OpenAI → Anthropic 流式转换中的内容块索引
type OpenAIToAnthropicState struct {
	ContentIndex int // 当前 Anthropic 内容块索引
	InToolCall   bool
}

// OpenAIChunkToAnthropicEvents 将 OpenAI 流式 chunk 转换为 Anthropic 格式的 SSE 文本
// state 用于跟踪跨 chunk 的内容块索引，调用方需在整个流中维持同一实例
func OpenAIChunkToAnthropicEvents(chunk *ChatCompletionChunk, isFirst bool, state *OpenAIToAnthropicState) string {
	var sb strings.Builder

	if isFirst {
		msgStart := fmt.Sprintf(
			`{"id":"%s","type":"message","role":"assistant","content":[],"model":"%s","stop_reason":null,"stop_sequence":null,"usage":{"input_tokens":0,"output_tokens":0}}`,
			chunk.ID, chunk.Model,
		)
		sb.WriteString(fmt.Sprintf("event: message_start\ndata: {\"type\":\"message_start\",\"message\":%s}\n\n", msgStart))
		sb.WriteString("event: ping\ndata: {\"type\":\"ping\"}\n\n")
	}

	if len(chunk.Choices) == 0 {
		return sb.String()
	}

	delta := chunk.Choices[0].Delta
	if delta != nil {
		// 文本增量
		text := delta.ContentString()
		if text != "" {
			if state.ContentIndex == 0 && !state.InToolCall && isFirst {
				sb.WriteString(fmt.Sprintf(
					"event: content_block_start\ndata: {\"type\":\"content_block_start\",\"index\":%d,\"content_block\":{\"type\":\"text\",\"text\":\"\"}}\n\n",
					state.ContentIndex,
				))
			}
			textJSON, _ := json.Marshal(text)
			sb.WriteString(fmt.Sprintf(
				"event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":%d,\"delta\":{\"type\":\"text_delta\",\"text\":%s}}\n\n",
				state.ContentIndex, string(textJSON),
			))
		}

		// 工具调用增量
		if len(delta.ToolCalls) > 0 {
			for _, tc := range delta.ToolCalls {
				if tc.Function.Name != "" {
					// 新工具调用开始：关闭前一个内容块，开启 tool_use 块
					if state.ContentIndex > 0 || state.InToolCall {
						sb.WriteString(fmt.Sprintf(
							"event: content_block_stop\ndata: {\"type\":\"content_block_stop\",\"index\":%d}\n\n",
							state.ContentIndex,
						))
						state.ContentIndex++
					} else if isFirst {
						state.ContentIndex = 0
					}
					state.InToolCall = true
					toolID := tc.ID
					if toolID == "" {
						toolID = "toolu_" + uuid.New().String()[:8]
					}
					nameJSON, _ := json.Marshal(tc.Function.Name)
					sb.WriteString(fmt.Sprintf(
						"event: content_block_start\ndata: {\"type\":\"content_block_start\",\"index\":%d,\"content_block\":{\"type\":\"tool_use\",\"id\":\"%s\",\"name\":%s,\"input\":{}}}\n\n",
						state.ContentIndex, toolID, string(nameJSON),
					))
				}
				// 工具调用参数增量
				if tc.Function.Arguments != "" {
					argsJSON, _ := json.Marshal(tc.Function.Arguments)
					sb.WriteString(fmt.Sprintf(
						"event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":%d,\"delta\":{\"type\":\"input_json_delta\",\"partial_json\":%s}}\n\n",
						state.ContentIndex, string(argsJSON),
					))
				}
			}
		}
	}

	// 结束事件
	if chunk.Choices[0].FinishReason != nil {
		stopReason := mapOpenAIStopReasonToAnthropic(*chunk.Choices[0].FinishReason)
		sb.WriteString(fmt.Sprintf(
			"event: content_block_stop\ndata: {\"type\":\"content_block_stop\",\"index\":%d}\n\n",
			state.ContentIndex,
		))
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

	return sb.String()
}

// AnthropicToOpenAIState 跟踪 Anthropic → OpenAI 流式转换中的工具调用状态
type AnthropicToOpenAIState struct {
	CurrentBlockType string // 当前 content_block 的类型
	CurrentBlockID   string // 当前 tool_use 块的 ID
	CurrentBlockName string // 当前 tool_use 块的工具名
	ToolCallIndex    int    // 工具调用累计索引
}

// AnthropicEventToOpenAIChunk 将 Anthropic 流式事件转换为 OpenAI 格式的 SSE 文本
// state 用于跟踪跨事件的工具调用索引，调用方需在整个流中维持同一实例
func AnthropicEventToOpenAIChunk(eventType string, event *AnthropicStreamEvent, model string, state *AnthropicToOpenAIState) string {
	chunkID := "chatcmpl-" + uuid.New().String()[:8]
	now := time.Now().Unix()

	switch eventType {
	case AnthropicEventMessageStart:
		chunk := ChatCompletionChunk{
			ID: chunkID, Object: "chat.completion.chunk", Created: now, Model: model,
			Choices: []ChatChoice{{Index: 0, Delta: &ChatMessage{Role: "assistant"}}},
		}
		data, _ := json.Marshal(chunk)
		return fmt.Sprintf("data: %s\n\n", string(data))

	case AnthropicEventContentBlockStart:
		if event.ContentBlock != nil {
			state.CurrentBlockType = event.ContentBlock.Type
			switch event.ContentBlock.Type {
			case "tool_use":
				state.CurrentBlockID = event.ContentBlock.ID
				state.CurrentBlockName = event.ContentBlock.Name
				idx := state.ToolCallIndex
				chunk := ChatCompletionChunk{
					ID: chunkID, Object: "chat.completion.chunk", Created: now, Model: model,
					Choices: []ChatChoice{{
						Index: 0,
						Delta: &ChatMessage{
							ToolCalls: []ToolCall{{
								Index:    &idx,
								ID:       event.ContentBlock.ID,
								Type:     "function",
								Function: ToolCallFunction{Name: event.ContentBlock.Name},
							}},
						},
					}},
				}
				data, _ := json.Marshal(chunk)
				return fmt.Sprintf("data: %s\n\n", string(data))
			case "thinking":
				// thinking 块无 OpenAI 对应物，不输出
			}
		}

	case AnthropicEventContentBlockDelta:
		if event.Delta == nil {
			return ""
		}
		switch event.Delta.Type {
		case "text_delta":
			if event.Delta.Text != "" {
				chunk := ChatCompletionChunk{
					ID: chunkID, Object: "chat.completion.chunk", Created: now, Model: model,
					Choices: []ChatChoice{{Index: 0, Delta: &ChatMessage{Content: event.Delta.Text}}},
				}
				data, _ := json.Marshal(chunk)
				return fmt.Sprintf("data: %s\n\n", string(data))
			}
		case "input_json_delta":
			if event.Delta.PartialJSON != "" {
				idx := state.ToolCallIndex
				chunk := ChatCompletionChunk{
					ID: chunkID, Object: "chat.completion.chunk", Created: now, Model: model,
					Choices: []ChatChoice{{
						Index: 0,
						Delta: &ChatMessage{
							ToolCalls: []ToolCall{{
								Index:    &idx,
								Function: ToolCallFunction{Arguments: event.Delta.PartialJSON},
							}},
						},
					}},
				}
				data, _ := json.Marshal(chunk)
				return fmt.Sprintf("data: %s\n\n", string(data))
			}
		case "thinking_delta", "signature_delta":
			// 扩展思考增量无 OpenAI 对应物，不输出
		}

	case AnthropicEventContentBlockStop:
		if state.CurrentBlockType == "tool_use" {
			state.ToolCallIndex++
		}
		state.CurrentBlockType = ""

	case AnthropicEventMessageDelta:
		if event.Delta != nil && event.Delta.StopReason != nil {
			reason := mapAnthropicStopReasonToOpenAI(*event.Delta.StopReason)
			chunk := ChatCompletionChunk{
				ID: chunkID, Object: "chat.completion.chunk", Created: now, Model: model,
				Choices: []ChatChoice{{Index: 0, Delta: &ChatMessage{}, FinishReason: &reason}},
			}
			if event.Usage != nil {
				chunk.Usage = event.Usage.ToUsage()
			}
			data, _ := json.Marshal(chunk)
			return fmt.Sprintf("data: %s\n\n", string(data))
		}

	case AnthropicEventMessageStop:
		return "data: [DONE]\n\n"
	}

	return ""
}

// ──────────────────────────────────
// tool_choice 双向映射
// ──────────────────────────────────

// mapOpenAIToolChoiceToAnthropic 将 OpenAI tool_choice 映射为 Anthropic 格式
// OpenAI: "none" | "auto" | "required" | {"type":"function","function":{"name":"..."}}
// Anthropic: {"type":"auto"} | {"type":"any"} | {"type":"tool","name":"..."}
func mapOpenAIToolChoiceToAnthropic(choice interface{}) interface{} {
	switch v := choice.(type) {
	case string:
		switch v {
		case "auto":
			return AnthropicToolChoice{Type: "auto"}
		case "required":
			return AnthropicToolChoice{Type: "any"}
		case "none":
			return nil
		}
	case map[string]interface{}:
		if fn, ok := v["function"].(map[string]interface{}); ok {
			if name, ok := fn["name"].(string); ok {
				return AnthropicToolChoice{Type: "tool", Name: name}
			}
		}
	}
	return nil
}

// mapAnthropicToolChoiceToOpenAI 将 Anthropic tool_choice 映射为 OpenAI 格式
func mapAnthropicToolChoiceToOpenAI(choice interface{}) interface{} {
	switch v := choice.(type) {
	case map[string]interface{}:
		tcType, _ := v["type"].(string)
		switch tcType {
		case "auto":
			return "auto"
		case "any":
			return "required"
		case "tool":
			if name, ok := v["name"].(string); ok {
				return map[string]interface{}{
					"type":     "function",
					"function": map[string]interface{}{"name": name},
				}
			}
		}
	}
	return nil
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
