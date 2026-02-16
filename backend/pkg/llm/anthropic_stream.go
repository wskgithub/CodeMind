package llm

import (
	"bufio"
	"encoding/json"
	"io"
	"strings"
)

// AnthropicStreamReader Anthropic SSE 流式响应读取器
// Anthropic 使用 "event: xxx\ndata: {...}\n\n" 格式
type AnthropicStreamReader struct {
	reader *bufio.Reader
	body   io.ReadCloser
	done   bool
}

// NewAnthropicStreamReader 创建 Anthropic SSE 流式读取器
func NewAnthropicStreamReader(body io.ReadCloser) *AnthropicStreamReader {
	return &AnthropicStreamReader{
		reader: bufio.NewReaderSize(body, 8192),
		body:   body,
	}
}

// ReadEvent 读取下一个 SSE 事件
// 返回事件类型、原始 SSE 文本（含 event: 和 data: 行）、以及解析后的事件数据
// 当流结束时返回 io.EOF
func (s *AnthropicStreamReader) ReadEvent() (eventType string, rawLines string, event *AnthropicStreamEvent, err error) {
	if s.done {
		return "", "", nil, io.EOF
	}

	var currentEvent string
	var currentData string
	var rawBuilder strings.Builder

	for {
		line, readErr := s.reader.ReadString('\n')
		if readErr != nil {
			if readErr == io.EOF {
				s.done = true
				// 如果有未处理的数据，先返回它
				if currentEvent != "" && currentData != "" {
					return s.parseEventData(currentEvent, currentData, rawBuilder.String())
				}
			}
			return "", "", nil, readErr
		}

		// 去除尾部换行
		trimmed := strings.TrimRight(line, "\r\n")

		// 空行表示事件结束
		if trimmed == "" {
			if currentEvent != "" && currentData != "" {
				return s.parseEventData(currentEvent, currentData, rawBuilder.String())
			}
			// 重置，继续读下一个事件
			currentEvent = ""
			currentData = ""
			rawBuilder.Reset()
			continue
		}

		rawBuilder.WriteString(line)

		// 解析 event: 行
		if strings.HasPrefix(trimmed, "event: ") {
			currentEvent = strings.TrimPrefix(trimmed, "event: ")
			continue
		}

		// 解析 data: 行
		if strings.HasPrefix(trimmed, "data: ") {
			currentData = strings.TrimPrefix(trimmed, "data: ")
			continue
		}
	}
}

// parseEventData 解析事件数据
func (s *AnthropicStreamReader) parseEventData(eventType, data, rawLines string) (string, string, *AnthropicStreamEvent, error) {
	// 消息结束事件
	if eventType == AnthropicEventMessageStop {
		s.done = true
		return eventType, rawLines, &AnthropicStreamEvent{Type: eventType}, nil
	}

	// ping 事件无需解析 data
	if eventType == AnthropicEventPing {
		return eventType, rawLines, &AnthropicStreamEvent{Type: eventType}, nil
	}

	// 解析 JSON 数据
	var event AnthropicStreamEvent
	if err := json.Unmarshal([]byte(data), &event); err != nil {
		// 解析失败仍返回原始数据，让调用方决定处理方式
		return eventType, rawLines, &AnthropicStreamEvent{Type: eventType}, nil
	}
	event.Type = eventType

	return eventType, rawLines, &event, nil
}

// Close 关闭底层连接
func (s *AnthropicStreamReader) Close() error {
	s.done = true
	return s.body.Close()
}

// IsDone 流是否已结束
func (s *AnthropicStreamReader) IsDone() bool {
	return s.done
}
