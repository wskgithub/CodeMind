package llm

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ChatCompletionRequest represents an OpenAI Chat Completions request body.
type ChatCompletionRequest struct {
	Stop                interface{}       `json:"stop,omitempty"`
	FunctionCall        interface{}       `json:"function_call,omitempty"`
	Prediction          interface{}       `json:"prediction,omitempty"`
	ToolChoice          interface{}       `json:"tool_choice,omitempty"`
	MaxTokens           *int              `json:"max_tokens,omitempty"`
	Store               *bool             `json:"store,omitempty"`
	Metadata            map[string]string `json:"metadata,omitempty"`
	MaxCompletionTokens *int              `json:"max_completion_tokens,omitempty"`
	Temperature         *float64          `json:"temperature,omitempty"`
	N                   *int              `json:"n,omitempty"`
	TopP                *float64          `json:"top_p,omitempty"`
	ParallelToolCalls   *bool             `json:"parallel_tool_calls,omitempty"`
	StreamOptions       *StreamOptions    `json:"stream_options,omitempty"`
	ResponseFormat      *ResponseFormat   `json:"response_format,omitempty"`
	Seed                *int64            `json:"seed,omitempty"`
	FrequencyPenalty    *float64          `json:"frequency_penalty,omitempty"`
	PresencePenalty     *float64          `json:"presence_penalty,omitempty"`
	LogProbs            *bool             `json:"logprobs,omitempty"`
	TopLogProbs         *int              `json:"top_logprobs,omitempty"`
	User                string            `json:"user,omitempty"`
	ReasoningEffort     string            `json:"reasoning_effort,omitempty"`
	ServiceTier         string            `json:"service_tier,omitempty"`
	Model               string            `json:"model"`
	Tools               []Tool            `json:"tools,omitempty"`
	Functions           []Function        `json:"functions,omitempty"`
	Messages            []ChatMessage     `json:"messages"`
	Stream              bool              `json:"stream,omitempty"`
}

// ChatMessage represents a chat message.
type ChatMessage struct {
	Content          interface{}   `json:"content,omitempty"`
	FunctionCall     *FunctionCall `json:"function_call,omitempty"`
	Refusal          *string       `json:"refusal,omitempty"`
	Role             string        `json:"role"`
	Name             string        `json:"name,omitempty"`
	ToolCallID       string        `json:"tool_call_id,omitempty"`
	ReasoningContent string        `json:"reasoning_content,omitempty"`
	ToolCalls        []ToolCall    `json:"tool_calls,omitempty"`
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
				if partMap["type"] == ContentTypeText {
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
	ImageURL   *ImageURL   `json:"image_url,omitempty"`
	InputAudio *InputAudio `json:"input_audio,omitempty"`
	Type       string      `json:"type"`
	Text       string      `json:"text,omitempty"`
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
	Function ToolFunction `json:"function"`
	Type     string       `json:"type"`
}

// ToolFunction represents a tool function definition.
type ToolFunction struct {
	Parameters  interface{} `json:"parameters,omitempty"`
	Strict      *bool       `json:"strict,omitempty"`
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
}

// ToolCall represents a tool call in assistant messages.
type ToolCall struct {
	Index    *int             `json:"index,omitempty"`
	Function ToolCallFunction `json:"function"`
	ID       string           `json:"id"`
	Type     string           `json:"type"`
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
	Parameters  interface{} `json:"parameters,omitempty"`
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
}

// ResponseFormat specifies the model output format.
type ResponseFormat struct {
	JSONSchema interface{} `json:"json_schema,omitempty"`
	Type       string      `json:"type"`
}

// StreamOptions represents streaming options.
type StreamOptions struct {
	IncludeUsage *bool `json:"include_usage,omitempty"`
}

// ChatCompletionResponse represents a Chat Completions response.
type ChatCompletionResponse struct {
	Usage             *Usage       `json:"usage,omitempty"`
	ID                string       `json:"id"`
	Object            string       `json:"object"`
	Model             string       `json:"model"`
	SystemFingerprint string       `json:"system_fingerprint,omitempty"`
	ServiceTier       string       `json:"service_tier,omitempty"`
	Choices           []ChatChoice `json:"choices"`
	Created           int64        `json:"created"`
}

// ChatChoice represents a chat completion choice.
type ChatChoice struct {
	Logprobs     interface{}  `json:"logprobs,omitempty"`
	Message      *ChatMessage `json:"message,omitempty"`
	Delta        *ChatMessage `json:"delta,omitempty"`
	FinishReason *string      `json:"finish_reason,omitempty"`
	Index        int          `json:"index"`
}

// ChatCompletionChunk represents a streaming chat response chunk.
type ChatCompletionChunk struct {
	Usage             *Usage       `json:"usage,omitempty"`
	ID                string       `json:"id"`
	Object            string       `json:"object"`
	Model             string       `json:"model"`
	SystemFingerprint string       `json:"system_fingerprint,omitempty"`
	ServiceTier       string       `json:"service_tier,omitempty"`
	Choices           []ChatChoice `json:"choices"`
	Created           int64        `json:"created"`
}

// CompletionRequest represents a Completions request body.
type CompletionRequest struct {
	Stop             interface{}    `json:"stop,omitempty"`
	Prompt           interface{}    `json:"prompt"`
	LogitBias        interface{}    `json:"logit_bias,omitempty"`
	N                *int           `json:"n,omitempty"`
	LogProbs         *int           `json:"logprobs,omitempty"`
	Temperature      *float64       `json:"temperature,omitempty"`
	TopP             *float64       `json:"top_p,omitempty"`
	StreamOptions    *StreamOptions `json:"stream_options,omitempty"`
	Seed             *int64         `json:"seed,omitempty"`
	PresencePenalty  *float64       `json:"presence_penalty,omitempty"`
	MaxTokens        *int           `json:"max_tokens,omitempty"`
	Echo             *bool          `json:"echo,omitempty"`
	BestOf           *int           `json:"best_of,omitempty"`
	FrequencyPenalty *float64       `json:"frequency_penalty,omitempty"`
	Suffix           string         `json:"suffix,omitempty"`
	Model            string         `json:"model"`
	User             string         `json:"user,omitempty"`
	Stream           bool           `json:"stream,omitempty"`
}

// CompletionResponse represents a Completions response.
type CompletionResponse struct {
	Usage             *Usage             `json:"usage,omitempty"`
	ID                string             `json:"id"`
	Object            string             `json:"object"`
	Model             string             `json:"model"`
	SystemFingerprint string             `json:"system_fingerprint,omitempty"`
	Choices           []CompletionChoice `json:"choices"`
	Created           int64              `json:"created"`
}

// CompletionChoice represents a completion choice.
type CompletionChoice struct {
	Logprobs     interface{} `json:"logprobs,omitempty"`
	FinishReason *string     `json:"finish_reason,omitempty"`
	Text         string      `json:"text"`
	Index        int         `json:"index"`
}

// Usage represents token usage information.
type Usage struct {
	CompletionTokensDetails *CompletionTokensDetails `json:"completion_tokens_details,omitempty"`
	PromptTokensDetails     *PromptTokensDetails     `json:"prompt_tokens_details,omitempty"`
	PromptTokens            int                      `json:"prompt_tokens"`
	CompletionTokens        int                      `json:"completion_tokens"`
	TotalTokens             int                      `json:"total_tokens"`
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
	OwnedBy string `json:"owned_by"`
	Created int64  `json:"created"`
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
	Usage  *EmbeddingUsage `json:"usage,omitempty"`
	Object string          `json:"object"`
	Model  string          `json:"model"`
	Data   []EmbeddingData `json:"data"`
}

// EmbeddingData represents a single embedding data item.
type EmbeddingData struct {
	Embedding interface{} `json:"embedding"`
	Object    string      `json:"object"`
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
