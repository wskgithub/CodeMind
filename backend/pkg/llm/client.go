package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
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
// baseURL 支持带或不带 /v1 后缀，内部统一去掉尾部的 /v1 及多余的斜杠，
// 后续调用时会再拼接 /v1/chat/completions 等路径，避免重复。
func NewClient(baseURL, apiKey string, timeoutSec, streamTimeoutSec int) *Client {
	baseURL = normalizeBaseURL(baseURL)
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

// normalizeBaseURL 去掉 baseURL 末尾的 /v1（及任意 /vN）和多余的斜杠，
// 使得拼接 /v1/chat/completions 等路径时不会出现双重 /v1。
// 示例：
//
//	"http://host:8000/v1"  → "http://host:8000"
//	"http://host:8000/v1/" → "http://host:8000"
//	"http://host:8000"     → "http://host:8000"
func normalizeBaseURL(u string) string {
	u = strings.TrimRight(u, "/")
	// 去掉末尾形如 /v1、/v2 … /v999 的版本段
	for {
		idx := strings.LastIndex(u, "/")
		if idx < 0 {
			break
		}
		seg := u[idx+1:]
		// 匹配 vN 格式（v1、v2 等）
		if len(seg) >= 2 && seg[0] == 'v' {
			isVer := true
			for _, c := range seg[1:] {
				if c < '0' || c > '9' {
					isVer = false
					break
				}
			}
			if isVer {
				u = u[:idx]
				continue
			}
		}
		break
	}
	return strings.TrimRight(u, "/")
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

// RetrieveModel 获取单个模型信息
func (c *Client) RetrieveModel(modelID string) (*ModelInfo, error) {
	body, err := c.doRequest("GET", "/v1/models/"+modelID, nil, false)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var resp ModelInfo
	if err := json.NewDecoder(body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("解析模型信息失败: %w", err)
	}

	return &resp, nil
}

// ──────────────────────────────────
// 原始请求体透传方法
// 直接转发客户端的完整 JSON，不做结构体序列化/反序列化，
// 确保 tools、stream_options 等所有字段完整保留
// ──────────────────────────────────

// ChatCompletionRawAll 非流式对话补全（原始请求体透传）
// 返回原始响应体和解析的 Usage，保留响应中的所有字段（包括 tool_calls 等）
func (c *Client) ChatCompletionRawAll(rawBody []byte) (rawResp []byte, usage *Usage, err error) {
	body, err := c.doRequestRaw("POST", "/v1/chat/completions", rawBody, false)
	if err != nil {
		return nil, nil, err
	}
	defer body.Close()

	rawResp, err = io.ReadAll(body)
	if err != nil {
		return nil, nil, fmt.Errorf("读取 LLM 响应失败: %w", err)
	}

	usage = ExtractUsageFromResponse(rawResp)
	return rawResp, usage, nil
}

// ChatCompletionStreamRaw 流式对话补全（原始请求体透传）
// 自动注入 stream_options.include_usage 以确保 LLM 返回用量信息
func (c *Client) ChatCompletionStreamRaw(rawBody []byte) (io.ReadCloser, error) {
	rawBody = EnsureStreamOptions(rawBody)
	return c.doRequestRaw("POST", "/v1/chat/completions", rawBody, true)
}

// CompletionRawAll 非流式文本补全（原始请求体透传）
func (c *Client) CompletionRawAll(rawBody []byte) (rawResp []byte, usage *Usage, err error) {
	body, err := c.doRequestRaw("POST", "/v1/completions", rawBody, false)
	if err != nil {
		return nil, nil, err
	}
	defer body.Close()

	rawResp, err = io.ReadAll(body)
	if err != nil {
		return nil, nil, fmt.Errorf("读取 LLM 响应失败: %w", err)
	}

	usage = ExtractUsageFromResponse(rawResp)
	return rawResp, usage, nil
}

// CompletionStreamRaw 流式文本补全（原始请求体透传）
func (c *Client) CompletionStreamRaw(rawBody []byte) (io.ReadCloser, error) {
	rawBody = EnsureStreamOptions(rawBody)
	return c.doRequestRaw("POST", "/v1/completions", rawBody, true)
}

// ResponsesRaw 非流式 Responses API（原始请求体透传）
func (c *Client) ResponsesRaw(rawBody []byte) (rawResp []byte, usage *Usage, err error) {
	body, err := c.doRequestRaw("POST", "/v1/responses", rawBody, false)
	if err != nil {
		return nil, nil, err
	}
	defer body.Close()

	rawResp, err = io.ReadAll(body)
	if err != nil {
		return nil, nil, fmt.Errorf("读取 LLM 响应失败: %w", err)
	}

	usage = ExtractUsageFromResponsesBody(rawResp)
	return rawResp, usage, nil
}

// ResponsesStreamRaw 流式 Responses API（原始请求体透传）
func (c *Client) ResponsesStreamRaw(rawBody []byte) (io.ReadCloser, error) {
	return c.doRequestRaw("POST", "/v1/responses", rawBody, true)
}

// EmbeddingRaw 向量嵌入（原始请求体透传）
func (c *Client) EmbeddingRaw(rawBody []byte) (rawResp []byte, usage *Usage, err error) {
	body, err := c.doRequestRaw("POST", "/v1/embeddings", rawBody, false)
	if err != nil {
		return nil, nil, err
	}
	defer body.Close()

	rawResp, err = io.ReadAll(body)
	if err != nil {
		return nil, nil, fmt.Errorf("读取 LLM 响应失败: %w", err)
	}

	usage = ExtractUsageFromResponse(rawResp)
	return rawResp, usage, nil
}

// ──────────────────────────────────
// HTTP 传输层
// ──────────────────────────────────

// doRequest 发送结构体请求到 LLM 服务（序列化后转发）
func (c *Client) doRequest(method, path string, payload interface{}, isStream bool) (io.ReadCloser, error) {
	var rawBody []byte
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("序列化请求体失败: %w", err)
		}
		rawBody = data
	}
	return c.doRequestRaw(method, path, rawBody, isStream)
}

// doRequestRaw 发送原始字节请求到 LLM 服务
// 这是所有 HTTP 请求的底层方法，doRequest 也基于此方法实现
func (c *Client) doRequestRaw(method, path string, rawBody []byte, isStream bool) (io.ReadCloser, error) {
	var bodyReader io.Reader
	if rawBody != nil {
		bodyReader = bytes.NewReader(rawBody)
	}

	url := c.baseURL + path
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("创建 HTTP 请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	// 标记此请求由 CodeMind 转发，用于下游检测自环
	req.Header.Set("X-CodeMind-Proxy", "1")
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
