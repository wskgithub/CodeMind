// Package repository 监控数据访问层
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

// MonitorRepository 监控数据仓库
type MonitorRepository struct {
	db *gorm.DB
}

// NewMonitorRepository 创建监控仓库实例
func NewMonitorRepository(db *gorm.DB) *MonitorRepository {
	return &MonitorRepository{db: db}
}

// ==================== System Metrics ====================

// CreateSystemMetric 创建系统指标记录
func (r *MonitorRepository) CreateSystemMetric(ctx context.Context, metric *monitor.SystemMetric) error {
	return r.db.WithContext(ctx).Create(metric).Error
}

// CreateSystemMetrics 批量创建系统指标记录
func (r *MonitorRepository) CreateSystemMetrics(ctx context.Context, metrics []*monitor.SystemMetric) error {
	return r.db.WithContext(ctx).CreateInBatches(metrics, 100).Error
}

// GetLatestSystemMetrics 获取最新的系统指标
func (r *MonitorRepository) GetLatestSystemMetrics(ctx context.Context, hostname string, limit int) ([]*monitor.SystemMetric, error) {
	var metrics []*monitor.SystemMetric
	err := r.db.WithContext(ctx).
		Where("host_name = ?", hostname).
		Order("created_at DESC").
		Limit(limit).
		Find(&metrics).Error
	return metrics, err
}

// GetSystemMetricsByTimeRange 按时间范围查询系统指标
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

// GetSystemMetricsSummary 获取系统指标汇总（用于仪表盘）
func (r *MonitorRepository) GetSystemMetricsSummary(ctx context.Context, hostname string) (*monitor.SystemMetricsSummary, error) {
	// 获取最新的 CPU 指标
	var cpuUsage monitor.SystemMetric
	err := r.db.WithContext(ctx).
		Where("host_name = ? AND metric_type = ? AND metric_name = ?", hostname, monitor.MetricTypeCPU, "usage_percent").
		Order("created_at DESC").
		First(&cpuUsage).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// 获取内存指标
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

	// 获取磁盘指标
	var diskMetrics []*monitor.SystemMetric
	r.db.WithContext(ctx).
		Where("host_name = ? AND metric_type = ?", hostname, monitor.MetricTypeDisk).
		Order("created_at DESC").
		Find(&diskMetrics)

	diskMap := make(map[string]*monitor.DiskMetrics)
	for _, m := range diskMetrics {
		// 从 labels 解析挂载点
		var labels map[string]string
		json.Unmarshal([]byte(m.Labels), &labels)
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

	// 获取负载指标
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
		// 从 labels 解析核心数
		var cpuLabels map[string]string
		json.Unmarshal([]byte(cpuUsage.Labels), &cpuLabels)
		if cores, ok := cpuLabels["core_count"]; ok {
			fmt.Sscanf(cores, "%d", &summary.CPUUsage.CoreCount)
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
		// 按挂载点排序
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

// CleanupOldSystemMetrics 清理旧的系统指标数据（保留最近7天）
func (r *MonitorRepository) CleanupOldSystemMetrics(ctx context.Context, retentionDays int) (int64, error) {
	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	result := r.db.WithContext(ctx).Where("created_at < ?", cutoff).Delete(&monitor.SystemMetric{})
	return result.RowsAffected, result.Error
}

// ==================== LLM Node Metrics ====================

// CreateLLMNodeMetric 创建 LLM 节点指标记录
func (r *MonitorRepository) CreateLLMNodeMetric(ctx context.Context, metric *monitor.LLMNodeMetric) error {
	return r.db.WithContext(ctx).Create(metric).Error
}

// GetLatestLLMNodeMetrics 获取所有节点的最新指标
func (r *MonitorRepository) GetLatestLLMNodeMetrics(ctx context.Context) ([]*monitor.LLMNodeMetric, error) {
	// 使用子查询获取每个节点最新的一条记录
	subQuery := r.db.Model(&monitor.LLMNodeMetric{}).
		Select("node_id, MAX(created_at) as max_created_at").
		Group("node_id")

	var metrics []*monitor.LLMNodeMetric
	err := r.db.WithContext(ctx).
		Joins("JOIN (?) AS latest ON llm_node_metrics.node_id = latest.node_id AND llm_node_metrics.created_at = latest.max_created_at", subQuery).
		Find(&metrics).Error
	return metrics, err
}

// GetLLMNodeMetricsByNodeID 获取指定节点的指标历史
func (r *MonitorRepository) GetLLMNodeMetricsByNodeID(ctx context.Context, nodeID string, limit int) ([]*monitor.LLMNodeMetric, error) {
	var metrics []*monitor.LLMNodeMetric
	err := r.db.WithContext(ctx).
		Where("node_id = ?", nodeID).
		Order("created_at DESC").
		Limit(limit).
		Find(&metrics).Error
	return metrics, err
}

// GetLLMNodeMetricSummary 获取 LLM 节点汇总信息
func (r *MonitorRepository) GetLLMNodeMetricSummary(ctx context.Context) ([]monitor.LLMNodeSummary, error) {
	metrics, err := r.GetLatestLLMNodeMetrics(ctx)
	if err != nil {
		return nil, err
	}

	var summaries []monitor.LLMNodeSummary
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
			summary.MemoryUsagePercent = (m.MemoryUsedGB / m.MemoryTotalGB) * 100
		}
		summary.CPUUsagePercent = m.CPUUsagePercent

		// 解析已加载模型
		var models []monitor.LoadedModelInfo
		json.Unmarshal([]byte(m.LoadedModels), &models)
		for _, model := range models {
			summary.LoadedModels = append(summary.LoadedModels, model.ModelName)
		}

		summaries = append(summaries, summary)
	}

	return summaries, nil
}

// GetActiveNodeCount 获取活跃节点数量（最近5分钟内有上报）
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

// GetTotalNodeCount 获取总节点数量
func (r *MonitorRepository) GetTotalNodeCount(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&monitor.LLMNodeMetric{}).
		Distinct("node_id").
		Count(&count).Error
	return count, err
}

// CleanupOldLLMNodeMetrics 清理旧的 LLM 节点指标数据
func (r *MonitorRepository) CleanupOldLLMNodeMetrics(ctx context.Context, retentionHours int) (int64, error) {
	cutoff := time.Now().Add(-time.Duration(retentionHours) * time.Hour)
	result := r.db.WithContext(ctx).Where("created_at < ?", cutoff).Delete(&monitor.LLMNodeMetric{})
	return result.RowsAffected, result.Error
}

// UpdateOrCreateLLMNodeMetric 更新或创建 LLM 节点指标
func (r *MonitorRepository) UpdateOrCreateLLMNodeMetric(ctx context.Context, metric *monitor.LLMNodeMetric) error {
	// 先尝试查找是否存在该节点最近的记录
	var existing monitor.LLMNodeMetric
	err := r.db.WithContext(ctx).
		Where("node_id = ?", metric.NodeID).
		Order("created_at DESC").
		First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		// 不存在则创建
		return r.db.WithContext(ctx).Create(metric).Error
	}
	
	if err != nil {
		return err
	}

	// 存在则创建新记录（保留历史）
	return r.db.WithContext(ctx).Create(metric).Error
}
