package service

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"codemind/internal/model"
	"codemind/internal/repository"

	"go.uber.org/zap"
)

const (
	// 累积达到此阈值后自动触发归档
	archiveThreshold = 100000
	// 归档检查间隔
	archiveCheckInterval = 30 * time.Minute
	// 归档导出时的批次大小
	archiveExportBatch = 500
	// 归档删除时的批次大小（较小值减少锁竞争）
	archiveDeleteBatch = 2000
)

// ArchiveMetadata 归档文件的元信息
type ArchiveMetadata struct {
	MinID        int64     `json:"min_id"`
	MaxID        int64     `json:"max_id"`
	RecordCount  int64     `json:"record_count"`
	ArchivedAt   time.Time `json:"archived_at"`
	OriginalSize int64     `json:"original_size"`
}

// TrainingDataArchiver 训练数据阶段性归档服务
// 当记录累积达到阈值后，自动将数据打包压缩为 .tar.gz 存档，
// 然后从数据库中清理已归档的记录
type TrainingDataArchiver struct {
	repo       *repository.TrainingDataRepository
	logger     *zap.Logger
	archiveDir string

	stopCh chan struct{}
	wg     sync.WaitGroup
	mu     sync.Mutex // 防止并发归档
}

// NewTrainingDataArchiver 创建训练数据归档服务并启动后台检查协程
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

	if err := os.MkdirAll(archiveDir, 0755); err != nil {
		logger.Error("创建归档目录失败", zap.String("dir", archiveDir), zap.Error(err))
	}

	a.wg.Add(1)
	go a.watchLoop()

	logger.Info("训练数据归档服务已启动",
		zap.String("archive_dir", archiveDir),
		zap.Int("threshold", archiveThreshold),
		zap.Duration("check_interval", archiveCheckInterval),
	)

	return a
}

// Close 停止归档服务并等待当前归档操作完成
func (a *TrainingDataArchiver) Close() {
	close(a.stopCh)
	a.wg.Wait()
	a.logger.Info("训练数据归档服务已停止")
}

// watchLoop 定期检查是否需要触发归档
func (a *TrainingDataArchiver) watchLoop() {
	defer a.wg.Done()

	// 启动后先执行一次检查
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

// tryArchive 检查数据量并在达到阈值时执行归档
func (a *TrainingDataArchiver) tryArchive() {
	// 防止并发归档
	if !a.mu.TryLock() {
		return
	}
	defer a.mu.Unlock()

	count, err := a.repo.CountAll()
	if err != nil {
		a.logger.Error("检查训练数据数量失败", zap.Error(err))
		return
	}

	if count < archiveThreshold {
		return
	}

	a.logger.Info("训练数据量达到归档阈值，开始归档",
		zap.Int64("current_count", count),
		zap.Int("threshold", archiveThreshold),
	)

	if err := a.doArchive(); err != nil {
		a.logger.Error("归档执行失败", zap.Error(err))
		return
	}

	a.logger.Info("训练数据归档完成")
}

// doArchive 执行一次完整的归档流程
// 流程：确定边界 → 导出到临时文件 → 压缩打包 → 分批删除已归档数据
func (a *TrainingDataArchiver) doArchive() error {
	// 第一步：确定归档边界（基于 ID 的快照语义）
	// 取第 archiveThreshold 条记录的 ID 作为边界，
	// 此后新插入的记录 ID 必然大于边界值，不会被误删
	boundaryID, err := a.repo.GetArchiveBoundaryID(archiveThreshold)
	if err != nil {
		return fmt.Errorf("获取归档边界 ID 失败: %w", err)
	}
	if boundaryID <= 0 {
		return fmt.Errorf("无效的边界 ID: %d", boundaryID)
	}

	minID, maxID, err := a.repo.GetIDRange(boundaryID)
	if err != nil {
		return fmt.Errorf("获取 ID 范围失败: %w", err)
	}

	a.logger.Info("归档范围确定",
		zap.Int64("min_id", minID),
		zap.Int64("max_id", maxID),
		zap.Int64("boundary_id", boundaryID),
	)

	// 第二步：导出数据到临时 JSONL 文件
	tmpFile, recordCount, err := a.exportToTempFile(boundaryID)
	if err != nil {
		return fmt.Errorf("导出训练数据失败: %w", err)
	}
	defer os.Remove(tmpFile)

	// 第三步：打包压缩为 .tar.gz
	archivePath, err := a.compressArchive(tmpFile, minID, maxID, recordCount)
	if err != nil {
		return fmt.Errorf("压缩归档文件失败: %w", err)
	}

	a.logger.Info("归档文件创建成功",
		zap.String("path", archivePath),
		zap.Int64("records", recordCount),
	)

	// 第四步：分批删除已归档的数据库记录
	deleted, err := a.repo.DeleteByIDRange(boundaryID, archiveDeleteBatch)
	if err != nil {
		// 归档文件已生成，删除失败不影响数据完整性
		// 下次归档时会重新检测到超阈值并继续
		return fmt.Errorf("删除已归档数据失败（已删除 %d 条）: %w", deleted, err)
	}

	a.logger.Info("已归档数据清理完成",
		zap.Int64("deleted", deleted),
		zap.String("archive", filepath.Base(archivePath)),
	)

	return nil
}

// exportToTempFile 将指定 ID 范围内的数据导出为临时 JSONL 文件
func (a *TrainingDataArchiver) exportToTempFile(maxID int64) (string, int64, error) {
	tmpFile, err := os.CreateTemp(a.archiveDir, "archive-export-*.jsonl")
	if err != nil {
		return "", 0, fmt.Errorf("创建临时文件失败: %w", err)
	}
	tmpPath := tmpFile.Name()

	var recordCount int64
	exportErr := a.repo.StreamByIDRange(maxID, archiveExportBatch, func(batch []model.LLMTrainingData) error {
		for _, record := range batch {
			line, err := json.Marshal(record)
			if err != nil {
				a.logger.Warn("序列化训练数据记录失败",
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

	tmpFile.Close()

	if exportErr != nil {
		os.Remove(tmpPath)
		return "", 0, exportErr
	}

	return tmpPath, recordCount, nil
}

// compressArchive 将 JSONL 文件打包压缩为 .tar.gz 归档
// 归档包含两个文件：metadata.json（元信息）和 training_data.jsonl（数据）
func (a *TrainingDataArchiver) compressArchive(jsonlPath string, minID, maxID, recordCount int64) (string, error) {
	ts := time.Now().Format("20060102_150405")
	archiveName := fmt.Sprintf("training_%d_%d_%s.tar.gz", minID, maxID, ts)
	archivePath := filepath.Join(a.archiveDir, archiveName)

	outFile, err := os.Create(archivePath)
	if err != nil {
		return "", err
	}
	defer outFile.Close()

	gzWriter := gzip.NewWriter(outFile)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	// 写入元信息
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
		Mode:    0644,
		ModTime: time.Now(),
	}); err != nil {
		return "", err
	}
	if _, err := tarWriter.Write(metaBytes); err != nil {
		return "", err
	}

	// 写入 JSONL 数据文件
	if err := tarWriter.WriteHeader(&tar.Header{
		Name:    "training_data.jsonl",
		Size:    jsonlInfo.Size(),
		Mode:    0644,
		ModTime: time.Now(),
	}); err != nil {
		return "", err
	}

	jsonlFile, err := os.Open(jsonlPath)
	if err != nil {
		return "", err
	}
	defer jsonlFile.Close()

	if _, err := io.Copy(tarWriter, jsonlFile); err != nil {
		return "", err
	}

	return archivePath, nil
}
