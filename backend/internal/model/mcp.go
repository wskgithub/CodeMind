package model

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// MCPService represents an MCP service registration.
type MCPService struct {
	ID            int64           `gorm:"primaryKey;autoIncrement" json:"id"`
	Name          string          `gorm:"size:100;not null;uniqueIndex" json:"name"`
	DisplayName   string          `gorm:"size:200;not null" json:"display_name"`
	Description   string          `gorm:"type:text" json:"description"`
	EndpointURL   string          `gorm:"size:500;not null" json:"endpoint_url"`
	TransportType string          `gorm:"size:20;not null;default:sse" json:"transport_type"`
	Status        string          `gorm:"size:20;not null;default:enabled" json:"status"`
	AuthType      string          `gorm:"size:20;not null;default:none" json:"auth_type"`
	AuthConfig    json.RawMessage `gorm:"type:jsonb" json:"auth_config,omitempty"`
	ToolsSchema   json.RawMessage `gorm:"type:jsonb" json:"tools_schema,omitempty"`
	CreatedAt     time.Time       `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time       `gorm:"not null;autoUpdateTime" json:"updated_at"`
	DeletedAt     gorm.DeletedAt  `gorm:"index" json:"-"`

	AccessRules []MCPAccessRule `gorm:"foreignKey:ServiceID" json:"access_rules,omitempty"`
}

// TableName returns the table name.
func (MCPService) TableName() string {
	return "mcp_services"
}

const (
	MCPServiceEnabled  = "enabled"
	MCPServiceDisabled = "disabled"
)

const (
	MCPTransportSSE            = "sse"
	MCPTransportStreamableHTTP = "streamable-http"
)

const (
	MCPAuthNone   = "none"
	MCPAuthBearer = "bearer"
	MCPAuthHeader = "header"
)

// MCPAuthConfigBearer represents bearer auth configuration.
type MCPAuthConfigBearer struct {
	Token string `json:"token"`
}

// MCPAuthConfigHeader represents custom header auth configuration.
type MCPAuthConfigHeader struct {
	HeaderName  string `json:"header_name"`
	HeaderValue string `json:"header_value"`
}

// MCPAccessRule represents an MCP service access control rule.
type MCPAccessRule struct {
	ID         int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ServiceID  int64     `gorm:"not null;index" json:"service_id"`
	TargetType string    `gorm:"size:20;not null" json:"target_type"`
	TargetID   int64     `gorm:"not null" json:"target_id"`
	Allowed    bool      `gorm:"not null;default:true" json:"allowed"`
	CreatedAt  time.Time `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time `gorm:"not null;autoUpdateTime" json:"updated_at"`

	Service *MCPService `gorm:"foreignKey:ServiceID" json:"service,omitempty"`
}

// TableName returns the table name.
func (MCPAccessRule) TableName() string {
	return "mcp_access_rules"
}

const (
	MCPTargetUser       = "user"
	MCPTargetDepartment = "department"
	MCPTargetRole       = "role"
)
