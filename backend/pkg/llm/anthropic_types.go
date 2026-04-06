package llm

// ──────────────────────────────────
// Anthropic 原生 API — 请求类型
// 兼容 Anthropic Messages API (2023-06-01)
// ──────────────────────────────────

// AnthropicMessagesRequest Anthropic Messages API 请求体
// 注意：直连 Anthropic 后端时使用原始请求体透传，此结构体仅用于跨格式转换场景
type AnthropicMessagesRequest struct {
	Model            string                  `json:"model"`
	Messages         []AnthropicMessage      `json:"messages"`
	System           interface{}             `json:"system,omitempty"`            // string 或 []AnthropicSystemBlock
	MaxTokens        int                     `json:"max_tokens"`                  // Anthropic 要求必填
	Stream           bool                    `json:"stream,omitempty"`
	Temperature      *float64                `json:"temperature,omitempty"`
	TopP             *float64                `json:"top_p,omitempty"`
	TopK             *int                    `json:"top_k,omitempty"`
	StopSequences    []string                `json:"stop_sequences,omitempty"`
	Metadata         *AnthropicMetadata      `json:"metadata,omitempty"`
	Tools            []AnthropicTool         `json:"tools,omitempty"`
	ToolChoice       interface{}             `json:"tool_choice,omitempty"`       // string 或 AnthropicToolChoice
	Thinking         *AnthropicThinking      `json:"thinking,omitempty"`          // 扩展思考配置
	ParallelToolUse  *bool                   `json:"parallel_tool_use,omitempty"` // 是否允许并行工具调用
}

// AnthropicThinking 扩展思考配置
type AnthropicThinking struct {
	Type         string `json:"type"`                    // "enabled" | "disabled" | "adaptive"
	BudgetTokens int    `json:"budget_tokens,omitempty"` // 思考预算 token 数
}

// AnthropicMessage Anthropic 对话消息
type AnthropicMessage struct {
	Role    string      `json:"role"`    // "user" 或 "assistant"
	Content interface{} `json:"content"` // string 或 []AnthropicContentBlock
}

// AnthropicSystemBlock Anthropic system 消息块（高级用法）
type AnthropicSystemBlock struct {
	Type         string `json:"type"`                    // "text"
	Text         string `json:"text"`
	CacheControl *struct {
		Type string `json:"type"` // "ephemeral"
	} `json:"cache_control,omitempty"`
}

// AnthropicContentBlock 内容块（消息中的一个元素）
// 支持: text, image, tool_use, tool_result, thinking, document, server_tool_use
type AnthropicContentBlock struct {
	Type  string      `json:"type"`            // 内容块类型
	Text  string      `json:"text,omitempty"`  // type="text" 时使用
	ID    string      `json:"id,omitempty"`    // type="tool_use" | "server_tool_use" 时的工具调用 ID
	Name  string      `json:"name,omitempty"`  // type="tool_use" | "server_tool_use" 时的工具名称
	Input interface{} `json:"input,omitempty"` // type="tool_use" | "server_tool_use" 时的输入参数

	// type="tool_result" 相关字段
	ToolUseID string      `json:"tool_use_id,omitempty"`
	Content   interface{} `json:"content,omitempty"` // 工具返回内容 (string 或 []ContentBlock)
	IsError   bool        `json:"is_error,omitempty"`

	// type="image" 相关字段
	Source *AnthropicImageSource `json:"source,omitempty"`

	// type="thinking" 相关字段（扩展思考）
	Thinking  string `json:"thinking,omitempty"`  // 思考内容文本
	Signature string `json:"signature,omitempty"` // 思考块完整性签名
}

// AnthropicImageSource 图片来源
type AnthropicImageSource struct {
	Type      string `json:"type"`       // "base64"
	MediaType string `json:"media_type"` // "image/jpeg" | "image/png" 等
	Data      string `json:"data"`       // Base64 编码数据
}

// AnthropicMetadata 请求元数据
type AnthropicMetadata struct {
	UserID string `json:"user_id,omitempty"`
}

// AnthropicTool 工具定义
type AnthropicTool struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	InputSchema interface{} `json:"input_schema"` // JSON Schema
}

// AnthropicToolChoice 工具选择策略
type AnthropicToolChoice struct {
	Type string `json:"type"` // "auto" | "any" | "tool"
	Name string `json:"name,omitempty"`
}

// ──────────────────────────────────
// Anthropic 原生 API — 响应类型
// ──────────────────────────────────

// AnthropicMessagesResponse Anthropic Messages API 非流式响应
type AnthropicMessagesResponse struct {
	ID           string                  `json:"id"`
	Type         string                  `json:"type"`           // "message"
	Role         string                  `json:"role"`           // "assistant"
	Content      []AnthropicContentBlock `json:"content"`
	Model        string                  `json:"model"`
	StopReason   *string                 `json:"stop_reason"`    // "end_turn" | "max_tokens" | "stop_sequence" | "tool_use"
	StopSequence *string                 `json:"stop_sequence"`
	Usage        *AnthropicUsage         `json:"usage"`
}

// AnthropicUsage Anthropic Token 用量
type AnthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	// 缓存相关（可选）
	CacheCreationInputTokens int `json:"cache_creation_input_tokens,omitempty"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens,omitempty"`
}

// ToUsage 转换为通用 Usage 格式
// 保留缓存相关字段：CacheCreationInputTokens 和 CacheReadInputTokens
func (u *AnthropicUsage) ToUsage() *Usage {
	if u == nil {
		return nil
	}
	usage := &Usage{
		PromptTokens:     u.InputTokens,
		CompletionTokens: u.OutputTokens,
		TotalTokens:      u.InputTokens + u.OutputTokens,
	}
	// 转换缓存相关字段
	if u.CacheCreationInputTokens > 0 || u.CacheReadInputTokens > 0 {
		usage.PromptTokensDetails = &PromptTokensDetails{
			CacheCreationInputTokens: u.CacheCreationInputTokens,
			CacheReadInputTokens:     u.CacheReadInputTokens,
			CachedTokens:             u.CacheCreationInputTokens + u.CacheReadInputTokens,
		}
	}
	return usage
}

// ──────────────────────────────────
// Anthropic 流式响应事件类型
// ──────────────────────────────────

// AnthropicStreamEvent 流式事件包装
type AnthropicStreamEvent struct {
	Type string `json:"type"` // 事件类型
	// 以下字段根据事件类型选择性填充
	Message      *AnthropicMessagesResponse `json:"message,omitempty"`       // message_start
	Index        *int                       `json:"index,omitempty"`         // content_block_start/delta
	ContentBlock *AnthropicContentBlock     `json:"content_block,omitempty"` // content_block_start
	Delta        *AnthropicStreamDelta      `json:"delta,omitempty"`         // content_block_delta / message_delta
	Usage        *AnthropicUsage            `json:"usage,omitempty"`         // message_delta (最终用量)
}

// AnthropicStreamDelta 流式增量数据
type AnthropicStreamDelta struct {
	Type         string  `json:"type,omitempty"`          // "text_delta" | "input_json_delta" | "thinking_delta" | "signature_delta"
	Text         string  `json:"text,omitempty"`          // text_delta 时的文本增量
	PartialJSON  string  `json:"partial_json,omitempty"`  // input_json_delta 时的 JSON 增量
	Thinking     string  `json:"thinking,omitempty"`      // thinking_delta 时的思考增量
	Signature    string  `json:"signature,omitempty"`     // signature_delta 时的签名数据
	StopReason   *string `json:"stop_reason,omitempty"`   // message_delta 中的停止原因
	StopSequence *string `json:"stop_sequence,omitempty"` // message_delta 中的停止序列
}

// ──────────────────────────────────
// Anthropic 错误响应
// ──────────────────────────────────

// AnthropicErrorResponse Anthropic 格式错误响应
type AnthropicErrorResponse struct {
	Type  string `json:"type"` // "error"
	Error struct {
		Type    string `json:"type"`    // "invalid_request_error" | "authentication_error" 等
		Message string `json:"message"`
	} `json:"error"`
}

// ──────────────────────────────────
// Anthropic 流式事件类型常量
// ──────────────────────────────────

const (
	// AnthropicEventMessageStart 消息开始事件
	AnthropicEventMessageStart = "message_start"
	// AnthropicEventContentBlockStart 内容块开始事件
	AnthropicEventContentBlockStart = "content_block_start"
	// AnthropicEventContentBlockDelta 内容块增量事件
	AnthropicEventContentBlockDelta = "content_block_delta"
	// AnthropicEventContentBlockStop 内容块结束事件
	AnthropicEventContentBlockStop = "content_block_stop"
	// AnthropicEventMessageDelta 消息增量事件（含最终用量）
	AnthropicEventMessageDelta = "message_delta"
	// AnthropicEventMessageStop 消息结束事件
	AnthropicEventMessageStop = "message_stop"
	// AnthropicEventPing 心跳事件
	AnthropicEventPing = "ping"
	// AnthropicEventError 错误事件
	AnthropicEventError = "error"
)
