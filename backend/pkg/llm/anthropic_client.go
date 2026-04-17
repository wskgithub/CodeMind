package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Anthropic API version constant.
const (
	AnthropicAPIVersion = "2023-06-01"
)

// AnthropicClient is the native Anthropic API client.
type AnthropicClient struct {
	httpClient   *http.Client
	streamClient *http.Client
	baseURL      string
	apiKey       string
}

// NewAnthropicClient creates an Anthropic client.
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

// Messages performs non-streaming message call.
func (c *AnthropicClient) Messages(req *AnthropicMessagesRequest) (*AnthropicMessagesResponse, error) {
	req.Stream = false

	body, err := c.doRequest("/v1/messages", req, false)
	if err != nil {
		return nil, err
	}
	defer func() { _ = body.Close() }()

	var resp AnthropicMessagesResponse
	if err := json.NewDecoder(body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("failed to parse Anthropic response: %w", err)
	}

	return &resp, nil
}

// MessagesStream performs streaming message call, returns SSE event stream.
func (c *AnthropicClient) MessagesStream(req *AnthropicMessagesRequest) (io.ReadCloser, error) {
	req.Stream = true
	return c.doRequest("/v1/messages", req, false)
}

// MessagesRaw performs non-streaming call with raw request body.
func (c *AnthropicClient) MessagesRaw(rawBody []byte, extraHeaders map[string]string) ([]byte, *AnthropicUsage, error) {
	body, err := c.doRequestRaw("/v1/messages", rawBody, false, extraHeaders)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = body.Close() }()

	respBytes, err := io.ReadAll(body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read Anthropic response: %w", err)
	}

	var partial struct {
		Usage *AnthropicUsage `json:"usage"`
	}
	_ = json.Unmarshal(respBytes, &partial)

	return respBytes, partial.Usage, nil
}

// MessagesStreamRaw performs streaming call with raw request body.
func (c *AnthropicClient) MessagesStreamRaw(rawBody []byte, extraHeaders map[string]string) (io.ReadCloser, error) {
	return c.doRequestRaw("/v1/messages", rawBody, true, extraHeaders)
}

func (c *AnthropicClient) doRequest(path string, payload interface{}, isStream bool) (io.ReadCloser, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}
	return c.doRequestRaw(path, data, isStream, nil)
}

func (c *AnthropicClient) doRequestRaw(path string, rawBody []byte, isStream bool, extraHeaders map[string]string) (io.ReadCloser, error) {
	url := c.baseURL + path
	req, err := http.NewRequest("POST", url, bytes.NewReader(rawBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("anthropic-version", AnthropicAPIVersion)
	if c.apiKey != "" {
		req.Header.Set("x-api-key", c.apiKey)
	}

	for k, v := range extraHeaders {
		req.Header.Set(k, v)
	}

	client := c.httpClient
	if isStream {
		client = c.streamClient
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to request Anthropic service: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer func() { _ = resp.Body.Close() }()
		bodyBytes, _ := io.ReadAll(resp.Body)

		switch {
		case resp.StatusCode == 429: //nolint:mnd // intentional constant.
			return nil, &Error{StatusCode: 529, Message: "Anthropic service overloaded, please retry later", Body: bodyBytes} //nolint:mnd // intentional constant.
		case resp.StatusCode >= 500: //nolint:mnd // intentional constant.
			return nil, &Error{StatusCode: 502, Message: "Anthropic service internal error", Body: bodyBytes} //nolint:mnd // intentional constant.
		default:
			return nil, &Error{StatusCode: resp.StatusCode, Message: "Anthropic service request failed", Body: bodyBytes}
		}
	}

	return resp.Body, nil
}
