package mcp

import "encoding/json"

const JSONRPCVersion = "2.0"

const MCPProtocolVersion = "2024-11-05"

// JSONRPCRequest represents a JSON-RPC request message.
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC response message.
type JSONRPCResponse struct {
	ID      interface{}     `json:"id"`
	Error   *JSONRPCError   `json:"error,omitempty"`
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result,omitempty"`
}

// JSONRPCNotification represents a JSON-RPC notification (no ID).
type JSONRPCNotification struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// JSONRPCError represents a JSON-RPC error object.
type JSONRPCError struct {
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message"`
	Code    int         `json:"code"`
}

const (
	ErrCodeParseError     = -32700
	ErrCodeInvalidRequest = -32600
	ErrCodeMethodNotFound = -32601
	ErrCodeInvalidParams  = -32602
	ErrCodeInternalError  = -32603
)

const (
	MethodInitialize    = "initialize"
	MethodInitialized   = "notifications/initialized"
	MethodToolsList     = "tools/list"
	MethodToolsCall     = "tools/call"
	MethodResourcesList = "resources/list"
	MethodResourcesRead = "resources/read"
	MethodPromptsList   = "prompts/list"
	MethodPromptsGet    = "prompts/get"
	MethodPing          = "ping"
)

// InitializeParams represents initialize request parameters.
type InitializeParams struct {
	ProtocolVersion string           `json:"protocolVersion"`
	Capabilities    ClientCapability `json:"capabilities"`
	ClientInfo      Implementation   `json:"clientInfo"`
}

// InitializeResult represents the initialize response.
type InitializeResult struct {
	ProtocolVersion string           `json:"protocolVersion"`
	Capabilities    ServerCapability `json:"capabilities"`
	ServerInfo      Implementation   `json:"serverInfo"`
}

// Implementation contains implementation info.
type Implementation struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ClientCapability declares client capabilities.
type ClientCapability struct {
	Roots    *RootsCapability    `json:"roots,omitempty"`
	Sampling *SamplingCapability `json:"sampling,omitempty"`
}

// ServerCapability declares server capabilities.
type ServerCapability struct {
	Tools     *ToolsCapability     `json:"tools,omitempty"`
	Resources *ResourcesCapability `json:"resources,omitempty"`
	Prompts   *PromptsCapability   `json:"prompts,omitempty"`
}

// RootsCapability declares roots capability.
type RootsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// SamplingCapability declares sampling capability.
type SamplingCapability struct{}

// ToolsCapability declares tools capability.
type ToolsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// ResourcesCapability declares resources capability.
type ResourcesCapability struct {
	Subscribe   bool `json:"subscribe,omitempty"`
	ListChanged bool `json:"listChanged,omitempty"`
}

// PromptsCapability declares prompts capability.
type PromptsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// Tool represents an MCP tool definition.
type Tool struct {
	InputSchema interface{} `json:"inputSchema"`
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
}

// ToolsListResult represents the tools/list response.
type ToolsListResult struct {
	Tools []Tool `json:"tools"`
}

// ToolCallParams represents tool call parameters.
type ToolCallParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments,omitempty"`
}

// ToolCallResult represents tool call result.
type ToolCallResult struct {
	Content []ToolContent `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

// ToolContent represents tool return content.
type ToolContent struct {
	Type string `json:"type"` // "text" | "image" | "resource"
	Text string `json:"text,omitempty"`
}

// Resource represents an MCP resource definition.
type Resource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	MimeType    string `json:"mimeType,omitempty"`
}

// ResourcesListResult represents the resources/list response.
type ResourcesListResult struct {
	Resources []Resource `json:"resources"`
}

// ResourceReadParams represents resource read parameters.
type ResourceReadParams struct {
	URI string `json:"uri"`
}

// ResourceReadResult represents resource read result.
type ResourceReadResult struct {
	Contents []ResourceContent `json:"contents"`
}

// ResourceContent represents resource content.
type ResourceContent struct {
	URI      string `json:"uri"`
	MimeType string `json:"mimeType,omitempty"`
	Text     string `json:"text,omitempty"`
	Blob     string `json:"blob,omitempty"`
}

// NewResponse creates a successful response.
func NewResponse(id interface{}, result interface{}) *JSONRPCResponse {
	data, _ := json.Marshal(result)
	return &JSONRPCResponse{
		JSONRPC: JSONRPCVersion,
		ID:      id,
		Result:  data,
	}
}

// NewErrorResponse creates an error response.
func NewErrorResponse(id interface{}, code int, message string) *JSONRPCResponse {
	return &JSONRPCResponse{
		JSONRPC: JSONRPCVersion,
		ID:      id,
		Error: &JSONRPCError{
			Code:    code,
			Message: message,
		},
	}
}
