package mcp

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ──────────────────────────────────
// SSE 传输层
// 实现 MCP 的 SSE 传输协议
// ──────────────────────────────────

// SSESession 单个 SSE 连接会话
type SSESession struct {
	ID         string
	Writer     http.ResponseWriter
	Flusher    http.Flusher
	Done       chan struct{}
	MessageURL string   // 对应的消息端点 URL
	CreatedAt  time.Time
}

// NewSSESession 创建新的 SSE 会话
func NewSSESession(w http.ResponseWriter, messageBaseURL string) (*SSESession, error) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, fmt.Errorf("ResponseWriter 不支持 Flusher 接口")
	}

	sessionID := uuid.New().String()

	return &SSESession{
		ID:         sessionID,
		Writer:     w,
		Flusher:    flusher,
		Done:       make(chan struct{}),
		MessageURL: fmt.Sprintf("%s?sessionId=%s", messageBaseURL, sessionID),
		CreatedAt:  time.Now(),
	}, nil
}

// SendEvent 发送 SSE 事件
func (s *SSESession) SendEvent(event, data string) error {
	_, err := fmt.Fprintf(s.Writer, "event: %s\ndata: %s\n\n", event, data)
	if err != nil {
		return err
	}
	s.Flusher.Flush()
	return nil
}

// SendEndpoint 发送消息端点 URL（SSE 连接建立后的首个事件）
func (s *SSESession) SendEndpoint() error {
	return s.SendEvent("endpoint", s.MessageURL)
}

// SendMessage 发送 JSON-RPC 消息
func (s *SSESession) SendMessage(msg interface{}) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %w", err)
	}
	return s.SendEvent("message", string(data))
}

// Close 关闭会话
func (s *SSESession) Close() {
	select {
	case <-s.Done:
		// 已关闭
	default:
		close(s.Done)
	}
}

// ──────────────────────────────────
// SSE 会话管理器
// ──────────────────────────────────

// SessionManager SSE 会话管理
type SessionManager struct {
	mu       sync.RWMutex
	sessions map[string]*SSESession
}

// NewSessionManager 创建会话管理器
func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*SSESession),
	}
}

// Add 注册会话
func (m *SessionManager) Add(session *SSESession) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[session.ID] = session
}

// Get 获取会话
func (m *SessionManager) Get(sessionID string) (*SSESession, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.sessions[sessionID]
	return s, ok
}

// Remove 移除会话
func (m *SessionManager) Remove(sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if s, ok := m.sessions[sessionID]; ok {
		s.Close()
		delete(m.sessions, sessionID)
	}
}

// Count 当前活跃会话数
func (m *SessionManager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.sessions)
}

// CleanExpired 清理过期会话
func (m *SessionManager) CleanExpired(maxAge time.Duration) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	cleaned := 0
	for id, s := range m.sessions {
		if now.Sub(s.CreatedAt) > maxAge {
			s.Close()
			delete(m.sessions, id)
			cleaned++
		}
	}
	return cleaned
}

// ──────────────────────────────────
// SSE 客户端（连接上游 MCP 服务）
// ──────────────────────────────────

// SSEClient 连接上游 MCP 服务的 SSE 客户端
type SSEClient struct {
	endpoint   string       // 上游 SSE 端点 URL
	httpClient *http.Client
	authType   string       // "none" | "bearer" | "header"
	authToken  string       // Bearer token 或自定义 header 值
	authHeader string       // 自定义 header 名
}

// NewSSEClient 创建上游 MCP SSE 客户端
func NewSSEClient(endpoint string, authType, authToken, authHeader string) *SSEClient {
	return &SSEClient{
		endpoint: endpoint,
		httpClient: &http.Client{
			Timeout: 0, // SSE 连接不设超时
		},
		authType:   authType,
		authToken:  authToken,
		authHeader: authHeader,
	}
}

// Connect 建立 SSE 连接，返回消息端点 URL 和事件流
func (c *SSEClient) Connect() (messageURL string, events <-chan SSEEvent, err error) {
	req, err := http.NewRequest("GET", c.endpoint, nil)
	if err != nil {
		return "", nil, fmt.Errorf("创建 SSE 请求失败: %w", err)
	}

	req.Header.Set("Accept", "text/event-stream")
	c.applyAuth(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("连接上游 MCP 服务失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return "", nil, fmt.Errorf("上游 MCP 服务返回非 200 状态: %d", resp.StatusCode)
	}

	eventCh := make(chan SSEEvent, 64)

	// 后台读取 SSE 事件
	go c.readEvents(resp.Body, eventCh)

	// 等待获取消息端点 URL（通常是第一个 endpoint 事件）
	select {
	case evt := <-eventCh:
		if evt.Event == "endpoint" {
			messageURL = evt.Data
		} else {
			return "", nil, fmt.Errorf("首个 SSE 事件不是 endpoint: %s", evt.Event)
		}
	case <-time.After(10 * time.Second):
		resp.Body.Close()
		return "", nil, fmt.Errorf("等待上游 endpoint 事件超时")
	}

	return messageURL, eventCh, nil
}

// SendMessage 向上游 MCP 服务发送 JSON-RPC 消息
func (c *SSEClient) SendMessage(messageURL string, msg interface{}) (*json.RawMessage, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("序列化消息失败: %w", err)
	}

	req, err := http.NewRequest("POST", messageURL, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	c.applyAuth(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送消息失败: %w", err)
	}
	defer resp.Body.Close()

	// 某些 MCP 服务直接在 POST 响应中返回结果
	if resp.StatusCode == http.StatusOK {
		var result json.RawMessage
		if err := json.NewDecoder(resp.Body).Decode(&result); err == nil {
			return &result, nil
		}
	}

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("上游 MCP 服务错误 (HTTP %d): %s", resp.StatusCode, string(body))
	}

	return nil, nil // 结果通过 SSE 事件返回
}

// applyAuth 应用认证信息到请求
func (c *SSEClient) applyAuth(req *http.Request) {
	switch c.authType {
	case "bearer":
		if c.authToken != "" {
			req.Header.Set("Authorization", "Bearer "+c.authToken)
		}
	case "header":
		if c.authHeader != "" && c.authToken != "" {
			req.Header.Set(c.authHeader, c.authToken)
		}
	}
}

// readEvents 从 SSE 流中读取事件
func (c *SSEClient) readEvents(body io.ReadCloser, ch chan<- SSEEvent) {
	defer close(ch)
	defer body.Close()

	scanner := bufio.NewScanner(body)
	var currentEvent, currentData string

	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			// 空行表示事件结束
			if currentData != "" {
				ch <- SSEEvent{
					Event: currentEvent,
					Data:  currentData,
				}
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
}

// SSEEvent SSE 事件
type SSEEvent struct {
	Event string
	Data  string
}
