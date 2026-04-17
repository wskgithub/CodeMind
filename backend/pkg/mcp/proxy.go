package mcp

import (
	"encoding/json"
	"fmt"
	"sync"

	"go.uber.org/zap"
)

// ServiceConnection represents an upstream MCP service connection.
type ServiceConnection struct {
	ServiceName string
	MessageURL  string
	Client      *SSEClient
	Events      <-chan SSEEvent
}

// Proxy forwards requests to upstream MCP services.
type Proxy struct {
	mu          sync.RWMutex
	connections map[string]*ServiceConnection // serviceName -> connection
	logger      *zap.Logger
}

// NewProxy creates a new MCP proxy.
func NewProxy(logger *zap.Logger) *Proxy {
	return &Proxy{
		connections: make(map[string]*ServiceConnection),
		logger:      logger,
	}
}

// ConnectService connects to an upstream MCP service.
func (p *Proxy) ConnectService(name, endpoint, authType, authToken, authHeader string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, ok := p.connections[name]; ok {
		return nil
	}

	client := NewSSEClient(endpoint, authType, authToken, authHeader)

	messageURL, events, err := client.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to MCP service '%s': %w", name, err)
	}

	conn := &ServiceConnection{
		ServiceName: name,
		MessageURL:  messageURL,
		Client:      client,
		Events:      events,
	}

	p.connections[name] = conn
	p.logger.Info("connected to upstream MCP service",
		zap.String("service", name),
		zap.String("message_url", messageURL),
	)

	if err := p.initializeService(conn); err != nil {
		p.logger.Warn("failed to initialize upstream MCP service (may still work)",
			zap.String("service", name),
			zap.Error(err),
		)
	}

	return nil
}

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

	notification := &JSONRPCNotification{
		JSONRPC: JSONRPCVersion,
		Method:  MethodInitialized,
	}

	_, err = conn.Client.SendMessage(conn.MessageURL, notification)
	return err
}

// ForwardRequest forwards a JSON-RPC request to the specified upstream service.
func (p *Proxy) ForwardRequest(serviceName string, request *JSONRPCRequest) (*JSONRPCResponse, error) {
	p.mu.RLock()
	conn, ok := p.connections[serviceName]
	p.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("MCP service '%s' not connected", serviceName)
	}

	result, err := conn.Client.SendMessage(conn.MessageURL, request)
	if err != nil {
		return nil, fmt.Errorf("failed to forward request to MCP service '%s': %w", serviceName, err)
	}

	if result != nil {
		var resp JSONRPCResponse
		if json.Unmarshal(*result, &resp) == nil {
			return &resp, nil
		}
		return &JSONRPCResponse{
			JSONRPC: JSONRPCVersion,
			ID:      request.ID,
			Result:  *result,
		}, nil
	}

	return p.waitForResponse(conn, request.ID)
}

func (p *Proxy) waitForResponse(conn *ServiceConnection, requestID interface{}) (*JSONRPCResponse, error) {
	for evt := range conn.Events {
		if evt.Event == "message" {
			var resp JSONRPCResponse
			if err := json.Unmarshal([]byte(evt.Data), &resp); err != nil {
				continue
			}
			if fmt.Sprintf("%v", resp.ID) == fmt.Sprintf("%v", requestID) {
				return &resp, nil
			}
		}
	}

	return nil, fmt.Errorf("timeout waiting for MCP service response")
}

// DisconnectService disconnects from an upstream MCP service.
func (p *Proxy) DisconnectService(name string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.connections, name)
	p.logger.Info("disconnected from upstream MCP service", zap.String("service", name))
}

// IsConnected checks if a service is connected.
func (p *Proxy) IsConnected(name string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	_, ok := p.connections[name]
	return ok
}

// ListConnectedServices returns a list of connected service names.
func (p *Proxy) ListConnectedServices() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var names []string
	for name := range p.connections {
		names = append(names, name)
	}
	return names
}

// HandleRequest routes incoming MCP requests to the appropriate upstream service.
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
		return NewErrorResponse(request.ID, ErrCodeMethodNotFound, "method not supported: "+request.Method), nil
	}
}

// handleInitialize handles the initialize request.
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

// handleToolsList aggregates tools from all services.
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

// handleToolsCall routes tool calls to the appropriate service.
func (p *Proxy) handleToolsCall(request *JSONRPCRequest, services []ServiceInfo) (*JSONRPCResponse, error) {
	var params ToolCallParams
	if err := json.Unmarshal(request.Params, &params); err != nil {
		return NewErrorResponse(request.ID, ErrCodeInvalidParams, "failed to parse params"), nil
	}

	serviceName := p.findServiceForTool(params.Name, services)
	if serviceName == "" {
		return NewErrorResponse(request.ID, ErrCodeMethodNotFound, "tool not found: "+params.Name), nil
	}

	return p.ForwardRequest(serviceName, request)
}

// handleResourcesList aggregates resources from all services.
func (p *Proxy) handleResourcesList(request *JSONRPCRequest, services []ServiceInfo) (*JSONRPCResponse, error) {
	var allResources []Resource

	for _, svc := range services {
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

// handleResourcesRead routes resource reads to the appropriate service.
func (p *Proxy) handleResourcesRead(request *JSONRPCRequest, services []ServiceInfo) (*JSONRPCResponse, error) {
	for _, svc := range services {
		if p.IsConnected(svc.Name) {
			return p.ForwardRequest(svc.Name, request)
		}
	}
	return NewErrorResponse(request.ID, ErrCodeInternalError, "no available MCP service"), nil
}

// findServiceForTool finds the service that owns the specified tool.
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

// ServiceInfo contains service metadata for routing decisions.
type ServiceInfo struct {
	Name        string
	ToolsSchema json.RawMessage
}
