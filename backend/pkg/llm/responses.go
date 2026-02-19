package llm

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// ──────────────────────────────────
// OpenAI Responses API 流式读取器
//
// Responses API 的 SSE 格式与 Chat Completions 不同：
//   - 使用命名事件: event: response.output_text.delta\ndata: {...}\n\n
//   - 最终事件 response.completed 包含完整的 usage 信息
//   - 没有 data: [DONE] 结束标记
//
// 参考: https://developers.openai.com/docs/api-reference/responses/streaming
// ──────────────────────────────────

// ResponsesStreamReader 读取并转发 Responses API 的 SSE 事件，
// 同时从 response.completed 事件中提取 usage 信息
type ResponsesStreamReader struct {
	reader *bufio.Reader
	body   io.ReadCloser
	done   bool
}

// NewResponsesStreamReader 创建 Responses API 流式读取器
func NewResponsesStreamReader(body io.ReadCloser) *ResponsesStreamReader {
	return &ResponsesStreamReader{
		reader: bufio.NewReaderSize(body, 16384),
		body:   body,
	}
}

// ReadEvent 读取下一个 SSE 事件
// 返回事件类型、原始 SSE 文本（可直接转发）和解析后的数据行
// 当流结束时返回 io.EOF
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
				// 处理缓冲区中剩余的事件
				if currentEvent != "" || currentData != "" {
					return currentEvent, sb.String(), []byte(currentData), nil
				}
			}
			return "", "", nil, err
		}

		sb.WriteString(line)
		trimmed := strings.TrimRight(line, "\r\n")

		// 空行表示事件结束
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

// Close 关闭底层连接
func (r *ResponsesStreamReader) Close() error {
	r.done = true
	return r.body.Close()
}

// IsDone 流是否已结束
func (r *ResponsesStreamReader) IsDone() bool {
	return r.done
}

// ──────────────────────────────────
// Responses API Usage 提取
// ──────────────────────────────────

// responsesUsageWrapper 用于从 response.completed 事件中提取 usage
type responsesUsageWrapper struct {
	Response struct {
		Usage *ResponsesUsage `json:"usage"`
	} `json:"response"`
}

// ResponsesUsage Responses API 的 token 用量格式
type ResponsesUsage struct {
	InputTokens        int                         `json:"input_tokens"`
	InputTokensDetails *ResponsesInputTokenDetails `json:"input_tokens_details,omitempty"`
	OutputTokens       int                         `json:"output_tokens"`
	OutputTokensDetails *ResponsesOutputTokenDetails `json:"output_tokens_details,omitempty"`
	TotalTokens        int                         `json:"total_tokens"`
}

// ResponsesInputTokenDetails Responses API 输入 token 详情
type ResponsesInputTokenDetails struct {
	CachedTokens int `json:"cached_tokens,omitempty"`
}

// ResponsesOutputTokenDetails Responses API 输出 token 详情
type ResponsesOutputTokenDetails struct {
	ReasoningTokens int `json:"reasoning_tokens,omitempty"`
}

// ToUsage 将 Responses API 用量转换为通用 Usage 格式
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

// ExtractUsageFromResponsesEvent 从 response.completed 事件的 data payload 中提取 usage
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

// ExtractUsageFromResponsesBody 从非流式 Responses API 响应体中提取 usage
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

// ExtractModelFromResponsesBody 从 Responses API 请求/响应体中提取模型名
func ExtractModelFromResponsesBody(rawBody []byte) string {
	var wrapper struct {
		Model string `json:"model"`
	}
	if err := json.Unmarshal(rawBody, &wrapper); err != nil {
		return ""
	}
	return wrapper.Model
}

// ──────────────────────────────────
// Responses API 常量
// ──────────────────────────────────

const (
	// ResponsesEventCreated 响应创建事件
	ResponsesEventCreated = "response.created"
	// ResponsesEventInProgress 响应处理中事件
	ResponsesEventInProgress = "response.in_progress"
	// ResponsesEventCompleted 响应完成事件（含 usage）
	ResponsesEventCompleted = "response.completed"
	// ResponsesEventFailed 响应失败事件
	ResponsesEventFailed = "response.failed"
	// ResponsesEventIncomplete 响应不完整事件
	ResponsesEventIncomplete = "response.incomplete"
	// ResponsesEventOutputTextDelta 文本增量事件
	ResponsesEventOutputTextDelta = "response.output_text.delta"
	// ResponsesEventOutputTextDone 文本完成事件
	ResponsesEventOutputTextDone = "response.output_text.done"
	// ResponsesEventOutputItemAdded 输出项添加事件
	ResponsesEventOutputItemAdded = "response.output_item.added"
	// ResponsesEventOutputItemDone 输出项完成事件
	ResponsesEventOutputItemDone = "response.output_item.done"
	// ResponsesEventFunctionCallArgsDelta 函数调用参数增量事件
	ResponsesEventFunctionCallArgsDelta = "response.function_call_arguments.delta"
	// ResponsesEventFunctionCallArgsDone 函数调用参数完成事件
	ResponsesEventFunctionCallArgsDone = "response.function_call_arguments.done"
	// ResponsesEventError 错误事件
	ResponsesEventError = "error"
)

// ResponsesErrorResponse Responses API 错误响应格式
type ResponsesErrorResponse struct {
	Type  string `json:"type"`
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// NewResponsesError 创建 Responses API 格式的错误响应
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

// FormatResponsesSSEError 格式化为 SSE 错误事件
func FormatResponsesSSEError(code, message string) string {
	errJSON, _ := json.Marshal(map[string]interface{}{
		"type":    "error",
		"code":    code,
		"message": message,
		"param":   nil,
	})
	return fmt.Sprintf("event: error\ndata: %s\n\n", string(errJSON))
}
