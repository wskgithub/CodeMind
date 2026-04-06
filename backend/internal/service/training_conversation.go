package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"github.com/google/uuid"
)

// ConversationExtractor 会话 ID 提取器
type ConversationExtractor struct{}

// NewConversationExtractor 创建提取器
func NewConversationExtractor() *ConversationExtractor {
	return &ConversationExtractor{}
}

// ExtractConversationID 从请求中提取或生成会话 ID
// 优先级：
// 1. 请求头 X-Conversation-ID
// 2. 请求体 conversation_id 字段
// 3. 基于首条 user 消息生成确定性 ID
// 4. 生成新的 UUID
func (e *ConversationExtractor) ExtractConversationID(
	requestBody json.RawMessage,
	headers map[string]string,
) string {
	// 1. 检查请求头
	if id := headers["X-Conversation-Id"]; id != "" {
		return e.normalizeID(id)
	}
	if id := headers["X-Session-Id"]; id != "" {
		return e.normalizeID(id)
	}

	// 2. 检查请求体
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

	// 3. 基于 messages 生成确定性 ID（同一对话历史生成相同 ID）
	if firstUserMsg := e.extractFirstUserMessage(requestBody); firstUserMsg != "" {
		hash := sha256.Sum256([]byte(firstUserMsg))
		return hex.EncodeToString(hash[:8]) // 16 字符
	}

	// 4. 生成新的 UUID（截取前 16 字符）
	return uuid.New().String()[:16]
}

// ExtractConversationIDFromMetadata 从元数据中提取会话 ID（简化版本）
func (e *ConversationExtractor) ExtractConversationIDFromMetadata(requestBody json.RawMessage) string {
	return e.ExtractConversationID(requestBody, nil)
}

// normalizeID 标准化 ID 格式
func (e *ConversationExtractor) normalizeID(id string) string {
	// 去除首尾空白
	id = trimSpace(id)
	// 限制长度
	if len(id) > 64 {
		// 对超长 ID 取 hash
		hash := sha256.Sum256([]byte(id))
		return hex.EncodeToString(hash[:8])
	}
	return id
}

// extractFirstUserMessage 提取首条 user 消息内容
func (e *ConversationExtractor) extractFirstUserMessage(body json.RawMessage) string {
	var req struct {
		Messages []struct {
			Role    string `json:"role"`
			Content interface{} `json:"content"` // 可能是 string 或 []interface{}
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

// contentToString 将 content 字段转换为字符串
func (e *ConversationExtractor) contentToString(content interface{}) string {
	switch v := content.(type) {
	case string:
		return v
	case []interface{}:
		// 处理多模态内容，提取文本
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
			return texts[0] // 返回第一个文本
		}
	}
	return ""
}

// trimSpace 去除空白
func trimSpace(s string) string {
	// 简单实现
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

// isWhitespace 检查是否是空白字符
func isWhitespace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}
