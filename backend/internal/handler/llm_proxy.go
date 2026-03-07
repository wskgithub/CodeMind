package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"codemind/internal/middleware"
	"codemind/internal/pkg/errcode"
	"codemind/internal/service"
	"codemind/pkg/llm"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// LLMProxyHandler LLM 代理控制器
type LLMProxyHandler struct {
	proxyService *service.LLMProxyService
	logger       *zap.Logger
}

// NewLLMProxyHandler 创建 LLM 代理 Handler
func NewLLMProxyHandler(proxyService *service.LLMProxyService, logger *zap.Logger) *LLMProxyHandler {
	return &LLMProxyHandler{
		proxyService: proxyService,
		logger:       logger,
	}
}

// ──────────────────────────────────
// OpenAI Chat Completions
// ──────────────────────────────────

// ChatCompletions 对话补全代理（OpenAI 格式）
// POST /v1/chat/completions
//
// 采用「原始请求体透传」模式：
//   - 只解析路由所需的最小字段（model、stream）
//   - 将完整的原始 JSON 转发给 LLM，确保 tools、stream_options 等所有字段保留
//   - 流式请求自动注入 stream_options.include_usage 以获取用量信息
func (h *LLMProxyHandler) ChatCompletions(c *gin.Context) {
	startTime := time.Now()
	userID := middleware.GetUserID(c)
	keyID, _ := c.Get(middleware.CtxKeyAPIKeyID)
	apiKeyID := keyID.(int64)
	deptID := middleware.GetDepartmentID(c)

	rawBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.sendOpenAIError(c, http.StatusBadRequest, "invalid_request_error", "请求体读取失败")
		return
	}

	// 清理对话历史中的 thinking 内容（Thinking 模型优化）
	rawBody = llm.CleanThinkingFromHistory(rawBody)

	meta, err := llm.ExtractRequestMeta(rawBody)
	if err != nil {
		h.sendOpenAIError(c, http.StatusBadRequest, "invalid_request_error", "请求体解析失败")
		return
	}

	allowed, err := h.proxyService.CheckTokenQuota(c.Request.Context(), userID, deptID)
	if err != nil {
		h.logger.Error("配额检查失败", zap.Error(err))
	}
	if !allowed {
		h.sendOpenAIError(c, http.StatusTooManyRequests, "rate_limit_exceeded", errcode.ErrTokenQuotaExceeded.Message)
		return
	}

	acquired, err := h.proxyService.AcquireConcurrency(c.Request.Context(), userID, deptID)
	if err != nil {
		h.logger.Error("并发控制失败", zap.Error(err))
	}
	if !acquired {
		h.sendOpenAIError(c, http.StatusTooManyRequests, "rate_limit_exceeded", errcode.ErrConcurrencyExceeded.Message)
		return
	}
	defer h.proxyService.ReleaseConcurrency(c.Request.Context(), userID)

	modelName := meta.Model
	ctx := c.Request.Context()
	provider, err := h.proxyService.GetProviderForModel(ctx, userID, modelName)
	if err != nil {
		h.logger.Error("获取 Provider 失败", zap.Error(err), zap.String("model", modelName))
		h.sendOpenAIError(c, http.StatusServiceUnavailable, "server_error", "没有可用的 LLM Provider")
		return
	}

	h.logger.Debug("路由请求到 Provider",
		zap.String("model", modelName),
		zap.String("provider", provider.Name()),
		zap.String("format", string(provider.Format())),
	)

	if meta.Stream {
		h.handleStreamChatRaw(c, ctx, provider, rawBody, userID, apiKeyID, deptID, modelName, startTime)
	} else {
		h.handleNonStreamChatRaw(c, ctx, provider, rawBody, userID, apiKeyID, deptID, modelName, startTime)
	}
}

// handleNonStreamChatRaw 非流式对话 — 原始请求体透传
func (h *LLMProxyHandler) handleNonStreamChatRaw(
	c *gin.Context, ctx context.Context, provider llm.Provider,
	rawBody []byte,
	userID, apiKeyID int64, deptID *int64, modelName string, startTime time.Time,
) {
	rawResp, usage, err := provider.ChatCompletionRaw(ctx, rawBody)
	durationMs := int(time.Since(startTime).Milliseconds())

	if err != nil {
		h.handleLLMError(c, err, userID, apiKeyID, modelName, "chat_completion", durationMs)
		return
	}

	go h.proxyService.RecordUsage(userID, apiKeyID, deptID, modelName, "chat_completion", usage, durationMs)
	go h.proxyService.RecordRequestLog(userID, apiKeyID, "chat_completion", modelName, 200, "", c.ClientIP(), c.Request.UserAgent(), durationMs)
	go h.proxyService.RecordTrainingData(userID, apiKeyID, "chat_completion", modelName, false, rawBody, rawResp, usage, 200, durationMs, c.ClientIP())

	c.Data(http.StatusOK, "application/json", rawResp)
}

// handleStreamChatRaw 流式对话（SSE）— 原始请求体透传
func (h *LLMProxyHandler) handleStreamChatRaw(
	c *gin.Context, ctx context.Context, provider llm.Provider,
	rawBody []byte,
	userID, apiKeyID int64, deptID *int64, modelName string, startTime time.Time,
) {
	body, err := provider.ChatCompletionStreamRaw(ctx, rawBody)
	if err != nil {
		durationMs := int(time.Since(startTime).Milliseconds())
		h.handleLLMError(c, err, userID, apiKeyID, modelName, "chat_completion", durationMs)
		return
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")
	c.Status(http.StatusOK)

	var streamResult *llm.StreamResult

	if provider.Format() == llm.FormatAnthropic {
		streamResult = h.pipeAnthropicStreamToOpenAI(c, body, modelName)
	} else {
		streamResult = h.pipeOpenAIStream(c, body)
	}

	durationMs := int(time.Since(startTime).Milliseconds())
	go h.proxyService.RecordUsage(userID, apiKeyID, deptID, modelName, "chat_completion", streamResult.Usage, durationMs)
	go h.proxyService.RecordRequestLog(userID, apiKeyID, "chat_completion", modelName, 200, "", c.ClientIP(), c.Request.UserAgent(), durationMs)
	go h.proxyService.RecordTrainingData(userID, apiKeyID, "chat_completion", modelName, true, rawBody, llm.AssembleChatResponse(streamResult), streamResult.Usage, 200, durationMs, c.ClientIP())
}

// ──────────────────────────────────
// OpenAI Completions
// ──────────────────────────────────

// Completions 文本补全代理（OpenAI 格式）
// POST /v1/completions
//
// 同样采用「原始请求体透传」模式，确保 suffix、logprobs 等所有字段完整保留
func (h *LLMProxyHandler) Completions(c *gin.Context) {
	startTime := time.Now()
	userID := middleware.GetUserID(c)
	keyID, _ := c.Get(middleware.CtxKeyAPIKeyID)
	apiKeyID := keyID.(int64)
	deptID := middleware.GetDepartmentID(c)

	rawBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.sendOpenAIError(c, http.StatusBadRequest, "invalid_request_error", "请求体读取失败")
		return
	}

	meta, err := llm.ExtractRequestMeta(rawBody)
	if err != nil {
		h.sendOpenAIError(c, http.StatusBadRequest, "invalid_request_error", "请求体解析失败")
		return
	}

	allowed, _ := h.proxyService.CheckTokenQuota(c.Request.Context(), userID, deptID)
	if !allowed {
		h.sendOpenAIError(c, http.StatusTooManyRequests, "rate_limit_exceeded", errcode.ErrTokenQuotaExceeded.Message)
		return
	}

	acquired, _ := h.proxyService.AcquireConcurrency(c.Request.Context(), userID, deptID)
	if !acquired {
		h.sendOpenAIError(c, http.StatusTooManyRequests, "rate_limit_exceeded", errcode.ErrConcurrencyExceeded.Message)
		return
	}
	defer h.proxyService.ReleaseConcurrency(c.Request.Context(), userID)

	modelName := meta.Model
	ctx := c.Request.Context()
	provider, err := h.proxyService.GetProviderForModel(ctx, userID, modelName)
	if err != nil {
		h.sendOpenAIError(c, http.StatusServiceUnavailable, "server_error", "没有可用的 LLM Provider")
		return
	}

	if meta.Stream {
		h.handleStreamCompletionRaw(c, ctx, provider, rawBody, userID, apiKeyID, deptID, modelName, startTime)
	} else {
		h.handleNonStreamCompletionRaw(c, ctx, provider, rawBody, userID, apiKeyID, deptID, modelName, startTime)
	}
}

// handleNonStreamCompletionRaw 非流式文本补全 — 原始请求体透传
func (h *LLMProxyHandler) handleNonStreamCompletionRaw(
	c *gin.Context, ctx context.Context, provider llm.Provider,
	rawBody []byte,
	userID, apiKeyID int64, deptID *int64, modelName string, startTime time.Time,
) {
	rawResp, usage, err := provider.CompletionRaw(ctx, rawBody)
	durationMs := int(time.Since(startTime).Milliseconds())

	if err != nil {
		h.handleLLMError(c, err, userID, apiKeyID, modelName, "completion", durationMs)
		return
	}

	go h.proxyService.RecordUsage(userID, apiKeyID, deptID, modelName, "completion", usage, durationMs)
	go h.proxyService.RecordRequestLog(userID, apiKeyID, "completion", modelName, 200, "", c.ClientIP(), c.Request.UserAgent(), durationMs)
	go h.proxyService.RecordTrainingData(userID, apiKeyID, "completion", modelName, false, rawBody, rawResp, usage, 200, durationMs, c.ClientIP())

	c.Data(http.StatusOK, "application/json", rawResp)
}

// handleStreamCompletionRaw 流式文本补全 — 原始请求体透传
func (h *LLMProxyHandler) handleStreamCompletionRaw(
	c *gin.Context, ctx context.Context, provider llm.Provider,
	rawBody []byte,
	userID, apiKeyID int64, deptID *int64, modelName string, startTime time.Time,
) {
	body, err := provider.CompletionStreamRaw(ctx, rawBody)
	if err != nil {
		durationMs := int(time.Since(startTime).Milliseconds())
		h.handleLLMError(c, err, userID, apiKeyID, modelName, "completion", durationMs)
		return
	}
	defer body.Close()

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")
	c.Status(http.StatusOK)

	// Completions 流直接透传原始 SSE 数据
	result := &llm.StreamResult{}
	var contentBuilder strings.Builder
	reader := llm.NewStreamReader(body)
	defer reader.Close()

	for {
		rawLine, chunk, err := reader.ReadChunk()
		if err != nil {
			if err == io.EOF {
				fmt.Fprintf(c.Writer, "data: [DONE]\n\n")
				c.Writer.Flush()
				break
			}
			h.logger.Error("读取 Completions SSE 流失败", zap.Error(err))
			break
		}

		fmt.Fprintf(c.Writer, "%s\n\n", rawLine)
		c.Writer.Flush()

		if chunk != nil {
			if result.ResponseID == "" {
				result.ResponseID = chunk.ID
				result.Model = chunk.Model
			}
			if len(chunk.Choices) > 0 && chunk.Choices[0].FinishReason != nil {
				result.FinishReason = *chunk.Choices[0].FinishReason
			}
			if chunk.Usage != nil {
				result.Usage = chunk.Usage
			}
		}
	}
	result.Content = contentBuilder.String()

	durationMs := int(time.Since(startTime).Milliseconds())
	go h.proxyService.RecordUsage(userID, apiKeyID, deptID, modelName, "completion", result.Usage, durationMs)
	go h.proxyService.RecordRequestLog(userID, apiKeyID, "completion", modelName, 200, "", c.ClientIP(), c.Request.UserAgent(), durationMs)
	go h.proxyService.RecordTrainingData(userID, apiKeyID, "completion", modelName, true, rawBody, llm.AssembleCompletionResponse(result), result.Usage, 200, durationMs, c.ClientIP())
}

// ──────────────────────────────────
// OpenAI Models
// ──────────────────────────────────

// ListModels 获取可用模型列表
// GET /v1/models
func (h *LLMProxyHandler) ListModels(c *gin.Context) {
	provider, err := h.proxyService.GetProviderManager().GetDefault()
	if err != nil {
		h.sendOpenAIError(c, http.StatusBadGateway, "server_error", "获取模型列表失败")
		return
	}
	resp, err := provider.ListModels(c.Request.Context())
	if err != nil {
		h.sendOpenAIError(c, http.StatusBadGateway, "server_error", "获取模型列表失败")
		return
	}
	c.JSON(http.StatusOK, resp)
}

// RetrieveModel 获取单个模型信息
// GET /v1/models/:model
func (h *LLMProxyHandler) RetrieveModel(c *gin.Context) {
	modelID := c.Param("model")
	if modelID == "" {
		h.sendOpenAIError(c, http.StatusBadRequest, "invalid_request_error", "模型 ID 不能为空")
		return
	}

	provider, err := h.proxyService.GetProviderManager().GetDefault()
	if err != nil {
		h.sendOpenAIError(c, http.StatusBadGateway, "server_error", "获取模型信息失败")
		return
	}

	resp, err := provider.RetrieveModel(c.Request.Context(), modelID)
	if err != nil {
		if llmErr, ok := err.(*llm.LLMError); ok && llmErr.StatusCode == 404 {
			h.sendOpenAIError(c, http.StatusNotFound, "invalid_request_error",
				fmt.Sprintf("The model '%s' does not exist", modelID))
			return
		}
		h.sendOpenAIError(c, http.StatusBadGateway, "server_error", "获取模型信息失败")
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ──────────────────────────────────
// OpenAI Embeddings
// ──────────────────────────────────

// Embeddings 向量嵌入代理
// POST /v1/embeddings
func (h *LLMProxyHandler) Embeddings(c *gin.Context) {
	startTime := time.Now()
	userID := middleware.GetUserID(c)
	keyID, _ := c.Get(middleware.CtxKeyAPIKeyID)
	apiKeyID := keyID.(int64)
	deptID := middleware.GetDepartmentID(c)

	rawBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.sendOpenAIError(c, http.StatusBadRequest, "invalid_request_error", "请求体读取失败")
		return
	}

	meta, err := llm.ExtractRequestMeta(rawBody)
	if err != nil {
		h.sendOpenAIError(c, http.StatusBadRequest, "invalid_request_error", "请求体解析失败")
		return
	}

	allowed, _ := h.proxyService.CheckTokenQuota(c.Request.Context(), userID, deptID)
	if !allowed {
		h.sendOpenAIError(c, http.StatusTooManyRequests, "rate_limit_exceeded", errcode.ErrTokenQuotaExceeded.Message)
		return
	}

	modelName := meta.Model
	ctx := c.Request.Context()
	provider, err := h.proxyService.GetProviderForModel(ctx, userID, modelName)
	if err != nil {
		h.sendOpenAIError(c, http.StatusServiceUnavailable, "server_error", "没有可用的 LLM Provider")
		return
	}

	rawResp, usage, err := provider.EmbeddingRaw(ctx, rawBody)
	durationMs := int(time.Since(startTime).Milliseconds())

	if err != nil {
		h.handleLLMError(c, err, userID, apiKeyID, modelName, "embedding", durationMs)
		return
	}

	// Embeddings 只有 prompt_tokens 和 total_tokens，映射到通用 Usage
	if usage != nil {
		go h.proxyService.RecordUsage(userID, apiKeyID, deptID, modelName, "embedding", usage, durationMs)
	}
	go h.proxyService.RecordRequestLog(userID, apiKeyID, "embedding", modelName, 200, "", c.ClientIP(), c.Request.UserAgent(), durationMs)
	// Embedding 响应体包含大量浮点向量，对训练无意义，仅记录请求体
	go h.proxyService.RecordTrainingData(userID, apiKeyID, "embedding", modelName, false, rawBody, nil, usage, 200, durationMs, c.ClientIP())

	c.Data(http.StatusOK, "application/json", rawResp)
}

// ──────────────────────────────────
// OpenAI Responses API
// ──────────────────────────────────

// Responses 代理（OpenAI Responses API）
// POST /v1/responses
//
// Responses API 是 OpenAI 2025 年引入的新接口，与 Chat Completions 并存。
// 采用原始请求体透传模式；流式输出使用命名事件格式（event: xxx\ndata: {...}），
// 而非 Chat Completions 的 data-only 格式。
func (h *LLMProxyHandler) Responses(c *gin.Context) {
	startTime := time.Now()
	userID := middleware.GetUserID(c)
	keyID, _ := c.Get(middleware.CtxKeyAPIKeyID)
	apiKeyID := keyID.(int64)
	deptID := middleware.GetDepartmentID(c)

	rawBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.sendResponsesError(c, http.StatusBadRequest, "invalid_request_error", "请求体读取失败")
		return
	}

	meta, err := llm.ExtractRequestMeta(rawBody)
	if err != nil {
		h.sendResponsesError(c, http.StatusBadRequest, "invalid_request_error", "请求体解析失败")
		return
	}

	allowed, err := h.proxyService.CheckTokenQuota(c.Request.Context(), userID, deptID)
	if err != nil {
		h.logger.Error("配额检查失败", zap.Error(err))
	}
	if !allowed {
		h.sendResponsesError(c, http.StatusTooManyRequests, "rate_limit_exceeded", errcode.ErrTokenQuotaExceeded.Message)
		return
	}

	acquired, err := h.proxyService.AcquireConcurrency(c.Request.Context(), userID, deptID)
	if err != nil {
		h.logger.Error("并发控制失败", zap.Error(err))
	}
	if !acquired {
		h.sendResponsesError(c, http.StatusTooManyRequests, "rate_limit_exceeded", errcode.ErrConcurrencyExceeded.Message)
		return
	}
	defer h.proxyService.ReleaseConcurrency(c.Request.Context(), userID)

	modelName := meta.Model
	ctx := c.Request.Context()
	provider, err := h.proxyService.GetProviderForModel(ctx, userID, modelName)
	if err != nil {
		h.logger.Error("获取 Provider 失败", zap.Error(err), zap.String("model", modelName))
		h.sendResponsesError(c, http.StatusServiceUnavailable, "server_error", "没有可用的 LLM Provider")
		return
	}

	h.logger.Debug("Responses API 路由请求到 Provider",
		zap.String("model", modelName),
		zap.String("provider", provider.Name()),
	)

	if meta.Stream {
		h.handleStreamResponses(c, ctx, provider, rawBody, userID, apiKeyID, deptID, modelName, startTime)
	} else {
		h.handleNonStreamResponses(c, ctx, provider, rawBody, userID, apiKeyID, deptID, modelName, startTime)
	}
}

// handleNonStreamResponses 非流式 Responses API
func (h *LLMProxyHandler) handleNonStreamResponses(
	c *gin.Context, ctx context.Context, provider llm.Provider,
	rawBody []byte,
	userID, apiKeyID int64, deptID *int64, modelName string, startTime time.Time,
) {
	rawResp, usage, err := provider.ResponsesRaw(ctx, rawBody)
	durationMs := int(time.Since(startTime).Milliseconds())

	if err != nil {
		h.handleLLMErrorResponses(c, err, userID, apiKeyID, modelName, "responses", durationMs)
		return
	}

	go h.proxyService.RecordUsage(userID, apiKeyID, deptID, modelName, "responses", usage, durationMs)
	go h.proxyService.RecordRequestLog(userID, apiKeyID, "responses", modelName, 200, "", c.ClientIP(), c.Request.UserAgent(), durationMs)
	go h.proxyService.RecordTrainingData(userID, apiKeyID, "responses", modelName, false, rawBody, rawResp, usage, 200, durationMs, c.ClientIP())

	c.Data(http.StatusOK, "application/json", rawResp)
}

// handleStreamResponses 流式 Responses API（SSE 命名事件格式）
func (h *LLMProxyHandler) handleStreamResponses(
	c *gin.Context, ctx context.Context, provider llm.Provider,
	rawBody []byte,
	userID, apiKeyID int64, deptID *int64, modelName string, startTime time.Time,
) {
	body, err := provider.ResponsesStreamRaw(ctx, rawBody)
	if err != nil {
		durationMs := int(time.Since(startTime).Milliseconds())
		h.handleLLMErrorResponses(c, err, userID, apiKeyID, modelName, "responses", durationMs)
		return
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")
	c.Status(http.StatusOK)

	streamResult := h.pipeResponsesStream(c, body)

	durationMs := int(time.Since(startTime).Milliseconds())
	go h.proxyService.RecordUsage(userID, apiKeyID, deptID, modelName, "responses", streamResult.Usage, durationMs)
	go h.proxyService.RecordRequestLog(userID, apiKeyID, "responses", modelName, 200, "", c.ClientIP(), c.Request.UserAgent(), durationMs)
	go h.proxyService.RecordTrainingData(userID, apiKeyID, "responses", modelName, true, rawBody, nil, streamResult.Usage, 200, durationMs, c.ClientIP())
}

// ──────────────────────────────────
// Anthropic 格式端点
// ──────────────────────────────────

// AnthropicMessages Anthropic 原生消息代理
// POST /v1/messages
// 采用原始请求体透传，避免结构体反序列化丢失 thinking 等未定义字段
func (h *LLMProxyHandler) AnthropicMessages(c *gin.Context) {
	startTime := time.Now()
	userID := middleware.GetUserID(c)
	keyID, _ := c.Get(middleware.CtxKeyAPIKeyID)
	apiKeyID := keyID.(int64)
	deptID := middleware.GetDepartmentID(c)

	rawBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.sendAnthropicError(c, http.StatusBadRequest, "invalid_request_error", "读取请求体失败")
		return
	}

	// 轻量解析：只提取路由和分流所需的最小字段
	var partial struct {
		Model  string `json:"model"`
		Stream bool   `json:"stream"`
	}
	if err := json.Unmarshal(rawBody, &partial); err != nil {
		h.sendAnthropicError(c, http.StatusBadRequest, "invalid_request_error", "请求体 JSON 格式无效")
		return
	}
	if partial.Model == "" {
		h.sendAnthropicError(c, http.StatusBadRequest, "invalid_request_error", "缺少必填字段 model")
		return
	}

	allowed, err := h.proxyService.CheckTokenQuota(c.Request.Context(), userID, deptID)
	if err != nil {
		h.logger.Error("配额检查失败", zap.Error(err))
	}
	if !allowed {
		h.sendAnthropicError(c, http.StatusTooManyRequests, "rate_limit_error", errcode.ErrTokenQuotaExceeded.Message)
		return
	}

	acquired, err := h.proxyService.AcquireConcurrency(c.Request.Context(), userID, deptID)
	if err != nil {
		h.logger.Error("并发控制失败", zap.Error(err))
	}
	if !acquired {
		h.sendAnthropicError(c, http.StatusTooManyRequests, "rate_limit_error", errcode.ErrConcurrencyExceeded.Message)
		return
	}
	defer h.proxyService.ReleaseConcurrency(c.Request.Context(), userID)

	modelName := partial.Model
	ctx := c.Request.Context()
	provider, err := h.proxyService.GetProviderForModel(ctx, userID, modelName)
	if err != nil {
		h.sendAnthropicError(c, http.StatusServiceUnavailable, "api_error", "没有可用的 LLM Provider")
		return
	}

	h.logger.Debug("Anthropic 请求路由到 Provider",
		zap.String("model", modelName),
		zap.String("provider", provider.Name()),
		zap.String("format", string(provider.Format())),
	)

	if partial.Stream {
		h.handleAnthropicStream(c, ctx, provider, rawBody, userID, apiKeyID, deptID, modelName, startTime)
	} else {
		h.handleAnthropicNonStream(c, ctx, provider, rawBody, userID, apiKeyID, deptID, modelName, startTime)
	}
}

// handleAnthropicNonStream 处理 Anthropic 非流式请求（原始请求体透传）
func (h *LLMProxyHandler) handleAnthropicNonStream(
	c *gin.Context, ctx context.Context, provider llm.Provider,
	rawBody []byte,
	userID, apiKeyID int64, deptID *int64, modelName string, startTime time.Time,
) {
	rawResp, usage, err := provider.AnthropicMessagesRaw(ctx, rawBody)
	durationMs := int(time.Since(startTime).Milliseconds())

	if err != nil {
		h.handleLLMErrorAnthropic(c, err, userID, apiKeyID, modelName, "anthropic_messages", durationMs)
		return
	}

	go h.proxyService.RecordUsage(userID, apiKeyID, deptID, modelName, "anthropic_messages", usage, durationMs)
	go h.proxyService.RecordRequestLog(userID, apiKeyID, "anthropic_messages", modelName, 200, "", c.ClientIP(), c.Request.UserAgent(), durationMs)
	go h.proxyService.RecordTrainingData(userID, apiKeyID, "anthropic_messages", modelName, false, rawBody, rawResp, usage, 200, durationMs, c.ClientIP())

	c.Data(http.StatusOK, "application/json", rawResp)
}

// handleAnthropicStream 处理 Anthropic 流式请求（原始请求体透传）
func (h *LLMProxyHandler) handleAnthropicStream(
	c *gin.Context, ctx context.Context, provider llm.Provider,
	rawBody []byte,
	userID, apiKeyID int64, deptID *int64, modelName string, startTime time.Time,
) {
	body, err := provider.AnthropicMessagesStreamRaw(ctx, rawBody)
	if err != nil {
		durationMs := int(time.Since(startTime).Milliseconds())
		h.handleLLMErrorAnthropic(c, err, userID, apiKeyID, modelName, "anthropic_messages", durationMs)
		return
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")
	c.Status(http.StatusOK)

	var streamResult *llm.StreamResult

	if provider.Format() == llm.FormatOpenAI {
		streamResult = h.pipeOpenAIStreamToAnthropic(c, body, modelName)
	} else {
		streamResult = h.pipeAnthropicStream(c, body)
	}

	durationMs := int(time.Since(startTime).Milliseconds())
	go h.proxyService.RecordUsage(userID, apiKeyID, deptID, modelName, "anthropic_messages", streamResult.Usage, durationMs)
	go h.proxyService.RecordRequestLog(userID, apiKeyID, "anthropic_messages", modelName, 200, "", c.ClientIP(), c.Request.UserAgent(), durationMs)
	go h.proxyService.RecordTrainingData(userID, apiKeyID, "anthropic_messages", modelName, true, rawBody, llm.AssembleChatResponse(streamResult), streamResult.Usage, 200, durationMs, c.ClientIP())
}

// ──────────────────────────────────
// 流式数据管道
// ──────────────────────────────────

// pipeOpenAIStream 直接转发 OpenAI 格式流，同时收集完整响应内容
func (h *LLMProxyHandler) pipeOpenAIStream(c *gin.Context, body io.ReadCloser) *llm.StreamResult {
	reader := llm.NewStreamReader(body)
	defer reader.Close()

	result := &llm.StreamResult{}
	var contentBuilder strings.Builder

	for {
		rawLine, chunk, err := reader.ReadChunk()
		if err != nil {
			if err == io.EOF {
				fmt.Fprintf(c.Writer, "data: [DONE]\n\n")
				c.Writer.Flush()
				break
			}
			h.logger.Error("读取 OpenAI SSE 流失败", zap.Error(err))
			break
		}

		fmt.Fprintf(c.Writer, "%s\n\n", rawLine)
		c.Writer.Flush()

		if chunk != nil {
			if result.ResponseID == "" {
				result.ResponseID = chunk.ID
				result.Model = chunk.Model
			}
			if len(chunk.Choices) > 0 {
				delta := chunk.Choices[0].Delta
				if delta != nil {
					contentBuilder.WriteString(delta.ContentString())
				}
				if chunk.Choices[0].FinishReason != nil {
					result.FinishReason = *chunk.Choices[0].FinishReason
				}
			}
			if chunk.Usage != nil {
				result.Usage = chunk.Usage
			}
		}
	}

	result.Content = contentBuilder.String()
	return result
}

// pipeAnthropicStream 直接转发 Anthropic 格式流，同时收集完整响应内容
func (h *LLMProxyHandler) pipeAnthropicStream(c *gin.Context, body io.ReadCloser) *llm.StreamResult {
	reader := llm.NewAnthropicStreamReader(body)
	defer reader.Close()

	result := &llm.StreamResult{}
	var contentBuilder strings.Builder

	for {
		eventType, rawLines, event, err := reader.ReadEvent()
		if err != nil {
			if err == io.EOF {
				break
			}
			h.logger.Error("读取 Anthropic SSE 流失败", zap.Error(err))
			break
		}

		fmt.Fprint(c.Writer, rawLines)
		if !endsWith(rawLines, "\n\n") {
			fmt.Fprint(c.Writer, "\n")
		}
		c.Writer.Flush()

		if event != nil {
			switch eventType {
			case llm.AnthropicEventMessageStart:
				if event.Message != nil {
					result.ResponseID = event.Message.ID
					result.Model = event.Message.Model
				}
			case llm.AnthropicEventContentBlockDelta:
				if event.Delta != nil {
					contentBuilder.WriteString(event.Delta.Text)
				}
			case llm.AnthropicEventMessageDelta:
				if event.Delta != nil && event.Delta.StopReason != nil {
					result.FinishReason = *event.Delta.StopReason
				}
				if event.Usage != nil {
					result.Usage = event.Usage.ToUsage()
				}
			}
		}
	}

	result.Content = contentBuilder.String()
	return result
}

// pipeAnthropicStreamToOpenAI 将 Anthropic 流转换为 OpenAI 格式输出，同时收集完整响应内容
func (h *LLMProxyHandler) pipeAnthropicStreamToOpenAI(c *gin.Context, body io.ReadCloser, model string) *llm.StreamResult {
	reader := llm.NewAnthropicStreamReader(body)
	defer reader.Close()

	result := &llm.StreamResult{}
	var contentBuilder strings.Builder
	state := &llm.AnthropicToOpenAIState{}

	for {
		eventType, _, event, err := reader.ReadEvent()
		if err != nil {
			if err == io.EOF {
				fmt.Fprintf(c.Writer, "data: [DONE]\n\n")
				c.Writer.Flush()
				break
			}
			h.logger.Error("读取 Anthropic SSE 流失败", zap.Error(err))
			break
		}

		openaiData := llm.AnthropicEventToOpenAIChunk(eventType, event, model, state)
		if openaiData != "" {
			fmt.Fprint(c.Writer, openaiData)
			c.Writer.Flush()
		}

		if event != nil {
			switch eventType {
			case llm.AnthropicEventMessageStart:
				if event.Message != nil {
					result.ResponseID = event.Message.ID
					result.Model = event.Message.Model
				}
			case llm.AnthropicEventContentBlockDelta:
				if event.Delta != nil {
					contentBuilder.WriteString(event.Delta.Text)
				}
			case llm.AnthropicEventMessageDelta:
				if event.Delta != nil && event.Delta.StopReason != nil {
					result.FinishReason = *event.Delta.StopReason
				}
				if event.Usage != nil {
					result.Usage = event.Usage.ToUsage()
				}
			}
		}
	}

	result.Content = contentBuilder.String()
	return result
}

// pipeOpenAIStreamToAnthropic 将 OpenAI 流转换为 Anthropic 格式输出，同时收集完整响应内容
func (h *LLMProxyHandler) pipeOpenAIStreamToAnthropic(c *gin.Context, body io.ReadCloser, model string) *llm.StreamResult {
	reader := llm.NewStreamReader(body)
	defer reader.Close()

	result := &llm.StreamResult{}
	var contentBuilder strings.Builder
	isFirst := true
	state := &llm.OpenAIToAnthropicState{}

	for {
		_, chunk, err := reader.ReadChunk()
		if err != nil {
			if err == io.EOF {
				break
			}
			h.logger.Error("读取 OpenAI SSE 流失败", zap.Error(err))
			break
		}

		if chunk != nil {
			anthropicData := llm.OpenAIChunkToAnthropicEvents(chunk, isFirst, state)
			if anthropicData != "" {
				fmt.Fprint(c.Writer, anthropicData)
				c.Writer.Flush()
			}
			isFirst = false

			if result.ResponseID == "" {
				result.ResponseID = chunk.ID
				result.Model = chunk.Model
			}
			if len(chunk.Choices) > 0 {
				delta := chunk.Choices[0].Delta
				if delta != nil {
					contentBuilder.WriteString(delta.ContentString())
				}
				if chunk.Choices[0].FinishReason != nil {
					result.FinishReason = *chunk.Choices[0].FinishReason
				}
			}
			if chunk.Usage != nil {
				result.Usage = chunk.Usage
			}
		}
	}

	result.Content = contentBuilder.String()
	return result
}

// pipeResponsesStream 转发 Responses API 的 SSE 命名事件流，同时收集完整响应内容
// Responses API 使用 event: xxx\ndata: {...}\n\n 格式，
// 从 response.completed 事件中提取最终 usage
func (h *LLMProxyHandler) pipeResponsesStream(c *gin.Context, body io.ReadCloser) *llm.StreamResult {
	reader := llm.NewResponsesStreamReader(body)
	defer reader.Close()

	result := &llm.StreamResult{}
	var contentBuilder strings.Builder

	for {
		eventType, rawLines, dataPayload, err := reader.ReadEvent()
		if err != nil {
			if err == io.EOF {
				break
			}
			h.logger.Error("读取 Responses SSE 流失败", zap.Error(err))
			break
		}

		fmt.Fprint(c.Writer, rawLines)
		if !endsWith(rawLines, "\n\n") {
			fmt.Fprint(c.Writer, "\n")
		}
		c.Writer.Flush()

		switch eventType {
		case llm.ResponsesEventOutputTextDelta:
			if len(dataPayload) > 0 {
				var delta struct {
					Delta string `json:"delta"`
				}
				if json.Unmarshal(dataPayload, &delta) == nil {
					contentBuilder.WriteString(delta.Delta)
				}
			}
		case llm.ResponsesEventCompleted:
			if len(dataPayload) > 0 {
				if usage := llm.ExtractUsageFromResponsesEvent(dataPayload); usage != nil {
					result.Usage = usage
				}
				var completed struct {
					Response struct {
						ID    string `json:"id"`
						Model string `json:"model"`
					} `json:"response"`
				}
				if json.Unmarshal(dataPayload, &completed) == nil {
					result.ResponseID = completed.Response.ID
					result.Model = completed.Response.Model
				}
			}
			result.FinishReason = "stop"
		}
	}

	result.Content = contentBuilder.String()
	return result
}

// ──────────────────────────────────
// 错误处理
// ──────────────────────────────────

// handleLLMError 处理 LLM 服务错误（OpenAI 格式响应）
func (h *LLMProxyHandler) handleLLMError(
	c *gin.Context, err error,
	userID, apiKeyID int64,
	modelName, requestType string,
	durationMs int,
) {
	var statusCode int
	var errMsg string

	if llmErr, ok := err.(*llm.LLMError); ok {
		statusCode = llmErr.StatusCode
		errMsg = llmErr.Message
		if len(llmErr.Body) > 0 {
			c.Data(statusCode, "application/json", llmErr.Body)
			go h.proxyService.RecordRequestLog(userID, apiKeyID, requestType, modelName, statusCode, errMsg, c.ClientIP(), c.Request.UserAgent(), durationMs)
			return
		}
	} else {
		statusCode = http.StatusBadGateway
		errMsg = "LLM 服务不可用"
	}

	go h.proxyService.RecordRequestLog(userID, apiKeyID, requestType, modelName, statusCode, errMsg, c.ClientIP(), c.Request.UserAgent(), durationMs)
	h.sendOpenAIError(c, statusCode, "server_error", errMsg)
}

// handleLLMErrorAnthropic 处理 LLM 服务错误（Anthropic 格式响应）
func (h *LLMProxyHandler) handleLLMErrorAnthropic(
	c *gin.Context, err error,
	userID, apiKeyID int64,
	modelName, requestType string,
	durationMs int,
) {
	var statusCode int
	var errMsg string

	if llmErr, ok := err.(*llm.LLMError); ok {
		statusCode = llmErr.StatusCode
		errMsg = llmErr.Message
		if len(llmErr.Body) > 0 {
			var anthropicErr llm.AnthropicErrorResponse
			if json.Unmarshal(llmErr.Body, &anthropicErr) == nil {
				c.JSON(statusCode, anthropicErr)
			} else {
				c.Data(statusCode, "application/json", llmErr.Body)
			}
			go h.proxyService.RecordRequestLog(userID, apiKeyID, requestType, modelName, statusCode, errMsg, c.ClientIP(), c.Request.UserAgent(), durationMs)
			return
		}
	} else {
		statusCode = http.StatusBadGateway
		errMsg = "LLM 服务不可用"
	}

	go h.proxyService.RecordRequestLog(userID, apiKeyID, requestType, modelName, statusCode, errMsg, c.ClientIP(), c.Request.UserAgent(), durationMs)
	h.sendAnthropicError(c, statusCode, "api_error", errMsg)
}

// handleLLMErrorResponses 处理 LLM 服务错误（Responses API 格式响应）
func (h *LLMProxyHandler) handleLLMErrorResponses(
	c *gin.Context, err error,
	userID, apiKeyID int64,
	modelName, requestType string,
	durationMs int,
) {
	var statusCode int
	var errMsg string

	if llmErr, ok := err.(*llm.LLMError); ok {
		statusCode = llmErr.StatusCode
		errMsg = llmErr.Message
		if len(llmErr.Body) > 0 {
			c.Data(statusCode, "application/json", llmErr.Body)
			go h.proxyService.RecordRequestLog(userID, apiKeyID, requestType, modelName, statusCode, errMsg, c.ClientIP(), c.Request.UserAgent(), durationMs)
			return
		}
	} else {
		statusCode = http.StatusBadGateway
		errMsg = "LLM 服务不可用"
	}

	go h.proxyService.RecordRequestLog(userID, apiKeyID, requestType, modelName, statusCode, errMsg, c.ClientIP(), c.Request.UserAgent(), durationMs)
	h.sendResponsesError(c, statusCode, "server_error", errMsg)
}

// sendResponsesError 发送 Responses API 格式错误响应
func (h *LLMProxyHandler) sendResponsesError(c *gin.Context, status int, code, msg string) {
	c.JSON(status, llm.NewResponsesError(code, msg))
}

// sendOpenAIError 发送 OpenAI 格式错误响应
func (h *LLMProxyHandler) sendOpenAIError(c *gin.Context, status int, errType, msg string) {
	c.JSON(status, llm.ErrorResponse{
		Error: llm.ErrorDetail{
			Message: msg,
			Type:    errType,
			Param:   nil,
			Code:    fmt.Sprintf("%d", status),
		},
	})
}

// sendAnthropicError 发送 Anthropic 格式错误响应
func (h *LLMProxyHandler) sendAnthropicError(c *gin.Context, status int, errType, msg string) {
	c.JSON(status, llm.AnthropicErrorResponse{
		Type: "error",
		Error: struct {
			Type    string `json:"type"`
			Message string `json:"message"`
		}{
			Type:    errType,
			Message: msg,
		},
	})
}

// endsWith 检查字符串是否以指定后缀结尾
func endsWith(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}
