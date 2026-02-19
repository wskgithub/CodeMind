package llm

import (
	"context"
	"encoding/json"
	"fmt"
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

	// ── Chat Completions ──

	// ChatCompletion 非流式对话补全（结构体方式）
	ChatCompletion(ctx context.Context, req *ChatCompletionRequest) (*ChatCompletionResponse, error)

	// ChatCompletionStream 流式对话补全（结构体方式）
	ChatCompletionStream(ctx context.Context, req *ChatCompletionRequest) (io.ReadCloser, error)

	// ChatCompletionRaw 非流式对话补全（原始请求体透传，推荐的代理方式）
	ChatCompletionRaw(ctx context.Context, rawBody []byte) (rawResp []byte, usage *Usage, err error)

	// ChatCompletionStreamRaw 流式对话补全（原始请求体透传）
	ChatCompletionStreamRaw(ctx context.Context, rawBody []byte) (io.ReadCloser, error)

	// ── Completions ──

	// Completion 非流式文本补全（结构体方式）
	Completion(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error)

	// CompletionStream 流式文本补全（结构体方式）
	CompletionStream(ctx context.Context, req *CompletionRequest) (io.ReadCloser, error)

	// CompletionRaw 非流式文本补全（原始请求体透传）
	CompletionRaw(ctx context.Context, rawBody []byte) (rawResp []byte, usage *Usage, err error)

	// CompletionStreamRaw 流式文本补全（原始请求体透传）
	CompletionStreamRaw(ctx context.Context, rawBody []byte) (io.ReadCloser, error)

	// ── Models ──

	// ListModels 获取可用模型列表
	ListModels(ctx context.Context) (*ModelListResponse, error)

	// RetrieveModel 获取单个模型信息
	RetrieveModel(ctx context.Context, modelID string) (*ModelInfo, error)

	// ── Embeddings ──

	// EmbeddingRaw 向量嵌入（原始请求体透传）
	EmbeddingRaw(ctx context.Context, rawBody []byte) (rawResp []byte, usage *Usage, err error)

	// ── Responses API ──

	// ResponsesRaw 非流式 Responses API（原始请求体透传）
	ResponsesRaw(ctx context.Context, rawBody []byte) (rawResp []byte, usage *Usage, err error)

	// ResponsesStreamRaw 流式 Responses API（原始请求体透传）
	ResponsesStreamRaw(ctx context.Context, rawBody []byte) (io.ReadCloser, error)

	// ── Anthropic 原生接口（原始请求体透传） ──

	// AnthropicMessagesRaw 非流式消息调用（原始请求体透传）
	AnthropicMessagesRaw(ctx context.Context, rawBody []byte) (rawResp []byte, usage *Usage, err error)

	// AnthropicMessagesStreamRaw 流式消息调用（原始请求体透传）
	AnthropicMessagesStreamRaw(ctx context.Context, rawBody []byte) (io.ReadCloser, error)
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

func (p *OpenAIProvider) ChatCompletionRaw(_ context.Context, rawBody []byte) ([]byte, *Usage, error) {
	return p.client.ChatCompletionRawAll(rawBody)
}

func (p *OpenAIProvider) ChatCompletionStreamRaw(_ context.Context, rawBody []byte) (io.ReadCloser, error) {
	return p.client.ChatCompletionStreamRaw(rawBody)
}

func (p *OpenAIProvider) Completion(_ context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	return p.client.Completion(req)
}

func (p *OpenAIProvider) CompletionStream(_ context.Context, req *CompletionRequest) (io.ReadCloser, error) {
	return p.client.CompletionStream(req)
}

func (p *OpenAIProvider) CompletionRaw(_ context.Context, rawBody []byte) ([]byte, *Usage, error) {
	return p.client.CompletionRawAll(rawBody)
}

func (p *OpenAIProvider) CompletionStreamRaw(_ context.Context, rawBody []byte) (io.ReadCloser, error) {
	return p.client.CompletionStreamRaw(rawBody)
}

func (p *OpenAIProvider) ListModels(_ context.Context) (*ModelListResponse, error) {
	return p.client.ListModels()
}

func (p *OpenAIProvider) RetrieveModel(_ context.Context, modelID string) (*ModelInfo, error) {
	return p.client.RetrieveModel(modelID)
}

func (p *OpenAIProvider) EmbeddingRaw(_ context.Context, rawBody []byte) ([]byte, *Usage, error) {
	return p.client.EmbeddingRaw(rawBody)
}

func (p *OpenAIProvider) ResponsesRaw(_ context.Context, rawBody []byte) ([]byte, *Usage, error) {
	return p.client.ResponsesRaw(rawBody)
}

func (p *OpenAIProvider) ResponsesStreamRaw(_ context.Context, rawBody []byte) (io.ReadCloser, error) {
	return p.client.ResponsesStreamRaw(rawBody)
}

// AnthropicMessagesRaw OpenAI Provider 收到 Anthropic 原始请求时，解析→转换→调用→转换回
func (p *OpenAIProvider) AnthropicMessagesRaw(_ context.Context, rawBody []byte) ([]byte, *Usage, error) {
	var req AnthropicMessagesRequest
	if err := json.Unmarshal(rawBody, &req); err != nil {
		return nil, nil, fmt.Errorf("解析 Anthropic 请求体失败: %w", err)
	}
	openaiReq := AnthropicToOpenAI(&req)
	resp, err := p.client.ChatCompletion(openaiReq)
	if err != nil {
		return nil, nil, err
	}
	anthropicResp := OpenAIResponseToAnthropic(resp)
	data, err := json.Marshal(anthropicResp)
	if err != nil {
		return nil, nil, fmt.Errorf("序列化 Anthropic 响应失败: %w", err)
	}
	return data, resp.Usage, nil
}

// AnthropicMessagesStreamRaw OpenAI Provider 收到 Anthropic 流式请求时，返回 OpenAI 格式流
// handler 层根据 provider.Format() 进行格式转换
func (p *OpenAIProvider) AnthropicMessagesStreamRaw(_ context.Context, rawBody []byte) (io.ReadCloser, error) {
	var req AnthropicMessagesRequest
	if err := json.Unmarshal(rawBody, &req); err != nil {
		return nil, fmt.Errorf("解析 Anthropic 请求体失败: %w", err)
	}
	openaiReq := AnthropicToOpenAI(&req)
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
	anthropicReq := OpenAIToAnthropic(req)
	resp, err := p.client.Messages(anthropicReq)
	if err != nil {
		return nil, err
	}
	return AnthropicResponseToOpenAI(resp), nil
}

// ChatCompletionStream Anthropic Provider 收到 OpenAI 流式请求时的处理
// 此处返回原始 Anthropic 流，handler 层负责格式转换
func (p *AnthropicProvider) ChatCompletionStream(_ context.Context, req *ChatCompletionRequest) (io.ReadCloser, error) {
	anthropicReq := OpenAIToAnthropic(req)
	return p.client.MessagesStream(anthropicReq)
}

// ChatCompletionRaw Anthropic Provider 收到原始 OpenAI 请求时，需解析→转换→调用→转换
// 跨格式转换会丢失部分 OpenAI 特有字段，这是协议差异的固有限制
func (p *AnthropicProvider) ChatCompletionRaw(ctx context.Context, rawBody []byte) ([]byte, *Usage, error) {
	var req ChatCompletionRequest
	if err := json.Unmarshal(rawBody, &req); err != nil {
		return nil, nil, fmt.Errorf("解析请求体失败: %w", err)
	}
	resp, err := p.ChatCompletion(ctx, &req)
	if err != nil {
		return nil, nil, err
	}
	data, err := json.Marshal(resp)
	if err != nil {
		return nil, nil, fmt.Errorf("序列化响应失败: %w", err)
	}
	return data, resp.Usage, nil
}

// ChatCompletionStreamRaw Anthropic Provider 收到原始 OpenAI 流式请求时的处理
func (p *AnthropicProvider) ChatCompletionStreamRaw(ctx context.Context, rawBody []byte) (io.ReadCloser, error) {
	var req ChatCompletionRequest
	if err := json.Unmarshal(rawBody, &req); err != nil {
		return nil, fmt.Errorf("解析请求体失败: %w", err)
	}
	return p.ChatCompletionStream(ctx, &req)
}

// Completion Anthropic 不支持原始 Completions 接口，转换为 Messages 调用
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

// CompletionStream Anthropic 不支持原始 Completions 流式接口
func (p *AnthropicProvider) CompletionStream(_ context.Context, req *CompletionRequest) (io.ReadCloser, error) {
	chatReq := completionToChatRequest(req)
	anthropicReq := OpenAIToAnthropic(chatReq)
	return p.client.MessagesStream(anthropicReq)
}

// CompletionRaw Anthropic Provider 的 Completions 原始透传
func (p *AnthropicProvider) CompletionRaw(ctx context.Context, rawBody []byte) ([]byte, *Usage, error) {
	var req CompletionRequest
	if err := json.Unmarshal(rawBody, &req); err != nil {
		return nil, nil, fmt.Errorf("解析请求体失败: %w", err)
	}
	resp, err := p.Completion(ctx, &req)
	if err != nil {
		return nil, nil, err
	}
	data, err := json.Marshal(resp)
	if err != nil {
		return nil, nil, fmt.Errorf("序列化响应失败: %w", err)
	}
	return data, resp.Usage, nil
}

// CompletionStreamRaw Anthropic Provider 的 Completions 流式原始透传
func (p *AnthropicProvider) CompletionStreamRaw(ctx context.Context, rawBody []byte) (io.ReadCloser, error) {
	var req CompletionRequest
	if err := json.Unmarshal(rawBody, &req); err != nil {
		return nil, fmt.Errorf("解析请求体失败: %w", err)
	}
	return p.CompletionStream(ctx, &req)
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

// RetrieveModel Anthropic 不支持查询单个模型，从预定义列表中查找
func (p *AnthropicProvider) RetrieveModel(_ context.Context, modelID string) (*ModelInfo, error) {
	resp, _ := p.ListModels(nil)
	for _, m := range resp.Data {
		if m.ID == modelID {
			return &m, nil
		}
	}
	return nil, &LLMError{StatusCode: 404, Message: fmt.Sprintf("模型 '%s' 不存在", modelID)}
}

// EmbeddingRaw Anthropic 不支持 Embeddings 接口
func (p *AnthropicProvider) EmbeddingRaw(_ context.Context, _ []byte) ([]byte, *Usage, error) {
	return nil, nil, &LLMError{StatusCode: 404, Message: "Anthropic 不支持 Embeddings 接口"}
}

func (p *AnthropicProvider) ResponsesRaw(_ context.Context, _ []byte) ([]byte, *Usage, error) {
	return nil, nil, &LLMError{StatusCode: 404, Message: "Anthropic 不支持 Responses 接口"}
}

func (p *AnthropicProvider) ResponsesStreamRaw(_ context.Context, _ []byte) (io.ReadCloser, error) {
	return nil, &LLMError{StatusCode: 404, Message: "Anthropic 不支持 Responses 接口"}
}

// AnthropicMessagesRaw 直接透传原始请求体到 Anthropic 后端
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

// AnthropicMessagesStreamRaw 直接透传原始请求体到 Anthropic 后端，返回 SSE 流
func (p *AnthropicProvider) AnthropicMessagesStreamRaw(_ context.Context, rawBody []byte) (io.ReadCloser, error) {
	return p.client.MessagesStreamRaw(rawBody, nil)
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
