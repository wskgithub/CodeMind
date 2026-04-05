package errcode

import "net/http"

// ErrCode 业务错误码结构
type ErrCode struct {
	Code    int    `json:"code"`    // 业务错误码
	Message string `json:"message"` // 错误描述
	HTTP    int    `json:"-"`       // 对应的 HTTP 状态码
}

// Error 实现 error 接口
func (e *ErrCode) Error() string {
	return e.Message
}

// WithMessage 创建带自定义消息的错误码副本
func (e *ErrCode) WithMessage(msg string) *ErrCode {
	return &ErrCode{
		Code:    e.Code,
		Message: msg,
		HTTP:    e.HTTP,
	}
}

// ──────────────────────────────────
// 成功
// ──────────────────────────────────

var Success = &ErrCode{Code: 0, Message: "success", HTTP: http.StatusOK}

// ──────────────────────────────────
// 认证相关错误 (40001-40099)
// ──────────────────────────────────

var (
	ErrInvalidCredentials = &ErrCode{Code: 40001, Message: "用户名或密码错误", HTTP: http.StatusUnauthorized}
	ErrTokenExpired       = &ErrCode{Code: 40002, Message: "Token 已过期", HTTP: http.StatusUnauthorized}
	ErrTokenInvalid       = &ErrCode{Code: 40003, Message: "Token 无效", HTTP: http.StatusUnauthorized}
	ErrAccountDisabled    = &ErrCode{Code: 40004, Message: "账号已被禁用", HTTP: http.StatusForbidden}
	ErrAPIKeyInvalid      = &ErrCode{Code: 40005, Message: "API Key 无效", HTTP: http.StatusUnauthorized}
	ErrAPIKeyExpired      = &ErrCode{Code: 40006, Message: "API Key 已过期", HTTP: http.StatusUnauthorized}
	ErrAPIKeyDisabled     = &ErrCode{Code: 40007, Message: "API Key 已被禁用", HTTP: http.StatusForbidden}
	ErrAccountLocked      = &ErrCode{Code: 40008, Message: "账号已被锁定", HTTP: http.StatusForbidden}
)

// ──────────────────────────────────
// 权限相关错误 (40100-40199)
// ──────────────────────────────────

var (
	ErrForbidden          = &ErrCode{Code: 40101, Message: "无权访问该资源", HTTP: http.StatusForbidden}
	ErrForbiddenUser      = &ErrCode{Code: 40102, Message: "无权操作该用户", HTTP: http.StatusForbidden}
	ErrForbiddenDept      = &ErrCode{Code: 40103, Message: "无权操作该部门", HTTP: http.StatusForbidden}
)

// ──────────────────────────────────
// 参数校验错误 (40200-40299)
// ──────────────────────────────────

var (
	ErrInvalidParams      = &ErrCode{Code: 40201, Message: "请求参数错误", HTTP: http.StatusBadRequest}
	ErrMissingParams      = &ErrCode{Code: 40202, Message: "必填参数缺失", HTTP: http.StatusBadRequest}
)

// ──────────────────────────────────
// 业务逻辑错误 (40300-40399)
// ──────────────────────────────────

var (
	ErrUsernameExists     = &ErrCode{Code: 40301, Message: "用户名已存在", HTTP: http.StatusConflict}
	ErrEmailExists        = &ErrCode{Code: 40302, Message: "邮箱已被使用", HTTP: http.StatusConflict}
	ErrDeptNotFound       = &ErrCode{Code: 40303, Message: "部门不存在", HTTP: http.StatusNotFound}
	ErrAPIKeyLimit        = &ErrCode{Code: 40304, Message: "API Key 数量已达上限", HTTP: http.StatusConflict}
	ErrDeptHasUsers       = &ErrCode{Code: 40305, Message: "部门下还有用户，无法删除", HTTP: http.StatusConflict}
	ErrUserNotFound       = &ErrCode{Code: 40306, Message: "用户不存在", HTTP: http.StatusNotFound}
	ErrOldPasswordWrong              = &ErrCode{Code: 40307, Message: "原密码错误", HTTP: http.StatusBadRequest}
	ErrAPIKeyNotFound                = &ErrCode{Code: 40308, Message: "API Key 不存在", HTTP: http.StatusNotFound}
	ErrRecordNotFound                = &ErrCode{Code: 40309, Message: "记录不存在", HTTP: http.StatusNotFound}
	ErrProviderNameExists            = &ErrCode{Code: 40310, Message: "服务名称已存在", HTTP: http.StatusConflict}
	ErrProviderTemplateNameExists    = &ErrCode{Code: 40311, Message: "模板名称已存在", HTTP: http.StatusConflict}
)

// ──────────────────────────────────
// 限流/限额错误 (42900-42999)
// ──────────────────────────────────

var (
	ErrTokenQuotaExceeded   = &ErrCode{Code: 42901, Message: "Token 用量已达限额", HTTP: http.StatusTooManyRequests}
	ErrConcurrencyExceeded  = &ErrCode{Code: 42902, Message: "并发请求数已达上限", HTTP: http.StatusTooManyRequests}
	ErrRateLimitExceeded    = &ErrCode{Code: 42903, Message: "请求频率过快", HTTP: http.StatusTooManyRequests}
)

// ──────────────────────────────────
// 系统内部错误 (50000-50099)
// ──────────────────────────────────

var (
	ErrInternal           = &ErrCode{Code: 50001, Message: "系统内部错误", HTTP: http.StatusInternalServerError}
	ErrLLMUnavailable     = &ErrCode{Code: 50002, Message: "LLM 服务不可用", HTTP: http.StatusBadGateway}
	ErrDatabase           = &ErrCode{Code: 50003, Message: "数据库错误", HTTP: http.StatusInternalServerError}
)
