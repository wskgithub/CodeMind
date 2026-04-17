package monitor

import "time"

// RequestMetricsSummary represents request metrics summary.
type RequestMetricsSummary struct {
	QPS             float64       `json:"qps"`
	AvgResponseTime float64       `json:"avg_response_time"`
	P95ResponseTime float64       `json:"p95_response_time"`
	P99ResponseTime float64       `json:"p99_response_time"`
	TotalRequests   int64         `json:"total_requests"`
	ErrorRate       float64       `json:"error_rate"`
	StatusCodes     map[int]int64 `json:"status_codes"`
	TimeRange       TimeRange     `json:"time_range"`
}

// TimeRange represents time range.
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// QPSDataPoint represents QPS data point for charts.
type QPSDataPoint struct {
	Timestamp int64   `json:"timestamp"`
	Value     float64 `json:"value"`
}

// ResponseTimeDataPoint represents response time data point.
type ResponseTimeDataPoint struct {
	Timestamp int64   `json:"timestamp"`
	Avg       float64 `json:"avg"`
	P95       float64 `json:"p95"`
	P99       float64 `json:"p99"`
}

// DashboardSummary represents dashboard summary data.
type DashboardSummary struct {
	SystemStatus   *SystemMetricsSummary  `json:"system_status"`
	RequestMetrics *RequestMetricsSummary `json:"request_metrics"`
	LLMNodes       []LLMNodeSummary       `json:"llm_nodes"`
	ActiveNodes    int                    `json:"active_nodes"`
	TotalNodes     int                    `json:"total_nodes"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

// MetricQueryParams represents metric query parameters.
type MetricQueryParams struct {
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	MetricType string    `json:"metric_type"`
	NodeID     string    `json:"node_id"`
	Interval   string    `json:"interval"`
}

// NodeReportRequest represents LLM node report request.
type NodeReportRequest struct {
	NodeID            string            `json:"node_id" binding:"required"`
	NodeName          string            `json:"node_name"`
	Status            string            `json:"status"`
	GPUInfo           []GPUInfo         `json:"gpu_info"`
	GPUUtilization    float64           `json:"gpu_utilization"`
	CPUCores          int               `json:"cpu_cores"`
	CPUUsagePercent   float64           `json:"cpu_usage_percent"`
	MemoryTotalGB     float64           `json:"memory_total_gb"`
	MemoryUsedGB      float64           `json:"memory_used_gb"`
	RequestsPerMin    int               `json:"requests_per_min"`
	AvgResponseTimeMs float64           `json:"avg_response_time_ms"`
	ActiveRequests    int               `json:"active_requests"`
	QueuedRequests    int               `json:"queued_requests"`
	LoadedModels      []LoadedModelInfo `json:"loaded_models"`
	Version           string            `json:"version"`
	Labels            map[string]string `json:"labels"`
	Timestamp         int64             `json:"timestamp"`
	APIKey            string            `json:"api_key"`
}
