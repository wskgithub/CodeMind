package response

import (
	"net/http"

	"codemind/internal/pkg/errcode"

	"github.com/gin-gonic/gin"
)

// Response represents a unified API response structure.
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// Pagination represents pagination information.
type Pagination struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// PageData represents paginated response data.
type PageData struct {
	List       interface{} `json:"list"`
	Pagination Pagination  `json:"pagination"`
}

// Success sends a successful response.
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// SuccessWithPage sends a successful response with pagination.
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

// Error sends an error response using predefined error codes.
func Error(c *gin.Context, err *errcode.ErrCode) {
	c.JSON(err.HTTP, Response{
		Code:    err.Code,
		Message: err.Message,
		Data:    nil,
	})
}

// ErrorWithMsg sends an error response with a custom message.
func ErrorWithMsg(c *gin.Context, err *errcode.ErrCode, msg string) {
	c.JSON(err.HTTP, Response{
		Code:    err.Code,
		Message: msg,
		Data:    nil,
	})
}

// BadRequest sends a bad request error response.
func BadRequest(c *gin.Context, msg string) {
	c.JSON(http.StatusBadRequest, Response{
		Code:    errcode.ErrInvalidParams.Code,
		Message: msg,
		Data:    nil,
	})
}

// Unauthorized sends an unauthorized error response.
func Unauthorized(c *gin.Context, msg string) {
	c.JSON(http.StatusUnauthorized, Response{
		Code:    errcode.ErrTokenInvalid.Code,
		Message: msg,
		Data:    nil,
	})
}

// Forbidden sends a forbidden error response.
func Forbidden(c *gin.Context, msg string) {
	c.JSON(http.StatusForbidden, Response{
		Code:    errcode.ErrForbidden.Code,
		Message: msg,
		Data:    nil,
	})
}

// InternalError sends an internal server error response.
func InternalError(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, Response{
		Code:    errcode.ErrInternal.Code,
		Message: errcode.ErrInternal.Message,
		Data:    nil,
	})
}
