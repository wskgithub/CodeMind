// Package monitor 监控相关模型
package monitor

import "time"

// SystemMetric 系统资源指标
// 用于存储服务器 CPU、内存、磁盘等资源使用情况
type SystemMetric struct {
	ID         int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	HostName   string    `gorm:"size:100;not null;index" json:"host_name"`              // 服务器主机名
	MetricType string    `gorm:"size:50;not null;index" json:"metric_type"`             // 指标类型: cpu, memory, disk, network
	MetricName string    `gorm:"size:100;not null" json:"metric_name"`                  // 指标名称: usage_percent, used_gb, total_gb, etc.
	Value      float64   `gorm:"not null" json:"value"`                                 // 指标数值
	Labels     string    `gorm:"size:500;default:''" json:"labels"`                     // 额外标签 JSON，如磁盘分区名、网卡名等
	CreatedAt  time.Time `gorm:"not null;autoCreateTime;index" json:"created_at"`       // 记录时间
}

// TableName 指定表名
func (SystemMetric) TableName() string {
	return "system_metrics"
}

// MetricType 常量定义
const (
	MetricTypeCPU     = "cpu"
	MetricTypeMemory  = "memory"
	MetricTypeDisk    = "disk"
	MetricTypeNetwork = "network"
	MetricTypeLoad    = "load"
)

// SystemMetricsSummary 系统指标汇总（用于实时展示）
type SystemMetricsSummary struct {
	CPUUsage     *CPUMetrics     `json:"cpu_usage"`     // CPU 使用情况
	MemoryUsage  *MemoryMetrics  `json:"memory_usage"`  // 内存使用情况
	DiskUsage    []DiskMetrics   `json:"disk_usage"`    // 磁盘使用情况
	NetworkIO    *NetworkMetrics `json:"network_io"`    // 网络 IO
	LoadAverage  *LoadMetrics    `json:"load_average"`  // 系统负载
	RecordedAt   time.Time       `json:"recorded_at"`   // 记录时间
}

// CPUMetrics CPU 指标
type CPUMetrics struct {
	UsagePercent float64 `json:"usage_percent"`       // CPU 使用率百分比
	CoreCount    int     `json:"core_count"`          // CPU 核心数
	ModelName    string  `json:"model_name"`          // CPU 型号
}

// MemoryMetrics 内存指标
type MemoryMetrics struct {
	TotalGB      float64 `json:"total_gb"`            // 总内存 GB
	UsedGB       float64 `json:"used_gb"`             // 已用内存 GB
	FreeGB       float64 `json:"free_gb"`             // 空闲内存 GB
	UsagePercent float64 `json:"usage_percent"`       // 使用率百分比
}

// DiskMetrics 磁盘指标
type DiskMetrics struct {
	MountPoint   string  `json:"mount_point"`         // 挂载点
	Device       string  `json:"device"`              // 设备名
	TotalGB      float64 `json:"total_gb"`            // 总容量 GB
	UsedGB       float64 `json:"used_gb"`             // 已用容量 GB
	FreeGB       float64 `json:"free_gb"`             // 空闲容量 GB
	UsagePercent float64 `json:"usage_percent"`       // 使用率百分比
}

// NetworkMetrics 网络指标
type NetworkMetrics struct {
	InterfaceName string  `json:"interface_name"`     // 网卡名
	BytesSentMB   float64 `json:"bytes_sent_mb"`      // 发送 MB
	BytesRecvMB   float64 `json:"bytes_recv_mb"`      // 接收 MB
	PacketsSent   uint64  `json:"packets_sent"`       // 发送包数
	PacketsRecv   uint64  `json:"packets_recv"`       // 接收包数
}

// LoadMetrics 系统负载指标
type LoadMetrics struct {
	Load1  float64 `json:"load_1"`                  // 1分钟负载
	Load5  float64 `json:"load_5"`                  // 5分钟负载
	Load15 float64 `json:"load_15"`                 // 15分钟负载
}
