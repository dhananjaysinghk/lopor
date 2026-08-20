package prompt

import (
	"context"
	"strings"

	"github.com/google/uuid"
)

type CreatePromptRequest struct {
	WorkspaceID uuid.UUID `json:"workspace_id"`
	Title       string    `json:"title"`
	Description *string   `json:"description,omitempty"`
	Category    string    `json:"category"`
	Template    string    `json:"template"`
	Variables   string    `json:"variables"`
}

type ExecutePromptRequest struct {
	VariableValues map[string]string `json:"variable_values"`
}

type Service interface {
	CreatePrompt(ctx context.Context, userID uuid.UUID, req CreatePromptRequest) (*PromptRecord, error)
	GetWorkspacePrompts(ctx context.Context, workspaceID uuid.UUID) ([]*PromptRecord, error)
	GetPromptByID(ctx context.Context, id uuid.UUID) (*PromptRecord, error)
	SubstituteVariables(ctx context.Context, promptID uuid.UUID, req ExecutePromptRequest) (string, error)
	DeletePrompt(ctx context.Context, id uuid.UUID) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) CreatePrompt(ctx context.Context, userID uuid.UUID, req CreatePromptRequest) (*PromptRecord, error) {
	p := &PromptRecord{
		WorkspaceID: req.WorkspaceID,
		Title:       req.Title,
		Description: req.Description,
		Category:    req.Category,
		Template:    req.Template,
		Variables:   req.Variables,
		CreatedBy:   userID,
	}

	if err := s.repo.CreatePrompt(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *service) GetWorkspacePrompts(ctx context.Context, workspaceID uuid.UUID) ([]*PromptRecord, error) {
	return s.repo.GetWorkspacePrompts(ctx, workspaceID)
}

func (s *service) GetPromptByID(ctx context.Context, id uuid.UUID) (*PromptRecord, error) {
	return s.repo.GetPromptByID(ctx, id)
}

func (s *service) SubstituteVariables(ctx context.Context, promptID uuid.UUID, req ExecutePromptRequest) (string, error) {
	p, err := s.repo.GetPromptByID(ctx, promptID)
	if err != nil {
		return "", err
	}

	result := p.Template
	for k, v := range req.VariableValues {
		placeholder := "{{" + k + "}}"
		result = strings.ReplaceAll(result, placeholder, v)
	}

	return result, nil
}

func (s *service) DeletePrompt(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeletePrompt(ctx, id)
}
