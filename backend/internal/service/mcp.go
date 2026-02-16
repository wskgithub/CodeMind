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

// MCPService MCP 服务管理业务逻辑
type MCPService struct {
	mcpRepo  *repository.MCPRepository
	proxy    *mcpPkg.Proxy
	logger   *zap.Logger
}

// NewMCPService 创建 MCP 服务管理实例
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

// GetProxy 获取 MCP 代理
func (s *MCPService) GetProxy() *mcpPkg.Proxy {
	return s.proxy
}

// ──────────────────────────────────
// 服务管理
// ──────────────────────────────────

// CreateService 创建 MCP 服务
func (s *MCPService) CreateService(req *dto.CreateMCPServiceRequest) (*model.MCPService, error) {
	// 检查名称是否重复
	if existing, _ := s.mcpRepo.GetServiceByName(req.Name); existing != nil {
		return nil, fmt.Errorf("服务名称 '%s' 已存在", req.Name)
	}

	// 序列化认证配置
	var authConfig json.RawMessage
	if req.AuthConfig != nil {
		data, err := json.Marshal(req.AuthConfig)
		if err != nil {
			return nil, fmt.Errorf("认证配置格式错误: %w", err)
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
		return nil, fmt.Errorf("创建 MCP 服务失败: %w", err)
	}

	s.logger.Info("创建 MCP 服务", zap.String("name", svc.Name), zap.Int64("id", svc.ID))
	return svc, nil
}

// GetService 获取服务详情
func (s *MCPService) GetService(id int64) (*model.MCPService, error) {
	return s.mcpRepo.GetServiceByID(id)
}

// ListServices 获取服务列表
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

// UpdateService 更新服务信息
func (s *MCPService) UpdateService(id int64, req *dto.UpdateMCPServiceRequest) error {
	svc, err := s.mcpRepo.GetServiceByID(id)
	if err != nil {
		return fmt.Errorf("服务不存在")
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

// DeleteService 删除服务
func (s *MCPService) DeleteService(id int64) error {
	svc, err := s.mcpRepo.GetServiceByID(id)
	if err != nil {
		return fmt.Errorf("服务不存在")
	}

	// 断开连接
	s.proxy.DisconnectService(svc.Name)

	// 删除访问规则
	_ = s.mcpRepo.DeleteAccessRulesByService(id)

	// 软删除服务
	return s.mcpRepo.DeleteService(id)
}

// SyncTools 同步服务工具列表
func (s *MCPService) SyncTools(id int64) error {
	svc, err := s.mcpRepo.GetServiceByID(id)
	if err != nil {
		return fmt.Errorf("服务不存在")
	}

	// 确保连接
	if !s.proxy.IsConnected(svc.Name) {
		authToken, authHeader := s.extractAuth(svc)
		if err := s.proxy.ConnectService(svc.Name, svc.EndpointURL, svc.AuthType, authToken, authHeader); err != nil {
			return fmt.Errorf("连接服务失败: %w", err)
		}
	}

	// 请求工具列表
	req := &mcpPkg.JSONRPCRequest{
		JSONRPC: mcpPkg.JSONRPCVersion,
		ID:      "sync-tools",
		Method:  mcpPkg.MethodToolsList,
	}

	resp, err := s.proxy.ForwardRequest(svc.Name, req)
	if err != nil {
		return fmt.Errorf("获取工具列表失败: %w", err)
	}

	if resp.Error != nil {
		return fmt.Errorf("上游服务错误: %s", resp.Error.Message)
	}

	// 提取工具列表并保存
	var result mcpPkg.ToolsListResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return fmt.Errorf("解析工具列表失败: %w", err)
	}

	toolsData, _ := json.Marshal(result.Tools)
	return s.mcpRepo.UpdateToolsSchema(id, toolsData)
}

// GetServiceTools 获取服务的工具列表
func (s *MCPService) GetServiceTools(id int64) ([]dto.MCPToolInfo, error) {
	svc, err := s.mcpRepo.GetServiceByID(id)
	if err != nil {
		return nil, fmt.Errorf("服务不存在")
	}

	if svc.ToolsSchema == nil {
		return []dto.MCPToolInfo{}, nil
	}

	var tools []mcpPkg.Tool
	if err := json.Unmarshal(svc.ToolsSchema, &tools); err != nil {
		return nil, fmt.Errorf("工具列表解析失败: %w", err)
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

// ──────────────────────────────────
// 访问控制
// ──────────────────────────────────

// SetAccessRule 设置访问规则
func (s *MCPService) SetAccessRule(req *dto.SetMCPAccessRuleRequest) error {
	rule := &model.MCPAccessRule{
		ServiceID:  req.ServiceID,
		TargetType: req.TargetType,
		TargetID:   req.TargetID,
		Allowed:    req.Allowed,
	}
	return s.mcpRepo.UpsertAccessRule(rule)
}

// ListAccessRules 获取访问规则列表
func (s *MCPService) ListAccessRules(serviceID int64) ([]model.MCPAccessRule, error) {
	return s.mcpRepo.ListAccessRules(serviceID)
}

// DeleteAccessRule 删除访问规则
func (s *MCPService) DeleteAccessRule(id int64) error {
	return s.mcpRepo.DeleteAccessRule(id)
}

// CheckAccess 检查用户是否有权访问服务
func (s *MCPService) CheckAccess(serviceID, userID int64, deptID *int64, role string) bool {
	return s.mcpRepo.CheckAccess(serviceID, userID, deptID, role)
}

// ──────────────────────────────────
// 获取网关所需的服务信息
// ──────────────────────────────────

// GetServiceInfosForGateway 获取所有启用服务的信息（供 MCP 网关使用）
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

// FilterAccessibleServices 过滤用户有权限访问的服务
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

// extractAuth 从服务配置中提取认证信息
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
