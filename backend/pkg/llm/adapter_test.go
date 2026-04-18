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

	result := OpenAIToAnthropic(req)

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

	result := OpenAIToAnthropic(req)

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

	result := OpenAIToAnthropic(req)
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

	result := OpenAIToAnthropic(req)
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

	result := OpenAIToAnthropic(req)

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

			result := OpenAIToAnthropic(req)
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

	result := OpenAIToAnthropic(req)

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
	result := OpenAIToAnthropic(req)
	if len(result.StopSequences) != 1 || result.StopSequences[0] != "END" {
		t.Errorf("incorrect StopSequences: %v", result.StopSequences)
	}
}

// TestOpenAIToAnthropicFirstMessageMustBeUser tests that first message must be user role.
func TestOpenAIToAnthropicFirstMessageMustBeUser(t *testing.T) {
	req := &ChatCompletionRequest{
		Model: "gpt-4",
		Messages: []ChatMessage{
			{Role: "assistant", Content: "I'm ready."},
		},
	}

	result := OpenAIToAnthropic(req)
	if result.Messages[0].Role != "user" {
		t.Error("first Anthropic message must be user role")
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

// ══════════════════════════════════
// Stream format conversion: OpenAI to Anthropic
// ══════════════════════════════════

// TestOpenAIChunkToAnthropicEventsFirstChunk tests first chunk generates message_start.
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
		t.Error("first chunk should contain message_start event")
	}
	if !strings.Contains(result, "event: ping") {
		t.Error("first chunk should contain ping event")
	}
	if !strings.Contains(result, "text_delta") {
		t.Error("should contain text_delta event")
	}
}

// TestOpenAIChunkToAnthropicEventsFinish tests finish events.
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
		t.Error("should contain content_block_stop event")
	}
	if !strings.Contains(result, "message_delta") {
		t.Error("should contain message_delta event")
	}
	if !strings.Contains(result, `"stop_reason":"end_turn"`) {
		t.Error("stop_reason should map to end_turn")
	}
	if !strings.Contains(result, "message_stop") {
		t.Error("should contain message_stop event")
	}
}

// TestOpenAIChunkToAnthropicEventsToolCall tests tool call chunk conversion.
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
		t.Error("should contain tool_use content_block_start")
	}
	if !strings.Contains(result, "get_weather") {
		t.Error("should contain tool name")
	}

	// Argument increments
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
		t.Error("should contain input_json_delta event")
	}
}

// ══════════════════════════════════
// Helper functions
// ══════════════════════════════════

func intPtr(i int) *int { return &i }
