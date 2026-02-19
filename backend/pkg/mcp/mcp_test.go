package mcp

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// ═══════════════════════════════════════════════════════════
// Protocol Tests - JSON-RPC 消息序列化和反序列化
// ═══════════════════════════════════════════════════════════

func TestJSONRPCRequest_MarshalUnmarshal(t *testing.T) {
	tests := []struct {
		name     string
		request  JSONRPCRequest
		expected string
	}{
		{
			name: "基本请求",
			request: JSONRPCRequest{
				JSONRPC: "2.0",
				ID:      "req-1",
				Method:  "tools/list",
				Params:  json.RawMessage(`{}`),
			},
			expected: `{"jsonrpc":"2.0","id":"req-1","method":"tools/list","params":{}}`,
		},
		{
			name: "数字ID请求",
			request: JSONRPCRequest{
				JSONRPC: "2.0",
				ID:      123,
				Method:  "initialize",
				Params:  nil,
			},
			expected: `{"jsonrpc":"2.0","id":123,"method":"initialize"}`,
		},
		{
			name: "带参数的请求",
			request: JSONRPCRequest{
				JSONRPC: "2.0",
				ID:      1,
				Method:  "tools/call",
				Params:  json.RawMessage(`{"name":"test-tool","arguments":{"key":"value"}}`),
			},
			expected: `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"test-tool","arguments":{"key":"value"}}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal
			data, err := json.Marshal(tt.request)
			require.NoError(t, err)
			assert.JSONEq(t, tt.expected, string(data))

			// Unmarshal
			var decoded JSONRPCRequest
			err = json.Unmarshal(data, &decoded)
			require.NoError(t, err)
			assert.Equal(t, tt.request.JSONRPC, decoded.JSONRPC)
			assert.Equal(t, tt.request.Method, decoded.Method)
			assert.Equal(t, string(tt.request.Params), string(decoded.Params))
		})
	}
}

func TestJSONRPCResponse_MarshalUnmarshal(t *testing.T) {
	tests := []struct {
		name     string
		response JSONRPCResponse
	}{
		{
			name: "成功响应",
			response: JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      "req-1",
				Result:  json.RawMessage(`{"tools":[]}`),
			},
		},
		{
			name: "错误响应",
			response: JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      float64(123), // JSON 数字解码为 float64
				Error: &JSONRPCError{
					Code:    ErrCodeMethodNotFound,
					Message: "方法不存在",
				},
			},
		},
		{
			name: "带数据的错误响应",
			response: JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      float64(456), // JSON 数字解码为 float64
				Error: &JSONRPCError{
					Code:    ErrCodeInvalidParams,
					Message: "参数错误",
					Data:    map[string]string{"field": "name"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal
			data, err := json.Marshal(tt.response)
			require.NoError(t, err)

			// Unmarshal
			var decoded JSONRPCResponse
			err = json.Unmarshal(data, &decoded)
			require.NoError(t, err)
			assert.Equal(t, tt.response.JSONRPC, decoded.JSONRPC)
			assert.Equal(t, tt.response.ID, decoded.ID)
			assert.Equal(t, string(tt.response.Result), string(decoded.Result))
			if tt.response.Error != nil {
				require.NotNil(t, decoded.Error)
				assert.Equal(t, tt.response.Error.Code, decoded.Error.Code)
				assert.Equal(t, tt.response.Error.Message, decoded.Error.Message)
			}
		})
	}
}

func TestJSONRPCNotification_MarshalUnmarshal(t *testing.T) {
	notification := JSONRPCNotification{
		JSONRPC: "2.0",
		Method:  "notifications/initialized",
		Params:  json.RawMessage(`{}`),
	}

	// Marshal
	data, err := json.Marshal(notification)
	require.NoError(t, err)
	assert.NotContains(t, string(data), "id") // 通知不应该有 id 字段

	// Unmarshal
	var decoded JSONRPCNotification
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, notification.Method, decoded.Method)
}

func TestNewResponse(t *testing.T) {
	result := ToolsListResult{
		Tools: []Tool{
			{Name: "test-tool", Description: "Test tool"},
		},
	}

	resp := NewResponse("req-1", result)
	require.NotNil(t, resp)
	assert.Equal(t, JSONRPCVersion, resp.JSONRPC)
	assert.Equal(t, "req-1", resp.ID)
	assert.NotNil(t, resp.Result)
	assert.Nil(t, resp.Error)

	// 验证 Result 内容
	var decodedResult ToolsListResult
	err := json.Unmarshal(resp.Result, &decodedResult)
	require.NoError(t, err)
	assert.Len(t, decodedResult.Tools, 1)
	assert.Equal(t, "test-tool", decodedResult.Tools[0].Name)
}

func TestNewErrorResponse(t *testing.T) {
	resp := NewErrorResponse("req-1", ErrCodeMethodNotFound, "方法不存在: test")
	require.NotNil(t, resp)
	assert.Equal(t, JSONRPCVersion, resp.JSONRPC)
	assert.Equal(t, "req-1", resp.ID)
	assert.Nil(t, resp.Result)
	require.NotNil(t, resp.Error)
	assert.Equal(t, ErrCodeMethodNotFound, resp.Error.Code)
	assert.Equal(t, "方法不存在: test", resp.Error.Message)
}

func TestErrorCodes(t *testing.T) {
	// 验证标准 JSON-RPC 错误码
	assert.Equal(t, -32700, ErrCodeParseError)
	assert.Equal(t, -32600, ErrCodeInvalidRequest)
	assert.Equal(t, -32601, ErrCodeMethodNotFound)
	assert.Equal(t, -32602, ErrCodeInvalidParams)
	assert.Equal(t, -32603, ErrCodeInternalError)
}

func TestMethodConstants(t *testing.T) {
	// 验证 MCP 标准方法名
	assert.Equal(t, "initialize", MethodInitialize)
	assert.Equal(t, "notifications/initialized", MethodInitialized)
	assert.Equal(t, "tools/list", MethodToolsList)
	assert.Equal(t, "tools/call", MethodToolsCall)
	assert.Equal(t, "resources/list", MethodResourcesList)
	assert.Equal(t, "resources/read", MethodResourcesRead)
	assert.Equal(t, "prompts/list", MethodPromptsList)
	assert.Equal(t, "prompts/get", MethodPromptsGet)
	assert.Equal(t, "ping", MethodPing)
}

// ═══════════════════════════════════════════════════════════
// MCP 类型序列化测试
// ═══════════════════════════════════════════════════════════

func TestInitializeParams_Marshal(t *testing.T) {
	params := InitializeParams{
		ProtocolVersion: MCPProtocolVersion,
		Capabilities:    ClientCapability{},
		ClientInfo: Implementation{
			Name:    "test-client",
			Version: "1.0.0",
		},
	}

	data, err := json.Marshal(params)
	require.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, MCPProtocolVersion, decoded["protocolVersion"])
	assert.Equal(t, "test-client", decoded["clientInfo"].(map[string]interface{})["name"])
}

func TestInitializeResult_Marshal(t *testing.T) {
	result := InitializeResult{
		ProtocolVersion: MCPProtocolVersion,
		Capabilities: ServerCapability{
			Tools:     &ToolsCapability{ListChanged: true},
			Resources: &ResourcesCapability{ListChanged: true},
		},
		ServerInfo: Implementation{
			Name:    "test-server",
			Version: "2.0.0",
		},
	}

	data, err := json.Marshal(result)
	require.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, MCPProtocolVersion, decoded["protocolVersion"])
	capabilities := decoded["capabilities"].(map[string]interface{})
	assert.NotNil(t, capabilities["tools"])
	assert.NotNil(t, capabilities["resources"])
}

func TestTool_Marshal(t *testing.T) {
	tool := Tool{
		Name:        "test-tool",
		Description: "A test tool",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"name": map[string]string{"type": "string"},
			},
		},
	}

	data, err := json.Marshal(tool)
	require.NoError(t, err)

	var decoded Tool
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, tool.Name, decoded.Name)
	assert.Equal(t, tool.Description, decoded.Description)
}

func TestToolCallParams_Marshal(t *testing.T) {
	params := ToolCallParams{
		Name:      "test-tool",
		Arguments: json.RawMessage(`{"key":"value","num":123}`),
	}

	data, err := json.Marshal(params)
	require.NoError(t, err)

	var decoded ToolCallParams
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, params.Name, decoded.Name)
	assert.JSONEq(t, string(params.Arguments), string(decoded.Arguments))
}

func TestToolCallResult_Marshal(t *testing.T) {
	result := ToolCallResult{
		Content: []ToolContent{
			{Type: "text", Text: "Hello"},
			{Type: "text", Text: "World"},
		},
		IsError: false,
	}

	data, err := json.Marshal(result)
	require.NoError(t, err)

	var decoded ToolCallResult
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Len(t, decoded.Content, 2)
	assert.Equal(t, "Hello", decoded.Content[0].Text)
	assert.False(t, decoded.IsError)
}

func TestResource_Marshal(t *testing.T) {
	resource := Resource{
		URI:         "file:///test.txt",
		Name:        "test.txt",
		Description: "Test file",
		MimeType:    "text/plain",
	}

	data, err := json.Marshal(resource)
	require.NoError(t, err)

	var decoded Resource
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, resource.URI, decoded.URI)
	assert.Equal(t, resource.MimeType, decoded.MimeType)
}

func TestResourceReadParams_Marshal(t *testing.T) {
	params := ResourceReadParams{
		URI: "file:///test.txt",
	}

	data, err := json.Marshal(params)
	require.NoError(t, err)

	var decoded ResourceReadParams
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, params.URI, decoded.URI)
}

func TestResourceReadResult_Marshal(t *testing.T) {
	result := ResourceReadResult{
		Contents: []ResourceContent{
			{
				URI:      "file:///test.txt",
				MimeType: "text/plain",
				Text:     "Hello, World!",
			},
		},
	}

	data, err := json.Marshal(result)
	require.NoError(t, err)

	var decoded ResourceReadResult
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Len(t, decoded.Contents, 1)
	assert.Equal(t, "Hello, World!", decoded.Contents[0].Text)
}

// ═══════════════════════════════════════════════════════════
// SSE Session Tests
// ═══════════════════════════════════════════════════════════

func TestNewSSESession(t *testing.T) {
	// 使用支持 Flusher 的 ResponseWriter
	w := httptest.NewRecorder()
	
	session, err := NewSSESession(w, "http://localhost/messages")
	require.NoError(t, err)
	assert.NotEmpty(t, session.ID)
	assert.NotNil(t, session.Writer)
	assert.NotNil(t, session.Flusher)
	assert.NotNil(t, session.Done)
	assert.Contains(t, session.MessageURL, "http://localhost/messages")
	assert.Contains(t, session.MessageURL, "sessionId=")
	assert.WithinDuration(t, time.Now(), session.CreatedAt, time.Second)
}

func TestSSESession_SendEvent(t *testing.T) {
	w := httptest.NewRecorder()
	
	session, err := NewSSESession(w, "http://localhost/messages")
	require.NoError(t, err)

	err = session.SendEvent("message", `{"test":"data"}`)
	require.NoError(t, err)

	// 验证输出格式
	body := w.Body.String()
	assert.Contains(t, body, "event: message")
	assert.Contains(t, body, `data: {"test":"data"}`)
}

func TestSSESession_SendEndpoint(t *testing.T) {
	w := httptest.NewRecorder()
	
	session, err := NewSSESession(w, "http://localhost/messages")
	require.NoError(t, err)

	err = session.SendEndpoint()
	require.NoError(t, err)

	body := w.Body.String()
	assert.Contains(t, body, "event: endpoint")
	assert.Contains(t, body, session.MessageURL)
}

func TestSSESession_SendMessage(t *testing.T) {
	w := httptest.NewRecorder()
	
	session, err := NewSSESession(w, "http://localhost/messages")
	require.NoError(t, err)

	msg := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      1,
		Result:  json.RawMessage(`{"tools":[]}`),
	}

	err = session.SendMessage(msg)
	require.NoError(t, err)

	body := w.Body.String()
	assert.Contains(t, body, "event: message")
	assert.Contains(t, body, `"jsonrpc":"2.0"`)
}

func TestSSESession_Close(t *testing.T) {
	w := httptest.NewRecorder()
	
	session, err := NewSSESession(w, "http://localhost/messages")
	require.NoError(t, err)

	// 第一次关闭应该成功
	session.Close()
	
	// 验证通道已关闭
	select {
	case <-session.Done:
		// 成功
	case <-time.After(time.Second):
		t.Fatal("通道未关闭")
	}

	// 第二次关闭不应该 panic
	assert.NotPanics(t, func() {
		session.Close()
	})
}

// ═══════════════════════════════════════════════════════════
// SessionManager Tests
// ═══════════════════════════════════════════════════════════

func TestNewSessionManager(t *testing.T) {
	sm := NewSessionManager()
	assert.NotNil(t, sm)
	assert.NotNil(t, sm.sessions)
	assert.Equal(t, 0, sm.Count())
}

func TestSessionManager_Add(t *testing.T) {
	sm := NewSessionManager()
	w := httptest.NewRecorder()
	
	session, err := NewSSESession(w, "http://localhost/messages")
	require.NoError(t, err)

	sm.Add(session)
	assert.Equal(t, 1, sm.Count())
}

func TestSessionManager_Get(t *testing.T) {
	sm := NewSessionManager()
	w := httptest.NewRecorder()
	
	session, err := NewSSESession(w, "http://localhost/messages")
	require.NoError(t, err)

	sm.Add(session)

	// 获取存在的会话
	got, ok := sm.Get(session.ID)
	assert.True(t, ok)
	assert.Equal(t, session.ID, got.ID)

	// 获取不存在的会话
	got, ok = sm.Get("non-existent")
	assert.False(t, ok)
	assert.Nil(t, got)
}

func TestSessionManager_Remove(t *testing.T) {
	sm := NewSessionManager()
	w := httptest.NewRecorder()
	
	session, err := NewSSESession(w, "http://localhost/messages")
	require.NoError(t, err)

	sm.Add(session)
	assert.Equal(t, 1, sm.Count())

	sm.Remove(session.ID)
	assert.Equal(t, 0, sm.Count())

	// 验证会话已关闭
	select {
	case <-session.Done:
		// 成功
	case <-time.After(time.Second):
		t.Fatal("会话未关闭")
	}

	// 删除不存在的会话不应该 panic
	sm.Remove("non-existent")
}

func TestSessionManager_Count(t *testing.T) {
	sm := NewSessionManager()
	assert.Equal(t, 0, sm.Count())

	// 添加多个会话
	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		session, err := NewSSESession(w, "http://localhost/messages")
		require.NoError(t, err)
		sm.Add(session)
	}

	assert.Equal(t, 5, sm.Count())
}

func TestSessionManager_CleanExpired(t *testing.T) {
	sm := NewSessionManager()
	
	// 创建会话并手动设置创建时间
	sessions := make([]*SSESession, 3)
	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		session, err := NewSSESession(w, "http://localhost/messages")
		require.NoError(t, err)
		sessions[i] = session
		sm.Add(session)
	}

	assert.Equal(t, 3, sm.Count())

	// 修改第一个会话的创建时间，使其过期
	sessions[0].CreatedAt = time.Now().Add(-2 * time.Hour)

	// 清理超过 1 小时的会话
	cleaned := sm.CleanExpired(time.Hour)
	assert.Equal(t, 1, cleaned)
	assert.Equal(t, 2, sm.Count())

	// 验证过期会话已关闭
	select {
	case <-sessions[0].Done:
		// 成功
	case <-time.After(time.Second):
		t.Fatal("过期会话未关闭")
	}

	// 其他会话应该仍然存在
	_, ok := sm.Get(sessions[1].ID)
	assert.True(t, ok)
	_, ok = sm.Get(sessions[2].ID)
	assert.True(t, ok)
}

func TestSessionManager_ConcurrentAccess(t *testing.T) {
	sm := NewSessionManager()
	var wg sync.WaitGroup

	// 并发添加会话
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			w := httptest.NewRecorder()
			session, err := NewSSESession(w, "http://localhost/messages")
			require.NoError(t, err)
			sm.Add(session)
		}()
	}

	wg.Wait()
	assert.Equal(t, 100, sm.Count())
}

// ═══════════════════════════════════════════════════════════
// SSE Client Tests
// ═══════════════════════════════════════════════════════════

func TestNewSSEClient(t *testing.T) {
	client := NewSSEClient("http://localhost/sse", "bearer", "token123", "")
	assert.NotNil(t, client)
	assert.Equal(t, "http://localhost/sse", client.endpoint)
	assert.Equal(t, "bearer", client.authType)
	assert.Equal(t, "token123", client.authToken)
	assert.NotNil(t, client.httpClient)
}

func TestSSEClient_applyAuth_Bearer(t *testing.T) {
	client := NewSSEClient("http://localhost/sse", "bearer", "my-token", "")
	
	req, err := http.NewRequest("GET", "http://localhost", nil)
	require.NoError(t, err)

	client.applyAuth(req)
	
	authHeader := req.Header.Get("Authorization")
	assert.Equal(t, "Bearer my-token", authHeader)
}

func TestSSEClient_applyAuth_Header(t *testing.T) {
	client := NewSSEClient("http://localhost/sse", "header", "api-key-123", "X-API-Key")
	
	req, err := http.NewRequest("GET", "http://localhost", nil)
	require.NoError(t, err)

	client.applyAuth(req)
	
	assert.Equal(t, "api-key-123", req.Header.Get("X-API-Key"))
}

func TestSSEClient_applyAuth_None(t *testing.T) {
	client := NewSSEClient("http://localhost/sse", "none", "", "")
	
	req, err := http.NewRequest("GET", "http://localhost", nil)
	require.NoError(t, err)

	client.applyAuth(req)
	
	assert.Empty(t, req.Header.Get("Authorization"))
}

func TestSSEClient_Connect(t *testing.T) {
	// 创建 mock MCP 服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证请求头
		assert.Equal(t, "text/event-stream", r.Header.Get("Accept"))

		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		
		flusher, ok := w.(http.Flusher)
		require.True(t, ok)

		// 发送 endpoint 事件
		fmt.Fprintf(w, "event: endpoint\ndata: http://localhost/messages?sessionId=test-123\n\n")
		flusher.Flush()

		// 保持连接打开一段时间
		time.Sleep(100 * time.Millisecond)
	}))
	defer server.Close()

	client := NewSSEClient(server.URL, "none", "", "")
	messageURL, events, err := client.Connect()

	require.NoError(t, err)
	assert.Equal(t, "http://localhost/messages?sessionId=test-123", messageURL)
	assert.NotNil(t, events)
}

func TestSSEClient_Connect_Non200Status(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	client := NewSSEClient(server.URL, "none", "", "")
	_, _, err := client.Connect()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "非 200 状态")
}

func TestSSEClient_Connect_Timeout(t *testing.T) {
	// 使用一个立即关闭的连接来模拟超时
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		// 立即关闭，不发送 endpoint 事件
		w.(http.Flusher).Flush()
	}))
	defer server.Close()

	client := NewSSEClient(server.URL, "none", "", "")
	// 设置非常短的读取超时
	client.httpClient.Timeout = 50 * time.Millisecond
	
	_, _, err := client.Connect()

	assert.Error(t, err)
}

func TestSSEClient_SendMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证请求
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		var msg JSONRPCRequest
		err = json.Unmarshal(body, &msg)
		require.NoError(t, err)
		assert.Equal(t, "tools/list", msg.Method)

		// 返回成功响应
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      msg.ID,
			Result:  json.RawMessage(`{"tools":[]}`),
		})
	}))
	defer server.Close()

	client := NewSSEClient(server.URL, "none", "", "")
	
	req := &JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/list",
	}

	result, err := client.SendMessage(server.URL, req)
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestSSEClient_SendMessage_ErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	client := NewSSEClient(server.URL, "none", "", "")
	
	req := &JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/list",
	}

	_, err := client.SendMessage(server.URL, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

func TestSSEClient_SendMessage_NoContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	client := NewSSEClient(server.URL, "none", "", "")
	
	req := &JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/list",
	}

	result, err := client.SendMessage(server.URL, req)
	require.NoError(t, err)
	assert.Nil(t, result) // 结果通过 SSE 返回
}

func TestSSEClient_readEvents(t *testing.T) {
	eventData := []SSEEvent{
		{Event: "endpoint", Data: "http://localhost/messages"},
		{Event: "message", Data: `{"jsonrpc":"2.0","id":1,"result":{}}`},
		{Event: "message", Data: `{"jsonrpc":"2.0","id":2,"result":{"tools":[]}}`},
	}

	// 创建 SSE 数据流
	var buf bytes.Buffer
	for _, evt := range eventData {
		fmt.Fprintf(&buf, "event: %s\ndata: %s\n\n", evt.Event, evt.Data)
	}

	client := NewSSEClient("http://localhost", "none", "", "")
	eventCh := make(chan SSEEvent, 10)

	// 使用 pipe 模拟 body
	pr, pw := io.Pipe()
	go func() {
		buf.WriteTo(pw)
		pw.Close()
	}()

	go client.readEvents(pr, eventCh)

	// 收集事件
	var received []SSEEvent
	done := make(chan struct{})
	go func() {
		for evt := range eventCh {
			received = append(received, evt)
		}
		close(done)
	}()

	select {
	case <-done:
		assert.Len(t, received, 3)
		assert.Equal(t, "endpoint", received[0].Event)
		assert.Equal(t, "http://localhost/messages", received[0].Data)
		assert.Equal(t, "message", received[1].Event)
	case <-time.After(2 * time.Second):
		t.Fatal("超时等待事件")
	}
}

// ═══════════════════════════════════════════════════════════
// Proxy Tests
// ═══════════════════════════════════════════════════════════

func TestNewProxy(t *testing.T) {
	logger := zap.NewNop()
	proxy := NewProxy(logger)
	assert.NotNil(t, proxy)
	assert.NotNil(t, proxy.connections)
	assert.Equal(t, logger, proxy.logger)
}

func TestProxy_ConnectService(t *testing.T) {
	var serverURL string
	// 创建 mock MCP 服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			// SSE 连接
			w.Header().Set("Content-Type", "text/event-stream")
			w.WriteHeader(http.StatusOK)
			flusher, _ := w.(http.Flusher)
			fmt.Fprintf(w, "event: endpoint\ndata: %s/messages\n\n", serverURL)
			flusher.Flush()
			time.Sleep(100 * time.Millisecond)
		case "POST":
			// 消息处理
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      1,
				Result:  json.RawMessage(`{"protocolVersion":"2024-11-05"}`),
			})
		}
	}))
	serverURL = server.URL
	defer server.Close()

	logger := zap.NewNop()
	proxy := NewProxy(logger)

	err := proxy.ConnectService("test-service", server.URL, "none", "", "")
	require.NoError(t, err)

	assert.True(t, proxy.IsConnected("test-service"))
}

func TestProxy_ConnectService_AlreadyConnected(t *testing.T) {
	var serverURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		flusher, _ := w.(http.Flusher)
		fmt.Fprintf(w, "event: endpoint\ndata: %s/messages\n\n", serverURL)
		flusher.Flush()
		time.Sleep(100 * time.Millisecond)
	}))
	serverURL = server.URL
	defer server.Close()

	logger := zap.NewNop()
	proxy := NewProxy(logger)

	// 第一次连接
	err := proxy.ConnectService("test-service", server.URL, "none", "", "")
	require.NoError(t, err)

	// 第二次连接应该直接返回
	err = proxy.ConnectService("test-service", server.URL, "none", "", "")
	require.NoError(t, err) // 不报错，直接返回
}

func TestProxy_ConnectService_Failure(t *testing.T) {
	logger := zap.NewNop()
	proxy := NewProxy(logger)

	// 连接一个不存在的地址
	err := proxy.ConnectService("test-service", "http://localhost:59999", "none", "", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "失败")
}

func TestProxy_ForwardRequest(t *testing.T) {
	var serverURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			w.Header().Set("Content-Type", "text/event-stream")
			w.WriteHeader(http.StatusOK)
			flusher, _ := w.(http.Flusher)
			fmt.Fprintf(w, "event: endpoint\ndata: %s/messages\n\n", serverURL)
			flusher.Flush()
			time.Sleep(100 * time.Millisecond)
		case "POST":
			var req JSONRPCRequest
			json.NewDecoder(r.Body).Decode(&req)
			
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result:  json.RawMessage(`{"tools":[{"name":"test-tool"}]}`),
			})
		}
	}))
	serverURL = server.URL
	defer server.Close()

	logger := zap.NewNop()
	proxy := NewProxy(logger)

	err := proxy.ConnectService("test-service", server.URL, "none", "", "")
	require.NoError(t, err)

	req := &JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      123,
		Method:  "tools/list",
	}

	resp, err := proxy.ForwardRequest("test-service", req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, float64(123), resp.ID) // JSON 解析后为 float64
}

func TestProxy_ForwardRequest_NotConnected(t *testing.T) {
	logger := zap.NewNop()
	proxy := NewProxy(logger)

	req := &JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/list",
	}

	_, err := proxy.ForwardRequest("non-existent", req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "未连接")
}

func TestProxy_DisconnectService(t *testing.T) {
	var serverURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		flusher, _ := w.(http.Flusher)
		fmt.Fprintf(w, "event: endpoint\ndata: %s/messages\n\n", serverURL)
		flusher.Flush()
		time.Sleep(100 * time.Millisecond)
	}))
	serverURL = server.URL
	defer server.Close()

	logger := zap.NewNop()
	proxy := NewProxy(logger)

	err := proxy.ConnectService("test-service", server.URL, "none", "", "")
	require.NoError(t, err)
	assert.True(t, proxy.IsConnected("test-service"))

	proxy.DisconnectService("test-service")
	assert.False(t, proxy.IsConnected("test-service"))
}

func TestProxy_ListConnectedServices(t *testing.T) {
	var serverURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		flusher, _ := w.(http.Flusher)
		fmt.Fprintf(w, "event: endpoint\ndata: %s/messages\n\n", serverURL)
		flusher.Flush()
		time.Sleep(50 * time.Millisecond)
	}))
	serverURL = server.URL
	defer server.Close()

	logger := zap.NewNop()
	proxy := NewProxy(logger)

	services := []string{"service-1", "service-2", "service-3"}
	for _, svc := range services {
		err := proxy.ConnectService(svc, server.URL, "none", "", "")
		require.NoError(t, err)
	}

	connected := proxy.ListConnectedServices()
	assert.Len(t, connected, 3)
	for _, svc := range services {
		assert.Contains(t, connected, svc)
	}
}

// ═══════════════════════════════════════════════════════════
// Proxy Handler Tests
// ═══════════════════════════════════════════════════════════

func TestProxy_HandleRequest_Initialize(t *testing.T) {
	logger := zap.NewNop()
	proxy := NewProxy(logger)

	req := &JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  MethodInitialize,
	}

	resp, err := proxy.HandleRequest(req, nil)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Nil(t, resp.Error)

	var result InitializeResult
	err = json.Unmarshal(resp.Result, &result)
	require.NoError(t, err)
	assert.Equal(t, MCPProtocolVersion, result.ProtocolVersion)
	assert.NotNil(t, result.Capabilities.Tools)
	assert.NotNil(t, result.Capabilities.Resources)
}

func TestProxy_HandleRequest_Ping(t *testing.T) {
	logger := zap.NewNop()
	proxy := NewProxy(logger)

	req := &JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  MethodPing,
	}

	resp, err := proxy.HandleRequest(req, nil)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Nil(t, resp.Error)
	assert.Equal(t, "{}", string(resp.Result))
}

func TestProxy_HandleRequest_MethodNotFound(t *testing.T) {
	logger := zap.NewNop()
	proxy := NewProxy(logger)

	req := &JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "unknown/method",
	}

	resp, err := proxy.HandleRequest(req, nil)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Error)
	assert.Equal(t, ErrCodeMethodNotFound, resp.Error.Code)
}

func TestProxy_handleToolsList(t *testing.T) {
	logger := zap.NewNop()
	proxy := NewProxy(logger)

	services := []ServiceInfo{
		{
			Name:        "service-1",
			ToolsSchema: json.RawMessage(`[{"name":"tool-1","description":"Tool 1"}]`),
		},
		{
			Name:        "service-2",
			ToolsSchema: json.RawMessage(`[{"name":"tool-2","description":"Tool 2"}]`),
		},
	}

	req := &JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  MethodToolsList,
	}

	resp, err := proxy.HandleRequest(req, services)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Nil(t, resp.Error)

	var result ToolsListResult
	err = json.Unmarshal(resp.Result, &result)
	require.NoError(t, err)
	assert.Len(t, result.Tools, 2)
}

func TestProxy_handleToolsList_InvalidSchema(t *testing.T) {
	logger := zap.NewNop()
	proxy := NewProxy(logger)

	services := []ServiceInfo{
		{
			Name:        "service-1",
			ToolsSchema: json.RawMessage(`invalid json`), // 无效的 JSON
		},
	}

	req := &JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  MethodToolsList,
	}

	resp, err := proxy.HandleRequest(req, services)
	require.NoError(t, err)
	assert.NotNil(t, resp)

	var result ToolsListResult
	err = json.Unmarshal(resp.Result, &result)
	require.NoError(t, err)
	assert.Len(t, result.Tools, 0) // 无效的 schema 被跳过
}

func TestProxy_handleToolsCall(t *testing.T) {
	var serverURL string
	// 创建 mock 服务
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			w.Header().Set("Content-Type", "text/event-stream")
			w.WriteHeader(http.StatusOK)
			flusher, _ := w.(http.Flusher)
			fmt.Fprintf(w, "event: endpoint\ndata: %s/messages\n\n", serverURL)
			flusher.Flush()
			time.Sleep(100 * time.Millisecond)
		case "POST":
			var req JSONRPCRequest
			json.NewDecoder(r.Body).Decode(&req)
			
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result:  json.RawMessage(`{"content":[{"type":"text","text":"Tool executed"}]}`),
			})
		}
	}))
	serverURL = server.URL
	defer server.Close()

	logger := zap.NewNop()
	proxy := NewProxy(logger)

	// 先连接服务
	err := proxy.ConnectService("test-service", server.URL, "none", "", "")
	require.NoError(t, err)

	services := []ServiceInfo{
		{
			Name:        "test-service",
			ToolsSchema: json.RawMessage(`[{"name":"test-tool","description":"Test tool"}]`),
		},
	}

	req := &JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  MethodToolsCall,
		Params:  json.RawMessage(`{"name":"test-tool","arguments":{}}`),
	}

	resp, err := proxy.HandleRequest(req, services)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Nil(t, resp.Error)

	var result ToolCallResult
	err = json.Unmarshal(resp.Result, &result)
	require.NoError(t, err)
	assert.Len(t, result.Content, 1)
	assert.Equal(t, "Tool executed", result.Content[0].Text)
}

func TestProxy_handleToolsCall_InvalidParams(t *testing.T) {
	logger := zap.NewNop()
	proxy := NewProxy(logger)

	req := &JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  MethodToolsCall,
		Params:  json.RawMessage(`invalid json`),
	}

	resp, err := proxy.HandleRequest(req, nil)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Error)
	assert.Equal(t, ErrCodeInvalidParams, resp.Error.Code)
}

func TestProxy_handleToolsCall_ToolNotFound(t *testing.T) {
	logger := zap.NewNop()
	proxy := NewProxy(logger)

	services := []ServiceInfo{
		{
			Name:        "test-service",
			ToolsSchema: json.RawMessage(`[{"name":"other-tool","description":"Other tool"}]`),
		},
	}

	req := &JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  MethodToolsCall,
		Params:  json.RawMessage(`{"name":"non-existent-tool","arguments":{}}`),
	}

	resp, err := proxy.HandleRequest(req, services)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Error)
	assert.Equal(t, ErrCodeMethodNotFound, resp.Error.Code)
	assert.Contains(t, resp.Error.Message, "工具不存在")
}

func TestProxy_handleResourcesList(t *testing.T) {
	var serverURL string
	// 创建 mock 服务
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			w.Header().Set("Content-Type", "text/event-stream")
			w.WriteHeader(http.StatusOK)
			flusher, _ := w.(http.Flusher)
			fmt.Fprintf(w, "event: endpoint\ndata: %s/messages\n\n", serverURL)
			flusher.Flush()
			time.Sleep(100 * time.Millisecond)
		case "POST":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      1,
				Result:  json.RawMessage(`{"resources":[{"uri":"file:///test.txt","name":"test.txt"}]}`),
			})
		}
	}))
	serverURL = server.URL
	defer server.Close()

	logger := zap.NewNop()
	proxy := NewProxy(logger)

	err := proxy.ConnectService("test-service", server.URL, "none", "", "")
	require.NoError(t, err)

	services := []ServiceInfo{
		{
			Name: "test-service",
		},
	}

	req := &JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  MethodResourcesList,
	}

	resp, err := proxy.HandleRequest(req, services)
	require.NoError(t, err)
	assert.NotNil(t, resp)

	var result ResourcesListResult
	err = json.Unmarshal(resp.Result, &result)
	require.NoError(t, err)
	assert.Len(t, result.Resources, 1)
}

func TestProxy_handleResourcesRead(t *testing.T) {
	var serverURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			w.Header().Set("Content-Type", "text/event-stream")
			w.WriteHeader(http.StatusOK)
			flusher, _ := w.(http.Flusher)
			fmt.Fprintf(w, "event: endpoint\ndata: %s/messages\n\n", serverURL)
			flusher.Flush()
			time.Sleep(100 * time.Millisecond)
		case "POST":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      1,
				Result:  json.RawMessage(`{"contents":[{"uri":"file:///test.txt","text":"Hello"}]}`),
			})
		}
	}))
	serverURL = server.URL
	defer server.Close()

	logger := zap.NewNop()
	proxy := NewProxy(logger)

	err := proxy.ConnectService("test-service", server.URL, "none", "", "")
	require.NoError(t, err)

	services := []ServiceInfo{
		{
			Name: "test-service",
		},
	}

	req := &JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  MethodResourcesRead,
	}

	resp, err := proxy.HandleRequest(req, services)
	require.NoError(t, err)
	assert.NotNil(t, resp)

	var result ResourceReadResult
	err = json.Unmarshal(resp.Result, &result)
	require.NoError(t, err)
	assert.Len(t, result.Contents, 1)
}

func TestProxy_handleResourcesRead_NoService(t *testing.T) {
	logger := zap.NewNop()
	proxy := NewProxy(logger)

	req := &JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  MethodResourcesRead,
	}

	resp, err := proxy.HandleRequest(req, nil)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Error)
	assert.Equal(t, ErrCodeInternalError, resp.Error.Code)
}

// ═══════════════════════════════════════════════════════════
// Proxy waitForResponse Tests
// ═══════════════════════════════════════════════════════════

func TestProxy_waitForResponse(t *testing.T) {
	logger := zap.NewNop()
	proxy := NewProxy(logger)

	eventCh := make(chan SSEEvent, 10)
	conn := &ServiceConnection{
		ServiceName: "test",
		Events:      eventCh,
	}

	// 在后台发送匹配的响应
	go func() {
		time.Sleep(50 * time.Millisecond)
		eventCh <- SSEEvent{
			Event: "message",
			Data:  `{"jsonrpc":"2.0","id":123,"result":{"tools":[]}}`,
		}
	}()

	resp, err := proxy.waitForResponse(conn, 123)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, float64(123), resp.ID)
}

func TestProxy_waitForResponse_NotMatching(t *testing.T) {
	logger := zap.NewNop()
	proxy := NewProxy(logger)

	eventCh := make(chan SSEEvent, 10)
	conn := &ServiceConnection{
		ServiceName: "test",
		Events:      eventCh,
	}

	// 发送不匹配的响应
	go func() {
		time.Sleep(50 * time.Millisecond)
		eventCh <- SSEEvent{
			Event: "message",
			Data:  `{"jsonrpc":"2.0","id":999,"result":{}}`, // 不同的 ID
		}
		// 关闭通道，模拟超时
		close(eventCh)
	}()

	_, err := proxy.waitForResponse(conn, 123)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "超时")
}

// ═══════════════════════════════════════════════════════════
// Mock and Integration Tests
// ═══════════════════════════════════════════════════════════

// MockSSEClient 用于测试的 SSE 客户端 mock
type MockSSEClient struct {
	ConnectFunc     func() (string, <-chan SSEEvent, error)
	SendMessageFunc func(string, interface{}) (*json.RawMessage, error)
}

func (m *MockSSEClient) Connect() (string, <-chan SSEEvent, error) {
	if m.ConnectFunc != nil {
		return m.ConnectFunc()
	}
	return "", nil, nil
}

func (m *MockSSEClient) SendMessage(url string, msg interface{}) (*json.RawMessage, error) {
	if m.SendMessageFunc != nil {
		return m.SendMessageFunc(url, msg)
	}
	return nil, nil
}

func TestProxy_ConcurrentRequests(t *testing.T) {
	var serverURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			w.Header().Set("Content-Type", "text/event-stream")
			w.WriteHeader(http.StatusOK)
			flusher, _ := w.(http.Flusher)
			fmt.Fprintf(w, "event: endpoint\ndata: %s/messages\n\n", serverURL)
			flusher.Flush()
			time.Sleep(500 * time.Millisecond)
		case "POST":
			var req JSONRPCRequest
			json.NewDecoder(r.Body).Decode(&req)
			
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			
			// 模拟处理延迟
			time.Sleep(10 * time.Millisecond)
			
			json.NewEncoder(w).Encode(JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result:  json.RawMessage(`{"result":"ok"}`),
			})
		}
	}))
	serverURL = server.URL
	defer server.Close()

	logger := zap.NewNop()
	proxy := NewProxy(logger)

	err := proxy.ConnectService("test-service", server.URL, "none", "", "")
	require.NoError(t, err)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			req := &JSONRPCRequest{
				JSONRPC: "2.0",
				ID:      id,
				Method:  "test/method",
			}
			resp, err := proxy.ForwardRequest("test-service", req)
			assert.NoError(t, err)
			assert.NotNil(t, resp)
		}(i)
	}

	wg.Wait()
}

// ═══════════════════════════════════════════════════════════
// SSE Server Handler Tests
// ═══════════════════════════════════════════════════════════

func TestSSEEndpointHandler(t *testing.T) {
	// 测试 SSE 端点处理器
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证支持 Flusher
		flusher, ok := w.(http.Flusher)
		require.True(t, ok)

		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		// 发送 endpoint 事件
		fmt.Fprintf(w, "event: endpoint\ndata: http://localhost/messages?sessionId=%s\n\n", uuid.New().String())
		flusher.Flush()

		// 保持连接
		<-r.Context().Done()
	})

	req := httptest.NewRequest("GET", "/sse", nil)
	req.Header.Set("Accept", "text/event-stream")
	w := httptest.NewRecorder()

	go handler(w, req)

	// 读取响应
	time.Sleep(100 * time.Millisecond)
	body := w.Body.String()
	assert.Contains(t, body, "event: endpoint")
	assert.Contains(t, body, "http://localhost/messages")
}

// ═══════════════════════════════════════════════════════════
// Utility Tests
// ═══════════════════════════════════════════════════════════

func TestUUIDGeneration(t *testing.T) {
	// 测试 UUID 生成
	id1 := uuid.New().String()
	id2 := uuid.New().String()
	
	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2)
	
	// 验证 UUID 格式
	_, err := uuid.Parse(id1)
	assert.NoError(t, err)
}

func TestJSONRawMessageHandling(t *testing.T) {
	// 测试 json.RawMessage 的处理
	original := map[string]interface{}{
		"key1": "value1",
		"key2": 123,
		"key3": true,
		"nested": map[string]string{
			"inner": "value",
		},
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	raw := json.RawMessage(data)

	// 验证可以直接序列化
	data2, err := json.Marshal(raw)
	require.NoError(t, err)
	assert.Equal(t, data, data2)

	// 验证可以反序列化
	var decoded map[string]interface{}
	err = json.Unmarshal(raw, &decoded)
	require.NoError(t, err)
	assert.Equal(t, original["key1"], decoded["key1"])
}

// ═══════════════════════════════════════════════════════════
// Benchmark Tests
// ═══════════════════════════════════════════════════════════

func BenchmarkJSONRPCRequest_Marshal(b *testing.B) {
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/list",
		Params:  json.RawMessage(`{}`),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJSONRPCResponse_Marshal(b *testing.B) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      1,
		Result:  json.RawMessage(`{"tools":[{"name":"tool1"},{"name":"tool2"}]}`),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(resp)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSessionManager_Concurrent(b *testing.B) {
	sm := NewSessionManager()
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			w := httptest.NewRecorder()
			session, _ := NewSSESession(w, "http://localhost/messages")
			sm.Add(session)
			if i%2 == 0 {
				sm.Get(session.ID)
			}
			i++
		}
	})
}

// ═══════════════════════════════════════════════════════════
// Error Response Tests
// ═══════════════════════════════════════════════════════════

func TestErrorResponseFormatting(t *testing.T) {
	tests := []struct {
		name       string
		code       int
		message    string
		wantCode   int
		wantMsg    string
	}{
		{
			name:     "解析错误",
			code:     ErrCodeParseError,
			message:  "Invalid JSON",
			wantCode: -32700,
			wantMsg:  "Invalid JSON",
		},
		{
			name:     "无效请求",
			code:     ErrCodeInvalidRequest,
			message:  "Missing jsonrpc field",
			wantCode: -32600,
			wantMsg:  "Missing jsonrpc field",
		},
		{
			name:     "方法未找到",
			code:     ErrCodeMethodNotFound,
			message:  "Method not found: test",
			wantCode: -32601,
			wantMsg:  "Method not found: test",
		},
		{
			name:     "无效参数",
			code:     ErrCodeInvalidParams,
			message:  "Missing required param",
			wantCode: -32602,
			wantMsg:  "Missing required param",
		},
		{
			name:     "内部错误",
			code:     ErrCodeInternalError,
			message:  "Database connection failed",
			wantCode: -32603,
			wantMsg:  "Database connection failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := NewErrorResponse("test-id", tt.code, tt.message)
			
			assert.Equal(t, JSONRPCVersion, resp.JSONRPC)
			assert.Equal(t, "test-id", resp.ID)
			require.NotNil(t, resp.Error)
			assert.Equal(t, tt.wantCode, resp.Error.Code)
			assert.Equal(t, tt.wantMsg, resp.Error.Message)
			assert.Nil(t, resp.Result)

			// 验证 JSON 序列化
			data, err := json.Marshal(resp)
			require.NoError(t, err)
			
			var decoded JSONRPCResponse
			err = json.Unmarshal(data, &decoded)
			require.NoError(t, err)
			assert.Equal(t, tt.wantCode, decoded.Error.Code)
		})
	}
}

// ═══════════════════════════════════════════════════════════
// SSE Event Parsing Tests
// ═══════════════════════════════════════════════════════════

func TestSSEEventParsing(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		events   []SSEEvent
	}{
		{
			name:  "单个事件",
			input: "event: message\ndata: hello\n\n",
			events: []SSEEvent{
				{Event: "message", Data: "hello"},
			},
		},
		{
			name:  "多个事件",
			input: "event: endpoint\ndata: url1\n\nevent: message\ndata: data1\n\n",
			events: []SSEEvent{
				{Event: "endpoint", Data: "url1"},
				{Event: "message", Data: "data1"},
			},
		},
		{
			name:   "空事件",
			input:  "\n",
			events: []SSEEvent{},
		},
		{
			name:  "带JSON数据的事件",
			input: "event: message\ndata: {\"jsonrpc\":\"2.0\",\"id\":1}\n\n",
			events: []SSEEvent{
				{Event: "message", Data: `{"jsonrpc":"2.0","id":1}`},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 使用 bufio.Scanner 模拟 SSE 解析
			scanner := bufio.NewScanner(strings.NewReader(tt.input))
			var events []SSEEvent
			var currentEvent, currentData string

			for scanner.Scan() {
				line := scanner.Text()
				if line == "" {
					if currentData != "" {
						events = append(events, SSEEvent{
							Event: currentEvent,
							Data:  currentData,
						})
						currentEvent = ""
						currentData = ""
					}
					continue
				}
				if strings.HasPrefix(line, "event: ") {
					currentEvent = strings.TrimPrefix(line, "event: ")
				} else if strings.HasPrefix(line, "data: ") {
					currentData = strings.TrimPrefix(line, "data: ")
				}
			}

			assert.Equal(t, len(tt.events), len(events))
			for i, evt := range tt.events {
				if i < len(events) {
					assert.Equal(t, evt.Event, events[i].Event)
					assert.Equal(t, evt.Data, events[i].Data)
				}
			}
		})
	}
}

// ═══════════════════════════════════════════════════════════
// Helper Functions for Tests
// ═══════════════════════════════════════════════════════════

func mustMarshalJSON(v interface{}) json.RawMessage {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return json.RawMessage(data)
}



// ═══════════════════════════════════════════════════════════
// Integration Tests
// ═══════════════════════════════════════════════════════════

func TestIntegration_FullRequestFlow(t *testing.T) {
	var serverURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			// SSE 连接
			w.Header().Set("Content-Type", "text/event-stream")
			w.WriteHeader(http.StatusOK)
			flusher, ok := w.(http.Flusher)
			if !ok {
				http.Error(w, "Streaming not supported", http.StatusInternalServerError)
				return
			}

			messageURL := fmt.Sprintf("%s/messages?sessionId=%s", serverURL, uuid.New().String())
			fmt.Fprintf(w, "event: endpoint\ndata: %s\n\n", messageURL)
			flusher.Flush()
			
			// 保持连接打开一段时间
			time.Sleep(200 * time.Millisecond)

		case "POST":
			// 处理消息
			body, _ := io.ReadAll(r.Body)
			
			var req JSONRPCRequest
			if err := json.Unmarshal(body, &req); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			
			var result interface{}
			switch req.Method {
			case "initialize":
				result = InitializeResult{
					ProtocolVersion: MCPProtocolVersion,
					Capabilities:    ServerCapability{},
					ServerInfo:      Implementation{Name: "mock-server", Version: "1.0"},
				}
			case "tools/list":
				result = ToolsListResult{Tools: []Tool{{Name: "mock-tool"}}}
			default:
				result = struct{}{}
			}

			json.NewEncoder(w).Encode(JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result:  mustMarshalJSON(result),
			})
		}
	}))
	serverURL = server.URL
	defer server.Close()

	logger := zap.NewNop()
	proxy := NewProxy(logger)

	// 连接服务
	err := proxy.ConnectService("mock-service", server.URL, "none", "", "")
	require.NoError(t, err)

	// 发送初始化请求
	initReq := &JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  MethodInitialize,
	}

	resp, err := proxy.ForwardRequest("mock-service", initReq)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Nil(t, resp.Error)

	// 发送工具列表请求
	toolsReq := &JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      2,
		Method:  MethodToolsList,
	}

	resp, err = proxy.ForwardRequest("mock-service", toolsReq)
	require.NoError(t, err)
	assert.NotNil(t, resp)

	var result ToolsListResult
	err = json.Unmarshal(resp.Result, &result)
	require.NoError(t, err)
	assert.Len(t, result.Tools, 1)
}
