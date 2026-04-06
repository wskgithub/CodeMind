package handler

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"codemind/internal/model"
	"codemind/pkg/llm"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// 第三方服务代理专用 HTTP 客户端（独立超时设置）
var thirdPartyHTTPClient = &http.Client{
	Timeout: 600 * time.Second,
	Transport: &http.Transport{
		MaxIdleConnsPerHost: 20,
		IdleConnTimeout:     120 * time.Second,
	},
}

// handleThirdPartyProxy 第三方模型服务透明代理
// 核心原则：仅做数据路由，不修改原始请求/响应数据包
// requestFormat: "openai" 或 "anthropic"，决定认证头的设置方式
func (h *LLMProxyHandler) handleThirdPartyProxy(
	c *gin.Context,
	rawBody []byte,
	meta *llm.RequestMeta,
	route *model.ThirdPartyRouteInfo,
	endpointPath string,
	requestFormat string,
	userID, apiKeyID int64,
	startTime time.Time,
) {
	// 解密第三方 API Key
	apiKey, err := h.proxyService.GetThirdPartyService().DecryptAPIKey(route.APIKeyEncrypted)
	if err != nil {
		h.logger.Error("解密第三方 API Key 失败",
			zap.Int64("provider_id", route.ProviderID),
			zap.Error(err),
		)
		h.sendFormatError(c, requestFormat, http.StatusInternalServerError, "第三方服务配置异常")
		return
	}

	// 根据请求协议格式选择对应的 Base URL
	baseURL := route.BaseURLForFormat(requestFormat)
	targetURL := strings.TrimRight(baseURL, "/") + endpointPath

	// 创建代理请求 — 透传原始请求体，不做任何修改
	proxyReq, err := http.NewRequestWithContext(c.Request.Context(), c.Request.Method, targetURL, bytes.NewReader(rawBody))
	if err != nil {
		h.sendFormatError(c, requestFormat, http.StatusInternalServerError, "构建代理请求失败")
		return
	}

	// 透传客户端请求头（保持客户端标识完整性）
	copyRequestHeaders(c.Request, proxyReq)

	// 根据请求协议格式设置认证头
	if requestFormat == "anthropic" {
		proxyReq.Header.Set("x-api-key", apiKey)
	} else {
		proxyReq.Header.Set("Authorization", "Bearer "+apiKey)
	}

	h.logger.Debug("代理请求到第三方服务",
		zap.String("model", meta.Model),
		zap.String("provider", route.ProviderName),
		zap.String("target_url", targetURL),
		zap.Bool("stream", meta.Stream),
	)

	// 发送请求到第三方服务
	resp, err := thirdPartyHTTPClient.Do(proxyReq)
	if err != nil {
		durationMs := int(time.Since(startTime).Milliseconds())
		h.logger.Error("第三方服务请求失败", zap.Error(err), zap.String("url", targetURL))
		go h.proxyService.RecordRequestLog(userID, apiKeyID, "chat_completion", meta.Model, 502, err.Error(), c.ClientIP(), c.Request.UserAgent(), durationMs)
		h.sendFormatError(c, requestFormat, http.StatusBadGateway, "第三方服务不可用: "+err.Error())
		return
	}
	defer resp.Body.Close()

	// 非 2xx 响应直接透传（保持第三方服务原始错误格式）
	if resp.StatusCode >= 400 {
		h.pipeThirdPartyErrorResponse(c, resp, route, meta, userID, apiKeyID, startTime)
		return
	}

	if meta.Stream {
		h.pipeThirdPartyStreamResponse(c, resp, route, meta, rawBody, requestFormat, userID, apiKeyID, startTime)
	} else {
		h.pipeThirdPartyNonStreamResponse(c, resp, route, meta, rawBody, requestFormat, userID, apiKeyID, startTime)
	}
}

// pipeThirdPartyNonStreamResponse 透传非流式响应
func (h *LLMProxyHandler) pipeThirdPartyNonStreamResponse(
	c *gin.Context, resp *http.Response,
	route *model.ThirdPartyRouteInfo, meta *llm.RequestMeta,
	requestBody []byte, requestFormat string, userID, apiKeyID int64, startTime time.Time,
) {
	respBody, err := io.ReadAll(resp.Body)
	durationMs := int(time.Since(startTime).Milliseconds())

	if err != nil {
		h.logger.Error("读取第三方响应失败", zap.Error(err))
		h.sendFormatError(c, requestFormat, http.StatusBadGateway, "读取第三方服务响应失败")
		return
	}

	// 尝试从响应中提取 usage（仅用于内部统计，兼容 OpenAI 和 Anthropic 两种字段名）
	usage := extractUsageFromResponse(respBody)

	// 异步记录：第三方用量（仅参考） + 训练数据
	go h.recordThirdPartyMetrics(route, meta, userID, apiKeyID, usage, durationMs, c, requestBody, respBody, false)

	// 透传原始响应（不修改 Content-Type 和响应体）
	copyResponseHeaders(resp, c)
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)
}

// pipeThirdPartyStreamResponse 透传流式 SSE 响应（逐字节转发，同时观测 usage 并提取文本内容）
func (h *LLMProxyHandler) pipeThirdPartyStreamResponse(
	c *gin.Context, resp *http.Response,
	route *model.ThirdPartyRouteInfo, meta *llm.RequestMeta,
	requestBody []byte, requestFormat string, userID, apiKeyID int64, startTime time.Time,
) {
	// 透传上游响应头
	for key, vals := range resp.Header {
		for _, v := range vals {
			c.Writer.Header().Add(key, v)
		}
	}
	c.Status(resp.StatusCode)

	reader := bufio.NewReaderSize(resp.Body, 32*1024)
	var lastUsage *llm.Usage
	var contentBuilder strings.Builder

	for {
		line, err := reader.ReadBytes('\n')
		if len(line) > 0 {
			c.Writer.Write(line)

			// 从 SSE data 行中提取 usage 和文本内容（仅用于内部统计和训练数据记录）
			trimmed := bytes.TrimSpace(line)
			if bytes.HasPrefix(trimmed, []byte("data: ")) {
				data := trimmed[6:]
				if !bytes.Equal(data, []byte("[DONE]")) {
					if u := extractUsageFromSSEData(data); u != nil {
						lastUsage = u
					}
					if text := extractContentFromSSEData(data); text != "" {
						contentBuilder.WriteString(text)
					}
				}
			}
		}

		if err != nil {
			break
		}

		// SSE 事件以空行分隔，空行时刷新输出
		if len(bytes.TrimRight(line, "\r\n")) == 0 {
			c.Writer.(http.Flusher).Flush()
		}
	}
	c.Writer.(http.Flusher).Flush()

	durationMs := int(time.Since(startTime).Milliseconds())

	// 从流式 SSE 中组装等效的非流式响应（用于训练数据记录）
	streamResult := &llm.StreamResult{
		Content:      contentBuilder.String(),
		Model:        meta.Model,
		Usage:        lastUsage,
		FinishReason: "stop",
	}
	assembledResp := llm.AssembleChatResponse(streamResult)

	// 异步记录第三方指标
	go h.recordThirdPartyMetrics(route, meta, userID, apiKeyID, lastUsage, durationMs, c, requestBody, assembledResp, true)
}

// pipeThirdPartyErrorResponse 透传第三方服务错误响应
func (h *LLMProxyHandler) pipeThirdPartyErrorResponse(
	c *gin.Context, resp *http.Response,
	route *model.ThirdPartyRouteInfo, meta *llm.RequestMeta,
	userID, apiKeyID int64, startTime time.Time,
) {
	respBody, _ := io.ReadAll(resp.Body)
	durationMs := int(time.Since(startTime).Milliseconds())

	errMsg := string(respBody)
	if len(errMsg) > 500 {
		errMsg = errMsg[:500]
	}

	go h.proxyService.RecordRequestLog(
		userID, apiKeyID, "chat_completion", meta.Model,
		resp.StatusCode, errMsg, c.ClientIP(), c.Request.UserAgent(), durationMs,
	)

	// 透传原始错误响应
	copyResponseHeaders(resp, c)
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)
}

// recordThirdPartyMetrics 异步记录第三方服务指标
func (h *LLMProxyHandler) recordThirdPartyMetrics(
	route *model.ThirdPartyRouteInfo, meta *llm.RequestMeta,
	userID, apiKeyID int64, usage *llm.Usage, durationMs int,
	c *gin.Context, requestBody, responseBody []byte, isStream bool,
) {
	tpService := h.proxyService.GetThirdPartyService()

	// 记录第三方用量（仅供参考）
	var promptTokens, completionTokens, totalTokens int
	var cacheCreationTokens, cacheReadTokens int
	if usage != nil {
		promptTokens = usage.PromptTokens
		completionTokens = usage.CompletionTokens
		totalTokens = usage.TotalTokens
		if usage.PromptTokensDetails != nil {
			cacheCreationTokens = usage.PromptTokensDetails.CacheCreationInputTokens
			cacheReadTokens = usage.PromptTokensDetails.CacheReadInputTokens
		}
	}
	tpService.RecordThirdPartyUsage(
		userID, route.ProviderID, apiKeyID,
		meta.Model, "chat_completion",
		promptTokens, completionTokens, totalTokens,
		cacheCreationTokens, cacheReadTokens,
		&durationMs,
	)

	// 记录请求日志
	h.proxyService.RecordRequestLog(
		userID, apiKeyID, "chat_completion", meta.Model,
		200, "", c.ClientIP(), c.Request.UserAgent(), durationMs,
	)

	// 记录训练数据（第三方来源）
	h.proxyService.RecordTrainingDataWithSource(
		userID, apiKeyID, "chat_completion", meta.Model,
		isStream, requestBody, responseBody, usage,
		200, durationMs, c.ClientIP(),
		"third_party", &route.ProviderID,
	)
}

// ──────────────────────────────────
// 请求/响应头处理
// ──────────────────────────────────

// 需要跳过的请求头（hop-by-hop 或安全相关）
var skipRequestHeaders = map[string]bool{
	"Authorization":    true,
	"X-Api-Key":        true,
	"Host":             true,
	"X-Codemind-Proxy": true,
	"Connection":       true,
	"Transfer-Encoding": true,
}

// copyRequestHeaders 透传客户端请求头到代理请求
func copyRequestHeaders(src *http.Request, dst *http.Request) {
	for key, values := range src.Header {
		if skipRequestHeaders[http.CanonicalHeaderKey(key)] {
			continue
		}
		for _, v := range values {
			dst.Header.Add(key, v)
		}
	}
}

// 需要跳过的响应头
var skipResponseHeaders = map[string]bool{
	"Transfer-Encoding": true,
	"Connection":        true,
}

// copyResponseHeaders 复制上游响应头到客户端响应（仅非流式场景）
func copyResponseHeaders(resp *http.Response, c *gin.Context) {
	for key, vals := range resp.Header {
		if skipResponseHeaders[http.CanonicalHeaderKey(key)] {
			continue
		}
		if http.CanonicalHeaderKey(key) == "Content-Type" || http.CanonicalHeaderKey(key) == "Content-Length" {
			continue
		}
		for _, v := range vals {
			c.Writer.Header().Add(key, v)
		}
	}
}

// extractUsageFromResponse 从 JSON 响应体中提取 usage 信息（兼容 OpenAI 和 Anthropic 格式）
func extractUsageFromResponse(body []byte) *llm.Usage {
	// OpenAI 格式：{"usage": {"prompt_tokens": ..., "completion_tokens": ...}}
	var openai struct {
		Usage *llm.Usage `json:"usage"`
	}
	if json.Unmarshal(body, &openai) == nil && openai.Usage != nil && openai.Usage.TotalTokens > 0 {
		return openai.Usage
	}

	// Anthropic 格式：{"usage": {"input_tokens": ..., "output_tokens": ...}}
	var anthropic struct {
		Usage *llm.AnthropicUsage `json:"usage"`
	}
	if json.Unmarshal(body, &anthropic) == nil && anthropic.Usage != nil {
		if anthropic.Usage.InputTokens+anthropic.Usage.OutputTokens > 0 {
			return anthropic.Usage.ToUsage()
		}
	}
	return nil
}

// extractUsageFromSSEData 从 SSE data 行中提取 usage 信息
// 兼容 OpenAI（prompt_tokens）和 Anthropic（input_tokens / message.usage）两种格式
func extractUsageFromSSEData(data []byte) *llm.Usage {
	// OpenAI 格式：data 行顶层 usage
	var openai struct {
		Usage *llm.Usage `json:"usage"`
	}
	if json.Unmarshal(data, &openai) == nil && openai.Usage != nil && openai.Usage.TotalTokens > 0 {
		return openai.Usage
	}

	// Anthropic 格式：顶层 usage（message_delta 事件）或 message.usage（message_start 事件）
	var anthropic struct {
		Usage   *llm.AnthropicUsage `json:"usage"`
		Message *struct {
			Usage *llm.AnthropicUsage `json:"usage"`
		} `json:"message"`
	}
	if json.Unmarshal(data, &anthropic) == nil {
		if anthropic.Usage != nil && anthropic.Usage.InputTokens+anthropic.Usage.OutputTokens > 0 {
			return anthropic.Usage.ToUsage()
		}
		if anthropic.Message != nil && anthropic.Message.Usage != nil &&
			anthropic.Message.Usage.InputTokens+anthropic.Message.Usage.OutputTokens > 0 {
			return anthropic.Message.Usage.ToUsage()
		}
	}
	return nil
}

// extractContentFromSSEData 从 SSE data 行中提取文本内容
// 兼容 OpenAI（choices[0].delta.content）和 Anthropic（delta.text）两种格式
func extractContentFromSSEData(data []byte) string {
	// OpenAI 格式：choices[0].delta.content
	var openai struct {
		Choices []struct {
			Delta *struct {
				Content string `json:"content"`
			} `json:"delta"`
		} `json:"choices"`
	}
	if json.Unmarshal(data, &openai) == nil && len(openai.Choices) > 0 && openai.Choices[0].Delta != nil {
		return openai.Choices[0].Delta.Content
	}

	// Anthropic 格式：content_block_delta 事件中的 delta.text
	var anthropic struct {
		Delta *struct {
			Text string `json:"text"`
		} `json:"delta"`
	}
	if json.Unmarshal(data, &anthropic) == nil && anthropic.Delta != nil {
		return anthropic.Delta.Text
	}

	return ""
}

// ──────────────────────────────────
// 集成点：在各代理方法中调用
// ──────────────────────────────────

// checkAndHandleThirdParty 检查是否为第三方模型并处理
// requestFormat: "openai" 或 "anthropic"，由端点决定
// 返回 true 表示已处理（第三方模型或协议不兼容），false 表示需要走内置路由
func (h *LLMProxyHandler) checkAndHandleThirdParty(
	c *gin.Context, rawBody []byte, meta *llm.RequestMeta,
	endpointPath string, requestFormat string,
	userID, apiKeyID int64, startTime time.Time,
) bool {
	tpService := h.proxyService.GetThirdPartyService()
	if tpService == nil {
		return false
	}

	route := tpService.ResolveThirdPartyModel(c.Request.Context(), userID, meta.Model, requestFormat)
	if route == nil {
		return false
	}

	// 校验协议格式兼容性：第三方服务必须支持客户端请求的协议
	if !route.IsFormatCompatible(requestFormat) {
		h.logger.Warn("第三方服务协议不兼容",
			zap.String("model", meta.Model),
			zap.String("provider", route.ProviderName),
			zap.String("provider_format", route.Format),
			zap.String("request_format", requestFormat),
		)
		h.sendFormatError(c, requestFormat, http.StatusBadRequest,
			fmt.Sprintf("模型 %q 所属的第三方服务 %q 不支持 %s 协议，请通过该服务支持的协议端点访问",
				meta.Model, route.ProviderName, strings.ToUpper(requestFormat)))
		return true
	}

	// 第三方模型 — 透明代理，不走配额和并发控制
	h.handleThirdPartyProxy(c, rawBody, meta, route, endpointPath, requestFormat, userID, apiKeyID, startTime)
	return true
}

// sendFormatError 根据客户端请求的协议格式发送对应格式的错误响应
func (h *LLMProxyHandler) sendFormatError(c *gin.Context, format string, status int, msg string) {
	if format == "anthropic" {
		h.sendAnthropicError(c, status, "api_error", msg)
	} else {
		h.sendOpenAIError(c, status, "server_error", msg)
	}
}
