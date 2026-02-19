package service

import (
	"strings"

	"codemind/internal/model"
	"codemind/internal/model/dto"
	"codemind/internal/pkg/errcode"
	"codemind/internal/repository"

	"go.uber.org/zap"
)

// DocumentService 文档管理业务逻辑
type DocumentService struct {
	repo   repository.DocumentRepository
	logger *zap.Logger
}

// NewDocumentService 创建文档服务
func NewDocumentService(repo repository.DocumentRepository, logger *zap.Logger) *DocumentService {
	return &DocumentService{repo: repo, logger: logger}
}

// List 获取已发布的文档列表（面向用户）
func (s *DocumentService) List() ([]model.DocumentListItem, error) {
	items, err := s.repo.List()
	if err != nil {
		s.logger.Error("获取文档列表失败", zap.Error(err))
		return nil, errcode.ErrDatabase
	}
	return items, nil
}

// GetBySlug 根据 slug 获取文档详情（面向用户，仅已发布）
func (s *DocumentService) GetBySlug(slug string) (*model.Document, error) {
	doc, err := s.repo.GetBySlug(slug)
	if err != nil {
		return nil, errcode.ErrRecordNotFound
	}
	return doc, nil
}

// ListAll 获取所有文档（管理用，包括未发布）
func (s *DocumentService) ListAll() ([]model.Document, error) {
	docs, err := s.repo.ListAll()
	if err != nil {
		s.logger.Error("获取全部文档列表失败", zap.Error(err))
		return nil, errcode.ErrDatabase
	}
	return docs, nil
}

// GetByID 根据 ID 获取文档（管理用）
func (s *DocumentService) GetByID(id int64) (*model.Document, error) {
	doc, err := s.repo.GetByID(id)
	if err != nil {
		return nil, errcode.ErrRecordNotFound
	}
	return doc, nil
}

// Create 创建文档
func (s *DocumentService) Create(req *dto.CreateDocumentRequest) (*model.Document, error) {
	slug := strings.ToLower(strings.TrimSpace(req.Slug))

	doc := &model.Document{
		Slug:        slug,
		Title:       req.Title,
		Subtitle:    req.Subtitle,
		Icon:        req.Icon,
		Content:     req.Content,
		SortOrder:   req.SortOrder,
		IsPublished: req.IsPublished,
	}

	if err := s.repo.Create(doc); err != nil {
		s.logger.Error("创建文档失败", zap.Error(err))
		return nil, errcode.ErrDatabase
	}
	return doc, nil
}

// Update 更新文档
func (s *DocumentService) Update(id int64, req *dto.UpdateDocumentRequest) (*model.Document, error) {
	doc, err := s.repo.GetByID(id)
	if err != nil {
		return nil, errcode.ErrRecordNotFound
	}

	doc.Title = req.Title
	doc.Subtitle = req.Subtitle
	doc.Icon = req.Icon
	doc.Content = req.Content
	doc.SortOrder = req.SortOrder
	doc.IsPublished = req.IsPublished

	if err := s.repo.Update(doc); err != nil {
		s.logger.Error("更新文档失败", zap.Int64("id", id), zap.Error(err))
		return nil, errcode.ErrDatabase
	}
	return doc, nil
}

// Delete 删除文档
func (s *DocumentService) Delete(id int64) error {
	if _, err := s.repo.GetByID(id); err != nil {
		return errcode.ErrRecordNotFound
	}
	if err := s.repo.Delete(id); err != nil {
		s.logger.Error("删除文档失败", zap.Int64("id", id), zap.Error(err))
		return errcode.ErrDatabase
	}
	return nil
}

// Initialize 初始化默认文档（仅在表为空时执行）
func (s *DocumentService) Initialize() (int, error) {
	docs, err := s.repo.ListAll()
	if err != nil {
		s.logger.Error("检查文档初始化状态失败", zap.Error(err))
		return 0, errcode.ErrDatabase
	}
	if len(docs) > 0 {
		return 0, errcode.ErrInvalidParams.WithMessage("文档已存在，无法初始化")
	}

	defaultDocs := model.DefaultTools
	for i := range defaultDocs {
		defaultDocs[i].Content = getDefaultContent(defaultDocs[i].Slug)
	}

	if err := s.repo.BatchCreate(defaultDocs); err != nil {
		s.logger.Error("初始化文档失败", zap.Error(err))
		return 0, errcode.ErrDatabase
	}
	return len(defaultDocs), nil
}
