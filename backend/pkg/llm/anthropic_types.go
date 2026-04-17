package llm

// AnthropicMessagesRequest represents an Anthropic Messages API request body.
type AnthropicMessagesRequest struct {
	System          interface{}        `json:"system,omitempty"`
	ToolChoice      interface{}        `json:"tool_choice,omitempty"`
	Metadata        *AnthropicMetadata `json:"metadata,omitempty"`
	Temperature     *float64           `json:"temperature,omitempty"`
	TopP            *float64           `json:"top_p,omitempty"`
	TopK            *int               `json:"top_k,omitempty"`
	Thinking        *AnthropicThinking `json:"thinking,omitempty"`
	ParallelToolUse *bool              `json:"parallel_tool_use,omitempty"`
	Model           string             `json:"model"`
	StopSequences   []string           `json:"stop_sequences,omitempty"`
	Tools           []AnthropicTool    `json:"tools,omitempty"`
	Messages        []AnthropicMessage `json:"messages"`
	MaxTokens       int                `json:"max_tokens"`
	Stream          bool               `json:"stream,omitempty"`
}

// AnthropicThinking represents extended thinking configuration.
type AnthropicThinking struct {
	Type         string `json:"type"`
	BudgetTokens int    `json:"budget_tokens,omitempty"`
}

// AnthropicMessage represents an Anthropic conversation message.
type AnthropicMessage struct {
	Content interface{} `json:"content"`
	Role    string      `json:"role"`
}

// AnthropicSystemBlock represents an Anthropic system message block.
type AnthropicSystemBlock struct {
	CacheControl *struct {
		Type string `json:"type"`
	} `json:"cache_control,omitempty"`
	Type string `json:"type"`
	Text string `json:"text"`
}

// AnthropicContentBlock represents a content block in a message.
type AnthropicContentBlock struct {
	Input     interface{}           `json:"input,omitempty"`
	Content   interface{}           `json:"content,omitempty"`
	Source    *AnthropicImageSource `json:"source,omitempty"`
	Type      string                `json:"type"`
	Text      string                `json:"text,omitempty"`
	ID        string                `json:"id,omitempty"`
	Name      string                `json:"name,omitempty"`
	ToolUseID string                `json:"tool_use_id,omitempty"`
	Thinking  string                `json:"thinking,omitempty"`
	Signature string                `json:"signature,omitempty"`
	IsError   bool                  `json:"is_error,omitempty"`
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
	InputSchema interface{} `json:"input_schema"`
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
}

// AnthropicToolChoice represents tool selection strategy.
type AnthropicToolChoice struct {
	Type string `json:"type"`
	Name string `json:"name,omitempty"`
}

// AnthropicMessagesResponse represents an Anthropic Messages API response.
type AnthropicMessagesResponse struct {
	StopReason   *string                 `json:"stop_reason"`
	StopSequence *string                 `json:"stop_sequence"`
	Usage        *AnthropicUsage         `json:"usage"`
	ID           string                  `json:"id"`
	Type         string                  `json:"type"`
	Role         string                  `json:"role"`
	Model        string                  `json:"model"`
	Content      []AnthropicContentBlock `json:"content"`
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
	Message      *AnthropicMessagesResponse `json:"message,omitempty"`
	Index        *int                       `json:"index,omitempty"`
	ContentBlock *AnthropicContentBlock     `json:"content_block,omitempty"`
	Delta        *AnthropicStreamDelta      `json:"delta,omitempty"`
	Usage        *AnthropicUsage            `json:"usage,omitempty"`
	Type         string                     `json:"type"`
}

// AnthropicStreamDelta represents streaming delta data.
type AnthropicStreamDelta struct {
	StopReason   *string `json:"stop_reason,omitempty"`
	StopSequence *string `json:"stop_sequence,omitempty"`
	Type         string  `json:"type,omitempty"`
	Text         string  `json:"text,omitempty"`
	PartialJSON  string  `json:"partial_json,omitempty"`
	Thinking     string  `json:"thinking,omitempty"`
	Signature    string  `json:"signature,omitempty"`
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
