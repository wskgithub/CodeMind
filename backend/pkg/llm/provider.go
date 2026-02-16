package llm

import (
	"context"
	"io"
)

// ──────────────────────────────────
// Provider 统一接口
// 抽象不同 LLM 服务商的调用方式
// ──────────────────────────────────

// ProviderFormat LLM Provider 协议格式
type ProviderFormat string

const (
	// FormatOpenAI OpenAI 兼容格式
	FormatOpenAI ProviderFormat = "openai"
	// FormatAnthropic Anthropic 原生格式
	FormatAnthropic ProviderFormat = "anthropic"
)

// Provider LLM 服务提供者统一接口
// 封装不同格式的 LLM 服务，对外提供一致的调用方式
type Provider interface {
	// Name 返回 Provider 名称（用于日志和路由标识）
	Name() string

	// Format 返回 Provider 支持的协议格式
	Format() ProviderFormat

	// ChatCompletion OpenAI 格式 — 非流式对话补全
	ChatCompletion(ctx context.Context, req *ChatCompletionRequest) (*ChatCompletionResponse, error)

	// ChatCompletionStream OpenAI 格式 — 流式对话补全
	ChatCompletionStream(ctx context.Context, req *ChatCompletionRequest) (io.ReadCloser, error)

	// Completion OpenAI 格式 — 非流式文本补全
	Completion(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error)

	// CompletionStream OpenAI 格式 — 流式文本补全
	CompletionStream(ctx context.Context, req *CompletionRequest) (io.ReadCloser, error)

	// ListModels 获取可用模型列表
	ListModels(ctx context.Context) (*ModelListResponse, error)

	// AnthropicMessages Anthropic 格式 — 非流式消息调用
	AnthropicMessages(ctx context.Context, req *AnthropicMessagesRequest) (*AnthropicMessagesResponse, error)

	// AnthropicMessagesStream Anthropic 格式 — 流式消息调用
	AnthropicMessagesStream(ctx context.Context, req *AnthropicMessagesRequest) (io.ReadCloser, error)
}

// ──────────────────────────────────
// OpenAI Provider 实现
// ──────────────────────────────────

// OpenAIProvider 包装现有 OpenAI 兼容客户端为 Provider 接口
type OpenAIProvider struct {
	name   string
	client *Client
}

// NewOpenAIProvider 创建 OpenAI Provider
func NewOpenAIProvider(name string, client *Client) *OpenAIProvider {
	return &OpenAIProvider{
		name:   name,
		client: client,
	}
}

func (p *OpenAIProvider) Name() string          { return p.name }
func (p *OpenAIProvider) Format() ProviderFormat { return FormatOpenAI }

func (p *OpenAIProvider) ChatCompletion(_ context.Context, req *ChatCompletionRequest) (*ChatCompletionResponse, error) {
	return p.client.ChatCompletion(req)
}

func (p *OpenAIProvider) ChatCompletionStream(_ context.Context, req *ChatCompletionRequest) (io.ReadCloser, error) {
	return p.client.ChatCompletionStream(req)
}

func (p *OpenAIProvider) Completion(_ context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	return p.client.Completion(req)
}

func (p *OpenAIProvider) CompletionStream(_ context.Context, req *CompletionRequest) (io.ReadCloser, error) {
	return p.client.CompletionStream(req)
}

func (p *OpenAIProvider) ListModels(_ context.Context) (*ModelListResponse, error) {
	return p.client.ListModels()
}

// AnthropicMessages OpenAI Provider 收到 Anthropic 格式请求时，先转换再调用
func (p *OpenAIProvider) AnthropicMessages(ctx context.Context, req *AnthropicMessagesRequest) (*AnthropicMessagesResponse, error) {
	// 将 Anthropic 请求转换为 OpenAI 格式
	openaiReq := AnthropicToOpenAI(req)
	resp, err := p.client.ChatCompletion(openaiReq)
	if err != nil {
		return nil, err
	}
	// 将 OpenAI 响应转换回 Anthropic 格式
	return OpenAIResponseToAnthropic(resp), nil
}

// AnthropicMessagesStream OpenAI Provider 收到 Anthropic 流式请求时的处理
// 注意：此场景需要在 handler 层进行流式格式转换，此处返回原始 OpenAI 流
func (p *OpenAIProvider) AnthropicMessagesStream(ctx context.Context, req *AnthropicMessagesRequest) (io.ReadCloser, error) {
	openaiReq := AnthropicToOpenAI(req)
	return p.client.ChatCompletionStream(openaiReq)
}

// ──────────────────────────────────
// Anthropic Provider 实现
// ──────────────────────────────────

// AnthropicProvider 包装 Anthropic 客户端为 Provider 接口
type AnthropicProvider struct {
	name   string
	client *AnthropicClient
}

// NewAnthropicProvider 创建 Anthropic Provider
func NewAnthropicProvider(name string, client *AnthropicClient) *AnthropicProvider {
	return &AnthropicProvider{
		name:   name,
		client: client,
	}
}

func (p *AnthropicProvider) Name() string          { return p.name }
func (p *AnthropicProvider) Format() ProviderFormat { return FormatAnthropic }

// ChatCompletion Anthropic Provider 收到 OpenAI 格式请求时，先转换再调用
func (p *AnthropicProvider) ChatCompletion(_ context.Context, req *ChatCompletionRequest) (*ChatCompletionResponse, error) {
	// 将 OpenAI 请求转换为 Anthropic 格式
	anthropicReq := OpenAIToAnthropic(req)
	resp, err := p.client.Messages(anthropicReq)
	if err != nil {
		return nil, err
	}
	// 将 Anthropic 响应转换回 OpenAI 格式
	return AnthropicResponseToOpenAI(resp), nil
}

// ChatCompletionStream Anthropic Provider 收到 OpenAI 流式请求时的处理
// 注意：此场景需要在 handler 层进行流式格式转换，此处返回原始 Anthropic 流
func (p *AnthropicProvider) ChatCompletionStream(_ context.Context, req *ChatCompletionRequest) (io.ReadCloser, error) {
	anthropicReq := OpenAIToAnthropic(req)
	return p.client.MessagesStream(anthropicReq)
}

// Completion Anthropic 不支持原始 Completions 接口，转换为 Messages 调用
func (p *AnthropicProvider) Completion(_ context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	// 将 Completion 请求简单转换为 Chat 格式再通过 Anthropic 调用
	chatReq := completionToChatRequest(req)
	anthropicReq := OpenAIToAnthropic(chatReq)
	resp, err := p.client.Messages(anthropicReq)
	if err != nil {
		return nil, err
	}
	chatResp := AnthropicResponseToOpenAI(resp)
	return chatToCompletionResponse(chatResp), nil
}

// CompletionStream Anthropic 不支持原始 Completions 流式接口
func (p *AnthropicProvider) CompletionStream(_ context.Context, req *CompletionRequest) (io.ReadCloser, error) {
	chatReq := completionToChatRequest(req)
	anthropicReq := OpenAIToAnthropic(chatReq)
	return p.client.MessagesStream(anthropicReq)
}

// ListModels Anthropic 无标准模型列表接口，返回预定义列表
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

func (p *AnthropicProvider) AnthropicMessages(_ context.Context, req *AnthropicMessagesRequest) (*AnthropicMessagesResponse, error) {
	return p.client.Messages(req)
}

func (p *AnthropicProvider) AnthropicMessagesStream(_ context.Context, req *AnthropicMessagesRequest) (io.ReadCloser, error) {
	return p.client.MessagesStream(req)
}

// ──────────────────────────────────
// 辅助转换函数
// ──────────────────────────────────

// completionToChatRequest 将 Completion 请求转换为 Chat 格式
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

// chatToCompletionResponse 将 Chat 响应转换为 Completion 格式
func chatToCompletionResponse(resp *ChatCompletionResponse) *CompletionResponse {
	var choices []CompletionChoice
	for _, c := range resp.Choices {
		text := ""
		if c.Message != nil {
			text = c.Message.Content
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
