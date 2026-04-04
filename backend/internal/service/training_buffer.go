package service

import (
	"sync"
	"time"

	"codemind/internal/model"
	"codemind/internal/repository"

	"go.uber.org/zap"
)

const (
	// 缓冲通道容量：积压超过此值时新记录将被丢弃（不阻塞调用方）
	defaultBufferCap = 2048
	// 累积到此数量时触发一次批量写入
	defaultBatchSize = 50
	// 无论是否凑满批次，最长等待此时间后强制刷盘
	defaultFlushInterval = 3 * time.Second
)

// TrainingDataBuffer 训练数据批量写入缓冲器
// 通过内存缓冲 + 定时/定量批量写入，将高频单条 INSERT
// 合并为低频批量 INSERT，显著减少数据库连接占用和 I/O 频率
type TrainingDataBuffer struct {
	repo   *repository.TrainingDataRepository
	logger *zap.Logger

	ch     chan *model.LLMTrainingData
	stopCh chan struct{}
	wg     sync.WaitGroup

	batchSize     int
	flushInterval time.Duration
}

// NewTrainingDataBuffer 创建训练数据缓冲器并启动后台消费协程
func NewTrainingDataBuffer(repo *repository.TrainingDataRepository, logger *zap.Logger) *TrainingDataBuffer {
	b := &TrainingDataBuffer{
		repo:          repo,
		logger:        logger,
		ch:            make(chan *model.LLMTrainingData, defaultBufferCap),
		stopCh:        make(chan struct{}),
		batchSize:     defaultBatchSize,
		flushInterval: defaultFlushInterval,
	}

	b.wg.Add(1)
	go b.consume()

	logger.Info("训练数据批量缓冲器已启动",
		zap.Int("buffer_cap", defaultBufferCap),
		zap.Int("batch_size", defaultBatchSize),
		zap.Duration("flush_interval", defaultFlushInterval),
	)

	return b
}

// Add 将训练数据记录放入缓冲区
// 非阻塞：缓冲区满时丢弃记录并打印警告
func (b *TrainingDataBuffer) Add(record *model.LLMTrainingData) {
	select {
	case b.ch <- record:
	default:
		b.logger.Warn("训练数据缓冲区已满，丢弃记录",
			zap.Int64("user_id", record.UserID),
			zap.String("model", record.Model),
		)
	}
}

// Close 关闭缓冲器，排空残留数据后返回
// 应在服务优雅关停时调用
func (b *TrainingDataBuffer) Close() {
	close(b.stopCh)
	b.wg.Wait()
	b.logger.Info("训练数据批量缓冲器已关闭")
}

// consume 后台消费协程：按批次大小或时间间隔触发批量写入
func (b *TrainingDataBuffer) consume() {
	defer b.wg.Done()

	batch := make([]*model.LLMTrainingData, 0, b.batchSize)
	timer := time.NewTimer(b.flushInterval)
	defer timer.Stop()

	flush := func() {
		if len(batch) == 0 {
			return
		}

		count := len(batch)
		if err := b.repo.BatchCreate(batch); err != nil {
			b.logger.Error("批量写入训练数据失败，降级为逐条写入",
				zap.Int("count", count),
				zap.Error(err),
			)
			for _, record := range batch {
				if e := b.repo.Create(record); e != nil {
					b.logger.Error("训练数据逐条写入也失败",
						zap.Int64("user_id", record.UserID),
						zap.Error(e),
					)
				}
			}
		}

		batch = batch[:0]
	}

	for {
		select {
		case record := <-b.ch:
			batch = append(batch, record)
			if len(batch) >= b.batchSize {
				flush()
				timer.Reset(b.flushInterval)
			}

		case <-timer.C:
			flush()
			timer.Reset(b.flushInterval)

		case <-b.stopCh:
			// 排空通道中残留的记录
			for {
				select {
				case record := <-b.ch:
					batch = append(batch, record)
				default:
					flush()
					return
				}
			}
		}
	}
}
