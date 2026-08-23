package graph

import (
	"context"

	"github.com/google/uuid"
)

type Service interface {
	GetWorkspaceGraph(ctx context.Context, workspaceID uuid.UUID) (*GraphData, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetWorkspaceGraph(ctx context.Context, workspaceID uuid.UUID) (*GraphData, error) {
	return s.repo.GetWorkspaceGraph(ctx, workspaceID)
}
