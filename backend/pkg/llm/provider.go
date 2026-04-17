package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
)

// ProviderFormat represents the LLM provider protocol format.
type ProviderFormat string

// LLM 服务商协议格式常量。
const (
	FormatOpenAI    ProviderFormat = "openai"
	FormatAnthropic ProviderFormat = "anthropic"
)

// Provider is the unified interface for LLM service providers.
type Provider interface {
	Name() string
	Format() ProviderFormat

	ChatCompletion(ctx context.Context, req *ChatCompletionRequest) (*ChatCompletionResponse, error)
	ChatCompletionStream(ctx context.Context, req *ChatCompletionRequest) (io.ReadCloser, error)
	ChatCompletionRaw(ctx context.Context, rawBody []byte) (rawResp []byte, usage *Usage, err error)
	ChatCompletionStreamRaw(ctx context.Context, rawBody []byte) (io.ReadCloser, error)

	Completion(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error)
	CompletionStream(ctx context.Context, req *CompletionRequest) (io.ReadCloser, error)
	CompletionRaw(ctx context.Context, rawBody []byte) (rawResp []byte, usage *Usage, err error)
	CompletionStreamRaw(ctx context.Context, rawBody []byte) (io.ReadCloser, error)

	ListModels(ctx context.Context) (*ModelListResponse, error)
	RetrieveModel(ctx context.Context, modelID string) (*ModelInfo, error)

	EmbeddingRaw(ctx context.Context, rawBody []byte) (rawResp []byte, usage *Usage, err error)

	ResponsesRaw(ctx context.Context, rawBody []byte) (rawResp []byte, usage *Usage, err error)
	ResponsesStreamRaw(ctx context.Context, rawBody []byte) (io.ReadCloser, error)

	AnthropicMessagesRaw(ctx context.Context, rawBody []byte) (rawResp []byte, usage *Usage, err error)
	AnthropicMessagesStreamRaw(ctx context.Context, rawBody []byte) (io.ReadCloser, error)
}

// OpenAIProvider wraps an OpenAI-compatible client as a Provider.
type OpenAIProvider struct {
	client *Client
	name   string
}

// NewOpenAIProvider creates an OpenAI provider.
func NewOpenAIProvider(name string, client *Client) *OpenAIProvider {
	return &OpenAIProvider{
		name:   name,
		client: client,
	}
}

// Name 返回服务商名称。
func (p *OpenAIProvider) Name() string { return p.name }

// Format 返回服务商协议格式。
func (p *OpenAIProvider) Format() ProviderFormat { return FormatOpenAI }

// ChatCompletion 执行非流式聊天补全。
func (p *OpenAIProvider) ChatCompletion(_ context.Context, req *ChatCompletionRequest) (*ChatCompletionResponse, error) {
	return p.client.ChatCompletion(req)
}

// ChatCompletionStream 执行流式聊天补全。
func (p *OpenAIProvider) ChatCompletionStream(_ context.Context, req *ChatCompletionRequest) (io.ReadCloser, error) {
	return p.client.ChatCompletionStream(req)
}

// ChatCompletionRaw 使用原始请求体执行非流式聊天补全。
func (p *OpenAIProvider) ChatCompletionRaw(_ context.Context, rawBody []byte) ([]byte, *Usage, error) {
	return p.client.ChatCompletionRawAll(rawBody)
}

// ChatCompletionStreamRaw 使用原始请求体执行流式聊天补全。
func (p *OpenAIProvider) ChatCompletionStreamRaw(_ context.Context, rawBody []byte) (io.ReadCloser, error) {
	return p.client.ChatCompletionStreamRaw(rawBody)
}

// Completion 执行非流式文本补全。
func (p *OpenAIProvider) Completion(_ context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	return p.client.Completion(req)
}

// CompletionStream 执行流式文本补全。
func (p *OpenAIProvider) CompletionStream(_ context.Context, req *CompletionRequest) (io.ReadCloser, error) {
	return p.client.CompletionStream(req)
}

// CompletionRaw 使用原始请求体执行非流式文本补全。
func (p *OpenAIProvider) CompletionRaw(_ context.Context, rawBody []byte) ([]byte, *Usage, error) {
	return p.client.CompletionRawAll(rawBody)
}

// CompletionStreamRaw 使用原始请求体执行流式文本补全。
func (p *OpenAIProvider) CompletionStreamRaw(_ context.Context, rawBody []byte) (io.ReadCloser, error) {
	return p.client.CompletionStreamRaw(rawBody)
}

// ListModels 获取可用模型列表。
func (p *OpenAIProvider) ListModels(_ context.Context) (*ModelListResponse, error) {
	return p.client.ListModels()
}

// RetrieveModel 获取单个模型信息。
func (p *OpenAIProvider) RetrieveModel(_ context.Context, modelID string) (*ModelInfo, error) {
	return p.client.RetrieveModel(modelID)
}

// EmbeddingRaw 使用原始请求体执行嵌入调用。
func (p *OpenAIProvider) EmbeddingRaw(_ context.Context, rawBody []byte) ([]byte, *Usage, error) {
	return p.client.EmbeddingRaw(rawBody)
}

// ResponsesRaw 使用原始请求体执行非流式 Responses API 调用。
func (p *OpenAIProvider) ResponsesRaw(_ context.Context, rawBody []byte) ([]byte, *Usage, error) {
	return p.client.ResponsesRaw(rawBody)
}

// ResponsesStreamRaw 使用原始请求体执行流式 Responses API 调用。
func (p *OpenAIProvider) ResponsesStreamRaw(_ context.Context, rawBody []byte) (io.ReadCloser, error) {
	return p.client.ResponsesStreamRaw(rawBody)
}

// AnthropicMessagesRaw converts Anthropic request to OpenAI and back.
func (p *OpenAIProvider) AnthropicMessagesRaw(_ context.Context, rawBody []byte) ([]byte, *Usage, error) {
	var req AnthropicMessagesRequest
	if err := json.Unmarshal(rawBody, &req); err != nil {
		return nil, nil, fmt.Errorf("failed to parse Anthropic request: %w", err)
	}
	openaiReq := AnthropicToOpenAI(&req)
	resp, err := p.client.ChatCompletion(openaiReq)
	if err != nil {
		return nil, nil, err
	}
	anthropicResp := OpenAIResponseToAnthropic(resp)
	data, err := json.Marshal(anthropicResp)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to serialize Anthropic response: %w", err)
	}
	return data, resp.Usage, nil
}

// AnthropicMessagesStreamRaw converts Anthropic streaming request to OpenAI format.
func (p *OpenAIProvider) AnthropicMessagesStreamRaw(_ context.Context, rawBody []byte) (io.ReadCloser, error) {
	var req AnthropicMessagesRequest
	if err := json.Unmarshal(rawBody, &req); err != nil {
		return nil, fmt.Errorf("failed to parse Anthropic request: %w", err)
	}
	openaiReq := AnthropicToOpenAI(&req)
	return p.client.ChatCompletionStream(openaiReq)
}

// AnthropicProvider wraps an Anthropic client as a Provider.
type AnthropicProvider struct {
	client *AnthropicClient
	name   string
}

// NewAnthropicProvider creates an Anthropic provider.
func NewAnthropicProvider(name string, client *AnthropicClient) *AnthropicProvider {
	return &AnthropicProvider{
		name:   name,
		client: client,
	}
}

// Name 返回服务商名称。
func (p *AnthropicProvider) Name() string { return p.name }

// Format 返回服务商协议格式。
func (p *AnthropicProvider) Format() ProviderFormat { return FormatAnthropic }

// ChatCompletion converts OpenAI request to Anthropic format.
func (p *AnthropicProvider) ChatCompletion(_ context.Context, req *ChatCompletionRequest) (*ChatCompletionResponse, error) {
	anthropicReq := OpenAIToAnthropic(req)
	resp, err := p.client.Messages(anthropicReq)
	if err != nil {
		return nil, err
	}
	return AnthropicResponseToOpenAI(resp), nil
}

// ChatCompletionStream returns Anthropic stream (handler converts format).
func (p *AnthropicProvider) ChatCompletionStream(_ context.Context, req *ChatCompletionRequest) (io.ReadCloser, error) {
	anthropicReq := OpenAIToAnthropic(req)
	return p.client.MessagesStream(anthropicReq)
}

// ChatCompletionRaw parses OpenAI request and converts to Anthropic.
func (p *AnthropicProvider) ChatCompletionRaw(ctx context.Context, rawBody []byte) ([]byte, *Usage, error) {
	var req ChatCompletionRequest
	if err := json.Unmarshal(rawBody, &req); err != nil {
		return nil, nil, fmt.Errorf("failed to parse request: %w", err)
	}
	resp, err := p.ChatCompletion(ctx, &req)
	if err != nil {
		return nil, nil, err
	}
	data, err := json.Marshal(resp)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to serialize response: %w", err)
	}
	return data, resp.Usage, nil
}

// ChatCompletionStreamRaw 将 OpenAI 流式请求转换为 Anthropic 格式并返回流。
func (p *AnthropicProvider) ChatCompletionStreamRaw(ctx context.Context, rawBody []byte) (io.ReadCloser, error) {
	var req ChatCompletionRequest
	if err := json.Unmarshal(rawBody, &req); err != nil {
		return nil, fmt.Errorf("failed to parse request: %w", err)
	}
	return p.ChatCompletionStream(ctx, &req)
}

// Completion converts to Messages call (Anthropic doesn't support raw Completions).
func (p *AnthropicProvider) Completion(_ context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	chatReq := completionToChatRequest(req)
	anthropicReq := OpenAIToAnthropic(chatReq)
	resp, err := p.client.Messages(anthropicReq)
	if err != nil {
		return nil, err
	}
	chatResp := AnthropicResponseToOpenAI(resp)
	return chatToCompletionResponse(chatResp), nil
}

// CompletionStream 执行流式文本补全（转换为 Anthropic Messages 调用）。
func (p *AnthropicProvider) CompletionStream(_ context.Context, req *CompletionRequest) (io.ReadCloser, error) {
	chatReq := completionToChatRequest(req)
	anthropicReq := OpenAIToAnthropic(chatReq)
	return p.client.MessagesStream(anthropicReq)
}

// CompletionRaw 使用原始请求体执行非流式文本补全。
func (p *AnthropicProvider) CompletionRaw(ctx context.Context, rawBody []byte) ([]byte, *Usage, error) {
	var req CompletionRequest
	if err := json.Unmarshal(rawBody, &req); err != nil {
		return nil, nil, fmt.Errorf("failed to parse request: %w", err)
	}
	resp, err := p.Completion(ctx, &req)
	if err != nil {
		return nil, nil, err
	}
	data, err := json.Marshal(resp)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to serialize response: %w", err)
	}
	return data, resp.Usage, nil
}

// CompletionStreamRaw 使用原始请求体执行流式文本补全。
func (p *AnthropicProvider) CompletionStreamRaw(ctx context.Context, rawBody []byte) (io.ReadCloser, error) {
	var req CompletionRequest
	if err := json.Unmarshal(rawBody, &req); err != nil {
		return nil, fmt.Errorf("failed to parse request: %w", err)
	}
	return p.CompletionStream(ctx, &req)
}

// ListModels returns a predefined model list (Anthropic has no standard API).
func (p *AnthropicProvider) ListModels(_ context.Context) (*ModelListResponse, error) {
	return &ModelListResponse{
		Object: "list",
		Data: []ModelInfo{
			{ID: "claude-sonnet-4-20250514", Object: "model", OwnedBy: "anthropic"},
			{ID: "claude-3-5-sonnet-20241022", Object: "model", OwnedBy: "anthropic"},
			{ID: "claude-3-5-haiku-20241022", Object: "model", OwnedBy: "anthropic"},
			{ID: "claude-3-opus-20240229", Object: "model", OwnedBy: "anthropic"},
		},
	}, nil
}

// RetrieveModel looks up model from predefined list.
func (p *AnthropicProvider) RetrieveModel(_ context.Context, modelID string) (*ModelInfo, error) {
	resp, _ := p.ListModels(context.TODO())
	for _, m := range resp.Data {
		if m.ID == modelID {
			return &m, nil
		}
	}
	return nil, &Error{StatusCode: 404, Message: fmt.Sprintf("model '%s' not found", modelID)} //nolint:mnd // intentional constant.
}

// EmbeddingRaw is not supported by Anthropic.
func (p *AnthropicProvider) EmbeddingRaw(_ context.Context, _ []byte) ([]byte, *Usage, error) {
	return nil, nil, &Error{StatusCode: 404, Message: "Anthropic does not support Embeddings API"} //nolint:mnd // intentional constant.
}

// ResponsesRaw 返回不支持错误（Anthropic 不支持 Responses API）。
func (p *AnthropicProvider) ResponsesRaw(_ context.Context, _ []byte) ([]byte, *Usage, error) {
	return nil, nil, &Error{StatusCode: 404, Message: "Anthropic does not support Responses API"} //nolint:mnd // intentional constant.
}

// ResponsesStreamRaw 返回不支持错误（Anthropic 不支持 Responses API）。
func (p *AnthropicProvider) ResponsesStreamRaw(_ context.Context, _ []byte) (io.ReadCloser, error) {
	return nil, &Error{StatusCode: 404, Message: "Anthropic does not support Responses API"} //nolint:mnd // intentional constant.
}

// AnthropicMessagesRaw passes raw request to Anthropic backend.
func (p *AnthropicProvider) AnthropicMessagesRaw(_ context.Context, rawBody []byte) ([]byte, *Usage, error) {
	respBytes, anthropicUsage, err := p.client.MessagesRaw(rawBody, nil)
	if err != nil {
		return nil, nil, err
	}
	var usage *Usage
	if anthropicUsage != nil {
		usage = anthropicUsage.ToUsage()
	}
	return respBytes, usage, nil
}

// AnthropicMessagesStreamRaw passes raw request to Anthropic backend and returns SSE stream.
func (p *AnthropicProvider) AnthropicMessagesStreamRaw(_ context.Context, rawBody []byte) (io.ReadCloser, error) {
	return p.client.MessagesStreamRaw(rawBody, nil)
}

func completionToChatRequest(req *CompletionRequest) *ChatCompletionRequest {
	var prompt string
	switch v := req.Prompt.(type) {
	case string:
		prompt = v
	case []interface{}:
		if len(v) > 0 {
			if s, ok := v[0].(string); ok {
				prompt = s
			}
		}
	}

	return &ChatCompletionRequest{
		Model:       req.Model,
		Messages:    []ChatMessage{{Role: "user", Content: prompt}},
		Stream:      req.Stream,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		Stop:        req.Stop,
	}
}

func chatToCompletionResponse(resp *ChatCompletionResponse) *CompletionResponse {
	choices := make([]CompletionChoice, 0, len(resp.Choices))
	for _, c := range resp.Choices {
		text := ""
		if c.Message != nil {
			text = c.Message.ContentString()
		}
		choices = append(choices, CompletionChoice{
			Index:        c.Index,
			Text:         text,
			FinishReason: c.FinishReason,
		})
	}
	return &CompletionResponse{
		ID:      resp.ID,
		Object:  "text_completion",
		Created: resp.Created,
		Model:   resp.Model,
		Choices: choices,
		Usage:   resp.Usage,
	}
}
