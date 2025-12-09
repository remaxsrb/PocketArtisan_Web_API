package validator

import (
	"errors"
	"unicode"
)

var (
	ErrPwdTooShort      = errors.New("password must be at least 12 characters long")
	ErrPwdMissingUpper  = errors.New("password must contain at least one uppercase letter")
	ErrPwdMissingDigit  = errors.New("password must contain at least one digit")
	ErrPwdMissingSymbol = errors.New("password must contain at least one special character")
)

func IsValidPassword(pwd string) bool {
	return ValidatePassword(pwd) == nil
}

func ValidatePassword(pwd string) error {
	if len([]rune(pwd)) < 12 {
		return ErrPwdTooShort
	}

	var hasUpper, hasDigit, hasSymbol bool
	for _, r := range pwd {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSymbol = true
		}
		if hasUpper && hasDigit && hasSymbol {
			break
		}
	}

	if !hasUpper {
		return ErrPwdMissingUpper
	}
	if !hasDigit {
		return ErrPwdMissingDigit
	}
	if !hasSymbol {
		return ErrPwdMissingSymbol
	}
	return nil
}
