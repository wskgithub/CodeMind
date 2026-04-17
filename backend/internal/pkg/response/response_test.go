package response

import (
	"codemind/internal/pkg/errcode"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	return c, w
}

func TestSuccess(t *testing.T) {
	c, w := setupTestContext()

	testData := map[string]string{"key": "value"}
	Success(c, testData)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "success", resp.Message)

	dataMap, ok := resp.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "value", dataMap["key"])
}

func TestSuccessWithNilData(t *testing.T) {
	c, w := setupTestContext()

	Success(c, nil)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "success", resp.Message)
	assert.Nil(t, resp.Data)
}

func TestSuccessWithPage(t *testing.T) {
	c, w := setupTestContext()

	list := []string{"item1", "item2", "item3"}
	var total int64 = 25
	page := 2
	pageSize := 10

	SuccessWithPage(c, list, total, page, pageSize)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "success", resp.Message)

	pageData, ok := resp.Data.(map[string]interface{})
	assert.True(t, ok)

	// 验证分页信息
	pagination, ok := pageData["pagination"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, float64(page), pagination["page"])
	assert.Equal(t, float64(pageSize), pagination["page_size"])
	assert.Equal(t, float64(total), pagination["total"])
	assert.Equal(t, float64(3), pagination["total_pages"]) // 25/10 = 2.5 -> 3 pages

	// 验证列表数据
	listData, ok := pageData["list"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, listData, 3)
}

func TestSuccessWithPage_TotalPagesCalculation(t *testing.T) {
	tests := []struct {
		name      string
		total     int64
		pageSize  int
		wantPages int
	}{
		{"exact division", 20, 10, 2},
		{"with remainder", 21, 10, 3},
		{"single page", 5, 10, 1},
		{"empty list", 0, 10, 0},
		{"one item", 1, 10, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, w := setupTestContext()

			SuccessWithPage(c, []string{}, tt.total, 1, tt.pageSize)

			var resp Response
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			assert.NoError(t, err)

			pageData := resp.Data.(map[string]interface{})
			pagination := pageData["pagination"].(map[string]interface{})
			assert.Equal(t, float64(tt.wantPages), pagination["total_pages"])
		})
	}
}

func TestError(t *testing.T) {
	c, w := setupTestContext()

	testErr := &errcode.ErrCode{
		Code:    40001,
		Message: "用户名或密码错误",
		HTTP:    http.StatusUnauthorized,
	}

	Error(c, testErr)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 40001, resp.Code)
	assert.Equal(t, "用户名或密码错误", resp.Message)
	assert.Nil(t, resp.Data)
}

func TestErrorWithDifferentErrorCodes(t *testing.T) {
	tests := []struct {
		errCode  *errcode.ErrCode
		name     string
		wantMsg  string
		wantHTTP int
		wantCode int
	}{
		{
			name:     "internal error",
			errCode:  errcode.ErrInternal,
			wantHTTP: http.StatusInternalServerError,
			wantCode: 50001,
			wantMsg:  "系统内部错误",
		},
		{
			name:     "invalid params",
			errCode:  errcode.ErrInvalidParams,
			wantHTTP: http.StatusBadRequest,
			wantCode: 40201,
			wantMsg:  "请求参数错误",
		},
		{
			name:     "forbidden",
			errCode:  errcode.ErrForbidden,
			wantHTTP: http.StatusForbidden,
			wantCode: 40101,
			wantMsg:  "无权访问该资源",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, w := setupTestContext()

			Error(c, tt.errCode)

			assert.Equal(t, tt.wantHTTP, w.Code)

			var resp Response
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantCode, resp.Code)
			assert.Equal(t, tt.wantMsg, resp.Message)
		})
	}
}

func TestErrorWithMsg(t *testing.T) {
	c, w := setupTestContext()

	testErr := &errcode.ErrCode{
		Code:    40001,
		Message: "用户名或密码错误",
		HTTP:    http.StatusUnauthorized,
	}
	customMsg := "自定义错误消息"

	ErrorWithMsg(c, testErr, customMsg)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 40001, resp.Code)
	assert.Equal(t, customMsg, resp.Message)
	assert.Nil(t, resp.Data)
}

func TestBadRequest(t *testing.T) {
	c, w := setupTestContext()

	msg := "缺少必填参数 user_id"
	BadRequest(c, msg)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, errcode.ErrInvalidParams.Code, resp.Code)
	assert.Equal(t, msg, resp.Message)
	assert.Nil(t, resp.Data)
}

func TestUnauthorized(t *testing.T) {
	c, w := setupTestContext()

	msg := "请先登录"
	Unauthorized(c, msg)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, errcode.ErrTokenInvalid.Code, resp.Code)
	assert.Equal(t, msg, resp.Message)
	assert.Nil(t, resp.Data)
}

func TestForbidden(t *testing.T) {
	c, w := setupTestContext()

	msg := "您没有权限执行此操作"
	Forbidden(c, msg)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, errcode.ErrForbidden.Code, resp.Code)
	assert.Equal(t, msg, resp.Message)
	assert.Nil(t, resp.Data)
}

func TestInternalError(t *testing.T) {
	c, w := setupTestContext()

	InternalError(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, errcode.ErrInternal.Code, resp.Code)
	assert.Equal(t, errcode.ErrInternal.Message, resp.Message)
	assert.Nil(t, resp.Data)
}

func TestResponseStructures(t *testing.T) {
	t.Run("Response struct", func(t *testing.T) {
		resp := Response{
			Code:    0,
			Message: "test",
			Data:    map[string]string{"key": "value"},
		}
		assert.Equal(t, 0, resp.Code)
		assert.Equal(t, "test", resp.Message)
	})

	t.Run("Pagination struct", func(t *testing.T) {
		pagination := Pagination{
			Page:       1,
			PageSize:   10,
			Total:      100,
			TotalPages: 10,
		}
		assert.Equal(t, 1, pagination.Page)
		assert.Equal(t, 10, pagination.PageSize)
		assert.Equal(t, int64(100), pagination.Total)
		assert.Equal(t, 10, pagination.TotalPages)
	})

	t.Run("PageData struct", func(t *testing.T) {
		pageData := PageData{
			List: []string{"item1"},
			Pagination: Pagination{
				Page:       1,
				PageSize:   10,
				Total:      1,
				TotalPages: 1,
			},
		}
		assert.NotNil(t, pageData.List)
		assert.Equal(t, 1, pageData.Pagination.Page)
	})
}

func TestSuccessWithComplexData(t *testing.T) {
	tests := []struct {
		data interface{}
		name string
	}{
		{
			name: "string slice",
			data: []string{"a", "b", "c"},
		},
		{
			name: "int slice",
			data: []int{1, 2, 3},
		},
		{
			name: "complex map",
			data: map[string]interface{}{
				"name": "test",
				"age":  18,
				"tags": []string{"tag1", "tag2"},
			},
		},
		{
			name: "struct",
			data: struct {
				Name string `json:"name"`
				ID   int    `json:"id"`
			}{ID: 1, Name: "test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, w := setupTestContext()

			Success(c, tt.data)

			assert.Equal(t, http.StatusOK, w.Code)

			var resp Response
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			assert.NoError(t, err)
			assert.Equal(t, 0, resp.Code)
			assert.NotNil(t, resp.Data)
		})
	}
}
