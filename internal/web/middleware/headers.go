package middleware

import (
	"github.com/TeDenis/bukhindor-backend/internal/app"
	"github.com/gofiber/fiber/v2"
)

// ValidateHeaders middleware проверяет наличие обязательных заголовков
func ValidateHeaders() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Проверяем наличие обязательных заголовков
		appVersion := c.Get(app.HeaderAppVersion)
		appType := c.Get(app.HeaderAppType)
		deviceID := c.Get(app.HeaderDeviceID)

		if appVersion == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Missing required header: " + app.HeaderAppVersion,
				"code":  fiber.StatusBadRequest,
			})
		}

		if appType == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Missing required header: " + app.HeaderAppType,
				"code":  fiber.StatusBadRequest,
			})
		}

		if !app.ValidateAppType(appType) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid app type. Must be one of: " + app.AppTypeIOS + ", " + app.AppTypeAndroid + ", " + app.AppTypeWeb,
				"code":  fiber.StatusBadRequest,
			})
		}

		if deviceID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Missing required header: " + app.HeaderDeviceID,
				"code":  fiber.StatusBadRequest,
			})
		}

		// Сохраняем заголовки в контексте для использования в handlers
		c.Locals("app_version", appVersion)
		c.Locals("app_type", appType)
		c.Locals("device_id", deviceID)

		return c.Next()
	}
}
