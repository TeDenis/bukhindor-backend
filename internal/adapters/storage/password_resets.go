package storage

import (
	"context"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/TeDenis/bukhindor-backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

// CreatePasswordReset создает новый запрос на сброс пароля
func (s *Service) CreatePasswordReset(ctx context.Context, reset *domain.PasswordReset) error {
	query, args, err := squirrel.Insert("password_resets").
		Columns("id", "user_id", "token", "expires_at", "used", "created_at").
		Values(reset.ID, reset.UserID, reset.Token, reset.ExpiresAt.Format("2006-01-02 15:04:05"), reset.Used, reset.CreatedAt.Format("2006-01-02 15:04:05")).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()

	if err != nil {
		s.logger.Error("Failed to build create password reset query", zap.Error(err))
		return err
	}

	_, err = s.db.Exec(ctx, query, args...)
	if err != nil {
		s.logger.Error("Failed to create password reset", zap.Error(err), zap.String("user_id", reset.UserID))
		return err
	}

	s.logger.Info("Password reset created successfully", zap.String("reset_id", reset.ID), zap.String("user_id", reset.UserID))
	return nil
}

// GetPasswordResetByToken получает запрос на сброс пароля по токену
func (s *Service) GetPasswordResetByToken(ctx context.Context, token string) (*domain.PasswordReset, error) {
	query, args, err := squirrel.Select("id", "user_id", "token", "expires_at", "used", "created_at").
		From("password_resets").
		Where(squirrel.Eq{"token": token}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()

	if err != nil {
		s.logger.Error("Failed to build get password reset by token query", zap.Error(err))
		return nil, err
	}

	var reset domain.PasswordReset
	err = s.db.QueryRow(ctx, query, args...).Scan(
		&reset.ID,
		&reset.UserID,
		&reset.Token,
		&reset.ExpiresAt,
		&reset.Used,
		&reset.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			s.logger.Debug("Password reset not found", zap.String("token", token))
			return nil, domain.ErrUserNotFound
		}
		s.logger.Error("Failed to get password reset by token", zap.Error(err), zap.String("token", token))
		return nil, err
	}

	return &reset, nil
}

// MarkPasswordResetAsUsed помечает запрос на сброс пароля как использованный
func (s *Service) MarkPasswordResetAsUsed(ctx context.Context, id string) error {
	query, args, err := squirrel.Update("password_resets").
		Set("used", true).
		Where(squirrel.Eq{"id": id}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()

	if err != nil {
		s.logger.Error("Failed to build mark password reset as used query", zap.Error(err))
		return err
	}

	tag, err := s.db.Exec(ctx, query, args...)
	if err != nil {
		s.logger.Error("Failed to mark password reset as used", zap.Error(err), zap.String("reset_id", id))
		return err
	}

	rowsAffected := int64(tag.RowsAffected())

	if rowsAffected == 0 {
		s.logger.Debug("Password reset not found for marking as used", zap.String("reset_id", id))
		return domain.ErrUserNotFound
	}

	s.logger.Info("Password reset marked as used", zap.String("reset_id", id))
	return nil
}

// DeleteExpiredPasswordResets удаляет истекшие запросы на сброс пароля
func (s *Service) DeleteExpiredPasswordResets(ctx context.Context) error {
	query, args, err := squirrel.Delete("password_resets").
		Where(squirrel.Lt{"expires_at": time.Now().Format("2006-01-02 15:04:05")}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()

	if err != nil {
		s.logger.Error("Failed to build delete expired password resets query", zap.Error(err))
		return err
	}

	tag, err := s.db.Exec(ctx, query, args...)
	if err != nil {
		s.logger.Error("Failed to delete expired password resets", zap.Error(err))
		return err
	}

	rowsAffected := int64(tag.RowsAffected())

	if rowsAffected > 0 {
		s.logger.Info("Expired password resets deleted", zap.Int64("count", rowsAffected))
	}

	return nil
}
