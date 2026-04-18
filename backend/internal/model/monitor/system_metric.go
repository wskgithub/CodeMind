package monitor

import "time"

// SystemMetric stores server resource metrics.
type SystemMetric struct {
	CreatedAt  time.Time `gorm:"not null;autoCreateTime;index" json:"created_at"`
	HostName   string    `gorm:"size:100;not null;index" json:"host_name"`
	MetricType string    `gorm:"size:50;not null;index" json:"metric_type"`
	MetricName string    `gorm:"size:100;not null" json:"metric_name"`
	Labels     string    `gorm:"size:500;default:''" json:"labels"`
	ID         int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Value      float64   `gorm:"not null" json:"value"`
}

// TableName returns table name.
func (SystemMetric) TableName() string {
	return "system_metrics"
}

// System metric type constants.
const (
	MetricTypeCPU     = "cpu"
	MetricTypeMemory  = "memory"
	MetricTypeDisk    = "disk"
	MetricTypeNetwork = "network"
	MetricTypeLoad    = "load"
)

// SystemMetricsSummary represents system metrics summary.
type SystemMetricsSummary struct {
	RecordedAt  time.Time       `json:"recorded_at"`
	CPUUsage    *CPUMetrics     `json:"cpu_usage"`
	MemoryUsage *MemoryMetrics  `json:"memory_usage"`
	NetworkIO   *NetworkMetrics `json:"network_io"`
	LoadAverage *LoadMetrics    `json:"load_average"`
	DiskUsage   []DiskMetrics   `json:"disk_usage"`
}

// CPUMetrics represents CPU metrics.
type CPUMetrics struct {
	ModelName    string  `json:"model_name"`
	UsagePercent float64 `json:"usage_percent"`
	CoreCount    int     `json:"core_count"`
}

// MemoryMetrics represents memory metrics.
type MemoryMetrics struct {
	TotalGB      float64 `json:"total_gb"`
	UsedGB       float64 `json:"used_gb"`
	FreeGB       float64 `json:"free_gb"`
	UsagePercent float64 `json:"usage_percent"`
}

// DiskMetrics represents disk metrics.
type DiskMetrics struct {
	MountPoint   string  `json:"mount_point"`
	Device       string  `json:"device"`
	TotalGB      float64 `json:"total_gb"`
	UsedGB       float64 `json:"used_gb"`
	FreeGB       float64 `json:"free_gb"`
	UsagePercent float64 `json:"usage_percent"`
}

// NetworkMetrics represents network metrics.
type NetworkMetrics struct {
	InterfaceName string  `json:"interface_name"`
	BytesSentMB   float64 `json:"bytes_sent_mb"`
	BytesRecvMB   float64 `json:"bytes_recv_mb"`
	PacketsSent   uint64  `json:"packets_sent"`
	PacketsRecv   uint64  `json:"packets_recv"`
}

// LoadMetrics represents system load metrics.
type LoadMetrics struct {
	Load1  float64 `json:"load_1"`
	Load5  float64 `json:"load_5"`
	Load15 float64 `json:"load_15"`
}
