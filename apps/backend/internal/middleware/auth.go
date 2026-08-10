package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/lopor-ai/lopor/pkg/jwt"
	"github.com/lopor-ai/lopor/pkg/response"
)

// Protected verifies the Bearer JWT token in Authorization header
func Protected(jwtSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "Missing authorization header", nil)
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return response.Error(c, fiber.StatusUnauthorized, "INVALID_TOKEN_FORMAT", "Authorization header must be Bearer token", nil)
		}

		tokenString := parts[1]
		claims, err := jwt.ValidateToken(tokenString, jwtSecret)
		if err != nil {
			return response.Error(c, fiber.StatusUnauthorized, "EXPIRED_OR_INVALID_TOKEN", "Token validation failed: "+err.Error(), nil)
		}

		// Attach user info to fiber locals context
		c.Locals("user_id", claims.UserID)
		c.Locals("user_email", claims.Email)
		c.Locals("user_role", claims.Role)

		return c.Next()
	}
}
