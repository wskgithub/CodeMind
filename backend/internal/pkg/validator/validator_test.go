package validator

import "testing"

// TestValidatePassword tests password strength validation.
func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"empty password", "", true},
		{"too short (less than 8 chars)", "Ab1@56", true},
		{"missing uppercase letter", "test@12345", true},
		{"missing lowercase letter", "TEST@12345", true},
		{"missing digit", "Test@abcde", true},
		{"missing special character", "Test12345a", false},
		{"valid password", "Test@12345", false},
		{"valid password (multiple special chars)", "P@ssw0rd!#", false},
		{"password too long", "A1@" + string(make([]byte, 200)), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, errMsg := ValidatePassword(tt.password)
			if valid == tt.wantErr { // valid=true means wantErr should be false
				t.Errorf("ValidatePassword(%q) result = %v, message = %s, expected wantErr = %v", tt.password, valid, errMsg, tt.wantErr)
			}
		})
	}
}

// TestValidateUsername tests username format validation.
func TestValidateUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		wantErr  bool
	}{
		{"empty username", "", true},
		{"too short", "a", true},
		{"normal username", "admin", false},
		{"with underscore", "user_test", false},
		{"with hyphen", "user-test", true},
		{"digits only", "12345", false},
		{"starts with digit", "1admin", false},
		{"contains space", "user name", true},
		{"contains Chinese", "用户名", false},
		{"contains special character", "user@name", true},
		{"minimum 2 chars", "ab", false},
		{"exceeds max 50 chars", string(make([]byte, 51)), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, errMsg := ValidateUsername(tt.username)
			if valid == tt.wantErr { // valid=true means wantErr should be false
				t.Errorf("ValidateUsername(%q) result = %v, message = %s, expected wantErr = %v", tt.username, valid, errMsg, tt.wantErr)
			}
		})
	}
}
