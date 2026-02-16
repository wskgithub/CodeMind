package llm

// ──────────────────────────────────
// OpenAI 兼容格式 — 请求类型
// ──────────────────────────────────

// ChatCompletionRequest Chat Completions 请求体
type ChatCompletionRequest struct {
	Model       string          `json:"model"`
	Messages    []ChatMessage   `json:"messages"`
	Stream      bool            `json:"stream,omitempty"`
	Temperature *float64        `json:"temperature,omitempty"`
	TopP        *float64        `json:"top_p,omitempty"`
	MaxTokens   *int            `json:"max_tokens,omitempty"`
	Stop        interface{}     `json:"stop,omitempty"`          // string 或 []string
	N           *int            `json:"n,omitempty"`
	User        string          `json:"user,omitempty"`
}

// ChatMessage 对话消息
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// CompletionRequest Completions 请求体
type CompletionRequest struct {
	Model       string      `json:"model"`
	Prompt      interface{} `json:"prompt"`                  // string 或 []string
	Stream      bool        `json:"stream,omitempty"`
	MaxTokens   *int        `json:"max_tokens,omitempty"`
	Temperature *float64    `json:"temperature,omitempty"`
	TopP        *float64    `json:"top_p,omitempty"`
	Stop        interface{} `json:"stop,omitempty"`
}

// ──────────────────────────────────
// OpenAI 兼容格式 — 响应类型
// ──────────────────────────────────

// ChatCompletionResponse Chat Completions 非流式响应
type ChatCompletionResponse struct {
	ID      string             `json:"id"`
	Object  string             `json:"object"`
	Created int64              `json:"created"`
	Model   string             `json:"model"`
	Choices []ChatChoice       `json:"choices"`
	Usage   *Usage             `json:"usage,omitempty"`
}

// ChatChoice 对话选择
type ChatChoice struct {
	Index        int          `json:"index"`
	Message      *ChatMessage `json:"message,omitempty"`
	Delta        *ChatMessage `json:"delta,omitempty"`
	FinishReason *string      `json:"finish_reason,omitempty"`
}

// CompletionResponse Completions 非流式响应
type CompletionResponse struct {
	ID      string             `json:"id"`
	Object  string             `json:"object"`
	Created int64              `json:"created"`
	Model   string             `json:"model"`
	Choices []CompletionChoice `json:"choices"`
	Usage   *Usage             `json:"usage,omitempty"`
}

// CompletionChoice 补全选择
type CompletionChoice struct {
	Index        int     `json:"index"`
	Text         string  `json:"text"`
	FinishReason *string `json:"finish_reason,omitempty"`
}

// Usage Token 用量信息
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ──────────────────────────────────
// 流式响应 Chunk 类型
// ──────────────────────────────────

// ChatCompletionChunk 流式对话响应块
type ChatCompletionChunk struct {
	ID      string       `json:"id"`
	Object  string       `json:"object"`
	Created int64        `json:"created"`
	Model   string       `json:"model"`
	Choices []ChatChoice `json:"choices"`
	Usage   *Usage       `json:"usage,omitempty"` // 最后一个 chunk 可能包含 usage
}

// ──────────────────────────────────
// Models API 响应
// ──────────────────────────────────

// ModelListResponse 模型列表响应
type ModelListResponse struct {
	Object string      `json:"object"`
	Data   []ModelInfo `json:"data"`
}

// ModelInfo 单个模型信息
type ModelInfo struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

// ──────────────────────────────────
// 错误响应
// ──────────────────────────────────

// ErrorResponse OpenAI 格式错误响应
type ErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}
