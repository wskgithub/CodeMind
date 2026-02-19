package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Anthropic API 版本常量
const (
	AnthropicAPIVersion = "2023-06-01"
)

// AnthropicClient Anthropic 原生 API 客户端
type AnthropicClient struct {
	baseURL      string
	apiKey       string
	httpClient   *http.Client
	streamClient *http.Client
}

// NewAnthropicClient 创建 Anthropic 客户端
func NewAnthropicClient(baseURL, apiKey string, timeoutSec, streamTimeoutSec int) *AnthropicClient {
	return &AnthropicClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: time.Duration(timeoutSec) * time.Second,
		},
		streamClient: &http.Client{
			Timeout: time.Duration(streamTimeoutSec) * time.Second,
		},
	}
}

// Messages 非流式消息调用
func (c *AnthropicClient) Messages(req *AnthropicMessagesRequest) (*AnthropicMessagesResponse, error) {
	req.Stream = false

	body, err := c.doRequest("/v1/messages", req, false)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var resp AnthropicMessagesResponse
	if err := json.NewDecoder(body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("解析 Anthropic 响应失败: %w", err)
	}

	return &resp, nil
}

// MessagesStream 流式消息调用，返回 SSE 事件流
func (c *AnthropicClient) MessagesStream(req *AnthropicMessagesRequest) (io.ReadCloser, error) {
	req.Stream = true
	return c.doRequest("/v1/messages", req, false)
}

// MessagesRaw 原始请求体非流式调用
// 直接转发原始 JSON 请求体，避免结构体序列化导致的字段丢失
func (c *AnthropicClient) MessagesRaw(rawBody []byte, extraHeaders map[string]string) ([]byte, *AnthropicUsage, error) {
	body, err := c.doRequestRaw("/v1/messages", rawBody, false, extraHeaders)
	if err != nil {
		return nil, nil, err
	}
	defer body.Close()

	respBytes, err := io.ReadAll(body)
	if err != nil {
		return nil, nil, fmt.Errorf("读取 Anthropic 响应失败: %w", err)
	}

	// 轻量提取 usage 用于用量统计，无需完整反序列化
	var partial struct {
		Usage *AnthropicUsage `json:"usage"`
	}
	json.Unmarshal(respBytes, &partial)

	return respBytes, partial.Usage, nil
}

// MessagesStreamRaw 原始请求体流式调用
// 直接转发原始 JSON 请求体，返回 SSE 事件流
func (c *AnthropicClient) MessagesStreamRaw(rawBody []byte, extraHeaders map[string]string) (io.ReadCloser, error) {
	return c.doRequestRaw("/v1/messages", rawBody, true, extraHeaders)
}

// doRequest 发送结构体请求到 Anthropic 服务（内部跨格式转换使用）
func (c *AnthropicClient) doRequest(path string, payload interface{}, isStream bool) (io.ReadCloser, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("序列化请求体失败: %w", err)
	}
	return c.doRequestRaw(path, data, isStream, nil)
}

// doRequestRaw 发送原始字节请求到 Anthropic 服务
// extraHeaders 用于转发客户端请求头（如 anthropic-beta）
func (c *AnthropicClient) doRequestRaw(path string, rawBody []byte, isStream bool, extraHeaders map[string]string) (io.ReadCloser, error) {
	url := c.baseURL + path
	req, err := http.NewRequest("POST", url, bytes.NewReader(rawBody))
	if err != nil {
		return nil, fmt.Errorf("创建 HTTP 请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("anthropic-version", AnthropicAPIVersion)
	if c.apiKey != "" {
		req.Header.Set("x-api-key", c.apiKey)
	}

	// 应用客户端转发的额外请求头（可覆盖默认值）
	for k, v := range extraHeaders {
		req.Header.Set(k, v)
	}

	client := c.httpClient
	if isStream {
		client = c.streamClient
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求 Anthropic 服务失败: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		bodyBytes, _ := io.ReadAll(resp.Body)

		switch {
		case resp.StatusCode == 429:
			return nil, &LLMError{StatusCode: 529, Message: "Anthropic 服务过载，请稍后重试", Body: bodyBytes}
		case resp.StatusCode >= 500:
			return nil, &LLMError{StatusCode: 502, Message: "Anthropic 服务内部错误", Body: bodyBytes}
		default:
			return nil, &LLMError{StatusCode: resp.StatusCode, Message: "Anthropic 服务请求失败", Body: bodyBytes}
		}
	}

	return resp.Body, nil
}
