package llm

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestChatCompletion 测试非流式对话补全.
func TestChatCompletion(t *testing.T) {
	// 模拟 LLM 服务
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证请求方法和路径
		if r.Method != "POST" {
			t.Errorf("预期 POST 方法, 实际: %s", r.Method)
		}
		if r.URL.Path != "/v1/chat/completions" {
			t.Errorf("预期路径 /v1/chat/completions, 实际: %s", r.URL.Path)
		}

		// 验证请求头
		if r.Header.Get("Content-Type") != "application/json" {
			t.Error("缺少 Content-Type 头")
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Error("Authorization 头不正确")
		}

		// 返回模拟响应
		resp := ChatCompletionResponse{
			ID:      "chatcmpl-123",
			Object:  "chat.completion",
			Created: 1677652288,
			Model:   "gpt-4",
			Choices: []ChatChoice{
				{
					Index:   0,
					Message: &ChatMessage{Role: "assistant", Content: "Hello!"},
				},
			},
			Usage: &Usage{
				PromptTokens:     10,
				CompletionTokens: 5,
				TotalTokens:      15,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key", 30, 60)
	req := &ChatCompletionRequest{
		Model:    "gpt-4",
		Messages: []ChatMessage{{Role: "user", Content: "Hi"}},
	}

	resp, err := client.ChatCompletion(req)
	if err != nil {
		t.Fatalf("请求失败: %v", err)
	}

	if resp.ID != "chatcmpl-123" {
		t.Errorf("响应 ID 不正确: %s", resp.ID)
	}
	if resp.Usage == nil || resp.Usage.TotalTokens != 15 {
		t.Error("Usage 信息不正确")
	}
	if len(resp.Choices) != 1 || resp.Choices[0].Message.Content != "Hello!" {
		t.Error("Choices 内容不正确")
	}
}

// TestListModels 测试获取模型列表.
func TestListModels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("预期 GET 方法, 实际: %s", r.Method)
		}

		resp := ModelListResponse{
			Object: "list",
			Data: []ModelInfo{
				{ID: "gpt-4", Object: "model", Created: 1677652288, OwnedBy: "openai"},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key", 30, 60)
	resp, err := client.ListModels()
	if err != nil {
		t.Fatalf("请求失败: %v", err)
	}

	if len(resp.Data) != 1 || resp.Data[0].ID != "gpt-4" {
		t.Error("模型列表内容不正确")
	}
}

// TestLLMErrorHandling 测试 LLM 错误处理.
func TestLLMErrorHandling(t *testing.T) {
	// 测试 500 错误
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": {"message": "Internal server error"}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key", 30, 60)
	_, err := client.ChatCompletion(&ChatCompletionRequest{
		Model:    "gpt-4",
		Messages: []ChatMessage{{Role: "user", Content: "Hi"}},
	})

	if err == nil {
		t.Fatal("应返回错误")
	}

	llmErr, ok := err.(*LLMError)
	if !ok {
		t.Fatalf("应返回 LLMError 类型, 实际: %T", err)
	}

	// 500 应映射为 502
	if llmErr.StatusCode != 502 {
		t.Errorf("500 应映射为 502, 实际: %d", llmErr.StatusCode)
	}
}

// TestLLMRateLimitHandling 测试 429 限流处理.
func TestLLMRateLimitHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"error": {"message": "Rate limit exceeded"}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key", 30, 60)
	_, err := client.ChatCompletion(&ChatCompletionRequest{
		Model:    "gpt-4",
		Messages: []ChatMessage{{Role: "user", Content: "Hi"}},
	})

	if err == nil {
		t.Fatal("应返回错误")
	}

	llmErr := err.(*LLMError)
	// 429 应映射为 503
	if llmErr.StatusCode != 503 {
		t.Errorf("429 应映射为 503, 实际: %d", llmErr.StatusCode)
	}
}

// TestChatCompletionStream 测试流式对话（基本连通性）.
func TestChatCompletionStream(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("data: {\"id\":\"1\",\"object\":\"chat.completion.chunk\",\"choices\":[{\"delta\":{\"content\":\"Hi\"}}]}\n\n"))
		w.Write([]byte("data: [DONE]\n\n"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key", 30, 60)
	body, err := client.ChatCompletionStream(&ChatCompletionRequest{
		Model:    "gpt-4",
		Messages: []ChatMessage{{Role: "user", Content: "Hi"}},
	})
	if err != nil {
		t.Fatalf("请求失败: %v", err)
	}
	defer body.Close()

	// 验证能够读取流
	data, err := io.ReadAll(body)
	if err != nil {
		t.Fatalf("读取流失败: %v", err)
	}

	if len(data) == 0 {
		t.Error("流数据不应为空")
	}
}
