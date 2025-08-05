package domain

import "errors"

// Ошибки домена
var (
	ErrUserNotFound = errors.New("user not found")
)
