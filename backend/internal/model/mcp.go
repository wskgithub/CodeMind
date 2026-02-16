package model

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// ──────────────────────────────────
// MCP 服务管理模型
// ──────────────────────────────────

// MCPService MCP 服务注册信息
type MCPService struct {
	ID            int64           `gorm:"primaryKey;autoIncrement" json:"id"`
	Name          string          `gorm:"size:100;not null;uniqueIndex" json:"name"`            // 服务唯一标识
	DisplayName   string          `gorm:"size:200;not null" json:"display_name"`                // 显示名称
	Description   string          `gorm:"type:text" json:"description"`                         // 服务描述
	EndpointURL   string          `gorm:"size:500;not null" json:"endpoint_url"`                // 后端 MCP 服务地址
	TransportType string          `gorm:"size:20;not null;default:sse" json:"transport_type"`   // sse | streamable-http
	Status        string          `gorm:"size:20;not null;default:enabled" json:"status"`       // enabled | disabled
	AuthType      string          `gorm:"size:20;not null;default:none" json:"auth_type"`       // none | bearer | header
	AuthConfig    json.RawMessage `gorm:"type:jsonb" json:"auth_config,omitempty"`              // 认证配置
	ToolsSchema   json.RawMessage `gorm:"type:jsonb" json:"tools_schema,omitempty"`             // 工具列表缓存
	CreatedAt     time.Time       `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time       `gorm:"not null;autoUpdateTime" json:"updated_at"`
	DeletedAt     gorm.DeletedAt  `gorm:"index" json:"-"`

	// 关联
	AccessRules []MCPAccessRule `gorm:"foreignKey:ServiceID" json:"access_rules,omitempty"`
}

// TableName 自定义表名
func (MCPService) TableName() string {
	return "mcp_services"
}

// MCP 服务状态常量
const (
	MCPServiceEnabled  = "enabled"
	MCPServiceDisabled = "disabled"
)

// MCP 服务传输类型常量
const (
	MCPTransportSSE            = "sse"
	MCPTransportStreamableHTTP = "streamable-http"
)

// MCP 认证类型常量
const (
	MCPAuthNone   = "none"
	MCPAuthBearer = "bearer"
	MCPAuthHeader = "header"
)

// MCPAuthConfigBearer Bearer 认证配置
type MCPAuthConfigBearer struct {
	Token string `json:"token"`
}

// MCPAuthConfigHeader 自定义头认证配置
type MCPAuthConfigHeader struct {
	HeaderName  string `json:"header_name"`
	HeaderValue string `json:"header_value"`
}

// ──────────────────────────────────
// MCP 访问规则模型
// ──────────────────────────────────

// MCPAccessRule MCP 服务访问控制规则
type MCPAccessRule struct {
	ID         int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ServiceID  int64     `gorm:"not null;index" json:"service_id"`                      // 关联 MCP 服务
	TargetType string    `gorm:"size:20;not null" json:"target_type"`                   // user | department | role
	TargetID   int64     `gorm:"not null" json:"target_id"`                             // 目标 ID（role 时为 0）
	Allowed    bool      `gorm:"not null;default:true" json:"allowed"`                  // 是否允许
	CreatedAt  time.Time `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time `gorm:"not null;autoUpdateTime" json:"updated_at"`

	// 关联
	Service *MCPService `gorm:"foreignKey:ServiceID" json:"service,omitempty"`
}

// TableName 自定义表名
func (MCPAccessRule) TableName() string {
	return "mcp_access_rules"
}

// MCP 访问规则目标类型常量
const (
	MCPTargetUser       = "user"
	MCPTargetDepartment = "department"
	MCPTargetRole       = "role"
)
