package storage

import (
	"context"
	"database/sql"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/TeDenis/bukhindor-backend/internal/domain"
	"go.uber.org/zap"
)

// CreateSession создает новую сессию пользователя
func (s *Service) CreateSession(ctx context.Context, session *domain.UserSession) error {
	query, args, err := squirrel.Insert("user_sessions").
		Columns("id", "user_id", "token_hash", "expires_at", "created_at").
		Values(session.ID, session.UserID, session.TokenHash, session.ExpiresAt, session.CreatedAt).
		ToSql()

	if err != nil {
		s.logger.Error("Failed to build create session query", zap.Error(err))
		return err
	}

	_, err = s.db.ExecContext(ctx, query, args...)
	if err != nil {
		s.logger.Error("Failed to create session", zap.Error(err), zap.String("user_id", session.UserID))
		return err
	}

	s.logger.Info("Session created successfully", zap.String("session_id", session.ID), zap.String("user_id", session.UserID))
	return nil
}

// GetSessionByID получает сессию по ID
func (s *Service) GetSessionByID(ctx context.Context, id string) (*domain.UserSession, error) {
	query, args, err := squirrel.Select("id", "user_id", "token_hash", "expires_at", "created_at").
		From("user_sessions").
		Where(squirrel.Eq{"id": id}).
		ToSql()

	if err != nil {
		s.logger.Error("Failed to build get session by ID query", zap.Error(err))
		return nil, err
	}

	var session domain.UserSession
	err = s.db.QueryRowContext(ctx, query, args...).Scan(
		&session.ID,
		&session.UserID,
		&session.TokenHash,
		&session.ExpiresAt,
		&session.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			s.logger.Debug("Session not found", zap.String("session_id", id))
			return nil, domain.ErrUserNotFound
		}
		s.logger.Error("Failed to get session by ID", zap.Error(err), zap.String("session_id", id))
		return nil, err
	}

	return &session, nil
}

// GetSessionsByUserID получает все сессии пользователя
func (s *Service) GetSessionsByUserID(ctx context.Context, userID string) ([]*domain.UserSession, error) {
	query, args, err := squirrel.Select("id", "user_id", "token_hash", "expires_at", "created_at").
		From("user_sessions").
		Where(squirrel.Eq{"user_id": userID}).
		OrderBy("created_at DESC").
		ToSql()

	if err != nil {
		s.logger.Error("Failed to build get sessions by user ID query", zap.Error(err))
		return nil, err
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		s.logger.Error("Failed to get sessions by user ID", zap.Error(err), zap.String("user_id", userID))
		return nil, err
	}
	defer rows.Close()

	var sessions []*domain.UserSession
	for rows.Next() {
		var session domain.UserSession
		err := rows.Scan(
			&session.ID,
			&session.UserID,
			&session.TokenHash,
			&session.ExpiresAt,
			&session.CreatedAt,
		)
		if err != nil {
			s.logger.Error("Failed to scan session", zap.Error(err))
			return nil, err
		}
		sessions = append(sessions, &session)
	}

	if err = rows.Err(); err != nil {
		s.logger.Error("Failed to iterate sessions", zap.Error(err))
		return nil, err
	}

	return sessions, nil
}

// DeleteSession удаляет сессию
func (s *Service) DeleteSession(ctx context.Context, id string) error {
	query, args, err := squirrel.Delete("user_sessions").
		Where(squirrel.Eq{"id": id}).
		ToSql()

	if err != nil {
		s.logger.Error("Failed to build delete session query", zap.Error(err))
		return err
	}

	result, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		s.logger.Error("Failed to delete session", zap.Error(err), zap.String("session_id", id))
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		s.logger.Error("Failed to get rows affected", zap.Error(err))
		return err
	}

	if rowsAffected == 0 {
		s.logger.Debug("Session not found for deletion", zap.String("session_id", id))
		return domain.ErrUserNotFound
	}

	s.logger.Info("Session deleted successfully", zap.String("session_id", id))
	return nil
}

// DeleteExpiredSessions удаляет истекшие сессии
func (s *Service) DeleteExpiredSessions(ctx context.Context) error {
	query, args, err := squirrel.Delete("user_sessions").
		Where(squirrel.Lt{"expires_at": time.Now()}).
		ToSql()

	if err != nil {
		s.logger.Error("Failed to build delete expired sessions query", zap.Error(err))
		return err
	}

	result, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		s.logger.Error("Failed to delete expired sessions", zap.Error(err))
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		s.logger.Error("Failed to get rows affected", zap.Error(err))
		return err
	}

	if rowsAffected > 0 {
		s.logger.Info("Expired sessions deleted", zap.Int64("count", rowsAffected))
	}

	return nil
}
