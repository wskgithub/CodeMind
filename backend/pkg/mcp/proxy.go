package mcp

import (
	"encoding/json"
	"fmt"
	"sync"

	"go.uber.org/zap"
)

// ──────────────────────────────────
// MCP 代理
// 负责将客户端请求转发到上游 MCP 服务
// 并维护连接池和会话映射
// ──────────────────────────────────

// ServiceConnection 上游 MCP 服务连接
type ServiceConnection struct {
	ServiceName string
	MessageURL  string
	Client      *SSEClient
	Events      <-chan SSEEvent
}

// Proxy MCP 代理转发器
type Proxy struct {
	mu          sync.RWMutex
	connections map[string]*ServiceConnection // serviceName -> connection
	logger      *zap.Logger
}

// NewProxy 创建 MCP 代理
func NewProxy(logger *zap.Logger) *Proxy {
	return &Proxy{
		connections: make(map[string]*ServiceConnection),
		logger:      logger,
	}
}

// ConnectService 连接上游 MCP 服务
func (p *Proxy) ConnectService(name, endpoint, authType, authToken, authHeader string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 检查是否已连接
	if _, ok := p.connections[name]; ok {
		return nil // 已连接
	}

	client := NewSSEClient(endpoint, authType, authToken, authHeader)

	messageURL, events, err := client.Connect()
	if err != nil {
		return fmt.Errorf("连接 MCP 服务 '%s' 失败: %w", name, err)
	}

	conn := &ServiceConnection{
		ServiceName: name,
		MessageURL:  messageURL,
		Client:      client,
		Events:      events,
	}

	p.connections[name] = conn
	p.logger.Info("已连接上游 MCP 服务",
		zap.String("service", name),
		zap.String("message_url", messageURL),
	)

	// 发送 initialize 请求
	if err := p.initializeService(conn); err != nil {
		p.logger.Warn("初始化上游 MCP 服务失败（可能仍然可用）",
			zap.String("service", name),
			zap.Error(err),
		)
	}

	return nil
}

// initializeService 向上游服务发送初始化请求
func (p *Proxy) initializeService(conn *ServiceConnection) error {
	initReq := &JSONRPCRequest{
		JSONRPC: JSONRPCVersion,
		ID:      1,
		Method:  MethodInitialize,
	}

	params := InitializeParams{
		ProtocolVersion: MCPProtocolVersion,
		Capabilities:    ClientCapability{},
		ClientInfo: Implementation{
			Name:    "codemind-mcp-gateway",
			Version: "1.0.0",
		},
	}

	paramsData, _ := json.Marshal(params)
	initReq.Params = paramsData

	_, err := conn.Client.SendMessage(conn.MessageURL, initReq)
	if err != nil {
		return err
	}

	// 发送 initialized 通知
	notification := &JSONRPCNotification{
		JSONRPC: JSONRPCVersion,
		Method:  MethodInitialized,
	}

	_, err = conn.Client.SendMessage(conn.MessageURL, notification)
	return err
}

// ForwardRequest 转发 JSON-RPC 请求到指定的上游服务
func (p *Proxy) ForwardRequest(serviceName string, request *JSONRPCRequest) (*JSONRPCResponse, error) {
	p.mu.RLock()
	conn, ok := p.connections[serviceName]
	p.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("MCP 服务 '%s' 未连接", serviceName)
	}

	// 转发请求
	result, err := conn.Client.SendMessage(conn.MessageURL, request)
	if err != nil {
		return nil, fmt.Errorf("转发请求到 MCP 服务 '%s' 失败: %w", serviceName, err)
	}

	if result != nil {
		// 直接从 POST 响应获取了结果
		var resp JSONRPCResponse
		if json.Unmarshal(*result, &resp) == nil {
			return &resp, nil
		}
		// 可能是纯结果对象
		return &JSONRPCResponse{
			JSONRPC: JSONRPCVersion,
			ID:      request.ID,
			Result:  *result,
		}, nil
	}

	// 结果通过 SSE 事件返回 — 等待事件
	return p.waitForResponse(conn, request.ID)
}

// waitForResponse 等待上游 SSE 事件中的响应
func (p *Proxy) waitForResponse(conn *ServiceConnection, requestID interface{}) (*JSONRPCResponse, error) {
	// 从事件通道中查找匹配的响应
	for evt := range conn.Events {
		if evt.Event == "message" {
			var resp JSONRPCResponse
			if err := json.Unmarshal([]byte(evt.Data), &resp); err != nil {
				continue
			}
			// 检查 ID 是否匹配
			if fmt.Sprintf("%v", resp.ID) == fmt.Sprintf("%v", requestID) {
				return &resp, nil
			}
		}
	}

	return nil, fmt.Errorf("等待 MCP 服务响应超时")
}

// DisconnectService 断开与上游 MCP 服务的连接
func (p *Proxy) DisconnectService(name string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.connections, name)
	p.logger.Info("已断开上游 MCP 服务", zap.String("service", name))
}

// IsConnected 检查服务是否已连接
func (p *Proxy) IsConnected(name string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	_, ok := p.connections[name]
	return ok
}

// ListConnectedServices 列出已连接的服务
func (p *Proxy) ListConnectedServices() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var names []string
	for name := range p.connections {
		names = append(names, name)
	}
	return names
}

// HandleRequest 处理来自客户端的 MCP 请求
// 根据方法名和工具/资源名路由到对应的上游服务
func (p *Proxy) HandleRequest(request *JSONRPCRequest, services []ServiceInfo) (*JSONRPCResponse, error) {
	switch request.Method {
	case MethodInitialize:
		return p.handleInitialize(request)
	case MethodPing:
		return NewResponse(request.ID, struct{}{}), nil
	case MethodToolsList:
		return p.handleToolsList(request, services)
	case MethodToolsCall:
		return p.handleToolsCall(request, services)
	case MethodResourcesList:
		return p.handleResourcesList(request, services)
	case MethodResourcesRead:
		return p.handleResourcesRead(request, services)
	default:
		return NewErrorResponse(request.ID, ErrCodeMethodNotFound, "方法不支持: "+request.Method), nil
	}
}

// handleInitialize 处理初始化请求
func (p *Proxy) handleInitialize(request *JSONRPCRequest) (*JSONRPCResponse, error) {
	result := InitializeResult{
		ProtocolVersion: MCPProtocolVersion,
		Capabilities: ServerCapability{
			Tools:     &ToolsCapability{ListChanged: true},
			Resources: &ResourcesCapability{ListChanged: true},
		},
		ServerInfo: Implementation{
			Name:    "codemind-mcp-gateway",
			Version: "1.0.0",
		},
	}
	return NewResponse(request.ID, result), nil
}

// handleToolsList 聚合所有服务的工具列表
func (p *Proxy) handleToolsList(request *JSONRPCRequest, services []ServiceInfo) (*JSONRPCResponse, error) {
	var allTools []Tool

	for _, svc := range services {
		if svc.ToolsSchema != nil {
			var tools []Tool
			if err := json.Unmarshal(svc.ToolsSchema, &tools); err == nil {
				allTools = append(allTools, tools...)
			}
		}
	}

	return NewResponse(request.ID, ToolsListResult{Tools: allTools}), nil
}

// handleToolsCall 根据工具名路由到对应服务
func (p *Proxy) handleToolsCall(request *JSONRPCRequest, services []ServiceInfo) (*JSONRPCResponse, error) {
	var params ToolCallParams
	if err := json.Unmarshal(request.Params, &params); err != nil {
		return NewErrorResponse(request.ID, ErrCodeInvalidParams, "参数解析失败"), nil
	}

	// 查找工具所属的服务
	serviceName := p.findServiceForTool(params.Name, services)
	if serviceName == "" {
		return NewErrorResponse(request.ID, ErrCodeMethodNotFound, "工具不存在: "+params.Name), nil
	}

	// 转发到上游服务
	return p.ForwardRequest(serviceName, request)
}

// handleResourcesList 聚合所有服务的资源列表
func (p *Proxy) handleResourcesList(request *JSONRPCRequest, services []ServiceInfo) (*JSONRPCResponse, error) {
	var allResources []Resource

	for _, svc := range services {
		// 向每个已连接的服务请求资源列表
		if p.IsConnected(svc.Name) {
			resp, err := p.ForwardRequest(svc.Name, request)
			if err == nil && resp.Result != nil {
				var result ResourcesListResult
				if json.Unmarshal(resp.Result, &result) == nil {
					allResources = append(allResources, result.Resources...)
				}
			}
		}
	}

	return NewResponse(request.ID, ResourcesListResult{Resources: allResources}), nil
}

// handleResourcesRead 根据资源 URI 路由到对应服务
func (p *Proxy) handleResourcesRead(request *JSONRPCRequest, services []ServiceInfo) (*JSONRPCResponse, error) {
	// 默认转发到第一个已连接的服务
	for _, svc := range services {
		if p.IsConnected(svc.Name) {
			return p.ForwardRequest(svc.Name, request)
		}
	}
	return NewErrorResponse(request.ID, ErrCodeInternalError, "没有可用的 MCP 服务"), nil
}

// findServiceForTool 查找工具所属的服务
func (p *Proxy) findServiceForTool(toolName string, services []ServiceInfo) string {
	for _, svc := range services {
		if svc.ToolsSchema != nil {
			var tools []Tool
			if err := json.Unmarshal(svc.ToolsSchema, &tools); err == nil {
				for _, t := range tools {
					if t.Name == toolName {
						return svc.Name
					}
				}
			}
		}
	}
	return ""
}

// ServiceInfo 服务简要信息（用于路由决策）
type ServiceInfo struct {
	Name        string
	ToolsSchema json.RawMessage
}
