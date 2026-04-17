package service

import (
	"codemind/internal/model/monitor"
	"codemind/internal/repository"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
	"go.uber.org/zap"
)

// MonitorService handles system and request monitoring.
type MonitorService struct {
	monitorRepo  *repository.MonitorRepository
	usageRepo    *repository.UsageRepository
	backendRepo  *repository.LLMBackendRepository
	rdb          *redis.Client
	logger       *zap.Logger
	requestStats *RequestStatsCollector
	hostname     string
}

// RequestStatsCollector collects request statistics.
type RequestStatsCollector struct {
	lastResetTime time.Time
	statusCodes   map[int]int64
	mutex         chan struct{}
	responseTimes []float64
	requestCount  int64
}

// NewMonitorService creates a new monitor service.
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
		monitorRepo: monitorRepo,
		usageRepo:   usageRepo,
		backendRepo: backendRepo,
		rdb:         rdb,
		logger:      logger,
		hostname:    hostname,
		requestStats: &RequestStatsCollector{
			responseTimes: make([]float64, 0, 1000), //nolint:mnd // intentional constant.
			statusCodes:   make(map[int]int64),
			lastResetTime: time.Now(),
			mutex:         make(chan struct{}, 1),
		},
	}

	go svc.startCollector()

	return svc
}

// startCollector starts background metrics collection.
func (s *MonitorService) startCollector() {
	ticker := time.NewTicker(30 * time.Second) //nolint:mnd // intentional constant.
	defer ticker.Stop()

	s.collectSystemMetrics()

	for range ticker.C {
		s.collectSystemMetrics()
	}
}

// collectSystemMetrics collects system metrics.
func (s *MonitorService) collectSystemMetrics() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) //nolint:mnd // intentional constant.
	defer cancel()

	metrics := make([]*monitor.SystemMetric, 0)
	now := time.Now()

	if info, err := cpu.Info(); err == nil && len(info) > 0 {
		modelName := info[0].ModelName
		cores := len(info)

		if percent, err := cpu.Percent(1*time.Second, false); err == nil && len(percent) > 0 {
			labels, _ := json.Marshal(map[string]string{
				"model_name": modelName,
				"core_count": fmt.Sprintf("%d", cores),
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

	if memInfo, err := mem.VirtualMemory(); err == nil {
		totalGB := float64(memInfo.Total) / 1024 / 1024 / 1024 //nolint:mnd // intentional constant.
		usedGB := float64(memInfo.Used) / 1024 / 1024 / 1024   //nolint:mnd // intentional constant.

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

	if partitions, err := disk.Partitions(false); err == nil {
		for _, part := range partitions {
			if part.Fstype == "tmpfs" || part.Fstype == "devtmpfs" || part.Fstype == "squashfs" {
				continue
			}

			if usage, err := disk.Usage(part.Mountpoint); err == nil {
				labels, _ := json.Marshal(map[string]string{
					"mount_point": part.Mountpoint,
					"device":      part.Device,
					"fstype":      part.Fstype,
				})

				totalGB := float64(usage.Total) / 1024 / 1024 / 1024 //nolint:mnd // intentional constant.
				usedGB := float64(usage.Used) / 1024 / 1024 / 1024   //nolint:mnd // intentional constant.

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

	if len(metrics) > 0 {
		if err := s.monitorRepo.CreateSystemMetrics(ctx, metrics); err != nil {
			s.logger.Error("failed to save system metrics", zap.Error(err))
		} else {
			s.logger.Debug("system metrics collected", zap.Int("count", len(metrics)))
		}
	}

	go s.cleanupOldMetrics()
}

// cleanupOldMetrics removes old metrics data.
func (s *MonitorService) cleanupOldMetrics() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second) //nolint:mnd // intentional constant.
	defer cancel()

	//nolint:mnd // magic number for configuration/defaults.
	if rows, err := s.monitorRepo.CleanupOldSystemMetrics(ctx, 7); err != nil {
		s.logger.Warn("failed to cleanup system metrics", zap.Error(err))
	} else if rows > 0 {
		s.logger.Info("cleaned up old system metrics", zap.Int64("rows", rows))
	}

	//nolint:mnd // magic number for configuration/defaults.
	if rows, err := s.monitorRepo.CleanupOldLLMNodeMetrics(ctx, 48); err != nil {
		s.logger.Warn("failed to cleanup LLM node metrics", zap.Error(err))
	} else if rows > 0 {
		s.logger.Info("cleaned up old LLM node metrics", zap.Int64("rows", rows))
	}
}

// RecordRequest records request statistics.
func (s *MonitorService) RecordRequest(statusCode int, responseTimeMs float64) {
	select {
	case s.requestStats.mutex <- struct{}{}:
		s.requestStats.requestCount++
		s.requestStats.statusCodes[statusCode]++
		s.requestStats.responseTimes = append(s.requestStats.responseTimes, responseTimeMs)

		//nolint:mnd // magic number for configuration/defaults.
		if len(s.requestStats.responseTimes) > 10000 {
			s.requestStats.responseTimes = s.requestStats.responseTimes[5000:]
		}
		<-s.requestStats.mutex
	default:
	}
}

// GetRequestMetrics returns request performance metrics.
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

	for code, count := range s.requestStats.statusCodes {
		summary.StatusCodes[code] = count
		summary.TotalRequests += count
		//nolint:mnd // magic number for configuration/defaults.
		if code >= 400 {
			summary.ErrorRate += float64(count)
		}
	}

	if summary.TotalRequests > 0 {
		summary.ErrorRate = (summary.ErrorRate / float64(summary.TotalRequests)) * 100 //nolint:mnd // intentional constant.
	}

	times := s.requestStats.responseTimes
	if len(times) > 0 {
		var total float64
		for _, t := range times {
			total += t
		}
		summary.AvgResponseTime = total / float64(len(times))
		summary.P95ResponseTime = calculatePercentile(times, 0.95) //nolint:mnd // intentional constant.
		summary.P99ResponseTime = calculatePercentile(times, 0.99) //nolint:mnd // intentional constant.
	}

	elapsed := time.Since(s.requestStats.lastResetTime).Seconds()
	if elapsed > 0 {
		summary.QPS = float64(s.requestStats.requestCount) / elapsed
	}

	return summary, nil
}

// calculatePercentile calculates the percentile value.
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

// ResetRequestStats resets request statistics.
func (s *MonitorService) ResetRequestStats() {
	s.requestStats.mutex <- struct{}{}
	defer func() { <-s.requestStats.mutex }()

	s.requestStats.requestCount = 0
	s.requestStats.responseTimes = make([]float64, 0, 1000) //nolint:mnd // intentional constant.
	s.requestStats.statusCodes = make(map[int]int64)
	s.requestStats.lastResetTime = time.Now()
}

// GetSystemMetricsSummary returns system metrics summary.
func (s *MonitorService) GetSystemMetricsSummary(ctx context.Context) (*monitor.SystemMetricsSummary, error) {
	return s.monitorRepo.GetSystemMetricsSummary(ctx, s.hostname)
}

// GetLLMNodeSummaries returns LLM node summaries.
func (s *MonitorService) GetLLMNodeSummaries(ctx context.Context) ([]monitor.LLMNodeSummary, error) {
	return s.monitorRepo.GetLLMNodeMetricSummary(ctx)
}

// GetDashboardSummary returns dashboard summary data with parallel queries.
func (s *MonitorService) GetDashboardSummary(ctx context.Context) (*monitor.DashboardSummary, error) {
	summary := &monitor.DashboardSummary{
		UpdatedAt: time.Now(),
	}

	var wg sync.WaitGroup
	wg.Add(5) //nolint:mnd // intentional constant.

	go func() {
		defer wg.Done()
		if metrics, err := s.GetSystemMetricsSummary(ctx); err == nil {
			summary.SystemStatus = metrics
		}
	}()

	go func() {
		defer wg.Done()
		//nolint:mnd // magic number for configuration/defaults.
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

// ReportLLMNodeMetrics handles LLM node metrics reporting.
func (s *MonitorService) ReportLLMNodeMetrics(ctx context.Context, req *monitor.NodeReportRequest) error {
	gpuInfoJSON, _ := json.Marshal(req.GPUInfo)
	modelsJSON, _ := json.Marshal(req.LoadedModels)
	labelsJSON, _ := json.Marshal(req.Labels)

	var gpuTotalMem, gpuUsedMem float64
	for _, gpu := range req.GPUInfo {
		gpuTotalMem += gpu.TotalMemoryGB
		gpuUsedMem += gpu.UsedMemoryGB
	}

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

	if req.Timestamp == 0 {
		metric.ReportedAt = time.Now()
	}

	return s.monitorRepo.CreateLLMNodeMetric(ctx, metric)
}

// GetHostname returns the current hostname.
func (s *MonitorService) GetHostname() string {
	return s.hostname
}
