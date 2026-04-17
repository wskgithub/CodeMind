package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"github.com/google/uuid"
)

// ConversationExtractor extracts conversation IDs from requests.
type ConversationExtractor struct{}

// NewConversationExtractor creates a new extractor.
func NewConversationExtractor() *ConversationExtractor {
	return &ConversationExtractor{}
}

// ExtractConversationID extracts or generates a conversation ID from request.
func (e *ConversationExtractor) ExtractConversationID(
	requestBody json.RawMessage,
	headers map[string]string,
) string {
	if id := headers["X-Conversation-Id"]; id != "" {
		return e.normalizeID(id)
	}
	if id := headers["X-Session-Id"]; id != "" {
		return e.normalizeID(id)
	}

	var body struct {
		ConversationID string `json:"conversation_id"`
		SessionID      string `json:"session_id"`
	}
	if json.Unmarshal(requestBody, &body) == nil {
		if body.ConversationID != "" {
			return e.normalizeID(body.ConversationID)
		}
		if body.SessionID != "" {
			return e.normalizeID(body.SessionID)
		}
	}

	if firstUserMsg := e.extractFirstUserMessage(requestBody); firstUserMsg != "" {
		hash := sha256.Sum256([]byte(firstUserMsg))
		return hex.EncodeToString(hash[:8])
	}

	return uuid.New().String()[:16]
}

// ExtractConversationIDFromMetadata extracts conversation ID from request body.
func (e *ConversationExtractor) ExtractConversationIDFromMetadata(requestBody json.RawMessage) string {
	return e.ExtractConversationID(requestBody, nil)
}

func (e *ConversationExtractor) normalizeID(id string) string {
	id = trimSpace(id)
	if len(id) > 64 {
		hash := sha256.Sum256([]byte(id))
		return hex.EncodeToString(hash[:8])
	}
	return id
}

func (e *ConversationExtractor) extractFirstUserMessage(body json.RawMessage) string {
	var req struct {
		Messages []struct {
			Role    string      `json:"role"`
			Content interface{} `json:"content"`
		} `json:"messages"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		return ""
	}

	for _, m := range req.Messages {
		if m.Role == "user" {
			return e.contentToString(m.Content)
		}
	}
	return ""
}

func (e *ConversationExtractor) contentToString(content interface{}) string {
	switch v := content.(type) {
	case string:
		return v
	case []interface{}:
		var texts []string
		for _, item := range v {
			if itemMap, ok := item.(map[string]interface{}); ok {
				if itemMap["type"] == "text" {
					if text, ok := itemMap["text"].(string); ok {
						texts = append(texts, text)
					}
				}
			}
		}
		if len(texts) > 0 {
			return texts[0]
		}
	}
	return ""
}

func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && isWhitespace(s[start]) {
		start++
	}
	for end > start && isWhitespace(s[end-1]) {
		end--
	}
	return s[start:end]
}

func isWhitespace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}
