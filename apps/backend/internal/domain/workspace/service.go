package workspace

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/lopor-ai/lopor/internal/domain/models"
)

type CreateWorkspaceRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Icon        *string `json:"icon,omitempty"`
}

type Service interface {
	CreateWorkspace(ctx context.Context, req CreateWorkspaceRequest, ownerID uuid.UUID) (*models.Workspace, error)
	GetUserWorkspaces(ctx context.Context, userID uuid.UUID) ([]*models.Workspace, error)
	GetWorkspaceByID(ctx context.Context, id uuid.UUID) (*models.Workspace, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) CreateWorkspace(ctx context.Context, req CreateWorkspaceRequest, ownerID uuid.UUID) (*models.Workspace, error) {
	slug := strings.ToLower(strings.ReplaceAll(req.Name, " ", "-"))
	ws := &models.Workspace{
		Name:        req.Name,
		Slug:        slug,
		Description: req.Description,
		Icon:        req.Icon,
	}

	if err := s.repo.CreateWorkspace(ctx, ws, ownerID); err != nil {
		return nil, err
	}
	return ws, nil
}

func (s *service) GetUserWorkspaces(ctx context.Context, userID uuid.UUID) ([]*models.Workspace, error) {
	return s.repo.GetUserWorkspaces(ctx, userID)
}

func (s *service) GetWorkspaceByID(ctx context.Context, id uuid.UUID) (*models.Workspace, error) {
	return s.repo.GetWorkspaceByID(ctx, id)
}
