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

// 允许上传的图片类型
var allowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
	"image/svg+xml": true,
}

// 图片类型对应的文件扩展名
var mimeToExt = map[string]string{
	"image/jpeg":    ".jpg",
	"image/png":     ".png",
	"image/gif":     ".gif",
	"image/webp":    ".webp",
	"image/svg+xml": ".svg",
}

// UploadResult 上传结果
type UploadResult struct {
	URL      string `json:"url"`
	Filename string `json:"filename"`
}

// UploadService 文件上传服务
type UploadService struct {
	uploadDir  string
	maxSize    int64
	urlPrefix  string
	logger     *zap.Logger
}

// NewUploadService 创建上传服务
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
		maxSize:   int64(maxSizeMB) << 20,
		urlPrefix: urlPrefix,
		logger:    logger,
	}
}

// UploadImage 上传图片文件
func (s *UploadService) UploadImage(file *multipart.FileHeader) (*UploadResult, error) {
	if file.Size > s.maxSize {
		return nil, fmt.Errorf("文件大小超过限制（最大 %dMB）", s.maxSize>>20)
	}

	contentType := file.Header.Get("Content-Type")
	if !allowedImageTypes[contentType] {
		return nil, fmt.Errorf("不支持的图片格式，仅允许: JPG, PNG, GIF, WebP, SVG")
	}

	ext, ok := mimeToExt[contentType]
	if !ok {
		ext = filepath.Ext(file.Filename)
		if ext == "" {
			ext = ".png"
		}
	}

	// 按日期分子目录：uploads/docs/2026/04/
	now := time.Now()
	subDir := filepath.Join("docs", fmt.Sprintf("%d", now.Year()), fmt.Sprintf("%02d", now.Month()))
	absDir := filepath.Join(s.uploadDir, subDir)

	if err := os.MkdirAll(absDir, 0755); err != nil {
		s.logger.Error("创建上传目录失败", zap.String("dir", absDir), zap.Error(err))
		return nil, fmt.Errorf("创建存储目录失败")
	}

	// 生成随机文件名，避免冲突和路径遍历
	randBytes := make([]byte, 12)
	if _, err := rand.Read(randBytes); err != nil {
		return nil, fmt.Errorf("生成文件名失败")
	}
	filename := hex.EncodeToString(randBytes) + ext
	absPath := filepath.Join(absDir, filename)

	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("读取上传文件失败")
	}
	defer src.Close()

	dst, err := os.Create(absPath)
	if err != nil {
		s.logger.Error("创建目标文件失败", zap.String("path", absPath), zap.Error(err))
		return nil, fmt.Errorf("保存文件失败")
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		s.logger.Error("写入文件失败", zap.String("path", absPath), zap.Error(err))
		os.Remove(absPath)
		return nil, fmt.Errorf("写入文件失败")
	}

	relPath := filepath.Join(subDir, filename)
	url := s.urlPrefix + "/" + strings.ReplaceAll(relPath, string(os.PathSeparator), "/")

	s.logger.Info("图片上传成功",
		zap.String("filename", file.Filename),
		zap.Int64("size", file.Size),
		zap.String("url", url),
	)

	return &UploadResult{
		URL:      url,
		Filename: filename,
	}, nil
}
