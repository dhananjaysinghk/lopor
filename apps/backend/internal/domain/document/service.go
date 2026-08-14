package document

import (
	"context"

	"github.com/google/uuid"
	"github.com/lopor-ai/lopor/internal/domain/models"
)

type CreateDocumentRequest struct {
	WorkspaceID uuid.UUID  `json:"workspace_id"`
	FolderID    *uuid.UUID `json:"folder_id,omitempty"`
	Title       string     `json:"title"`
	Content     *string    `json:"content,omitempty"`
}

type UpdateDocumentRequest struct {
	Title    string     `json:"title"`
	Content  *string    `json:"content,omitempty"`
	FolderID *uuid.UUID `json:"folder_id,omitempty"`
}

type Service interface {
	CreateDocument(ctx context.Context, userID uuid.UUID, req CreateDocumentRequest) (*models.Document, error)
	GetWorkspaceDocuments(ctx context.Context, workspaceID uuid.UUID) ([]*models.Document, error)
	GetDocumentByID(ctx context.Context, id uuid.UUID) (*models.Document, error)
	UpdateDocument(ctx context.Context, userID, docID uuid.UUID, req UpdateDocumentRequest) (*models.Document, error)
	CreateFolder(ctx context.Context, userID, workspaceID uuid.UUID, name string) (*models.Folder, error)
	GetWorkspaceFolders(ctx context.Context, workspaceID uuid.UUID) ([]*models.Folder, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) CreateDocument(ctx context.Context, userID uuid.UUID, req CreateDocumentRequest) (*models.Document, error) {
	doc := &models.Document{
		WorkspaceID:  req.WorkspaceID,
		FolderID:     req.FolderID,
		Title:        req.Title,
		Content:      req.Content,
		CreatedBy:    userID,
		LastEditedBy: userID,
	}

	if err := s.repo.CreateDocument(ctx, doc); err != nil {
		return nil, err
	}
	return doc, nil
}

func (s *service) GetWorkspaceDocuments(ctx context.Context, workspaceID uuid.UUID) ([]*models.Document, error) {
	return s.repo.GetWorkspaceDocuments(ctx, workspaceID)
}

func (s *service) GetDocumentByID(ctx context.Context, id uuid.UUID) (*models.Document, error) {
	return s.repo.GetDocumentByID(ctx, id)
}

func (s *service) UpdateDocument(ctx context.Context, userID, docID uuid.UUID, req UpdateDocumentRequest) (*models.Document, error) {
	doc, err := s.repo.GetDocumentByID(ctx, docID)
	if err != nil {
		return nil, err
	}

	if req.Title != "" {
		doc.Title = req.Title
	}
	if req.Content != nil {
		doc.Content = req.Content
	}
	doc.FolderID = req.FolderID
	doc.LastEditedBy = userID

	if err := s.repo.UpdateDocument(ctx, doc); err != nil {
		return nil, err
	}
	return doc, nil
}

func (s *service) CreateFolder(ctx context.Context, userID, workspaceID uuid.UUID, name string) (*models.Folder, error) {
	folder := &models.Folder{
		WorkspaceID: workspaceID,
		Name:        name,
		CreatedBy:   userID,
	}
	if err := s.repo.CreateFolder(ctx, folder); err != nil {
		return nil, err
	}
	return folder, nil
}

func (s *service) GetWorkspaceFolders(ctx context.Context, workspaceID uuid.UUID) ([]*models.Folder, error) {
	return s.repo.GetWorkspaceFolders(ctx, workspaceID)
}
