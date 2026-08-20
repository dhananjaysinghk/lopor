package prompt

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PromptRecord struct {
	ID          uuid.UUID `json:"id"`
	WorkspaceID uuid.UUID `json:"workspace_id"`
	Title       string    `json:"title"`
	Description *string   `json:"description,omitempty"`
	Category    string    `json:"category"`
	Template    string    `json:"template"`
	Variables   string    `json:"variables"`
	CreatedBy   uuid.UUID `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Repository interface {
	CreatePrompt(ctx context.Context, prompt *PromptRecord) error
	GetWorkspacePrompts(ctx context.Context, workspaceID uuid.UUID) ([]*PromptRecord, error)
	GetPromptByID(ctx context.Context, id uuid.UUID) (*PromptRecord, error)
	DeletePrompt(ctx context.Context, id uuid.UUID) error
}

type repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{pool: pool}
}

func (r *repository) CreatePrompt(ctx context.Context, p *PromptRecord) error {
	query := `
		INSERT INTO prompts (id, workspace_id, title, description, category, template, variables, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb, $8, $9, $10)
	`
	p.ID = uuid.New()
	now := time.Now()
	p.CreatedAt = now
	p.UpdatedAt = now
	if p.Category == "" {
		p.Category = "General"
	}
	if p.Variables == "" {
		p.Variables = "[]"
	}

	_, err := r.pool.Exec(ctx, query, p.ID, p.WorkspaceID, p.Title, p.Description, p.Category, p.Template, p.Variables, p.CreatedBy, p.CreatedAt, p.UpdatedAt)
	return err
}

func (r *repository) GetWorkspacePrompts(ctx context.Context, workspaceID uuid.UUID) ([]*PromptRecord, error) {
	query := `
		SELECT id, workspace_id, title, description, category, template, variables::text, created_by, created_at, updated_at
		FROM prompts WHERE workspace_id = $1 ORDER BY updated_at DESC
	`
	rows, err := r.pool.Query(ctx, query, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prompts []*PromptRecord
	for rows.Next() {
		var p PromptRecord
		if err := rows.Scan(&p.ID, &p.WorkspaceID, &p.Title, &p.Description, &p.Category, &p.Template, &p.Variables, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		prompts = append(prompts, &p)
	}
	return prompts, nil
}

func (r *repository) GetPromptByID(ctx context.Context, id uuid.UUID) (*PromptRecord, error) {
	query := `
		SELECT id, workspace_id, title, description, category, template, variables::text, created_by, created_at, updated_at
		FROM prompts WHERE id = $1
	`
	var p PromptRecord
	err := r.pool.QueryRow(ctx, query, id).Scan(&p.ID, &p.WorkspaceID, &p.Title, &p.Description, &p.Category, &p.Template, &p.Variables, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *repository) DeletePrompt(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM prompts WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}
