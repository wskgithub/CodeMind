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
	return c.doRequest("/v1/messages", req, true)
}

// doRequest 发送 HTTP 请求到 Anthropic 服务
func (c *AnthropicClient) doRequest(path string, payload interface{}, isStream bool) (io.ReadCloser, error) {
	var bodyReader io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("序列化请求体失败: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	url := c.baseURL + path
	req, err := http.NewRequest("POST", url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("创建 HTTP 请求失败: %w", err)
	}

	// 设置 Anthropic 专用请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("anthropic-version", AnthropicAPIVersion)
	if c.apiKey != "" {
		req.Header.Set("x-api-key", c.apiKey)
	}

	// 流式请求使用长超时客户端
	client := c.httpClient
	if isStream {
		client = c.streamClient
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求 Anthropic 服务失败: %w", err)
	}

	// 处理非 2xx 响应
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
