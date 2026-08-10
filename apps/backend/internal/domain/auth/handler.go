package auth

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/lopor-ai/lopor/pkg/response"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(c *fiber.Ctx) error {
	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Invalid request body", nil)
	}

	if req.Email == "" || req.Password == "" || req.FullName == "" {
		return response.Error(c, fiber.StatusBadRequest, "VALIDATION_FAILED", "Email, password, and full name are required", nil)
	}

	res, err := h.service.Register(c.Context(), req)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "REGISTRATION_FAILED", err.Error(), nil)
	}

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    res.RefreshToken,
		HTTPOnly: true,
		Secure:   false,
		SameSite: "Lax",
	})

	return response.Success(c, fiber.StatusCreated, "User registered successfully", res)
}

func (h *Handler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Invalid request body", nil)
	}

	ip := c.IP()
	res, err := h.service.Login(c.Context(), req, ip)
	if err != nil {
		return response.Error(c, fiber.StatusUnauthorized, "LOGIN_FAILED", err.Error(), nil)
	}

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    res.RefreshToken,
		HTTPOnly: true,
		Secure:   false,
		SameSite: "Lax",
	})

	return response.Success(c, fiber.StatusOK, "Login successful", res)
}

func (h *Handler) Refresh(c *fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")
	if refreshToken == "" {
		type refreshReq struct {
			RefreshToken string `json:"refresh_token"`
		}
		var req refreshReq
		_ = c.BodyParser(&req)
		refreshToken = req.RefreshToken
	}

	if refreshToken == "" {
		return response.Error(c, fiber.StatusBadRequest, "MISSING_REFRESH_TOKEN", "Refresh token is required", nil)
	}

	res, err := h.service.RefreshToken(c.Context(), refreshToken)
	if err != nil {
		return response.Error(c, fiber.StatusUnauthorized, "REFRESH_FAILED", err.Error(), nil)
	}

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    res.RefreshToken,
		HTTPOnly: true,
		Secure:   false,
		SameSite: "Lax",
	})

	return response.Success(c, fiber.StatusOK, "Token refreshed successfully", res)
}

func (h *Handler) Logout(c *fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")
	if refreshToken != "" {
		_ = h.service.Logout(c.Context(), refreshToken)
	}

	c.ClearCookie("refresh_token")
	return response.Success(c, fiber.StatusOK, "Logged out successfully", nil)
}

func (h *Handler) GetMe(c *fiber.Ctx) error {
	rawUserID := c.Locals("user_id")
	userID, ok := rawUserID.(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "Invalid session context", nil)
	}

	user, err := h.service.GetMe(c.Context(), userID)
	if err != nil {
		return response.Error(c, fiber.StatusNotFound, "USER_NOT_FOUND", "User profile not found", nil)
	}

	return response.Success(c, fiber.StatusOK, "User fetched successfully", user)
}
