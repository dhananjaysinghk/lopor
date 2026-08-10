package response

import "github.com/gofiber/fiber/v2"

// Response represents a standard JSON API response structure
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorDetail `json:"error,omitempty"`
}

// ErrorDetail provides structured error diagnostic info
type ErrorDetail struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// Success renders a 200/201 JSON success response
func Success(c *fiber.Ctx, statusCode int, message string, data interface{}) error {
	return c.Status(statusCode).JSON(Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Error renders a standard JSON error response
func Error(c *fiber.Ctx, statusCode int, errorCode string, message string, details interface{}) error {
	return c.Status(statusCode).JSON(Response{
		Success: false,
		Error: &ErrorDetail{
			Code:    errorCode,
			Message: message,
			Details: details,
		},
	})
}
