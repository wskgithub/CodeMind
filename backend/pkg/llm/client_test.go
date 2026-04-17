package llm

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestChatCompletion tests non-streaming chat completion.
func TestChatCompletion(t *testing.T) {
	// Mock LLM service
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and path
		if r.Method != "POST" {
			t.Errorf("expected POST method, got: %s", r.Method)
		}
		if r.URL.Path != "/v1/chat/completions" {
			t.Errorf("expected path /v1/chat/completions, got: %s", r.URL.Path)
		}

		// Verify request headers
		if r.Header.Get("Content-Type") != "application/json" {
			t.Error("missing Content-Type header")
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Error("incorrect Authorization header")
		}

		// Return mock response
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
		t.Fatalf("request failed: %v", err)
	}

	if resp.ID != "chatcmpl-123" {
		t.Errorf("incorrect response ID: %s", resp.ID)
	}
	if resp.Usage == nil || resp.Usage.TotalTokens != 15 {
		t.Error("incorrect Usage info")
	}
	if len(resp.Choices) != 1 || resp.Choices[0].Message.Content != "Hello!" {
		t.Error("incorrect Choices content")
	}
}

// TestListModels tests model list retrieval.
func TestListModels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET method, got: %s", r.Method)
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
		t.Fatalf("request failed: %v", err)
	}

	if len(resp.Data) != 1 || resp.Data[0].ID != "gpt-4" {
		t.Error("incorrect model list content")
	}
}

// TestLLMErrorHandling tests LLM error handling.
func TestLLMErrorHandling(t *testing.T) {
	// Test 500 error
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
		t.Fatal("should return error")
	}

	llmErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("should return Error type, got: %T", err)
	}

	// 500 should map to 502
	if llmErr.StatusCode != 502 {
		t.Errorf("500 should map to 502, got: %d", llmErr.StatusCode)
	}
}

// TestLLMRateLimitHandling tests 429 rate limit handling.
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
		t.Fatal("should return error")
	}

	llmErr := err.(*Error)
	// 429 should map to 503
	if llmErr.StatusCode != 503 {
		t.Errorf("429 should map to 503, got: %d", llmErr.StatusCode)
	}
}

// TestChatCompletionStream tests streaming chat (basic connectivity).
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
		t.Fatalf("request failed: %v", err)
	}
	defer body.Close()

	// Verify stream can be read
	data, err := io.ReadAll(body)
	if err != nil {
		t.Fatalf("failed to read stream: %v", err)
	}

	if len(data) == 0 {
		t.Error("stream data should not be empty")
	}
}
