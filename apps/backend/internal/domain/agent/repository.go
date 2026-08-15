package agent

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AgentRecord struct {
	ID           uuid.UUID `json:"id"`
	WorkspaceID  uuid.UUID `json:"workspace_id"`
	Name         string    `json:"name"`
	Description  *string   `json:"description,omitempty"`
	SystemPrompt string    `json:"system_prompt"`
	Tools        string    `json:"tools"`
	CreatedBy    uuid.UUID `json:"created_by"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Repository interface {
	CreateAgent(ctx context.Context, agent *AgentRecord) error
	GetWorkspaceAgents(ctx context.Context, workspaceID uuid.UUID) ([]*AgentRecord, error)
	GetAgentByID(ctx context.Context, id uuid.UUID) (*AgentRecord, error)
	DeleteAgent(ctx context.Context, id uuid.UUID) error
}

type repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{pool: pool}
}

func (r *repository) CreateAgent(ctx context.Context, agent *AgentRecord) error {
	query := `
		INSERT INTO agents (id, workspace_id, name, description, system_prompt, tools, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7, $8, $9)
	`
	agent.ID = uuid.New()
	now := time.Now()
	agent.CreatedAt = now
	agent.UpdatedAt = now

	_, err := r.pool.Exec(ctx, query, agent.ID, agent.WorkspaceID, agent.Name, agent.Description, agent.SystemPrompt, agent.Tools, agent.CreatedBy, agent.CreatedAt, agent.UpdatedAt)
	return err
}

func (r *repository) GetWorkspaceAgents(ctx context.Context, workspaceID uuid.UUID) ([]*AgentRecord, error) {
	query := `
		SELECT id, workspace_id, name, description, system_prompt, tools::text, created_by, created_at, updated_at
		FROM agents WHERE workspace_id = $1 ORDER BY updated_at DESC
	`
	rows, err := r.pool.Query(ctx, query, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var agents []*AgentRecord
	for rows.Next() {
		var a AgentRecord
		if err := rows.Scan(&a.ID, &a.WorkspaceID, &a.Name, &a.Description, &a.SystemPrompt, &a.Tools, &a.CreatedBy, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		agents = append(agents, &a)
	}
	return agents, nil
}

func (r *repository) GetAgentByID(ctx context.Context, id uuid.UUID) (*AgentRecord, error) {
	query := `
		SELECT id, workspace_id, name, description, system_prompt, tools::text, created_by, created_at, updated_at
		FROM agents WHERE id = $1
	`
	var a AgentRecord
	err := r.pool.QueryRow(ctx, query, id).Scan(&a.ID, &a.WorkspaceID, &a.Name, &a.Description, &a.SystemPrompt, &a.Tools, &a.CreatedBy, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *repository) DeleteAgent(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM agents WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}
