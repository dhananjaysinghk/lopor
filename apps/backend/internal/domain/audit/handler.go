package audit

import (
	"fmt"

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

func (h *Handler) GetWorkspaceAuditLogs(c *fiber.Ctx) error {
	wsIDParam := c.Params("wsId")
	wsID, err := uuid.Parse(wsIDParam)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_UUID", "Invalid workspace ID", nil)
	}

	logs, err := h.service.GetWorkspaceLogs(c.Context(), wsID, 50)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "FETCH_FAILED", err.Error(), nil)
	}

	return response.Success(c, fiber.StatusOK, "Audit logs retrieved", logs)
}

func (h *Handler) ExportAuditLogsCSV(c *fiber.Ctx) error {
	wsIDParam := c.Params("wsId")
	wsID, err := uuid.Parse(wsIDParam)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_UUID", "Invalid workspace ID", nil)
	}

	csvData, err := h.service.ExportCSV(c.Context(), wsID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "EXPORT_FAILED", err.Error(), nil)
	}

	c.Set("Content-Type", "text/csv")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"audit-logs-%s.csv\"", wsID))
	return c.Send(csvData)
}
