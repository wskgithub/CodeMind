package llm

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// ══════════════════════════════════
// Anthropic Messages API - type serialization tests
// ══════════════════════════════════

// TestAnthropicRequestSerialization verifies request body contains all key fields.
func TestAnthropicRequestSerialization(t *testing.T) {
	temp := 0.7
	topP := 0.9
	topK := 40
	parallel := true

	req := AnthropicMessagesRequest{
		Model: "claude-sonnet-4-20250514",
		Messages: []AnthropicMessage{
			{Role: "user", Content: "Hello"},
		},
		System:          "You are helpful.",
		MaxTokens:       4096,
		Stream:          true,
		Temperature:     &temp,
		TopP:            &topP,
		TopK:            &topK,
		StopSequences:   []string{"\n\nHuman:"},
		ParallelToolUse: &parallel,
		Tools: []AnthropicTool{{
			Name:        "get_weather",
			Description: "get weather",
			InputSchema: map[string]interface{}{"type": "object"},
		}},
		ToolChoice: AnthropicToolChoice{Type: "auto"},
		Thinking:   &AnthropicThinking{Type: "enabled", BudgetTokens: 10000},
		Metadata:   &AnthropicMetadata{UserID: "user-123"},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("serialization failed: %v", err)
	}

	var raw map[string]interface{}
	json.Unmarshal(data, &raw)

	checks := []string{
		"model", "messages", "system", "max_tokens", "stream",
		"temperature", "top_p", "top_k", "stop_sequences", "tools",
		"tool_choice", "thinking", "metadata", "parallel_tool_use",
	}
	for _, key := range checks {
		if _, ok := raw[key]; !ok {
			t.Errorf("serialized result missing field: %s", key)
		}
	}

	// Verify thinking structure
	thinking, _ := raw["thinking"].(map[string]interface{})
	if thinking["type"] != "enabled" {
		t.Errorf("incorrect thinking.type: %v", thinking["type"])
	}
	if thinking["budget_tokens"].(float64) != 10000 {
		t.Errorf("incorrect thinking.budget_tokens: %v", thinking["budget_tokens"])
	}
}

// TestAnthropicResponseDeserialization verifies full response body deserialization.
func TestAnthropicResponseDeserialization(t *testing.T) {
	respJSON := `{
		"id": "msg_01abc",
		"type": "message",
		"role": "assistant",
		"content": [
			{"type": "text", "text": "Hello!"},
			{"type": "tool_use", "id": "toolu_01xyz", "name": "get_weather", "input": {"city": "Beijing"}}
		],
		"model": "claude-sonnet-4-20250514",
		"stop_reason": "tool_use",
		"stop_sequence": null,
		"usage": {
			"input_tokens": 100,
			"output_tokens": 50,
			"cache_creation_input_tokens": 10,
			"cache_read_input_tokens": 5
		}
	}`

	var resp AnthropicMessagesResponse
	if err := json.Unmarshal([]byte(respJSON), &resp); err != nil {
		t.Fatalf("deserialization failed: %v", err)
	}

	if resp.ID != "msg_01abc" {
		t.Errorf("incorrect ID: %s", resp.ID)
	}
	if resp.Type != "message" {
		t.Errorf("incorrect Type: %s", resp.Type)
	}
	if resp.Model != "claude-sonnet-4-20250514" {
		t.Errorf("incorrect Model: %s", resp.Model)
	}
	if len(resp.Content) != 2 {
		t.Fatalf("incorrect Content count: %d", len(resp.Content))
	}

	// Verify text block
	if resp.Content[0].Type != "text" || resp.Content[0].Text != "Hello!" {
		t.Error("incorrect first content block")
	}

	// Verify tool_use block
	tc := resp.Content[1]
	if tc.Type != "tool_use" || tc.ID != "toolu_01xyz" || tc.Name != "get_weather" {
		t.Errorf("incorrect tool_use block: type=%s, id=%s, name=%s", tc.Type, tc.ID, tc.Name)
	}

	// Verify stop_reason
	if resp.StopReason == nil || *resp.StopReason != "tool_use" {
		t.Error("incorrect stop_reason")
	}

	// Verify usage
	if resp.Usage == nil {
		t.Fatal("Usage should not be nil")
	}
	if resp.Usage.InputTokens != 100 || resp.Usage.OutputTokens != 50 {
		t.Errorf("incorrect token count: input=%d, output=%d", resp.Usage.InputTokens, resp.Usage.OutputTokens)
	}
	if resp.Usage.CacheCreationInputTokens != 10 || resp.Usage.CacheReadInputTokens != 5 {
		t.Error("incorrect cache token count")
	}
}

// TestAnthropicContentBlockThinking verifies thinking content block serialization.
func TestAnthropicContentBlockThinking(t *testing.T) {
	block := AnthropicContentBlock{
		Type:      "thinking",
		Thinking:  "Let me analyze this step by step...",
		Signature: "EqQBCgIYAhIM...",
	}

	data, err := json.Marshal(block)
	if err != nil {
		t.Fatalf("serialization failed: %v", err)
	}

	var raw map[string]interface{}
	json.Unmarshal(data, &raw)

	if raw["type"] != "thinking" {
		t.Errorf("incorrect type: %v", raw["type"])
	}
	if raw["thinking"] != "Let me analyze this step by step..." {
		t.Error("incorrect thinking content")
	}
	if raw["signature"] != "EqQBCgIYAhIM..." {
		t.Error("incorrect signature")
	}
}

// TestAnthropicUsageToUsage verifies Anthropic Usage to common Usage conversion.
func TestAnthropicUsageToUsage(t *testing.T) {
	au := &AnthropicUsage{
		InputTokens:              100,
		OutputTokens:             50,
		CacheCreationInputTokens: 10,
		CacheReadInputTokens:     5,
	}

	usage := au.ToUsage()
	if usage.PromptTokens != 100 {
		t.Errorf("incorrect PromptTokens: %d", usage.PromptTokens)
	}
	if usage.CompletionTokens != 50 {
		t.Errorf("incorrect CompletionTokens: %d", usage.CompletionTokens)
	}
	if usage.TotalTokens != 150 {
		t.Errorf("incorrect TotalTokens: %d", usage.TotalTokens)
	}

	// nil safety
	var nilUsage *AnthropicUsage
	if nilUsage.ToUsage() != nil {
		t.Error("nil AnthropicUsage.ToUsage() should return nil")
	}
}

// ══════════════════════════════════
// Anthropic Client - request/response tests
// ══════════════════════════════════

// TestAnthropicClientMessages tests non-streaming messages call.
func TestAnthropicClientMessages(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and path
		if r.Method != "POST" {
			t.Errorf("expected POST, got: %s", r.Method)
		}
		if r.URL.Path != "/v1/messages" {
			t.Errorf("expected path /v1/messages, got: %s", r.URL.Path)
		}

		// Verify Anthropic-specific request headers
		if r.Header.Get("anthropic-version") != AnthropicAPIVersion {
			t.Errorf("incorrect anthropic-version header: %s", r.Header.Get("anthropic-version"))
		}
		if r.Header.Get("x-api-key") != "test-anthropic-key" {
			t.Errorf("incorrect x-api-key header: %s", r.Header.Get("x-api-key"))
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Error("missing Content-Type header")
		}

		// Verify request body
		body, _ := io.ReadAll(r.Body)
		var req AnthropicMessagesRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Errorf("failed to parse request body: %v", err)
		}
		if req.Model != "claude-sonnet-4-20250514" {
			t.Errorf("incorrect model: %s", req.Model)
		}
		if req.Stream {
			t.Error("stream should be false for non-streaming request")
		}

		stopReason := "end_turn"
		resp := AnthropicMessagesResponse{
			ID:         "msg_test",
			Type:       "message",
			Role:       "assistant",
			Content:    []AnthropicContentBlock{{Type: "text", Text: "Hello!"}},
			Model:      "claude-sonnet-4-20250514",
			StopReason: &stopReason,
			Usage:      &AnthropicUsage{InputTokens: 10, OutputTokens: 5},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewAnthropicClient(server.URL, "test-anthropic-key", 30, 60)
	resp, err := client.Messages(&AnthropicMessagesRequest{
		Model:     "claude-sonnet-4-20250514",
		Messages:  []AnthropicMessage{{Role: "user", Content: "Hi"}},
		MaxTokens: 1024,
	})
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.ID != "msg_test" {
		t.Errorf("incorrect response ID: %s", resp.ID)
	}
	if len(resp.Content) != 1 || resp.Content[0].Text != "Hello!" {
		t.Error("incorrect response content")
	}
	if resp.Usage == nil || resp.Usage.InputTokens != 10 {
		t.Error("incorrect Usage")
	}
}

// TestAnthropicClientMessagesRaw tests raw request body passthrough.
func TestAnthropicClientMessagesRaw(t *testing.T) {
	var receivedBody []byte
	var receivedHeaders http.Header

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders = r.Header
		receivedBody, _ = io.ReadAll(r.Body)

		w.Write([]byte(`{
			"id": "msg_raw",
			"type": "message",
			"role": "assistant",
			"content": [{"type": "text", "text": "Hello!"}],
			"model": "claude-sonnet-4-20250514",
			"stop_reason": "end_turn",
			"usage": {"input_tokens": 20, "output_tokens": 10}
		}`))
	}))
	defer server.Close()

	client := NewAnthropicClient(server.URL, "test-key", 30, 60)

	// Construct raw request with thinking and other extra fields to verify passthrough preserves them
	rawReq := `{
		"model": "claude-sonnet-4-20250514",
		"messages": [{"role": "user", "content": "Hi"}],
		"max_tokens": 4096,
		"thinking": {"type": "enabled", "budget_tokens": 10000},
		"parallel_tool_use": true
	}`
	extraHeaders := map[string]string{
		"anthropic-beta": "extended-thinking-2025-04-11",
	}

	respBytes, usage, err := client.MessagesRaw([]byte(rawReq), extraHeaders)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	// Verify raw request body is fully forwarded (including thinking field)
	if !strings.Contains(string(receivedBody), `"thinking"`) {
		t.Error("raw passthrough should preserve thinking field")
	}
	if !strings.Contains(string(receivedBody), `"parallel_tool_use"`) {
		t.Error("raw passthrough should preserve parallel_tool_use field")
	}

	// Verify extra request headers are forwarded
	if receivedHeaders.Get("anthropic-beta") != "extended-thinking-2025-04-11" {
		t.Errorf("anthropic-beta header not forwarded: %s", receivedHeaders.Get("anthropic-beta"))
	}

	// Verify usage extraction
	if usage == nil {
		t.Fatal("Usage should not be nil")
	}
	if usage.InputTokens != 20 || usage.OutputTokens != 10 {
		t.Errorf("incorrect Usage: input=%d, output=%d", usage.InputTokens, usage.OutputTokens)
	}

	// Verify raw response is complete
	if !strings.Contains(string(respBytes), "msg_raw") {
		t.Error("raw response incomplete")
	}
}

// TestAnthropicClientMessagesStreamRaw tests raw request body streaming passthrough.
func TestAnthropicClientMessagesStreamRaw(t *testing.T) {
	sseData := `event: message_start
data: {"type":"message_start","message":{"id":"msg_stream","type":"message","role":"assistant","content":[],"model":"claude-sonnet-4-20250514","stop_reason":null,"usage":{"input_tokens":25,"output_tokens":1}}}

event: content_block_start
data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Hello"}}

event: message_stop
data: {"type":"message_stop"}

`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(sseData))
	}))
	defer server.Close()

	client := NewAnthropicClient(server.URL, "test-key", 30, 60)
	body, err := client.MessagesStreamRaw(
		[]byte(`{"model":"claude-sonnet-4-20250514","messages":[{"role":"user","content":"Hi"}],"max_tokens":1024,"stream":true}`),
		nil,
	)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer body.Close()

	data, err := io.ReadAll(body)
	if err != nil {
		t.Fatalf("failed to read stream: %v", err)
	}
	if !strings.Contains(string(data), "message_start") {
		t.Error("stream should contain message_start event")
	}
}

// TestAnthropicClientErrorHandling tests error handling.
func TestAnthropicClientErrorHandling(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		expectCode int
	}{
		{"429 → 529 overloaded", http.StatusTooManyRequests, 529},
		{"500 → 502 server error", http.StatusInternalServerError, 502},
		{"400 → 400 pass through", http.StatusBadRequest, 400},
		{"401 → 401 pass through", http.StatusUnauthorized, 401},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(`{"type":"error","error":{"type":"api_error","message":"test error"}}`))
			}))
			defer server.Close()

			client := NewAnthropicClient(server.URL, "test-key", 30, 60)
			_, err := client.Messages(&AnthropicMessagesRequest{
				Model:     "claude-sonnet-4-20250514",
				Messages:  []AnthropicMessage{{Role: "user", Content: "Hi"}},
				MaxTokens: 1024,
			})

			if err == nil {
				t.Fatal("should return error")
			}
			llmErr, ok := err.(*Error)
			if !ok {
				t.Fatalf("should return Error type, got: %T", err)
			}
			if llmErr.StatusCode != tt.expectCode {
				t.Errorf("incorrect status code mapping: expected %d, got %d", tt.expectCode, llmErr.StatusCode)
			}
		})
	}
}

// ══════════════════════════════════
// Anthropic SSE stream reader tests
// ══════════════════════════════════

// TestAnthropicStreamReaderBasic tests basic SSE event reading.
func TestAnthropicStreamReaderBasic(t *testing.T) {
	sseData := `event: message_start
data: {"type":"message_start","message":{"id":"msg_1","type":"message","role":"assistant","content":[],"model":"claude-sonnet-4-20250514","stop_reason":null,"usage":{"input_tokens":25,"output_tokens":1}}}

event: content_block_start
data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}

event: ping
data: {"type":"ping"}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Hello"}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"!"}}

event: content_block_stop
data: {"type":"content_block_stop","index":0}

event: message_delta
data: {"type":"message_delta","delta":{"stop_reason":"end_turn","stop_sequence":null},"usage":{"output_tokens":15}}

event: message_stop
data: {"type":"message_stop"}

`

	body := io.NopCloser(strings.NewReader(sseData))
	reader := NewAnthropicStreamReader(body)
	defer reader.Close()

	expectedEvents := []string{
		AnthropicEventMessageStart,
		AnthropicEventContentBlockStart,
		AnthropicEventPing,
		AnthropicEventContentBlockDelta,
		AnthropicEventContentBlockDelta,
		AnthropicEventContentBlockStop,
		AnthropicEventMessageDelta,
		AnthropicEventMessageStop,
	}

	for i, expected := range expectedEvents {
		eventType, rawLines, _, err := reader.ReadEvent()
		if err != nil {
			t.Fatalf("failed to read event %d: %v", i, err)
		}
		if eventType != expected {
			t.Errorf("incorrect event %d type: expected %s, got %s", i, expected, eventType)
		}
		if rawLines == "" {
			t.Errorf("event %d raw text should not be empty", i)
		}
	}

	// Should be marked done after message_stop
	if !reader.IsDone() {
		t.Error("reader should be marked as done after message_stop")
	}

	// Reading again should return EOF
	_, _, _, err := reader.ReadEvent()
	if err != io.EOF {
		t.Errorf("should return io.EOF, got: %v", err)
	}
}

// TestAnthropicStreamReaderToolUse tests tool call streaming events.
func TestAnthropicStreamReaderToolUse(t *testing.T) {
	sseData := `event: message_start
data: {"type":"message_start","message":{"id":"msg_tool","type":"message","role":"assistant","content":[],"model":"claude-sonnet-4-20250514","stop_reason":null,"usage":{"input_tokens":50,"output_tokens":2}}}

event: content_block_start
data: {"type":"content_block_start","index":0,"content_block":{"type":"tool_use","id":"toolu_01","name":"get_weather","input":{}}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"input_json_delta","partial_json":"{\"city\":"}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"input_json_delta","partial_json":" \"Beijing\"}"}}

event: content_block_stop
data: {"type":"content_block_stop","index":0}

event: message_delta
data: {"type":"message_delta","delta":{"stop_reason":"tool_use","stop_sequence":null},"usage":{"output_tokens":30}}

event: message_stop
data: {"type":"message_stop"}

`

	body := io.NopCloser(strings.NewReader(sseData))
	reader := NewAnthropicStreamReader(body)
	defer reader.Close()

	// message_start
	_, _, _, _ = reader.ReadEvent()

	// content_block_start (tool_use)
	eventType, _, event, err := reader.ReadEvent()
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if eventType != AnthropicEventContentBlockStart {
		t.Errorf("incorrect event type: %s", eventType)
	}
	if event.ContentBlock == nil || event.ContentBlock.Type != "tool_use" {
		t.Error("content_block should be tool_use type")
	}
	if event.ContentBlock.Name != "get_weather" {
		t.Errorf("incorrect tool name: %s", event.ContentBlock.Name)
	}

	// content_block_delta (input_json_delta)
	eventType, _, event, err = reader.ReadEvent()
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if eventType != AnthropicEventContentBlockDelta {
		t.Errorf("incorrect event type: %s", eventType)
	}
	if event.Delta == nil || event.Delta.Type != "input_json_delta" {
		t.Error("delta should be input_json_delta type")
	}
	if event.Delta.PartialJSON != `{"city":` {
		t.Errorf("incorrect partial_json: %s", event.Delta.PartialJSON)
	}
}

// TestAnthropicStreamReaderThinking tests extended thinking streaming events.
func TestAnthropicStreamReaderThinking(t *testing.T) {
	sseData := `event: message_start
data: {"type":"message_start","message":{"id":"msg_think","type":"message","role":"assistant","content":[],"model":"claude-sonnet-4-20250514","stop_reason":null}}

event: content_block_start
data: {"type":"content_block_start","index":0,"content_block":{"type":"thinking","thinking":""}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"thinking_delta","thinking":"Let me think about this..."}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"signature_delta","signature":"EqQBCgIYAhIM..."}}

event: content_block_stop
data: {"type":"content_block_stop","index":0}

event: content_block_start
data: {"type":"content_block_start","index":1,"content_block":{"type":"text","text":""}}

event: content_block_delta
data: {"type":"content_block_delta","index":1,"delta":{"type":"text_delta","text":"The answer is 42."}}

event: content_block_stop
data: {"type":"content_block_stop","index":1}

event: message_delta
data: {"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"output_tokens":100}}

event: message_stop
data: {"type":"message_stop"}

`

	body := io.NopCloser(strings.NewReader(sseData))
	reader := NewAnthropicStreamReader(body)
	defer reader.Close()

	// message_start
	_, _, _, _ = reader.ReadEvent()

	// content_block_start (thinking)
	eventType, _, event, _ := reader.ReadEvent()
	if eventType != AnthropicEventContentBlockStart {
		t.Errorf("incorrect event type: %s", eventType)
	}
	if event.ContentBlock == nil || event.ContentBlock.Type != "thinking" {
		t.Error("content_block should be thinking type")
	}

	// thinking_delta
	eventType, _, event, _ = reader.ReadEvent()
	if eventType != AnthropicEventContentBlockDelta {
		t.Errorf("incorrect event type: %s", eventType)
	}
	if event.Delta == nil || event.Delta.Type != "thinking_delta" {
		t.Error("delta should be thinking_delta type")
	}
	if event.Delta.Thinking != "Let me think about this..." {
		t.Errorf("incorrect thinking content: %s", event.Delta.Thinking)
	}

	// signature_delta
	_, _, event, _ = reader.ReadEvent()
	if event.Delta == nil || event.Delta.Type != "signature_delta" {
		t.Error("delta should be signature_delta type")
	}
	if event.Delta.Signature != "EqQBCgIYAhIM..." {
		t.Errorf("incorrect signature: %s", event.Delta.Signature)
	}

	// content_block_stop
	reader.ReadEvent()

	// content_block_start (text)
	reader.ReadEvent()

	// text_delta
	_, _, event, _ = reader.ReadEvent()
	if event.Delta.Text != "The answer is 42." {
		t.Errorf("incorrect text: %s", event.Delta.Text)
	}

	// content_block_stop
	reader.ReadEvent()

	// message_delta (with usage)
	eventType, _, event, _ = reader.ReadEvent()
	if eventType != AnthropicEventMessageDelta {
		t.Errorf("incorrect event type: %s", eventType)
	}
	if event.Usage == nil || event.Usage.OutputTokens != 100 {
		t.Error("incorrect usage in message_delta")
	}
}

// TestAnthropicStreamReaderEmpty tests empty stream.
func TestAnthropicStreamReaderEmpty(t *testing.T) {
	body := io.NopCloser(strings.NewReader(""))
	reader := NewAnthropicStreamReader(body)
	defer reader.Close()

	_, _, _, err := reader.ReadEvent()
	if err != io.EOF {
		t.Errorf("empty stream should return io.EOF, got: %v", err)
	}
}

// ══════════════════════════════════
// Anthropic error response tests
// ══════════════════════════════════

// TestAnthropicErrorResponseSerialization verifies error response format.
func TestAnthropicErrorResponseSerialization(t *testing.T) {
	errResp := AnthropicErrorResponse{
		Type: "error",
	}
	errResp.Error.Type = "invalid_request_error"
	errResp.Error.Message = "messages: Required"

	data, _ := json.Marshal(errResp)
	s := string(data)

	if !strings.Contains(s, `"type":"error"`) {
		t.Error("error response missing type:error")
	}
	if !strings.Contains(s, `"invalid_request_error"`) {
		t.Error("error response missing error type")
	}
	if !strings.Contains(s, `messages: Required`) {
		t.Error("error response missing error message")
	}
}
