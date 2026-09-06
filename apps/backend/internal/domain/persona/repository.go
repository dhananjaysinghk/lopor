package persona

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PersonaRecord struct {
	ID           uuid.UUID `json:"id"`
	WorkspaceID  uuid.UUID `json:"workspace_id"`
	Name         string    `json:"name"`
	Description  *string   `json:"description,omitempty"`
	SystemPrompt string    `json:"system_prompt"`
	Tone         string    `json:"tone"`
	Temperature  float64   `json:"temperature"`
	MaxTokens    int       `json:"max_tokens"`
	Guardrails   string    `json:"guardrails"` // JSON string list
	IsDefault    bool      `json:"is_default"`
	CreatedBy    uuid.UUID `json:"created_by"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Repository interface {
	CreatePersona(ctx context.Context, p *PersonaRecord) error
	GetWorkspacePersonas(ctx context.Context, workspaceID uuid.UUID) ([]*PersonaRecord, error)
	GetPersonaByID(ctx context.Context, id uuid.UUID) (*PersonaRecord, error)
	UpdatePersona(ctx context.Context, p *PersonaRecord) error
	SetDefaultPersona(ctx context.Context, workspaceID, personaID uuid.UUID) error
	DeletePersona(ctx context.Context, id uuid.UUID) error
}

type repository struct {
	pool    *pgxpool.Pool
	mu      sync.RWMutex
	memMap  map[uuid.UUID]*PersonaRecord
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{
		pool:   pool,
		memMap: make(map[uuid.UUID]*PersonaRecord),
	}
}

func (r *repository) CreatePersona(ctx context.Context, p *PersonaRecord) error {
	p.ID = uuid.New()
	now := time.Now()
	p.CreatedAt = now
	p.UpdatedAt = now

	if p.Tone == "" {
		p.Tone = "professional"
	}
	if p.Temperature <= 0 {
		p.Temperature = 0.7
	}
	if p.MaxTokens <= 0 {
		p.MaxTokens = 4096
	}
	if p.Guardrails == "" {
		p.Guardrails = "[]"
	}

	if r.pool == nil {
		r.mu.Lock()
		defer r.mu.Unlock()
		if p.IsDefault {
			for _, item := range r.memMap {
				if item.WorkspaceID == p.WorkspaceID {
					item.IsDefault = false
				}
			}
		}
		r.memMap[p.ID] = p
		return nil
	}

	query := `
		INSERT INTO ai_personas (id, workspace_id, name, description, system_prompt, tone, temperature, max_tokens, guardrails, is_default, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9::jsonb, $10, $11, $12, $13)
	`
	_, err := r.pool.Exec(ctx, query, p.ID, p.WorkspaceID, p.Name, p.Description, p.SystemPrompt, p.Tone, p.Temperature, p.MaxTokens, p.Guardrails, p.IsDefault, p.CreatedBy, p.CreatedAt, p.UpdatedAt)
	return err
}

func (r *repository) GetWorkspacePersonas(ctx context.Context, workspaceID uuid.UUID) ([]*PersonaRecord, error) {
	if r.pool == nil {
		r.mu.RLock()
		defer r.mu.RUnlock()
		var res []*PersonaRecord
		for _, p := range r.memMap {
			if p.WorkspaceID == workspaceID {
				res = append(res, p)
			}
		}
		return res, nil
	}

	query := `
		SELECT id, workspace_id, name, description, system_prompt, tone, temperature, max_tokens, guardrails::text, is_default, created_by, created_at, updated_at
		FROM ai_personas WHERE workspace_id = $1 ORDER BY is_default DESC, name ASC
	`
	rows, err := r.pool.Query(ctx, query, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var personas []*PersonaRecord
	for rows.Next() {
		var p PersonaRecord
		if err := rows.Scan(&p.ID, &p.WorkspaceID, &p.Name, &p.Description, &p.SystemPrompt, &p.Tone, &p.Temperature, &p.MaxTokens, &p.Guardrails, &p.IsDefault, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		personas = append(personas, &p)
	}
	return personas, nil
}

func (r *repository) GetPersonaByID(ctx context.Context, id uuid.UUID) (*PersonaRecord, error) {
	if r.pool == nil {
		r.mu.RLock()
		defer r.mu.RUnlock()
		p, ok := r.memMap[id]
		if !ok {
			return nil, fmt.Errorf("persona not found")
		}
		return p, nil
	}

	query := `
		SELECT id, workspace_id, name, description, system_prompt, tone, temperature, max_tokens, guardrails::text, is_default, created_by, created_at, updated_at
		FROM ai_personas WHERE id = $1
	`
	var p PersonaRecord
	err := r.pool.QueryRow(ctx, query, id).Scan(&p.ID, &p.WorkspaceID, &p.Name, &p.Description, &p.SystemPrompt, &p.Tone, &p.Temperature, &p.MaxTokens, &p.Guardrails, &p.IsDefault, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *repository) UpdatePersona(ctx context.Context, p *PersonaRecord) error {
	p.UpdatedAt = time.Now()
	if r.pool == nil {
		r.mu.Lock()
		defer r.mu.Unlock()
		if _, ok := r.memMap[p.ID]; !ok {
			return fmt.Errorf("persona not found")
		}
		r.memMap[p.ID] = p
		return nil
	}

	query := `
		UPDATE ai_personas
		SET name = $1, description = $2, system_prompt = $3, tone = $4, temperature = $5, max_tokens = $6, guardrails = $7::jsonb, is_default = $8, updated_at = $9
		WHERE id = $10
	`
	_, err := r.pool.Exec(ctx, query, p.Name, p.Description, p.SystemPrompt, p.Tone, p.Temperature, p.MaxTokens, p.Guardrails, p.IsDefault, p.UpdatedAt, p.ID)
	return err
}

func (r *repository) SetDefaultPersona(ctx context.Context, workspaceID, personaID uuid.UUID) error {
	if r.pool == nil {
		r.mu.Lock()
		defer r.mu.Unlock()
		for _, p := range r.memMap {
			if p.WorkspaceID == workspaceID {
				p.IsDefault = (p.ID == personaID)
			}
		}
		return nil
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `UPDATE ai_personas SET is_default = false WHERE workspace_id = $1`, workspaceID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `UPDATE ai_personas SET is_default = true WHERE id = $1 AND workspace_id = $2`, personaID, workspaceID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *repository) DeletePersona(ctx context.Context, id uuid.UUID) error {
	if r.pool == nil {
		r.mu.Lock()
		defer r.mu.Unlock()
		delete(r.memMap, id)
		return nil
	}

	query := `DELETE FROM ai_personas WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}
