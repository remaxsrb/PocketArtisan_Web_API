package login

import "errors"

var (
	ErrUsernameNotFound = errors.New("username not found")
	ErrInvalidPassword  = errors.New("invalid password")
)
