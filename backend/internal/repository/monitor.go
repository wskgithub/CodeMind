package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"codemind/internal/model/monitor"

	"gorm.io/gorm"
)

// MonitorRepository handles monitor data access.
type MonitorRepository struct {
	db *gorm.DB
}

// NewMonitorRepository creates a new monitor repository.
func NewMonitorRepository(db *gorm.DB) *MonitorRepository {
	return &MonitorRepository{db: db}
}

// CreateSystemMetric creates a system metric record.
func (r *MonitorRepository) CreateSystemMetric(ctx context.Context, metric *monitor.SystemMetric) error {
	return r.db.WithContext(ctx).Create(metric).Error
}

// CreateSystemMetrics batch creates system metric records.
func (r *MonitorRepository) CreateSystemMetrics(ctx context.Context, metrics []*monitor.SystemMetric) error {
	return r.db.WithContext(ctx).CreateInBatches(metrics, 100).Error //nolint:mnd // batch size
}

// GetLatestSystemMetrics returns the latest system metrics.
func (r *MonitorRepository) GetLatestSystemMetrics(ctx context.Context, hostname string, limit int) ([]*monitor.SystemMetric, error) {
	var metrics []*monitor.SystemMetric
	err := r.db.WithContext(ctx).
		Where("host_name = ?", hostname).
		Order("created_at DESC").
		Limit(limit).
		Find(&metrics).Error
	return metrics, err
}

// GetSystemMetricsByTimeRange returns system metrics by time range.
func (r *MonitorRepository) GetSystemMetricsByTimeRange(ctx context.Context, hostname, metricType string, start, end time.Time) ([]*monitor.SystemMetric, error) {
	var metrics []*monitor.SystemMetric
	query := r.db.WithContext(ctx).
		Where("host_name = ?", hostname).
		Where("created_at BETWEEN ? AND ?", start, end)

	if metricType != "" {
		query = query.Where("metric_type = ?", metricType)
	}

	err := query.Order("created_at ASC").Find(&metrics).Error
	return metrics, err
}

// GetSystemMetricsSummary returns system metrics summary for dashboard.
//
//nolint:gocyclo // complex dashboard aggregation logic
func (r *MonitorRepository) GetSystemMetricsSummary(ctx context.Context, hostname string) (*monitor.SystemMetricsSummary, error) {
	var cpuUsage monitor.SystemMetric
	err := r.db.WithContext(ctx).
		Where("host_name = ? AND metric_type = ? AND metric_name = ?", hostname, monitor.MetricTypeCPU, "usage_percent").
		Order("created_at DESC").
		First(&cpuUsage).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	var memTotal, memUsed, memUsage monitor.SystemMetric
	r.db.WithContext(ctx).
		Where("host_name = ? AND metric_type = ? AND metric_name = ?", hostname, monitor.MetricTypeMemory, "total_gb").
		Order("created_at DESC").
		First(&memTotal)
	r.db.WithContext(ctx).
		Where("host_name = ? AND metric_type = ? AND metric_name = ?", hostname, monitor.MetricTypeMemory, "used_gb").
		Order("created_at DESC").
		First(&memUsed)
	r.db.WithContext(ctx).
		Where("host_name = ? AND metric_type = ? AND metric_name = ?", hostname, monitor.MetricTypeMemory, "usage_percent").
		Order("created_at DESC").
		First(&memUsage)

	var diskMetrics []*monitor.SystemMetric
	r.db.WithContext(ctx).
		Where("host_name = ? AND metric_type = ?", hostname, monitor.MetricTypeDisk).
		Order("created_at DESC").
		Find(&diskMetrics)

	diskMap := make(map[string]*monitor.DiskMetrics)
	for _, m := range diskMetrics {
		var labels map[string]string
		_ = json.Unmarshal([]byte(m.Labels), &labels)
		mountPoint := labels["mount_point"]
		if mountPoint == "" {
			mountPoint = "unknown"
		}

		if _, ok := diskMap[mountPoint]; !ok {
			diskMap[mountPoint] = &monitor.DiskMetrics{MountPoint: mountPoint}
		}

		switch m.MetricName {
		case "total_gb":
			diskMap[mountPoint].TotalGB = m.Value
		case "used_gb":
			diskMap[mountPoint].UsedGB = m.Value
		case "usage_percent":
			diskMap[mountPoint].UsagePercent = m.Value
		}
	}

	var load1, load5, load15 monitor.SystemMetric
	r.db.WithContext(ctx).
		Where("host_name = ? AND metric_type = ? AND metric_name = ?", hostname, monitor.MetricTypeLoad, "load_1").
		Order("created_at DESC").
		First(&load1)
	r.db.WithContext(ctx).
		Where("host_name = ? AND metric_type = ? AND metric_name = ?", hostname, monitor.MetricTypeLoad, "load_5").
		Order("created_at DESC").
		First(&load5)
	r.db.WithContext(ctx).
		Where("host_name = ? AND metric_type = ? AND metric_name = ?", hostname, monitor.MetricTypeLoad, "load_15").
		Order("created_at DESC").
		First(&load15)

	summary := &monitor.SystemMetricsSummary{
		RecordedAt: time.Now(),
	}

	if cpuUsage.ID > 0 {
		summary.CPUUsage = &monitor.CPUMetrics{
			UsagePercent: cpuUsage.Value,
		}
		var cpuLabels map[string]string
		_ = json.Unmarshal([]byte(cpuUsage.Labels), &cpuLabels)
		if cores, ok := cpuLabels["core_count"]; ok {
			_, _ = fmt.Sscanf(cores, "%d", &summary.CPUUsage.CoreCount)
		}
		summary.CPUUsage.ModelName = cpuLabels["model_name"]
	}

	if memTotal.ID > 0 || memUsage.ID > 0 {
		summary.MemoryUsage = &monitor.MemoryMetrics{
			TotalGB:      memTotal.Value,
			UsedGB:       memUsed.Value,
			UsagePercent: memUsage.Value,
		}
		if summary.MemoryUsage.TotalGB > summary.MemoryUsage.UsedGB {
			summary.MemoryUsage.FreeGB = summary.MemoryUsage.TotalGB - summary.MemoryUsage.UsedGB
		}
	}

	if len(diskMap) > 0 {
		for _, d := range diskMap {
			d.FreeGB = d.TotalGB - d.UsedGB
			summary.DiskUsage = append(summary.DiskUsage, *d)
		}
		sort.Slice(summary.DiskUsage, func(i, j int) bool {
			return summary.DiskUsage[i].MountPoint < summary.DiskUsage[j].MountPoint
		})
	}

	if load1.ID > 0 || load5.ID > 0 || load15.ID > 0 {
		summary.LoadAverage = &monitor.LoadMetrics{
			Load1:  load1.Value,
			Load5:  load5.Value,
			Load15: load15.Value,
		}
	}

	return summary, nil
}

// CleanupOldSystemMetrics removes old system metrics data.
func (r *MonitorRepository) CleanupOldSystemMetrics(ctx context.Context, retentionDays int) (int64, error) {
	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	result := r.db.WithContext(ctx).Where("created_at < ?", cutoff).Delete(&monitor.SystemMetric{})
	return result.RowsAffected, result.Error
}

// CreateLLMNodeMetric creates an LLM node metric record.
func (r *MonitorRepository) CreateLLMNodeMetric(ctx context.Context, metric *monitor.LLMNodeMetric) error {
	return r.db.WithContext(ctx).Create(metric).Error
}

// GetLatestLLMNodeMetrics returns the latest metrics for all nodes.
func (r *MonitorRepository) GetLatestLLMNodeMetrics(ctx context.Context) ([]*monitor.LLMNodeMetric, error) {
	subQuery := r.db.Model(&monitor.LLMNodeMetric{}).
		Select("node_id, MAX(created_at) as max_created_at").
		Group("node_id")

	var metrics []*monitor.LLMNodeMetric
	err := r.db.WithContext(ctx).
		Joins("JOIN (?) AS latest ON llm_node_metrics.node_id = latest.node_id AND llm_node_metrics.created_at = latest.max_created_at", subQuery).
		Find(&metrics).Error
	return metrics, err
}

// GetLLMNodeMetricsByNodeID returns metrics history for a specific node.
func (r *MonitorRepository) GetLLMNodeMetricsByNodeID(ctx context.Context, nodeID string, limit int) ([]*monitor.LLMNodeMetric, error) {
	var metrics []*monitor.LLMNodeMetric
	err := r.db.WithContext(ctx).
		Where("node_id = ?", nodeID).
		Order("created_at DESC").
		Limit(limit).
		Find(&metrics).Error
	return metrics, err
}

// GetLLMNodeMetricSummary returns LLM node summary info.
func (r *MonitorRepository) GetLLMNodeMetricSummary(ctx context.Context) ([]monitor.LLMNodeSummary, error) {
	metrics, err := r.GetLatestLLMNodeMetrics(ctx)
	if err != nil {
		return nil, err
	}

	summaries := make([]monitor.LLMNodeSummary, 0, len(metrics))
	for _, m := range metrics {
		summary := monitor.LLMNodeSummary{
			NodeID:            m.NodeID,
			NodeName:          m.NodeName,
			Status:            m.Status,
			GPUUtilization:    m.GPUUtilization,
			GPUTotalMemoryGB:  m.GPUTotalMemoryGB,
			GPUUsedMemoryGB:   m.GPUUsedMemoryGB,
			ActiveRequests:    m.ActiveRequests,
			RequestsPerMin:    m.RequestsPerMin,
			AvgResponseTimeMs: m.AvgResponseTimeMs,
			ModelCount:        m.ModelCount,
			Version:           m.Version,
			LastSeenAt:        m.ReportedAt,
		}

		if m.MemoryTotalGB > 0 {
			summary.MemoryUsagePercent = (m.MemoryUsedGB / m.MemoryTotalGB) * 100 //nolint:mnd // percentage calculation
		}
		summary.CPUUsagePercent = m.CPUUsagePercent

		var models []monitor.LoadedModelInfo
		_ = json.Unmarshal([]byte(m.LoadedModels), &models)
		for _, model := range models {
			summary.LoadedModels = append(summary.LoadedModels, model.ModelName)
		}

		summaries = append(summaries, summary)
	}

	return summaries, nil
}

// GetActiveNodeCount returns active node count (reported in last 5 minutes).
func (r *MonitorRepository) GetActiveNodeCount(ctx context.Context) (int64, error) {
	cutoff := time.Now().Add(-5 * time.Minute)
	var count int64
	err := r.db.WithContext(ctx).
		Model(&monitor.LLMNodeMetric{}).
		Where("reported_at > ?", cutoff).
		Distinct("node_id").
		Count(&count).Error
	return count, err
}

// GetTotalNodeCount returns total node count.
func (r *MonitorRepository) GetTotalNodeCount(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&monitor.LLMNodeMetric{}).
		Distinct("node_id").
		Count(&count).Error
	return count, err
}

// CleanupOldLLMNodeMetrics removes old LLM node metrics data.
func (r *MonitorRepository) CleanupOldLLMNodeMetrics(ctx context.Context, retentionHours int) (int64, error) {
	cutoff := time.Now().Add(-time.Duration(retentionHours) * time.Hour)
	result := r.db.WithContext(ctx).Where("created_at < ?", cutoff).Delete(&monitor.LLMNodeMetric{})
	return result.RowsAffected, result.Error
}

// UpdateOrCreateLLMNodeMetric updates or creates an LLM node metric.
func (r *MonitorRepository) UpdateOrCreateLLMNodeMetric(ctx context.Context, metric *monitor.LLMNodeMetric) error {
	var existing monitor.LLMNodeMetric
	err := r.db.WithContext(ctx).
		Where("node_id = ?", metric.NodeID).
		Order("created_at DESC").
		First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		return r.db.WithContext(ctx).Create(metric).Error
	}

	if err != nil {
		return err
	}

	return r.db.WithContext(ctx).Create(metric).Error
}
