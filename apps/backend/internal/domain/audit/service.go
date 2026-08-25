package audit

import (
	"bytes"
	"context"
	"encoding/csv"

	"github.com/google/uuid"
)

type Service interface {
	RecordLog(ctx context.Context, log *AuditRecord) error
	GetWorkspaceLogs(ctx context.Context, workspaceID uuid.UUID, limit int) ([]*AuditRecord, error)
	ExportCSV(ctx context.Context, workspaceID uuid.UUID) ([]byte, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) RecordLog(ctx context.Context, log *AuditRecord) error {
	return s.repo.RecordAuditLog(ctx, log)
}

func (s *service) GetWorkspaceLogs(ctx context.Context, workspaceID uuid.UUID, limit int) ([]*AuditRecord, error) {
	return s.repo.GetWorkspaceAuditLogs(ctx, workspaceID, limit)
}

func (s *service) ExportCSV(ctx context.Context, workspaceID uuid.UUID) ([]byte, error) {
	logs, err := s.repo.GetWorkspaceAuditLogs(ctx, workspaceID, 1000)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write CSV Header
	_ = writer.Write([]string{"ID", "Action", "EntityType", "IPAddress", "Timestamp"})

	for _, l := range logs {
		_ = writer.Write([]string{
			l.ID.String(),
			l.Action,
			l.EntityType,
			l.IPAddress,
			l.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	writer.Flush()

	return buf.Bytes(), nil
}
