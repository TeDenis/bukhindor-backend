package auth

import (
	"context"
	"time"

	"github.com/TeDenis/bukhindor-backend/internal/app"
	"github.com/TeDenis/bukhindor-backend/internal/domain"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// LoginInput представляет входные данные для входа
type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterInput представляет входные данные для регистрации
type RegisterInput struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// ResetPasswordInput представляет входные данные для сброса пароля
type ResetPasswordInput struct {
	Email string `json:"email"`
}

// Login выполняет аутентификацию пользователя
func (s *Service) Login(ctx context.Context, input LoginInput) (*domain.AuthTokens, error) {
	// Валидация входных данных
	if !app.ValidateEmail(input.Email) {
		s.logger.Warn("Invalid email format", zap.String("email", input.Email))
		return nil, app.ErrInvalidInput
	}

	if !app.ValidatePassword(input.Password) {
		s.logger.Warn("Invalid password format")
		return nil, app.ErrInvalidInput
	}

	// Получаем пользователя по email
	user, err := s.userRepo.GetUserByEmail(ctx, input.Email)
	if err != nil {
		s.logger.Warn("User not found during login", zap.String("email", input.Email))
		return nil, app.ErrInvalidCredentials
	}

	// Проверяем активность пользователя
	if !user.IsActive {
		s.logger.Warn("Inactive user attempted login", zap.String("user_id", user.ID))
		return nil, app.ErrInvalidCredentials
	}

	// Проверяем пароль
	if !app.CheckPasswordHash(input.Password, user.PasswordHash) {
		s.logger.Warn("Invalid password for user", zap.String("user_id", user.ID))
		return nil, app.ErrInvalidCredentials
	}

	// Генерируем токены
	tokens, err := s.generateTokens(user.ID)
	if err != nil {
		s.logger.Error("Failed to generate tokens", zap.Error(err), zap.String("user_id", user.ID))
		return nil, app.ErrInternalServer
	}

	// Сохраняем refresh токен в Redis
	err = s.redisRepo.SetRefreshToken(ctx, user.ID, tokens.RefreshToken, s.config.RefreshTokenExpiration)
	if err != nil {
		s.logger.Error("Failed to save refresh token", zap.Error(err), zap.String("user_id", user.ID))
		return nil, app.ErrInternalServer
	}

	// Создаем сессию в БД
	session := &domain.UserSession{
		ID:        app.GenerateUUID(),
		UserID:    user.ID,
		TokenHash: app.HashToken(tokens.RefreshToken), // Хешируем refresh токен для БД
		ExpiresAt: time.Now().Add(s.config.RefreshTokenExpiration),
		CreatedAt: time.Now(),
	}

	err = s.sessionRepo.CreateSession(ctx, session)
	if err != nil {
		s.logger.Error("Failed to create session", zap.Error(err), zap.String("user_id", user.ID))
		// Удаляем refresh токен из Redis если не удалось создать сессию
		s.redisRepo.DeleteRefreshToken(ctx, user.ID)
		return nil, app.ErrInternalServer
	}

	s.logger.Info("User logged in successfully", zap.String("user_id", user.ID), zap.String("email", user.Email))
	return tokens, nil
}

// Register регистрирует нового пользователя
func (s *Service) Register(ctx context.Context, input RegisterInput) (*domain.User, error) {
	// Валидация входных данных
	if !app.ValidateName(input.Name) {
		s.logger.Warn("Invalid name format", zap.String("name", input.Name))
		return nil, app.ErrInvalidInput
	}

	if !app.ValidateEmail(input.Email) {
		s.logger.Warn("Invalid email format", zap.String("email", input.Email))
		return nil, app.ErrInvalidInput
	}

	if !app.ValidatePassword(input.Password) {
		s.logger.Warn("Invalid password format")
		return nil, app.ErrInvalidInput
	}

	// Проверяем, существует ли пользователь с таким email
	existingUser, err := s.userRepo.GetUserByEmail(ctx, input.Email)
	if err == nil && existingUser != nil {
		s.logger.Warn("User already exists", zap.String("email", input.Email))
		return nil, app.ErrUserExists
	}

	// Хешируем пароль
	passwordHash, err := app.HashPassword(input.Password)
	if err != nil {
		s.logger.Error("Failed to hash password", zap.Error(err))
		return nil, app.ErrInternalServer
	}

	// Создаем пользователя
	user := &domain.User{
		ID:           app.GenerateUUID(),
		Email:        input.Email,
		Name:         input.Name,
		PasswordHash: passwordHash,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err = s.userRepo.CreateUser(ctx, user)
	if err != nil {
		s.logger.Error("Failed to create user", zap.Error(err), zap.String("email", input.Email))
		return nil, app.ErrInternalServer
	}

	s.logger.Info("User registered successfully", zap.String("user_id", user.ID), zap.String("email", user.Email))
	return user, nil
}

// RequestPasswordReset создает запрос на сброс пароля
func (s *Service) RequestPasswordReset(ctx context.Context, input ResetPasswordInput) error {
	// Валидация email
	if !app.ValidateEmail(input.Email) {
		s.logger.Warn("Invalid email format", zap.String("email", input.Email))
		return app.ErrInvalidInput
	}

	// Получаем пользователя по email
	user, err := s.userRepo.GetUserByEmail(ctx, input.Email)
	if err != nil {
		// Не раскрываем информацию о существовании пользователя
		s.logger.Debug("User not found for password reset", zap.String("email", input.Email))
		return nil // Возвращаем успех даже если пользователь не найден
	}

	// Проверяем активность пользователя
	if !user.IsActive {
		s.logger.Debug("Inactive user requested password reset", zap.String("user_id", user.ID))
		return nil // Возвращаем успех даже для неактивных пользователей
	}

	// Генерируем токен для сброса пароля
	token, err := app.GenerateRandomToken(app.PasswordResetTokenLength)
	if err != nil {
		s.logger.Error("Failed to generate password reset token", zap.Error(err), zap.String("user_id", user.ID))
		return app.ErrInternalServer
	}

	// Создаем запрос на сброс пароля
	reset := &domain.PasswordReset{
		ID:        app.GenerateUUID(),
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(time.Duration(app.PasswordResetExpiration) * time.Hour),
		Used:      false,
		CreatedAt: time.Now(),
	}

	err = s.passwordResetRepo.CreatePasswordReset(ctx, reset)
	if err != nil {
		s.logger.Error("Failed to create password reset", zap.Error(err), zap.String("user_id", user.ID))
		return app.ErrInternalServer
	}

	// TODO: Отправить email с токеном для сброса пароля
	// В реальном приложении здесь должна быть отправка email

	s.logger.Info("Password reset requested", zap.String("user_id", user.ID), zap.String("email", user.Email))
	return nil
}

// generateTokens генерирует пару токенов (access и refresh)
func (s *Service) generateTokens(userID string) (*domain.AuthTokens, error) {
	// Генерируем access токен
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(s.config.JWTExpiration).Unix(),
		"iat":     time.Now().Unix(),
		"type":    "access",
	})

	accessTokenString, err := accessToken.SignedString([]byte(s.config.JWTSecret))
	if err != nil {
		return nil, err
	}

	// Генерируем refresh токен
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(s.config.RefreshTokenExpiration).Unix(),
		"iat":     time.Now().Unix(),
		"type":    "refresh",
	})

	refreshTokenString, err := refreshToken.SignedString([]byte(s.config.JWTSecret))
	if err != nil {
		return nil, err
	}

	return &domain.AuthTokens{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
	}, nil
}
