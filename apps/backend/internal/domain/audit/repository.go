package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuditRecord struct {
	ID          uuid.UUID       `json:"id"`
	WorkspaceID *uuid.UUID      `json:"workspace_id,omitempty"`
	UserID      *uuid.UUID      `json:"user_id,omitempty"`
	Action      string          `json:"action"`
	EntityType  string          `json:"entity_type"`
	EntityID    *uuid.UUID      `json:"entity_id,omitempty"`
	IPAddress   string          `json:"ip_address"`
	Details     json.RawMessage `json:"details"`
	CreatedAt   time.Time       `json:"created_at"`
}

type Repository interface {
	RecordAuditLog(ctx context.Context, log *AuditRecord) error
	GetWorkspaceAuditLogs(ctx context.Context, workspaceID uuid.UUID, limit int) ([]*AuditRecord, error)
}

type repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{pool: pool}
}

func (r *repository) RecordAuditLog(ctx context.Context, log *AuditRecord) error {
	query := `
		INSERT INTO audit_logs (id, workspace_id, user_id, action, entity_type, entity_id, ip_address, details, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	log.ID = uuid.New()
	log.CreatedAt = time.Now()
	if log.Details == nil {
		log.Details = json.RawMessage("{}")
	}

	_, err := r.pool.Exec(ctx, query, log.ID, log.WorkspaceID, log.UserID, log.Action, log.EntityType, log.EntityID, log.IPAddress, log.Details, log.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert audit log: %w", err)
	}
	return nil
}

func (r *repository) GetWorkspaceAuditLogs(ctx context.Context, workspaceID uuid.UUID, limit int) ([]*AuditRecord, error) {
	if limit <= 0 {
		limit = 50
	}

	query := `
		SELECT id, workspace_id, user_id, action, entity_type, entity_id, ip_address, details, created_at
		FROM audit_logs WHERE workspace_id = $1 ORDER BY created_at DESC LIMIT $2
	`
	rows, err := r.pool.Query(ctx, query, workspaceID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*AuditRecord
	for rows.Next() {
		var l AuditRecord
		if err := rows.Scan(&l.ID, &l.WorkspaceID, &l.UserID, &l.Action, &l.EntityType, &l.EntityID, &l.IPAddress, &l.Details, &l.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, &l)
	}
	return logs, nil
}
