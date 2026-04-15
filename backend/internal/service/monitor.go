// Package service 监控服务层
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"sync"
	"time"

	"codemind/internal/model/monitor"
	"codemind/internal/repository"

	"github.com/redis/go-redis/v9"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
	"go.uber.org/zap"
)

// MonitorService 监控服务
type MonitorService struct {
	monitorRepo  *repository.MonitorRepository
	usageRepo    *repository.UsageRepository
	backendRepo  *repository.LLMBackendRepository
	rdb          *redis.Client
	logger       *zap.Logger
	hostname     string
	
	// 性能统计数据
	requestStats *RequestStatsCollector
}

// RequestStatsCollector 请求统计收集器
type RequestStatsCollector struct {
	requestCount   int64
	responseTimes  []float64 // 响应时间列表（毫秒）
	statusCodes    map[int]int64
	lastResetTime  time.Time
	mutex          chan struct{}
}

// NewMonitorService 创建监控服务
func NewMonitorService(
	monitorRepo *repository.MonitorRepository,
	usageRepo *repository.UsageRepository,
	backendRepo *repository.LLMBackendRepository,
	rdb *redis.Client,
	logger *zap.Logger,
) *MonitorService {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "unknown"
	}

	svc := &MonitorService{
		monitorRepo:  monitorRepo,
		usageRepo:    usageRepo,
		backendRepo:  backendRepo,
		rdb:          rdb,
		logger:       logger,
		hostname:     hostname,
		requestStats: &RequestStatsCollector{
			responseTimes: make([]float64, 0, 1000),
			statusCodes:   make(map[int]int64),
			lastResetTime: time.Now(),
			mutex:         make(chan struct{}, 1),
		},
	}

	// 启动后台收集任务
	go svc.startCollector()
	
	return svc
}

// ==================== 系统指标收集 ====================

// startCollector 启动后台指标收集
func (s *MonitorService) startCollector() {
	ticker := time.NewTicker(30 * time.Second) // 每30秒收集一次
	defer ticker.Stop()

	// 立即执行一次
	s.collectSystemMetrics()

	for range ticker.C {
		s.collectSystemMetrics()
	}
}

// collectSystemMetrics 收集系统指标
func (s *MonitorService) collectSystemMetrics() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	metrics := make([]*monitor.SystemMetric, 0)
	now := time.Now()

	// 1. 收集 CPU 信息
	if info, err := cpu.Info(); err == nil && len(info) > 0 {
		modelName := info[0].ModelName
		cores := len(info)
		
		if percent, err := cpu.Percent(1*time.Second, false); err == nil && len(percent) > 0 {
			labels, _ := json.Marshal(map[string]string{
				"model_name":  modelName,
				"core_count":  fmt.Sprintf("%d", cores),
			})
			metrics = append(metrics, &monitor.SystemMetric{
				HostName:   s.hostname,
				MetricType: monitor.MetricTypeCPU,
				MetricName: "usage_percent",
				Value:      percent[0],
				Labels:     string(labels),
				CreatedAt:  now,
			})
		}
	}

	// 2. 收集内存信息
	if memInfo, err := mem.VirtualMemory(); err == nil {
		totalGB := float64(memInfo.Total) / 1024 / 1024 / 1024
		usedGB := float64(memInfo.Used) / 1024 / 1024 / 1024
		
		metrics = append(metrics,
			&monitor.SystemMetric{
				HostName:   s.hostname,
				MetricType: monitor.MetricTypeMemory,
				MetricName: "total_gb",
				Value:      totalGB,
				CreatedAt:  now,
			},
			&monitor.SystemMetric{
				HostName:   s.hostname,
				MetricType: monitor.MetricTypeMemory,
				MetricName: "used_gb",
				Value:      usedGB,
				CreatedAt:  now,
			},
			&monitor.SystemMetric{
				HostName:   s.hostname,
				MetricType: monitor.MetricTypeMemory,
				MetricName: "usage_percent",
				Value:      memInfo.UsedPercent,
				CreatedAt:  now,
			},
		)
	}

	// 3. 收集磁盘信息
	if partitions, err := disk.Partitions(false); err == nil {
		for _, part := range partitions {
			// 跳过虚拟文件系统
			if part.Fstype == "tmpfs" || part.Fstype == "devtmpfs" || part.Fstype == "squashfs" {
				continue
			}
			
			if usage, err := disk.Usage(part.Mountpoint); err == nil {
				labels, _ := json.Marshal(map[string]string{
					"mount_point": part.Mountpoint,
					"device":      part.Device,
					"fstype":      part.Fstype,
				})
				
				totalGB := float64(usage.Total) / 1024 / 1024 / 1024
				usedGB := float64(usage.Used) / 1024 / 1024 / 1024
				
				metrics = append(metrics,
					&monitor.SystemMetric{
						HostName:   s.hostname,
						MetricType: monitor.MetricTypeDisk,
						MetricName: "total_gb",
						Value:      totalGB,
						Labels:     string(labels),
						CreatedAt:  now,
					},
					&monitor.SystemMetric{
						HostName:   s.hostname,
						MetricType: monitor.MetricTypeDisk,
						MetricName: "used_gb",
						Value:      usedGB,
						Labels:     string(labels),
						CreatedAt:  now,
					},
					&monitor.SystemMetric{
						HostName:   s.hostname,
						MetricType: monitor.MetricTypeDisk,
						MetricName: "usage_percent",
						Value:      usage.UsedPercent,
						Labels:     string(labels),
						CreatedAt:  now,
					},
				)
			}
		}
	}

	// 4. 收集系统负载
	if loadAvg, err := load.Avg(); err == nil {
		metrics = append(metrics,
			&monitor.SystemMetric{
				HostName:   s.hostname,
				MetricType: monitor.MetricTypeLoad,
				MetricName: "load_1",
				Value:      loadAvg.Load1,
				CreatedAt:  now,
			},
			&monitor.SystemMetric{
				HostName:   s.hostname,
				MetricType: monitor.MetricTypeLoad,
				MetricName: "load_5",
				Value:      loadAvg.Load5,
				CreatedAt:  now,
			},
			&monitor.SystemMetric{
				HostName:   s.hostname,
				MetricType: monitor.MetricTypeLoad,
				MetricName: "load_15",
				Value:      loadAvg.Load15,
				CreatedAt:  now,
			},
		)
	}

	// 批量保存指标
	if len(metrics) > 0 {
		if err := s.monitorRepo.CreateSystemMetrics(ctx, metrics); err != nil {
			s.logger.Error("保存系统指标失败", zap.Error(err))
		} else {
			s.logger.Debug("系统指标收集完成", zap.Int("count", len(metrics)))
		}
	}

	// 清理旧数据（保留7天）
	go s.cleanupOldMetrics()
}

// cleanupOldMetrics 清理旧数据
func (s *MonitorService) cleanupOldMetrics() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if rows, err := s.monitorRepo.CleanupOldSystemMetrics(ctx, 7); err != nil {
		s.logger.Warn("清理系统指标旧数据失败", zap.Error(err))
	} else if rows > 0 {
		s.logger.Info("清理系统指标旧数据", zap.Int64("rows", rows))
	}

	if rows, err := s.monitorRepo.CleanupOldLLMNodeMetrics(ctx, 48); err != nil {
		s.logger.Warn("清理 LLM 节点指标旧数据失败", zap.Error(err))
	} else if rows > 0 {
		s.logger.Info("清理 LLM 节点指标旧数据", zap.Int64("rows", rows))
	}
}

// ==================== 请求统计 ====================

// RecordRequest 记录请求统计数据
func (s *MonitorService) RecordRequest(statusCode int, responseTimeMs float64) {
	select {
	case s.requestStats.mutex <- struct{}{}:
		s.requestStats.requestCount++
		s.requestStats.statusCodes[statusCode]++
		s.requestStats.responseTimes = append(s.requestStats.responseTimes, responseTimeMs)
		
		// 限制数组大小，防止内存无限增长
		if len(s.requestStats.responseTimes) > 10000 {
			s.requestStats.responseTimes = s.requestStats.responseTimes[5000:]
		}
		<-s.requestStats.mutex
	default:
		// 如果获取不到锁，则跳过（不阻塞主流程）
	}
}

// GetRequestMetrics 获取请求性能指标
func (s *MonitorService) GetRequestMetrics(ctx context.Context, duration time.Duration) (*monitor.RequestMetricsSummary, error) {
	s.requestStats.mutex <- struct{}{}
	defer func() { <-s.requestStats.mutex }()

	summary := &monitor.RequestMetricsSummary{
		StatusCodes: make(map[int]int64),
		TimeRange: monitor.TimeRange{
			Start: s.requestStats.lastResetTime,
			End:   time.Now(),
		},
	}

	// 复制状态码统计
	for code, count := range s.requestStats.statusCodes {
		summary.StatusCodes[code] = count
		summary.TotalRequests += count
		if code >= 400 {
			summary.ErrorRate += float64(count)
		}
	}

	if summary.TotalRequests > 0 {
		summary.ErrorRate = (summary.ErrorRate / float64(summary.TotalRequests)) * 100
	}

	// 计算响应时间统计
	times := s.requestStats.responseTimes
	if len(times) > 0 {
		var total float64
		for _, t := range times {
			total += t
		}
		summary.AvgResponseTime = total / float64(len(times))
		summary.P95ResponseTime = calculatePercentile(times, 0.95)
		summary.P99ResponseTime = calculatePercentile(times, 0.99)
	}

	// 计算 QPS
	elapsed := time.Since(s.requestStats.lastResetTime).Seconds()
	if elapsed > 0 {
		summary.QPS = float64(s.requestStats.requestCount) / elapsed
	}

	return summary, nil
}

// calculatePercentile 计算百分位数
// 使用标准库 O(n log n) 排序替代原先 O(n²) 冒泡排序
func calculatePercentile(values []float64, percentile float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	index := int(float64(len(sorted)-1) * percentile)
	return sorted[index]
}

// ResetRequestStats 重置请求统计
func (s *MonitorService) ResetRequestStats() {
	s.requestStats.mutex <- struct{}{}
	defer func() { <-s.requestStats.mutex }()

	s.requestStats.requestCount = 0
	s.requestStats.responseTimes = make([]float64, 0, 1000)
	s.requestStats.statusCodes = make(map[int]int64)
	s.requestStats.lastResetTime = time.Now()
}

// ==================== 数据查询接口 ====================

// GetSystemMetricsSummary 获取系统指标汇总
func (s *MonitorService) GetSystemMetricsSummary(ctx context.Context) (*monitor.SystemMetricsSummary, error) {
	return s.monitorRepo.GetSystemMetricsSummary(ctx, s.hostname)
}

// GetLLMNodeSummaries 获取 LLM 节点汇总
func (s *MonitorService) GetLLMNodeSummaries(ctx context.Context) ([]monitor.LLMNodeSummary, error) {
	return s.monitorRepo.GetLLMNodeMetricSummary(ctx)
}

// GetDashboardSummary 获取仪表盘汇总数据
// 所有独立查询并行执行，将延迟从串行累加降低到最长单次查询
func (s *MonitorService) GetDashboardSummary(ctx context.Context) (*monitor.DashboardSummary, error) {
	summary := &monitor.DashboardSummary{
		UpdatedAt: time.Now(),
	}

	var wg sync.WaitGroup
	wg.Add(5)

	go func() {
		defer wg.Done()
		if metrics, err := s.GetSystemMetricsSummary(ctx); err == nil {
			summary.SystemStatus = metrics
		}
	}()

	go func() {
		defer wg.Done()
		if metrics, err := s.GetRequestMetrics(ctx, 5*time.Minute); err == nil {
			summary.RequestMetrics = metrics
		}
	}()

	go func() {
		defer wg.Done()
		if nodes, err := s.GetLLMNodeSummaries(ctx); err == nil {
			summary.LLMNodes = nodes
		}
	}()

	go func() {
		defer wg.Done()
		if count, err := s.backendRepo.CountEnabled(); err == nil {
			summary.ActiveNodes = int(count)
		}
	}()

	go func() {
		defer wg.Done()
		if count, err := s.backendRepo.CountAll(); err == nil {
			summary.TotalNodes = int(count)
		}
	}()

	wg.Wait()
	return summary, nil
}

// ==================== LLM 节点上报 ====================

// ReportLLMNodeMetrics 处理 LLM 节点指标上报
func (s *MonitorService) ReportLLMNodeMetrics(ctx context.Context, req *monitor.NodeReportRequest) error {
	// 解析 GPU 信息
	gpuInfoJSON, _ := json.Marshal(req.GPUInfo)
	
	// 解析已加载模型
	modelsJSON, _ := json.Marshal(req.LoadedModels)
	
	// 解析标签
	labelsJSON, _ := json.Marshal(req.Labels)

	// 计算 GPU 总显存和已用显存
	var gpuTotalMem, gpuUsedMem float64
	for _, gpu := range req.GPUInfo {
		gpuTotalMem += gpu.TotalMemoryGB
		gpuUsedMem += gpu.UsedMemoryGB
	}

	// 构建指标记录
	metric := &monitor.LLMNodeMetric{
		NodeID:            req.NodeID,
		NodeName:          req.NodeName,
		Status:            req.Status,
		GPUInfo:           string(gpuInfoJSON),
		GPUCount:          len(req.GPUInfo),
		GPUTotalMemoryGB:  gpuTotalMem,
		GPUUsedMemoryGB:   gpuUsedMem,
		GPUUtilization:    req.GPUUtilization,
		CPUCores:          req.CPUCores,
		CPUUsagePercent:   req.CPUUsagePercent,
		MemoryTotalGB:     req.MemoryTotalGB,
		MemoryUsedGB:      req.MemoryUsedGB,
		RequestsPerMin:    req.RequestsPerMin,
		AvgResponseTimeMs: req.AvgResponseTimeMs,
		ActiveRequests:    req.ActiveRequests,
		QueuedRequests:    req.QueuedRequests,
		LoadedModels:      string(modelsJSON),
		ModelCount:        len(req.LoadedModels),
		Version:           req.Version,
		Labels:            string(labelsJSON),
		ReportedAt:        time.Unix(req.Timestamp, 0),
	}

	// 如果上报时间为空，使用当前时间
	if req.Timestamp == 0 {
		metric.ReportedAt = time.Now()
	}

	return s.monitorRepo.CreateLLMNodeMetric(ctx, metric)
}

// GetHostname 获取当前主机名
func (s *MonitorService) GetHostname() string {
	return s.hostname
}
