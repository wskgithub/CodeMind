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

// Client is an HTTP client for LLM services.
type Client struct {
	httpClient   *http.Client
	streamClient *http.Client
	baseURL      string
	apiKey       string
}

// NewClient creates an LLM client (normalizes baseURL to avoid duplicate /v1).
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

// normalizeBaseURL strips trailing /vN and slashes from baseURL.
func normalizeBaseURL(u string) string {
	u = strings.TrimRight(u, "/")
	for {
		idx := strings.LastIndex(u, "/")
		if idx < 0 {
			break
		}
		seg := u[idx+1:]
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

// ChatCompletion performs a non-streaming chat completion.
func (c *Client) ChatCompletion(req *ChatCompletionRequest) (*ChatCompletionResponse, error) {
	req.Stream = false

	body, err := c.doRequest("POST", "/v1/chat/completions", req, false)
	if err != nil {
		return nil, err
	}
	defer func() { _ = body.Close() }()

	var resp ChatCompletionResponse
	if err := json.NewDecoder(body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	return &resp, nil
}

// ChatCompletionStream performs a streaming chat completion.
func (c *Client) ChatCompletionStream(req *ChatCompletionRequest) (io.ReadCloser, error) {
	req.Stream = true
	return c.doRequest("POST", "/v1/chat/completions", req, true)
}

// Completion performs a non-streaming text completion.
func (c *Client) Completion(req *CompletionRequest) (*CompletionResponse, error) {
	req.Stream = false

	body, err := c.doRequest("POST", "/v1/completions", req, false)
	if err != nil {
		return nil, err
	}
	defer func() { _ = body.Close() }()

	var resp CompletionResponse
	if err := json.NewDecoder(body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	return &resp, nil
}

// CompletionStream performs a streaming text completion.
func (c *Client) CompletionStream(req *CompletionRequest) (io.ReadCloser, error) {
	req.Stream = true
	return c.doRequest("POST", "/v1/completions", req, true)
}

// ListModels retrieves available models.
func (c *Client) ListModels() (*ModelListResponse, error) {
	body, err := c.doRequest("GET", "/v1/models", nil, false)
	if err != nil {
		return nil, err
	}
	defer func() { _ = body.Close() }()

	var resp ModelListResponse
	if err := json.NewDecoder(body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("failed to parse model list: %w", err)
	}

	return &resp, nil
}

// RetrieveModel retrieves a single model's info.
func (c *Client) RetrieveModel(modelID string) (*ModelInfo, error) {
	body, err := c.doRequest("GET", "/v1/models/"+modelID, nil, false)
	if err != nil {
		return nil, err
	}
	defer func() { _ = body.Close() }()

	var resp ModelInfo
	if err := json.NewDecoder(body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("failed to parse model info: %w", err)
	}

	return &resp, nil
}

// ChatCompletionRawAll performs non-streaming chat completion with raw request passthrough.
func (c *Client) ChatCompletionRawAll(rawBody []byte) (rawResp []byte, usage *Usage, err error) {
	body, err := c.doRequestRaw("POST", "/v1/chat/completions", rawBody, false)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = body.Close() }()

	rawResp, err = io.ReadAll(body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read LLM response: %w", err)
	}

	usage = ExtractUsageFromResponse(rawResp)
	return rawResp, usage, nil
}

// ChatCompletionStreamRaw performs streaming chat completion with raw request passthrough.
func (c *Client) ChatCompletionStreamRaw(rawBody []byte) (io.ReadCloser, error) {
	rawBody = EnsureStreamOptions(rawBody)
	return c.doRequestRaw("POST", "/v1/chat/completions", rawBody, true)
}

// CompletionRawAll performs non-streaming text completion with raw request passthrough.
func (c *Client) CompletionRawAll(rawBody []byte) (rawResp []byte, usage *Usage, err error) {
	body, err := c.doRequestRaw("POST", "/v1/completions", rawBody, false)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = body.Close() }()

	rawResp, err = io.ReadAll(body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read LLM response: %w", err)
	}

	usage = ExtractUsageFromResponse(rawResp)
	return rawResp, usage, nil
}

// CompletionStreamRaw performs streaming text completion with raw request passthrough.
func (c *Client) CompletionStreamRaw(rawBody []byte) (io.ReadCloser, error) {
	rawBody = EnsureStreamOptions(rawBody)
	return c.doRequestRaw("POST", "/v1/completions", rawBody, true)
}

// ResponsesRaw performs non-streaming Responses API call with raw request passthrough.
func (c *Client) ResponsesRaw(rawBody []byte) (rawResp []byte, usage *Usage, err error) {
	body, err := c.doRequestRaw("POST", "/v1/responses", rawBody, false)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = body.Close() }()

	rawResp, err = io.ReadAll(body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read LLM response: %w", err)
	}

	usage = ExtractUsageFromResponsesBody(rawResp)
	return rawResp, usage, nil
}

// ResponsesStreamRaw performs streaming Responses API call with raw request passthrough.
func (c *Client) ResponsesStreamRaw(rawBody []byte) (io.ReadCloser, error) {
	return c.doRequestRaw("POST", "/v1/responses", rawBody, true)
}

// EmbeddingRaw performs embedding with raw request passthrough.
func (c *Client) EmbeddingRaw(rawBody []byte) (rawResp []byte, usage *Usage, err error) {
	body, err := c.doRequestRaw("POST", "/v1/embeddings", rawBody, false)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = body.Close() }()

	rawResp, err = io.ReadAll(body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read LLM response: %w", err)
	}

	usage = ExtractUsageFromResponse(rawResp)
	return rawResp, usage, nil
}

func (c *Client) doRequest(method, path string, payload interface{}, isStream bool) (io.ReadCloser, error) {
	var rawBody []byte
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize request body: %w", err)
		}
		rawBody = data
	}
	return c.doRequestRaw(method, path, rawBody, isStream)
}

func (c *Client) doRequestRaw(method, path string, rawBody []byte, isStream bool) (io.ReadCloser, error) {
	var bodyReader io.Reader
	if rawBody != nil {
		bodyReader = bytes.NewReader(rawBody)
	}

	url := c.baseURL + path
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CodeMind-Proxy", "1")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	client := c.httpClient
	if isStream {
		client = c.streamClient
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("LLM request failed: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer func() { _ = resp.Body.Close() }()
		bodyBytes, _ := io.ReadAll(resp.Body)

		switch {
		case resp.StatusCode == 429: //nolint:mnd // intentional constant.
			return nil, &LLMError{StatusCode: 503, Message: "LLM service busy, try again later", Body: bodyBytes} //nolint:mnd // intentional constant.
		case resp.StatusCode >= 500: //nolint:mnd // intentional constant.
			return nil, &LLMError{StatusCode: 502, Message: "LLM service internal error", Body: bodyBytes} //nolint:mnd // intentional constant.
		default:
			return nil, &LLMError{StatusCode: resp.StatusCode, Message: "LLM request failed", Body: bodyBytes}
		}
	}

	return resp.Body, nil
}

// LLMError represents an LLM service error.
type LLMError struct {
	Message    string
	Body       []byte
	StatusCode int
}

func (e *LLMError) Error() string {
	return fmt.Sprintf("LLM error (HTTP %d): %s", e.StatusCode, e.Message)
}
