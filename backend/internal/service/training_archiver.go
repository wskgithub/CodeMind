package service

import (
	"archive/tar"
	"codemind/internal/model"
	"codemind/internal/repository"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
)

const (
	archiveThreshold     = 100000
	archiveCheckInterval = 30 * time.Minute
	archiveExportBatch   = 500
	archiveDeleteBatch   = 2000
)

// ArchiveMetadata contains metadata for archived files.
type ArchiveMetadata struct {
	ArchivedAt   time.Time `json:"archived_at"`
	MinID        int64     `json:"min_id"`
	MaxID        int64     `json:"max_id"`
	RecordCount  int64     `json:"record_count"`
	OriginalSize int64     `json:"original_size"`
}

// TrainingDataArchiver archives training data when threshold is reached.
type TrainingDataArchiver struct {
	repo       *repository.TrainingDataRepository
	logger     *zap.Logger
	stopCh     chan struct{}
	archiveDir string
	wg         sync.WaitGroup
	mu         sync.Mutex
}

// NewTrainingDataArchiver creates and starts the archiver service.
func NewTrainingDataArchiver(
	repo *repository.TrainingDataRepository,
	logger *zap.Logger,
	archiveDir string,
) *TrainingDataArchiver {
	if archiveDir == "" {
		archiveDir = "./data/training-archives"
	}

	a := &TrainingDataArchiver{
		repo:       repo,
		logger:     logger,
		archiveDir: archiveDir,
		stopCh:     make(chan struct{}),
	}

	//nolint:mnd // magic number for configuration/defaults.
	if err := os.MkdirAll(archiveDir, 0o755); err != nil {
		logger.Error("failed to create archive directory", zap.String("dir", archiveDir), zap.Error(err))
	}

	a.wg.Add(1)
	go a.watchLoop()

	logger.Info("training data archiver started",
		zap.String("archive_dir", archiveDir),
		zap.Int("threshold", archiveThreshold),
		zap.Duration("check_interval", archiveCheckInterval),
	)

	return a
}

// Close stops the archiver and waits for current operation to complete.
func (a *TrainingDataArchiver) Close() {
	close(a.stopCh)
	a.wg.Wait()
	a.logger.Info("training data archiver stopped")
}

func (a *TrainingDataArchiver) watchLoop() {
	defer a.wg.Done()

	a.tryArchive()

	ticker := time.NewTicker(archiveCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			a.tryArchive()
		case <-a.stopCh:
			return
		}
	}
}

func (a *TrainingDataArchiver) tryArchive() {
	if !a.mu.TryLock() {
		return
	}
	defer a.mu.Unlock()

	count, err := a.repo.CountAll()
	if err != nil {
		a.logger.Error("failed to count training data", zap.Error(err))
		return
	}

	if count < archiveThreshold {
		return
	}

	a.logger.Info("training data threshold reached, starting archive",
		zap.Int64("current_count", count),
		zap.Int("threshold", archiveThreshold),
	)

	if err := a.doArchive(); err != nil {
		a.logger.Error("archive failed", zap.Error(err))
		return
	}

	a.logger.Info("training data archive completed")
}

func (a *TrainingDataArchiver) doArchive() error {
	boundaryID, err := a.repo.GetArchiveBoundaryID(archiveThreshold)
	if err != nil {
		return fmt.Errorf("failed to get archive boundary ID: %w", err)
	}
	if boundaryID <= 0 {
		return fmt.Errorf("invalid boundary ID: %d", boundaryID)
	}

	minID, maxID, err := a.repo.GetIDRange(boundaryID)
	if err != nil {
		return fmt.Errorf("failed to get ID range: %w", err)
	}

	a.logger.Info("archive range determined",
		zap.Int64("min_id", minID),
		zap.Int64("max_id", maxID),
		zap.Int64("boundary_id", boundaryID),
	)

	tmpFile, recordCount, err := a.exportToTempFile(boundaryID)
	if err != nil {
		return fmt.Errorf("failed to export training data: %w", err)
	}
	defer func() { _ = os.Remove(tmpFile) }()

	archivePath, err := a.compressArchive(tmpFile, minID, maxID, recordCount)
	if err != nil {
		return fmt.Errorf("failed to compress archive: %w", err)
	}

	a.logger.Info("archive file created",
		zap.String("path", archivePath),
		zap.Int64("records", recordCount),
	)

	deleted, err := a.repo.DeleteByIDRange(boundaryID, archiveDeleteBatch)
	if err != nil {
		return fmt.Errorf("failed to delete archived data (deleted %d): %w", deleted, err)
	}

	a.logger.Info("archived data cleanup completed",
		zap.Int64("deleted", deleted),
		zap.String("archive", filepath.Base(archivePath)),
	)

	return nil
}

func (a *TrainingDataArchiver) exportToTempFile(maxID int64) (string, int64, error) {
	tmpFile, err := os.CreateTemp(a.archiveDir, "archive-export-*.jsonl")
	if err != nil {
		return "", 0, fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	var recordCount int64
	exportErr := a.repo.StreamByIDRange(maxID, archiveExportBatch, func(batch []model.LLMTrainingData) error {
		for _, record := range batch {
			line, err := json.Marshal(record)
			if err != nil {
				a.logger.Warn("failed to serialize training record",
					zap.Int64("id", record.ID),
					zap.Error(err),
				)
				continue
			}
			if _, err := tmpFile.Write(line); err != nil {
				return err
			}
			if _, err := tmpFile.Write([]byte("\n")); err != nil {
				return err
			}
			recordCount++
		}
		return nil
	})

	_ = tmpFile.Close()

	if exportErr != nil {
		_ = os.Remove(tmpPath)
		return "", 0, exportErr
	}

	return tmpPath, recordCount, nil
}

func (a *TrainingDataArchiver) compressArchive(jsonlPath string, minID, maxID, recordCount int64) (string, error) {
	ts := time.Now().Format("20060102_150405")
	archiveName := fmt.Sprintf("training_%d_%d_%s.tar.gz", minID, maxID, ts)
	archivePath := filepath.Join(a.archiveDir, archiveName)

	outFile, err := os.Create(archivePath)
	if err != nil {
		return "", err
	}
	defer func() { _ = outFile.Close() }()

	gzWriter := gzip.NewWriter(outFile)
	defer func() { _ = gzWriter.Close() }()

	tarWriter := tar.NewWriter(gzWriter)
	defer func() { _ = tarWriter.Close() }()

	jsonlInfo, err := os.Stat(jsonlPath)
	if err != nil {
		return "", err
	}

	meta := ArchiveMetadata{
		MinID:        minID,
		MaxID:        maxID,
		RecordCount:  recordCount,
		ArchivedAt:   time.Now(),
		OriginalSize: jsonlInfo.Size(),
	}
	metaBytes, _ := json.MarshalIndent(meta, "", "  ")

	if err := tarWriter.WriteHeader(&tar.Header{
		Name:    "metadata.json",
		Size:    int64(len(metaBytes)),
		Mode:    0o644, //nolint:mnd // intentional constant.
		ModTime: time.Now(),
	}); err != nil {
		return "", err
	}
	if _, err := tarWriter.Write(metaBytes); err != nil {
		return "", err
	}

	if err := tarWriter.WriteHeader(&tar.Header{
		Name:    "training_data.jsonl",
		Size:    jsonlInfo.Size(),
		Mode:    0o644, //nolint:mnd // intentional constant.
		ModTime: time.Now(),
	}); err != nil {
		return "", err
	}

	jsonlFile, err := os.Open(jsonlPath)
	if err != nil {
		return "", err
	}
	defer func() { _ = jsonlFile.Close() }()

	if _, err := io.Copy(tarWriter, jsonlFile); err != nil {
		return "", err
	}

	return archivePath, nil
}
