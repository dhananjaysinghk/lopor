package workspace

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lopor-ai/lopor/internal/domain/models"
)

type Repository interface {
	CreateWorkspace(ctx context.Context, ws *models.Workspace, ownerID uuid.UUID) error
	GetUserWorkspaces(ctx context.Context, userID uuid.UUID) ([]*models.Workspace, error)
	GetWorkspaceByID(ctx context.Context, id uuid.UUID) (*models.Workspace, error)
}

type repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{pool: pool}
}

func (r *repository) CreateWorkspace(ctx context.Context, ws *models.Workspace, ownerID uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Create default organization if none specified
	if ws.OrganizationID == uuid.Nil {
		var orgID uuid.UUID
		orgQuery := `INSERT INTO organizations (name, slug, owner_id) VALUES ($1, $2, $3) RETURNING id`
		err := tx.QueryRow(ctx, orgQuery, ws.Name+" Org", ws.Slug+"-org", ownerID).Scan(&orgID)
		if err != nil {
			return fmt.Errorf("failed to create default organization: %w", err)
		}
		ws.OrganizationID = orgID
	}

	ws.ID = uuid.New()
	ws.CreatedAt = time.Now()
	ws.UpdatedAt = time.Now()

	wsQuery := `
		INSERT INTO workspaces (id, organization_id, name, slug, description, icon, is_public, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err = tx.Exec(ctx, wsQuery, ws.ID, ws.OrganizationID, ws.Name, ws.Slug, ws.Description, ws.Icon, ws.IsPublic, ws.CreatedAt, ws.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert workspace: %w", err)
	}

	memberQuery := `INSERT INTO workspace_members (workspace_id, user_id, role) VALUES ($1, $2, 'owner')`
	_, err = tx.Exec(ctx, memberQuery, ws.ID, ownerID)
	if err != nil {
		return fmt.Errorf("failed to assign workspace owner: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *repository) GetUserWorkspaces(ctx context.Context, userID uuid.UUID) ([]*models.Workspace, error) {
	query := `
		SELECT w.id, w.organization_id, w.name, w.slug, w.description, w.icon, w.is_public, w.created_at, w.updated_at
		FROM workspaces w
		JOIN workspace_members wm ON w.id = wm.workspace_id
		WHERE wm.user_id = $1
		ORDER BY w.updated_at DESC
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workspaces []*models.Workspace
	for rows.Next() {
		var w models.Workspace
		if err := rows.Scan(&w.ID, &w.OrganizationID, &w.Name, &w.Slug, &w.Description, &w.Icon, &w.IsPublic, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, err
		}
		workspaces = append(workspaces, &w)
	}
	return workspaces, nil
}

func (r *repository) GetWorkspaceByID(ctx context.Context, id uuid.UUID) (*models.Workspace, error) {
	query := `
		SELECT id, organization_id, name, slug, description, icon, is_public, created_at, updated_at
		FROM workspaces WHERE id = $1
	`
	var w models.Workspace
	err := r.pool.QueryRow(ctx, query, id).Scan(&w.ID, &w.OrganizationID, &w.Name, &w.Slug, &w.Description, &w.Icon, &w.IsPublic, &w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &w, nil
}
