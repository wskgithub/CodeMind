package llm

// AnthropicMessagesRequest represents an Anthropic Messages API request body.
type AnthropicMessagesRequest struct {
	Model           string             `json:"model"`
	Messages        []AnthropicMessage `json:"messages"`
	System          interface{}        `json:"system,omitempty"`
	MaxTokens       int                `json:"max_tokens"`
	Stream          bool               `json:"stream,omitempty"`
	Temperature     *float64           `json:"temperature,omitempty"`
	TopP            *float64           `json:"top_p,omitempty"`
	TopK            *int               `json:"top_k,omitempty"`
	StopSequences   []string           `json:"stop_sequences,omitempty"`
	Metadata        *AnthropicMetadata `json:"metadata,omitempty"`
	Tools           []AnthropicTool    `json:"tools,omitempty"`
	ToolChoice      interface{}        `json:"tool_choice,omitempty"`
	Thinking        *AnthropicThinking `json:"thinking,omitempty"`
	ParallelToolUse *bool              `json:"parallel_tool_use,omitempty"`
}

// AnthropicThinking represents extended thinking configuration.
type AnthropicThinking struct {
	Type         string `json:"type"`
	BudgetTokens int    `json:"budget_tokens,omitempty"`
}

// AnthropicMessage represents an Anthropic conversation message.
type AnthropicMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

// AnthropicSystemBlock represents an Anthropic system message block.
type AnthropicSystemBlock struct {
	Type         string `json:"type"`
	Text         string `json:"text"`
	CacheControl *struct {
		Type string `json:"type"`
	} `json:"cache_control,omitempty"`
}

// AnthropicContentBlock represents a content block in a message.
type AnthropicContentBlock struct {
	Type      string                `json:"type"`
	Text      string                `json:"text,omitempty"`
	ID        string                `json:"id,omitempty"`
	Name      string                `json:"name,omitempty"`
	Input     interface{}           `json:"input,omitempty"`
	ToolUseID string                `json:"tool_use_id,omitempty"`
	Content   interface{}           `json:"content,omitempty"`
	IsError   bool                  `json:"is_error,omitempty"`
	Source    *AnthropicImageSource `json:"source,omitempty"`
	Thinking  string                `json:"thinking,omitempty"`
	Signature string                `json:"signature,omitempty"`
}

// AnthropicImageSource represents image source data.
type AnthropicImageSource struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}

// AnthropicMetadata represents request metadata.
type AnthropicMetadata struct {
	UserID string `json:"user_id,omitempty"`
}

// AnthropicTool represents a tool definition.
type AnthropicTool struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	InputSchema interface{} `json:"input_schema"` // JSON Schema
}

// AnthropicToolChoice represents tool selection strategy.
type AnthropicToolChoice struct {
	Type string `json:"type"`
	Name string `json:"name,omitempty"`
}

// AnthropicMessagesResponse represents an Anthropic Messages API response.
type AnthropicMessagesResponse struct {
	ID           string                  `json:"id"`
	Type         string                  `json:"type"`
	Role         string                  `json:"role"`
	Content      []AnthropicContentBlock `json:"content"`
	Model        string                  `json:"model"`
	StopReason   *string                 `json:"stop_reason"`
	StopSequence *string                 `json:"stop_sequence"`
	Usage        *AnthropicUsage         `json:"usage"`
}

// AnthropicUsage represents Anthropic token usage.
type AnthropicUsage struct {
	InputTokens              int `json:"input_tokens"`
	OutputTokens             int `json:"output_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens,omitempty"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens,omitempty"`
}

// ToUsage converts to the common Usage format.
func (u *AnthropicUsage) ToUsage() *Usage {
	if u == nil {
		return nil
	}
	usage := &Usage{
		PromptTokens:     u.InputTokens,
		CompletionTokens: u.OutputTokens,
		TotalTokens:      u.InputTokens + u.OutputTokens,
	}
	if u.CacheCreationInputTokens > 0 || u.CacheReadInputTokens > 0 {
		usage.PromptTokensDetails = &PromptTokensDetails{
			CacheCreationInputTokens: u.CacheCreationInputTokens,
			CacheReadInputTokens:     u.CacheReadInputTokens,
			CachedTokens:             u.CacheCreationInputTokens + u.CacheReadInputTokens,
		}
	}
	return usage
}

// AnthropicStreamEvent represents a streaming event wrapper.
type AnthropicStreamEvent struct {
	Type         string                     `json:"type"`
	Message      *AnthropicMessagesResponse `json:"message,omitempty"`
	Index        *int                       `json:"index,omitempty"`
	ContentBlock *AnthropicContentBlock     `json:"content_block,omitempty"`
	Delta        *AnthropicStreamDelta      `json:"delta,omitempty"`
	Usage        *AnthropicUsage            `json:"usage,omitempty"`
}

// AnthropicStreamDelta represents streaming delta data.
type AnthropicStreamDelta struct {
	Type         string  `json:"type,omitempty"`
	Text         string  `json:"text,omitempty"`
	PartialJSON  string  `json:"partial_json,omitempty"`
	Thinking     string  `json:"thinking,omitempty"`
	Signature    string  `json:"signature,omitempty"`
	StopReason   *string `json:"stop_reason,omitempty"`
	StopSequence *string `json:"stop_sequence,omitempty"`
}

// AnthropicErrorResponse represents an Anthropic error response.
type AnthropicErrorResponse struct {
	Type  string `json:"type"`
	Error struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

const (
	AnthropicEventMessageStart      = "message_start"
	AnthropicEventContentBlockStart = "content_block_start"
	AnthropicEventContentBlockDelta = "content_block_delta"
	AnthropicEventContentBlockStop  = "content_block_stop"
	AnthropicEventMessageDelta      = "message_delta"
	AnthropicEventMessageStop       = "message_stop"
	AnthropicEventPing              = "ping"
	AnthropicEventError             = "error"
)
