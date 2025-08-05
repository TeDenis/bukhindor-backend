package auth

import (
	"github.com/TeDenis/bukhindor-backend/internal/config"
	"go.uber.org/zap"
)

// Service представляет сервис аутентификации
type Service struct {
	userRepo          UserRepository
	sessionRepo       SessionRepository
	passwordResetRepo PasswordResetRepository
	redisRepo         RedisRepository
	config            *config.Config
	logger            *zap.Logger
}

// NewService создает новый сервис аутентификации
func NewService(
	userRepo UserRepository,
	sessionRepo SessionRepository,
	passwordResetRepo PasswordResetRepository,
	redisRepo RedisRepository,
	cfg *config.Config,
	logger *zap.Logger,
) *Service {
	return &Service{
		userRepo:          userRepo,
		sessionRepo:       sessionRepo,
		passwordResetRepo: passwordResetRepo,
		redisRepo:         redisRepo,
		config:            cfg,
		logger:            logger,
	}
}
