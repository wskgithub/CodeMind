package llm

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ──────────────────────────────────
// OpenAI Chat Completions — 请求类型
// 参考: https://platform.openai.com/docs/api-reference/chat/create
// ──────────────────────────────────

// ChatCompletionRequest Chat Completions 请求体
type ChatCompletionRequest struct {
	Model               string          `json:"model"`
	Messages            []ChatMessage   `json:"messages"`
	Stream              bool            `json:"stream,omitempty"`
	StreamOptions       *StreamOptions  `json:"stream_options,omitempty"`
	Temperature         *float64        `json:"temperature,omitempty"`
	TopP                *float64        `json:"top_p,omitempty"`
	MaxTokens           *int            `json:"max_tokens,omitempty"`
	MaxCompletionTokens *int            `json:"max_completion_tokens,omitempty"`
	Stop                interface{}     `json:"stop,omitempty"`
	N                   *int            `json:"n,omitempty"`
	User                string          `json:"user,omitempty"`
	Tools               []Tool          `json:"tools,omitempty"`
	ToolChoice          interface{}     `json:"tool_choice,omitempty"`
	ResponseFormat      *ResponseFormat `json:"response_format,omitempty"`
	Seed                *int64          `json:"seed,omitempty"`
	FrequencyPenalty    *float64        `json:"frequency_penalty,omitempty"`
	PresencePenalty     *float64        `json:"presence_penalty,omitempty"`
	LogProbs            *bool           `json:"logprobs,omitempty"`
	TopLogProbs         *int            `json:"top_logprobs,omitempty"`
	ParallelToolCalls   *bool           `json:"parallel_tool_calls,omitempty"`
	ReasoningEffort     string          `json:"reasoning_effort,omitempty"`
	ServiceTier         string          `json:"service_tier,omitempty"`
	Store               *bool           `json:"store,omitempty"`
	Metadata            map[string]string `json:"metadata,omitempty"`
	Prediction          interface{}     `json:"prediction,omitempty"`

	// 已废弃字段（保持向后兼容）
	Functions    []Function  `json:"functions,omitempty"`
	FunctionCall interface{} `json:"function_call,omitempty"`
}

// ChatMessage 对话消息
// Content 类型为 interface{}，可以是 string 或 []ContentPart（多模态场景）
type ChatMessage struct {
	Role             string        `json:"role"`
	Content          interface{}   `json:"content,omitempty"`
	Name             string        `json:"name,omitempty"`
	ToolCalls        []ToolCall    `json:"tool_calls,omitempty"`
	ToolCallID       string        `json:"tool_call_id,omitempty"`
	FunctionCall     *FunctionCall `json:"function_call,omitempty"`
	ReasoningContent string        `json:"reasoning_content,omitempty"`
	Refusal          *string       `json:"refusal,omitempty"`
}

// ContentString 获取消息的纯文本内容
// 当 Content 为数组（多模态）时，仅提取文本部分并拼接
func (m *ChatMessage) ContentString() string {
	if m.Content == nil {
		return ""
	}
	switch v := m.Content.(type) {
	case string:
		return v
	case []interface{}:
		var parts []string
		for _, part := range v {
			if partMap, ok := part.(map[string]interface{}); ok {
				if partMap["type"] == "text" {
					if text, ok := partMap["text"].(string); ok {
						parts = append(parts, text)
					}
				}
			}
		}
		return strings.Join(parts, "")
	default:
		return fmt.Sprintf("%v", v)
	}
}

// ──────────────────────────────────
// 消息内容部分（多模态支持）
// ──────────────────────────────────

// ContentPart 内容部分
type ContentPart struct {
	Type       string      `json:"type"`
	Text       string      `json:"text,omitempty"`
	ImageURL   *ImageURL   `json:"image_url,omitempty"`
	InputAudio *InputAudio `json:"input_audio,omitempty"`
}

// ImageURL 图片 URL
type ImageURL struct {
	URL    string `json:"url"`
	Detail string `json:"detail,omitempty"`
}

// InputAudio 音频输入
type InputAudio struct {
	Data   string `json:"data"`
	Format string `json:"format"`
}

// ──────────────────────────────────
// 工具定义与调用
// ──────────────────────────────────

// Tool 工具定义
type Tool struct {
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

// ToolFunction 工具函数定义
type ToolFunction struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	Parameters  interface{} `json:"parameters,omitempty"`
	Strict      *bool       `json:"strict,omitempty"`
}

// ToolCall 工具调用（assistant 消息中返回）
type ToolCall struct {
	ID       string           `json:"id"`
	Type     string           `json:"type"`
	Function ToolCallFunction `json:"function"`
	Index    *int             `json:"index,omitempty"`
}

// ToolCallFunction 工具调用的函数部分
type ToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// FunctionCall 函数调用（已废弃，兼容旧版 function_call）
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// Function 函数定义（已废弃，兼容旧版 functions）
type Function struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	Parameters  interface{} `json:"parameters,omitempty"`
}

// ──────────────────────────────────
// 响应格式
// ──────────────────────────────────

// ResponseFormat 指定模型输出格式
type ResponseFormat struct {
	Type       string      `json:"type"`
	JSONSchema interface{} `json:"json_schema,omitempty"`
}

// StreamOptions 流式选项
type StreamOptions struct {
	IncludeUsage *bool `json:"include_usage,omitempty"`
}

// ──────────────────────────────────
// OpenAI Chat Completions — 响应类型
// ──────────────────────────────────

// ChatCompletionResponse Chat Completions 非流式响应
type ChatCompletionResponse struct {
	ID                string       `json:"id"`
	Object            string       `json:"object"`
	Created           int64        `json:"created"`
	Model             string       `json:"model"`
	Choices           []ChatChoice `json:"choices"`
	Usage             *Usage       `json:"usage,omitempty"`
	SystemFingerprint string       `json:"system_fingerprint,omitempty"`
	ServiceTier       string       `json:"service_tier,omitempty"`
}

// ChatChoice 对话选择
type ChatChoice struct {
	Index        int          `json:"index"`
	Message      *ChatMessage `json:"message,omitempty"`
	Delta        *ChatMessage `json:"delta,omitempty"`
	FinishReason *string      `json:"finish_reason,omitempty"`
	Logprobs     interface{}  `json:"logprobs,omitempty"`
}

// ChatCompletionChunk 流式对话响应块
type ChatCompletionChunk struct {
	ID                string       `json:"id"`
	Object            string       `json:"object"`
	Created           int64        `json:"created"`
	Model             string       `json:"model"`
	Choices           []ChatChoice `json:"choices"`
	Usage             *Usage       `json:"usage,omitempty"`
	SystemFingerprint string       `json:"system_fingerprint,omitempty"`
	ServiceTier       string       `json:"service_tier,omitempty"`
}

// ──────────────────────────────────
// OpenAI Completions — 请求/响应类型
// 参考: https://platform.openai.com/docs/api-reference/completions
// ──────────────────────────────────

// CompletionRequest Completions 请求体
type CompletionRequest struct {
	Model            string      `json:"model"`
	Prompt           interface{} `json:"prompt"`
	Stream           bool        `json:"stream,omitempty"`
	StreamOptions    *StreamOptions `json:"stream_options,omitempty"`
	MaxTokens        *int        `json:"max_tokens,omitempty"`
	Temperature      *float64    `json:"temperature,omitempty"`
	TopP             *float64    `json:"top_p,omitempty"`
	Stop             interface{} `json:"stop,omitempty"`
	N                *int        `json:"n,omitempty"`
	Suffix           string      `json:"suffix,omitempty"`
	LogProbs         *int        `json:"logprobs,omitempty"`
	Echo             *bool       `json:"echo,omitempty"`
	BestOf           *int        `json:"best_of,omitempty"`
	LogitBias        interface{} `json:"logit_bias,omitempty"`
	FrequencyPenalty *float64    `json:"frequency_penalty,omitempty"`
	PresencePenalty  *float64    `json:"presence_penalty,omitempty"`
	Seed             *int64      `json:"seed,omitempty"`
	User             string      `json:"user,omitempty"`
}

// CompletionResponse Completions 非流式响应
type CompletionResponse struct {
	ID                string             `json:"id"`
	Object            string             `json:"object"`
	Created           int64              `json:"created"`
	Model             string             `json:"model"`
	Choices           []CompletionChoice `json:"choices"`
	Usage             *Usage             `json:"usage,omitempty"`
	SystemFingerprint string             `json:"system_fingerprint,omitempty"`
}

// CompletionChoice 补全选择
type CompletionChoice struct {
	Index        int         `json:"index"`
	Text         string      `json:"text"`
	FinishReason *string     `json:"finish_reason,omitempty"`
	Logprobs     interface{} `json:"logprobs,omitempty"`
}

// ──────────────────────────────────
// Token 用量
// ──────────────────────────────────

// Usage Token 用量信息
type Usage struct {
	PromptTokens            int                      `json:"prompt_tokens"`
	CompletionTokens        int                      `json:"completion_tokens"`
	TotalTokens             int                      `json:"total_tokens"`
	CompletionTokensDetails *CompletionTokensDetails `json:"completion_tokens_details,omitempty"`
	PromptTokensDetails     *PromptTokensDetails     `json:"prompt_tokens_details,omitempty"`
}

// CompletionTokensDetails 补全 Token 分类详情
type CompletionTokensDetails struct {
	ReasoningTokens          int `json:"reasoning_tokens,omitempty"`
	AudioTokens              int `json:"audio_tokens,omitempty"`
	AcceptedPredictionTokens int `json:"accepted_prediction_tokens,omitempty"`
	RejectedPredictionTokens int `json:"rejected_prediction_tokens,omitempty"`
}

// PromptTokensDetails 提示 Token 分类详情
type PromptTokensDetails struct {
	CachedTokens int `json:"cached_tokens,omitempty"`
	AudioTokens  int `json:"audio_tokens,omitempty"`
}

// ──────────────────────────────────
// Models API
// 参考: https://platform.openai.com/docs/api-reference/models
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
// Embeddings API
// 参考: https://platform.openai.com/docs/api-reference/embeddings
// ──────────────────────────────────

// EmbeddingRequest 向量嵌入请求
type EmbeddingRequest struct {
	Input          interface{} `json:"input"`
	Model          string      `json:"model"`
	EncodingFormat string      `json:"encoding_format,omitempty"`
	Dimensions     *int        `json:"dimensions,omitempty"`
	User           string      `json:"user,omitempty"`
}

// EmbeddingResponse 向量嵌入响应
type EmbeddingResponse struct {
	Object string          `json:"object"`
	Data   []EmbeddingData `json:"data"`
	Model  string          `json:"model"`
	Usage  *EmbeddingUsage `json:"usage,omitempty"`
}

// EmbeddingData 单条向量嵌入数据
type EmbeddingData struct {
	Object    string      `json:"object"`
	Embedding interface{} `json:"embedding"`
	Index     int         `json:"index"`
}

// EmbeddingUsage 嵌入请求的 Token 用量
type EmbeddingUsage struct {
	PromptTokens int `json:"prompt_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

// ──────────────────────────────────
// 错误响应
// ──────────────────────────────────

// ErrorResponse OpenAI 格式错误响应
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail 错误详情
type ErrorDetail struct {
	Message string  `json:"message"`
	Type    string  `json:"type"`
	Param   *string `json:"param"`
	Code    string  `json:"code"`
}

// ──────────────────────────────────
// 流式响应聚合结果
// ──────────────────────────────────

// StreamResult 流式响应的聚合结果
// pipe 函数在转发 SSE 流的同时收集内容，最终返回此结构体
type StreamResult struct {
	Usage        *Usage
	Content      string // 拼接后的完整文本内容
	ResponseID   string // 响应 ID（来自第一个 chunk）
	Model        string // 实际使用的模型名称
	FinishReason string // 终止原因（stop / length 等）
}

// ──────────────────────────────────
// 辅助类型与方法
// ──────────────────────────────────

// RequestMeta 请求路由元数据
// 从原始请求体中提取代理路由所需的最小字段
type RequestMeta struct {
	Model  string `json:"model"`
	Stream bool   `json:"stream"`
}

// ExtractRequestMeta 从原始请求体中提取路由所需的元数据
func ExtractRequestMeta(rawBody []byte) (*RequestMeta, error) {
	var meta RequestMeta
	if err := json.Unmarshal(rawBody, &meta); err != nil {
		return nil, err
	}
	return &meta, nil
}
