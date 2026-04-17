package llm

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// StreamReader parses SSE stream responses from LLM services
type StreamReader struct {
	reader *bufio.Reader
	body   io.ReadCloser
	done   bool
}

// NewStreamReader creates an SSE stream reader
func NewStreamReader(body io.ReadCloser) *StreamReader {
	return &StreamReader{
		reader: bufio.NewReaderSize(body, 8192),
		body:   body,
	}
}

// ReadChunk reads the next SSE data chunk, returns io.EOF when stream ends
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

		line = strings.TrimRight(line, "\r\n")

		if line == "" {
			continue
		}

		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")

		if data == "[DONE]" {
			s.done = true
			return "data: [DONE]", nil, io.EOF
		}

		var c ChatCompletionChunk
		if err := json.Unmarshal([]byte(data), &c); err != nil {
			return fmt.Sprintf("data: %s", data), nil, nil
		}

		return fmt.Sprintf("data: %s", data), &c, nil
	}
}

// Close closes the underlying connection
func (s *StreamReader) Close() error {
	s.done = true
	return s.body.Close()
}

// IsDone returns whether the stream has ended
func (s *StreamReader) IsDone() bool {
	return s.done
}
