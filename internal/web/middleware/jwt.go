package middleware

import (
	"strings"

	"github.com/TeDenis/bukhindor-backend/internal/app"
	"github.com/TeDenis/bukhindor-backend/internal/config"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// JWTAuth middleware проверяет JWT токен
func JWTAuth(cfg *config.Config, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Получаем токен из куки или заголовка Authorization
		tokenString := c.Cookies(app.JWTCookieName)

		if tokenString == "" {
			// Пробуем получить из заголовка Authorization
			authHeader := c.Get("Authorization")
			if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
				tokenString = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		if tokenString == "" {
			logger.Debug("No JWT token found")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "No token provided",
				"code":  fiber.StatusUnauthorized,
			})
		}

		// Парсим и валидируем токен
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Проверяем алгоритм подписи
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(cfg.JWTSecret), nil
		})

		if err != nil {
			logger.Debug("JWT token validation failed", zap.Error(err))
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token",
				"code":  fiber.StatusUnauthorized,
			})
		}

		// Проверяем, что токен валиден
		if !token.Valid {
			logger.Debug("JWT token is invalid")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token",
				"code":  fiber.StatusUnauthorized,
			})
		}

		// Извлекаем claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			logger.Debug("Failed to extract JWT claims")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token claims",
				"code":  fiber.StatusUnauthorized,
			})
		}

		// Проверяем тип токена
		tokenType, ok := claims["type"].(string)
		if !ok || tokenType != "access" {
			logger.Debug("Invalid token type", zap.String("type", tokenType))
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token type",
				"code":  fiber.StatusUnauthorized,
			})
		}

		// Извлекаем user_id
		userID, ok := claims["user_id"].(string)
		if !ok || userID == "" {
			logger.Debug("No user_id in JWT claims")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token claims",
				"code":  fiber.StatusUnauthorized,
			})
		}

		// Сохраняем user_id в контексте
		c.Locals("user_id", userID)

		logger.Debug("JWT authentication successful", zap.String("user_id", userID))
		return c.Next()
	}
}
