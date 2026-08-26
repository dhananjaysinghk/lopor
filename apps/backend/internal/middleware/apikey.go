package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/lopor-ai/lopor/pkg/response"
)

// APIKeyAuth validates public developer API requests matching X-API-Key header
func APIKeyAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		apiKey := c.Get("X-API-Key")
		if apiKey == "" {
			// Fallback check query parameter
			apiKey = c.Query("api_key")
		}

		if apiKey == "" || !strings.HasPrefix(apiKey, "lopor_live_") {
			return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED_API_KEY", "Invalid or missing X-API-Key header", nil)
		}

		c.Locals("authenticated_via", "api_key")
		return c.Next()
	}
}
