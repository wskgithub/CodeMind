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
// Anthropic Messages API — 类型序列化测试
// ══════════════════════════════════

// TestAnthropicRequestSerialization 验证请求体包含所有关键字段.
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
			Description: "获取天气",
			InputSchema: map[string]interface{}{"type": "object"},
		}},
		ToolChoice: AnthropicToolChoice{Type: "auto"},
		Thinking:   &AnthropicThinking{Type: "enabled", BudgetTokens: 10000},
		Metadata:   &AnthropicMetadata{UserID: "user-123"},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("序列化失败: %v", err)
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
			t.Errorf("序列化结果缺少字段: %s", key)
		}
	}

	// 验证 thinking 结构
	thinking, _ := raw["thinking"].(map[string]interface{})
	if thinking["type"] != "enabled" {
		t.Errorf("thinking.type 不正确: %v", thinking["type"])
	}
	if thinking["budget_tokens"].(float64) != 10000 {
		t.Errorf("thinking.budget_tokens 不正确: %v", thinking["budget_tokens"])
	}
}

// TestAnthropicResponseDeserialization 验证完整响应体反序列化.
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
		t.Fatalf("反序列化失败: %v", err)
	}

	if resp.ID != "msg_01abc" {
		t.Errorf("ID 不正确: %s", resp.ID)
	}
	if resp.Type != "message" {
		t.Errorf("Type 不正确: %s", resp.Type)
	}
	if resp.Model != "claude-sonnet-4-20250514" {
		t.Errorf("Model 不正确: %s", resp.Model)
	}
	if len(resp.Content) != 2 {
		t.Fatalf("Content 数量不正确: %d", len(resp.Content))
	}

	// 验证 text 块
	if resp.Content[0].Type != "text" || resp.Content[0].Text != "Hello!" {
		t.Error("第一个内容块不正确")
	}

	// 验证 tool_use 块
	tc := resp.Content[1]
	if tc.Type != "tool_use" || tc.ID != "toolu_01xyz" || tc.Name != "get_weather" {
		t.Errorf("tool_use 块不正确: type=%s, id=%s, name=%s", tc.Type, tc.ID, tc.Name)
	}

	// 验证 stop_reason
	if resp.StopReason == nil || *resp.StopReason != "tool_use" {
		t.Error("stop_reason 不正确")
	}

	// 验证 usage
	if resp.Usage == nil {
		t.Fatal("Usage 不应为 nil")
	}
	if resp.Usage.InputTokens != 100 || resp.Usage.OutputTokens != 50 {
		t.Errorf("Token 数量不正确: input=%d, output=%d", resp.Usage.InputTokens, resp.Usage.OutputTokens)
	}
	if resp.Usage.CacheCreationInputTokens != 10 || resp.Usage.CacheReadInputTokens != 5 {
		t.Error("缓存 token 数量不正确")
	}
}

// TestAnthropicContentBlockThinking 验证 thinking 内容块序列化.
func TestAnthropicContentBlockThinking(t *testing.T) {
	block := AnthropicContentBlock{
		Type:      "thinking",
		Thinking:  "Let me analyze this step by step...",
		Signature: "EqQBCgIYAhIM...",
	}

	data, err := json.Marshal(block)
	if err != nil {
		t.Fatalf("序列化失败: %v", err)
	}

	var raw map[string]interface{}
	json.Unmarshal(data, &raw)

	if raw["type"] != "thinking" {
		t.Errorf("type 不正确: %v", raw["type"])
	}
	if raw["thinking"] != "Let me analyze this step by step..." {
		t.Error("thinking 内容不正确")
	}
	if raw["signature"] != "EqQBCgIYAhIM..." {
		t.Error("signature 不正确")
	}
}

// TestAnthropicUsageToUsage 验证 Anthropic Usage 到通用 Usage 的转换.
func TestAnthropicUsageToUsage(t *testing.T) {
	au := &AnthropicUsage{
		InputTokens:              100,
		OutputTokens:             50,
		CacheCreationInputTokens: 10,
		CacheReadInputTokens:     5,
	}

	usage := au.ToUsage()
	if usage.PromptTokens != 100 {
		t.Errorf("PromptTokens 不正确: %d", usage.PromptTokens)
	}
	if usage.CompletionTokens != 50 {
		t.Errorf("CompletionTokens 不正确: %d", usage.CompletionTokens)
	}
	if usage.TotalTokens != 150 {
		t.Errorf("TotalTokens 不正确: %d", usage.TotalTokens)
	}

	// nil 安全
	var nilUsage *AnthropicUsage
	if nilUsage.ToUsage() != nil {
		t.Error("nil AnthropicUsage.ToUsage() 应返回 nil")
	}
}

// ══════════════════════════════════
// Anthropic Client — 请求/响应测试
// ══════════════════════════════════

// TestAnthropicClientMessages 测试非流式消息调用.
func TestAnthropicClientMessages(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证请求方法和路径
		if r.Method != "POST" {
			t.Errorf("预期 POST, 实际: %s", r.Method)
		}
		if r.URL.Path != "/v1/messages" {
			t.Errorf("预期路径 /v1/messages, 实际: %s", r.URL.Path)
		}

		// 验证 Anthropic 专用请求头
		if r.Header.Get("anthropic-version") != AnthropicAPIVersion {
			t.Errorf("anthropic-version 头不正确: %s", r.Header.Get("anthropic-version"))
		}
		if r.Header.Get("x-api-key") != "test-anthropic-key" {
			t.Errorf("x-api-key 头不正确: %s", r.Header.Get("x-api-key"))
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Error("缺少 Content-Type 头")
		}

		// 验证请求体
		body, _ := io.ReadAll(r.Body)
		var req AnthropicMessagesRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Errorf("请求体解析失败: %v", err)
		}
		if req.Model != "claude-sonnet-4-20250514" {
			t.Errorf("模型不正确: %s", req.Model)
		}
		if req.Stream {
			t.Error("非流式请求 stream 应为 false")
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
		t.Fatalf("请求失败: %v", err)
	}

	if resp.ID != "msg_test" {
		t.Errorf("响应 ID 不正确: %s", resp.ID)
	}
	if len(resp.Content) != 1 || resp.Content[0].Text != "Hello!" {
		t.Error("响应内容不正确")
	}
	if resp.Usage == nil || resp.Usage.InputTokens != 10 {
		t.Error("Usage 不正确")
	}
}

// TestAnthropicClientMessagesRaw 测试原始请求体透传.
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

	// 构造包含 thinking 等额外字段的原始请求——验证透传不丢失
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
		t.Fatalf("请求失败: %v", err)
	}

	// 验证原始请求体被完整转发（含 thinking 字段）
	if !strings.Contains(string(receivedBody), `"thinking"`) {
		t.Error("原始透传应保留 thinking 字段")
	}
	if !strings.Contains(string(receivedBody), `"parallel_tool_use"`) {
		t.Error("原始透传应保留 parallel_tool_use 字段")
	}

	// 验证额外请求头被转发
	if receivedHeaders.Get("anthropic-beta") != "extended-thinking-2025-04-11" {
		t.Errorf("anthropic-beta 头未被转发: %s", receivedHeaders.Get("anthropic-beta"))
	}

	// 验证 usage 提取
	if usage == nil {
		t.Fatal("Usage 不应为 nil")
	}
	if usage.InputTokens != 20 || usage.OutputTokens != 10 {
		t.Errorf("Usage 不正确: input=%d, output=%d", usage.InputTokens, usage.OutputTokens)
	}

	// 验证原始响应完整
	if !strings.Contains(string(respBytes), "msg_raw") {
		t.Error("原始响应不完整")
	}
}

// TestAnthropicClientMessagesStreamRaw 测试原始请求体流式透传.
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
		t.Fatalf("请求失败: %v", err)
	}
	defer body.Close()

	data, err := io.ReadAll(body)
	if err != nil {
		t.Fatalf("读取流失败: %v", err)
	}
	if !strings.Contains(string(data), "message_start") {
		t.Error("流应包含 message_start 事件")
	}
}

// TestAnthropicClientErrorHandling 测试错误处理.
func TestAnthropicClientErrorHandling(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		expectCode int
	}{
		{"429 → 529 过载", http.StatusTooManyRequests, 529},
		{"500 → 502 服务错误", http.StatusInternalServerError, 502},
		{"400 → 400 原样返回", http.StatusBadRequest, 400},
		{"401 → 401 原样返回", http.StatusUnauthorized, 401},
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
				t.Fatal("应返回错误")
			}
			llmErr, ok := err.(*Error)
			if !ok {
				t.Fatalf("应返回 Error 类型, 实际: %T", err)
			}
			if llmErr.StatusCode != tt.expectCode {
				t.Errorf("状态码映射不正确: 预期 %d, 实际 %d", tt.expectCode, llmErr.StatusCode)
			}
		})
	}
}

// ══════════════════════════════════
// Anthropic SSE 流式读取器测试
// ══════════════════════════════════

// TestAnthropicStreamReaderBasic 测试基本 SSE 事件读取.
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
			t.Fatalf("第 %d 个事件读取失败: %v", i, err)
		}
		if eventType != expected {
			t.Errorf("第 %d 个事件类型不正确: 预期 %s, 实际 %s", i, expected, eventType)
		}
		if rawLines == "" {
			t.Errorf("第 %d 个事件原始文本不应为空", i)
		}
	}

	// message_stop 后应标记 done
	if !reader.IsDone() {
		t.Error("message_stop 后 reader 应标记为 done")
	}

	// 再次读取应返回 EOF
	_, _, _, err := reader.ReadEvent()
	if err != io.EOF {
		t.Errorf("应返回 io.EOF, 实际: %v", err)
	}
}

// TestAnthropicStreamReaderToolUse 测试工具调用流式事件.
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
		t.Fatalf("读取失败: %v", err)
	}
	if eventType != AnthropicEventContentBlockStart {
		t.Errorf("事件类型不正确: %s", eventType)
	}
	if event.ContentBlock == nil || event.ContentBlock.Type != "tool_use" {
		t.Error("content_block 应为 tool_use 类型")
	}
	if event.ContentBlock.Name != "get_weather" {
		t.Errorf("工具名不正确: %s", event.ContentBlock.Name)
	}

	// content_block_delta (input_json_delta)
	eventType, _, event, err = reader.ReadEvent()
	if err != nil {
		t.Fatalf("读取失败: %v", err)
	}
	if eventType != AnthropicEventContentBlockDelta {
		t.Errorf("事件类型不正确: %s", eventType)
	}
	if event.Delta == nil || event.Delta.Type != "input_json_delta" {
		t.Error("delta 应为 input_json_delta 类型")
	}
	if event.Delta.PartialJSON != `{"city":` {
		t.Errorf("partial_json 不正确: %s", event.Delta.PartialJSON)
	}
}

// TestAnthropicStreamReaderThinking 测试扩展思考流式事件.
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
		t.Errorf("事件类型不正确: %s", eventType)
	}
	if event.ContentBlock == nil || event.ContentBlock.Type != "thinking" {
		t.Error("content_block 应为 thinking 类型")
	}

	// thinking_delta
	eventType, _, event, _ = reader.ReadEvent()
	if eventType != AnthropicEventContentBlockDelta {
		t.Errorf("事件类型不正确: %s", eventType)
	}
	if event.Delta == nil || event.Delta.Type != "thinking_delta" {
		t.Error("delta 应为 thinking_delta 类型")
	}
	if event.Delta.Thinking != "Let me think about this..." {
		t.Errorf("thinking 内容不正确: %s", event.Delta.Thinking)
	}

	// signature_delta
	_, _, event, _ = reader.ReadEvent()
	if event.Delta == nil || event.Delta.Type != "signature_delta" {
		t.Error("delta 应为 signature_delta 类型")
	}
	if event.Delta.Signature != "EqQBCgIYAhIM..." {
		t.Errorf("signature 不正确: %s", event.Delta.Signature)
	}

	// content_block_stop
	reader.ReadEvent()

	// content_block_start (text)
	reader.ReadEvent()

	// text_delta
	_, _, event, _ = reader.ReadEvent()
	if event.Delta.Text != "The answer is 42." {
		t.Errorf("文本不正确: %s", event.Delta.Text)
	}

	// content_block_stop
	reader.ReadEvent()

	// message_delta (含 usage)
	eventType, _, event, _ = reader.ReadEvent()
	if eventType != AnthropicEventMessageDelta {
		t.Errorf("事件类型不正确: %s", eventType)
	}
	if event.Usage == nil || event.Usage.OutputTokens != 100 {
		t.Error("message_delta 中的 usage 不正确")
	}
}

// TestAnthropicStreamReaderEmpty 测试空流.
func TestAnthropicStreamReaderEmpty(t *testing.T) {
	body := io.NopCloser(strings.NewReader(""))
	reader := NewAnthropicStreamReader(body)
	defer reader.Close()

	_, _, _, err := reader.ReadEvent()
	if err != io.EOF {
		t.Errorf("空流应返回 io.EOF, 实际: %v", err)
	}
}

// ══════════════════════════════════
// Anthropic 错误响应测试
// ══════════════════════════════════

// TestAnthropicErrorResponseSerialization 验证错误响应格式.
func TestAnthropicErrorResponseSerialization(t *testing.T) {
	errResp := AnthropicErrorResponse{
		Type: "error",
	}
	errResp.Error.Type = "invalid_request_error"
	errResp.Error.Message = "messages: Required"

	data, _ := json.Marshal(errResp)
	s := string(data)

	if !strings.Contains(s, `"type":"error"`) {
		t.Error("错误响应缺少 type:error")
	}
	if !strings.Contains(s, `"invalid_request_error"`) {
		t.Error("错误响应缺少 error type")
	}
	if !strings.Contains(s, `messages: Required`) {
		t.Error("错误响应缺少 error message")
	}
}
