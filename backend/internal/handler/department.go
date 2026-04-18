package handler

import (
	"codemind/internal/middleware"
	"codemind/internal/model/dto"
	"codemind/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// DepartmentHandler handles department management endpoints.
type DepartmentHandler struct {
	deptService DepartmentService
}

// NewDepartmentHandler creates a department handler.
func NewDepartmentHandler(deptService DepartmentService) *DepartmentHandler {
	return &DepartmentHandler{deptService: deptService}
}

// List returns departments as a tree structure.
func (h *DepartmentHandler) List(c *gin.Context) {
	tree, err := h.deptService.ListTree()
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, tree)
}

// Create creates a new department.
func (h *DepartmentHandler) Create(c *gin.Context) {
	var req dto.CreateDepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	operatorID := middleware.GetUserID(c)

	dept, err := h.deptService.Create(&req, operatorID, c.ClientIP())
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, dept)
}

// GetDetail returns department details.
func (h *DepartmentHandler) GetDetail(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "invalid department ID")
		return
	}

	dept, err := h.deptService.GetByID(id)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, dept)
}

// Update updates department information.
func (h *DepartmentHandler) Update(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "invalid department ID")
		return
	}

	var req dto.UpdateDepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request")
		return
	}

	operatorID := middleware.GetUserID(c)

	if err := h.deptService.Update(id, &req, operatorID, c.ClientIP()); err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, nil)
}

// Delete deletes a department.
func (h *DepartmentHandler) Delete(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "invalid department ID")
		return
	}

	operatorID := middleware.GetUserID(c)

	if err := h.deptService.Delete(id, operatorID, c.ClientIP()); err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, nil)
}
