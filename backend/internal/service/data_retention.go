package service

import (
	"sync"
	"time"

	"codemind/internal/repository"

	"go.uber.org/zap"
)

const (
	// 明细数据默认保留天数（90 天）
	// token_usage 和 request_logs 已有每日汇总表，明细仅用于排查，无需永久保留
	defaultRetentionDays = 90
	// 清理检查间隔
	retentionCheckInterval = 6 * time.Hour
	// 每批删除行数
	retentionDeleteBatch = 5000
)

// DataRetentionCleaner 数据保留清理服务
// 定期清理超过保留期的 token_usage 和 request_logs 明细记录，
// 防止大表无限膨胀导致写入和索引维护性能退化
type DataRetentionCleaner struct {
	usageRepo     *repository.UsageRepository
	logger        *zap.Logger
	retentionDays int

	stopCh chan struct{}
	wg     sync.WaitGroup
}

// NewDataRetentionCleaner 创建数据保留清理服务并启动后台协程
func NewDataRetentionCleaner(
	usageRepo *repository.UsageRepository,
	logger *zap.Logger,
	retentionDays int,
) *DataRetentionCleaner {
	if retentionDays <= 0 {
		retentionDays = defaultRetentionDays
	}

	c := &DataRetentionCleaner{
		usageRepo:     usageRepo,
		logger:        logger,
		retentionDays: retentionDays,
		stopCh:        make(chan struct{}),
	}

	c.wg.Add(1)
	go c.watchLoop()

	logger.Info("数据保留清理服务已启动",
		zap.Int("retention_days", retentionDays),
		zap.Duration("check_interval", retentionCheckInterval),
	)

	return c
}

// Close 停止清理服务
func (c *DataRetentionCleaner) Close() {
	close(c.stopCh)
	c.wg.Wait()
	c.logger.Info("数据保留清理服务已停止")
}

// watchLoop 定期执行清理
func (c *DataRetentionCleaner) watchLoop() {
	defer c.wg.Done()

	// 启动后延迟 5 分钟再执行首次清理，避免与启动初始化竞争
	select {
	case <-time.After(5 * time.Minute):
		c.cleanup()
	case <-c.stopCh:
		return
	}

	ticker := time.NewTicker(retentionCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.cleanup()
		case <-c.stopCh:
			return
		}
	}
}

// cleanup 执行一次清理
func (c *DataRetentionCleaner) cleanup() {
	cutoff := time.Now().AddDate(0, 0, -c.retentionDays)

	c.logger.Info("开始清理过期明细数据",
		zap.Time("cutoff", cutoff),
		zap.Int("retention_days", c.retentionDays),
	)

	// 清理 token_usage 明细表
	usageDeleted, err := c.usageRepo.DeleteOldUsageRecords(cutoff, retentionDeleteBatch)
	if err != nil {
		c.logger.Error("清理 token_usage 失败", zap.Error(err))
	} else if usageDeleted > 0 {
		c.logger.Info("token_usage 清理完成", zap.Int64("deleted", usageDeleted))
	}

	// 清理 request_logs 表
	logsDeleted, err := c.usageRepo.DeleteOldRequestLogs(cutoff, retentionDeleteBatch)
	if err != nil {
		c.logger.Error("清理 request_logs 失败", zap.Error(err))
	} else if logsDeleted > 0 {
		c.logger.Info("request_logs 清理完成", zap.Int64("deleted", logsDeleted))
	}
}
