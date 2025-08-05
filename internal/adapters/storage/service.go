package storage

import (
	"database/sql"

	"github.com/TeDenis/bukhindor-backend/internal/config"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Service представляет storage сервис
type Service struct {
	db     *sql.DB
	redis  *redis.Client
	config *config.Config
	logger *zap.Logger
}

// NewService создает новый storage сервис
func NewService(db *sql.DB, redis *redis.Client, cfg *config.Config, logger *zap.Logger) *Service {
	return &Service{
		db:     db,
		redis:  redis,
		config: cfg,
		logger: logger,
	}
}
