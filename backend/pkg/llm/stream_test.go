package llm

import (
	"io"
	"strings"
	"testing"
)

// TestStreamReader 测试 SSE 流式读取器
func TestStreamReader(t *testing.T) {
	// 模拟 SSE 数据流
	sseData := `data: {"id":"1","object":"chat.completion.chunk","model":"gpt-4","choices":[{"index":0,"delta":{"role":"assistant","content":"Hello"}}]}

data: {"id":"1","object":"chat.completion.chunk","model":"gpt-4","choices":[{"index":0,"delta":{"content":" world"}}]}

data: [DONE]

`
	body := io.NopCloser(strings.NewReader(sseData))
	reader := NewStreamReader(body)
	defer reader.Close()

	// 第一个 chunk
	rawLine, chunk, err := reader.ReadChunk()
	if err != nil {
		t.Fatalf("读取第一个 chunk 失败: %v", err)
	}
	if !strings.HasPrefix(rawLine, "data: ") {
		t.Errorf("原始行应以 'data: ' 开头: %s", rawLine)
	}
	if chunk == nil {
		t.Fatal("chunk 不应为 nil")
	}
	if chunk.ID != "1" {
		t.Errorf("chunk ID 不正确: %s", chunk.ID)
	}

	// 第二个 chunk
	_, chunk2, err := reader.ReadChunk()
	if err != nil {
		t.Fatalf("读取第二个 chunk 失败: %v", err)
	}
	if chunk2 == nil {
		t.Fatal("第二个 chunk 不应为 nil")
	}

	// [DONE] 标记
	_, _, err = reader.ReadChunk()
	if err != io.EOF {
		t.Errorf("应返回 io.EOF, 实际: %v", err)
	}

	// 已结束
	if !reader.IsDone() {
		t.Error("reader 应已标记为 done")
	}
}

// TestStreamReaderEmpty 测试空流
func TestStreamReaderEmpty(t *testing.T) {
	body := io.NopCloser(strings.NewReader(""))
	reader := NewStreamReader(body)
	defer reader.Close()

	_, _, err := reader.ReadChunk()
	if err != io.EOF {
		t.Errorf("空流应返回 io.EOF, 实际: %v", err)
	}
}

// TestStreamReaderInvalidJSON 测试无法解析的 JSON
func TestStreamReaderInvalidJSON(t *testing.T) {
	sseData := "data: {invalid json}\n\n"
	body := io.NopCloser(strings.NewReader(sseData))
	reader := NewStreamReader(body)
	defer reader.Close()

	rawLine, chunk, err := reader.ReadChunk()
	if err != nil {
		t.Fatalf("不应返回错误: %v", err)
	}
	// 应返回原始行但 chunk 为 nil
	if rawLine == "" {
		t.Error("原始行不应为空")
	}
	if chunk != nil {
		t.Error("无效 JSON 的 chunk 应为 nil")
	}
}

// TestStreamReaderSkipsComments 测试跳过非 data 行
func TestStreamReaderSkipsComments(t *testing.T) {
	sseData := `: this is a comment
event: message
data: {"id":"1","object":"chat.completion.chunk","model":"gpt-4","choices":[{"index":0,"delta":{"content":"Hi"}}]}

data: [DONE]

`
	body := io.NopCloser(strings.NewReader(sseData))
	reader := NewStreamReader(body)
	defer reader.Close()

	// 应跳过注释和 event 行，直接读到第一个 data 行
	_, chunk, err := reader.ReadChunk()
	if err != nil {
		t.Fatalf("读取失败: %v", err)
	}
	if chunk == nil {
		t.Fatal("chunk 不应为 nil")
	}
}
