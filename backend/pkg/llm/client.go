package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client LLM 服务 HTTP 客户端
type Client struct {
	baseURL       string
	apiKey        string
	httpClient    *http.Client
	streamClient  *http.Client
}

// NewClient 创建 LLM 客户端
func NewClient(baseURL, apiKey string, timeoutSec, streamTimeoutSec int) *Client {
	return &Client{
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

// ChatCompletion 非流式对话补全
func (c *Client) ChatCompletion(req *ChatCompletionRequest) (*ChatCompletionResponse, error) {
	req.Stream = false

	body, err := c.doRequest("POST", "/v1/chat/completions", req, false)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var resp ChatCompletionResponse
	if err := json.NewDecoder(body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("解析 LLM 响应失败: %w", err)
	}

	return &resp, nil
}

// ChatCompletionStream 流式对话补全，返回响应体供逐行读取
func (c *Client) ChatCompletionStream(req *ChatCompletionRequest) (io.ReadCloser, error) {
	req.Stream = true
	return c.doRequest("POST", "/v1/chat/completions", req, true)
}

// Completion 非流式文本补全
func (c *Client) Completion(req *CompletionRequest) (*CompletionResponse, error) {
	req.Stream = false

	body, err := c.doRequest("POST", "/v1/completions", req, false)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var resp CompletionResponse
	if err := json.NewDecoder(body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("解析 LLM 响应失败: %w", err)
	}

	return &resp, nil
}

// CompletionStream 流式文本补全
func (c *Client) CompletionStream(req *CompletionRequest) (io.ReadCloser, error) {
	req.Stream = true
	return c.doRequest("POST", "/v1/completions", req, true)
}

// ListModels 获取模型列表
func (c *Client) ListModels() (*ModelListResponse, error) {
	body, err := c.doRequest("GET", "/v1/models", nil, false)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var resp ModelListResponse
	if err := json.NewDecoder(body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("解析模型列表失败: %w", err)
	}

	return &resp, nil
}

// doRequest 发送 HTTP 请求到 LLM 服务
func (c *Client) doRequest(method, path string, payload interface{}, isStream bool) (io.ReadCloser, error) {
	var bodyReader io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("序列化请求体失败: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	url := c.baseURL + path
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("创建 HTTP 请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	// 流式请求使用长超时客户端
	client := c.httpClient
	if isStream {
		client = c.streamClient
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求 LLM 服务失败: %w", err)
	}

	// 处理非 2xx 响应
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		bodyBytes, _ := io.ReadAll(resp.Body)

		// 根据 LLM 返回状态码映射错误
		switch {
		case resp.StatusCode == 429:
			return nil, &LLMError{StatusCode: 503, Message: "LLM 服务繁忙，请稍后重试", Body: bodyBytes}
		case resp.StatusCode >= 500:
			return nil, &LLMError{StatusCode: 502, Message: "LLM 服务内部错误", Body: bodyBytes}
		default:
			return nil, &LLMError{StatusCode: resp.StatusCode, Message: "LLM 服务请求失败", Body: bodyBytes}
		}
	}

	return resp.Body, nil
}

// LLMError LLM 服务错误
type LLMError struct {
	StatusCode int
	Message    string
	Body       []byte
}

func (e *LLMError) Error() string {
	return fmt.Sprintf("LLM 错误 (HTTP %d): %s", e.StatusCode, e.Message)
}
