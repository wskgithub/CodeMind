package service

import (
	"strings"

	"codemind/internal/model"
	"codemind/internal/model/dto"
	"codemind/internal/pkg/errcode"
	"codemind/internal/repository"

	"go.uber.org/zap"
)

// DocumentService provides document management operations.
type DocumentService struct {
	repo   repository.DocumentRepository
	logger *zap.Logger
}

// NewDocumentService creates a new DocumentService instance.
func NewDocumentService(repo repository.DocumentRepository, logger *zap.Logger) *DocumentService {
	return &DocumentService{repo: repo, logger: logger}
}

// List retrieves published documents (user-facing).
func (s *DocumentService) List() ([]model.DocumentListItem, error) {
	items, err := s.repo.List()
	if err != nil {
		s.logger.Error("failed to list documents", zap.Error(err))
		return nil, errcode.ErrDatabase
	}
	return items, nil
}

// GetBySlug retrieves a published document by its slug.
func (s *DocumentService) GetBySlug(slug string) (*model.Document, error) {
	doc, err := s.repo.GetBySlug(slug)
	if err != nil {
		return nil, errcode.ErrRecordNotFound
	}
	return doc, nil
}

// ListAll retrieves all documents including unpublished ones (admin).
func (s *DocumentService) ListAll() ([]model.Document, error) {
	docs, err := s.repo.ListAll()
	if err != nil {
		s.logger.Error("failed to list all documents", zap.Error(err))
		return nil, errcode.ErrDatabase
	}
	return docs, nil
}

// GetByID retrieves a document by ID (admin).
func (s *DocumentService) GetByID(id int64) (*model.Document, error) {
	doc, err := s.repo.GetByID(id)
	if err != nil {
		return nil, errcode.ErrRecordNotFound
	}
	return doc, nil
}

// Create creates a new document.
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
		s.logger.Error("failed to create document", zap.Error(err))
		return nil, errcode.ErrDatabase
	}
	return doc, nil
}

// Update updates an existing document.
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
		s.logger.Error("failed to update document", zap.Int64("id", id), zap.Error(err))
		return nil, errcode.ErrDatabase
	}
	return doc, nil
}

// Delete soft-deletes a document.
func (s *DocumentService) Delete(id int64) error {
	if _, err := s.repo.GetByID(id); err != nil {
		return errcode.ErrRecordNotFound
	}
	if err := s.repo.Delete(id); err != nil {
		s.logger.Error("failed to delete document", zap.Int64("id", id), zap.Error(err))
		return errcode.ErrDatabase
	}
	return nil
}
