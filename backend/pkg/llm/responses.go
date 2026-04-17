package llm

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// ResponsesStreamReader reads and forwards Responses API SSE events,
// extracting usage info from response.completed event.
type ResponsesStreamReader struct {
	reader *bufio.Reader
	body   io.ReadCloser
	done   bool
}

// NewResponsesStreamReader creates a Responses API stream reader.
func NewResponsesStreamReader(body io.ReadCloser) *ResponsesStreamReader {
	return &ResponsesStreamReader{
		reader: bufio.NewReaderSize(body, 16384), //nolint:mnd // intentional constant.
		body:   body,
	}
}

// ReadEvent reads the next SSE event.
// Returns event type, raw SSE text, and parsed data line.
// Returns io.EOF when stream ends.
func (r *ResponsesStreamReader) ReadEvent() (eventType string, rawLines string, dataPayload []byte, err error) {
	if r.done {
		return "", "", nil, io.EOF
	}

	var sb strings.Builder
	var currentEvent string
	var currentData string

	for {
		line, err := r.reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				r.done = true
				if currentEvent != "" || currentData != "" {
					return currentEvent, sb.String(), []byte(currentData), nil
				}
			}
			return "", "", nil, err
		}

		sb.WriteString(line)
		trimmed := strings.TrimRight(line, "\r\n")

		if trimmed == "" {
			if currentEvent != "" || currentData != "" {
				return currentEvent, sb.String(), []byte(currentData), nil
			}
			continue
		}

		if strings.HasPrefix(trimmed, "event: ") {
			currentEvent = strings.TrimPrefix(trimmed, "event: ")
		} else if strings.HasPrefix(trimmed, "data: ") {
			currentData = strings.TrimPrefix(trimmed, "data: ")
		}
	}
}

// Close closes the underlying connection.
func (r *ResponsesStreamReader) Close() error {
	r.done = true
	return r.body.Close()
}

// IsDone returns whether the stream has ended.
func (r *ResponsesStreamReader) IsDone() bool {
	return r.done
}

type responsesUsageWrapper struct {
	Response struct {
		Usage *ResponsesUsage `json:"usage"`
	} `json:"response"`
}

// ResponsesUsage represents Responses API token usage format.
type ResponsesUsage struct {
	InputTokensDetails  *ResponsesInputTokenDetails  `json:"input_tokens_details,omitempty"`
	OutputTokensDetails *ResponsesOutputTokenDetails `json:"output_tokens_details,omitempty"`
	InputTokens         int                          `json:"input_tokens"`
	OutputTokens        int                          `json:"output_tokens"`
	TotalTokens         int                          `json:"total_tokens"`
}

// ResponsesInputTokenDetails represents Responses API input token details.
type ResponsesInputTokenDetails struct {
	CachedTokens int `json:"cached_tokens,omitempty"`
}

// ResponsesOutputTokenDetails represents Responses API output token details.
type ResponsesOutputTokenDetails struct {
	ReasoningTokens int `json:"reasoning_tokens,omitempty"`
}

// ToUsage converts Responses API usage to common Usage format.
func (u *ResponsesUsage) ToUsage() *Usage {
	if u == nil {
		return nil
	}
	usage := &Usage{
		PromptTokens:     u.InputTokens,
		CompletionTokens: u.OutputTokens,
		TotalTokens:      u.TotalTokens,
	}
	if u.OutputTokensDetails != nil && u.OutputTokensDetails.ReasoningTokens > 0 {
		usage.CompletionTokensDetails = &CompletionTokensDetails{
			ReasoningTokens: u.OutputTokensDetails.ReasoningTokens,
		}
	}
	if u.InputTokensDetails != nil && u.InputTokensDetails.CachedTokens > 0 {
		usage.PromptTokensDetails = &PromptTokensDetails{
			CachedTokens: u.InputTokensDetails.CachedTokens,
		}
	}
	return usage
}

// ExtractUsageFromResponsesEvent extracts usage from response.completed event data payload.
func ExtractUsageFromResponsesEvent(dataPayload []byte) *Usage {
	var wrapper responsesUsageWrapper
	if err := json.Unmarshal(dataPayload, &wrapper); err != nil {
		return nil
	}
	if wrapper.Response.Usage == nil {
		return nil
	}
	return wrapper.Response.Usage.ToUsage()
}

// ExtractUsageFromResponsesBody extracts usage from non-streaming Responses API response body.
func ExtractUsageFromResponsesBody(rawResp []byte) *Usage {
	var wrapper struct {
		Usage *ResponsesUsage `json:"usage"`
	}
	if err := json.Unmarshal(rawResp, &wrapper); err != nil {
		return nil
	}
	if wrapper.Usage == nil {
		return nil
	}
	return wrapper.Usage.ToUsage()
}

// ExtractModelFromResponsesBody extracts model name from Responses API request/response body.
func ExtractModelFromResponsesBody(rawBody []byte) string {
	var wrapper struct {
		Model string `json:"model"`
	}
	if err := json.Unmarshal(rawBody, &wrapper); err != nil {
		return ""
	}
	return wrapper.Model
}

// Responses API SSE 事件类型常量。
const (
	ResponsesEventCreated               = "response.created"
	ResponsesEventInProgress            = "response.in_progress"
	ResponsesEventCompleted             = "response.completed"
	ResponsesEventFailed                = "response.failed"
	ResponsesEventIncomplete            = "response.incomplete"
	ResponsesEventOutputTextDelta       = "response.output_text.delta"
	ResponsesEventOutputTextDone        = "response.output_text.done"
	ResponsesEventOutputItemAdded       = "response.output_item.added"
	ResponsesEventOutputItemDone        = "response.output_item.done"
	ResponsesEventFunctionCallArgsDelta = "response.function_call_arguments.delta"
	ResponsesEventFunctionCallArgsDone  = "response.function_call_arguments.done"
	ResponsesEventError                 = "error"
)

// ResponsesErrorResponse represents Responses API error response format.
type ResponsesErrorResponse struct {
	Type  string `json:"type"`
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// NewResponsesError creates a Responses API format error response.
func NewResponsesError(code, message string) ResponsesErrorResponse {
	return ResponsesErrorResponse{
		Type: "error",
		Error: struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		}{
			Code:    code,
			Message: message,
		},
	}
}

// FormatResponsesSSEError formats error as SSE error event.
func FormatResponsesSSEError(code, message string) string {
	errJSON, _ := json.Marshal(map[string]interface{}{
		"type":    "error",
		"code":    code,
		"message": message,
		"param":   nil,
	})
	return fmt.Sprintf("event: error\ndata: %s\n\n", string(errJSON))
}
