package graph

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

func (h *Handler) GetWorkspaceGraph(c *fiber.Ctx) error {
	wsIDParam := c.Params("wsId")
	wsID, err := uuid.Parse(wsIDParam)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_UUID", "Invalid workspace ID", nil)
	}

	graph, err := h.service.GetWorkspaceGraph(c.Context(), wsID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "FETCH_FAILED", err.Error(), nil)
	}

	return response.Success(c, fiber.StatusOK, "Workspace knowledge graph retrieved", graph)
}
