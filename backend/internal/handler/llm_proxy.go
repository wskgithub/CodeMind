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

const errMsgLLMUnavailable = "LLM service unavailable"

// LLMProxyHandler handles LLM proxy requests.
type LLMProxyHandler struct {
	proxyService *service.LLMProxyService
	logger       *zap.Logger
}

// NewLLMProxyHandler creates a new LLM proxy handler.
func NewLLMProxyHandler(proxyService *service.LLMProxyService, logger *zap.Logger) *LLMProxyHandler {
	return &LLMProxyHandler{
		proxyService: proxyService,
		logger:       logger,
	}
}

// ChatCompletions handles chat completion requests (POST /api/openai/v1/chat/completions).
func (h *LLMProxyHandler) ChatCompletions(c *gin.Context) {
	startTime := time.Now()
	userID := middleware.GetUserID(c)
	keyID, _ := c.Get(middleware.CtxKeyAPIKeyID)
	apiKeyID := keyID.(int64)
	deptID := middleware.GetDepartmentID(c)

	rawBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.sendOpenAIError(c, http.StatusBadRequest, "invalid_request_error", "failed to read request body")
		return
	}

	meta, err := llm.ExtractRequestMeta(rawBody)
	if err != nil {
		h.sendOpenAIError(c, http.StatusBadRequest, "invalid_request_error", "failed to parse request body")
		return
	}

	if h.checkAndHandleThirdParty(c, rawBody, meta, "/chat/completions", "openai", userID, apiKeyID, startTime) {
		return
	}

	rawBody = llm.CleanThinkingFromHistory(rawBody)

	allowed, err := h.proxyService.CheckTokenQuota(c.Request.Context(), userID, deptID)
	if err != nil {
		h.logger.Error("quota check failed", zap.Error(err))
	}
	if !allowed {
		h.sendOpenAIError(c, http.StatusTooManyRequests, "rate_limit_exceeded", errcode.ErrTokenQuotaExceeded.Message)
		return
	}

	acquired, err := h.proxyService.AcquireConcurrency(c.Request.Context(), userID, deptID)
	if err != nil {
		h.logger.Error("concurrency control failed", zap.Error(err))
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
		h.logger.Error("failed to get provider", zap.Error(err), zap.String("model", modelName))
		h.sendOpenAIError(c, http.StatusBadRequest, "invalid_request_error",
			fmt.Sprintf("model %q is not supported by any configured service", modelName))
		return
	}

	h.logger.Debug("routing request to provider",
		zap.String("model", modelName),
		zap.String("provider", provider.Name()),
		zap.String("format", string(provider.Format())),
	)

	if meta.Stream {
		h.handleStreamChatRaw(ctx, c, provider, rawBody, userID, apiKeyID, deptID, modelName, startTime)
	} else {
		h.handleNonStreamChatRaw(ctx, c, provider, rawBody, userID, apiKeyID, deptID, modelName, startTime)
	}
}

// handleNonStreamChatRaw handles non-streaming chat with raw body passthrough.
func (h *LLMProxyHandler) handleNonStreamChatRaw(
	ctx context.Context, c *gin.Context, provider llm.Provider,
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
	go h.proxyService.RecordRequestLog(userID, apiKeyID, "chat_completion", modelName, 200, "", c.ClientIP(), c.Request.UserAgent(), durationMs)        //nolint:mnd // intentional constant.
	go h.proxyService.RecordTrainingData(userID, apiKeyID, "chat_completion", modelName, false, rawBody, rawResp, usage, 200, durationMs, c.ClientIP()) //nolint:mnd // intentional constant.

	c.Data(http.StatusOK, "application/json", rawResp)
}

// handleStreamChatRaw handles streaming chat (SSE) with raw body passthrough.
func (h *LLMProxyHandler) handleStreamChatRaw(
	ctx context.Context, c *gin.Context, provider llm.Provider,
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
	go h.proxyService.RecordRequestLog(userID, apiKeyID, "chat_completion", modelName, 200, "", c.ClientIP(), c.Request.UserAgent(), durationMs)                                                   //nolint:mnd // intentional constant.
	go h.proxyService.RecordTrainingData(userID, apiKeyID, "chat_completion", modelName, true, rawBody, llm.AssembleChatResponse(streamResult), streamResult.Usage, 200, durationMs, c.ClientIP()) //nolint:mnd // intentional constant.
}

// Completions handles text completion requests (POST /api/openai/v1/completions).
func (h *LLMProxyHandler) Completions(c *gin.Context) {
	startTime := time.Now()
	userID := middleware.GetUserID(c)
	keyID, _ := c.Get(middleware.CtxKeyAPIKeyID)
	apiKeyID := keyID.(int64)
	deptID := middleware.GetDepartmentID(c)

	rawBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.sendOpenAIError(c, http.StatusBadRequest, "invalid_request_error", "failed to read request body")
		return
	}

	meta, err := llm.ExtractRequestMeta(rawBody)
	if err != nil {
		h.sendOpenAIError(c, http.StatusBadRequest, "invalid_request_error", "failed to parse request body")
		return
	}

	if h.checkAndHandleThirdParty(c, rawBody, meta, "/completions", "openai", userID, apiKeyID, startTime) {
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
		h.sendOpenAIError(c, http.StatusBadRequest, "invalid_request_error",
			fmt.Sprintf("model %q is not supported by any configured service", modelName))
		return
	}

	if meta.Stream {
		h.handleStreamCompletionRaw(ctx, c, provider, rawBody, userID, apiKeyID, deptID, modelName, startTime)
	} else {
		h.handleNonStreamCompletionRaw(ctx, c, provider, rawBody, userID, apiKeyID, deptID, modelName, startTime)
	}
}

// handleNonStreamCompletionRaw handles non-streaming text completion.
func (h *LLMProxyHandler) handleNonStreamCompletionRaw(
	ctx context.Context, c *gin.Context, provider llm.Provider,
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
	go h.proxyService.RecordRequestLog(userID, apiKeyID, "completion", modelName, 200, "", c.ClientIP(), c.Request.UserAgent(), durationMs)        //nolint:mnd // intentional constant.
	go h.proxyService.RecordTrainingData(userID, apiKeyID, "completion", modelName, false, rawBody, rawResp, usage, 200, durationMs, c.ClientIP()) //nolint:mnd // intentional constant.

	c.Data(http.StatusOK, "application/json", rawResp)
}

// handleStreamCompletionRaw handles streaming text completion.
func (h *LLMProxyHandler) handleStreamCompletionRaw(
	ctx context.Context, c *gin.Context, provider llm.Provider,
	rawBody []byte,
	userID, apiKeyID int64, deptID *int64, modelName string, startTime time.Time,
) {
	body, err := provider.CompletionStreamRaw(ctx, rawBody)
	if err != nil {
		durationMs := int(time.Since(startTime).Milliseconds())
		h.handleLLMError(c, err, userID, apiKeyID, modelName, "completion", durationMs)
		return
	}
	defer func() { _ = body.Close() }()

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")
	c.Status(http.StatusOK)

	result := &llm.StreamResult{}
	var contentBuilder strings.Builder
	reader := llm.NewStreamReader(body)
	defer func() { _ = reader.Close() }()

	for {
		rawLine, chunk, err := reader.ReadChunk()
		if err != nil {
			if err == io.EOF {
				_, _ = fmt.Fprintf(c.Writer, "data: [DONE]\n\n")
				c.Writer.Flush()
				break
			}
			h.logger.Error("failed to read completions SSE stream", zap.Error(err))
			break
		}

		_, _ = fmt.Fprintf(c.Writer, "%s\n\n", rawLine)
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
	go h.proxyService.RecordRequestLog(userID, apiKeyID, "completion", modelName, 200, "", c.ClientIP(), c.Request.UserAgent(), durationMs)                                             //nolint:mnd // intentional constant.
	go h.proxyService.RecordTrainingData(userID, apiKeyID, "completion", modelName, true, rawBody, llm.AssembleCompletionResponse(result), result.Usage, 200, durationMs, c.ClientIP()) //nolint:mnd // intentional constant.
}

// ListModels returns the list of available models (GET /api/openai/v1/models).
func (h *LLMProxyHandler) ListModels(c *gin.Context) {
	provider, err := h.proxyService.GetProviderManager().GetDefault()
	if err != nil {
		h.sendOpenAIError(c, http.StatusBadGateway, "server_error", "failed to get model list")
		return
	}
	resp, err := provider.ListModels(c.Request.Context())
	if err != nil {
		h.sendOpenAIError(c, http.StatusBadGateway, "server_error", "failed to get model list")
		return
	}
	c.JSON(http.StatusOK, resp)
}

// RetrieveModel returns information for a specific model (GET /api/openai/v1/models/:model).
func (h *LLMProxyHandler) RetrieveModel(c *gin.Context) {
	modelID := c.Param("model")
	if modelID == "" {
		h.sendOpenAIError(c, http.StatusBadRequest, "invalid_request_error", "model ID is required")
		return
	}

	provider, err := h.proxyService.GetProviderManager().GetDefault()
	if err != nil {
		h.sendOpenAIError(c, http.StatusBadGateway, "server_error", "failed to get model info")
		return
	}

	resp, err := provider.RetrieveModel(c.Request.Context(), modelID)
	if err != nil {
		if llmErr, ok := err.(*llm.Error); ok && llmErr.StatusCode == 404 {
			h.sendOpenAIError(c, http.StatusNotFound, "invalid_request_error",
				fmt.Sprintf("The model '%s' does not exist", modelID))
			return
		}
		h.sendOpenAIError(c, http.StatusBadGateway, "server_error", "failed to get model info")
		return
	}
	c.JSON(http.StatusOK, resp)
}

// Embeddings handles embedding requests (POST /api/openai/v1/embeddings).
func (h *LLMProxyHandler) Embeddings(c *gin.Context) {
	startTime := time.Now()
	userID := middleware.GetUserID(c)
	keyID, _ := c.Get(middleware.CtxKeyAPIKeyID)
	apiKeyID := keyID.(int64)
	deptID := middleware.GetDepartmentID(c)

	rawBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.sendOpenAIError(c, http.StatusBadRequest, "invalid_request_error", "failed to read request body")
		return
	}

	meta, err := llm.ExtractRequestMeta(rawBody)
	if err != nil {
		h.sendOpenAIError(c, http.StatusBadRequest, "invalid_request_error", "failed to parse request body")
		return
	}

	if h.checkAndHandleThirdParty(c, rawBody, meta, "/embeddings", "openai", userID, apiKeyID, startTime) {
		return
	}

	allowed, _ := h.proxyService.CheckTokenQuota(c.Request.Context(), userID, deptID)
	if !allowed {
		h.sendOpenAIError(c, http.StatusTooManyRequests, "rate_limit_exceeded", errcode.ErrTokenQuotaExceeded.Message)
		return
	}

	acquired, err := h.proxyService.AcquireConcurrency(c.Request.Context(), userID, deptID)
	if err != nil {
		h.logger.Error("concurrency control failed", zap.Error(err))
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
		h.sendOpenAIError(c, http.StatusBadRequest, "invalid_request_error",
			fmt.Sprintf("model %q is not supported by any configured service", modelName))
		return
	}

	rawResp, usage, err := provider.EmbeddingRaw(ctx, rawBody)
	durationMs := int(time.Since(startTime).Milliseconds())

	if err != nil {
		h.handleLLMError(c, err, userID, apiKeyID, modelName, "embedding", durationMs)
		return
	}

	if usage != nil {
		go h.proxyService.RecordUsage(userID, apiKeyID, deptID, modelName, "embedding", usage, durationMs)
	}
	go h.proxyService.RecordRequestLog(userID, apiKeyID, "embedding", modelName, 200, "", c.ClientIP(), c.Request.UserAgent(), durationMs)    //nolint:mnd // intentional constant.
	go h.proxyService.RecordTrainingData(userID, apiKeyID, "embedding", modelName, false, rawBody, nil, usage, 200, durationMs, c.ClientIP()) //nolint:mnd // intentional constant.

	c.Data(http.StatusOK, "application/json", rawResp)
}

// ──────────────────────────────────
// OpenAI Responses API
// ──────────────────────────────────

// Responses handles Responses API requests (POST /api/openai/v1/responses).
func (h *LLMProxyHandler) Responses(c *gin.Context) {
	startTime := time.Now()
	userID := middleware.GetUserID(c)
	keyID, _ := c.Get(middleware.CtxKeyAPIKeyID)
	apiKeyID := keyID.(int64)
	deptID := middleware.GetDepartmentID(c)

	rawBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.sendResponsesError(c, http.StatusBadRequest, "invalid_request_error", "failed to read request body")
		return
	}

	meta, err := llm.ExtractRequestMeta(rawBody)
	if err != nil {
		h.sendResponsesError(c, http.StatusBadRequest, "invalid_request_error", "failed to parse request body")
		return
	}

	if h.checkAndHandleThirdParty(c, rawBody, meta, "/responses", "openai", userID, apiKeyID, startTime) {
		return
	}

	allowed, err := h.proxyService.CheckTokenQuota(c.Request.Context(), userID, deptID)
	if err != nil {
		h.logger.Error("quota check failed", zap.Error(err))
	}
	if !allowed {
		h.sendResponsesError(c, http.StatusTooManyRequests, "rate_limit_exceeded", errcode.ErrTokenQuotaExceeded.Message)
		return
	}

	acquired, err := h.proxyService.AcquireConcurrency(c.Request.Context(), userID, deptID)
	if err != nil {
		h.logger.Error("concurrency control failed", zap.Error(err))
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
		h.logger.Error("failed to get provider", zap.Error(err), zap.String("model", modelName))
		h.sendResponsesError(c, http.StatusBadRequest, "invalid_request_error",
			fmt.Sprintf("model %q is not supported by any configured service", modelName))
		return
	}

	h.logger.Debug("Responses API routing request to provider",
		zap.String("model", modelName),
		zap.String("provider", provider.Name()),
	)

	if meta.Stream {
		h.handleStreamResponses(ctx, c, provider, rawBody, userID, apiKeyID, deptID, modelName, startTime)
	} else {
		h.handleNonStreamResponses(ctx, c, provider, rawBody, userID, apiKeyID, deptID, modelName, startTime)
	}
}

// handleNonStreamResponses handles non-streaming Responses API.
func (h *LLMProxyHandler) handleNonStreamResponses(
	ctx context.Context, c *gin.Context, provider llm.Provider,
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
	go h.proxyService.RecordRequestLog(userID, apiKeyID, "responses", modelName, 200, "", c.ClientIP(), c.Request.UserAgent(), durationMs)        //nolint:mnd // intentional constant.
	go h.proxyService.RecordTrainingData(userID, apiKeyID, "responses", modelName, false, rawBody, rawResp, usage, 200, durationMs, c.ClientIP()) //nolint:mnd // intentional constant.

	c.Data(http.StatusOK, "application/json", rawResp)
}

// handleStreamResponses handles streaming Responses API (SSE named events).
func (h *LLMProxyHandler) handleStreamResponses(
	ctx context.Context, c *gin.Context, provider llm.Provider,
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
	go h.proxyService.RecordRequestLog(userID, apiKeyID, "responses", modelName, 200, "", c.ClientIP(), c.Request.UserAgent(), durationMs)                //nolint:mnd // intentional constant.
	go h.proxyService.RecordTrainingData(userID, apiKeyID, "responses", modelName, true, rawBody, nil, streamResult.Usage, 200, durationMs, c.ClientIP()) //nolint:mnd // intentional constant.
}

// AnthropicMessages handles Anthropic message requests (POST /api/anthropic/v1/messages).
func (h *LLMProxyHandler) AnthropicMessages(c *gin.Context) {
	startTime := time.Now()
	userID := middleware.GetUserID(c)
	keyID, _ := c.Get(middleware.CtxKeyAPIKeyID)
	apiKeyID := keyID.(int64)
	deptID := middleware.GetDepartmentID(c)

	rawBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.sendAnthropicError(c, http.StatusBadRequest, "invalid_request_error", "failed to read request body")
		return
	}

	var partial struct {
		Model  string `json:"model"`
		Stream bool   `json:"stream"`
	}
	if err = json.Unmarshal(rawBody, &partial); err != nil {
		h.sendAnthropicError(c, http.StatusBadRequest, "invalid_request_error", "invalid JSON in request body")
		return
	}
	if partial.Model == "" {
		h.sendAnthropicError(c, http.StatusBadRequest, "invalid_request_error", "missing required field: model")
		return
	}

	meta := &llm.RequestMeta{Model: partial.Model, Stream: partial.Stream}
	if h.checkAndHandleThirdParty(c, rawBody, meta, "/v1/messages", "anthropic", userID, apiKeyID, startTime) {
		return
	}

	allowed, err := h.proxyService.CheckTokenQuota(c.Request.Context(), userID, deptID)
	if err != nil {
		h.logger.Error("quota check failed", zap.Error(err))
	}
	if !allowed {
		h.sendAnthropicError(c, http.StatusTooManyRequests, "rate_limit_error", errcode.ErrTokenQuotaExceeded.Message)
		return
	}

	acquired, err := h.proxyService.AcquireConcurrency(c.Request.Context(), userID, deptID)
	if err != nil {
		h.logger.Error("concurrency control failed", zap.Error(err))
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
		h.sendAnthropicError(c, http.StatusBadRequest, "invalid_request_error",
			fmt.Sprintf("model %q is not supported by any configured service", modelName))
		return
	}

	h.logger.Debug("Anthropic request routing to provider",
		zap.String("model", modelName),
		zap.String("provider", provider.Name()),
		zap.String("format", string(provider.Format())),
	)

	if partial.Stream {
		h.handleAnthropicStream(ctx, c, provider, rawBody, userID, apiKeyID, deptID, modelName, startTime)
	} else {
		h.handleAnthropicNonStream(ctx, c, provider, rawBody, userID, apiKeyID, deptID, modelName, startTime)
	}
}

// handleAnthropicNonStream handles Anthropic non-streaming requests.
func (h *LLMProxyHandler) handleAnthropicNonStream(
	ctx context.Context, c *gin.Context, provider llm.Provider,
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
	go h.proxyService.RecordRequestLog(userID, apiKeyID, "anthropic_messages", modelName, 200, "", c.ClientIP(), c.Request.UserAgent(), durationMs)        //nolint:mnd // intentional constant.
	go h.proxyService.RecordTrainingData(userID, apiKeyID, "anthropic_messages", modelName, false, rawBody, rawResp, usage, 200, durationMs, c.ClientIP()) //nolint:mnd // intentional constant.

	c.Data(http.StatusOK, "application/json", rawResp)
}

// handleAnthropicStream handles Anthropic streaming requests.
func (h *LLMProxyHandler) handleAnthropicStream(
	ctx context.Context, c *gin.Context, provider llm.Provider,
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
	go h.proxyService.RecordRequestLog(userID, apiKeyID, "anthropic_messages", modelName, 200, "", c.ClientIP(), c.Request.UserAgent(), durationMs)                                                   //nolint:mnd // intentional constant.
	go h.proxyService.RecordTrainingData(userID, apiKeyID, "anthropic_messages", modelName, true, rawBody, llm.AssembleChatResponse(streamResult), streamResult.Usage, 200, durationMs, c.ClientIP()) //nolint:mnd // intentional constant.
}

// pipeOpenAIStream forwards OpenAI format stream while collecting full response content.
func (h *LLMProxyHandler) pipeOpenAIStream(c *gin.Context, body io.ReadCloser) *llm.StreamResult {
	reader := llm.NewStreamReader(body)
	defer func() { _ = reader.Close() }()

	result := &llm.StreamResult{}
	var contentBuilder strings.Builder

	for {
		rawLine, chunk, err := reader.ReadChunk()
		if err != nil {
			if err == io.EOF {
				_, _ = fmt.Fprintf(c.Writer, "data: [DONE]\n\n")
				c.Writer.Flush()
				break
			}
			h.logger.Error("failed to read OpenAI SSE stream", zap.Error(err))
			break
		}

		_, _ = fmt.Fprintf(c.Writer, "%s\n\n", rawLine)
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

// pipeAnthropicStream forwards Anthropic format stream while collecting full response content.
func (h *LLMProxyHandler) pipeAnthropicStream(c *gin.Context, body io.ReadCloser) *llm.StreamResult {
	reader := llm.NewAnthropicStreamReader(body)
	defer func() { _ = reader.Close() }()

	result := &llm.StreamResult{}
	var contentBuilder strings.Builder

	for {
		eventType, rawLines, event, err := reader.ReadEvent()
		if err != nil {
			if err == io.EOF {
				break
			}
			h.logger.Error("failed to read Anthropic SSE stream", zap.Error(err))
			break
		}

		_, _ = fmt.Fprint(c.Writer, rawLines)
		if !endsWith(rawLines, "\n\n") {
			_, _ = fmt.Fprint(c.Writer, "\n")
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

// pipeAnthropicStreamToOpenAI converts Anthropic stream to OpenAI format output.
func (h *LLMProxyHandler) pipeAnthropicStreamToOpenAI(c *gin.Context, body io.ReadCloser, model string) *llm.StreamResult {
	reader := llm.NewAnthropicStreamReader(body)
	defer func() { _ = reader.Close() }()

	result := &llm.StreamResult{}
	var contentBuilder strings.Builder
	state := &llm.AnthropicToOpenAIState{}

	for {
		eventType, _, event, err := reader.ReadEvent()
		if err != nil {
			if err == io.EOF {
				_, _ = fmt.Fprintf(c.Writer, "data: [DONE]\n\n")
				c.Writer.Flush()
				break
			}
			h.logger.Error("failed to read Anthropic SSE stream", zap.Error(err))
			break
		}

		openaiData := llm.AnthropicEventToOpenAIChunk(eventType, event, model, state)
		if openaiData != "" {
			_, _ = fmt.Fprint(c.Writer, openaiData)
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

// pipeOpenAIStreamToAnthropic converts OpenAI stream to Anthropic format output.
func (h *LLMProxyHandler) pipeOpenAIStreamToAnthropic(c *gin.Context, body io.ReadCloser, _ string) *llm.StreamResult {
	reader := llm.NewStreamReader(body)
	defer func() { _ = reader.Close() }()

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
			h.logger.Error("failed to read OpenAI SSE stream", zap.Error(err))
			break
		}

		if chunk != nil {
			anthropicData := llm.OpenAIChunkToAnthropicEvents(chunk, isFirst, state)
			if anthropicData != "" {
				_, _ = fmt.Fprint(c.Writer, anthropicData)
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

// pipeResponsesStream forwards Responses API SSE named event stream.
func (h *LLMProxyHandler) pipeResponsesStream(c *gin.Context, body io.ReadCloser) *llm.StreamResult {
	reader := llm.NewResponsesStreamReader(body)
	defer func() { _ = reader.Close() }()

	result := &llm.StreamResult{}
	var contentBuilder strings.Builder

	for {
		eventType, rawLines, dataPayload, err := reader.ReadEvent()
		if err != nil {
			if err == io.EOF {
				break
			}
			h.logger.Error("failed to read Responses SSE stream", zap.Error(err))
			break
		}

		_, _ = fmt.Fprint(c.Writer, rawLines)
		if !endsWith(rawLines, "\n\n") {
			_, _ = fmt.Fprint(c.Writer, "\n")
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

// handleLLMError handles LLM service errors (OpenAI format response).
func (h *LLMProxyHandler) handleLLMError(
	c *gin.Context, err error,
	userID, apiKeyID int64,
	modelName, requestType string,
	durationMs int,
) {
	var statusCode int
	var errMsg string

	if llmErr, ok := err.(*llm.Error); ok {
		statusCode = llmErr.StatusCode
		errMsg = llmErr.Message
	} else {
		statusCode = http.StatusBadGateway
		errMsg = errMsgLLMUnavailable
	}

	go h.proxyService.RecordRequestLog(userID, apiKeyID, requestType, modelName, statusCode, errMsg, c.ClientIP(), c.Request.UserAgent(), durationMs)
	h.sendOpenAIError(c, statusCode, "server_error", errMsg)
}

// handleLLMErrorAnthropic handles LLM service errors (Anthropic format response).
func (h *LLMProxyHandler) handleLLMErrorAnthropic(
	c *gin.Context, err error,
	userID, apiKeyID int64,
	modelName, requestType string,
	durationMs int,
) {
	var statusCode int
	var errMsg string

	if llmErr, ok := err.(*llm.Error); ok {
		statusCode = llmErr.StatusCode
		errMsg = llmErr.Message
	} else {
		statusCode = http.StatusBadGateway
		errMsg = errMsgLLMUnavailable
	}

	go h.proxyService.RecordRequestLog(userID, apiKeyID, requestType, modelName, statusCode, errMsg, c.ClientIP(), c.Request.UserAgent(), durationMs)
	h.sendAnthropicError(c, statusCode, "api_error", errMsg)
}

// handleLLMErrorResponses handles LLM service errors (Responses API format).
func (h *LLMProxyHandler) handleLLMErrorResponses(
	c *gin.Context, err error,
	userID, apiKeyID int64,
	modelName, requestType string,
	durationMs int,
) {
	var statusCode int
	var errMsg string

	if llmErr, ok := err.(*llm.Error); ok {
		statusCode = llmErr.StatusCode
		errMsg = llmErr.Message
	} else {
		statusCode = http.StatusBadGateway
		errMsg = errMsgLLMUnavailable
	}

	go h.proxyService.RecordRequestLog(userID, apiKeyID, requestType, modelName, statusCode, errMsg, c.ClientIP(), c.Request.UserAgent(), durationMs)
	h.sendResponsesError(c, statusCode, "server_error", errMsg)
}

// sendResponsesError sends Responses API format error response.
func (h *LLMProxyHandler) sendResponsesError(c *gin.Context, status int, code, msg string) {
	c.JSON(status, llm.NewResponsesError(code, msg))
}

// sendOpenAIError sends OpenAI format error response.
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

// sendAnthropicError sends Anthropic format error response.
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

// endsWith checks if a string ends with the specified suffix.
func endsWith(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}
