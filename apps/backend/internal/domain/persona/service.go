package persona

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type Service interface {
	CreatePersona(ctx context.Context, p *PersonaRecord) (*PersonaRecord, error)
	GetWorkspacePersonas(ctx context.Context, workspaceID uuid.UUID) ([]*PersonaRecord, error)
	GetPersonaByID(ctx context.Context, id uuid.UUID) (*PersonaRecord, error)
	UpdatePersona(ctx context.Context, p *PersonaRecord) (*PersonaRecord, error)
	SetDefaultPersona(ctx context.Context, workspaceID, personaID uuid.UUID) error
	DeletePersona(ctx context.Context, id uuid.UUID) error
	FormatSystemPrompt(p *PersonaRecord) string
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) CreatePersona(ctx context.Context, p *PersonaRecord) (*PersonaRecord, error) {
	if strings.TrimSpace(p.Name) == "" {
		return nil, fmt.Errorf("persona name is required")
	}
	if strings.TrimSpace(p.SystemPrompt) == "" {
		return nil, fmt.Errorf("system prompt is required")
	}

	err := s.repo.CreatePersona(ctx, p)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (s *service) GetWorkspacePersonas(ctx context.Context, workspaceID uuid.UUID) ([]*PersonaRecord, error) {
	return s.repo.GetWorkspacePersonas(ctx, workspaceID)
}

func (s *service) GetPersonaByID(ctx context.Context, id uuid.UUID) (*PersonaRecord, error) {
	return s.repo.GetPersonaByID(ctx, id)
}

func (s *service) UpdatePersona(ctx context.Context, p *PersonaRecord) (*PersonaRecord, error) {
	if strings.TrimSpace(p.Name) == "" {
		return nil, fmt.Errorf("persona name is required")
	}
	if strings.TrimSpace(p.SystemPrompt) == "" {
		return nil, fmt.Errorf("system prompt is required")
	}

	err := s.repo.UpdatePersona(ctx, p)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (s *service) SetDefaultPersona(ctx context.Context, workspaceID, personaID uuid.UUID) error {
	return s.repo.SetDefaultPersona(ctx, workspaceID, personaID)
}

func (s *service) DeletePersona(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeletePersona(ctx, id)
}

func (s *service) FormatSystemPrompt(p *PersonaRecord) string {
	if p == nil {
		return "You are Lopor AI, a helpful enterprise workspace assistant."
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("SYSTEM PERSONA: %s\n", p.Name))
	if p.Tone != "" {
		sb.WriteString(fmt.Sprintf("TONE & STYLE: %s\n", p.Tone))
	}
	sb.WriteString("\nINSTRUCTIONS:\n")
	sb.WriteString(p.SystemPrompt)
	sb.WriteString("\n")

	if p.Guardrails != "" && p.Guardrails != "[]" {
		sb.WriteString(fmt.Sprintf("\nSTRICT GUARDRAILS:\n%s\n", p.Guardrails))
	}

	return sb.String()
}
