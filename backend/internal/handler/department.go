package handler

import (
	"codemind/internal/middleware"
	"codemind/internal/model/dto"
	"codemind/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// DepartmentHandler 部门管理控制器
type DepartmentHandler struct {
	deptService DepartmentService
}

// NewDepartmentHandler 创建部门 Handler
func NewDepartmentHandler(deptService DepartmentService) *DepartmentHandler {
	return &DepartmentHandler{deptService: deptService}
}

// List 获取部门列表（树形结构）
// GET /api/v1/departments
func (h *DepartmentHandler) List(c *gin.Context) {
	tree, err := h.deptService.ListTree()
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, tree)
}

// Create 创建部门
// POST /api/v1/departments
func (h *DepartmentHandler) Create(c *gin.Context) {
	var req dto.CreateDepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误: "+err.Error())
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

// GetDetail 获取部门详情
// GET /api/v1/departments/:id
func (h *DepartmentHandler) GetDetail(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "无效的部门 ID")
		return
	}

	dept, err := h.deptService.GetByID(id)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, dept)
}

// Update 更新部门信息
// PUT /api/v1/departments/:id
func (h *DepartmentHandler) Update(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "无效的部门 ID")
		return
	}

	var req dto.UpdateDepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误")
		return
	}

	operatorID := middleware.GetUserID(c)

	if err := h.deptService.Update(id, &req, operatorID, c.ClientIP()); err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, nil)
}

// Delete 删除部门
// DELETE /api/v1/departments/:id
func (h *DepartmentHandler) Delete(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "无效的部门 ID")
		return
	}

	operatorID := middleware.GetUserID(c)

	if err := h.deptService.Delete(id, operatorID, c.ClientIP()); err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, nil)
}
