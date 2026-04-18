package llm

import (
	"io"
	"strings"
	"testing"
)

// TestStreamReader tests SSE stream reader.
func TestStreamReader(t *testing.T) {
	// Simulate SSE data stream
	sseData := `data: {"id":"1","object":"chat.completion.chunk","model":"gpt-4","choices":[{"index":0,"delta":{"role":"assistant","content":"Hello"}}]}

data: {"id":"1","object":"chat.completion.chunk","model":"gpt-4","choices":[{"index":0,"delta":{"content":" world"}}]}

data: [DONE]

`
	body := io.NopCloser(strings.NewReader(sseData))
	reader := NewStreamReader(body)
	defer reader.Close()

	// First chunk
	rawLine, chunk, err := reader.ReadChunk()
	if err != nil {
		t.Fatalf("failed to read first chunk: %v", err)
	}
	if !strings.HasPrefix(rawLine, "data: ") {
		t.Errorf("raw line should start with 'data: ': %s", rawLine)
	}
	if chunk == nil {
		t.Fatal("chunk should not be nil")
	}
	if chunk.ID != "1" {
		t.Errorf("incorrect chunk ID: %s", chunk.ID)
	}

	// Second chunk
	_, chunk2, err := reader.ReadChunk()
	if err != nil {
		t.Fatalf("failed to read second chunk: %v", err)
	}
	if chunk2 == nil {
		t.Fatal("second chunk should not be nil")
	}

	// [DONE] marker
	_, _, err = reader.ReadChunk()
	if err != io.EOF {
		t.Errorf("should return io.EOF, got: %v", err)
	}

	// Stream ended
	if !reader.IsDone() {
		t.Error("reader should be marked as done")
	}
}

// TestStreamReaderEmpty tests empty stream.
func TestStreamReaderEmpty(t *testing.T) {
	body := io.NopCloser(strings.NewReader(""))
	reader := NewStreamReader(body)
	defer reader.Close()

	_, _, err := reader.ReadChunk()
	if err != io.EOF {
		t.Errorf("empty stream should return io.EOF, got: %v", err)
	}
}

// TestStreamReaderInvalidJSON tests unparseable JSON.
func TestStreamReaderInvalidJSON(t *testing.T) {
	sseData := "data: {invalid json}\n\n"
	body := io.NopCloser(strings.NewReader(sseData))
	reader := NewStreamReader(body)
	defer reader.Close()

	rawLine, chunk, err := reader.ReadChunk()
	if err != nil {
		t.Fatalf("should not return error: %v", err)
	}
	// Should return raw line but chunk is nil
	if rawLine == "" {
		t.Error("raw line should not be empty")
	}
	if chunk != nil {
		t.Error("chunk should be nil for invalid JSON")
	}
}

// TestStreamReaderSkipsComments tests skipping non-data lines.
func TestStreamReaderSkipsComments(t *testing.T) {
	sseData := `: this is a comment
event: message
data: {"id":"1","object":"chat.completion.chunk","model":"gpt-4","choices":[{"index":0,"delta":{"content":"Hi"}}]}

data: [DONE]

`
	body := io.NopCloser(strings.NewReader(sseData))
	reader := NewStreamReader(body)
	defer reader.Close()

	// Should skip comments and event lines, read directly to first data line
	_, chunk, err := reader.ReadChunk()
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if chunk == nil {
		t.Fatal("chunk should not be nil")
	}
}
