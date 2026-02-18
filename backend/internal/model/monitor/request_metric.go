package monitor

import "time"

// RequestMetricsSummary 请求性能指标汇总
type RequestMetricsSummary struct {
	QPS             float64                 `json:"qps"`              // 当前 QPS
	AvgResponseTime float64                 `json:"avg_response_time"` // 平均响应时间 ms
	P95ResponseTime float64                 `json:"p95_response_time"` // P95 响应时间 ms
	P99ResponseTime float64                 `json:"p99_response_time"` // P99 响应时间 ms
	TotalRequests   int64                   `json:"total_requests"`   // 总请求数
	ErrorRate       float64                 `json:"error_rate"`       // 错误率
	StatusCodes     map[int]int64           `json:"status_codes"`     // HTTP 状态码分布
	TimeRange       TimeRange               `json:"time_range"`       // 时间范围
}

// TimeRange 时间范围
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// QPSDataPoint QPS 数据点（用于图表）
type QPSDataPoint struct {
	Timestamp int64   `json:"timestamp"`          // 时间戳
	Value     float64 `json:"value"`              // QPS 值
}

// ResponseTimeDataPoint 响应时间数据点
type ResponseTimeDataPoint struct {
	Timestamp int64   `json:"timestamp"`          // 时间戳
	Avg       float64 `json:"avg"`                // 平均值
	P95       float64 `json:"p95"`                // P95
	P99       float64 `json:"p99"`                // P99
}

// DashboardSummary 仪表盘汇总数据
type DashboardSummary struct {
	SystemStatus    *SystemMetricsSummary    `json:"system_status"`    // 系统状态
	RequestMetrics  *RequestMetricsSummary   `json:"request_metrics"`  // 请求指标
	LLMNodes        []LLMNodeSummary         `json:"llm_nodes"`        // LLM 节点列表
	ActiveNodes     int                      `json:"active_nodes"`     // 活跃节点数
	TotalNodes      int                      `json:"total_nodes""`     // 总节点数
	UpdatedAt       time.Time                `json:"updated_at"`       // 更新时间
}

// MetricQueryParams 指标查询参数
type MetricQueryParams struct {
	StartTime   time.Time `json:"start_time"`    // 开始时间
	EndTime     time.Time `json:"end_time"`      // 结束时间
	MetricType  string    `json:"metric_type"`   // 指标类型（可选）
	NodeID      string    `json:"node_id"`       // 节点 ID（可选）
	Interval    string    `json:"interval"`      // 聚合间隔: 1m, 5m, 1h
}

// NodeReportRequest LLM节点上报请求
type NodeReportRequest struct {
	NodeID            string            `json:"node_id" binding:"required"`            // 节点 ID
	NodeName          string            `json:"node_name"`                              // 节点名称
	Status            string            `json:"status"`                                 // 状态
	GPUInfo           []GPUInfo         `json:"gpu_info"`                               // GPU 信息
	GPUUtilization    float64           `json:"gpu_utilization"`                        // GPU 利用率
	CPUCores          int               `json:"cpu_cores"`                              // CPU 核心数
	CPUUsagePercent   float64           `json:"cpu_usage_percent"`                      // CPU 使用率
	MemoryTotalGB     float64           `json:"memory_total_gb"`                        // 总内存
	MemoryUsedGB      float64           `json:"memory_used_gb"`                         // 已用内存
	RequestsPerMin    int               `json:"requests_per_min"`                       // 每分钟请求数
	AvgResponseTimeMs float64           `json:"avg_response_time_ms"`                   // 平均响应时间
	ActiveRequests    int               `json:"active_requests"`                        // 活跃请求数
	QueuedRequests    int               `json:"queued_requests"`                        // 队列请求数
	LoadedModels      []LoadedModelInfo `json:"loaded_models"`                          // 已加载模型
	Version           string            `json:"version"`                                // 版本
	Labels            map[string]string `json:"labels"`                                 // 标签
	Timestamp         int64             `json:"timestamp"`                              // 上报时间戳
	APIKey            string            `json:"api_key"`                                // 认证密钥
}
