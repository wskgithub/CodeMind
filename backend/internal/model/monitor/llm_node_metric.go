package monitor

import "time"

// LLMNodeMetric stores metrics reported by LLM nodes.
type LLMNodeMetric struct {
	ID               int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	NodeID           string    `gorm:"size:100;not null;index" json:"node_id"`
	NodeName         string    `gorm:"size:200;not null" json:"node_name"`
	Status           string    `gorm:"size:20;not null;default:unknown" json:"status"`

	GPUInfo          string    `gorm:"type:text" json:"gpu_info"`
	GPUCount         int       `json:"gpu_count"`
	GPUTotalMemoryGB float64   `json:"gpu_total_memory_gb"`
	GPUUsedMemoryGB  float64   `json:"gpu_used_memory_gb"`
	GPUUtilization   float64   `json:"gpu_utilization"`

	CPUCores         int       `json:"cpu_cores"`
	CPUUsagePercent  float64   `json:"cpu_usage_percent"`
	MemoryTotalGB    float64   `json:"memory_total_gb"`
	MemoryUsedGB     float64   `json:"memory_used_gb"`

	RequestsPerMin    int      `json:"requests_per_min"`
	AvgResponseTimeMs float64  `json:"avg_response_time_ms"`
	ActiveRequests    int      `json:"active_requests"`
	QueuedRequests    int      `json:"queued_requests"`

	LoadedModels     string    `gorm:"type:text" json:"loaded_models"`
	ModelCount       int       `json:"model_count"`

	Version          string    `gorm:"size:50" json:"version"`
	Labels           string    `gorm:"size:500" json:"labels"`
	ReportedAt       time.Time `gorm:"not null;index" json:"reported_at"`
	CreatedAt        time.Time `gorm:"not null;autoCreateTime" json:"created_at"`
}

// TableName returns table name.
func (LLMNodeMetric) TableName() string {
	return "llm_node_metrics"
}

const (
	NodeStatusOnline  = "online"
	NodeStatusOffline = "offline"
	NodeStatusBusy    = "busy"
	NodeStatusError   = "error"
	NodeStatusIdle    = "idle"
)

// GPUInfo represents GPU info structure.
type GPUInfo struct {
	Index         int     `json:"index"`
	Name          string  `json:"name"`
	TotalMemoryGB float64 `json:"total_memory_gb"`
	UsedMemoryGB  float64 `json:"used_memory_gb"`
	Temperature   int     `json:"temperature"`
	Utilization   float64 `json:"utilization"`
}

// LoadedModelInfo represents loaded model info.
type LoadedModelInfo struct {
	ModelID      string    `json:"model_id"`
	ModelName    string    `json:"model_name"`
	LoadedAt     time.Time `json:"loaded_at"`
	RequestCount int       `json:"request_count"`
}

// LLMNodeSummary represents LLM node summary info.
type LLMNodeSummary struct {
	NodeID            string            `json:"node_id"`
	NodeName          string            `json:"node_name"`
	Status            string            `json:"status"`
	GPUUtilization    float64           `json:"gpu_utilization"`
	GPUTotalMemoryGB  float64           `json:"gpu_total_memory_gb"`
	GPUUsedMemoryGB   float64           `json:"gpu_used_memory_gb"`
	CPUUsagePercent   float64           `json:"cpu_usage_percent"`
	MemoryUsagePercent float64          `json:"memory_usage_percent"`
	RequestsPerMin    int               `json:"requests_per_min"`
	AvgResponseTimeMs float64           `json:"avg_response_time_ms"`
	ActiveRequests    int               `json:"active_requests"`
	ModelCount        int               `json:"model_count"`
	LoadedModels      []string          `json:"loaded_models"`
	Version           string            `json:"version"`
	LastSeenAt        time.Time         `json:"last_seen_at"`
}
