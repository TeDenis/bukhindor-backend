package app

import "errors"

// Общие ошибки приложения
var (
	ErrInvalidInput         = errors.New("invalid input data")
	ErrUserNotFound         = errors.New("user not found")
	ErrUserExists           = errors.New("user already exists")
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrUnauthorized         = errors.New("unauthorized")
	ErrForbidden            = errors.New("forbidden")
	ErrInternalServer       = errors.New("internal server error")
	ErrInvalidToken         = errors.New("invalid token")
	ErrTokenExpired         = errors.New("token expired")
	ErrPasswordResetExpired = errors.New("password reset token expired")
	ErrPasswordResetUsed    = errors.New("password reset token already used")
	ErrMissingHeaders       = errors.New("missing required headers")
	ErrInvalidAppType       = errors.New("invalid app type")
)
