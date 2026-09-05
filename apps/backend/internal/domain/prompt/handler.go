package prompt

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

func (h *Handler) CreatePrompt(c *fiber.Ctx) error {
	rawUserID := c.Locals("user_id")
	userID, ok := rawUserID.(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "Invalid session", nil)
	}

	var req CreatePromptRequest
	if err := c.BodyParser(&req); err != nil || req.Title == "" || req.Template == "" {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Title and Template content are required", nil)
	}

	p, err := h.service.CreatePrompt(c.Context(), userID, req)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "CREATE_FAILED", err.Error(), nil)
	}

	return response.Success(c, fiber.StatusCreated, "Prompt template created", p)
}

func (h *Handler) GetWorkspacePrompts(c *fiber.Ctx) error {
	wsIDParam := c.Params("wsId")
	wsID, err := uuid.Parse(wsIDParam)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_UUID", "Invalid workspace ID", nil)
	}

	prompts, err := h.service.GetWorkspacePrompts(c.Context(), wsID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "FETCH_FAILED", err.Error(), nil)
	}

	return response.Success(c, fiber.StatusOK, "Prompts retrieved", prompts)
}

func (h *Handler) SubstituteVariables(c *fiber.Ctx) error {
	promptIDParam := c.Params("promptId")
	promptID, err := uuid.Parse(promptIDParam)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_UUID", "Invalid prompt ID", nil)
	}

	var req ExecutePromptRequest
	_ = c.BodyParser(&req)

	resultText, err := h.service.SubstituteVariables(c.Context(), promptID, req)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "SUBSTITUTION_FAILED", err.Error(), nil)
	}

	return response.Success(c, fiber.StatusOK, "Variable substitution executed", fiber.Map{
		"prompt_id":     promptID,
		"hydrated_text": resultText,
	})
}

func (h *Handler) DeletePrompt(c *fiber.Ctx) error {
	promptIDParam := c.Params("promptId")
	promptID, err := uuid.Parse(promptIDParam)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_UUID", "Invalid prompt ID", nil)
	}

	if err := h.service.DeletePrompt(c.Context(), promptID); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "DELETE_FAILED", err.Error(), nil)
	}

	return response.Success(c, fiber.StatusOK, "Prompt template deleted", nil)
}
