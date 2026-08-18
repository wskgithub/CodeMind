package llm

import (
	"encoding/json"
	"strings"
	"testing"
)

// ══════════════════════════════════
// OpenAI to Anthropic request conversion
// ══════════════════════════════════

// TestOpenAIToAnthropicBasic tests basic request conversion.
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

	result, err := OpenAIToAnthropic(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Model != "gpt-4" {
		t.Errorf("incorrect Model: %s", result.Model)
	}
	if result.MaxTokens != 2048 {
		t.Errorf("incorrect MaxTokens: %d", result.MaxTokens)
	}
	if *result.Temperature != 0.7 {
		t.Errorf("incorrect Temperature: %f", *result.Temperature)
	}
	if !result.Stream {
		t.Error("Stream should be true")
	}
	if len(result.Messages) != 1 || result.Messages[0].Role != "user" {
		t.Error("incorrect Messages")
	}
}

// TestOpenAIToAnthropicSystemMessage tests system message extraction.
func TestOpenAIToAnthropicSystemMessage(t *testing.T) {
	req := &ChatCompletionRequest{
		Model: "gpt-4",
		Messages: []ChatMessage{
			{Role: "system", Content: "You are helpful."},
			{Role: "user", Content: "Hello"},
		},
		MaxTokens: intPtr(1024),
	}

	result, err := OpenAIToAnthropic(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// system message should be extracted to System field
	if result.System == nil || result.System != "You are helpful." {
		t.Errorf("incorrect System: %v", result.System)
	}
	// Messages should not contain system role
	for _, msg := range result.Messages {
		if msg.Role == "system" {
			t.Error("Anthropic Messages should not contain system role")
		}
	}
}

// TestOpenAIToAnthropicMaxTokensDefault tests default max_tokens.
func TestOpenAIToAnthropicMaxTokensDefault(t *testing.T) {
	req := &ChatCompletionRequest{
		Model:    "gpt-4",
		Messages: []ChatMessage{{Role: "user", Content: "Hi"}},
	}

	result, err := OpenAIToAnthropic(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.MaxTokens != 4096 {
		t.Errorf("default MaxTokens should be 4096, got: %d", result.MaxTokens)
	}
}

// TestOpenAIToAnthropicMaxCompletionTokensPriority tests max_completion_tokens priority.
func TestOpenAIToAnthropicMaxCompletionTokensPriority(t *testing.T) {
	maxTokens := 1024
	maxCompletionTokens := 2048
	req := &ChatCompletionRequest{
		Model:               "gpt-4",
		Messages:            []ChatMessage{{Role: "user", Content: "Hi"}},
		MaxTokens:           &maxTokens,
		MaxCompletionTokens: &maxCompletionTokens,
	}

	result, err := OpenAIToAnthropic(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.MaxTokens != 2048 {
		t.Errorf("should prioritize max_completion_tokens: %d", result.MaxTokens)
	}
}

// TestOpenAIToAnthropicToolConversion tests tool definition conversion.
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

	result, err := OpenAIToAnthropic(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Tools) != 1 {
		t.Fatalf("incorrect Tools count: %d", len(result.Tools))
	}
	tool := result.Tools[0]
	if tool.Name != "get_weather" || tool.Description != "获取天气" {
		t.Errorf("incorrect tool definition: name=%s, desc=%s", tool.Name, tool.Description)
	}
}

// TestOpenAIToAnthropicToolChoice tests tool_choice conversion.
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

			result, err := OpenAIToAnthropic(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
			if result.ToolChoice == nil {
				t.Fatal("ToolChoice should not be nil")
			}
			tc, ok := result.ToolChoice.(AnthropicToolChoice)
			if !ok {
				t.Fatalf("incorrect ToolChoice type: %T", result.ToolChoice)
			}
			if tc.Type != tt.expected {
				t.Errorf("incorrect ToolChoice.Type: expected %s, got %s", tt.expected, tc.Type)
			}
		})
	}
}

// TestOpenAIToAnthropicToolCallConversion tests tool_calls to tool_use conversion.
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

	result, err := OpenAIToAnthropic(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// assistant message should contain tool_use content blocks
	assistantMsg := result.Messages[1]
	blocks, ok := assistantMsg.Content.([]AnthropicContentBlock)
	if !ok {
		t.Fatal("assistant content should be []AnthropicContentBlock")
	}
	if len(blocks) == 0 || blocks[0].Type != "tool_use" {
		t.Error("should contain tool_use content block")
	}

	// tool message should convert to user role + tool_result content block
	toolResultMsg := result.Messages[2]
	if toolResultMsg.Role != "user" {
		t.Errorf("tool message should convert to user role: %s", toolResultMsg.Role)
	}
}

// TestOpenAIToAnthropicStopSequence tests stop sequence conversion.
func TestOpenAIToAnthropicStopSequence(t *testing.T) {
	// string type
	req := &ChatCompletionRequest{
		Model:    "gpt-4",
		Messages: []ChatMessage{{Role: "user", Content: "Hi"}},
		Stop:     "END",
	}
	result, err := OpenAIToAnthropic(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.StopSequences) != 1 || result.StopSequences[0] != "END" {
		t.Errorf("incorrect StopSequences: %v", result.StopSequences)
	}
}

// TestOpenAIToAnthropicFirstMessageAssistant 首条 assistant 消息合法，不插入假 user 消息。
func TestOpenAIToAnthropicFirstMessageAssistant(t *testing.T) {
	req := &ChatCompletionRequest{
		Model: "gpt-4",
		Messages: []ChatMessage{
			{Role: "assistant", Content: "I'm ready."},
		},
	}

	result, err := OpenAIToAnthropic(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Messages) != 1 || result.Messages[0].Role != "assistant" {
		t.Errorf("first message should remain assistant, got: %+v", result.Messages)
	}
}

// TestOpenAIToAnthropicEmptyMessages 空 messages 应返回错误。
func TestOpenAIToAnthropicEmptyMessages(t *testing.T) {
	req := &ChatCompletionRequest{
		Model:    "gpt-4",
		Messages: []ChatMessage{},
	}
	_, err := OpenAIToAnthropic(req)
	if err == nil {
		t.Error("expected error for empty messages")
	}
}

// TestOpenAIToAnthropicMergeToolMessages 连续 tool 消息合并为一条 user 消息。
func TestOpenAIToAnthropicMergeToolMessages(t *testing.T) {
	req := &ChatCompletionRequest{
		Model: "gpt-4",
		Messages: []ChatMessage{
			{Role: "user", Content: "query"},
			{Role: "tool", ToolCallID: "call_1", Content: "result1"},
			{Role: "tool", ToolCallID: "call_2", Content: "result2"},
		},
	}

	result, err := OpenAIToAnthropic(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(result.Messages))
	}
	blocks, ok := result.Messages[1].Content.([]AnthropicContentBlock)
	if !ok || len(blocks) != 2 {
		t.Fatalf("expected merged tool_result blocks, got: %+v", result.Messages[1].Content)
	}
	if blocks[0].ToolUseID != "call_1" || blocks[1].ToolUseID != "call_2" {
		t.Error("tool_use_id mismatch in merged blocks")
	}
}

// TestOpenAIToAnthropicImageURL 含 image_url 的 user 消息转为 Anthropic image block。
func TestOpenAIToAnthropicImageURL(t *testing.T) {
	req := &ChatCompletionRequest{
		Model: "gpt-4",
		Messages: []ChatMessage{{
			Role: RoleUser,
			Content: []ContentPart{{
				Type: "image_url",
				ImageURL: &ImageURL{
					URL: "data:image/png;base64,abc123",
				},
			}},
		}},
	}

	result, err := OpenAIToAnthropic(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	blocks, ok := result.Messages[0].Content.([]AnthropicContentBlock)
	if !ok || len(blocks) != 1 {
		t.Fatalf("expected image content block, got: %+v", result.Messages[0].Content)
	}
	if blocks[0].Type != "image" || blocks[0].Source == nil {
		t.Fatalf("expected image block, got: %+v", blocks[0])
	}
	if blocks[0].Source.Type != "base64" || blocks[0].Source.MediaType != "image/png" {
		t.Errorf("unexpected image source: %+v", blocks[0].Source)
	}
}

// ══════════════════════════════════
// Anthropic to OpenAI request conversion
// ══════════════════════════════════

// TestAnthropicToOpenAIBasic tests basic request conversion.
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
		t.Errorf("incorrect Model: %s", result.Model)
	}
	if result.MaxTokens == nil || *result.MaxTokens != 2048 {
		t.Error("incorrect MaxTokens")
	}
	if !result.Stream {
		t.Error("Stream should be true")
	}

	// system should convert to a system message
	if len(result.Messages) < 2 {
		t.Fatal("should have at least 2 messages (system + user)")
	}
	if result.Messages[0].Role != "system" {
		t.Error("first message should be system role")
	}
	if result.Messages[0].ContentString() != "You are helpful." {
		t.Error("incorrect system content")
	}
}

// TestAnthropicToOpenAICollapseSystemMessages 验证 messages 中间的 role=system
// 消息会被合并到首位。这是 vLLM 等严格上游的要求（"System message must be at the beginning"）。
// 复现 Claude Code 2.1.x 在多轮工具调用中向 messages 中间插入 system 消息的场景。
func TestAnthropicToOpenAICollapseSystemMessages(t *testing.T) {
	req := &AnthropicMessagesRequest{
		Model: "qwen3.6-27b",
		// 模拟 Claude Code 的真实结构：顶层 system + messages 中间混入 system
		System: "You are a helpful assistant.",
		Messages: []AnthropicMessage{
			{Role: "user", Content: "first question"},
			{Role: "system", Content: "ToolSearch hints available now"}, // 中间的 system（关键）
			{Role: "assistant", Content: "answer"},
			{Role: "user", Content: "second question"},
		},
		MaxTokens: 1024,
	}

	result := AnthropicToOpenAI(req)

	// 所有 system 消息必须合并到第一位
	if len(result.Messages) == 0 || result.Messages[0].Role != "system" {
		t.Fatalf("first message must be system, got role=%q (len=%d)",
			safeRole(result.Messages), len(result.Messages))
	}

	// 后续消息中不得再出现 system
	for i, msg := range result.Messages[1:] {
		if msg.Role == "system" {
			t.Errorf("message[%d] should not be system after collapse; roles=%v", i+1, rolesOf(result.Messages))
		}
	}

	// 合并后的 system 内容必须包含两段 system 文本
	systemContent := result.Messages[0].ContentString()
	if !strings.Contains(systemContent, "You are a helpful assistant.") {
		t.Errorf("merged system missing top-level system text: %q", systemContent)
	}
	if !strings.Contains(systemContent, "ToolSearch hints available now") {
		t.Errorf("merged system missing mid-stream system text: %q", systemContent)
	}

	// 非系统消息的相对顺序必须保持：user → assistant → user
	nonSystemRoles := rolesOfNonSystem(result.Messages)
	expected := []string{"user", "assistant", "user"}
	if len(nonSystemRoles) != len(expected) {
		t.Fatalf("non-system message count = %v, want %v", nonSystemRoles, expected)
	}
	for i, role := range expected {
		if nonSystemRoles[i] != role {
			t.Errorf("non-system[%d] role = %q, want %q; full=%v", i, nonSystemRoles[i], role, nonSystemRoles)
		}
	}
}

func safeRole(msgs []ChatMessage) string {
	if len(msgs) == 0 {
		return "<empty>"
	}
	return msgs[0].Role
}

func rolesOf(msgs []ChatMessage) []string {
	out := make([]string, len(msgs))
	for i, m := range msgs {
		out[i] = m.Role
	}
	return out
}

func rolesOfNonSystem(msgs []ChatMessage) []string {
	var out []string
	for _, m := range msgs {
		if m.Role != "system" {
			out = append(out, m.Role)
		}
	}
	return out
}

// TestAnthropicToOpenAIToolChoice tests tool_choice conversion.
func TestAnthropicToOpenAIToolChoice(t *testing.T) {
	tests := []struct {
		choice   interface{}
		expected interface{}
		name     string
	}{
		{
			name:     "auto",
			choice:   map[string]interface{}{"type": "auto"},
			expected: "auto",
		},
		{
			name:     "required",
			choice:   map[string]interface{}{"type": "any"},
			expected: "required",
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
				t.Errorf("incorrect ToolChoice: expected %v, got %v", tt.expected, result.ToolChoice)
			}
		})
	}
}

// TestAnthropicToOpenAIToolResultConversion tests tool_result to tool message conversion.
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

	// Find tool role message
	var toolMsg *ChatMessage
	for i := range result.Messages {
		if result.Messages[i].Role == "tool" {
			toolMsg = &result.Messages[i]
			break
		}
	}
	if toolMsg == nil {
		t.Fatal("should generate tool role message")
	}
	if toolMsg.ToolCallID != "toolu_1" {
		t.Errorf("incorrect ToolCallID: %s", toolMsg.ToolCallID)
	}

	// Find assistant message's tool_calls
	var assistantMsg *ChatMessage
	for i := range result.Messages {
		if result.Messages[i].Role == "assistant" {
			assistantMsg = &result.Messages[i]
			break
		}
	}
	if assistantMsg == nil {
		t.Fatal("should have assistant message")
	}
	if len(assistantMsg.ToolCalls) != 1 {
		t.Fatalf("should have 1 tool_call, got: %d", len(assistantMsg.ToolCalls))
	}
	if assistantMsg.ToolCalls[0].Function.Name != "get_weather" {
		t.Errorf("incorrect tool_call name: %s", assistantMsg.ToolCalls[0].Function.Name)
	}
}

// ══════════════════════════════════
// Response conversion
// ══════════════════════════════════

// TestOpenAIResponseToAnthropic tests OpenAI response to Anthropic format.
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
		t.Errorf("incorrect ID: %s", result.ID)
	}
	if result.Type != "message" {
		t.Errorf("incorrect Type: %s", result.Type)
	}
	if result.Role != "assistant" {
		t.Errorf("incorrect Role: %s", result.Role)
	}

	// Should have text + tool_use content blocks
	if len(result.Content) != 2 {
		t.Fatalf("incorrect Content count: %d", len(result.Content))
	}
	if result.Content[0].Type != "text" || result.Content[0].Text != "我来查天气" {
		t.Error("incorrect text content block")
	}
	if result.Content[1].Type != "tool_use" || result.Content[1].Name != "get_weather" {
		t.Error("incorrect tool_use content block")
	}

	// stop_reason should be mapped
	if result.StopReason == nil || *result.StopReason != "tool_use" {
		t.Error("incorrect StopReason mapping")
	}

	// usage should be converted
	if result.Usage == nil || result.Usage.InputTokens != 50 || result.Usage.OutputTokens != 30 {
		t.Error("incorrect Usage conversion")
	}
}

// TestAnthropicResponseToOpenAI tests Anthropic response to OpenAI format.
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
		t.Errorf("incorrect ID: %s", result.ID)
	}
	if result.Object != "chat.completion" {
		t.Errorf("incorrect Object: %s", result.Object)
	}

	if len(result.Choices) != 1 {
		t.Fatalf("incorrect Choices count: %d", len(result.Choices))
	}
	choice := result.Choices[0]

	// thinking blocks should be skipped, only text content
	if choice.Message.ContentString() != "The answer is 42." {
		t.Errorf("incorrect Content (thinking should be skipped): %s", choice.Message.ContentString())
	}

	// tool_use → tool_calls
	if len(choice.Message.ToolCalls) != 1 {
		t.Fatalf("incorrect ToolCalls count: %d", len(choice.Message.ToolCalls))
	}
	if choice.Message.ToolCalls[0].Function.Name != "calc" {
		t.Error("incorrect ToolCall name")
	}

	// stop_reason → finish_reason
	if choice.FinishReason == nil || *choice.FinishReason != "stop" {
		t.Error("incorrect FinishReason mapping")
	}
}

// ══════════════════════════════════
// Stop reason mapping
// ══════════════════════════════════

// TestStopReasonMapping tests bidirectional stop reason mapping.
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
			t.Errorf("OpenAI→Anthropic: %s → %s (expected %s)", tt.input, result, tt.expected)
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
			t.Errorf("Anthropic→OpenAI: %s → %s (expected %s)", tt.input, result, tt.expected)
		}
	}
}

// ══════════════════════════════════
// tool_choice mapping
// ══════════════════════════════════

// TestToolChoiceMapping tests bidirectional tool_choice mapping.
func TestToolChoiceMapping(t *testing.T) {
	// OpenAI → Anthropic
	t.Run("OpenAI auto → Anthropic auto", func(t *testing.T) {
		result := mapOpenAIToolChoiceToAnthropic("auto")
		tc, ok := result.(AnthropicToolChoice)
		if !ok || tc.Type != "auto" {
			t.Errorf("expected {type:auto}, got: %v", result)
		}
	})

	t.Run("OpenAI required → Anthropic any", func(t *testing.T) {
		result := mapOpenAIToolChoiceToAnthropic("required")
		tc, ok := result.(AnthropicToolChoice)
		if !ok || tc.Type != "any" {
			t.Errorf("expected {type:any}, got: %v", result)
		}
	})

	t.Run("OpenAI none → nil", func(t *testing.T) {
		result := mapOpenAIToolChoiceToAnthropic("none")
		if result != nil {
			t.Errorf("expected nil, got: %v", result)
		}
	})

	t.Run("OpenAI specific function → Anthropic tool", func(t *testing.T) {
		input := map[string]interface{}{
			"type":     "function",
			"function": map[string]interface{}{"name": "get_weather"},
		}
		result := mapOpenAIToolChoiceToAnthropic(input)
		tc, ok := result.(AnthropicToolChoice)
		if !ok || tc.Type != "tool" || tc.Name != "get_weather" {
			t.Errorf("expected {type:tool, name:get_weather}, got: %v", result)
		}
	})

	// Anthropic → OpenAI
	t.Run("Anthropic auto → OpenAI auto", func(t *testing.T) {
		result := mapAnthropicToolChoiceToOpenAI(map[string]interface{}{"type": "auto"})
		if result != "auto" {
			t.Errorf("expected auto, got: %v", result)
		}
	})

	t.Run("Anthropic any → OpenAI required", func(t *testing.T) {
		result := mapAnthropicToolChoiceToOpenAI(map[string]interface{}{"type": "any"})
		if result != "required" {
			t.Errorf("expected required, got: %v", result)
		}
	})

	t.Run("Anthropic none → OpenAI none", func(t *testing.T) {
		result := mapAnthropicToolChoiceToOpenAI(map[string]interface{}{"type": "none"})
		if result != "none" {
			t.Errorf("expected none, got: %v", result)
		}
	})

	t.Run("Anthropic tool → OpenAI specific function", func(t *testing.T) {
		result := mapAnthropicToolChoiceToOpenAI(map[string]interface{}{
			"type": "tool",
			"name": "get_weather",
		})
		m, ok := result.(map[string]interface{})
		if !ok {
			t.Fatalf("expected map, got: %T", result)
		}
		fn, _ := m["function"].(map[string]interface{})
		if fn["name"] != "get_weather" {
			t.Errorf("incorrect function name: %v", fn["name"])
		}
	})
}

// ══════════════════════════════════
// Stream format conversion: Anthropic to OpenAI
// ══════════════════════════════════

// TestAnthropicEventToOpenAIChunkText tests text delta conversion.
func TestAnthropicEventToOpenAIChunkText(t *testing.T) {
	state := &AnthropicToOpenAIState{}

	// message_start → role chunk
	event := &AnthropicStreamEvent{Type: AnthropicEventMessageStart}
	result := AnthropicEventToOpenAIChunk(AnthropicEventMessageStart, event, "claude-sonnet-4-20250514", state)
	if !strings.Contains(result, `"role":"assistant"`) {
		t.Error("message_start should generate chunk with role:assistant")
	}

	// text_delta
	event = &AnthropicStreamEvent{
		Type:  AnthropicEventContentBlockDelta,
		Delta: &AnthropicStreamDelta{Type: "text_delta", Text: "Hello"},
	}
	result = AnthropicEventToOpenAIChunk(AnthropicEventContentBlockDelta, event, "claude-sonnet-4-20250514", state)
	if !strings.Contains(result, "Hello") {
		t.Error("text_delta should contain text content")
	}
	if !strings.Contains(result, "chat.completion.chunk") {
		t.Error("should be chat.completion.chunk format")
	}

	// message_stop → [DONE]
	event = &AnthropicStreamEvent{Type: AnthropicEventMessageStop}
	result = AnthropicEventToOpenAIChunk(AnthropicEventMessageStop, event, "claude-sonnet-4-20250514", state)
	if result != "data: [DONE]\n\n" {
		t.Errorf("message_stop should generate [DONE], got: %s", result)
	}
}

// TestAnthropicEventToOpenAIChunkToolUse tests tool call stream conversion.
func TestAnthropicEventToOpenAIChunkToolUse(t *testing.T) {
	state := &AnthropicToOpenAIState{}

	// content_block_start (tool_use)
	event := &AnthropicStreamEvent{
		Type:         AnthropicEventContentBlockStart,
		ContentBlock: &AnthropicContentBlock{Type: "tool_use", ID: "toolu_01", Name: "get_weather"},
	}
	result := AnthropicEventToOpenAIChunk(AnthropicEventContentBlockStart, event, "claude-sonnet-4-20250514", state)
	if !strings.Contains(result, "get_weather") {
		t.Error("tool_use start should contain tool name")
	}
	if !strings.Contains(result, "toolu_01") {
		t.Error("tool_use start should contain tool ID")
	}
	if !strings.Contains(result, "tool_calls") {
		t.Error("should contain tool_calls field")
	}

	// input_json_delta
	event = &AnthropicStreamEvent{
		Type:  AnthropicEventContentBlockDelta,
		Delta: &AnthropicStreamDelta{Type: "input_json_delta", PartialJSON: `{"city":`},
	}
	result = AnthropicEventToOpenAIChunk(AnthropicEventContentBlockDelta, event, "claude-sonnet-4-20250514", state)
	if !strings.Contains(result, "tool_calls") {
		t.Error("input_json_delta should convert to tool_calls arguments")
	}

	// content_block_stop increments index
	event = &AnthropicStreamEvent{Type: AnthropicEventContentBlockStop}
	AnthropicEventToOpenAIChunk(AnthropicEventContentBlockStop, event, "claude-sonnet-4-20250514", state)
	if state.ToolCallIndex != 1 {
		t.Errorf("index should increment after tool_use ends: %d", state.ToolCallIndex)
	}
}

// TestAnthropicEventToOpenAIChunkThinkingIgnored tests thinking events are ignored.
func TestAnthropicEventToOpenAIChunkThinkingIgnored(t *testing.T) {
	state := &AnthropicToOpenAIState{}

	// content_block_start (thinking) produces no output
	event := &AnthropicStreamEvent{
		Type:         AnthropicEventContentBlockStart,
		ContentBlock: &AnthropicContentBlock{Type: "thinking"},
	}
	result := AnthropicEventToOpenAIChunk(AnthropicEventContentBlockStart, event, "claude-sonnet-4-20250514", state)
	if result != "" {
		t.Errorf("thinking block_start should not produce any output: %s", result)
	}

	// thinking_delta produces no output
	event = &AnthropicStreamEvent{
		Type:  AnthropicEventContentBlockDelta,
		Delta: &AnthropicStreamDelta{Type: "thinking_delta", Thinking: "thinking..."},
	}
	result = AnthropicEventToOpenAIChunk(AnthropicEventContentBlockDelta, event, "claude-sonnet-4-20250514", state)
	if result != "" {
		t.Errorf("thinking_delta should not produce any output: %s", result)
	}

	// signature_delta produces no output
	event = &AnthropicStreamEvent{
		Type:  AnthropicEventContentBlockDelta,
		Delta: &AnthropicStreamDelta{Type: "signature_delta", Signature: "sig..."},
	}
	result = AnthropicEventToOpenAIChunk(AnthropicEventContentBlockDelta, event, "claude-sonnet-4-20250514", state)
	if result != "" {
		t.Errorf("signature_delta should not produce any output: %s", result)
	}
}

// TestAnthropicEventToOpenAIChunkMessageDelta tests message_delta (with usage) conversion.
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
		t.Error("should contain mapped finish_reason: stop")
	}

	// Verify usage attachment
	var chunk ChatCompletionChunk
	// Extract JSON after data: prefix
	jsonStr := strings.TrimPrefix(strings.TrimSpace(result), "data: ")
	jsonStr = strings.TrimSuffix(jsonStr, "\n\n")
	if err := json.Unmarshal([]byte(jsonStr), &chunk); err != nil {
		t.Fatalf("failed to parse chunk JSON: %v", err)
	}
	if chunk.Usage == nil {
		t.Fatal("chunk should contain usage")
	}
	if chunk.Usage.CompletionTokens != 50 {
		t.Errorf("incorrect CompletionTokens: %d", chunk.Usage.CompletionTokens)
	}
}

// TestOpenAIUsageToAnthropicCacheDeduction 验证缓存 token 扣减。
func TestOpenAIUsageToAnthropicCacheDeduction(t *testing.T) {
	usage := openAIUsageToAnthropic(&Usage{
		PromptTokens:     100,
		CompletionTokens: 20,
		PromptTokensDetails: &PromptTokensDetails{
			CachedTokens: 60,
		},
	})
	if usage.InputTokens != 40 {
		t.Errorf("expected input_tokens=40, got %d", usage.InputTokens)
	}
	if usage.CacheReadInputTokens != 60 {
		t.Errorf("expected cache_read_input_tokens=60, got %d", usage.CacheReadInputTokens)
	}
}

// TestAnthropicToOpenAIStreamOptions 流式请求应注入 stream_options.include_usage。
func TestAnthropicToOpenAIStreamOptions(t *testing.T) {
	req := &AnthropicMessagesRequest{
		Model:     "claude-sonnet-4-20250514",
		Messages:  []AnthropicMessage{{Role: "user", Content: "Hi"}},
		MaxTokens: 1024,
		Stream:    true,
	}
	result := AnthropicToOpenAI(req)
	if result.StreamOptions == nil || result.StreamOptions.IncludeUsage == nil || !*result.StreamOptions.IncludeUsage {
		t.Error("stream request should include stream_options.include_usage")
	}
}

// ══════════════════════════════════
// Stream format conversion: OpenAI to Anthropic
// ══════════════════════════════════

// TestOpenAIChunkToAnthropicEventsLazyTextStart 首 chunk 无 content 时，后续 text 前必须先 content_block_start。
func TestOpenAIChunkToAnthropicEventsLazyTextStart(t *testing.T) {
	state := &OpenAIToAnthropicState{}

	// 首 chunk 仅 role
	roleChunk := &ChatCompletionChunk{
		ID: "chatcmpl-1", Model: "gpt-4",
		Choices: []ChatChoice{{Index: 0, Delta: &ChatMessage{Role: RoleAssistant}}},
	}
	result := OpenAIChunkToAnthropicEvents(roleChunk, true, state)
	if !strings.Contains(result, "message_start") {
		t.Error("should contain message_start")
	}
	if strings.Contains(result, "content_block_start") {
		t.Error("role-only chunk should not start text block")
	}

	// 后续 text chunk
	textChunk := &ChatCompletionChunk{
		ID: "chatcmpl-1", Model: "gpt-4",
		Choices: []ChatChoice{{Index: 0, Delta: &ChatMessage{Content: "Hi"}}},
	}
	result = OpenAIChunkToAnthropicEvents(textChunk, false, state)
	if !strings.Contains(result, "content_block_start") {
		t.Error("text chunk should start text block")
	}
	if !strings.Contains(result, "text_delta") {
		t.Error("should contain text_delta")
	}
}

// TestOpenAIChunkToAnthropicEventsTextThenTool text 后接 tool_call，text 块应先 stop。
func TestOpenAIChunkToAnthropicEventsTextThenTool(t *testing.T) {
	state := &OpenAIToAnthropicState{}
	idx := 0

	OpenAIChunkToAnthropicEvents(&ChatCompletionChunk{
		ID: "c1", Model: "gpt-4",
		Choices: []ChatChoice{{Delta: &ChatMessage{Content: "thinking"}}},
	}, true, state)

	toolChunk := &ChatCompletionChunk{
		ID: "c1", Model: "gpt-4",
		Choices: []ChatChoice{{Delta: &ChatMessage{
			ToolCalls: []ToolCall{{
				Index: &idx, ID: "call_1", Type: "function",
				Function: ToolCallFunction{Name: "get_weather"},
			}},
		}}},
	}
	result := OpenAIChunkToAnthropicEvents(toolChunk, false, state)
	if !strings.Contains(result, "content_block_stop") {
		t.Error("should stop text block before tool_use")
	}
	if !strings.Contains(result, "tool_use") {
		t.Error("should start tool_use block")
	}
}

// TestOpenAIChunkToAnthropicEventsParallelTools 并行 tool calls 按 index 独立管理。
func TestOpenAIChunkToAnthropicEventsParallelTools(t *testing.T) {
	state := &OpenAIToAnthropicState{}
	idx0, idx1 := 0, 1

	OpenAIChunkToAnthropicEvents(&ChatCompletionChunk{
		ID: "c1", Model: "gpt-4",
		Choices: []ChatChoice{{Delta: &ChatMessage{Role: RoleAssistant}}},
	}, true, state)

	// tool 0 name
	OpenAIChunkToAnthropicEvents(&ChatCompletionChunk{
		ID: "c1", Model: "gpt-4",
		Choices: []ChatChoice{{Delta: &ChatMessage{ToolCalls: []ToolCall{{
			Index: &idx0, ID: "call_0", Function: ToolCallFunction{Name: "tool_a"},
		}}}}},
	}, false, state)

	// tool 1 name
	OpenAIChunkToAnthropicEvents(&ChatCompletionChunk{
		ID: "c1", Model: "gpt-4",
		Choices: []ChatChoice{{Delta: &ChatMessage{ToolCalls: []ToolCall{{
			Index: &idx1, ID: "call_1", Function: ToolCallFunction{Name: "tool_b"},
		}}}}},
	}, false, state)

	// tool 0 args
	r0 := OpenAIChunkToAnthropicEvents(&ChatCompletionChunk{
		ID: "c1", Model: "gpt-4",
		Choices: []ChatChoice{{Delta: &ChatMessage{ToolCalls: []ToolCall{{
			Index: &idx0, Function: ToolCallFunction{Arguments: `{"a":1}`},
		}}}}},
	}, false, state)
	if !strings.Contains(r0, "input_json_delta") {
		t.Error("tool 0 should receive args delta")
	}

	// tool 1 args
	r1 := OpenAIChunkToAnthropicEvents(&ChatCompletionChunk{
		ID: "c1", Model: "gpt-4",
		Choices: []ChatChoice{{Delta: &ChatMessage{ToolCalls: []ToolCall{{
			Index: &idx1, Function: ToolCallFunction{Arguments: `{"b":2}`},
		}}}}},
	}, false, state)
	if !strings.Contains(r1, "input_json_delta") {
		t.Error("tool 1 should receive args delta")
	}
	if len(state.ToolBlocks) != 2 {
		t.Errorf("expected 2 tool blocks, got %d", len(state.ToolBlocks))
	}
}

// TestOpenAIChunkToAnthropicEventsPendingArgs args 先于 name 到达时应缓存后补发。
func TestOpenAIChunkToAnthropicEventsPendingArgs(t *testing.T) {
	state := &OpenAIToAnthropicState{}
	idx := 0

	OpenAIChunkToAnthropicEvents(&ChatCompletionChunk{
		ID: "c1", Model: "gpt-4",
		Choices: []ChatChoice{{Delta: &ChatMessage{Role: RoleAssistant}}},
	}, true, state)

	// args 先到
	OpenAIChunkToAnthropicEvents(&ChatCompletionChunk{
		ID: "c1", Model: "gpt-4",
		Choices: []ChatChoice{{Delta: &ChatMessage{ToolCalls: []ToolCall{{
			Index: &idx, Function: ToolCallFunction{Arguments: `{"x":1}`},
		}}}}},
	}, false, state)

	// name/id 后到
	result := OpenAIChunkToAnthropicEvents(&ChatCompletionChunk{
		ID: "c1", Model: "gpt-4",
		Choices: []ChatChoice{{Delta: &ChatMessage{ToolCalls: []ToolCall{{
			Index: &idx, ID: "call_1", Function: ToolCallFunction{Name: "my_tool"},
		}}}}},
	}, false, state)
	if !strings.Contains(result, "content_block_start") {
		t.Error("should start tool block when id+name arrive")
	}
	if !strings.Contains(result, "input_json_delta") {
		t.Error("should flush pending args after start")
	}
}

// TestOpenAIChunkToAnthropicEventsFinishDeferred finish 不立即发 message_delta，由 finalize 发出。
func TestOpenAIChunkToAnthropicEventsFinishDeferred(t *testing.T) {
	state := &OpenAIToAnthropicState{}
	finishReason := "stop"

	result := OpenAIChunkToAnthropicEvents(&ChatCompletionChunk{
		ID: "c1", Model: "gpt-4",
		Choices: []ChatChoice{{Delta: &ChatMessage{Content: "done"}, FinishReason: &finishReason}},
	}, true, state)
	if strings.Contains(result, "message_delta") {
		t.Error("finish chunk should not emit message_delta immediately")
	}

	// usage chunk（choices 为空）
	OpenAIChunkToAnthropicEvents(&ChatCompletionChunk{
		ID: "c1", Model: "gpt-4", Choices: []ChatChoice{},
		Usage: &Usage{PromptTokens: 10, CompletionTokens: 5},
	}, false, state)

	final := OpenAIStreamFinalize(state)
	if !strings.Contains(final, "message_delta") {
		t.Error("finalize should emit message_delta")
	}
	if !strings.Contains(final, `"input_tokens":10`) {
		t.Error("finalize should include usage")
	}
	if !strings.Contains(final, "message_stop") {
		t.Error("finalize should emit message_stop")
	}
}

// TestOpenAIStreamFinalizeEOFWithoutFinish 无 finish_reason 时 EOF 收尾仍发出完整序列。
func TestOpenAIStreamFinalizeEOFWithoutFinish(t *testing.T) {
	state := &OpenAIToAnthropicState{}
	OpenAIChunkToAnthropicEvents(&ChatCompletionChunk{
		ID: "c1", Model: "gpt-4",
		Choices: []ChatChoice{{Delta: &ChatMessage{Content: "partial"}}},
	}, true, state)

	final := OpenAIStreamFinalize(state)
	if !strings.Contains(final, `"stop_reason":"end_turn"`) {
		t.Error("should default stop_reason to end_turn")
	}
	if !strings.Contains(final, "message_stop") {
		t.Error("should emit message_stop")
	}
}

// TestOpenAIStreamFinalizeNoDuplicate 重复 finalize 只发一次 message_delta。
func TestOpenAIStreamFinalizeNoDuplicate(t *testing.T) {
	state := &OpenAIToAnthropicState{}
	finishReason := "stop"
	OpenAIChunkToAnthropicEvents(&ChatCompletionChunk{
		ID: "c1", Model: "gpt-4",
		Choices: []ChatChoice{{FinishReason: &finishReason}},
	}, true, state)

	first := OpenAIStreamFinalize(state)
	second := OpenAIStreamFinalize(state)
	if first == "" {
		t.Fatal("first finalize should produce output")
	}
	if second != "" {
		t.Error("second finalize should be empty")
	}
}

// TestOpenAIChunkToAnthropicEventsDuplicateFinish 重复 finish chunk 不重复发 message_delta。
func TestOpenAIChunkToAnthropicEventsDuplicateFinish(t *testing.T) {
	state := &OpenAIToAnthropicState{}
	finishReason := "stop"

	OpenAIChunkToAnthropicEvents(&ChatCompletionChunk{
		ID: "c1", Model: "gpt-4",
		Choices: []ChatChoice{{FinishReason: &finishReason}},
	}, true, state)

	// 第二个 finish chunk 不应在 chunk 处理阶段发 message_delta
	result := OpenAIChunkToAnthropicEvents(&ChatCompletionChunk{
		ID: "c1", Model: "gpt-4",
		Choices: []ChatChoice{{FinishReason: &finishReason}},
	}, false, state)
	if strings.Contains(result, "message_delta") {
		t.Error("duplicate finish should not emit message_delta in chunk handler")
	}

	final := OpenAIStreamFinalize(state)
	if strings.Count(final, "event: message_delta") != 1 {
		t.Error("finalize should emit exactly one message_delta")
	}
}

// TestOpenAIResponseToAnthropicCacheDeduction 非流式响应 usage 缓存 token 扣减。
func TestOpenAIResponseToAnthropicCacheDeduction(t *testing.T) {
	resp := &ChatCompletionResponse{
		ID:    "chatcmpl-1",
		Model: "gpt-4",
		Choices: []ChatChoice{{
			Message: &ChatMessage{Role: RoleAssistant, Content: "ok"},
		}},
		Usage: &Usage{
			PromptTokens:     100,
			CompletionTokens: 20,
			PromptTokensDetails: &PromptTokensDetails{
				CachedTokens: 60,
			},
		},
	}
	result := OpenAIResponseToAnthropic(resp)
	if result.Usage == nil || result.Usage.InputTokens != 40 {
		t.Errorf("expected input_tokens=40, got %+v", result.Usage)
	}
	if result.Usage.CacheReadInputTokens != 60 {
		t.Errorf("expected cache_read_input_tokens=60, got %d", result.Usage.CacheReadInputTokens)
	}
}

// TestAnthropicEventToOpenAIChunkFixedID 流式 chunk id 应在整条流中固定复用。
func TestAnthropicEventToOpenAIChunkFixedID(t *testing.T) {
	state := &AnthropicToOpenAIState{}
	model := "claude-sonnet-4-20250514"

	start := AnthropicEventToOpenAIChunk(AnthropicEventMessageStart, &AnthropicStreamEvent{Type: AnthropicEventMessageStart}, model, state)
	text := AnthropicEventToOpenAIChunk(AnthropicEventContentBlockDelta, &AnthropicStreamEvent{
		Type:  AnthropicEventContentBlockDelta,
		Delta: &AnthropicStreamDelta{Type: "text_delta", Text: "Hi"},
	}, model, state)

	extractID := func(s string) string {
		const prefix = `"id":"`
		idx := strings.Index(s, prefix)
		if idx < 0 {
			return ""
		}
		rest := s[idx+len(prefix):]
		end := strings.Index(rest, `"`)
		if end < 0 {
			return ""
		}
		return rest[:end]
	}

	id1 := extractID(start)
	id2 := extractID(text)
	if id1 == "" || id2 == "" {
		t.Fatal("failed to extract chunk id")
	}
	if id1 != id2 {
		t.Errorf("chunk id should be stable across events: %s vs %s", id1, id2)
	}
}

// TestAnthropicToOpenAIOSeriesModel o 系列模型应将 max_tokens 映射为 max_completion_tokens。
func TestAnthropicToOpenAIOSeriesModel(t *testing.T) {
	req := &AnthropicMessagesRequest{
		Model:     "o3-mini",
		Messages:  []AnthropicMessage{{Role: "user", Content: "Hi"}},
		MaxTokens: 2048,
	}
	result := AnthropicToOpenAI(req)
	if result.MaxCompletionTokens == nil || *result.MaxCompletionTokens != 2048 {
		t.Errorf("expected max_completion_tokens=2048, got %+v", result.MaxCompletionTokens)
	}
	if result.MaxTokens != nil {
		t.Error("o-series model should not set max_tokens")
	}
}

// ══════════════════════════════════
// Claude Code multi-turn edge cases
// ══════════════════════════════════

// TestAnthropicToOpenAIToolResultMissingContent 验证 tool_result 缺失 content 字段时，
// 转换后的 tool 消息仍保留 content 字段（空字符串），而非被 omitempty 省略。
// 严格上游（如部分 vLLM/DashScope）要求 tool 消息必须携带 content，缺失会返回 400。
func TestAnthropicToOpenAIToolResultMissingContent(t *testing.T) {
	reqJSON := `{
		"model": "glm-4.6",
		"max_tokens": 1024,
		"messages": [
			{"role": "user", "content": "hi"},
			{"role": "assistant", "content": [
				{"type": "tool_use", "id": "toolu_1", "name": "Read", "input": {"file_path": "/tmp/a"}}
			]},
			{"role": "user", "content": [
				{"type": "tool_result", "tool_use_id": "toolu_1"}
			]},
			{"role": "user", "content": "第二句"}
		]
	}`

	var req AnthropicMessagesRequest
	if err := json.Unmarshal([]byte(reqJSON), &req); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	result := AnthropicToOpenAI(&req)

	var toolMsgs []ChatMessage
	for _, m := range result.Messages {
		if m.Role == RoleTool {
			toolMsgs = append(toolMsgs, m)
		}
	}
	if len(toolMsgs) != 1 {
		t.Fatalf("should have 1 tool message, got %d", len(toolMsgs))
	}
	if toolMsgs[0].Content != "" {
		t.Errorf("tool content should be empty string, got %q", toolMsgs[0].Content)
	}

	// 序列化后 content 字段必须存在（值为 ""）
	data, err := json.Marshal(toolMsgs[0])
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var asMap map[string]interface{}
	if err := json.Unmarshal(data, &asMap); err != nil {
		t.Fatalf("unmarshal tool message failed: %v", err)
	}
	if _, ok := asMap["content"]; !ok {
		t.Errorf("content field must survive serialization, got: %s", string(data))
	}
	if asMap["content"] != "" {
		t.Errorf("content should serialize to empty string, got: %v", asMap["content"])
	}
}

// TestAnthropicToOpenAIThinkingOnlyAssistantDropped 验证纯 thinking 的 assistant 消息
// 被直接丢弃，不产生空 content 消息（避免连续 assistant 角色触发上游 400）。
func TestAnthropicToOpenAIThinkingOnlyAssistantDropped(t *testing.T) {
	reqJSON := `{
		"model": "glm-4.6",
		"max_tokens": 1024,
		"messages": [
			{"role": "user", "content": "hi"},
			{"role": "assistant", "content": [
				{"type": "thinking", "thinking": "思考中", "signature": "sig"}
			]},
			{"role": "assistant", "content": [
				{"type": "tool_use", "id": "toolu_1", "name": "Read", "input": {"file_path": "/tmp/a"}}
			]},
			{"role": "user", "content": [
				{"type": "tool_result", "tool_use_id": "toolu_1", "content": "文件内容"}
			]},
			{"role": "user", "content": "第二句"}
		]
	}`

	var req AnthropicMessagesRequest
	if err := json.Unmarshal([]byte(reqJSON), &req); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	result := AnthropicToOpenAI(&req)

	var assistantCount int
	for _, m := range result.Messages {
		if m.Role == RoleAssistant {
			assistantCount++
		}
	}
	// 纯 thinking 消息应被丢弃，仅保留带 tool_use 的 assistant 消息
	if assistantCount != 1 {
		t.Fatalf("should have exactly 1 assistant message, got %d", assistantCount)
	}
}

// TestAnthropicToOpenAIAssistantToolCallsOmitEmptyContent 验证仅含 tool_use 的
// assistant 消息：content 字段省略、arguments 兜底 "{}"、不携带流式专用的 index 字段。
func TestAnthropicToOpenAIAssistantToolCallsOmitEmptyContent(t *testing.T) {
	reqJSON := `{
		"model": "glm-4.6",
		"max_tokens": 1024,
		"messages": [
			{"role": "user", "content": "hi"},
			{"role": "assistant", "content": [
				{"type": "tool_use", "id": "toolu_a", "name": "Bash", "input": {"command": "ls"}},
				{"type": "tool_use", "id": "toolu_b", "name": "NoInputTool"}
			]}
		]
	}`

	var req AnthropicMessagesRequest
	if err := json.Unmarshal([]byte(reqJSON), &req); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	result := AnthropicToOpenAI(&req)

	var assistantMsg *ChatMessage
	for i := range result.Messages {
		if result.Messages[i].Role == RoleAssistant {
			assistantMsg = &result.Messages[i]
			break
		}
	}
	if assistantMsg == nil {
		t.Fatal("should have assistant message")
	}
	if len(assistantMsg.ToolCalls) != 2 {
		t.Fatalf("should have 2 tool_calls, got %d", len(assistantMsg.ToolCalls))
	}
	// input 缺失时 arguments 兜底 "{}"
	if assistantMsg.ToolCalls[1].Function.Arguments != "{}" {
		t.Errorf("missing input should default arguments to {}, got %q",
			assistantMsg.ToolCalls[1].Function.Arguments)
	}

	data, err := json.Marshal(assistantMsg)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	payload := string(data)
	// 有 tool_calls 且无文本时，不应输出 "content":""
	if strings.Contains(payload, `"content":""`) {
		t.Errorf("assistant with only tool_calls should omit content, got: %s", payload)
	}
	// 请求体中的 tool_calls 不应携带流式专用的 index 字段
	if strings.Contains(payload, `"index"`) {
		t.Errorf("tool_calls in request body should not carry index field, got: %s", payload)
	}
}

// TestAnthropicToOpenAIEmptyUserBlocksDropped 验证 content 为空数组的 user 消息
// 不再生成空占位消息。
func TestAnthropicToOpenAIEmptyUserBlocksDropped(t *testing.T) {
	req := &AnthropicMessagesRequest{
		Model:     "glm-4.6",
		MaxTokens: 1024,
		Messages: []AnthropicMessage{
			{Role: RoleUser, Content: "第一句"},
			{Role: RoleAssistant, Content: "第一轮回复"},
			{Role: RoleUser, Content: []interface{}{}},
			{Role: RoleUser, Content: "第二句"},
		},
	}

	result := AnthropicToOpenAI(req)

	// user + assistant + user = 3 条，空数组消息被丢弃
	if len(result.Messages) != 3 {
		t.Fatalf("should have 3 messages, got %d: %+v", len(result.Messages), result.Messages)
	}
	if result.Messages[2].Content != "第二句" {
		t.Errorf("last message should be 第二句, got %v", result.Messages[2].Content)
	}
}

// ══════════════════════════════════
// Helper functions
// ══════════════════════════════════

func intPtr(i int) *int { return &i }
