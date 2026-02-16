package validator

import (
	"regexp"
	"unicode"
)

// 密码强度正则：至少 8 位，包含大写字母、小写字母和数字
var passwordRegex = regexp.MustCompile(`^[a-zA-Z\d@$!%*?&]{8,128}$`)

// ValidatePassword 验证密码强度
// 规则：最少 8 位，包含大小写字母和数字
func ValidatePassword(password string) (bool, string) {
	if len(password) < 8 {
		return false, "密码长度不能少于 8 位"
	}
	if len(password) > 128 {
		return false, "密码长度不能超过 128 位"
	}

	var hasUpper, hasLower, hasDigit bool
	for _, c := range password {
		switch {
		case unicode.IsUpper(c):
			hasUpper = true
		case unicode.IsLower(c):
			hasLower = true
		case unicode.IsDigit(c):
			hasDigit = true
		}
	}

	if !hasUpper {
		return false, "密码必须包含大写字母"
	}
	if !hasLower {
		return false, "密码必须包含小写字母"
	}
	if !hasDigit {
		return false, "密码必须包含数字"
	}

	return true, ""
}

// ValidateUsername 验证用户名格式
// 规则：2-50 位，只能包含字母、数字、下划线
func ValidateUsername(username string) (bool, string) {
	if len(username) < 2 || len(username) > 50 {
		return false, "用户名长度必须在 2-50 位之间"
	}

	for _, c := range username {
		if !unicode.IsLetter(c) && !unicode.IsDigit(c) && c != '_' {
			return false, "用户名只能包含字母、数字和下划线"
		}
	}

	return true, ""
}
