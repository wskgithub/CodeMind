package llm

import (
	"encoding/json"
	"regexp"
	"strings"
)

// ──────────────────────────────────
// 原始请求体透传工具
// 代理层应当尽可能透明：只提取路由所需的最小信息，
// 将完整的原始请求体转发给 LLM，避免丢失任何字段。
// ──────────────────────────────────

// EnsureStreamOptions 确保流式请求包含 stream_options.include_usage
// 这是让 LLM 在流式响应的最后一个 chunk 中返回 token 用量信息的必要条件
//
// 策略：
//   - 若请求中已有 stream_options 且 include_usage=true，直接返回原始数据
//   - 若请求中已有 stream_options 但缺少 include_usage，补充该字段
//   - 若请求中没有 stream_options，新增该字段
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

// ExtractUsageFromResponse 从原始 JSON 响应中仅提取 usage 字段
// 用于非流式响应的用量统计，避免完整解析响应体
func ExtractUsageFromResponse(rawResp []byte) *Usage {
	var wrapper struct {
		Usage *Usage `json:"usage"`
	}
	if err := json.Unmarshal(rawResp, &wrapper); err != nil {
		return nil
	}
	return wrapper.Usage
}

// ──────────────────────────────────
// Thinking 模型对话历史优化
//
// 问题背景：
// Thinking 模型（如 Qwen3-*-Thinking）在流式输出时会生成大量 thinking/reasoning 内容。
// 当客户端（如 OpenCode）将上一轮 assistant 的完整响应（含 thinking）放入对话历史时：
//   1. 巨量 thinking 内容填满 LLM 上下文窗口，导致真正的对话内容被截断（上下文丢失）
//   2. 模型看到自己的 thinking 内容可能产生混乱行为
//   3. 浪费大量 prompt tokens
//
// 解决方案：
// 代理层在转发请求前，自动清理对话历史中 assistant 消息的 thinking 内容，
// 同时保留 tools、stream_options 等所有其他字段不变。
// ──────────────────────────────────

// thinkTagRegex 匹配 <think>...</think> 标签及其内容（含跨行）
var thinkTagRegex = regexp.MustCompile(`(?s)<think>.*?</think>\s*`)

// CleanThinkingFromHistory 清理对话历史中 assistant 消息的 thinking 内容
//
// 处理两种 thinking 格式：
//   - 内联标签：<think>...</think> 嵌入在 content 字符串中
//   - 独立字段：reasoning_content 作为消息的独立字段
//
// 只修改 messages 数组中的 assistant 消息，其他所有字段（tools 等）完整保留
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

		// 清理 1：移除 reasoning_content 字段
		if _, has := msg["reasoning_content"]; has {
			delete(messages[i], "reasoning_content")
			modified = true
		}

		// 清理 2：移除 content 中的 <think>...</think> 标签
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
