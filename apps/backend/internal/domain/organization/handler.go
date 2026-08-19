package organization

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

func (h *Handler) CreateOrganization(c *fiber.Ctx) error {
	rawUserID := c.Locals("user_id")
	userID, ok := rawUserID.(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "Invalid session", nil)
	}

	var req CreateOrgRequest
	if err := c.BodyParser(&req); err != nil || req.Name == "" {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Organization name is required", nil)
	}

	org, err := h.service.CreateOrganization(c.Context(), userID, req)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "CREATE_FAILED", err.Error(), nil)
	}

	return response.Success(c, fiber.StatusCreated, "Organization created", org)
}

func (h *Handler) GetUserOrganizations(c *fiber.Ctx) error {
	rawUserID := c.Locals("user_id")
	userID := rawUserID.(uuid.UUID)

	orgs, err := h.service.GetUserOrganizations(c.Context(), userID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "FETCH_FAILED", err.Error(), nil)
	}

	return response.Success(c, fiber.StatusOK, "Organizations retrieved", orgs)
}

func (h *Handler) GetOrganizationMembers(c *fiber.Ctx) error {
	orgIDParam := c.Params("orgId")
	orgID, err := uuid.Parse(orgIDParam)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_UUID", "Invalid org ID", nil)
	}

	members, err := h.service.GetOrganizationMembers(c.Context(), orgID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "FETCH_FAILED", err.Error(), nil)
	}

	return response.Success(c, fiber.StatusOK, "Organization members retrieved", members)
}

func (h *Handler) InviteMember(c *fiber.Ctx) error {
	orgIDParam := c.Params("orgId")
	orgID, err := uuid.Parse(orgIDParam)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_UUID", "Invalid org ID", nil)
	}

	var req InviteMemberRequest
	if err := c.BodyParser(&req); err != nil || req.Email == "" {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_EMAIL", "Recipient email address is required", nil)
	}

	if err := h.service.InviteMember(c.Context(), orgID, req); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "INVITE_FAILED", err.Error(), nil)
	}

	return response.Success(c, fiber.StatusOK, "Invitation email dispatched successfully", nil)
}
