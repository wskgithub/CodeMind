package mcp

import "encoding/json"

// ──────────────────────────────────
// MCP 协议 — JSON-RPC 2.0 消息类型
// 遵循 Model Context Protocol 规范
// ──────────────────────────────────

// JSON-RPC 版本常量
const JSONRPCVersion = "2.0"

// MCP 协议版本
const MCPProtocolVersion = "2024-11-05"

// ──────────────────────────────────
// 基础 JSON-RPC 消息类型
// ──────────────────────────────────

// JSONRPCRequest JSON-RPC 请求消息
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`     // string 或 number
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// JSONRPCResponse JSON-RPC 响应消息
type JSONRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *JSONRPCError   `json:"error,omitempty"`
}

// JSONRPCNotification JSON-RPC 通知消息（无 ID）
type JSONRPCNotification struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// JSONRPCError JSON-RPC 错误对象
type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ──────────────────────────────────
// 标准 JSON-RPC 错误码
// ──────────────────────────────────

const (
	// ErrCodeParseError 解析错误
	ErrCodeParseError = -32700
	// ErrCodeInvalidRequest 无效请求
	ErrCodeInvalidRequest = -32600
	// ErrCodeMethodNotFound 方法未找到
	ErrCodeMethodNotFound = -32601
	// ErrCodeInvalidParams 无效参数
	ErrCodeInvalidParams = -32602
	// ErrCodeInternalError 内部错误
	ErrCodeInternalError = -32603
)

// ──────────────────────────────────
// MCP 标准方法名
// ──────────────────────────────────

const (
	// MethodInitialize 客户端初始化
	MethodInitialize = "initialize"
	// MethodInitialized 初始化完成通知
	MethodInitialized = "notifications/initialized"
	// MethodToolsList 列出工具
	MethodToolsList = "tools/list"
	// MethodToolsCall 调用工具
	MethodToolsCall = "tools/call"
	// MethodResourcesList 列出资源
	MethodResourcesList = "resources/list"
	// MethodResourcesRead 读取资源
	MethodResourcesRead = "resources/read"
	// MethodPromptsList 列出提示
	MethodPromptsList = "prompts/list"
	// MethodPromptsGet 获取提示
	MethodPromptsGet = "prompts/get"
	// MethodPing Ping
	MethodPing = "ping"
)

// ──────────────────────────────────
// MCP 初始化相关类型
// ──────────────────────────────────

// InitializeParams 初始化请求参数
type InitializeParams struct {
	ProtocolVersion string           `json:"protocolVersion"`
	Capabilities    ClientCapability `json:"capabilities"`
	ClientInfo      Implementation   `json:"clientInfo"`
}

// InitializeResult 初始化响应
type InitializeResult struct {
	ProtocolVersion string           `json:"protocolVersion"`
	Capabilities    ServerCapability `json:"capabilities"`
	ServerInfo      Implementation   `json:"serverInfo"`
}

// Implementation 实现信息
type Implementation struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ClientCapability 客户端能力声明
type ClientCapability struct {
	Roots    *RootsCapability    `json:"roots,omitempty"`
	Sampling *SamplingCapability `json:"sampling,omitempty"`
}

// ServerCapability 服务端能力声明
type ServerCapability struct {
	Tools     *ToolsCapability     `json:"tools,omitempty"`
	Resources *ResourcesCapability `json:"resources,omitempty"`
	Prompts   *PromptsCapability   `json:"prompts,omitempty"`
}

// RootsCapability 根目录能力
type RootsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// SamplingCapability 采样能力
type SamplingCapability struct{}

// ToolsCapability 工具能力
type ToolsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// ResourcesCapability 资源能力
type ResourcesCapability struct {
	Subscribe   bool `json:"subscribe,omitempty"`
	ListChanged bool `json:"listChanged,omitempty"`
}

// PromptsCapability 提示能力
type PromptsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// ──────────────────────────────────
// MCP 工具相关类型
// ──────────────────────────────────

// Tool MCP 工具定义
type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	InputSchema interface{} `json:"inputSchema"` // JSON Schema
}

// ToolsListResult 工具列表响应
type ToolsListResult struct {
	Tools []Tool `json:"tools"`
}

// ToolCallParams 工具调用参数
type ToolCallParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments,omitempty"`
}

// ToolCallResult 工具调用结果
type ToolCallResult struct {
	Content []ToolContent `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

// ToolContent 工具返回内容
type ToolContent struct {
	Type string `json:"type"` // "text" | "image" | "resource"
	Text string `json:"text,omitempty"`
}

// ──────────────────────────────────
// MCP 资源相关类型
// ──────────────────────────────────

// Resource MCP 资源定义
type Resource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	MimeType    string `json:"mimeType,omitempty"`
}

// ResourcesListResult 资源列表响应
type ResourcesListResult struct {
	Resources []Resource `json:"resources"`
}

// ResourceReadParams 资源读取参数
type ResourceReadParams struct {
	URI string `json:"uri"`
}

// ResourceReadResult 资源读取结果
type ResourceReadResult struct {
	Contents []ResourceContent `json:"contents"`
}

// ResourceContent 资源内容
type ResourceContent struct {
	URI      string `json:"uri"`
	MimeType string `json:"mimeType,omitempty"`
	Text     string `json:"text,omitempty"`
	Blob     string `json:"blob,omitempty"` // Base64 编码
}

// ──────────────────────────────────
// 辅助构造函数
// ──────────────────────────────────

// NewResponse 创建成功响应
func NewResponse(id interface{}, result interface{}) *JSONRPCResponse {
	data, _ := json.Marshal(result)
	return &JSONRPCResponse{
		JSONRPC: JSONRPCVersion,
		ID:      id,
		Result:  data,
	}
}

// NewErrorResponse 创建错误响应
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
