package service

import (
	"sync"
	"time"

	"codemind/internal/repository"

	"go.uber.org/zap"
)

const (
	defaultRetentionDays   = 90
	retentionCheckInterval = 6 * time.Hour
	retentionDeleteBatch   = 5000
)

// DataRetentionCleaner periodically cleans up old token_usage and request_logs records.
type DataRetentionCleaner struct {
	usageRepo     *repository.UsageRepository
	logger        *zap.Logger
	retentionDays int

	stopCh chan struct{}
	wg     sync.WaitGroup
}

// NewDataRetentionCleaner creates a data retention cleaner and starts background cleanup.
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

	logger.Info("data retention cleaner started",
		zap.Int("retention_days", retentionDays),
		zap.Duration("check_interval", retentionCheckInterval),
	)

	return c
}

// Close stops the cleaner service.
func (c *DataRetentionCleaner) Close() {
	close(c.stopCh)
	c.wg.Wait()
	c.logger.Info("data retention cleaner stopped")
}

func (c *DataRetentionCleaner) watchLoop() {
	defer c.wg.Done()

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

func (c *DataRetentionCleaner) cleanup() {
	cutoff := time.Now().AddDate(0, 0, -c.retentionDays)

	c.logger.Info("starting cleanup of expired data",
		zap.Time("cutoff", cutoff),
		zap.Int("retention_days", c.retentionDays),
	)

	usageDeleted, err := c.usageRepo.DeleteOldUsageRecords(cutoff, retentionDeleteBatch)
	if err != nil {
		c.logger.Error("failed to cleanup token_usage", zap.Error(err))
	} else if usageDeleted > 0 {
		c.logger.Info("token_usage cleanup completed", zap.Int64("deleted", usageDeleted))
	}

	logsDeleted, err := c.usageRepo.DeleteOldRequestLogs(cutoff, retentionDeleteBatch)
	if err != nil {
		c.logger.Error("failed to cleanup request_logs", zap.Error(err))
	} else if logsDeleted > 0 {
		c.logger.Info("request_logs cleanup completed", zap.Int64("deleted", logsDeleted))
	}
}
