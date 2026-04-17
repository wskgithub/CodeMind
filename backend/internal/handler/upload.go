package handler

import (
	"codemind/internal/pkg/response"
	"codemind/internal/service"

	"github.com/gin-gonic/gin"
)

// UploadHandler 文件上传处理器
type UploadHandler struct {
	svc *service.UploadService
}

// NewUploadHandler 创建上传处理器
func NewUploadHandler(svc *service.UploadService) *UploadHandler {
	return &UploadHandler{svc: svc}
}

// UploadImage 处理文档图片上传
func (h *UploadHandler) UploadImage(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请选择要上传的文件")
		return
	}

	result, err := h.svc.UploadImage(file)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, result)
}
