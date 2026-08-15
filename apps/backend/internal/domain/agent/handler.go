package agent

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

func (h *Handler) CreateAgent(c *fiber.Ctx) error {
	rawUserID := c.Locals("user_id")
	userID, ok := rawUserID.(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "Invalid session", nil)
	}

	var req CreateAgentRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Invalid payload", nil)
	}

	if req.Name == "" || req.SystemPrompt == "" {
		return response.Error(c, fiber.StatusBadRequest, "VALIDATION_FAILED", "Name and System Prompt are required", nil)
	}

	agent, err := h.service.CreateAgent(c.Context(), userID, req)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "CREATE_FAILED", err.Error(), nil)
	}

	return response.Success(c, fiber.StatusCreated, "Agent created", agent)
}

func (h *Handler) GetWorkspaceAgents(c *fiber.Ctx) error {
	wsIDParam := c.Params("wsId")
	wsID, err := uuid.Parse(wsIDParam)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_UUID", "Invalid workspace ID", nil)
	}

	agents, err := h.service.GetWorkspaceAgents(c.Context(), wsID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "FETCH_FAILED", err.Error(), nil)
	}

	return response.Success(c, fiber.StatusOK, "Agents retrieved", agents)
}

func (h *Handler) ExecuteAgent(c *fiber.Ctx) error {
	agentIDParam := c.Params("agentId")
	agentID, err := uuid.Parse(agentIDParam)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_UUID", "Invalid agent ID", nil)
	}

	var req ExecuteAgentRequest
	if err := c.BodyParser(&req); err != nil || req.Task == "" {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_TASK", "Task instruction text is required", nil)
	}

	res, err := h.service.ExecuteAgentTask(c.Context(), agentID, req)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "EXECUTION_FAILED", err.Error(), nil)
	}

	return response.Success(c, fiber.StatusOK, "Agent task executed", res)
}

func (h *Handler) DeleteAgent(c *fiber.Ctx) error {
	agentIDParam := c.Params("agentId")
	agentID, err := uuid.Parse(agentIDParam)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_UUID", "Invalid agent ID", nil)
	}

	if err := h.service.DeleteAgent(c.Context(), agentID); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "DELETE_FAILED", err.Error(), nil)
	}

	return response.Success(c, fiber.StatusOK, "Agent deleted", nil)
}
