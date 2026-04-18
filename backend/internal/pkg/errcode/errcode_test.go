package errcode

import (
	"net/http"
	"testing"
)

// TestErrCodeError tests error code implements the error interface.
func TestErrCodeError(t *testing.T) {
	err := ErrInvalidCredentials
	if err.Error() != "invalid username or password" {
		t.Errorf("Error() returned incorrect value: %s", err.Error())
	}
}

// TestWithMessage tests custom error message override.
func TestWithMessage(t *testing.T) {
	original := ErrInvalidParams
	custom := original.WithMessage("custom message")

	// Custom message should take effect
	if custom.Message != "custom message" {
		t.Errorf("custom message not applied: %s", custom.Message)
	}

	// Code should remain unchanged
	if custom.Code != original.Code {
		t.Errorf("Code should not change: %d != %d", custom.Code, original.Code)
	}

	// HTTP status code should remain unchanged
	if custom.HTTP != original.HTTP {
		t.Errorf("HTTP status code should not change: %d != %d", custom.HTTP, original.HTTP)
	}

	// Original error should not be modified
	if original.Message != "invalid parameters" {
		t.Error("original error should not be modified")
	}
}

// TestHTTPStatusMapping tests error code to HTTP status code mapping.
func TestHTTPStatusMapping(t *testing.T) {
	tests := []struct {
		name   string
		err    *ErrCode
		status int
	}{
		{"invalid credentials", ErrInvalidCredentials, http.StatusUnauthorized},
		{"account disabled", ErrAccountDisabled, http.StatusForbidden},
		{"access denied", ErrForbidden, http.StatusForbidden},
		{"invalid params", ErrInvalidParams, http.StatusBadRequest},
		{"username exists", ErrUsernameExists, http.StatusConflict},
		{"token quota exceeded", ErrTokenQuotaExceeded, http.StatusTooManyRequests},
		{"internal error", ErrInternal, http.StatusInternalServerError},
		{"LLM unavailable", ErrLLMUnavailable, http.StatusBadGateway},
		{"API key not copyable", ErrAPIKeyNotCopyable, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.HTTP != tt.status {
				t.Errorf("%s: HTTP status = %d, expected = %d", tt.name, tt.err.HTTP, tt.status)
			}
		})
	}
}
