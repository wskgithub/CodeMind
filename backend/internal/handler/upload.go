package handler

import (
	"codemind/internal/pkg/response"
	"codemind/internal/service"

	"github.com/gin-gonic/gin"
)

// UploadHandler handles file upload requests.
type UploadHandler struct {
	svc *service.UploadService
}

// NewUploadHandler creates a new upload handler.
func NewUploadHandler(svc *service.UploadService) *UploadHandler {
	return &UploadHandler{svc: svc}
}

// UploadImage handles document image uploads.
func (h *UploadHandler) UploadImage(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "no file selected for upload")
		return
	}

	result, err := h.svc.UploadImage(file)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, result)
}
