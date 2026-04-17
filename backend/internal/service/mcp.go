package service

import (
	"encoding/json"
	"fmt"

	"codemind/internal/model"
	"codemind/internal/model/dto"
	"codemind/internal/repository"
	mcpPkg "codemind/pkg/mcp"

	"go.uber.org/zap"
)

// MCPService handles MCP service management.
type MCPService struct {
	mcpRepo  *repository.MCPRepository
	proxy    *mcpPkg.Proxy
	logger   *zap.Logger
}

// NewMCPService creates a new MCP service instance.
func NewMCPService(
	mcpRepo *repository.MCPRepository,
	proxy *mcpPkg.Proxy,
	logger *zap.Logger,
) *MCPService {
	return &MCPService{
		mcpRepo: mcpRepo,
		proxy:   proxy,
		logger:  logger,
	}
}

// GetProxy returns the MCP proxy.
func (s *MCPService) GetProxy() *mcpPkg.Proxy {
	return s.proxy
}

// CreateService creates an MCP service.
func (s *MCPService) CreateService(req *dto.CreateMCPServiceRequest) (*model.MCPService, error) {
	if existing, _ := s.mcpRepo.GetServiceByName(req.Name); existing != nil {
		return nil, fmt.Errorf("service name '%s' already exists", req.Name)
	}

	var authConfig json.RawMessage
	if req.AuthConfig != nil {
		data, err := json.Marshal(req.AuthConfig)
		if err != nil {
			return nil, fmt.Errorf("invalid auth config format: %w", err)
		}
		authConfig = data
	}

	svc := &model.MCPService{
		Name:          req.Name,
		DisplayName:   req.DisplayName,
		Description:   req.Description,
		EndpointURL:   req.EndpointURL,
		TransportType: req.TransportType,
		Status:        model.MCPServiceEnabled,
		AuthType:      req.AuthType,
		AuthConfig:    authConfig,
	}

	if err := s.mcpRepo.CreateService(svc); err != nil {
		return nil, fmt.Errorf("failed to create MCP service: %w", err)
	}

	s.logger.Info("MCP service created", zap.String("name", svc.Name), zap.Int64("id", svc.ID))
	return svc, nil
}

// GetService returns service details.
func (s *MCPService) GetService(id int64) (*model.MCPService, error) {
	return s.mcpRepo.GetServiceByID(id)
}

// ListServices returns service list.
func (s *MCPService) ListServices(status string) ([]dto.MCPServiceResponse, error) {
	services, err := s.mcpRepo.ListServices(status)
	if err != nil {
		return nil, err
	}

	var result []dto.MCPServiceResponse
	for _, svc := range services {
		toolsCount := 0
		if svc.ToolsSchema != nil {
			var tools []interface{}
			if json.Unmarshal(svc.ToolsSchema, &tools) == nil {
				toolsCount = len(tools)
			}
		}

		result = append(result, dto.MCPServiceResponse{
			ID:            svc.ID,
			Name:          svc.Name,
			DisplayName:   svc.DisplayName,
			Description:   svc.Description,
			EndpointURL:   svc.EndpointURL,
			TransportType: svc.TransportType,
			Status:        svc.Status,
			AuthType:      svc.AuthType,
			ToolsCount:    toolsCount,
			Connected:     s.proxy.IsConnected(svc.Name),
			CreatedAt:     svc.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:     svc.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return result, nil
}

// UpdateService updates service info.
func (s *MCPService) UpdateService(id int64, req *dto.UpdateMCPServiceRequest) error {
	svc, err := s.mcpRepo.GetServiceByID(id)
	if err != nil {
		return fmt.Errorf("service not found")
	}

	if req.DisplayName != nil {
		svc.DisplayName = *req.DisplayName
	}
	if req.Description != nil {
		svc.Description = *req.Description
	}
	if req.EndpointURL != nil {
		svc.EndpointURL = *req.EndpointURL
	}
	if req.TransportType != nil {
		svc.TransportType = *req.TransportType
	}
	if req.Status != nil {
		svc.Status = *req.Status
	}
	if req.AuthType != nil {
		svc.AuthType = *req.AuthType
	}
	if req.AuthConfig != nil {
		data, _ := json.Marshal(req.AuthConfig)
		svc.AuthConfig = data
	}

	return s.mcpRepo.UpdateService(svc)
}

// DeleteService deletes a service.
func (s *MCPService) DeleteService(id int64) error {
	svc, err := s.mcpRepo.GetServiceByID(id)
	if err != nil {
		return fmt.Errorf("service not found")
	}

	s.proxy.DisconnectService(svc.Name)
	_ = s.mcpRepo.DeleteAccessRulesByService(id)
	return s.mcpRepo.DeleteService(id)
}

// SyncTools synchronizes service tools.
func (s *MCPService) SyncTools(id int64) error {
	svc, err := s.mcpRepo.GetServiceByID(id)
	if err != nil {
		return fmt.Errorf("service not found")
	}

	if !s.proxy.IsConnected(svc.Name) {
		authToken, authHeader := s.extractAuth(svc)
		if err := s.proxy.ConnectService(svc.Name, svc.EndpointURL, svc.AuthType, authToken, authHeader); err != nil {
			return fmt.Errorf("failed to connect service: %w", err)
		}
	}

	req := &mcpPkg.JSONRPCRequest{
		JSONRPC: mcpPkg.JSONRPCVersion,
		ID:      "sync-tools",
		Method:  mcpPkg.MethodToolsList,
	}

	resp, err := s.proxy.ForwardRequest(svc.Name, req)
	if err != nil {
		return fmt.Errorf("failed to get tools list: %w", err)
	}

	if resp.Error != nil {
		return fmt.Errorf("upstream service error: %s", resp.Error.Message)
	}

	var result mcpPkg.ToolsListResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return fmt.Errorf("failed to parse tools list: %w", err)
	}

	toolsData, _ := json.Marshal(result.Tools)
	return s.mcpRepo.UpdateToolsSchema(id, toolsData)
}

// GetServiceTools returns service tools.
func (s *MCPService) GetServiceTools(id int64) ([]dto.MCPToolInfo, error) {
	svc, err := s.mcpRepo.GetServiceByID(id)
	if err != nil {
		return nil, fmt.Errorf("service not found")
	}

	if svc.ToolsSchema == nil {
		return []dto.MCPToolInfo{}, nil
	}

	var tools []mcpPkg.Tool
	if err := json.Unmarshal(svc.ToolsSchema, &tools); err != nil {
		return nil, fmt.Errorf("failed to parse tools list: %w", err)
	}

	var result []dto.MCPToolInfo
	for _, t := range tools {
		result = append(result, dto.MCPToolInfo{
			Name:        t.Name,
			Description: t.Description,
			ServiceName: svc.Name,
		})
	}

	return result, nil
}

// SetAccessRule sets an access rule.
func (s *MCPService) SetAccessRule(req *dto.SetMCPAccessRuleRequest) error {
	rule := &model.MCPAccessRule{
		ServiceID:  req.ServiceID,
		TargetType: req.TargetType,
		TargetID:   req.TargetID,
		Allowed:    req.Allowed,
	}
	return s.mcpRepo.UpsertAccessRule(rule)
}

// ListAccessRules returns access rules for a service.
func (s *MCPService) ListAccessRules(serviceID int64) ([]model.MCPAccessRule, error) {
	return s.mcpRepo.ListAccessRules(serviceID)
}

// DeleteAccessRule deletes an access rule.
func (s *MCPService) DeleteAccessRule(id int64) error {
	return s.mcpRepo.DeleteAccessRule(id)
}

// CheckAccess checks if user has access to a service.
func (s *MCPService) CheckAccess(serviceID, userID int64, deptID *int64, role string) bool {
	return s.mcpRepo.CheckAccess(serviceID, userID, deptID, role)
}

// GetServiceInfosForGateway returns enabled service info for MCP gateway.
func (s *MCPService) GetServiceInfosForGateway() ([]mcpPkg.ServiceInfo, error) {
	services, err := s.mcpRepo.ListEnabledServices()
	if err != nil {
		return nil, err
	}

	var infos []mcpPkg.ServiceInfo
	for _, svc := range services {
		infos = append(infos, mcpPkg.ServiceInfo{
			Name:        svc.Name,
			ToolsSchema: svc.ToolsSchema,
		})
	}
	return infos, nil
}

// FilterAccessibleServices filters services user has access to.
func (s *MCPService) FilterAccessibleServices(services []mcpPkg.ServiceInfo, userID int64, deptID *int64, role string) []mcpPkg.ServiceInfo {
	var allowed []mcpPkg.ServiceInfo
	for _, info := range services {
		svc, err := s.mcpRepo.GetServiceByName(info.Name)
		if err != nil {
			continue
		}
		if s.mcpRepo.CheckAccess(svc.ID, userID, deptID, role) {
			allowed = append(allowed, info)
		}
	}
	return allowed
}

func (s *MCPService) extractAuth(svc *model.MCPService) (token, header string) {
	switch svc.AuthType {
	case model.MCPAuthBearer:
		var cfg model.MCPAuthConfigBearer
		if json.Unmarshal(svc.AuthConfig, &cfg) == nil {
			return cfg.Token, ""
		}
	case model.MCPAuthHeader:
		var cfg model.MCPAuthConfigHeader
		if json.Unmarshal(svc.AuthConfig, &cfg) == nil {
			return cfg.HeaderValue, cfg.HeaderName
		}
	}
	return "", ""
}
