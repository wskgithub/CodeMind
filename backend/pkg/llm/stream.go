package llm

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// StreamReader SSE 流式响应读取器
// 逐行解析 LLM 返回的 Server-Sent Events 数据
type StreamReader struct {
	reader  *bufio.Reader
	body    io.ReadCloser
	done    bool
}

// NewStreamReader 创建 SSE 流式读取器
func NewStreamReader(body io.ReadCloser) *StreamReader {
	return &StreamReader{
		reader: bufio.NewReaderSize(body, 8192),
		body:   body,
	}
}

// ReadChunk 读取下一个 SSE 数据块
// 返回原始 SSE 行（含 "data: " 前缀）和解析后的 chunk
// 当流结束时返回 io.EOF
func (s *StreamReader) ReadChunk() (rawLine string, chunk *ChatCompletionChunk, err error) {
	if s.done {
		return "", nil, io.EOF
	}

	for {
		line, err := s.reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				s.done = true
			}
			return "", nil, err
		}

		// 去除尾部换行
		line = strings.TrimRight(line, "\r\n")

		// 跳过空行
		if line == "" {
			continue
		}

		// SSE 数据行以 "data: " 开头
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")

		// 检查流结束标记
		if data == "[DONE]" {
			s.done = true
			return "data: [DONE]", nil, io.EOF
		}

		// 解析 JSON chunk
		var c ChatCompletionChunk
		if err := json.Unmarshal([]byte(data), &c); err != nil {
			// 解析失败时仍然返回原始行，让调用方决定是否转发
			return fmt.Sprintf("data: %s", data), nil, nil
		}

		return fmt.Sprintf("data: %s", data), &c, nil
	}
}

// Close 关闭底层连接
func (s *StreamReader) Close() error {
	s.done = true
	return s.body.Close()
}

// IsDone 流是否已结束
func (s *StreamReader) IsDone() bool {
	return s.done
}
