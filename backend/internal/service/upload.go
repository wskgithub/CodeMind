package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
)

// Allowed image types for upload.
var allowedImageTypes = map[string]bool{
	"image/jpeg":    true,
	"image/png":     true,
	"image/gif":     true,
	"image/webp":    true,
	"image/svg+xml": true,
}

// File extension mapping for image MIME types.
var mimeToExt = map[string]string{
	"image/jpeg":    ".jpg",
	"image/png":     ".png",
	"image/gif":     ".gif",
	"image/webp":    ".webp",
	"image/svg+xml": ".svg",
}

// UploadResult represents the result of a file upload.
type UploadResult struct {
	URL      string `json:"url"`
	Filename string `json:"filename"`
}

// UploadService handles file upload operations.
type UploadService struct {
	logger    *zap.Logger
	uploadDir string
	urlPrefix string
	maxSize   int64
}

// NewUploadService creates a new UploadService.
func NewUploadService(uploadDir string, maxSizeMB int, urlPrefix string, logger *zap.Logger) *UploadService {
	if uploadDir == "" {
		uploadDir = "./uploads"
	}
	if maxSizeMB <= 0 {
		maxSizeMB = 5
	}
	if urlPrefix == "" {
		urlPrefix = "/uploads"
	}

	return &UploadService{
		uploadDir: uploadDir,
		maxSize:   int64(maxSizeMB) << 20, //nolint:mnd // intentional constant.
		urlPrefix: urlPrefix,
		logger:    logger,
	}
}

// UploadImage uploads an image file.
func (s *UploadService) UploadImage(file *multipart.FileHeader) (*UploadResult, error) {
	if file.Size > s.maxSize {
		return nil, fmt.Errorf("file size exceeds limit (max %dMB)", s.maxSize>>20) //nolint:mnd // intentional constant.
	}

	contentType := file.Header.Get("Content-Type")
	if !allowedImageTypes[contentType] {
		return nil, fmt.Errorf("unsupported image format, allowed: JPG, PNG, GIF, WebP, SVG")
	}

	ext, ok := mimeToExt[contentType]
	if !ok {
		ext = filepath.Ext(file.Filename)
		if ext == "" {
			ext = ".png"
		}
	}

	// Organize into date-based subdirectories: uploads/docs/2026/04/
	now := time.Now()
	subDir := filepath.Join("docs", fmt.Sprintf("%d", now.Year()), fmt.Sprintf("%02d", now.Month()))
	absDir := filepath.Join(s.uploadDir, subDir)

	//nolint:mnd // magic number for configuration/defaults.
	if err := os.MkdirAll(absDir, 0o755); err != nil {
		s.logger.Error("failed to create upload directory", zap.String("dir", absDir), zap.Error(err))
		return nil, fmt.Errorf("failed to create storage directory")
	}

	// Generate random filename to prevent conflicts and path traversal
	randBytes := make([]byte, 12) //nolint:mnd // intentional constant.
	if _, err := rand.Read(randBytes); err != nil {
		return nil, fmt.Errorf("failed to generate filename")
	}
	filename := hex.EncodeToString(randBytes) + ext
	absPath := filepath.Join(absDir, filename)

	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to read uploaded file")
	}
	defer func() { _ = src.Close() }()

	dst, err := os.Create(absPath)
	if err != nil {
		s.logger.Error("failed to create destination file", zap.String("path", absPath), zap.Error(err))
		return nil, fmt.Errorf("failed to save file")
	}
	defer func() { _ = dst.Close() }()

	if _, err := io.Copy(dst, src); err != nil {
		s.logger.Error("failed to write file", zap.String("path", absPath), zap.Error(err))
		_ = os.Remove(absPath)
		return nil, fmt.Errorf("failed to write file")
	}

	relPath := filepath.Join(subDir, filename)
	url := s.urlPrefix + "/" + strings.ReplaceAll(relPath, string(os.PathSeparator), "/")

	s.logger.Info("image uploaded successfully",
		zap.String("filename", file.Filename),
		zap.Int64("size", file.Size),
		zap.String("url", url),
	)

	return &UploadResult{
		URL:      url,
		Filename: filename,
	}, nil
}
