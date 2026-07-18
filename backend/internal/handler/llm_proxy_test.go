package handler

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// TestPipeOpenAIStreamToAnthropicFinalizeOnEOF 验证 EOF 时发出 message_delta + message_stop 收尾事件。
func TestPipeOpenAIStreamToAnthropicFinalizeOnEOF(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	streamBody := strings.Join([]string{
		`data: {"id":"chatcmpl-1","object":"chat.completion.chunk","created":0,"model":"gpt-4","choices":[{"index":0,"delta":{"role":"assistant"}}]}`,
		``,
		`data: {"id":"chatcmpl-1","object":"chat.completion.chunk","created":0,"model":"gpt-4","choices":[{"index":0,"delta":{"content":"Hi"}}]}`,
		``,
		`data: {"id":"chatcmpl-1","object":"chat.completion.chunk","created":0,"model":"gpt-4","choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}`,
		``,
		`data: {"id":"chatcmpl-1","object":"chat.completion.chunk","created":0,"model":"gpt-4","choices":[],"usage":{"prompt_tokens":10,"completion_tokens":5,"total_tokens":15}}`,
		``,
	}, "\n")

	h := &LLMProxyHandler{logger: zap.NewNop()}
	result := h.pipeOpenAIStreamToAnthropic(c, io.NopCloser(strings.NewReader(streamBody)), "gpt-4")

	body := w.Body.String()
	if !strings.Contains(body, "event: message_delta") {
		t.Error("EOF 收尾应发出 message_delta")
	}
	if !strings.Contains(body, "event: message_stop") {
		t.Error("EOF 收尾应发出 message_stop")
	}
	if !strings.Contains(body, `"input_tokens":10`) {
		t.Error("message_delta 应包含非零 usage")
	}
	if result.Usage == nil || result.Usage.PromptTokens != 10 {
		t.Errorf("应记录 usage，got: %+v", result.Usage)
	}
	if result.Content != "Hi" {
		t.Errorf("应累积文本内容，got: %q", result.Content)
	}
}
