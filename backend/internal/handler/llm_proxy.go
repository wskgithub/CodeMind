package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
// OpenAI 格式端点
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

	// 1. 读取原始请求体（保留所有字段，用于透传给 LLM）
	rawBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.sendOpenAIError(c, http.StatusBadRequest, "invalid_request_error", "请求体读取失败")
		return
	}

	// 2. 清理对话历史中的 thinking 内容
	// Thinking 模型（如 Qwen3-*-Thinking）会在 assistant 消息中留下大量 <think> 内容，
	// 这些内容在多轮对话中会填满上下文窗口，导致实际对话被截断、上下文丢失
	rawBody = llm.CleanThinkingFromHistory(rawBody)

	// 3. 仅提取路由所需的最小元数据（model、stream）
	meta, err := llm.ExtractRequestMeta(rawBody)
	if err != nil {
		h.sendOpenAIError(c, http.StatusBadRequest, "invalid_request_error", "请求体解析失败")
		return
	}

	// 4. 检查 Token 配额
	allowed, err := h.proxyService.CheckTokenQuota(c.Request.Context(), userID, deptID)
	if err != nil {
		h.logger.Error("配额检查失败", zap.Error(err))
	}
	if !allowed {
		h.sendOpenAIError(c, http.StatusTooManyRequests, "rate_limit_exceeded", errcode.ErrTokenQuotaExceeded.Message)
		return
	}

	// 5. 获取并发槽位
	acquired, err := h.proxyService.AcquireConcurrency(c.Request.Context(), userID, deptID)
	if err != nil {
		h.logger.Error("并发控制失败", zap.Error(err))
	}
	if !acquired {
		h.sendOpenAIError(c, http.StatusTooManyRequests, "rate_limit_exceeded", errcode.ErrConcurrencyExceeded.Message)
		return
	}
	defer h.proxyService.ReleaseConcurrency(c.Request.Context(), userID)

	// 6. 根据模型名路由到合适的 Provider（集成负载均衡）
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

	// 7. 根据流式模式分别处理（透传原始请求体）
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

	// 设置 SSE 响应头
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")
	c.Status(http.StatusOK)

	var totalUsage *llm.Usage

	// 根据上游 Provider 格式选择不同的流读取策略
	if provider.Format() == llm.FormatAnthropic {
		// 上游返回 Anthropic 格式流 → 转换为 OpenAI 格式再返回
		totalUsage = h.pipeAnthropicStreamToOpenAI(c, body, modelName)
	} else {
		// 上游返回 OpenAI 格式流 → 直接转发（SSE 行本身就是原始数据，tool_calls 等自然保留）
		totalUsage = h.pipeOpenAIStream(c, body)
	}

	durationMs := int(time.Since(startTime).Milliseconds())
	go h.proxyService.RecordUsage(userID, apiKeyID, deptID, modelName, "chat_completion", totalUsage, durationMs)
	go h.proxyService.RecordRequestLog(userID, apiKeyID, "chat_completion", modelName, 200, "", c.ClientIP(), c.Request.UserAgent(), durationMs)
}

// Completions 文本补全代理（OpenAI 格式）
// POST /v1/completions
func (h *LLMProxyHandler) Completions(c *gin.Context) {
	startTime := time.Now()
	userID := middleware.GetUserID(c)
	keyID, _ := c.Get(middleware.CtxKeyAPIKeyID)
	apiKeyID := keyID.(int64)
	deptID := middleware.GetDepartmentID(c)

	var req llm.CompletionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
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

	modelName := req.Model
	ctx := c.Request.Context()
	provider, err := h.proxyService.GetProviderForModel(ctx, userID, modelName)
	if err != nil {
		h.sendOpenAIError(c, http.StatusServiceUnavailable, "server_error", "没有可用的 LLM Provider")
		return
	}

	if req.Stream {
		body, err := provider.CompletionStream(ctx, &req)
		if err != nil {
			durationMs := int(time.Since(startTime).Milliseconds())
			h.handleLLMError(c, err, userID, apiKeyID, modelName, "completion", durationMs)
			return
		}

		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")
		c.Status(http.StatusOK)

		defer body.Close()
		buf := make([]byte, 4096)
		for {
			n, readErr := body.Read(buf)
			if n > 0 {
				c.Writer.Write(buf[:n])
				c.Writer.Flush()
			}
			if readErr != nil {
				break
			}
		}

		durationMs := int(time.Since(startTime).Milliseconds())
		go h.proxyService.RecordRequestLog(userID, apiKeyID, "completion", modelName, 200, "", c.ClientIP(), c.Request.UserAgent(), durationMs)
	} else {
		resp, err := provider.Completion(ctx, &req)
		durationMs := int(time.Since(startTime).Milliseconds())
		if err != nil {
			h.handleLLMError(c, err, userID, apiKeyID, modelName, "completion", durationMs)
			return
		}
		go h.proxyService.RecordUsage(userID, apiKeyID, deptID, modelName, "completion", resp.Usage, durationMs)
		go h.proxyService.RecordRequestLog(userID, apiKeyID, "completion", modelName, 200, "", c.ClientIP(), c.Request.UserAgent(), durationMs)
		c.JSON(http.StatusOK, resp)
	}
}

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

// ──────────────────────────────────
// Anthropic 格式端点
// ──────────────────────────────────

// AnthropicMessages Anthropic 原生消息代理
// POST /v1/messages
func (h *LLMProxyHandler) AnthropicMessages(c *gin.Context) {
	startTime := time.Now()
	userID := middleware.GetUserID(c)
	keyID, _ := c.Get(middleware.CtxKeyAPIKeyID)
	apiKeyID := keyID.(int64)
	deptID := middleware.GetDepartmentID(c)

	// 1. 解析 Anthropic 格式请求
	var req llm.AnthropicMessagesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.sendAnthropicError(c, http.StatusBadRequest, "invalid_request_error", "请求体解析失败")
		return
	}

	// 2. 检查配额
	allowed, err := h.proxyService.CheckTokenQuota(c.Request.Context(), userID, deptID)
	if err != nil {
		h.logger.Error("配额检查失败", zap.Error(err))
	}
	if !allowed {
		h.sendAnthropicError(c, http.StatusTooManyRequests, "rate_limit_error", errcode.ErrTokenQuotaExceeded.Message)
		return
	}

	// 3. 获取并发槽位
	acquired, err := h.proxyService.AcquireConcurrency(c.Request.Context(), userID, deptID)
	if err != nil {
		h.logger.Error("并发控制失败", zap.Error(err))
	}
	if !acquired {
		h.sendAnthropicError(c, http.StatusTooManyRequests, "rate_limit_error", errcode.ErrConcurrencyExceeded.Message)
		return
	}
	defer h.proxyService.ReleaseConcurrency(c.Request.Context(), userID)

	// 4. 路由到 Provider（集成负载均衡）
	modelName := req.Model
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

	// 5. 流式/非流式处理
	if req.Stream {
		h.handleAnthropicStream(c, ctx, provider, &req, userID, apiKeyID, deptID, modelName, startTime)
	} else {
		h.handleAnthropicNonStream(c, ctx, provider, &req, userID, apiKeyID, deptID, modelName, startTime)
	}
}

// handleAnthropicNonStream 处理 Anthropic 非流式请求
func (h *LLMProxyHandler) handleAnthropicNonStream(
	c *gin.Context, ctx context.Context, provider llm.Provider,
	req *llm.AnthropicMessagesRequest,
	userID, apiKeyID int64, deptID *int64, modelName string, startTime time.Time,
) {
	resp, err := provider.AnthropicMessages(ctx, req)
	durationMs := int(time.Since(startTime).Milliseconds())

	if err != nil {
		h.handleLLMErrorAnthropic(c, err, userID, apiKeyID, modelName, "anthropic_messages", durationMs)
		return
	}

	var usage *llm.Usage
	if resp.Usage != nil {
		usage = resp.Usage.ToUsage()
	}

	go h.proxyService.RecordUsage(userID, apiKeyID, deptID, modelName, "anthropic_messages", usage, durationMs)
	go h.proxyService.RecordRequestLog(userID, apiKeyID, "anthropic_messages", modelName, 200, "", c.ClientIP(), c.Request.UserAgent(), durationMs)

	c.JSON(http.StatusOK, resp)
}

// handleAnthropicStream 处理 Anthropic 流式请求
func (h *LLMProxyHandler) handleAnthropicStream(
	c *gin.Context, ctx context.Context, provider llm.Provider,
	req *llm.AnthropicMessagesRequest,
	userID, apiKeyID int64, deptID *int64, modelName string, startTime time.Time,
) {
	body, err := provider.AnthropicMessagesStream(ctx, req)
	if err != nil {
		durationMs := int(time.Since(startTime).Milliseconds())
		h.handleLLMErrorAnthropic(c, err, userID, apiKeyID, modelName, "anthropic_messages", durationMs)
		return
	}

	// 设置 Anthropic SSE 响应头
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")
	c.Status(http.StatusOK)

	var totalUsage *llm.Usage

	if provider.Format() == llm.FormatOpenAI {
		// 上游返回 OpenAI 格式流 → 转换为 Anthropic 格式再返回
		totalUsage = h.pipeOpenAIStreamToAnthropic(c, body, modelName)
	} else {
		// 上游返回 Anthropic 格式流 → 直接转发
		totalUsage = h.pipeAnthropicStream(c, body)
	}

	durationMs := int(time.Since(startTime).Milliseconds())
	go h.proxyService.RecordUsage(userID, apiKeyID, deptID, modelName, "anthropic_messages", totalUsage, durationMs)
	go h.proxyService.RecordRequestLog(userID, apiKeyID, "anthropic_messages", modelName, 200, "", c.ClientIP(), c.Request.UserAgent(), durationMs)
}

// ──────────────────────────────────
// 流式数据管道
// ──────────────────────────────────

// pipeOpenAIStream 直接转发 OpenAI 格式流
func (h *LLMProxyHandler) pipeOpenAIStream(c *gin.Context, body io.ReadCloser) *llm.Usage {
	reader := llm.NewStreamReader(body)
	defer reader.Close()

	var totalUsage *llm.Usage

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

		if chunk != nil && chunk.Usage != nil {
			totalUsage = chunk.Usage
		}
	}

	return totalUsage
}

// pipeAnthropicStream 直接转发 Anthropic 格式流
func (h *LLMProxyHandler) pipeAnthropicStream(c *gin.Context, body io.ReadCloser) *llm.Usage {
	reader := llm.NewAnthropicStreamReader(body)
	defer reader.Close()

	var totalUsage *llm.Usage

	for {
		eventType, rawLines, event, err := reader.ReadEvent()
		if err != nil {
			if err == io.EOF {
				break
			}
			h.logger.Error("读取 Anthropic SSE 流失败", zap.Error(err))
			break
		}

		// 直接转发原始 SSE 文本
		fmt.Fprint(c.Writer, rawLines)
		if !endsWith(rawLines, "\n\n") {
			fmt.Fprint(c.Writer, "\n")
		}
		c.Writer.Flush()

		// 从 message_delta 事件提取用量
		if eventType == llm.AnthropicEventMessageDelta && event != nil && event.Usage != nil {
			totalUsage = event.Usage.ToUsage()
		}
	}

	return totalUsage
}

// pipeAnthropicStreamToOpenAI 将 Anthropic 流转换为 OpenAI 格式输出
func (h *LLMProxyHandler) pipeAnthropicStreamToOpenAI(c *gin.Context, body io.ReadCloser, model string) *llm.Usage {
	reader := llm.NewAnthropicStreamReader(body)
	defer reader.Close()

	var totalUsage *llm.Usage

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

		// 将 Anthropic 事件转换为 OpenAI chunk 格式
		openaiData := llm.AnthropicEventToOpenAIChunk(eventType, event, model)
		if openaiData != "" {
			fmt.Fprint(c.Writer, openaiData)
			c.Writer.Flush()
		}

		// 提取用量
		if eventType == llm.AnthropicEventMessageDelta && event != nil && event.Usage != nil {
			totalUsage = event.Usage.ToUsage()
		}
	}

	return totalUsage
}

// pipeOpenAIStreamToAnthropic 将 OpenAI 流转换为 Anthropic 格式输出
func (h *LLMProxyHandler) pipeOpenAIStreamToAnthropic(c *gin.Context, body io.ReadCloser, model string) *llm.Usage {
	reader := llm.NewStreamReader(body)
	defer reader.Close()

	var totalUsage *llm.Usage
	isFirst := true

	for {
		_, chunk, err := reader.ReadChunk()
		if err != nil {
			if err == io.EOF {
				// 如果 OpenAI 流以 [DONE] 结束但没有 finish_reason
				// 确保发送 Anthropic 的结束事件
				break
			}
			h.logger.Error("读取 OpenAI SSE 流失败", zap.Error(err))
			break
		}

		if chunk != nil {
			anthropicData := llm.OpenAIChunkToAnthropicEvents(chunk, isFirst)
			if anthropicData != "" {
				fmt.Fprint(c.Writer, anthropicData)
				c.Writer.Flush()
			}
			isFirst = false

			if chunk.Usage != nil {
				totalUsage = chunk.Usage
			}
		}
	}

	return totalUsage
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
			// 尝试转发原始错误体，但包装为 Anthropic 格式
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

// sendOpenAIError 发送 OpenAI 格式错误响应
func (h *LLMProxyHandler) sendOpenAIError(c *gin.Context, status int, errType, msg string) {
	c.JSON(status, llm.ErrorResponse{
		Error: struct {
			Message string `json:"message"`
			Type    string `json:"type"`
			Code    string `json:"code"`
		}{
			Message: msg,
			Type:    errType,
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

