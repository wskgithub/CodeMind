package llm

import (
	"encoding/json"
	"strings"
	"testing"
)

// ══════════════════════════════════
// OpenAI → Anthropic 请求转换
// ══════════════════════════════════

// TestOpenAIToAnthropicBasic 测试基本请求转换
func TestOpenAIToAnthropicBasic(t *testing.T) {
	temp := 0.7
	maxTokens := 2048
	req := &ChatCompletionRequest{
		Model:       "gpt-4",
		Messages:    []ChatMessage{{Role: "user", Content: "Hello"}},
		Temperature: &temp,
		MaxTokens:   &maxTokens,
		Stream:      true,
	}

	result := OpenAIToAnthropic(req)

	if result.Model != "gpt-4" {
		t.Errorf("Model 不正确: %s", result.Model)
	}
	if result.MaxTokens != 2048 {
		t.Errorf("MaxTokens 不正确: %d", result.MaxTokens)
	}
	if *result.Temperature != 0.7 {
		t.Errorf("Temperature 不正确: %f", *result.Temperature)
	}
	if !result.Stream {
		t.Error("Stream 应为 true")
	}
	if len(result.Messages) != 1 || result.Messages[0].Role != "user" {
		t.Error("Messages 不正确")
	}
}

// TestOpenAIToAnthropicSystemMessage 测试 system 消息提取
func TestOpenAIToAnthropicSystemMessage(t *testing.T) {
	req := &ChatCompletionRequest{
		Model: "gpt-4",
		Messages: []ChatMessage{
			{Role: "system", Content: "You are helpful."},
			{Role: "user", Content: "Hello"},
		},
		MaxTokens: intPtr(1024),
	}

	result := OpenAIToAnthropic(req)

	// system 消息应被提取到 System 字段
	if result.System == nil || result.System != "You are helpful." {
		t.Errorf("System 不正确: %v", result.System)
	}
	// Messages 中不应包含 system 角色
	for _, msg := range result.Messages {
		if msg.Role == "system" {
			t.Error("Anthropic Messages 不应包含 system 角色")
		}
	}
}

// TestOpenAIToAnthropicMaxTokensDefault 测试 max_tokens 默认值
func TestOpenAIToAnthropicMaxTokensDefault(t *testing.T) {
	req := &ChatCompletionRequest{
		Model:    "gpt-4",
		Messages: []ChatMessage{{Role: "user", Content: "Hi"}},
	}

	result := OpenAIToAnthropic(req)
	if result.MaxTokens != 4096 {
		t.Errorf("默认 MaxTokens 应为 4096, 实际: %d", result.MaxTokens)
	}
}

// TestOpenAIToAnthropicMaxCompletionTokensPriority 测试 max_completion_tokens 优先级
func TestOpenAIToAnthropicMaxCompletionTokensPriority(t *testing.T) {
	maxTokens := 1024
	maxCompletionTokens := 2048
	req := &ChatCompletionRequest{
		Model:               "gpt-4",
		Messages:            []ChatMessage{{Role: "user", Content: "Hi"}},
		MaxTokens:           &maxTokens,
		MaxCompletionTokens: &maxCompletionTokens,
	}

	result := OpenAIToAnthropic(req)
	if result.MaxTokens != 2048 {
		t.Errorf("应优先使用 max_completion_tokens: %d", result.MaxTokens)
	}
}

// TestOpenAIToAnthropicToolConversion 测试工具定义转换
func TestOpenAIToAnthropicToolConversion(t *testing.T) {
	req := &ChatCompletionRequest{
		Model:    "gpt-4",
		Messages: []ChatMessage{{Role: "user", Content: "天气"}},
		Tools: []Tool{{
			Type: "function",
			Function: ToolFunction{
				Name:        "get_weather",
				Description: "获取天气",
				Parameters:  map[string]interface{}{"type": "object"},
			},
		}},
	}

	result := OpenAIToAnthropic(req)

	if len(result.Tools) != 1 {
		t.Fatalf("Tools 数量不正确: %d", len(result.Tools))
	}
	tool := result.Tools[0]
	if tool.Name != "get_weather" || tool.Description != "获取天气" {
		t.Errorf("工具定义不正确: name=%s, desc=%s", tool.Name, tool.Description)
	}
}

// TestOpenAIToAnthropicToolChoice 测试 tool_choice 转换
func TestOpenAIToAnthropicToolChoice(t *testing.T) {
	tests := []struct {
		name     string
		choice   interface{}
		expected string // AnthropicToolChoice.Type
	}{
		{"auto → auto", "auto", "auto"},
		{"required → any", "required", "any"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &ChatCompletionRequest{
				Model:      "gpt-4",
				Messages:   []ChatMessage{{Role: "user", Content: "Hi"}},
				Tools:      []Tool{{Type: "function", Function: ToolFunction{Name: "test"}}},
				ToolChoice: tt.choice,
			}

			result := OpenAIToAnthropic(req)
			if result.ToolChoice == nil {
				t.Fatal("ToolChoice 不应为 nil")
			}
			tc, ok := result.ToolChoice.(AnthropicToolChoice)
			if !ok {
				t.Fatalf("ToolChoice 类型不正确: %T", result.ToolChoice)
			}
			if tc.Type != tt.expected {
				t.Errorf("ToolChoice.Type 不正确: 预期 %s, 实际 %s", tt.expected, tc.Type)
			}
		})
	}
}

// TestOpenAIToAnthropicToolCallConversion 测试 tool_calls → tool_use 转换
func TestOpenAIToAnthropicToolCallConversion(t *testing.T) {
	req := &ChatCompletionRequest{
		Model: "gpt-4",
		Messages: []ChatMessage{
			{Role: "user", Content: "天气"},
			{
				Role: "assistant",
				ToolCalls: []ToolCall{{
					ID:       "call_1",
					Type:     "function",
					Function: ToolCallFunction{Name: "get_weather", Arguments: `{"city":"Beijing"}`},
				}},
			},
			{Role: "tool", ToolCallID: "call_1", Content: "晴天 25°C"},
		},
	}

	result := OpenAIToAnthropic(req)

	// assistant 消息应包含 tool_use 内容块
	assistantMsg := result.Messages[1]
	blocks, ok := assistantMsg.Content.([]AnthropicContentBlock)
	if !ok {
		t.Fatal("assistant content 应为 []AnthropicContentBlock")
	}
	if len(blocks) == 0 || blocks[0].Type != "tool_use" {
		t.Error("应包含 tool_use 内容块")
	}

	// tool 消息应转为 user 角色 + tool_result 内容块
	toolResultMsg := result.Messages[2]
	if toolResultMsg.Role != "user" {
		t.Errorf("tool 消息应转为 user 角色: %s", toolResultMsg.Role)
	}
}

// TestOpenAIToAnthropicStopSequence 测试 stop 序列转换
func TestOpenAIToAnthropicStopSequence(t *testing.T) {
	// string 类型
	req := &ChatCompletionRequest{
		Model:    "gpt-4",
		Messages: []ChatMessage{{Role: "user", Content: "Hi"}},
		Stop:     "END",
	}
	result := OpenAIToAnthropic(req)
	if len(result.StopSequences) != 1 || result.StopSequences[0] != "END" {
		t.Errorf("StopSequences 不正确: %v", result.StopSequences)
	}
}

// TestOpenAIToAnthropicFirstMessageMustBeUser 测试第一条消息必须是 user 角色
func TestOpenAIToAnthropicFirstMessageMustBeUser(t *testing.T) {
	req := &ChatCompletionRequest{
		Model: "gpt-4",
		Messages: []ChatMessage{
			{Role: "assistant", Content: "I'm ready."},
		},
	}

	result := OpenAIToAnthropic(req)
	if result.Messages[0].Role != "user" {
		t.Error("第一条 Anthropic 消息必须是 user 角色")
	}
}

// ══════════════════════════════════
// Anthropic → OpenAI 请求转换
// ══════════════════════════════════

// TestAnthropicToOpenAIBasic 测试基本请求转换
func TestAnthropicToOpenAIBasic(t *testing.T) {
	temp := 0.5
	req := &AnthropicMessagesRequest{
		Model:       "claude-sonnet-4-20250514",
		Messages:    []AnthropicMessage{{Role: "user", Content: "Hello"}},
		System:      "You are helpful.",
		MaxTokens:   2048,
		Temperature: &temp,
		Stream:      true,
	}

	result := AnthropicToOpenAI(req)

	if result.Model != "claude-sonnet-4-20250514" {
		t.Errorf("Model 不正确: %s", result.Model)
	}
	if result.MaxTokens == nil || *result.MaxTokens != 2048 {
		t.Error("MaxTokens 不正确")
	}
	if !result.Stream {
		t.Error("Stream 应为 true")
	}

	// system 应转为 system 消息
	if len(result.Messages) < 2 {
		t.Fatal("应至少有 2 条消息（system + user）")
	}
	if result.Messages[0].Role != "system" {
		t.Error("第一条消息应为 system 角色")
	}
	if result.Messages[0].ContentString() != "You are helpful." {
		t.Error("system 内容不正确")
	}
}

// TestAnthropicToOpenAIToolChoice 测试 tool_choice 转换
func TestAnthropicToOpenAIToolChoice(t *testing.T) {
	tests := []struct {
		name     string
		choice   interface{}
		expected interface{}
	}{
		{
			"auto → auto",
			map[string]interface{}{"type": "auto"},
			"auto",
		},
		{
			"any → required",
			map[string]interface{}{"type": "any"},
			"required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &AnthropicMessagesRequest{
				Model:      "claude-sonnet-4-20250514",
				Messages:   []AnthropicMessage{{Role: "user", Content: "Hi"}},
				MaxTokens:  1024,
				Tools:      []AnthropicTool{{Name: "test", InputSchema: map[string]interface{}{}}},
				ToolChoice: tt.choice,
			}

			result := AnthropicToOpenAI(req)
			if result.ToolChoice != tt.expected {
				t.Errorf("ToolChoice 不正确: 预期 %v, 实际 %v", tt.expected, result.ToolChoice)
			}
		})
	}
}

// TestAnthropicToOpenAIToolResultConversion 测试 tool_result → tool 消息转换
func TestAnthropicToOpenAIToolResultConversion(t *testing.T) {
	reqJSON := `{
		"model": "claude-sonnet-4-20250514",
		"max_tokens": 1024,
		"messages": [
			{"role": "user", "content": "天气怎样"},
			{"role": "assistant", "content": [
				{"type": "text", "text": "我来查查"},
				{"type": "tool_use", "id": "toolu_1", "name": "get_weather", "input": {"city": "Beijing"}}
			]},
			{"role": "user", "content": [
				{"type": "tool_result", "tool_use_id": "toolu_1", "content": "晴天 25°C"}
			]}
		]
	}`

	var req AnthropicMessagesRequest
	json.Unmarshal([]byte(reqJSON), &req)

	result := AnthropicToOpenAI(&req)

	// 查找 tool 角色消息
	var toolMsg *ChatMessage
	for i := range result.Messages {
		if result.Messages[i].Role == "tool" {
			toolMsg = &result.Messages[i]
			break
		}
	}
	if toolMsg == nil {
		t.Fatal("应生成 tool 角色消息")
	}
	if toolMsg.ToolCallID != "toolu_1" {
		t.Errorf("ToolCallID 不正确: %s", toolMsg.ToolCallID)
	}

	// 查找 assistant 消息的 tool_calls
	var assistantMsg *ChatMessage
	for i := range result.Messages {
		if result.Messages[i].Role == "assistant" {
			assistantMsg = &result.Messages[i]
			break
		}
	}
	if assistantMsg == nil {
		t.Fatal("应有 assistant 消息")
	}
	if len(assistantMsg.ToolCalls) != 1 {
		t.Fatalf("应有 1 个 tool_call, 实际: %d", len(assistantMsg.ToolCalls))
	}
	if assistantMsg.ToolCalls[0].Function.Name != "get_weather" {
		t.Errorf("tool_call 名称不正确: %s", assistantMsg.ToolCalls[0].Function.Name)
	}
}

// ══════════════════════════════════
// 响应转换
// ══════════════════════════════════

// TestOpenAIResponseToAnthropic 测试 OpenAI 响应→ Anthropic 格式
func TestOpenAIResponseToAnthropic(t *testing.T) {
	finishReason := "tool_calls"
	resp := &ChatCompletionResponse{
		ID:    "chatcmpl-123",
		Model: "gpt-4",
		Choices: []ChatChoice{{
			Index: 0,
			Message: &ChatMessage{
				Role:    "assistant",
				Content: "我来查天气",
				ToolCalls: []ToolCall{{
					ID:       "call_1",
					Type:     "function",
					Function: ToolCallFunction{Name: "get_weather", Arguments: `{"city":"Beijing"}`},
				}},
			},
			FinishReason: &finishReason,
		}},
		Usage: &Usage{PromptTokens: 50, CompletionTokens: 30, TotalTokens: 80},
	}

	result := OpenAIResponseToAnthropic(resp)

	if result.ID != "chatcmpl-123" {
		t.Errorf("ID 不正确: %s", result.ID)
	}
	if result.Type != "message" {
		t.Errorf("Type 不正确: %s", result.Type)
	}
	if result.Role != "assistant" {
		t.Errorf("Role 不正确: %s", result.Role)
	}

	// 应有 text + tool_use 两个内容块
	if len(result.Content) != 2 {
		t.Fatalf("Content 数量不正确: %d", len(result.Content))
	}
	if result.Content[0].Type != "text" || result.Content[0].Text != "我来查天气" {
		t.Error("text 内容块不正确")
	}
	if result.Content[1].Type != "tool_use" || result.Content[1].Name != "get_weather" {
		t.Error("tool_use 内容块不正确")
	}

	// stop_reason 应映射
	if result.StopReason == nil || *result.StopReason != "tool_use" {
		t.Error("StopReason 映射不正确")
	}

	// usage 应转换
	if result.Usage == nil || result.Usage.InputTokens != 50 || result.Usage.OutputTokens != 30 {
		t.Error("Usage 转换不正确")
	}
}

// TestAnthropicResponseToOpenAI 测试 Anthropic 响应 → OpenAI 格式
func TestAnthropicResponseToOpenAI(t *testing.T) {
	stopReason := "end_turn"
	resp := &AnthropicMessagesResponse{
		ID:    "msg_abc",
		Type:  "message",
		Role:  "assistant",
		Model: "claude-sonnet-4-20250514",
		Content: []AnthropicContentBlock{
			{Type: "thinking", Thinking: "Let me think..."},
			{Type: "text", Text: "The answer is 42."},
			{Type: "tool_use", ID: "toolu_1", Name: "calc", Input: map[string]interface{}{"x": 1}},
		},
		StopReason: &stopReason,
		Usage:      &AnthropicUsage{InputTokens: 100, OutputTokens: 50},
	}

	result := AnthropicResponseToOpenAI(resp)

	if result.ID != "msg_abc" {
		t.Errorf("ID 不正确: %s", result.ID)
	}
	if result.Object != "chat.completion" {
		t.Errorf("Object 不正确: %s", result.Object)
	}

	if len(result.Choices) != 1 {
		t.Fatalf("Choices 数量不正确: %d", len(result.Choices))
	}
	choice := result.Choices[0]

	// thinking 块应被跳过，只有 text 内容
	if choice.Message.ContentString() != "The answer is 42." {
		t.Errorf("Content 不正确（thinking 应被跳过）: %s", choice.Message.ContentString())
	}

	// tool_use → tool_calls
	if len(choice.Message.ToolCalls) != 1 {
		t.Fatalf("ToolCalls 数量不正确: %d", len(choice.Message.ToolCalls))
	}
	if choice.Message.ToolCalls[0].Function.Name != "calc" {
		t.Error("ToolCall 名称不正确")
	}

	// stop_reason → finish_reason
	if choice.FinishReason == nil || *choice.FinishReason != "stop" {
		t.Error("FinishReason 映射不正确")
	}
}

// ══════════════════════════════════
// 停止原因映射
// ══════════════════════════════════

// TestStopReasonMapping 测试双向停止原因映射
func TestStopReasonMapping(t *testing.T) {
	// OpenAI → Anthropic
	o2aTests := []struct{ input, expected string }{
		{"stop", "end_turn"},
		{"length", "max_tokens"},
		{"tool_calls", "tool_use"},
		{"function_call", "tool_use"},
		{"unknown", "end_turn"},
	}
	for _, tt := range o2aTests {
		result := mapOpenAIStopReasonToAnthropic(tt.input)
		if result != tt.expected {
			t.Errorf("OpenAI→Anthropic: %s → %s (预期 %s)", tt.input, result, tt.expected)
		}
	}

	// Anthropic → OpenAI
	a2oTests := []struct{ input, expected string }{
		{"end_turn", "stop"},
		{"max_tokens", "length"},
		{"tool_use", "tool_calls"},
		{"stop_sequence", "stop"},
		{"unknown", "stop"},
	}
	for _, tt := range a2oTests {
		result := mapAnthropicStopReasonToOpenAI(tt.input)
		if result != tt.expected {
			t.Errorf("Anthropic→OpenAI: %s → %s (预期 %s)", tt.input, result, tt.expected)
		}
	}
}

// ══════════════════════════════════
// tool_choice 映射
// ══════════════════════════════════

// TestToolChoiceMapping 测试双向 tool_choice 映射
func TestToolChoiceMapping(t *testing.T) {
	// OpenAI → Anthropic
	t.Run("OpenAI auto → Anthropic auto", func(t *testing.T) {
		result := mapOpenAIToolChoiceToAnthropic("auto")
		tc, ok := result.(AnthropicToolChoice)
		if !ok || tc.Type != "auto" {
			t.Errorf("预期 {type:auto}, 实际: %v", result)
		}
	})

	t.Run("OpenAI required → Anthropic any", func(t *testing.T) {
		result := mapOpenAIToolChoiceToAnthropic("required")
		tc, ok := result.(AnthropicToolChoice)
		if !ok || tc.Type != "any" {
			t.Errorf("预期 {type:any}, 实际: %v", result)
		}
	})

	t.Run("OpenAI none → nil", func(t *testing.T) {
		result := mapOpenAIToolChoiceToAnthropic("none")
		if result != nil {
			t.Errorf("预期 nil, 实际: %v", result)
		}
	})

	t.Run("OpenAI 指定函数 → Anthropic tool", func(t *testing.T) {
		input := map[string]interface{}{
			"type":     "function",
			"function": map[string]interface{}{"name": "get_weather"},
		}
		result := mapOpenAIToolChoiceToAnthropic(input)
		tc, ok := result.(AnthropicToolChoice)
		if !ok || tc.Type != "tool" || tc.Name != "get_weather" {
			t.Errorf("预期 {type:tool, name:get_weather}, 实际: %v", result)
		}
	})

	// Anthropic → OpenAI
	t.Run("Anthropic auto → OpenAI auto", func(t *testing.T) {
		result := mapAnthropicToolChoiceToOpenAI(map[string]interface{}{"type": "auto"})
		if result != "auto" {
			t.Errorf("预期 auto, 实际: %v", result)
		}
	})

	t.Run("Anthropic any → OpenAI required", func(t *testing.T) {
		result := mapAnthropicToolChoiceToOpenAI(map[string]interface{}{"type": "any"})
		if result != "required" {
			t.Errorf("预期 required, 实际: %v", result)
		}
	})

	t.Run("Anthropic tool → OpenAI 指定函数", func(t *testing.T) {
		result := mapAnthropicToolChoiceToOpenAI(map[string]interface{}{
			"type": "tool",
			"name": "get_weather",
		})
		m, ok := result.(map[string]interface{})
		if !ok {
			t.Fatalf("预期 map, 实际: %T", result)
		}
		fn, _ := m["function"].(map[string]interface{})
		if fn["name"] != "get_weather" {
			t.Errorf("函数名不正确: %v", fn["name"])
		}
	})
}

// ══════════════════════════════════
// 流式格式转换：Anthropic → OpenAI
// ══════════════════════════════════

// TestAnthropicEventToOpenAIChunkText 测试文本增量转换
func TestAnthropicEventToOpenAIChunkText(t *testing.T) {
	state := &AnthropicToOpenAIState{}

	// message_start → role chunk
	event := &AnthropicStreamEvent{Type: AnthropicEventMessageStart}
	result := AnthropicEventToOpenAIChunk(AnthropicEventMessageStart, event, "claude-sonnet-4-20250514", state)
	if !strings.Contains(result, `"role":"assistant"`) {
		t.Error("message_start 应生成包含 role:assistant 的 chunk")
	}

	// text_delta
	event = &AnthropicStreamEvent{
		Type:  AnthropicEventContentBlockDelta,
		Delta: &AnthropicStreamDelta{Type: "text_delta", Text: "Hello"},
	}
	result = AnthropicEventToOpenAIChunk(AnthropicEventContentBlockDelta, event, "claude-sonnet-4-20250514", state)
	if !strings.Contains(result, "Hello") {
		t.Error("text_delta 应包含文本内容")
	}
	if !strings.Contains(result, "chat.completion.chunk") {
		t.Error("应为 chat.completion.chunk 格式")
	}

	// message_stop → [DONE]
	event = &AnthropicStreamEvent{Type: AnthropicEventMessageStop}
	result = AnthropicEventToOpenAIChunk(AnthropicEventMessageStop, event, "claude-sonnet-4-20250514", state)
	if result != "data: [DONE]\n\n" {
		t.Errorf("message_stop 应生成 [DONE], 实际: %s", result)
	}
}

// TestAnthropicEventToOpenAIChunkToolUse 测试工具调用流式转换
func TestAnthropicEventToOpenAIChunkToolUse(t *testing.T) {
	state := &AnthropicToOpenAIState{}

	// content_block_start (tool_use)
	event := &AnthropicStreamEvent{
		Type:         AnthropicEventContentBlockStart,
		ContentBlock: &AnthropicContentBlock{Type: "tool_use", ID: "toolu_01", Name: "get_weather"},
	}
	result := AnthropicEventToOpenAIChunk(AnthropicEventContentBlockStart, event, "claude-sonnet-4-20250514", state)
	if !strings.Contains(result, "get_weather") {
		t.Error("tool_use start 应包含工具名")
	}
	if !strings.Contains(result, "toolu_01") {
		t.Error("tool_use start 应包含工具 ID")
	}
	if !strings.Contains(result, "tool_calls") {
		t.Error("应包含 tool_calls 字段")
	}

	// input_json_delta
	event = &AnthropicStreamEvent{
		Type:  AnthropicEventContentBlockDelta,
		Delta: &AnthropicStreamDelta{Type: "input_json_delta", PartialJSON: `{"city":`},
	}
	result = AnthropicEventToOpenAIChunk(AnthropicEventContentBlockDelta, event, "claude-sonnet-4-20250514", state)
	if !strings.Contains(result, "tool_calls") {
		t.Error("input_json_delta 应转为 tool_calls arguments")
	}

	// content_block_stop → 索引递增
	event = &AnthropicStreamEvent{Type: AnthropicEventContentBlockStop}
	AnthropicEventToOpenAIChunk(AnthropicEventContentBlockStop, event, "claude-sonnet-4-20250514", state)
	if state.ToolCallIndex != 1 {
		t.Errorf("tool_use 结束后索引应递增: %d", state.ToolCallIndex)
	}
}

// TestAnthropicEventToOpenAIChunkThinkingIgnored 测试 thinking 事件被忽略
func TestAnthropicEventToOpenAIChunkThinkingIgnored(t *testing.T) {
	state := &AnthropicToOpenAIState{}

	// content_block_start (thinking) → 不输出
	event := &AnthropicStreamEvent{
		Type:         AnthropicEventContentBlockStart,
		ContentBlock: &AnthropicContentBlock{Type: "thinking"},
	}
	result := AnthropicEventToOpenAIChunk(AnthropicEventContentBlockStart, event, "claude-sonnet-4-20250514", state)
	if result != "" {
		t.Errorf("thinking block_start 不应输出任何内容: %s", result)
	}

	// thinking_delta → 不输出
	event = &AnthropicStreamEvent{
		Type:  AnthropicEventContentBlockDelta,
		Delta: &AnthropicStreamDelta{Type: "thinking_delta", Thinking: "thinking..."},
	}
	result = AnthropicEventToOpenAIChunk(AnthropicEventContentBlockDelta, event, "claude-sonnet-4-20250514", state)
	if result != "" {
		t.Errorf("thinking_delta 不应输出任何内容: %s", result)
	}

	// signature_delta → 不输出
	event = &AnthropicStreamEvent{
		Type:  AnthropicEventContentBlockDelta,
		Delta: &AnthropicStreamDelta{Type: "signature_delta", Signature: "sig..."},
	}
	result = AnthropicEventToOpenAIChunk(AnthropicEventContentBlockDelta, event, "claude-sonnet-4-20250514", state)
	if result != "" {
		t.Errorf("signature_delta 不应输出任何内容: %s", result)
	}
}

// TestAnthropicEventToOpenAIChunkMessageDelta 测试 message_delta（含 usage）转换
func TestAnthropicEventToOpenAIChunkMessageDelta(t *testing.T) {
	state := &AnthropicToOpenAIState{}
	stopReason := "end_turn"

	event := &AnthropicStreamEvent{
		Type:  AnthropicEventMessageDelta,
		Delta: &AnthropicStreamDelta{StopReason: &stopReason},
		Usage: &AnthropicUsage{OutputTokens: 50},
	}

	result := AnthropicEventToOpenAIChunk(AnthropicEventMessageDelta, event, "claude-sonnet-4-20250514", state)
	if !strings.Contains(result, `"finish_reason":"stop"`) {
		t.Error("应包含映射后的 finish_reason: stop")
	}

	// 验证 usage 附加
	var chunk ChatCompletionChunk
	// 提取 data: 后面的 JSON
	jsonStr := strings.TrimPrefix(strings.TrimSpace(result), "data: ")
	jsonStr = strings.TrimSuffix(jsonStr, "\n\n")
	if err := json.Unmarshal([]byte(jsonStr), &chunk); err != nil {
		t.Fatalf("无法解析 chunk JSON: %v", err)
	}
	if chunk.Usage == nil {
		t.Fatal("chunk 应包含 usage")
	}
	if chunk.Usage.CompletionTokens != 50 {
		t.Errorf("CompletionTokens 不正确: %d", chunk.Usage.CompletionTokens)
	}
}

// ══════════════════════════════════
// 流式格式转换：OpenAI → Anthropic
// ══════════════════════════════════

// TestOpenAIChunkToAnthropicEventsFirstChunk 测试首个 chunk 生成 message_start
func TestOpenAIChunkToAnthropicEventsFirstChunk(t *testing.T) {
	state := &OpenAIToAnthropicState{}
	chunk := &ChatCompletionChunk{
		ID:    "chatcmpl-1",
		Model: "gpt-4",
		Choices: []ChatChoice{{
			Index: 0,
			Delta: &ChatMessage{Content: "Hi"},
		}},
	}

	result := OpenAIChunkToAnthropicEvents(chunk, true, state)

	if !strings.Contains(result, "event: message_start") {
		t.Error("首个 chunk 应包含 message_start 事件")
	}
	if !strings.Contains(result, "event: ping") {
		t.Error("首个 chunk 应包含 ping 事件")
	}
	if !strings.Contains(result, "text_delta") {
		t.Error("应包含 text_delta 事件")
	}
}

// TestOpenAIChunkToAnthropicEventsFinish 测试结束事件
func TestOpenAIChunkToAnthropicEventsFinish(t *testing.T) {
	state := &OpenAIToAnthropicState{}
	finishReason := "stop"
	chunk := &ChatCompletionChunk{
		ID:    "chatcmpl-1",
		Model: "gpt-4",
		Choices: []ChatChoice{{
			Index:        0,
			Delta:        &ChatMessage{},
			FinishReason: &finishReason,
		}},
		Usage: &Usage{CompletionTokens: 20},
	}

	result := OpenAIChunkToAnthropicEvents(chunk, false, state)

	if !strings.Contains(result, "content_block_stop") {
		t.Error("应包含 content_block_stop 事件")
	}
	if !strings.Contains(result, "message_delta") {
		t.Error("应包含 message_delta 事件")
	}
	if !strings.Contains(result, `"stop_reason":"end_turn"`) {
		t.Error("stop_reason 应映射为 end_turn")
	}
	if !strings.Contains(result, "message_stop") {
		t.Error("应包含 message_stop 事件")
	}
}

// TestOpenAIChunkToAnthropicEventsToolCall 测试工具调用 chunk 转换
func TestOpenAIChunkToAnthropicEventsToolCall(t *testing.T) {
	state := &OpenAIToAnthropicState{}
	idx := 0
	chunk := &ChatCompletionChunk{
		ID:    "chatcmpl-tc",
		Model: "gpt-4",
		Choices: []ChatChoice{{
			Index: 0,
			Delta: &ChatMessage{
				ToolCalls: []ToolCall{{
					Index:    &idx,
					ID:       "call_1",
					Type:     "function",
					Function: ToolCallFunction{Name: "get_weather", Arguments: ""},
				}},
			},
		}},
	}

	result := OpenAIChunkToAnthropicEvents(chunk, true, state)

	if !strings.Contains(result, "tool_use") {
		t.Error("应包含 tool_use content_block_start")
	}
	if !strings.Contains(result, "get_weather") {
		t.Error("应包含工具名称")
	}

	// 参数增量
	chunk2 := &ChatCompletionChunk{
		ID:    "chatcmpl-tc",
		Model: "gpt-4",
		Choices: []ChatChoice{{
			Index: 0,
			Delta: &ChatMessage{
				ToolCalls: []ToolCall{{
					Index:    &idx,
					Function: ToolCallFunction{Arguments: `{"city":"Beijing"}`},
				}},
			},
		}},
	}

	result2 := OpenAIChunkToAnthropicEvents(chunk2, false, state)
	if !strings.Contains(result2, "input_json_delta") {
		t.Error("应包含 input_json_delta 事件")
	}
}

// ══════════════════════════════════
// 辅助函数
// ══════════════════════════════════

func intPtr(i int) *int { return &i }
