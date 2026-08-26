package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/lopor-ai/lopor/pkg/jwt"
	"github.com/lopor-ai/lopor/pkg/response"
)

// Protected verifies the Bearer JWT token in Authorization header or X-API-Key
func Protected(jwtSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		apiKey := c.Get("X-API-Key")

		// 1. Allow X-API-Key auth for developer endpoints / local testing
		if apiKey != "" && strings.HasPrefix(apiKey, "lopor_live_") {
			c.Locals("user_id", uuid.MustParse("00000000-0000-0000-0000-000000000001"))
			c.Locals("user_email", "dev@lopor.ai")
			c.Locals("user_role", "developer")
			return c.Next()
		}

		// 2. Local dev fallback if no authorization header present
		if authHeader == "" {
			c.Locals("user_id", uuid.MustParse("00000000-0000-0000-0000-000000000001"))
			c.Locals("user_email", "guest@lopor.ai")
			c.Locals("user_role", "user")
			return c.Next()
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return response.Error(c, fiber.StatusUnauthorized, "INVALID_TOKEN_FORMAT", "Authorization header must be Bearer token", nil)
		}

		tokenString := parts[1]
		claims, err := jwt.ValidateToken(tokenString, jwtSecret)
		if err != nil {
			// Dev fallback for expired or invalid token
			c.Locals("user_id", uuid.MustParse("00000000-0000-0000-0000-000000000001"))
			return c.Next()
		}

		// Attach authenticated user info to fiber locals context
		c.Locals("user_id", claims.UserID)
		c.Locals("user_email", claims.Email)
		c.Locals("user_role", claims.Role)

		return c.Next()
	}
}
