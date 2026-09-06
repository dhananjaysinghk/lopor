package persona

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

type CreatePersonaReq struct {
	Name         string   `json:"name"`
	Description  *string  `json:"description,omitempty"`
	SystemPrompt string   `json:"system_prompt"`
	Tone         string   `json:"tone"`
	Temperature  float64  `json:"temperature"`
	MaxTokens    int      `json:"max_tokens"`
	Guardrails   string   `json:"guardrails"`
	IsDefault    bool     `json:"is_default"`
}

func (h *Handler) CreatePersona(c *fiber.Ctx) error {
	wsIDStr := c.Params("wsId")
	wsID, err := uuid.Parse(wsIDStr)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_WORKSPACE_ID", "Workspace ID is invalid", nil)
	}

	userIDVal := c.Locals("userId")
	userID, _ := uuid.Parse(fmtString(userIDVal))

	var req CreatePersonaReq
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Invalid JSON payload", nil)
	}

	record := &PersonaRecord{
		WorkspaceID:  wsID,
		Name:         req.Name,
		Description:  req.Description,
		SystemPrompt: req.SystemPrompt,
		Tone:         req.Tone,
		Temperature:  req.Temperature,
		MaxTokens:    req.MaxTokens,
		Guardrails:   req.Guardrails,
		IsDefault:    req.IsDefault,
		CreatedBy:    userID,
	}

	created, err := h.service.CreatePersona(c.Context(), record)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "CREATE_FAILED", err.Error(), nil)
	}

	return response.Success(c, fiber.StatusCreated, "AI Persona created successfully", created)
}

func (h *Handler) GetWorkspacePersonas(c *fiber.Ctx) error {
	wsIDStr := c.Params("wsId")
	wsID, err := uuid.Parse(wsIDStr)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_WORKSPACE_ID", "Workspace ID is invalid", nil)
	}

	personas, err := h.service.GetWorkspacePersonas(c.Context(), wsID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "FETCH_FAILED", err.Error(), nil)
	}

	return response.Success(c, fiber.StatusOK, "AI Personas retrieved successfully", personas)
}

func (h *Handler) GetPersonaByID(c *fiber.Ctx) error {
	personaIDStr := c.Params("personaId")
	personaID, err := uuid.Parse(personaIDStr)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_PERSONA_ID", "Persona ID is invalid", nil)
	}

	persona, err := h.service.GetPersonaByID(c.Context(), personaID)
	if err != nil {
		return response.Error(c, fiber.StatusNotFound, "NOT_FOUND", "Persona not found", nil)
	}

	return response.Success(c, fiber.StatusOK, "AI Persona retrieved", persona)
}

func (h *Handler) UpdatePersona(c *fiber.Ctx) error {
	personaIDStr := c.Params("personaId")
	personaID, err := uuid.Parse(personaIDStr)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_PERSONA_ID", "Persona ID is invalid", nil)
	}

	var req CreatePersonaReq
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Invalid JSON payload", nil)
	}

	existing, err := h.service.GetPersonaByID(c.Context(), personaID)
	if err != nil {
		return response.Error(c, fiber.StatusNotFound, "NOT_FOUND", "Persona not found", nil)
	}

	existing.Name = req.Name
	existing.Description = req.Description
	existing.SystemPrompt = req.SystemPrompt
	existing.Tone = req.Tone
	existing.Temperature = req.Temperature
	existing.MaxTokens = req.MaxTokens
	existing.Guardrails = req.Guardrails
	existing.IsDefault = req.IsDefault

	updated, err := h.service.UpdatePersona(c.Context(), existing)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "UPDATE_FAILED", err.Error(), nil)
	}

	return response.Success(c, fiber.StatusOK, "AI Persona updated successfully", updated)
}

func (h *Handler) SetDefaultPersona(c *fiber.Ctx) error {
	wsIDStr := c.Params("wsId")
	wsID, err := uuid.Parse(wsIDStr)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_WORKSPACE_ID", "Workspace ID is invalid", nil)
	}

	personaIDStr := c.Params("personaId")
	personaID, err := uuid.Parse(personaIDStr)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_PERSONA_ID", "Persona ID is invalid", nil)
	}

	err = h.service.SetDefaultPersona(c.Context(), wsID, personaID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "SET_DEFAULT_FAILED", err.Error(), nil)
	}

	return response.Success(c, fiber.StatusOK, "Default AI Persona set successfully", nil)
}

func (h *Handler) DeletePersona(c *fiber.Ctx) error {
	personaIDStr := c.Params("personaId")
	personaID, err := uuid.Parse(personaIDStr)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_PERSONA_ID", "Persona ID is invalid", nil)
	}

	err = h.service.DeletePersona(c.Context(), personaID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "DELETE_FAILED", err.Error(), nil)
	}

	return response.Success(c, fiber.StatusOK, "AI Persona deleted successfully", nil)
}

func fmtString(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
