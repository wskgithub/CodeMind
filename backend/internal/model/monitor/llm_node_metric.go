package monitor

import "time"

// LLMNodeMetric LLM节点上报的指标数据
// 由各个 LLM 服务节点定期上报，用于监控分布式 LLM 资源状态
type LLMNodeMetric struct {
	ID               int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	NodeID           string    `gorm:"size:100;not null;index" json:"node_id"`              // 节点唯一标识
	NodeName         string    `gorm:"size:200;not null" json:"node_name"`                  // 节点显示名称
	Status           string    `gorm:"size:20;not null;default:unknown" json:"status"`      // 状态: online, offline, busy, error
	
	// GPU 信息（JSON 存储多卡信息）
	GPUInfo          string    `gorm:"type:text" json:"gpu_info"`                         // GPU 信息 JSON 数组
	GPUCount         int       `json:"gpu_count"`                                           // GPU 数量
	GPUTotalMemoryGB float64   `json:"gpu_total_memory_gb"`                                 // GPU 总显存 GB
	GPUUsedMemoryGB  float64   `json:"gpu_used_memory_gb"`                                  // GPU 已用显存 GB
	GPUUtilization   float64   `json:"gpu_utilization"`                                     // GPU 平均利用率
	
	// CPU 和内存
	CPUCores         int       `json:"cpu_cores"`                                           // CPU 核心数
	CPUUsagePercent  float64   `json:"cpu_usage_percent"`                                   // CPU 使用率
	MemoryTotalGB    float64   `json:"memory_total_gb"`                                     // 总内存 GB
	MemoryUsedGB     float64   `json:"memory_used_gb"`                                      // 已用内存 GB
	
	// 请求处理统计（最近一分钟）
	RequestsPerMin   int       `json:"requests_per_min"`                                    // 每分钟请求数
	AvgResponseTimeMs float64  `json:"avg_response_time_ms"`                                // 平均响应时间 ms
	ActiveRequests   int       `json:"active_requests"`                                     // 当前活跃请求数
	QueuedRequests   int       `json:"queued_requests"`                                     // 队列中请求数
	
	// 模型信息
	LoadedModels     string    `gorm:"type:text" json:"loaded_models"`                    // 已加载模型 JSON
	ModelCount       int       `json:"model_count"`                                         // 加载的模型数量
	
	// 附加信息
	Version          string    `gorm:"size:50" json:"version"`                              // LLM 服务版本
	Labels           string    `gorm:"size:500" json:"labels"`                              // 标签（JSON）
	ReportedAt       time.Time `gorm:"not null;index" json:"reported_at"`                   // 节点上报时间
	CreatedAt        time.Time `gorm:"not null;autoCreateTime" json:"created_at"`           // 记录创建时间
}

// TableName 指定表名
func (LLMNodeMetric) TableName() string {
	return "llm_node_metrics"
}

// LLMNodeStatus 节点状态常量
const (
	NodeStatusOnline  = "online"   // 在线正常运行
	NodeStatusOffline = "offline"  // 离线
	NodeStatusBusy    = "busy"     // 忙碌（高负载）
	NodeStatusError   = "error"    // 错误状态
	NodeStatusIdle    = "idle"     // 空闲
)

// GPUInfo GPU 信息结构
type GPUInfo struct {
	Index        int     `json:"index"`          // GPU 索引
	Name         string  `json:"name"`           // GPU 型号
	TotalMemoryGB float64 `json:"total_memory_gb"` // 总显存
	UsedMemoryGB  float64 `json:"used_memory_gb"`  // 已用显存
	Temperature  int     `json:"temperature"`    // 温度
	Utilization  float64 `json:"utilization"`    // 利用率
}

// LoadedModelInfo 已加载模型信息
type LoadedModelInfo struct {
	ModelID      string  `json:"model_id"`       // 模型 ID
	ModelName    string  `json:"model_name"`     // 模型名称
	LoadedAt     time.Time `json:"loaded_at"`    // 加载时间
	RequestCount int     `json:"request_count"`  // 请求计数
}

// LLMNodeSummary LLM节点汇总信息
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
