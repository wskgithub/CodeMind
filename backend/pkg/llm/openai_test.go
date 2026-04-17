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
// OpenAI Chat Completions 协议测试
// ══════════════════════════════════

// TestChatCompletionRequestSerialization 验证 ChatCompletionRequest 序列化包含所有关键字段
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
				Description: "获取天气",
				Parameters:  map[string]interface{}{"type": "object"},
			},
		}},
		ToolChoice:     "auto",
		ResponseFormat: &ResponseFormat{Type: "json_object"},
		StreamOptions:  &StreamOptions{IncludeUsage: boolPtr(true)},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("序列化失败: %v", err)
	}

	var raw map[string]interface{}
	json.Unmarshal(data, &raw)

	checks := []string{"model", "messages", "stream", "temperature", "top_p",
		"max_tokens", "seed", "parallel_tool_calls", "tools", "tool_choice",
		"response_format", "stream_options"}
	for _, key := range checks {
		if _, ok := raw[key]; !ok {
			t.Errorf("序列化结果缺少字段: %s", key)
		}
	}
}

// TestChatCompletionResponseDeserialization 验证完整响应体反序列化
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
		t.Fatalf("反序列化失败: %v", err)
	}

	if resp.ID != "chatcmpl-abc123" {
		t.Errorf("ID 不正确: %s", resp.ID)
	}
	if resp.Model != "gpt-4" {
		t.Errorf("Model 不正确: %s", resp.Model)
	}
	if len(resp.Choices) != 1 {
		t.Fatalf("Choices 数量不正确: %d", len(resp.Choices))
	}

	choice := resp.Choices[0]
	if choice.Message.ContentString() != "Hello!" {
		t.Errorf("Content 不正确: %s", choice.Message.ContentString())
	}
	if len(choice.Message.ToolCalls) != 1 {
		t.Fatalf("ToolCalls 数量不正确: %d", len(choice.Message.ToolCalls))
	}
	tc := choice.Message.ToolCalls[0]
	if tc.ID != "call_123" || tc.Function.Name != "get_weather" {
		t.Errorf("ToolCall 不正确: id=%s, name=%s", tc.ID, tc.Function.Name)
	}
	if *choice.FinishReason != "tool_calls" {
		t.Errorf("FinishReason 不正确: %s", *choice.FinishReason)
	}

	if resp.Usage == nil {
		t.Fatal("Usage 不应为 nil")
	}
	if resp.Usage.TotalTokens != 80 {
		t.Errorf("TotalTokens 不正确: %d", resp.Usage.TotalTokens)
	}
	if resp.Usage.CompletionTokensDetails == nil || resp.Usage.CompletionTokensDetails.ReasoningTokens != 10 {
		t.Error("CompletionTokensDetails 不正确")
	}
}

// TestChatCompletionWithToolCalls 测试包含工具调用的完整请求→响应流程
func TestChatCompletionWithToolCalls(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证请求体包含 tools
		body, _ := io.ReadAll(r.Body)
		var req ChatCompletionRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Errorf("无法解析请求: %v", err)
		}
		if len(req.Tools) == 0 {
			t.Error("请求中缺少 tools")
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
		Messages: []ChatMessage{{Role: "user", Content: "北京天气"}},
		Tools: []Tool{{
			Type:     "function",
			Function: ToolFunction{Name: "get_weather", Parameters: map[string]interface{}{"type": "object"}},
		}},
		ToolChoice: "auto",
	})
	if err != nil {
		t.Fatalf("请求失败: %v", err)
	}

	if len(resp.Choices) == 0 || len(resp.Choices[0].Message.ToolCalls) == 0 {
		t.Fatal("响应中缺少 tool_calls")
	}
	if resp.Choices[0].Message.ToolCalls[0].Function.Name != "get_weather" {
		t.Error("tool_call 名称不正确")
	}
}

// TestMultimodalContentMessage 验证多模态消息 Content 的 interface{} 类型处理
func TestMultimodalContentMessage(t *testing.T) {
	msgJSON := `{
		"role": "user",
		"content": [
			{"type": "text", "text": "这是什么图片？"},
			{"type": "image_url", "image_url": {"url": "https://example.com/img.png"}}
		]
	}`

	var msg ChatMessage
	if err := json.Unmarshal([]byte(msgJSON), &msg); err != nil {
		t.Fatalf("反序列化失败: %v", err)
	}

	if msg.Role != "user" {
		t.Errorf("角色不正确: %s", msg.Role)
	}
	if msg.ContentString() != "这是什么图片？" {
		t.Errorf("ContentString 应只返回文本部分: %s", msg.ContentString())
	}
}

// ══════════════════════════════════
// OpenAI Completions 协议测试
// ══════════════════════════════════

// TestCompletionRequestResponse 测试 Completions API 完整请求→响应
func TestCompletionRequestResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/completions" {
			t.Errorf("路径不正确: %s", r.URL.Path)
		}

		finishReason := "stop"
		resp := CompletionResponse{
			ID:    "cmpl-123",
			Model: "gpt-3.5-turbo-instruct",
			Choices: []CompletionChoice{{
				Index:        0,
				Text:         "世界！",
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
		Prompt: "你好",
	})
	if err != nil {
		t.Fatalf("请求失败: %v", err)
	}

	if resp.Choices[0].Text != "世界！" {
		t.Errorf("补全文本不正确: %s", resp.Choices[0].Text)
	}
	if resp.Usage.TotalTokens != 8 {
		t.Errorf("TotalTokens 不正确: %d", resp.Usage.TotalTokens)
	}
}

// ══════════════════════════════════
// OpenAI Embeddings 协议测试
// ══════════════════════════════════

// TestEmbeddingRaw 测试 Embeddings 原始透传请求→响应
func TestEmbeddingRaw(t *testing.T) {
	expectedResp := `{
		"object": "list",
		"data": [{"object": "embedding", "embedding": [0.1, 0.2, 0.3], "index": 0}],
		"model": "text-embedding-ada-002",
		"usage": {"prompt_tokens": 5, "total_tokens": 5}
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/embeddings" {
			t.Errorf("路径不正确: %s", r.URL.Path)
		}
		w.Write([]byte(expectedResp))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key", 30, 60)
	rawReq := []byte(`{"input":"hello","model":"text-embedding-ada-002"}`)
	rawResp, usage, err := client.EmbeddingRaw(rawReq)
	if err != nil {
		t.Fatalf("请求失败: %v", err)
	}

	if usage == nil {
		t.Fatal("Usage 不应为 nil")
	}
	if usage.PromptTokens != 5 {
		t.Errorf("PromptTokens 不正确: %d", usage.PromptTokens)
	}

	var resp EmbeddingResponse
	if err := json.Unmarshal(rawResp, &resp); err != nil {
		t.Fatalf("无法解析响应: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Errorf("Embedding 数据数量不正确: %d", len(resp.Data))
	}
}

// ══════════════════════════════════
// OpenAI Responses API 协议测试
// ══════════════════════════════════

// TestResponsesRaw 测试 Responses API 非流式透传
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
			t.Errorf("路径不正确: %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("方法不正确: %s", r.Method)
		}
		w.Write([]byte(respJSON))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key", 30, 60)
	rawResp, usage, err := client.ResponsesRaw([]byte(`{"model":"gpt-4o","input":"Hello"}`))
	if err != nil {
		t.Fatalf("请求失败: %v", err)
	}

	if usage == nil {
		t.Fatal("Usage 不应为 nil")
	}
	if usage.TotalTokens != 15 {
		t.Errorf("TotalTokens 不正确: %d", usage.TotalTokens)
	}
	if !strings.Contains(string(rawResp), "resp_123") {
		t.Error("原始响应应包含 response ID")
	}
}

// TestResponsesStreamReader 测试 Responses API 流式 SSE 读取
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

	// 第一个事件：response.created
	eventType, _, _, err := reader.ReadEvent()
	if err != nil {
		t.Fatalf("读取事件失败: %v", err)
	}
	if eventType != "response.created" {
		t.Errorf("事件类型不正确: %s", eventType)
	}

	// 第二个事件：text delta
	eventType, _, _, err = reader.ReadEvent()
	if err != nil {
		t.Fatalf("读取事件失败: %v", err)
	}
	if eventType != "response.output_text.delta" {
		t.Errorf("事件类型不正确: %s", eventType)
	}

	// 第三个事件：text delta
	_, _, _, err = reader.ReadEvent()
	if err != nil {
		t.Fatalf("读取事件失败: %v", err)
	}

	// 第四个事件：response.completed（含 usage）
	eventType, _, payload, err := reader.ReadEvent()
	if err != nil {
		t.Fatalf("读取事件失败: %v", err)
	}
	if eventType != "response.completed" {
		t.Errorf("事件类型不正确: %s", eventType)
	}

	usage := ExtractUsageFromResponsesEvent(payload)
	if usage == nil {
		t.Fatal("应能从 response.completed 提取 usage")
	}
	if usage.TotalTokens != 15 {
		t.Errorf("TotalTokens 不正确: %d", usage.TotalTokens)
	}
}

// ══════════════════════════════════
// OpenAI Raw Proxy 工具测试
// ══════════════════════════════════

// TestEnsureStreamOptions 验证自动注入 stream_options
func TestEnsureStreamOptions(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(t *testing.T, result []byte)
	}{
		{
			name:  "无 stream_options 时应新增",
			input: `{"model":"gpt-4","stream":true}`,
			check: func(t *testing.T, result []byte) {
				if !strings.Contains(string(result), `"include_usage":true`) {
					t.Error("应包含 include_usage:true")
				}
			},
		},
		{
			name:  "已有 include_usage=true 时不修改",
			input: `{"model":"gpt-4","stream_options":{"include_usage":true}}`,
			check: func(t *testing.T, result []byte) {
				if !strings.Contains(string(result), `"include_usage":true`) {
					t.Error("应保留 include_usage:true")
				}
			},
		},
		{
			name:  "已有 stream_options 但缺少 include_usage 时应补充",
			input: `{"model":"gpt-4","stream_options":{"other":1}}`,
			check: func(t *testing.T, result []byte) {
				if !strings.Contains(string(result), `"include_usage":true`) {
					t.Error("应补充 include_usage:true")
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

// TestExtractUsageFromResponse 验证从原始响应中提取 usage
func TestExtractUsageFromResponse(t *testing.T) {
	resp := `{"id":"x","usage":{"prompt_tokens":10,"completion_tokens":20,"total_tokens":30}}`
	usage := ExtractUsageFromResponse([]byte(resp))
	if usage == nil {
		t.Fatal("Usage 不应为 nil")
	}
	if usage.PromptTokens != 10 || usage.CompletionTokens != 20 || usage.TotalTokens != 30 {
		t.Errorf("Usage 值不正确: %+v", usage)
	}

	// 无 usage 字段
	usage2 := ExtractUsageFromResponse([]byte(`{"id":"x"}`))
	if usage2 != nil {
		t.Error("无 usage 时应返回 nil")
	}
}

// TestChatCompletionRawPassthrough 测试原始请求体透传保留所有字段
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
		t.Fatalf("请求失败: %v", err)
	}

	// 验证原始请求中的自定义字段被保留
	if !strings.Contains(string(receivedBody), "custom_field") {
		t.Error("原始透传应保留所有字段（含自定义字段）")
	}
}

// ══════════════════════════════════
// OpenAI 错误响应测试
// ══════════════════════════════════

// TestErrorResponseSerialization 验证错误响应格式
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
		t.Error("错误响应序列化不正确")
	}
}

// ══════════════════════════════════
// 辅助函数
// ══════════════════════════════════

func boolPtr(b bool) *bool { return &b }
