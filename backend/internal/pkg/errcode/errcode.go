package errcode

import "net/http"

// ErrCode represents a business error code
type ErrCode struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	HTTP    int    `json:"-"`
}

// Error implements the error interface
func (e *ErrCode) Error() string {
	return e.Message
}

// WithMessage creates a copy with a custom message
func (e *ErrCode) WithMessage(msg string) *ErrCode {
	return &ErrCode{
		Code:    e.Code,
		Message: msg,
		HTTP:    e.HTTP,
	}
}

var Success = &ErrCode{Code: 0, Message: "success", HTTP: http.StatusOK}

// Authentication errors (40001-40099)
var (
	ErrInvalidCredentials = &ErrCode{Code: 40001, Message: "invalid username or password", HTTP: http.StatusUnauthorized}
	ErrTokenExpired       = &ErrCode{Code: 40002, Message: "token expired", HTTP: http.StatusUnauthorized}
	ErrTokenInvalid       = &ErrCode{Code: 40003, Message: "invalid token", HTTP: http.StatusUnauthorized}
	ErrAccountDisabled    = &ErrCode{Code: 40004, Message: "account disabled", HTTP: http.StatusForbidden}
	ErrAPIKeyInvalid      = &ErrCode{Code: 40005, Message: "invalid API key", HTTP: http.StatusUnauthorized}
	ErrAPIKeyExpired      = &ErrCode{Code: 40006, Message: "API key expired", HTTP: http.StatusUnauthorized}
	ErrAPIKeyDisabled     = &ErrCode{Code: 40007, Message: "API key disabled", HTTP: http.StatusForbidden}
	ErrAccountLocked      = &ErrCode{Code: 40008, Message: "account locked", HTTP: http.StatusForbidden}
)

// Permission errors (40100-40199)
var (
	ErrForbidden     = &ErrCode{Code: 40101, Message: "access denied", HTTP: http.StatusForbidden}
	ErrForbiddenUser = &ErrCode{Code: 40102, Message: "not authorized to manage this user", HTTP: http.StatusForbidden}
	ErrForbiddenDept = &ErrCode{Code: 40103, Message: "not authorized to manage this department", HTTP: http.StatusForbidden}
)

// Validation errors (40200-40299)
var (
	ErrInvalidParams = &ErrCode{Code: 40201, Message: "invalid parameters", HTTP: http.StatusBadRequest}
	ErrMissingParams = &ErrCode{Code: 40202, Message: "required parameters missing", HTTP: http.StatusBadRequest}
)

// Business logic errors (40300-40399)
var (
	ErrUsernameExists             = &ErrCode{Code: 40301, Message: "username already exists", HTTP: http.StatusConflict}
	ErrEmailExists                = &ErrCode{Code: 40302, Message: "email already in use", HTTP: http.StatusConflict}
	ErrDeptNotFound               = &ErrCode{Code: 40303, Message: "department not found", HTTP: http.StatusNotFound}
	ErrAPIKeyLimit                = &ErrCode{Code: 40304, Message: "API key limit reached", HTTP: http.StatusConflict}
	ErrDeptHasUsers               = &ErrCode{Code: 40305, Message: "cannot delete department with users", HTTP: http.StatusConflict}
	ErrUserNotFound               = &ErrCode{Code: 40306, Message: "user not found", HTTP: http.StatusNotFound}
	ErrOldPasswordWrong           = &ErrCode{Code: 40307, Message: "incorrect current password", HTTP: http.StatusBadRequest}
	ErrAPIKeyNotFound             = &ErrCode{Code: 40308, Message: "API key not found", HTTP: http.StatusNotFound}
	ErrRecordNotFound             = &ErrCode{Code: 40309, Message: "record not found", HTTP: http.StatusNotFound}
	ErrProviderNameExists         = &ErrCode{Code: 40310, Message: "provider name already exists", HTTP: http.StatusConflict}
	ErrProviderTemplateNameExists = &ErrCode{Code: 40311, Message: "template name already exists", HTTP: http.StatusConflict}
	ErrAPIKeyNotCopyable          = &ErrCode{Code: 40312, Message: "this API key cannot be copied, please delete and recreate", HTTP: http.StatusBadRequest}
)

// Rate limit errors (42900-42999)
var (
	ErrTokenQuotaExceeded  = &ErrCode{Code: 42901, Message: "token quota exceeded", HTTP: http.StatusTooManyRequests}
	ErrConcurrencyExceeded = &ErrCode{Code: 42902, Message: "concurrency limit exceeded", HTTP: http.StatusTooManyRequests}
	ErrRateLimitExceeded   = &ErrCode{Code: 42903, Message: "rate limit exceeded", HTTP: http.StatusTooManyRequests}
)

// System errors (50000-50099)
var (
	ErrInternal       = &ErrCode{Code: 50001, Message: "internal server error", HTTP: http.StatusInternalServerError}
	ErrLLMUnavailable = &ErrCode{Code: 50002, Message: "LLM service unavailable", HTTP: http.StatusBadGateway}
	ErrDatabase       = &ErrCode{Code: 50003, Message: "database error", HTTP: http.StatusInternalServerError}
)
