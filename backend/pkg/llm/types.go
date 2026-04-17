package llm

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ChatCompletionRequest represents an OpenAI Chat Completions request body.
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

	Functions    []Function  `json:"functions,omitempty"`
	FunctionCall interface{} `json:"function_call,omitempty"`
}

// ChatMessage represents a chat message.
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

// ContentString returns the plain text content of the message.
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

// ContentPart represents a part of multimodal message content.
type ContentPart struct {
	Type       string      `json:"type"`
	Text       string      `json:"text,omitempty"`
	ImageURL   *ImageURL   `json:"image_url,omitempty"`
	InputAudio *InputAudio `json:"input_audio,omitempty"`
}

// ImageURL represents an image URL.
type ImageURL struct {
	URL    string `json:"url"`
	Detail string `json:"detail,omitempty"`
}

// InputAudio represents audio input.
type InputAudio struct {
	Data   string `json:"data"`
	Format string `json:"format"`
}

// Tool represents a tool definition.
type Tool struct {
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

// ToolFunction represents a tool function definition.
type ToolFunction struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	Parameters  interface{} `json:"parameters,omitempty"`
	Strict      *bool       `json:"strict,omitempty"`
}

// ToolCall represents a tool call in assistant messages.
type ToolCall struct {
	ID       string           `json:"id"`
	Type     string           `json:"type"`
	Function ToolCallFunction `json:"function"`
	Index    *int             `json:"index,omitempty"`
}

// ToolCallFunction represents the function part of a tool call.
type ToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// FunctionCall represents a function call (deprecated).
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// Function represents a function definition (deprecated).
type Function struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	Parameters  interface{} `json:"parameters,omitempty"`
}

// ResponseFormat specifies the model output format.
type ResponseFormat struct {
	Type       string      `json:"type"`
	JSONSchema interface{} `json:"json_schema,omitempty"`
}

// StreamOptions represents streaming options.
type StreamOptions struct {
	IncludeUsage *bool `json:"include_usage,omitempty"`
}

// ChatCompletionResponse represents a Chat Completions response.
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

// ChatChoice represents a chat completion choice.
type ChatChoice struct {
	Index        int          `json:"index"`
	Message      *ChatMessage `json:"message,omitempty"`
	Delta        *ChatMessage `json:"delta,omitempty"`
	FinishReason *string      `json:"finish_reason,omitempty"`
	Logprobs     interface{}  `json:"logprobs,omitempty"`
}

// ChatCompletionChunk represents a streaming chat response chunk.
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

// CompletionRequest represents a Completions request body.
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

// CompletionResponse represents a Completions response.
type CompletionResponse struct {
	ID                string             `json:"id"`
	Object            string             `json:"object"`
	Created           int64              `json:"created"`
	Model             string             `json:"model"`
	Choices           []CompletionChoice `json:"choices"`
	Usage             *Usage             `json:"usage,omitempty"`
	SystemFingerprint string             `json:"system_fingerprint,omitempty"`
}

// CompletionChoice represents a completion choice.
type CompletionChoice struct {
	Index        int         `json:"index"`
	Text         string      `json:"text"`
	FinishReason *string     `json:"finish_reason,omitempty"`
	Logprobs     interface{} `json:"logprobs,omitempty"`
}

// Usage represents token usage information.
type Usage struct {
	PromptTokens            int                      `json:"prompt_tokens"`
	CompletionTokens        int                      `json:"completion_tokens"`
	TotalTokens             int                      `json:"total_tokens"`
	CompletionTokensDetails *CompletionTokensDetails `json:"completion_tokens_details,omitempty"`
	PromptTokensDetails     *PromptTokensDetails     `json:"prompt_tokens_details,omitempty"`
}

// CompletionTokensDetails contains detailed completion token breakdown.
type CompletionTokensDetails struct {
	ReasoningTokens          int `json:"reasoning_tokens,omitempty"`
	AudioTokens              int `json:"audio_tokens,omitempty"`
	AcceptedPredictionTokens int `json:"accepted_prediction_tokens,omitempty"`
	RejectedPredictionTokens int `json:"rejected_prediction_tokens,omitempty"`
}

// PromptTokensDetails contains detailed prompt token breakdown.
type PromptTokensDetails struct {
	CacheCreationInputTokens int `json:"cache_creation_input_tokens,omitempty"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens,omitempty"`
	CachedTokens             int `json:"cached_tokens,omitempty"`
	AudioTokens              int `json:"audio_tokens,omitempty"`
}

// ModelListResponse represents a model list response.
type ModelListResponse struct {
	Object string      `json:"object"`
	Data   []ModelInfo `json:"data"`
}

// ModelInfo represents a single model's information.
type ModelInfo struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

// EmbeddingRequest represents an embedding request.
type EmbeddingRequest struct {
	Input          interface{} `json:"input"`
	Model          string      `json:"model"`
	EncodingFormat string      `json:"encoding_format,omitempty"`
	Dimensions     *int        `json:"dimensions,omitempty"`
	User           string      `json:"user,omitempty"`
}

// EmbeddingResponse represents an embedding response.
type EmbeddingResponse struct {
	Object string          `json:"object"`
	Data   []EmbeddingData `json:"data"`
	Model  string          `json:"model"`
	Usage  *EmbeddingUsage `json:"usage,omitempty"`
}

// EmbeddingData represents a single embedding data item.
type EmbeddingData struct {
	Object    string      `json:"object"`
	Embedding interface{} `json:"embedding"`
	Index     int         `json:"index"`
}

// EmbeddingUsage represents token usage for embedding requests.
type EmbeddingUsage struct {
	PromptTokens int `json:"prompt_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

// ErrorResponse represents an OpenAI format error response.
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains error details.
type ErrorDetail struct {
	Message string  `json:"message"`
	Type    string  `json:"type"`
	Param   *string `json:"param"`
	Code    string  `json:"code"`
}

// StreamResult contains aggregated streaming response data.
type StreamResult struct {
	Usage        *Usage
	Content      string
	ResponseID   string
	Model        string
	FinishReason string
}

// RequestMeta contains minimal routing metadata extracted from requests.
type RequestMeta struct {
	Model  string `json:"model"`
	Stream bool   `json:"stream"`
}

// ExtractRequestMeta extracts routing metadata from a raw request body.
func ExtractRequestMeta(rawBody []byte) (*RequestMeta, error) {
	var meta RequestMeta
	if err := json.Unmarshal(rawBody, &meta); err != nil {
		return nil, err
	}
	return &meta, nil
}
