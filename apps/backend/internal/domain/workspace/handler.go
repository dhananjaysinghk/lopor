package workspace

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

func (h *Handler) CreateWorkspace(c *fiber.Ctx) error {
	rawUserID := c.Locals("user_id")
	userID, ok := rawUserID.(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "Invalid session context", nil)
	}

	var req CreateWorkspaceRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Invalid JSON payload", nil)
	}

	if req.Name == "" {
		return response.Error(c, fiber.StatusBadRequest, "VALIDATION_FAILED", "Workspace name is required", nil)
	}

	ws, err := h.service.CreateWorkspace(c.Context(), req, userID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "CREATE_FAILED", err.Error(), nil)
	}

	return response.Success(c, fiber.StatusCreated, "Workspace created successfully", ws)
}

func (h *Handler) GetUserWorkspaces(c *fiber.Ctx) error {
	rawUserID := c.Locals("user_id")
	userID, ok := rawUserID.(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "Invalid session context", nil)
	}

	workspaces, err := h.service.GetUserWorkspaces(c.Context(), userID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "FETCH_FAILED", err.Error(), nil)
	}

	return response.Success(c, fiber.StatusOK, "Workspaces retrieved successfully", workspaces)
}

func (h *Handler) GetWorkspaceByID(c *fiber.Ctx) error {
	idParam := c.Params("id")
	wsID, err := uuid.Parse(idParam)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_UUID", "Invalid workspace ID", nil)
	}

	ws, err := h.service.GetWorkspaceByID(c.Context(), wsID)
	if err != nil {
		return response.Error(c, fiber.StatusNotFound, "NOT_FOUND", "Workspace not found", nil)
	}

	return response.Success(c, fiber.StatusOK, "Workspace retrieved successfully", ws)
}
