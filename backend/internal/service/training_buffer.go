package service

import (
	"sync"
	"time"

	"codemind/internal/model"
	"codemind/internal/repository"

	"go.uber.org/zap"
)

const (
	defaultBufferCap     = 2048
	defaultBatchSize     = 50
	defaultFlushInterval = 3 * time.Second
)

// TrainingDataBuffer batches training data writes to reduce DB I/O.
type TrainingDataBuffer struct {
	repo   *repository.TrainingDataRepository
	logger *zap.Logger

	ch     chan *model.LLMTrainingData
	stopCh chan struct{}
	wg     sync.WaitGroup

	batchSize     int
	flushInterval time.Duration
}

// NewTrainingDataBuffer creates and starts a new buffer.
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

	logger.Info("training data buffer started",
		zap.Int("buffer_cap", defaultBufferCap),
		zap.Int("batch_size", defaultBatchSize),
		zap.Duration("flush_interval", defaultFlushInterval),
	)

	return b
}

// Add adds a record to the buffer. Non-blocking: drops record if buffer is full.
func (b *TrainingDataBuffer) Add(record *model.LLMTrainingData) {
	select {
	case b.ch <- record:
	default:
		b.logger.Warn("training data buffer full, dropping record",
			zap.Int64("user_id", record.UserID),
			zap.String("model", record.Model),
		)
	}
}

// Close closes the buffer and flushes remaining data.
func (b *TrainingDataBuffer) Close() {
	close(b.stopCh)
	b.wg.Wait()
	b.logger.Info("training data buffer closed")
}

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
			b.logger.Error("batch write failed, falling back to single writes",
				zap.Int("count", count),
				zap.Error(err),
			)
			for _, record := range batch {
				if e := b.repo.Create(record); e != nil {
					b.logger.Error("single write also failed",
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
