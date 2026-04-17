package llm

import (
	"encoding/json"
	"regexp"
	"strings"
)

// EnsureStreamOptions ensures streaming request includes stream_options.include_usage.
func EnsureStreamOptions(rawBody []byte) []byte {
	var body map[string]json.RawMessage
	if err := json.Unmarshal(rawBody, &body); err != nil {
		return rawBody
	}

	if existing, ok := body["stream_options"]; ok {
		var opts map[string]interface{}
		if err := json.Unmarshal(existing, &opts); err == nil {
			if v, ok := opts["include_usage"].(bool); ok && v {
				return rawBody
			}
			opts["include_usage"] = true
			if data, err := json.Marshal(opts); err == nil {
				body["stream_options"] = json.RawMessage(data)
			}
		}
	} else {
		body["stream_options"] = json.RawMessage(`{"include_usage":true}`)
	}

	result, err := json.Marshal(body)
	if err != nil {
		return rawBody
	}
	return result
}

// ExtractUsageFromResponse extracts only the usage field from raw JSON response.
func ExtractUsageFromResponse(rawResp []byte) *Usage {
	var wrapper struct {
		Usage *Usage `json:"usage"`
	}
	if err := json.Unmarshal(rawResp, &wrapper); err != nil {
		return nil
	}
	return wrapper.Usage
}

var thinkTagRegex = regexp.MustCompile(`(?s)<think>.*?</think>\s*`)

// CleanThinkingFromHistory removes thinking content from assistant messages in conversation history.
// Handles both inline <think>...</think> tags and separate reasoning_content fields.
func CleanThinkingFromHistory(rawBody []byte) []byte {
	var body map[string]json.RawMessage
	if err := json.Unmarshal(rawBody, &body); err != nil {
		return rawBody
	}

	messagesRaw, ok := body["messages"]
	if !ok {
		return rawBody
	}

	var messages []map[string]json.RawMessage
	if err := json.Unmarshal(messagesRaw, &messages); err != nil {
		return rawBody
	}

	modified := false
	for i, msg := range messages {
		var role string
		if roleRaw, ok := msg["role"]; ok {
			json.Unmarshal(roleRaw, &role)
		}
		if role != "assistant" {
			continue
		}

		if _, has := msg["reasoning_content"]; has {
			delete(messages[i], "reasoning_content")
			modified = true
		}

		if contentRaw, ok := msg["content"]; ok {
			var content string
			if json.Unmarshal(contentRaw, &content) == nil && strings.Contains(content, "<think>") {
				cleaned := thinkTagRegex.ReplaceAllString(content, "")
				cleaned = strings.TrimSpace(cleaned)
				if cleaned != content {
					newContentJSON, err := json.Marshal(cleaned)
					if err == nil {
						messages[i]["content"] = json.RawMessage(newContentJSON)
						modified = true
					}
				}
			}
		}
	}

	if !modified {
		return rawBody
	}

	newMessagesJSON, err := json.Marshal(messages)
	if err != nil {
		return rawBody
	}
	body["messages"] = json.RawMessage(newMessagesJSON)

	result, err := json.Marshal(body)
	if err != nil {
		return rawBody
	}
	return result
}
