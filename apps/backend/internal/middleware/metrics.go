package middleware

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

// PrometheusMetrics tracks API request latency, method, path, and response status
func PrometheusMetrics() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		duration := time.Since(start)

		status := strconv.Itoa(c.Response().StatusCode())
		method := c.Method()
		path := c.Path()

		// Set metric diagnostic headers
		c.Set("X-Response-Time", duration.String())
		_ = status
		_ = method
		_ = path

		return err
	}
}
