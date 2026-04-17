package monitor

import "time"

// LLMNodeMetric stores metrics reported by LLM nodes.
type LLMNodeMetric struct {
	CreatedAt         time.Time `gorm:"not null;autoCreateTime" json:"created_at"`
	ReportedAt        time.Time `gorm:"not null;index" json:"reported_at"`
	LoadedModels      string    `gorm:"type:text" json:"loaded_models"`
	NodeID            string    `gorm:"size:100;not null;index" json:"node_id"`
	NodeName          string    `gorm:"size:200;not null" json:"node_name"`
	Status            string    `gorm:"size:20;not null;default:unknown" json:"status"`
	GPUInfo           string    `gorm:"type:text" json:"gpu_info"`
	Labels            string    `gorm:"size:500" json:"labels"`
	Version           string    `gorm:"size:50" json:"version"`
	CPUUsagePercent   float64   `json:"cpu_usage_percent"`
	QueuedRequests    int       `json:"queued_requests"`
	MemoryTotalGB     float64   `json:"memory_total_gb"`
	MemoryUsedGB      float64   `json:"memory_used_gb"`
	RequestsPerMin    int       `json:"requests_per_min"`
	AvgResponseTimeMs float64   `json:"avg_response_time_ms"`
	ActiveRequests    int       `json:"active_requests"`
	ID                int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	CPUCores          int       `json:"cpu_cores"`
	ModelCount        int       `json:"model_count"`
	GPUUtilization    float64   `json:"gpu_utilization"`
	GPUUsedMemoryGB   float64   `json:"gpu_used_memory_gb"`
	GPUTotalMemoryGB  float64   `json:"gpu_total_memory_gb"`
	GPUCount          int       `json:"gpu_count"`
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
	Name          string  `json:"name"`
	Index         int     `json:"index"`
	TotalMemoryGB float64 `json:"total_memory_gb"`
	UsedMemoryGB  float64 `json:"used_memory_gb"`
	Temperature   int     `json:"temperature"`
	Utilization   float64 `json:"utilization"`
}

// LoadedModelInfo represents loaded model info.
type LoadedModelInfo struct {
	LoadedAt     time.Time `json:"loaded_at"`
	ModelID      string    `json:"model_id"`
	ModelName    string    `json:"model_name"`
	RequestCount int       `json:"request_count"`
}

// LLMNodeSummary represents LLM node summary info.
type LLMNodeSummary struct {
	LastSeenAt         time.Time `json:"last_seen_at"`
	NodeID             string    `json:"node_id"`
	NodeName           string    `json:"node_name"`
	Status             string    `json:"status"`
	Version            string    `json:"version"`
	LoadedModels       []string  `json:"loaded_models"`
	CPUUsagePercent    float64   `json:"cpu_usage_percent"`
	MemoryUsagePercent float64   `json:"memory_usage_percent"`
	RequestsPerMin     int       `json:"requests_per_min"`
	AvgResponseTimeMs  float64   `json:"avg_response_time_ms"`
	ActiveRequests     int       `json:"active_requests"`
	ModelCount         int       `json:"model_count"`
	GPUUsedMemoryGB    float64   `json:"gpu_used_memory_gb"`
	GPUTotalMemoryGB   float64   `json:"gpu_total_memory_gb"`
	GPUUtilization     float64   `json:"gpu_utilization"`
}
