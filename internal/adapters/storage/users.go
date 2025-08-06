package storage

import (
	"context"
	"database/sql"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/TeDenis/bukhindor-backend/internal/domain"
	"go.uber.org/zap"
)

// CreateUser создает нового пользователя
func (s *Service) CreateUser(ctx context.Context, user *domain.User) error {
	query := `INSERT INTO users (id, email, name, password_hash, is_active, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`

	s.logger.Debug("Generated SQL query", zap.String("query", query), zap.Any("args", []interface{}{user.ID, user.Email, user.Name, user.PasswordHash, user.IsActive, user.CreatedAt, user.UpdatedAt}))

	_, err := s.db.ExecContext(ctx, query, user.ID, user.Email, user.Name, user.PasswordHash, user.IsActive, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		s.logger.Error("Failed to create user", zap.Error(err), zap.String("email", user.Email))
		return err
	}

	s.logger.Info("User created successfully", zap.String("user_id", user.ID), zap.String("email", user.Email))
	return nil
}

// GetUserByID получает пользователя по ID
func (s *Service) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	query, args, err := squirrel.Select("id", "email", "name", "password_hash", "is_active", "created_at", "updated_at").
		From("users").
		Where(squirrel.Eq{"id": id}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()

	if err != nil {
		s.logger.Error("Failed to build get user by ID query", zap.Error(err))
		return nil, err
	}

	var user domain.User
	err = s.db.QueryRowContext(ctx, query, args...).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.PasswordHash,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			s.logger.Debug("User not found", zap.String("user_id", id))
			return nil, domain.ErrUserNotFound
		}
		s.logger.Error("Failed to get user by ID", zap.Error(err), zap.String("user_id", id))
		return nil, err
	}

	return &user, nil
}

// GetUserByEmail получает пользователя по email
func (s *Service) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `SELECT id, email, name, password_hash, is_active, created_at, updated_at FROM users WHERE email = $1`

	s.logger.Debug("Generated SQL query for GetUserByEmail", zap.String("query", query), zap.String("email", email))

	var user domain.User
	err := s.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.PasswordHash,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			s.logger.Debug("User not found", zap.String("email", email))
			return nil, domain.ErrUserNotFound
		}
		s.logger.Error("Failed to get user by email", zap.Error(err), zap.String("email", email))
		return nil, err
	}

	return &user, nil
}

// UpdateUser обновляет пользователя
func (s *Service) UpdateUser(ctx context.Context, user *domain.User) error {
	user.UpdatedAt = time.Now()

	query, args, err := squirrel.Update("users").
		Set("email", user.Email).
		Set("name", user.Name).
		Set("is_active", user.IsActive).
		Set("updated_at", user.UpdatedAt.Format("2006-01-02 15:04:05")).
		Where(squirrel.Eq{"id": user.ID}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()

	if err != nil {
		s.logger.Error("Failed to build update user query", zap.Error(err))
		return err
	}

	result, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		s.logger.Error("Failed to update user", zap.Error(err), zap.String("user_id", user.ID))
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		s.logger.Error("Failed to get rows affected", zap.Error(err))
		return err
	}

	if rowsAffected == 0 {
		s.logger.Debug("User not found for update", zap.String("user_id", user.ID))
		return domain.ErrUserNotFound
	}

	s.logger.Info("User updated successfully", zap.String("user_id", user.ID))
	return nil
}

// DeleteUser удаляет пользователя
func (s *Service) DeleteUser(ctx context.Context, id string) error {
	query, args, err := squirrel.Delete("users").
		Where(squirrel.Eq{"id": id}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()

	if err != nil {
		s.logger.Error("Failed to build delete user query", zap.Error(err))
		return err
	}

	result, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		s.logger.Error("Failed to delete user", zap.Error(err), zap.String("user_id", id))
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		s.logger.Error("Failed to get rows affected", zap.Error(err))
		return err
	}

	if rowsAffected == 0 {
		s.logger.Debug("User not found for deletion", zap.String("user_id", id))
		return domain.ErrUserNotFound
	}

	s.logger.Info("User deleted successfully", zap.String("user_id", id))
	return nil
}

// UpdatePassword обновляет пароль пользователя
func (s *Service) UpdatePassword(ctx context.Context, userID, passwordHash string) error {
	query, args, err := squirrel.Update("users").
		Set("password_hash", passwordHash).
		Set("updated_at", time.Now().Format("2006-01-02 15:04:05")).
		Where(squirrel.Eq{"id": userID}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()

	if err != nil {
		s.logger.Error("Failed to build update password query", zap.Error(err))
		return err
	}

	result, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		s.logger.Error("Failed to update password", zap.Error(err), zap.String("user_id", userID))
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		s.logger.Error("Failed to get rows affected", zap.Error(err))
		return err
	}

	if rowsAffected == 0 {
		s.logger.Debug("User not found for password update", zap.String("user_id", userID))
		return domain.ErrUserNotFound
	}

	s.logger.Info("Password updated successfully", zap.String("user_id", userID))
	return nil
}
