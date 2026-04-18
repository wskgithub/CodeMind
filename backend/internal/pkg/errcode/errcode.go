// Package errcode defines unified error codes.
package errcode

import "net/http"

// ErrCode represents a business error code.
type ErrCode struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
	HTTP    int    `json:"-"`
}

// Error implements the error interface.
func (e *ErrCode) Error() string {
	return e.Message
}

// WithMessage creates a copy with a custom message.
func (e *ErrCode) WithMessage(msg string) *ErrCode {
	return &ErrCode{
		Code:    e.Code,
		Message: msg,
		HTTP:    e.HTTP,
	}
}

// Success indicates a successful operation.
var Success = &ErrCode{Code: 0, Message: "success", HTTP: http.StatusOK}

// Authentication errors (40001-40099).
var (
	ErrInvalidCredentials = &ErrCode{Code: 40001, Message: "invalid username or password", HTTP: http.StatusUnauthorized} //nolint:mnd // intentional constant.
	ErrTokenExpired       = &ErrCode{Code: 40002, Message: "token expired", HTTP: http.StatusUnauthorized}                //nolint:mnd // intentional constant.
	ErrTokenInvalid       = &ErrCode{Code: 40003, Message: "invalid token", HTTP: http.StatusUnauthorized}                //nolint:mnd // intentional constant.
	ErrAccountDisabled    = &ErrCode{Code: 40004, Message: "account disabled", HTTP: http.StatusForbidden}                //nolint:mnd // intentional constant.
	ErrAPIKeyInvalid      = &ErrCode{Code: 40005, Message: "invalid API key", HTTP: http.StatusUnauthorized}              //nolint:mnd // intentional constant.
	ErrAPIKeyExpired      = &ErrCode{Code: 40006, Message: "API key expired", HTTP: http.StatusUnauthorized}              //nolint:mnd // intentional constant.
	ErrAPIKeyDisabled     = &ErrCode{Code: 40007, Message: "API key disabled", HTTP: http.StatusForbidden}                //nolint:mnd // intentional constant.
	ErrAccountLocked      = &ErrCode{Code: 40008, Message: "account locked", HTTP: http.StatusForbidden}                  //nolint:mnd // intentional constant.
)

// Permission errors (40100-40199).
var (
	ErrForbidden     = &ErrCode{Code: 40101, Message: "access denied", HTTP: http.StatusForbidden}                            //nolint:mnd // intentional constant.
	ErrForbiddenUser = &ErrCode{Code: 40102, Message: "not authorized to manage this user", HTTP: http.StatusForbidden}       //nolint:mnd // intentional constant.
	ErrForbiddenDept = &ErrCode{Code: 40103, Message: "not authorized to manage this department", HTTP: http.StatusForbidden} //nolint:mnd // intentional constant.
)

// Validation errors (40200-40299).
var (
	ErrInvalidParams = &ErrCode{Code: 40201, Message: "invalid parameters", HTTP: http.StatusBadRequest}          //nolint:mnd // intentional constant.
	ErrMissingParams = &ErrCode{Code: 40202, Message: "required parameters missing", HTTP: http.StatusBadRequest} //nolint:mnd // intentional constant.
)

// Business logic errors (40300-40399).
var (
	ErrUsernameExists             = &ErrCode{Code: 40301, Message: "username already exists", HTTP: http.StatusConflict}                                     //nolint:mnd // intentional constant.
	ErrEmailExists                = &ErrCode{Code: 40302, Message: "email already in use", HTTP: http.StatusConflict}                                        //nolint:mnd // intentional constant.
	ErrDeptNotFound               = &ErrCode{Code: 40303, Message: "department not found", HTTP: http.StatusNotFound}                                        //nolint:mnd // intentional constant.
	ErrAPIKeyLimit                = &ErrCode{Code: 40304, Message: "API key limit reached", HTTP: http.StatusConflict}                                       //nolint:mnd // intentional constant.
	ErrDeptHasUsers               = &ErrCode{Code: 40305, Message: "cannot delete department with users", HTTP: http.StatusConflict}                         //nolint:mnd // intentional constant.
	ErrUserNotFound               = &ErrCode{Code: 40306, Message: "user not found", HTTP: http.StatusNotFound}                                              //nolint:mnd // intentional constant.
	ErrOldPasswordWrong           = &ErrCode{Code: 40307, Message: "incorrect current password", HTTP: http.StatusBadRequest}                                //nolint:mnd // intentional constant.
	ErrAPIKeyNotFound             = &ErrCode{Code: 40308, Message: "API key not found", HTTP: http.StatusNotFound}                                           //nolint:mnd // intentional constant.
	ErrRecordNotFound             = &ErrCode{Code: 40309, Message: "record not found", HTTP: http.StatusNotFound}                                            //nolint:mnd // intentional constant.
	ErrProviderNameExists         = &ErrCode{Code: 40310, Message: "provider name already exists", HTTP: http.StatusConflict}                                //nolint:mnd // intentional constant.
	ErrProviderTemplateNameExists = &ErrCode{Code: 40311, Message: "template name already exists", HTTP: http.StatusConflict}                                //nolint:mnd // intentional constant.
	ErrAPIKeyNotCopyable          = &ErrCode{Code: 40312, Message: "this API key cannot be copied, please delete and recreate", HTTP: http.StatusBadRequest} //nolint:mnd // intentional constant.
)

// Rate limit errors (42900-42999).
var (
	ErrTokenQuotaExceeded  = &ErrCode{Code: 42901, Message: "token quota exceeded", HTTP: http.StatusTooManyRequests}       //nolint:mnd // intentional constant.
	ErrConcurrencyExceeded = &ErrCode{Code: 42902, Message: "concurrency limit exceeded", HTTP: http.StatusTooManyRequests} //nolint:mnd // intentional constant.
	ErrRateLimitExceeded   = &ErrCode{Code: 42903, Message: "rate limit exceeded", HTTP: http.StatusTooManyRequests}        //nolint:mnd // intentional constant.
)

// System errors (50000-50099).
var (
	ErrInternal       = &ErrCode{Code: 50001, Message: "internal server error", HTTP: http.StatusInternalServerError} //nolint:mnd // intentional constant.
	ErrLLMUnavailable = &ErrCode{Code: 50002, Message: "LLM service unavailable", HTTP: http.StatusBadGateway}        //nolint:mnd // intentional constant.
	ErrDatabase       = &ErrCode{Code: 50003, Message: "database error", HTTP: http.StatusInternalServerError}        //nolint:mnd // intentional constant.
)
