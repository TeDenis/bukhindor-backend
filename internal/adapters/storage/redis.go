package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/TeDenis/bukhindor-backend/internal/app"
	"go.uber.org/zap"
)

// SetRefreshToken сохраняет refresh токен в Redis
func (s *Service) SetRefreshToken(ctx context.Context, userID, refreshToken string, expiration time.Duration) error {
	key := fmt.Sprintf("%s%s", app.RefreshTokenPrefix, userID)

	err := s.redis.Set(ctx, key, refreshToken, expiration).Err()
	if err != nil {
		s.logger.Error("Failed to set refresh token in Redis", zap.Error(err), zap.String("user_id", userID))
		return err
	}

	s.logger.Debug("Refresh token set in Redis", zap.String("user_id", userID))
	return nil
}

// GetRefreshToken получает refresh токен из Redis
func (s *Service) GetRefreshToken(ctx context.Context, userID string) (string, error) {
	key := fmt.Sprintf("%s%s", app.RefreshTokenPrefix, userID)

	token, err := s.redis.Get(ctx, key).Result()
	if err != nil {
		s.logger.Debug("Refresh token not found in Redis", zap.Error(err), zap.String("user_id", userID))
		return "", err
	}

	s.logger.Debug("Refresh token retrieved from Redis", zap.String("user_id", userID))
	return token, nil
}

// DeleteRefreshToken удаляет refresh токен из Redis
func (s *Service) DeleteRefreshToken(ctx context.Context, userID string) error {
	key := fmt.Sprintf("%s%s", app.RefreshTokenPrefix, userID)

	err := s.redis.Del(ctx, key).Err()
	if err != nil {
		s.logger.Error("Failed to delete refresh token from Redis", zap.Error(err), zap.String("user_id", userID))
		return err
	}

	s.logger.Debug("Refresh token deleted from Redis", zap.String("user_id", userID))
	return nil
}

// DeleteAllUserRefreshTokens удаляет все refresh токены пользователя из Redis
func (s *Service) DeleteAllUserRefreshTokens(ctx context.Context, userID string) error {
	pattern := fmt.Sprintf("%s%s*", app.RefreshTokenPrefix, userID)

	keys, err := s.redis.Keys(ctx, pattern).Result()
	if err != nil {
		s.logger.Error("Failed to get refresh token keys from Redis", zap.Error(err), zap.String("user_id", userID))
		return err
	}

	if len(keys) > 0 {
		err = s.redis.Del(ctx, keys...).Err()
		if err != nil {
			s.logger.Error("Failed to delete refresh tokens from Redis", zap.Error(err), zap.String("user_id", userID))
			return err
		}
		s.logger.Debug("All refresh tokens deleted from Redis", zap.String("user_id", userID), zap.Int("count", len(keys)))
	}

	return nil
}
