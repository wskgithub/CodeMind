package llm

import (
	"bufio"
	"encoding/json"
	"io"
	"strings"
)

// AnthropicStreamReader reads Anthropic SSE streaming responses.
type AnthropicStreamReader struct {
	reader *bufio.Reader
	body   io.ReadCloser
	done   bool
}

// NewAnthropicStreamReader creates an Anthropic SSE stream reader.
func NewAnthropicStreamReader(body io.ReadCloser) *AnthropicStreamReader {
	return &AnthropicStreamReader{
		reader: bufio.NewReaderSize(body, 8192),
		body:   body,
	}
}

// ReadEvent reads the next SSE event.
// Returns event type, raw SSE text, and parsed event data.
// Returns io.EOF when stream ends.
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
				if currentEvent != "" && currentData != "" {
					return s.parseEventData(currentEvent, currentData, rawBuilder.String())
				}
			}
			return "", "", nil, readErr
		}

		trimmed := strings.TrimRight(line, "\r\n")

		if trimmed == "" {
			if currentEvent != "" && currentData != "" {
				return s.parseEventData(currentEvent, currentData, rawBuilder.String())
			}
			currentEvent = ""
			currentData = ""
			rawBuilder.Reset()
			continue
		}

		rawBuilder.WriteString(line)

		if strings.HasPrefix(trimmed, "event: ") {
			currentEvent = strings.TrimPrefix(trimmed, "event: ")
			continue
		}

		if strings.HasPrefix(trimmed, "data: ") {
			currentData = strings.TrimPrefix(trimmed, "data: ")
			continue
		}
	}
}

func (s *AnthropicStreamReader) parseEventData(eventType, data, rawLines string) (string, string, *AnthropicStreamEvent, error) {
	if eventType == AnthropicEventMessageStop {
		s.done = true
		return eventType, rawLines, &AnthropicStreamEvent{Type: eventType}, nil
	}

	if eventType == AnthropicEventPing {
		return eventType, rawLines, &AnthropicStreamEvent{Type: eventType}, nil
	}

	var event AnthropicStreamEvent
	if err := json.Unmarshal([]byte(data), &event); err != nil {
		return eventType, rawLines, &AnthropicStreamEvent{Type: eventType}, nil
	}
	event.Type = eventType

	return eventType, rawLines, &event, nil
}

// Close closes the underlying connection.
func (s *AnthropicStreamReader) Close() error {
	s.done = true
	return s.body.Close()
}

// IsDone returns whether the stream has ended.
func (s *AnthropicStreamReader) IsDone() bool {
	return s.done
}
