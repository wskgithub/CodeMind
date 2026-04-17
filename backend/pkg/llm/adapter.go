// Package llm provides LLM provider client wrappers and format adapters.
package llm

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Role constants.
const (
	RoleSystem    = "system"
	RoleUser      = "user"
	RoleAssistant = "assistant"
	RoleTool      = "tool"
)

// Content block type constants.
const (
	ContentTypeText     = "text"
	ContentTypeToolUse  = "tool_use"
	ContentTypeThinking = "thinking"
)

// Finish reason constants.
const (
	FinishReasonStop      = "stop"
	FinishReasonEndTurn   = "end_turn"
	FinishReasonToolCalls = "tool_calls"
)

// Tool choice strategy constants.
const (
	ToolChoiceAuto     = "auto"
	ToolChoiceRequired = "required"
)

// OpenAIToAnthropic converts an OpenAI ChatCompletion request to Anthropic Messages format.
func OpenAIToAnthropic(req *ChatCompletionRequest) *AnthropicMessagesRequest { //nolint:gocyclo // complex business logic.
	anthropicReq := &AnthropicMessagesRequest{
		Model:       req.Model,
		Stream:      req.Stream,
		Temperature: req.Temperature,
		TopP:        req.TopP,
	}

	switch {
	case req.MaxCompletionTokens != nil:
		anthropicReq.MaxTokens = *req.MaxCompletionTokens
	case req.MaxTokens != nil:
		anthropicReq.MaxTokens = *req.MaxTokens
	default:
		anthropicReq.MaxTokens = 4096
	}

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

	for _, tool := range req.Tools {
		if tool.Type == "function" {
			anthropicReq.Tools = append(anthropicReq.Tools, AnthropicTool{
				Name:        tool.Function.Name,
				Description: tool.Function.Description,
				InputSchema: tool.Function.Parameters,
			})
		}
	}

	if req.ToolChoice != nil && len(anthropicReq.Tools) > 0 {
		anthropicReq.ToolChoice = mapOpenAIToolChoiceToAnthropic(req.ToolChoice)
	}

	var systemParts []string
	var messages []AnthropicMessage

	for _, msg := range req.Messages {
		switch msg.Role {
		case RoleSystem, "developer":
			systemParts = append(systemParts, msg.ContentString())
		case RoleUser:
			messages = append(messages, AnthropicMessage{
				Role:    RoleUser,
				Content: msg.Content,
			})
		case RoleAssistant:
			anthropicMsg := AnthropicMessage{
				Role: RoleAssistant,
			}
			if len(msg.ToolCalls) > 0 {
				var blocks []AnthropicContentBlock
				text := msg.ContentString()
				if text != "" {
					blocks = append(blocks, AnthropicContentBlock{Type: ContentTypeText, Text: text})
				}
				for _, tc := range msg.ToolCalls {
					var input interface{}
					_ = json.Unmarshal([]byte(tc.Function.Arguments), &input)
					blocks = append(blocks, AnthropicContentBlock{
						Type:  ContentTypeToolUse,
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
		case RoleTool:
			messages = append(messages, AnthropicMessage{
				Role: RoleUser,
				Content: []AnthropicContentBlock{{
					Type:      "tool_result",
					ToolUseID: msg.ToolCallID,
					Content:   msg.Content,
				}},
			})
		default:
			messages = append(messages, AnthropicMessage{
				Role:    RoleUser,
				Content: fmt.Sprintf("[%s]: %s", msg.Role, msg.ContentString()),
			})
		}
	}

	if len(systemParts) > 0 {
		anthropicReq.System = strings.Join(systemParts, "\n\n")
	}

	if len(messages) == 0 {
		messages = append(messages, AnthropicMessage{
			Role:    RoleUser,
			Content: "Hello",
		})
	}

	if messages[0].Role != RoleUser {
		messages = append([]AnthropicMessage{{
			Role:    RoleUser,
			Content: "Continue",
		}}, messages...)
	}

	anthropicReq.Messages = messages
	return anthropicReq
}

// AnthropicToOpenAI converts an Anthropic Messages request to OpenAI ChatCompletion format.
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

	if req.ToolChoice != nil && len(openaiReq.Tools) > 0 {
		openaiReq.ToolChoice = mapAnthropicToolChoiceToOpenAI(req.ToolChoice)
	}

	var messages []ChatMessage

	if req.System != nil {
		switch v := req.System.(type) {
		case string:
			if v != "" {
				messages = append(messages, ChatMessage{Role: RoleSystem, Content: v})
			}
		case []interface{}:
			var systemText []string
			for _, block := range v {
				if blockMap, ok := block.(map[string]interface{}); ok {
					if text, ok := blockMap[ContentTypeText].(string); ok {
						systemText = append(systemText, text)
					}
				}
			}
			if len(systemText) > 0 {
				messages = append(messages, ChatMessage{Role: RoleSystem, Content: strings.Join(systemText, "\n\n")})
			}
		}
	}

	for _, msg := range req.Messages {
		openaiMsg := convertAnthropicMessageToOpenAI(msg)
		messages = append(messages, openaiMsg...)
	}

	openaiReq.Messages = messages
	return openaiReq
}

func convertAnthropicMessageToOpenAI(msg AnthropicMessage) []ChatMessage { //nolint:gocyclo // complex business logic.
	if textContent, ok := msg.Content.(string); ok {
		return []ChatMessage{{
			Role:    msg.Role,
			Content: textContent,
		}}
	}

	blocks, ok := msg.Content.([]interface{})
	if !ok {
		return []ChatMessage{{
			Role:    msg.Role,
			Content: extractAnthropicMessageContent(msg.Content),
		}}
	}

	if msg.Role == RoleAssistant {
		var textParts []string
		var toolCalls []ToolCall
		tcIdx := 0
		for _, block := range blocks {
			blockMap, ok := block.(map[string]interface{})
			if !ok {
				continue
			}
			switch blockMap["type"] {
			case ContentTypeText:
				if text, ok := blockMap[ContentTypeText].(string); ok {
					textParts = append(textParts, text)
				}
			case ContentTypeToolUse:
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
			case ContentTypeThinking:
			}
		}
		result := ChatMessage{
			Role:    RoleAssistant,
			Content: strings.Join(textParts, ""),
		}
		if len(toolCalls) > 0 {
			result.ToolCalls = toolCalls
		}
		return []ChatMessage{result}
	}

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
				Role: RoleTool,
			}
			if id, ok := blockMap["tool_use_id"].(string); ok {
				toolMsg.ToolCallID = id
			}
			if content, ok := blockMap["content"]; ok {
				toolMsg.Content = extractAnthropicMessageContent(content)
			}
			toolResults = append(toolResults, toolMsg)
		case ContentTypeText:
			if text, ok := blockMap[ContentTypeText].(string); ok {
				regularParts = append(regularParts, text)
			}
		case "image":
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
	if len(imageParts) > 0 {
		var parts []ContentPart
		for _, text := range regularParts {
			parts = append(parts, ContentPart{Type: ContentTypeText, Text: text})
		}
		parts = append(parts, imageParts...)
		result = append(result, ChatMessage{
			Role:    RoleUser,
			Content: parts,
		})
	} else if len(regularParts) > 0 {
		result = append(result, ChatMessage{
			Role:    RoleUser,
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

func extractAnthropicMessageContent(content interface{}) string {
	switch v := content.(type) {
	case string:
		return v
	case []interface{}:
		var parts []string
		for _, block := range v {
			if blockMap, ok := block.(map[string]interface{}); ok {
				if blockMap["type"] == ContentTypeText {
					if text, ok := blockMap[ContentTypeText].(string); ok {
						parts = append(parts, text)
					}
				}
			}
		}
		return strings.Join(parts, "")
	default:
		if data, err := json.Marshal(v); err == nil {
			return string(data)
		}
		return fmt.Sprintf("%v", v)
	}
}

// OpenAIResponseToAnthropic converts an OpenAI Chat response to Anthropic Messages format.
func OpenAIResponseToAnthropic(resp *ChatCompletionResponse) *AnthropicMessagesResponse {
	var content []AnthropicContentBlock

	for _, choice := range resp.Choices {
		if choice.Message == nil {
			continue
		}
		text := choice.Message.ContentString()
		if text != "" {
			content = append(content, AnthropicContentBlock{
				Type: ContentTypeText,
				Text: text,
			})
		}
		for _, tc := range choice.Message.ToolCalls {
			var input interface{}
			_ = json.Unmarshal([]byte(tc.Function.Arguments), &input)
			content = append(content, AnthropicContentBlock{
				Type:  ContentTypeToolUse,
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
		if resp.Usage.PromptTokensDetails != nil {
			usage.CacheCreationInputTokens = resp.Usage.PromptTokensDetails.CacheCreationInputTokens
			usage.CacheReadInputTokens = resp.Usage.PromptTokensDetails.CacheReadInputTokens
		}
	}

	return &AnthropicMessagesResponse{
		ID:         resp.ID,
		Type:       "message",
		Role:       RoleAssistant,
		Content:    content,
		Model:      resp.Model,
		StopReason: stopReason,
		Usage:      usage,
	}
}

// AnthropicResponseToOpenAI converts an Anthropic Messages response to OpenAI Chat format.
func AnthropicResponseToOpenAI(resp *AnthropicMessagesResponse) *ChatCompletionResponse {
	var textParts []string
	var toolCalls []ToolCall
	tcIndex := 0

	for _, block := range resp.Content {
		switch block.Type {
		case ContentTypeText:
			textParts = append(textParts, block.Text)
		case ContentTypeToolUse:
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
		case ContentTypeThinking:
		}
	}

	message := &ChatMessage{
		Role:    RoleAssistant,
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

// OpenAIToAnthropicState tracks content block index during OpenAI to Anthropic stream conversion.
type OpenAIToAnthropicState struct {
	ContentIndex int
	InToolCall   bool
}

// OpenAIChunkToAnthropicEvents converts an OpenAI stream chunk to Anthropic SSE format.
func OpenAIChunkToAnthropicEvents(chunk *ChatCompletionChunk, isFirst bool, state *OpenAIToAnthropicState) string { //nolint:gocyclo // complex business logic.
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

		if len(delta.ToolCalls) > 0 {
			for _, tc := range delta.ToolCalls {
				if tc.Function.Name != "" {
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

	if chunk.Choices[0].FinishReason != nil {
		stopReason := mapOpenAIStopReasonToAnthropic(*chunk.Choices[0].FinishReason)
		sb.WriteString(fmt.Sprintf(
			"event: content_block_stop\ndata: {\"type\":\"content_block_stop\",\"index\":%d}\n\n",
			state.ContentIndex,
		))
		usage := AnthropicUsage{}
		if chunk.Usage != nil {
			usage.InputTokens = chunk.Usage.PromptTokens
			usage.OutputTokens = chunk.Usage.CompletionTokens
			if chunk.Usage.PromptTokensDetails != nil {
				usage.CacheCreationInputTokens = chunk.Usage.PromptTokensDetails.CacheCreationInputTokens
				usage.CacheReadInputTokens = chunk.Usage.PromptTokensDetails.CacheReadInputTokens
			}
		}
		usageJSON, _ := json.Marshal(usage)
		sb.WriteString(fmt.Sprintf(
			"event: message_delta\ndata: {\"type\":\"message_delta\",\"delta\":{\"stop_reason\":\"%s\",\"stop_sequence\":null},\"usage\":%s}\n\n",
			stopReason, string(usageJSON),
		))
		sb.WriteString("event: message_stop\ndata: {\"type\":\"message_stop\"}\n\n")
	}

	return sb.String()
}

// AnthropicToOpenAIState tracks tool call state during Anthropic to OpenAI stream conversion.
type AnthropicToOpenAIState struct {
	CurrentBlockType string
	CurrentBlockID   string
	CurrentBlockName string
	ToolCallIndex    int
}

// AnthropicEventToOpenAIChunk converts an Anthropic stream event to OpenAI SSE format.
func AnthropicEventToOpenAIChunk(eventType string, event *AnthropicStreamEvent, model string, state *AnthropicToOpenAIState) string { //nolint:gocyclo // complex business logic.
	chunkID := "chatcmpl-" + uuid.New().String()[:8]
	now := time.Now().Unix()

	switch eventType {
	case AnthropicEventMessageStart:
		chunk := ChatCompletionChunk{
			ID: chunkID, Object: "chat.completion.chunk", Created: now, Model: model,
			Choices: []ChatChoice{{Index: 0, Delta: &ChatMessage{Role: RoleAssistant}}},
		}
		data, _ := json.Marshal(chunk)
		return fmt.Sprintf("data: %s\n\n", string(data))

	case AnthropicEventContentBlockStart:
		if event.ContentBlock != nil {
			state.CurrentBlockType = event.ContentBlock.Type
			switch event.ContentBlock.Type {
			case ContentTypeToolUse:
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
			case ContentTypeThinking:
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
		}

	case AnthropicEventContentBlockStop:
		if state.CurrentBlockType == ContentTypeToolUse {
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

// mapOpenAIToolChoiceToAnthropic maps OpenAI tool_choice to Anthropic format.
func mapOpenAIToolChoiceToAnthropic(choice interface{}) interface{} {
	switch v := choice.(type) {
	case string:
		switch v {
		case ToolChoiceAuto:
			return AnthropicToolChoice{Type: ToolChoiceAuto}
		case ToolChoiceRequired:
			return AnthropicToolChoice{Type: "any"}
		case "none":
			return nil
		}
	case map[string]interface{}:
		if fn, ok := v["function"].(map[string]interface{}); ok {
			if name, ok := fn["name"].(string); ok {
				return AnthropicToolChoice{Type: RoleTool, Name: name}
			}
		}
	}
	return nil
}

// mapAnthropicToolChoiceToOpenAI maps Anthropic tool_choice to OpenAI format.
func mapAnthropicToolChoiceToOpenAI(choice interface{}) interface{} {
	if v, ok := choice.(map[string]interface{}); ok {
		tcType, _ := v["type"].(string)
		switch tcType {
		case ToolChoiceAuto:
			return ToolChoiceAuto
		case "any":
			return ToolChoiceRequired
		case RoleTool:
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

// mapOpenAIStopReasonToAnthropic maps OpenAI stop reason to Anthropic format.
func mapOpenAIStopReasonToAnthropic(reason string) string {
	switch reason {
	case FinishReasonStop:
		return FinishReasonEndTurn
	case "length":
		return "max_tokens"
	case FinishReasonToolCalls, "function_call":
		return ContentTypeToolUse
	default:
		return FinishReasonEndTurn
	}
}

// mapAnthropicStopReasonToOpenAI maps Anthropic stop reason to OpenAI format.
func mapAnthropicStopReasonToOpenAI(reason string) string {
	switch reason {
	case FinishReasonEndTurn:
		return FinishReasonStop
	case "max_tokens":
		return "length"
	case ContentTypeToolUse:
		return FinishReasonToolCalls
	case "stop_sequence":
		return FinishReasonStop
	default:
		return FinishReasonStop
	}
}
