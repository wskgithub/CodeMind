package errcode

import (
	"net/http"
	"testing"
)

// TestErrCodeError 测试错误码实现 error 接口
func TestErrCodeError(t *testing.T) {
	err := ErrInvalidCredentials
	if err.Error() != "用户名或密码错误" {
		t.Errorf("Error() 返回值不正确: %s", err.Error())
	}
}

// TestWithMessage 测试自定义消息
func TestWithMessage(t *testing.T) {
	original := ErrInvalidParams
	custom := original.WithMessage("自定义消息")

	// 自定义消息应生效
	if custom.Message != "自定义消息" {
		t.Errorf("自定义消息未生效: %s", custom.Message)
	}

	// Code 应保持不变
	if custom.Code != original.Code {
		t.Errorf("Code 不应改变: %d != %d", custom.Code, original.Code)
	}

	// HTTP 状态码应保持不变
	if custom.HTTP != original.HTTP {
		t.Errorf("HTTP 状态码不应改变: %d != %d", custom.HTTP, original.HTTP)
	}

	// 原始错误不应被修改
	if original.Message != "请求参数错误" {
		t.Error("原始错误不应被修改")
	}
}

// TestHTTPStatusMapping 测试错误码到 HTTP 状态码映射
func TestHTTPStatusMapping(t *testing.T) {
	tests := []struct {
		name   string
		err    *ErrCode
		status int
	}{
		{"认证失败", ErrInvalidCredentials, http.StatusUnauthorized},
		{"账号禁用", ErrAccountDisabled, http.StatusForbidden},
		{"无权访问", ErrForbidden, http.StatusForbidden},
		{"参数错误", ErrInvalidParams, http.StatusBadRequest},
		{"用户名已存在", ErrUsernameExists, http.StatusConflict},
		{"Token 超限", ErrTokenQuotaExceeded, http.StatusTooManyRequests},
		{"内部错误", ErrInternal, http.StatusInternalServerError},
		{"LLM 不可用", ErrLLMUnavailable, http.StatusBadGateway},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.HTTP != tt.status {
				t.Errorf("%s: HTTP 状态码 = %d, 预期 = %d", tt.name, tt.err.HTTP, tt.status)
			}
		})
	}
}
