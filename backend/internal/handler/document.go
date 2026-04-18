package handler

import (
	"strconv"

	"codemind/internal/model/dto"
	"codemind/internal/pkg/errcode"
	"codemind/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// DocumentHandler handles document-related requests.
type DocumentHandler struct {
	svc DocumentService
}

// NewDocumentHandler creates a new document handler.
func NewDocumentHandler(svc DocumentService) *DocumentHandler {
	return &DocumentHandler{svc: svc}
}

// ListDocuments returns the list of published documents (user-facing).
func (h *DocumentHandler) ListDocuments(c *gin.Context) {
	docs, err := h.svc.List()
	if err != nil {
		response.InternalError(c)
		return
	}
	response.Success(c, docs)
}

// GetDocument retrieves a document by slug (user-facing).
func (h *DocumentHandler) GetDocument(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		response.BadRequest(c, "document slug cannot be empty")
		return
	}

	doc, err := h.svc.GetBySlug(slug)
	if err != nil {
		response.ErrorWithMsg(c, errcode.ErrRecordNotFound, "document not found or unpublished")
		return
	}
	response.Success(c, doc)
}

// ListAllDocuments returns all documents (admin).
func (h *DocumentHandler) ListAllDocuments(c *gin.Context) {
	docs, err := h.svc.ListAll()
	if err != nil {
		response.InternalError(c)
		return
	}
	response.Success(c, docs)
}

// GetDocumentByID retrieves a document by ID (admin).
func (h *DocumentHandler) GetDocumentByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid document ID")
		return
	}

	doc, err := h.svc.GetByID(id)
	if err != nil {
		response.ErrorWithMsg(c, errcode.ErrRecordNotFound, "document not found")
		return
	}
	response.Success(c, doc)
}

// CreateDocument creates a new document (admin).
func (h *DocumentHandler) CreateDocument(c *gin.Context) {
	var req dto.CreateDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	doc, err := h.svc.Create(&req)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.Success(c, doc)
}

// UpdateDocument updates an existing document (admin).
func (h *DocumentHandler) UpdateDocument(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid document ID")
		return
	}

	var req dto.UpdateDocumentRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	doc, err := h.svc.Update(id, &req)
	if err != nil {
		response.ErrorWithMsg(c, errcode.ErrRecordNotFound, "document not found")
		return
	}
	response.Success(c, doc)
}

// DeleteDocument deletes a document (admin).
func (h *DocumentHandler) DeleteDocument(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid document ID")
		return
	}

	if err := h.svc.Delete(id); err != nil {
		response.InternalError(c)
		return
	}
	response.Success(c, gin.H{"message": "deleted"})
}
