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
// OpenAI Chat Completions protocol tests
// ══════════════════════════════════

// TestChatCompletionRequestSerialization verifies ChatCompletionRequest serialization contains all key fields.
func TestChatCompletionRequestSerialization(t *testing.T) {
	temp := 0.7
	topP := 0.9
	maxTokens := 1024
	seed := int64(42)
	parallel := true

	req := ChatCompletionRequest{
		Model: "gpt-4",
		Messages: []ChatMessage{
			{Role: "system", Content: "You are helpful."},
			{Role: "user", Content: "Hello"},
		},
		Stream:            true,
		Temperature:       &temp,
		TopP:              &topP,
		MaxTokens:         &maxTokens,
		Seed:              &seed,
		ParallelToolCalls: &parallel,
		Tools: []Tool{{
			Type: "function",
			Function: ToolFunction{
				Name:        "get_weather",
				Description: "get weather",
				Parameters:  map[string]interface{}{"type": "object"},
			},
		}},
		ToolChoice:     "auto",
		ResponseFormat: &ResponseFormat{Type: "json_object"},
		StreamOptions:  &StreamOptions{IncludeUsage: boolPtr(true)},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("serialization failed: %v", err)
	}

	var raw map[string]interface{}
	json.Unmarshal(data, &raw)

	checks := []string{
		"model", "messages", "stream", "temperature", "top_p",
		"max_tokens", "seed", "parallel_tool_calls", "tools", "tool_choice",
		"response_format", "stream_options",
	}
	for _, key := range checks {
		if _, ok := raw[key]; !ok {
			t.Errorf("serialized result missing field: %s", key)
		}
	}
}

// TestChatCompletionResponseDeserialization verifies full response body deserialization.
func TestChatCompletionResponseDeserialization(t *testing.T) {
	respJSON := `{
		"id": "chatcmpl-abc123",
		"object": "chat.completion",
		"created": 1700000000,
		"model": "gpt-4",
		"choices": [{
			"index": 0,
			"message": {
				"role": "assistant",
				"content": "Hello!",
				"tool_calls": [{
					"id": "call_123",
					"type": "function",
					"function": {"name": "get_weather", "arguments": "{\"city\":\"Beijing\"}"}
				}]
			},
			"finish_reason": "tool_calls"
		}],
		"usage": {
			"prompt_tokens": 50,
			"completion_tokens": 30,
			"total_tokens": 80,
			"completion_tokens_details": {"reasoning_tokens": 10}
		},
		"system_fingerprint": "fp_abc123"
	}`

	var resp ChatCompletionResponse
	if err := json.Unmarshal([]byte(respJSON), &resp); err != nil {
		t.Fatalf("deserialization failed: %v", err)
	}

	if resp.ID != "chatcmpl-abc123" {
		t.Errorf("incorrect ID: %s", resp.ID)
	}
	if resp.Model != "gpt-4" {
		t.Errorf("incorrect Model: %s", resp.Model)
	}
	if len(resp.Choices) != 1 {
		t.Fatalf("incorrect Choices count: %d", len(resp.Choices))
	}

	choice := resp.Choices[0]
	if choice.Message.ContentString() != "Hello!" {
		t.Errorf("incorrect Content: %s", choice.Message.ContentString())
	}
	if len(choice.Message.ToolCalls) != 1 {
		t.Fatalf("incorrect ToolCalls count: %d", len(choice.Message.ToolCalls))
	}
	tc := choice.Message.ToolCalls[0]
	if tc.ID != "call_123" || tc.Function.Name != "get_weather" {
		t.Errorf("incorrect ToolCall: id=%s, name=%s", tc.ID, tc.Function.Name)
	}
	if *choice.FinishReason != "tool_calls" {
		t.Errorf("incorrect FinishReason: %s", *choice.FinishReason)
	}

	if resp.Usage == nil {
		t.Fatal("Usage should not be nil")
	}
	if resp.Usage.TotalTokens != 80 {
		t.Errorf("incorrect TotalTokens: %d", resp.Usage.TotalTokens)
	}
	if resp.Usage.CompletionTokensDetails == nil || resp.Usage.CompletionTokensDetails.ReasoningTokens != 10 {
		t.Error("incorrect CompletionTokensDetails")
	}
}

// TestChatCompletionWithToolCalls tests full request-response flow with tool calls.
func TestChatCompletionWithToolCalls(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request body contains tools
		body, _ := io.ReadAll(r.Body)
		var req ChatCompletionRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Errorf("failed to parse request: %v", err)
		}
		if len(req.Tools) == 0 {
			t.Error("request missing tools")
		}

		finishReason := "tool_calls"
		resp := ChatCompletionResponse{
			ID:    "chatcmpl-tool",
			Model: "gpt-4",
			Choices: []ChatChoice{{
				Index: 0,
				Message: &ChatMessage{
					Role: "assistant",
					ToolCalls: []ToolCall{{
						ID:   "call_abc",
						Type: "function",
						Function: ToolCallFunction{
							Name:      "get_weather",
							Arguments: `{"city":"Beijing"}`,
						},
					}},
				},
				FinishReason: &finishReason,
			}},
			Usage: &Usage{PromptTokens: 20, CompletionTokens: 15, TotalTokens: 35},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key", 30, 60)
	resp, err := client.ChatCompletion(&ChatCompletionRequest{
		Model:    "gpt-4",
		Messages: []ChatMessage{{Role: "user", Content: "Beijing weather"}},
		Tools: []Tool{{
			Type:     "function",
			Function: ToolFunction{Name: "get_weather", Parameters: map[string]interface{}{"type": "object"}},
		}},
		ToolChoice: "auto",
	})
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if len(resp.Choices) == 0 || len(resp.Choices[0].Message.ToolCalls) == 0 {
		t.Fatal("response missing tool_calls")
	}
	if resp.Choices[0].Message.ToolCalls[0].Function.Name != "get_weather" {
		t.Error("incorrect tool_call name")
	}
}

// TestMultimodalContentMessage verifies multimodal message Content interface{} type handling.
func TestMultimodalContentMessage(t *testing.T) {
	msgJSON := `{
		"role": "user",
		"content": [
			{"type": "text", "text": "What is this image?"},
			{"type": "image_url", "image_url": {"url": "https://example.com/img.png"}}
		]
	}`

	var msg ChatMessage
	if err := json.Unmarshal([]byte(msgJSON), &msg); err != nil {
		t.Fatalf("deserialization failed: %v", err)
	}

	if msg.Role != "user" {
		t.Errorf("incorrect role: %s", msg.Role)
	}
	if msg.ContentString() != "What is this image?" {
		t.Errorf("ContentString should return only text part: %s", msg.ContentString())
	}
}

// ══════════════════════════════════
// OpenAI Completions protocol tests
// ══════════════════════════════════

// TestCompletionRequestResponse tests Completions API full request-response.
func TestCompletionRequestResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/completions" {
			t.Errorf("incorrect path: %s", r.URL.Path)
		}

		finishReason := "stop"
		resp := CompletionResponse{
			ID:    "cmpl-123",
			Model: "gpt-3.5-turbo-instruct",
			Choices: []CompletionChoice{{
				Index:        0,
				Text:         "World!",
				FinishReason: &finishReason,
			}},
			Usage: &Usage{PromptTokens: 5, CompletionTokens: 3, TotalTokens: 8},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key", 30, 60)
	resp, err := client.Completion(&CompletionRequest{
		Model:  "gpt-3.5-turbo-instruct",
		Prompt: "Hello",
	})
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.Choices[0].Text != "World!" {
		t.Errorf("incorrect completion text: %s", resp.Choices[0].Text)
	}
	if resp.Usage.TotalTokens != 8 {
		t.Errorf("incorrect TotalTokens: %d", resp.Usage.TotalTokens)
	}
}

// ══════════════════════════════════
// OpenAI Embeddings protocol tests
// ══════════════════════════════════

// TestEmbeddingRaw tests Embeddings raw passthrough request-response.
func TestEmbeddingRaw(t *testing.T) {
	expectedResp := `{
		"object": "list",
		"data": [{"object": "embedding", "embedding": [0.1, 0.2, 0.3], "index": 0}],
		"model": "text-embedding-ada-002",
		"usage": {"prompt_tokens": 5, "total_tokens": 5}
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/embeddings" {
			t.Errorf("incorrect path: %s", r.URL.Path)
		}
		w.Write([]byte(expectedResp))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key", 30, 60)
	rawReq := []byte(`{"input":"hello","model":"text-embedding-ada-002"}`)
	rawResp, usage, err := client.EmbeddingRaw(rawReq)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if usage == nil {
		t.Fatal("Usage should not be nil")
	}
	if usage.PromptTokens != 5 {
		t.Errorf("incorrect PromptTokens: %d", usage.PromptTokens)
	}

	var resp EmbeddingResponse
	if err := json.Unmarshal(rawResp, &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Errorf("incorrect Embedding data count: %d", len(resp.Data))
	}
}

// ══════════════════════════════════
// OpenAI Responses API protocol tests
// ══════════════════════════════════

// TestResponsesRaw tests Responses API non-streaming passthrough.
func TestResponsesRaw(t *testing.T) {
	respJSON := `{
		"id": "resp_123",
		"object": "response",
		"model": "gpt-4o",
		"output": [{"type": "message", "content": [{"type": "output_text", "text": "Hello!"}]}],
		"usage": {"input_tokens": 10, "output_tokens": 5, "total_tokens": 15}
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/responses" {
			t.Errorf("incorrect path: %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("incorrect method: %s", r.Method)
		}
		w.Write([]byte(respJSON))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key", 30, 60)
	rawResp, usage, err := client.ResponsesRaw([]byte(`{"model":"gpt-4o","input":"Hello"}`))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if usage == nil {
		t.Fatal("Usage should not be nil")
	}
	if usage.TotalTokens != 15 {
		t.Errorf("incorrect TotalTokens: %d", usage.TotalTokens)
	}
	if !strings.Contains(string(rawResp), "resp_123") {
		t.Error("raw response should contain response ID")
	}
}

// TestResponsesStreamReader tests Responses API streaming SSE reader.
func TestResponsesStreamReader(t *testing.T) {
	sseData := `event: response.created
data: {"type":"response.created","response":{"id":"resp_001"}}

event: response.output_text.delta
data: {"type":"response.output_text.delta","delta":"Hello"}

event: response.output_text.delta
data: {"type":"response.output_text.delta","delta":" world"}

event: response.completed
data: {"type":"response.completed","response":{"id":"resp_001","usage":{"input_tokens":10,"output_tokens":5,"total_tokens":15}}}

`

	body := io.NopCloser(strings.NewReader(sseData))
	reader := NewResponsesStreamReader(body)
	defer reader.Close()

	// First event: response.created
	eventType, _, _, err := reader.ReadEvent()
	if err != nil {
		t.Fatalf("failed to read event: %v", err)
	}
	if eventType != "response.created" {
		t.Errorf("incorrect event type: %s", eventType)
	}

	// Second event: text delta
	eventType, _, _, err = reader.ReadEvent()
	if err != nil {
		t.Fatalf("failed to read event: %v", err)
	}
	if eventType != "response.output_text.delta" {
		t.Errorf("incorrect event type: %s", eventType)
	}

	// Third event: text delta
	_, _, _, err = reader.ReadEvent()
	if err != nil {
		t.Fatalf("failed to read event: %v", err)
	}

	// Fourth event: response.completed (with usage)
	eventType, _, payload, err := reader.ReadEvent()
	if err != nil {
		t.Fatalf("failed to read event: %v", err)
	}
	if eventType != "response.completed" {
		t.Errorf("incorrect event type: %s", eventType)
	}

	usage := ExtractUsageFromResponsesEvent(payload)
	if usage == nil {
		t.Fatal("should be able to extract usage from response.completed")
	}
	if usage.TotalTokens != 15 {
		t.Errorf("incorrect TotalTokens: %d", usage.TotalTokens)
	}
}

// ══════════════════════════════════
// OpenAI Raw Proxy utility tests
// ══════════════════════════════════

// TestEnsureStreamOptions verifies automatic stream_options injection.
func TestEnsureStreamOptions(t *testing.T) {
	tests := []struct {
		check func(t *testing.T, result []byte)
		name  string
		input string
	}{
		{
			name:  "should add when no stream_options",
			input: `{"model":"gpt-4","stream":true}`,
			check: func(t *testing.T, result []byte) {
				if !strings.Contains(string(result), `"include_usage":true`) {
					t.Error("should contain include_usage:true")
				}
			},
		},
		{
			name:  "should not modify when include_usage=true exists",
			input: `{"model":"gpt-4","stream_options":{"include_usage":true}}`,
			check: func(t *testing.T, result []byte) {
				if !strings.Contains(string(result), `"include_usage":true`) {
					t.Error("should preserve include_usage:true")
				}
			},
		},
		{
			name:  "should add include_usage when stream_options exists but lacks it",
			input: `{"model":"gpt-4","stream_options":{"other":1}}`,
			check: func(t *testing.T, result []byte) {
				if !strings.Contains(string(result), `"include_usage":true`) {
					t.Error("should add include_usage:true")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EnsureStreamOptions([]byte(tt.input))
			tt.check(t, result)
		})
	}
}

// TestExtractUsageFromResponse verifies usage extraction from raw response.
func TestExtractUsageFromResponse(t *testing.T) {
	resp := `{"id":"x","usage":{"prompt_tokens":10,"completion_tokens":20,"total_tokens":30}}`
	usage := ExtractUsageFromResponse([]byte(resp))
	if usage == nil {
		t.Fatal("Usage should not be nil")
	}
	if usage.PromptTokens != 10 || usage.CompletionTokens != 20 || usage.TotalTokens != 30 {
		t.Errorf("incorrect Usage values: %+v", usage)
	}

	// No usage field
	usage2 := ExtractUsageFromResponse([]byte(`{"id":"x"}`))
	if usage2 != nil {
		t.Error("should return nil when no usage")
	}
}

// TestChatCompletionRawPassthrough tests raw request body passthrough preserves all fields.
func TestChatCompletionRawPassthrough(t *testing.T) {
	var receivedBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedBody, _ = io.ReadAll(r.Body)
		w.Write([]byte(`{"id":"x","usage":{"prompt_tokens":1,"completion_tokens":2,"total_tokens":3}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key", 30, 60)
	rawReq := `{"model":"gpt-4","messages":[{"role":"user","content":"hi"}],"custom_field":"should_preserve","temperature":0.5}`
	_, _, err := client.ChatCompletionRawAll([]byte(rawReq))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	// Verify custom fields in raw request are preserved
	if !strings.Contains(string(receivedBody), "custom_field") {
		t.Error("raw passthrough should preserve all fields (including custom fields)")
	}
}

// ══════════════════════════════════
// OpenAI error response tests
// ══════════════════════════════════

// TestErrorResponseSerialization verifies error response format.
func TestErrorResponseSerialization(t *testing.T) {
	errResp := ErrorResponse{
		Error: ErrorDetail{
			Message: "Model not found",
			Type:    "invalid_request_error",
			Code:    "model_not_found",
		},
	}

	data, _ := json.Marshal(errResp)
	if !strings.Contains(string(data), "model_not_found") {
		t.Error("incorrect error response serialization")
	}
}

// ══════════════════════════════════
// Helper functions
// ══════════════════════════════════

func boolPtr(b bool) *bool { return &b }
