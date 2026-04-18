// Package validator provides input validation utilities.
package validator

import (
	"unicode"
)

const (
	minPasswordLength = 8   //nolint:mnd // minimum password length
	maxPasswordLength = 128 //nolint:mnd // maximum password length
)

// ValidatePassword checks password strength (min 8 chars, requires upper, lower, digit).
func ValidatePassword(password string) (bool, string) {
	if len(password) < minPasswordLength {
		return false, "password must be at least 8 characters"
	}
	if len(password) > maxPasswordLength {
		return false, "password cannot exceed 128 characters"
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
		return false, "password must contain an uppercase letter"
	}
	if !hasLower {
		return false, "password must contain a lowercase letter"
	}
	if !hasDigit {
		return false, "password must contain a digit"
	}

	return true, ""
}

// ValidateUsername checks username format (2-50 chars, alphanumeric + underscore).
func ValidateUsername(username string) (bool, string) {
	if len(username) < 2 || len(username) > 50 {
		return false, "username must be 2-50 characters"
	}

	for _, c := range username {
		if !unicode.IsLetter(c) && !unicode.IsDigit(c) && c != '_' {
			return false, "username can only contain letters, digits, and underscores"
		}
	}

	return true, ""
}
