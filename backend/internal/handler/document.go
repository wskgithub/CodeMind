package handler

import (
	"codemind/internal/model/dto"
	"codemind/internal/pkg/errcode"
	"codemind/internal/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// DocumentHandler 文档请求处理器.
type DocumentHandler struct {
	svc DocumentService
}

// NewDocumentHandler 创建文档处理器.
func NewDocumentHandler(svc DocumentService) *DocumentHandler {
	return &DocumentHandler{svc: svc}
}

// ListDocuments 获取已发布的文档列表（用户端）.
func (h *DocumentHandler) ListDocuments(c *gin.Context) {
	docs, err := h.svc.List()
	if err != nil {
		response.InternalError(c)
		return
	}
	response.Success(c, docs)
}

// GetDocument 根据 slug 获取文档详情（用户端）.
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

// ListAllDocuments 获取全部文档（管理端）.
func (h *DocumentHandler) ListAllDocuments(c *gin.Context) {
	docs, err := h.svc.ListAll()
	if err != nil {
		response.InternalError(c)
		return
	}
	response.Success(c, docs)
}

// GetDocumentByID 根据 ID 获取文档（管理端）.
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

// CreateDocument 创建文档（管理端）.
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

// UpdateDocument 更新文档（管理端）.
func (h *DocumentHandler) UpdateDocument(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid document ID")
		return
	}

	var req dto.UpdateDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
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

// DeleteDocument 删除文档（管理端）.
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
