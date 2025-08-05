package storage

import (
	"context"
	"time"

	"github.com/TeDenis/bukhindor-backend/internal/domain"
)

// UserRepository определяет интерфейс для работы с пользователями
type UserRepository interface {
	CreateUser(ctx context.Context, user *domain.User) error
	GetUserByID(ctx context.Context, id string) (*domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	UpdateUser(ctx context.Context, user *domain.User) error
	DeleteUser(ctx context.Context, id string) error
	UpdatePassword(ctx context.Context, userID, passwordHash string) error
}

// SessionRepository определяет интерфейс для работы с сессиями
type SessionRepository interface {
	CreateSession(ctx context.Context, session *domain.UserSession) error
	GetSessionByID(ctx context.Context, id string) (*domain.UserSession, error)
	GetSessionsByUserID(ctx context.Context, userID string) ([]*domain.UserSession, error)
	DeleteSession(ctx context.Context, id string) error
	DeleteExpiredSessions(ctx context.Context) error
}

// PasswordResetRepository определяет интерфейс для работы со сбросом паролей
type PasswordResetRepository interface {
	CreatePasswordReset(ctx context.Context, reset *domain.PasswordReset) error
	GetPasswordResetByToken(ctx context.Context, token string) (*domain.PasswordReset, error)
	MarkPasswordResetAsUsed(ctx context.Context, id string) error
	DeleteExpiredPasswordResets(ctx context.Context) error
}

// RedisRepository определяет интерфейс для работы с Redis
type RedisRepository interface {
	SetRefreshToken(ctx context.Context, userID, refreshToken string, expiration time.Duration) error
	GetRefreshToken(ctx context.Context, userID string) (string, error)
	DeleteRefreshToken(ctx context.Context, userID string) error
	DeleteAllUserRefreshTokens(ctx context.Context, userID string) error
}
