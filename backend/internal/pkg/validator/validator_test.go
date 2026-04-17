package validator

import "testing"

// TestValidatePassword 测试密码强度验证.
func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"空密码", "", true},
		{"太短（少于8位）", "Ab1@56", true},
		{"缺少大写字母", "test@12345", true},
		{"缺少小写字母", "TEST@12345", true},
		{"缺少数字", "Test@abcde", true},
		{"缺少特殊字符", "Test12345a", false},
		{"合格密码", "Test@12345", false},
		{"合格密码（含多种特殊字符）", "P@ssw0rd!#", false},
		{"超长密码", "A1@" + string(make([]byte, 200)), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, errMsg := ValidatePassword(tt.password)
			if valid == tt.wantErr { // valid 为 true 时 wantErr 应为 false
				t.Errorf("ValidatePassword(%q) 结果 = %v, 消息 = %s, 预期 wantErr = %v", tt.password, valid, errMsg, tt.wantErr)
			}
		})
	}
}

// TestValidateUsername 测试用户名格式验证.
func TestValidateUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		wantErr  bool
	}{
		{"空用户名", "", true},
		{"太短", "a", true},
		{"正常用户名", "admin", false},
		{"带下划线", "user_test", false},
		{"带连字符", "user-test", true},
		{"纯数字", "12345", false},
		{"以数字开头", "1admin", false},
		{"含空格", "user name", true},
		{"含中文", "用户名", false},
		{"含特殊字符", "user@name", true},
		{"最短2位", "ab", false},
		{"最长50位", string(make([]byte, 51)), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, errMsg := ValidateUsername(tt.username)
			if valid == tt.wantErr { // valid 为 true 时 wantErr 应为 false
				t.Errorf("ValidateUsername(%q) 结果 = %v, 消息 = %s, 预期 wantErr = %v", tt.username, valid, errMsg, tt.wantErr)
			}
		})
	}
}
