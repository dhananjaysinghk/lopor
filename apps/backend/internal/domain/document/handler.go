package document

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

func (h *Handler) CreateDocument(c *fiber.Ctx) error {
	rawUserID := c.Locals("user_id")
	userID, ok := rawUserID.(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "Invalid session context", nil)
	}

	var req CreateDocumentRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Invalid payload", nil)
	}

	doc, err := h.service.CreateDocument(c.Context(), userID, req)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "CREATE_FAILED", err.Error(), nil)
	}

	return response.Success(c, fiber.StatusCreated, "Document created", doc)
}

func (h *Handler) GetWorkspaceDocuments(c *fiber.Ctx) error {
	wsIDParam := c.Params("wsId")
	wsID, err := uuid.Parse(wsIDParam)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_UUID", "Invalid workspace ID", nil)
	}

	docs, err := h.service.GetWorkspaceDocuments(c.Context(), wsID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "FETCH_FAILED", err.Error(), nil)
	}

	return response.Success(c, fiber.StatusOK, "Documents retrieved", docs)
}

func (h *Handler) GetDocumentByID(c *fiber.Ctx) error {
	docIDParam := c.Params("docId")
	docID, err := uuid.Parse(docIDParam)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_UUID", "Invalid document ID", nil)
	}

	doc, err := h.service.GetDocumentByID(c.Context(), docID)
	if err != nil {
		return response.Error(c, fiber.StatusNotFound, "NOT_FOUND", "Document not found", nil)
	}

	return response.Success(c, fiber.StatusOK, "Document details retrieved", doc)
}

func (h *Handler) UpdateDocument(c *fiber.Ctx) error {
	rawUserID := c.Locals("user_id")
	userID := rawUserID.(uuid.UUID)

	docIDParam := c.Params("docId")
	docID, err := uuid.Parse(docIDParam)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_UUID", "Invalid document ID", nil)
	}

	var req UpdateDocumentRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Invalid payload", nil)
	}

	doc, err := h.service.UpdateDocument(c.Context(), userID, docID, req)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "UPDATE_FAILED", err.Error(), nil)
	}

	return response.Success(c, fiber.StatusOK, "Document saved successfully", doc)
}

func (h *Handler) CreateFolder(c *fiber.Ctx) error {
	rawUserID := c.Locals("user_id")
	userID := rawUserID.(uuid.UUID)

	wsIDParam := c.Params("wsId")
	wsID, err := uuid.Parse(wsIDParam)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_UUID", "Invalid workspace ID", nil)
	}

	type FolderReq struct {
		Name string `json:"name"`
	}

	var req FolderReq
	if err := c.BodyParser(&req); err != nil || req.Name == "" {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_NAME", "Folder name is required", nil)
	}

	folder, err := h.service.CreateFolder(c.Context(), userID, wsID, req.Name)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "CREATE_FAILED", err.Error(), nil)
	}

	return response.Success(c, fiber.StatusCreated, "Folder created", folder)
}

func (h *Handler) GetWorkspaceFolders(c *fiber.Ctx) error {
	wsIDParam := c.Params("wsId")
	wsID, err := uuid.Parse(wsIDParam)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_UUID", "Invalid workspace ID", nil)
	}

	folders, err := h.service.GetWorkspaceFolders(c.Context(), wsID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "FETCH_FAILED", err.Error(), nil)
	}

	return response.Success(c, fiber.StatusOK, "Workspace folders retrieved", folders)
}
