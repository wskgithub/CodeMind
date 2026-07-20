// Package llm provides LLM provider client wrappers and format adapters.
package llm

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

var errEmptyMessages = errors.New("messages cannot be empty")

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
func OpenAIToAnthropic(req *ChatCompletionRequest) (*AnthropicMessagesRequest, error) { //nolint:gocyclo // complex business logic.
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

	for i := 0; i < len(req.Messages); i++ {
		msg := req.Messages[i]
		switch msg.Role {
		case RoleSystem, "developer":
			systemParts = append(systemParts, msg.ContentString())
		case RoleUser:
			messages = append(messages, AnthropicMessage{
				Role:    RoleUser,
				Content: convertOpenAIUserContent(msg.Content),
			})
		case RoleTool:
			// 连续 tool 消息合并为一条 user 消息（Anthropic 要求 tool_result 在同一条 user 消息中）
			var toolBlocks []AnthropicContentBlock
			for i < len(req.Messages) && req.Messages[i].Role == RoleTool {
				toolMsg := req.Messages[i]
				toolBlocks = append(toolBlocks, AnthropicContentBlock{
					Type:      "tool_result",
					ToolUseID: toolMsg.ToolCallID,
					Content:   toolMsg.Content,
				})
				i++
			}
			i--
			messages = append(messages, AnthropicMessage{
				Role:    RoleUser,
				Content: toolBlocks,
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
		return nil, errEmptyMessages
	}

	anthropicReq.Messages = messages
	return anthropicReq, nil
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

	if req.Stream {
		includeUsage := true
		openaiReq.StreamOptions = &StreamOptions{IncludeUsage: &includeUsage}
	}

	if isOSeriesModel(req.Model) {
		if openaiReq.MaxTokens != nil {
			openaiReq.MaxCompletionTokens = openaiReq.MaxTokens
			openaiReq.MaxTokens = nil
		}
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

	// 部分 OpenAI 兼容上游（如 vLLM）严格要求 system 消息只能出现在 messages 首位，
	// 中间出现 role=system 会返回 400 "System message must be at the beginning."。
	// Claude Code 2.1.x 会在多轮工具调用中向 messages 中间插入 role=system 消息
	// （如 ToolSearch 提示），这里把所有 system 消息合并到首位，对宽松上游无损。
	messages = collapseSystemMessagesToHead(messages)

	openaiReq.Messages = messages
	return openaiReq
}

// collapseSystemMessagesToHead 把 messages 中所有 role=system 的消息合并为一条放在首位，
// 其余消息保持原相对顺序。content 仅取字符串形式（数组形式的多模态 system 消息保持原位）。
// 参考 cc-switch 的同名实现（transform_codex_chat.rs）。
func collapseSystemMessagesToHead(messages []ChatMessage) []ChatMessage {
	var systemChunks []string
	rest := make([]ChatMessage, 0, len(messages))

	for _, msg := range messages {
		if msg.Role == RoleSystem {
			if text, ok := msg.Content.(string); ok {
				trimmed := strings.TrimSpace(text)
				if trimmed != "" {
					systemChunks = append(systemChunks, text)
				}
				continue
			}
		}
		rest = append(rest, msg)
	}

	out := make([]ChatMessage, 0, len(rest)+1)
	if len(systemChunks) > 0 {
		out = append(out, ChatMessage{Role: RoleSystem, Content: strings.Join(systemChunks, "\n\n")})
	}
	out = append(out, rest...)
	return out
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
		usage = openAIUsageToAnthropic(resp.Usage)
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

// ToolBlockState 跟踪单个 tool_call 的流式转换状态。
type ToolBlockState struct {
	Started        bool
	AnthropicIndex int
	ID             string
	Name           string
	PendingArgs    string
	Open           bool
}

// OpenAIToAnthropicState 跟踪 OpenAI → Anthropic 流式转换状态。
type OpenAIToAnthropicState struct {
	MessageStarted      bool
	MessageID           string
	Model               string
	NextBlockIndex      int
	OpenTextBlock       bool
	TextBlockIndex      int
	ToolBlocks          map[int]*ToolBlockState
	PendingStopReason   *string
	PendingUsage        *AnthropicUsage
	MessageDeltaEmitted bool
}

// OpenAIChunkToAnthropicEvents 将 OpenAI 流式 chunk 转换为 Anthropic SSE 事件。
func OpenAIChunkToAnthropicEvents(chunk *ChatCompletionChunk, isFirst bool, state *OpenAIToAnthropicState) string { //nolint:gocyclo // complex business logic.
	var sb strings.Builder

	if state.ToolBlocks == nil {
		state.ToolBlocks = make(map[int]*ToolBlockState)
	}

	// usage chunk 通常 choices 为空，需先处理 usage
	if chunk.Usage != nil {
		state.PendingUsage = mergeAnthropicUsage(state.PendingUsage, openAIUsageToAnthropic(chunk.Usage))
	}

	if !state.MessageStarted && (isFirst || chunk.ID != "" || len(chunk.Choices) > 0) {
		if chunk.ID != "" {
			state.MessageID = chunk.ID
		}
		if chunk.Model != "" {
			state.Model = chunk.Model
		}
		writeMessageStart(&sb, state)
		state.MessageStarted = true
	}

	if len(chunk.Choices) == 0 {
		return sb.String()
	}

	choice := chunk.Choices[0]
	if choice.Delta != nil {
		text := choice.Delta.ContentString()
		if text != "" {
			if !state.OpenTextBlock {
				state.TextBlockIndex = state.NextBlockIndex
				state.NextBlockIndex++
				state.OpenTextBlock = true
				writeContentBlockStart(&sb, state.TextBlockIndex, ContentTypeText)
			}
			textJSON, _ := json.Marshal(text)
			sb.WriteString(fmt.Sprintf(
				"event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":%d,\"delta\":{\"type\":\"text_delta\",\"text\":%s}}\n\n",
				state.TextBlockIndex, string(textJSON),
			))
		}

		if len(choice.Delta.ToolCalls) > 0 {
			processOpenAIToolCallDeltas(&sb, state, choice.Delta.ToolCalls)
		}
	}

	if choice.FinishReason != nil {
		if state.PendingStopReason == nil {
			closeOpenTextBlock(&sb, state)
			lateStartPendingToolBlocks(&sb, state)
			closeOpenToolBlocks(&sb, state)
			reason := mapOpenAIStopReasonToAnthropic(*choice.FinishReason)
			state.PendingStopReason = &reason
		} else if chunk.Usage != nil {
			// 重复 finish chunk：仅更新 usage（已在开头合并）
		}
	}

	return sb.String()
}

// OpenAIStreamFinalize 在流结束时发出缓存的 message_delta 与 message_stop。
func OpenAIStreamFinalize(state *OpenAIToAnthropicState) string {
	if state.MessageDeltaEmitted {
		return ""
	}
	var sb strings.Builder

	if state.ToolBlocks == nil {
		state.ToolBlocks = make(map[int]*ToolBlockState)
	}

	closeOpenTextBlock(&sb, state)
	lateStartPendingToolBlocks(&sb, state)
	closeOpenToolBlocks(&sb, state)

	stopReason := FinishReasonEndTurn
	if state.PendingStopReason != nil {
		stopReason = *state.PendingStopReason
	}

	usage := state.PendingUsage
	if usage == nil {
		usage = &AnthropicUsage{}
	}
	usageJSON, _ := json.Marshal(usage)
	sb.WriteString(fmt.Sprintf(
		"event: message_delta\ndata: {\"type\":\"message_delta\",\"delta\":{\"stop_reason\":\"%s\",\"stop_sequence\":null},\"usage\":%s}\n\n",
		stopReason, string(usageJSON),
	))
	sb.WriteString("event: message_stop\ndata: {\"type\":\"message_stop\"}\n\n")
	state.MessageDeltaEmitted = true

	return sb.String()
}

func writeMessageStart(sb *strings.Builder, state *OpenAIToAnthropicState) {
	usage := state.PendingUsage
	if usage == nil {
		usage = &AnthropicUsage{}
	}
	usageJSON, _ := json.Marshal(usage)
	msgStart := fmt.Sprintf(
		`{"id":"%s","type":"message","role":"assistant","content":[],"model":"%s","stop_reason":null,"stop_sequence":null,"usage":%s}`,
		state.MessageID, state.Model, string(usageJSON),
	)
	sb.WriteString(fmt.Sprintf("event: message_start\ndata: {\"type\":\"message_start\",\"message\":%s}\n\n", msgStart))
	sb.WriteString("event: ping\ndata: {\"type\":\"ping\"}\n\n")
}

func writeContentBlockStart(sb *strings.Builder, index int, blockType string) {
	if blockType == ContentTypeText {
		sb.WriteString(fmt.Sprintf(
			"event: content_block_start\ndata: {\"type\":\"content_block_start\",\"index\":%d,\"content_block\":{\"type\":\"text\",\"text\":\"\"}}\n\n",
			index,
		))
	}
}

func writeToolBlockStart(sb *strings.Builder, index int, id, name string) {
	nameJSON, _ := json.Marshal(name)
	sb.WriteString(fmt.Sprintf(
		"event: content_block_start\ndata: {\"type\":\"content_block_start\",\"index\":%d,\"content_block\":{\"type\":\"tool_use\",\"id\":\"%s\",\"name\":%s,\"input\":{}}}\n\n",
		index, id, string(nameJSON),
	))
}

func closeOpenTextBlock(sb *strings.Builder, state *OpenAIToAnthropicState) {
	if !state.OpenTextBlock {
		return
	}
	sb.WriteString(fmt.Sprintf(
		"event: content_block_stop\ndata: {\"type\":\"content_block_stop\",\"index\":%d}\n\n",
		state.TextBlockIndex,
	))
	state.OpenTextBlock = false
}

func closeOpenToolBlocks(sb *strings.Builder, state *OpenAIToAnthropicState) {
	indices := make([]int, 0, len(state.ToolBlocks))
	for _, tb := range state.ToolBlocks {
		if tb.Open {
			indices = append(indices, tb.AnthropicIndex)
			tb.Open = false
		}
	}
	sortToolBlockIndices(indices)
	for _, idx := range indices {
		sb.WriteString(fmt.Sprintf(
			"event: content_block_stop\ndata: {\"type\":\"content_block_stop\",\"index\":%d}\n\n",
			idx,
		))
	}
}

func sortToolBlockIndices(indices []int) {
	for i := 0; i < len(indices); i++ {
		for j := i + 1; j < len(indices); j++ {
			if indices[j] < indices[i] {
				indices[i], indices[j] = indices[j], indices[i]
			}
		}
	}
}

func getToolCallIndex(tc ToolCall) int {
	if tc.Index != nil {
		return *tc.Index
	}
	return 0
}

func getOrCreateToolBlock(state *OpenAIToAnthropicState, toolIdx int) *ToolBlockState {
	if tb, ok := state.ToolBlocks[toolIdx]; ok {
		return tb
	}
	tb := &ToolBlockState{
		AnthropicIndex: state.NextBlockIndex,
	}
	state.NextBlockIndex++
	state.ToolBlocks[toolIdx] = tb
	return tb
}

func processOpenAIToolCallDeltas(sb *strings.Builder, state *OpenAIToAnthropicState, toolCalls []ToolCall) {
	closeOpenTextBlock(sb, state)

	for _, tc := range toolCalls {
		toolIdx := getToolCallIndex(tc)
		tb := getOrCreateToolBlock(state, toolIdx)

		if tc.ID != "" {
			tb.ID = tc.ID
		}
		if tc.Function.Name != "" {
			tb.Name = tc.Function.Name
		}

		if !tb.Started && tb.ID != "" && tb.Name != "" {
			startToolBlock(sb, state, tb)
		}

		if tc.Function.Arguments != "" {
			if tb.Started {
				argsJSON, _ := json.Marshal(tc.Function.Arguments)
				sb.WriteString(fmt.Sprintf(
					"event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":%d,\"delta\":{\"type\":\"input_json_delta\",\"partial_json\":%s}}\n\n",
					tb.AnthropicIndex, string(argsJSON),
				))
			} else {
				tb.PendingArgs += tc.Function.Arguments
			}
		}
	}
}

func startToolBlock(sb *strings.Builder, state *OpenAIToAnthropicState, tb *ToolBlockState) {
	tb.Started = true
	tb.Open = true
	writeToolBlockStart(sb, tb.AnthropicIndex, tb.ID, tb.Name)
	if tb.PendingArgs != "" {
		argsJSON, _ := json.Marshal(tb.PendingArgs)
		sb.WriteString(fmt.Sprintf(
			"event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":%d,\"delta\":{\"type\":\"input_json_delta\",\"partial_json\":%s}}\n\n",
			tb.AnthropicIndex, string(argsJSON),
		))
		tb.PendingArgs = ""
	}
	_ = state
}

func lateStartPendingToolBlocks(sb *strings.Builder, state *OpenAIToAnthropicState) {
	type lateStart struct {
		index   int
		id      string
		name    string
		pending string
	}
	var starts []lateStart
	for toolIdx, tb := range state.ToolBlocks {
		if tb.Started {
			continue
		}
		if tb.PendingArgs == "" && tb.ID == "" && tb.Name == "" {
			continue
		}
		id := tb.ID
		if id == "" {
			id = "toolu_" + uuid.New().String()[:8]
		}
		name := tb.Name
		if name == "" {
			name = "unknown_tool"
		}
		tb.Started = true
		tb.Open = true
		starts = append(starts, lateStart{
			index:   tb.AnthropicIndex,
			id:      id,
			name:    name,
			pending: tb.PendingArgs,
		})
		tb.ID = id
		tb.Name = name
		tb.PendingArgs = ""
		_ = toolIdx
	}
	sort.Slice(starts, func(i, j int) bool {
		return starts[i].index < starts[j].index
	})
	for _, s := range starts {
		writeToolBlockStart(sb, s.index, s.id, s.name)
		if s.pending != "" {
			argsJSON, _ := json.Marshal(s.pending)
			sb.WriteString(fmt.Sprintf(
				"event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":%d,\"delta\":{\"type\":\"input_json_delta\",\"partial_json\":%s}}\n\n",
				s.index, string(argsJSON),
			))
		}
	}
}

// openAIUsageToAnthropic 将 OpenAI usage 转为 Anthropic 格式（扣减缓存 token）。
func openAIUsageToAnthropic(u *Usage) *AnthropicUsage {
	if u == nil {
		return nil
	}
	cached := 0
	cacheCreation := 0
	if u.PromptTokensDetails != nil {
		cacheCreation = u.PromptTokensDetails.CacheCreationInputTokens
		cached = u.PromptTokensDetails.CacheReadInputTokens
		if cached == 0 {
			cached = u.PromptTokensDetails.CachedTokens
		}
	}
	inputTokens := u.PromptTokens - cached - cacheCreation
	if inputTokens < 0 {
		inputTokens = 0
	}
	usage := &AnthropicUsage{
		InputTokens:  inputTokens,
		OutputTokens: u.CompletionTokens,
	}
	if cached > 0 {
		usage.CacheReadInputTokens = cached
	}
	if cacheCreation > 0 {
		usage.CacheCreationInputTokens = cacheCreation
	}
	return usage
}

func mergeAnthropicUsage(dst, src *AnthropicUsage) *AnthropicUsage {
	if src == nil {
		return dst
	}
	if dst == nil {
		return src
	}
	if src.InputTokens > 0 {
		dst.InputTokens = src.InputTokens
	}
	if src.OutputTokens > 0 {
		dst.OutputTokens = src.OutputTokens
	}
	if src.CacheReadInputTokens > 0 {
		dst.CacheReadInputTokens = src.CacheReadInputTokens
	}
	if src.CacheCreationInputTokens > 0 {
		dst.CacheCreationInputTokens = src.CacheCreationInputTokens
	}
	return dst
}

// convertOpenAIUserContent 将 OpenAI user 消息 content 转为 Anthropic 格式。
func convertOpenAIUserContent(content interface{}) interface{} {
	if content == nil {
		return ""
	}
	switch v := content.(type) {
	case string:
		return v
	case []ContentPart:
		return convertContentPartsToAnthropic(v)
	case []interface{}:
		var parts []ContentPart
		for _, item := range v {
			if partMap, ok := item.(map[string]interface{}); ok {
				part := ContentPart{Type: fmt.Sprintf("%v", partMap["type"])}
				if text, ok := partMap["text"].(string); ok {
					part.Text = text
				}
				if imgURL, ok := partMap["image_url"].(map[string]interface{}); ok {
					if url, ok := imgURL["url"].(string); ok {
						part.ImageURL = &ImageURL{URL: url}
					}
				}
				parts = append(parts, part)
			}
		}
		if len(parts) == 0 {
			return content
		}
		return convertContentPartsToAnthropic(parts)
	default:
		return content
	}
}

func convertContentPartsToAnthropic(parts []ContentPart) []AnthropicContentBlock {
	var blocks []AnthropicContentBlock
	for _, part := range parts {
		switch part.Type {
		case ContentTypeText:
			if part.Text != "" {
				blocks = append(blocks, AnthropicContentBlock{Type: ContentTypeText, Text: part.Text})
			}
		case "image_url":
			if part.ImageURL != nil {
				if block := imageURLToAnthropicBlock(part.ImageURL.URL); block != nil {
					blocks = append(blocks, *block)
				}
			}
		}
	}
	if len(blocks) == 1 && blocks[0].Type == ContentTypeText {
		return blocks
	}
	if len(blocks) == 0 {
		return nil
	}
	return blocks
}

func imageURLToAnthropicBlock(url string) *AnthropicContentBlock {
	if strings.HasPrefix(url, "data:") {
		// data:<media_type>;base64,<data>
		rest := strings.TrimPrefix(url, "data:")
		mediaType, data, ok := strings.Cut(rest, ";base64,")
		if !ok || mediaType == "" || data == "" {
			return nil
		}
		return &AnthropicContentBlock{
			Type: "image",
			Source: &AnthropicImageSource{
				Type:      "base64",
				MediaType: mediaType,
				Data:      data,
			},
		}
	}
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		return &AnthropicContentBlock{
			Type: "image",
			Source: &AnthropicImageSource{
				Type: "url",
				URL:  url,
			},
		}
	}
	return nil
}

// isOSeriesModel 判断是否为 OpenAI o 系列模型。
func isOSeriesModel(model string) bool {
	if len(model) < 2 || model[0] != 'o' {
		return false
	}
	for _, c := range model[1:] {
		if c < '0' || c > '9' {
			if c == '-' {
				break
			}
			return false
		}
	}
	return model[1] >= '0' && model[1] <= '9'
}

// AnthropicToOpenAIState tracks tool call state during Anthropic to OpenAI stream conversion.
type AnthropicToOpenAIState struct {
	CurrentBlockType string
	CurrentBlockID   string
	CurrentBlockName string
	ToolCallIndex    int
	ChunkID          string
	Created          int64
}

// AnthropicEventToOpenAIChunk converts an Anthropic stream event to OpenAI SSE format.
func AnthropicEventToOpenAIChunk(eventType string, event *AnthropicStreamEvent, model string, state *AnthropicToOpenAIState) string { //nolint:gocyclo // complex business logic.
	if state.ChunkID == "" {
		state.ChunkID = "chatcmpl-" + uuid.New().String()[:8]
		state.Created = time.Now().Unix()
	}
	chunkID := state.ChunkID
	now := state.Created

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
		case "none":
			return "none"
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
