package api

import (
	"github.com/TeDenis/bukhindor-backend/internal/config"
	"github.com/TeDenis/bukhindor-backend/internal/service/auth"
	"github.com/TeDenis/bukhindor-backend/internal/web/middleware"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// Service представляет API сервис
type Service struct {
	config      *config.Config
	logger      *zap.Logger
	authService *auth.Service
}

// NewService создает новый API сервис
func NewService(cfg *config.Config, logger *zap.Logger, authService *auth.Service) *Service {
	return &Service{
		config:      cfg,
		logger:      logger,
		authService: authService,
	}
}

// SetupRoutes настраивает API роуты
func (s *Service) SetupRoutes(app *fiber.App) {
	// API группа с валидацией заголовков
	api := app.Group("/api/v1", middleware.ValidateHeaders())

	// Аутентификация (без авторизации)
	auth := api.Group("/auth")
	auth.Post("/login", s.login)
	auth.Post("/register", s.register)
	auth.Post("/reset-password", s.resetPassword)
	auth.Post("/refresh", s.refreshTokens)

	// Защищенные роуты (с авторизацией)
	// Применяем JWT только к конкретному маршруту, чтобы не требовать токен на public-ручках
	auth.Get("/me", middleware.JWTAuth(s.config, s.logger), s.getCurrentUser)

	// Пользователи (защищенные)
	users := api.Group("/users", middleware.JWTAuth(s.config, s.logger))
	users.Get("/", s.getUsers)
	users.Post("/", s.createUser)
	users.Get("/:id", s.getUser)
	users.Put("/:id", s.updateUser)
	users.Delete("/:id", s.deleteUser)
}

// login выполняет аутентификацию пользователя
// @Summary Войти в систему
// @Description Аутентифицирует пользователя и возвращает токен
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body LoginRequest true "Данные для входа"
// @Success 200 {object} LoginResponse "Успешная аутентификация"
// @Failure 400 {object} ErrorResponse "Ошибка валидации"
// @Failure 401 {object} ErrorResponse "Неверные учетные данные"
// @Failure 500 {object} ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/auth/login [post]
func (s *Service) login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		s.logger.Warn("Failed to parse login request", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
			"code":  fiber.StatusBadRequest,
		})
	}

	input := auth.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	}

	tokens, err := s.authService.Login(c.Context(), input)
	if err != nil {
		s.logger.Warn("Login failed", zap.Error(err), zap.String("email", req.Email))
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
			"code":  fiber.StatusUnauthorized,
		})
	}

	// Устанавливаем куки с access токеном
	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    tokens.AccessToken,
		HTTPOnly: true,
		Secure:   s.config.ServerHost != "localhost", // Secure только для продакшена
		SameSite: "Lax",
		MaxAge:   int(s.config.JWTExpiration.Seconds()),
	})

	return c.JSON(fiber.Map{
		"message": "Login successful",
		"user": fiber.Map{
			"access_token":  tokens.AccessToken,
			"refresh_token": tokens.RefreshToken,
		},
	})
}

// register регистрирует нового пользователя
// @Summary Зарегистрироваться
// @Description Регистрирует нового пользователя в системе
// @Tags auth
// @Accept json
// @Produce json
// @Param user body RegisterRequest true "Данные для регистрации"
// @Success 201 {object} RegisterResponse "Пользователь зарегистрирован"
// @Failure 400 {object} ErrorResponse "Ошибка валидации"
// @Failure 409 {object} ErrorResponse "Пользователь уже существует"
// @Failure 500 {object} ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/auth/register [post]
func (s *Service) register(c *fiber.Ctx) error {
	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		s.logger.Warn("Failed to parse register request", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
			"code":  fiber.StatusBadRequest,
		})
	}

	input := auth.RegisterInput{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	}

	user, err := s.authService.Register(c.Context(), input)
	if err != nil {
		s.logger.Warn("Registration failed", zap.Error(err), zap.String("email", req.Email))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
			"code":  fiber.StatusBadRequest,
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User registered successfully",
		"user": fiber.Map{
			"id":    user.ID,
			"email": user.Email,
			"name":  user.Name,
		},
	})
}

// resetPassword создает запрос на сброс пароля
// @Summary Запросить сброс пароля
// @Description Создает запрос на сброс пароля для указанного email
// @Tags auth
// @Accept json
// @Produce json
// @Param request body ResetPasswordRequest true "Email для сброса пароля"
// @Success 200 {object} MessageResponse "Запрос на сброс пароля создан"
// @Failure 400 {object} ErrorResponse "Ошибка валидации"
// @Failure 500 {object} ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/auth/reset-password [post]
func (s *Service) resetPassword(c *fiber.Ctx) error {
	var req ResetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		s.logger.Warn("Failed to parse reset password request", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
			"code":  fiber.StatusBadRequest,
		})
	}

	input := auth.ResetPasswordInput{
		Email: req.Email,
	}

	err := s.authService.RequestPasswordReset(c.Context(), input)
	if err != nil {
		s.logger.Warn("Password reset request failed", zap.Error(err), zap.String("email", req.Email))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
			"code":  fiber.StatusBadRequest,
		})
	}

	return c.JSON(fiber.Map{
		"message": "Password reset request created successfully",
	})
}

// getCurrentUser возвращает текущего пользователя
// @Summary Получить текущего пользователя
// @Description Возвращает информацию о текущем авторизованном пользователе
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} UserResponse "Информация о пользователе"
// @Failure 401 {object} ErrorResponse "Не авторизован"
// @Failure 500 {object} ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/auth/me [get]
func (s *Service) getCurrentUser(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	// TODO: Получить пользователя из БД через UserService
	// Пока возвращаем заглушку
	return c.JSON(fiber.Map{
		"user": fiber.Map{
			"id":    userID,
			"email": "user@example.com", // TODO: получить из БД
			"name":  "User Name",        // TODO: получить из БД
		},
	})
}

// getUsers возвращает список пользователей
// @Summary Получить список пользователей
// @Description Возвращает список всех пользователей в системе
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {array} UserResponse "Список пользователей"
// @Failure 500 {object} ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/users [get]
func (s *Service) getUsers(c *fiber.Ctx) error {
	// TODO: Реализовать получение пользователей
	return c.JSON(fiber.Map{
		"users": []fiber.Map{},
		"total": 0,
	})
}

// createUser создает нового пользователя
// @Summary Создать пользователя
// @Description Создает нового пользователя в системе
// @Tags users
// @Accept json
// @Produce json
// @Param user body CreateUserRequest true "Данные пользователя"
// @Success 201 {object} UserResponse "Пользователь создан"
// @Failure 400 {object} ErrorResponse "Ошибка валидации"
// @Failure 500 {object} ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/users [post]
func (s *Service) createUser(c *fiber.Ctx) error {
	// TODO: Реализовать создание пользователя
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User created successfully",
	})
}

// getUser возвращает пользователя по ID
// @Summary Получить пользователя
// @Description Возвращает пользователя по его ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "ID пользователя"
// @Success 200 {object} UserResponse "Пользователь найден"
// @Failure 404 {object} ErrorResponse "Пользователь не найден"
// @Failure 500 {object} ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/users/{id} [get]
func (s *Service) getUser(c *fiber.Ctx) error {
	// TODO: Реализовать получение пользователя
	id := c.Params("id")
	return c.JSON(fiber.Map{
		"id":      id,
		"message": "User details",
	})
}

// updateUser обновляет пользователя
// @Summary Обновить пользователя
// @Description Обновляет данные пользователя
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "ID пользователя"
// @Param user body UpdateUserRequest true "Данные для обновления"
// @Success 200 {object} UserResponse "Пользователь обновлен"
// @Failure 400 {object} ErrorResponse "Ошибка валидации"
// @Failure 404 {object} ErrorResponse "Пользователь не найден"
// @Failure 500 {object} ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/users/{id} [put]
func (s *Service) updateUser(c *fiber.Ctx) error {
	// TODO: Реализовать обновление пользователя
	id := c.Params("id")
	return c.JSON(fiber.Map{
		"id":      id,
		"message": "User updated successfully",
	})
}

// deleteUser удаляет пользователя
// @Summary Удалить пользователя
// @Description Удаляет пользователя из системы
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "ID пользователя"
// @Success 200 {object} MessageResponse "Пользователь удален"
// @Failure 404 {object} ErrorResponse "Пользователь не найден"
// @Failure 500 {object} ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/users/{id} [delete]
func (s *Service) deleteUser(c *fiber.Ctx) error {
	// TODO: Реализовать удаление пользователя
	id := c.Params("id")
	return c.JSON(fiber.Map{
		"message": "User deleted successfully",
		"id":      id,
	})
}

// refreshTokens обновляет access токен используя refresh токен
// @Summary Обновить токены
// @Description Обновляет access токен используя refresh токен
// @Tags auth
// @Accept json
// @Produce json
// @Param tokens body RefreshTokenRequest true "Refresh токен"
// @Success 200 {object} RefreshTokenResponse "Токены обновлены"
// @Failure 400 {object} ErrorResponse "Ошибка валидации"
// @Failure 401 {object} ErrorResponse "Неверный refresh токен"
// @Failure 500 {object} ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/auth/refresh [post]
func (s *Service) refreshTokens(c *fiber.Ctx) error {
	var req RefreshTokenRequest
	if err := c.BodyParser(&req); err != nil {
		s.logger.Warn("Failed to parse refresh token request", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
			"code":  fiber.StatusBadRequest,
		})
	}

	input := auth.RefreshTokensInput{
		RefreshToken: req.RefreshToken,
	}

	tokens, err := s.authService.RefreshTokens(c.Context(), input)
	if err != nil {
		s.logger.Warn("Token refresh failed", zap.Error(err))
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
			"code":  fiber.StatusUnauthorized,
		})
	}

	// Устанавливаем куки с новым access токеном
	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    tokens.AccessToken,
		HTTPOnly: true,
		Secure:   s.config.ServerHost != "localhost", // Secure только для продакшена
		SameSite: "Lax",
		MaxAge:   int(s.config.JWTExpiration.Seconds()),
	})

	return c.JSON(fiber.Map{
		"message": "Tokens refreshed successfully",
		"tokens": fiber.Map{
			"access_token":  tokens.AccessToken,
			"refresh_token": tokens.RefreshToken,
		},
	})
}
