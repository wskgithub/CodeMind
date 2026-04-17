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

// SSESession represents a single SSE connection session.
type SSESession struct {
	ID         string
	Writer     http.ResponseWriter
	Flusher    http.Flusher
	Done       chan struct{}
	MessageURL string
	CreatedAt  time.Time
}

// NewSSESession creates a new SSE session.
func NewSSESession(w http.ResponseWriter, messageBaseURL string) (*SSESession, error) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, fmt.Errorf("ResponseWriter does not support Flusher interface")
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

// SendEvent sends an SSE event.
func (s *SSESession) SendEvent(event, data string) error {
	_, err := fmt.Fprintf(s.Writer, "event: %s\ndata: %s\n\n", event, data)
	if err != nil {
		return err
	}
	s.Flusher.Flush()
	return nil
}

// SendEndpoint sends the message endpoint URL as the first SSE event.
func (s *SSESession) SendEndpoint() error {
	return s.SendEvent("endpoint", s.MessageURL)
}

// SendMessage sends a JSON-RPC message.
func (s *SSESession) SendMessage(msg interface{}) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}
	return s.SendEvent("message", string(data))
}

// Close closes the session.
func (s *SSESession) Close() {
	select {
	case <-s.Done:
	default:
		close(s.Done)
	}
}

// SessionManager manages SSE sessions.
type SessionManager struct {
	mu       sync.RWMutex
	sessions map[string]*SSESession
}

// NewSessionManager creates a new session manager.
func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*SSESession),
	}
}

// Add registers a session.
func (m *SessionManager) Add(session *SSESession) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[session.ID] = session
}

// Get retrieves a session by ID.
func (m *SessionManager) Get(sessionID string) (*SSESession, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.sessions[sessionID]
	return s, ok
}

// Remove removes and closes a session.
func (m *SessionManager) Remove(sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if s, ok := m.sessions[sessionID]; ok {
		s.Close()
		delete(m.sessions, sessionID)
	}
}

// Count returns the number of active sessions.
func (m *SessionManager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.sessions)
}

// CleanExpired removes sessions older than maxAge.
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

// SSEClient connects to upstream MCP services via SSE.
type SSEClient struct {
	endpoint   string
	httpClient *http.Client
	authType   string
	authToken  string
	authHeader string
}

// NewSSEClient creates an SSE client for upstream MCP services.
func NewSSEClient(endpoint string, authType, authToken, authHeader string) *SSEClient {
	return &SSEClient{
		endpoint: endpoint,
		httpClient: &http.Client{
			Timeout: 0,
		},
		authType:   authType,
		authToken:  authToken,
		authHeader: authHeader,
	}
}

// Connect establishes an SSE connection and returns the message URL and event channel.
func (c *SSEClient) Connect() (messageURL string, events <-chan SSEEvent, err error) {
	req, err := http.NewRequest("GET", c.endpoint, nil)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create SSE request: %w", err)
	}

	req.Header.Set("Accept", "text/event-stream")
	c.applyAuth(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("failed to connect to upstream MCP service: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return "", nil, fmt.Errorf("upstream MCP service returned non-200 status: %d", resp.StatusCode)
	}

	eventCh := make(chan SSEEvent, 64)

	go c.readEvents(resp.Body, eventCh)

	select {
	case evt := <-eventCh:
		if evt.Event == "endpoint" {
			messageURL = evt.Data
		} else {
			return "", nil, fmt.Errorf("first SSE event is not endpoint: %s", evt.Event)
		}
	case <-time.After(10 * time.Second):
		resp.Body.Close()
		return "", nil, fmt.Errorf("timeout waiting for upstream endpoint event")
	}

	return messageURL, eventCh, nil
}

// SendMessage sends a JSON-RPC message to the upstream MCP service.
func (c *SSEClient) SendMessage(messageURL string, msg interface{}) (*json.RawMessage, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	req, err := http.NewRequest("POST", messageURL, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	c.applyAuth(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var result json.RawMessage
		if err := json.NewDecoder(resp.Body).Decode(&result); err == nil {
			return &result, nil
		}
	}

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("upstream MCP service error (HTTP %d): %s", resp.StatusCode, string(body))
	}

	return nil, nil
}

// applyAuth applies authentication to the request.
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

func (c *SSEClient) readEvents(body io.ReadCloser, ch chan<- SSEEvent) {
	defer close(ch)
	defer body.Close()

	scanner := bufio.NewScanner(body)
	var currentEvent, currentData string

	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
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

// SSEEvent represents an SSE event.
type SSEEvent struct {
	Event string
	Data  string
}
