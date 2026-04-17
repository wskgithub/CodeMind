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

var thirdPartyHTTPClient = &http.Client{
	Timeout: 600 * time.Second,
	Transport: &http.Transport{
		MaxIdleConnsPerHost: 20,
		IdleConnTimeout:     120 * time.Second,
	},
}

// handleThirdPartyProxy handles third-party model service transparent proxy.
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
	apiKey, err := h.proxyService.GetThirdPartyService().DecryptAPIKey(route.APIKeyEncrypted)
	if err != nil {
		h.logger.Error("failed to decrypt third-party API key",
			zap.Int64("provider_id", route.ProviderID),
			zap.Error(err),
		)
		h.sendFormatError(c, requestFormat, http.StatusInternalServerError, "third-party service configuration error")
		return
	}

	baseURL := route.BaseURLForFormat(requestFormat)
	targetURL := strings.TrimRight(baseURL, "/") + endpointPath

	proxyReq, err := http.NewRequestWithContext(c.Request.Context(), c.Request.Method, targetURL, bytes.NewReader(rawBody))
	if err != nil {
		h.sendFormatError(c, requestFormat, http.StatusInternalServerError, "failed to build proxy request")
		return
	}

	copyRequestHeaders(c.Request, proxyReq)

	if requestFormat == "anthropic" {
		proxyReq.Header.Set("x-api-key", apiKey)
	} else {
		proxyReq.Header.Set("Authorization", "Bearer "+apiKey)
	}

	h.logger.Debug("proxying request to third-party service",
		zap.String("model", meta.Model),
		zap.String("provider", route.ProviderName),
		zap.String("target_url", targetURL),
		zap.Bool("stream", meta.Stream),
	)

	resp, err := thirdPartyHTTPClient.Do(proxyReq)
	if err != nil {
		durationMs := int(time.Since(startTime).Milliseconds())
		h.logger.Error("third-party service request failed", zap.Error(err), zap.String("url", targetURL))
		go h.proxyService.RecordRequestLog(userID, apiKeyID, "chat_completion", meta.Model, 502, err.Error(), c.ClientIP(), c.Request.UserAgent(), durationMs)
		h.sendFormatError(c, requestFormat, http.StatusBadGateway, "third-party service unavailable: "+err.Error())
		return
	}
	defer resp.Body.Close()

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

// pipeThirdPartyNonStreamResponse pipes non-streaming response.
func (h *LLMProxyHandler) pipeThirdPartyNonStreamResponse(
	c *gin.Context, resp *http.Response,
	route *model.ThirdPartyRouteInfo, meta *llm.RequestMeta,
	requestBody []byte, requestFormat string, userID, apiKeyID int64, startTime time.Time,
) {
	respBody, err := io.ReadAll(resp.Body)
	durationMs := int(time.Since(startTime).Milliseconds())

	if err != nil {
		h.logger.Error("failed to read third-party response", zap.Error(err))
		h.sendFormatError(c, requestFormat, http.StatusBadGateway, "failed to read third-party service response")
		return
	}

	usage := extractUsageFromResponse(respBody)

	go h.recordThirdPartyMetrics(route, meta, userID, apiKeyID, usage, durationMs, c, requestBody, respBody, false)

	copyResponseHeaders(resp, c)
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)
}

// pipeThirdPartyStreamResponse pipes streaming SSE response.
func (h *LLMProxyHandler) pipeThirdPartyStreamResponse(
	c *gin.Context, resp *http.Response,
	route *model.ThirdPartyRouteInfo, meta *llm.RequestMeta,
	requestBody []byte, requestFormat string, userID, apiKeyID int64, startTime time.Time,
) {
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

		if len(bytes.TrimRight(line, "\r\n")) == 0 {
			c.Writer.(http.Flusher).Flush()
		}
	}
	c.Writer.(http.Flusher).Flush()

	durationMs := int(time.Since(startTime).Milliseconds())

	streamResult := &llm.StreamResult{
		Content:      contentBuilder.String(),
		Model:        meta.Model,
		Usage:        lastUsage,
		FinishReason: "stop",
	}
	assembledResp := llm.AssembleChatResponse(streamResult)

	go h.recordThirdPartyMetrics(route, meta, userID, apiKeyID, lastUsage, durationMs, c, requestBody, assembledResp, true)
}

// pipeThirdPartyErrorResponse pipes third-party service error response.
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

	copyResponseHeaders(resp, c)
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)
}

// recordThirdPartyMetrics records third-party service metrics.
func (h *LLMProxyHandler) recordThirdPartyMetrics(
	route *model.ThirdPartyRouteInfo, meta *llm.RequestMeta,
	userID, apiKeyID int64, usage *llm.Usage, durationMs int,
	c *gin.Context, requestBody, responseBody []byte, isStream bool,
) {
	tpService := h.proxyService.GetThirdPartyService()

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

	h.proxyService.RecordRequestLog(
		userID, apiKeyID, "chat_completion", meta.Model,
		200, "", c.ClientIP(), c.Request.UserAgent(), durationMs,
	)

	h.proxyService.RecordTrainingDataWithSource(
		userID, apiKeyID, "chat_completion", meta.Model,
		isStream, requestBody, responseBody, usage,
		200, durationMs, c.ClientIP(),
		"third_party", &route.ProviderID,
	)
}

var skipRequestHeaders = map[string]bool{
	"Authorization":    true,
	"X-Api-Key":        true,
	"Host":             true,
	"X-Codemind-Proxy": true,
	"Connection":       true,
	"Transfer-Encoding": true,
}

// copyRequestHeaders copies client request headers to proxy request.
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

var skipResponseHeaders = map[string]bool{
	"Transfer-Encoding": true,
	"Connection":        true,
}

// copyResponseHeaders copies upstream response headers to client response.
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

// extractUsageFromResponse extracts usage info from JSON response body.
func extractUsageFromResponse(body []byte) *llm.Usage {
	var openai struct {
		Usage *llm.Usage `json:"usage"`
	}
	if json.Unmarshal(body, &openai) == nil && openai.Usage != nil && openai.Usage.TotalTokens > 0 {
		return openai.Usage
	}

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

// extractUsageFromSSEData extracts usage info from SSE data line.
func extractUsageFromSSEData(data []byte) *llm.Usage {
	var openai struct {
		Usage *llm.Usage `json:"usage"`
	}
	if json.Unmarshal(data, &openai) == nil && openai.Usage != nil && openai.Usage.TotalTokens > 0 {
		return openai.Usage
	}

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

// extractContentFromSSEData extracts text content from SSE data line.
func extractContentFromSSEData(data []byte) string {
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

// checkAndHandleThirdParty checks if model is third-party and handles it.
// Returns true if handled (third-party model), false if needs built-in routing.
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

	if !route.IsFormatCompatible(requestFormat) {
		h.logger.Warn("third-party service protocol incompatible",
			zap.String("model", meta.Model),
			zap.String("provider", route.ProviderName),
			zap.String("provider_format", route.Format),
			zap.String("request_format", requestFormat),
		)
		h.sendFormatError(c, requestFormat, http.StatusBadRequest,
			fmt.Sprintf("model %q's third-party service %q does not support %s protocol",
				meta.Model, route.ProviderName, strings.ToUpper(requestFormat)))
		return true
	}

	h.handleThirdPartyProxy(c, rawBody, meta, route, endpointPath, requestFormat, userID, apiKeyID, startTime)
	return true
}

// sendFormatError sends error response in the appropriate format.
func (h *LLMProxyHandler) sendFormatError(c *gin.Context, format string, status int, msg string) {
	if format == "anthropic" {
		h.sendAnthropicError(c, status, "api_error", msg)
	} else {
		h.sendOpenAIError(c, status, "server_error", msg)
	}
}
