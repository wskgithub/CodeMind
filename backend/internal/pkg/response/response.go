package response

import (
	"net/http"

	"codemind/internal/pkg/errcode"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// Pagination 分页信息
type Pagination struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// PageData 分页响应数据
type PageData struct {
	List       interface{} `json:"list"`
	Pagination Pagination  `json:"pagination"`
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// SuccessWithPage 成功响应（带分页）
func SuccessWithPage(c *gin.Context, list interface{}, total int64, page, pageSize int) {
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data: PageData{
			List: list,
			Pagination: Pagination{
				Page:       page,
				PageSize:   pageSize,
				Total:      total,
				TotalPages: totalPages,
			},
		},
	})
}

// Error 错误响应（使用预定义错误码）
func Error(c *gin.Context, err *errcode.ErrCode) {
	c.JSON(err.HTTP, Response{
		Code:    err.Code,
		Message: err.Message,
		Data:    nil,
	})
}

// ErrorWithMsg 错误响应（自定义消息）
func ErrorWithMsg(c *gin.Context, err *errcode.ErrCode, msg string) {
	c.JSON(err.HTTP, Response{
		Code:    err.Code,
		Message: msg,
		Data:    nil,
	})
}

// BadRequest 参数错误的快捷方法
func BadRequest(c *gin.Context, msg string) {
	c.JSON(http.StatusBadRequest, Response{
		Code:    errcode.ErrInvalidParams.Code,
		Message: msg,
		Data:    nil,
	})
}

// Unauthorized 未认证的快捷方法
func Unauthorized(c *gin.Context, msg string) {
	c.JSON(http.StatusUnauthorized, Response{
		Code:    errcode.ErrTokenInvalid.Code,
		Message: msg,
		Data:    nil,
	})
}

// Forbidden 无权限的快捷方法
func Forbidden(c *gin.Context, msg string) {
	c.JSON(http.StatusForbidden, Response{
		Code:    errcode.ErrForbidden.Code,
		Message: msg,
		Data:    nil,
	})
}

// InternalError 内部错误的快捷方法
func InternalError(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, Response{
		Code:    errcode.ErrInternal.Code,
		Message: errcode.ErrInternal.Message,
		Data:    nil,
	})
}
