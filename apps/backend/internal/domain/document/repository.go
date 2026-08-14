package document

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lopor-ai/lopor/internal/domain/models"
)

type Repository interface {
	CreateDocument(ctx context.Context, doc *models.Document) error
	GetWorkspaceDocuments(ctx context.Context, workspaceID uuid.UUID) ([]*models.Document, error)
	GetDocumentByID(ctx context.Context, id uuid.UUID) (*models.Document, error)
	UpdateDocument(ctx context.Context, doc *models.Document) error
	CreateFolder(ctx context.Context, folder *models.Folder) error
	GetWorkspaceFolders(ctx context.Context, workspaceID uuid.UUID) ([]*models.Folder, error)
}

type repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{pool: pool}
}

func (r *repository) CreateDocument(ctx context.Context, doc *models.Document) error {
	query := `
		INSERT INTO documents (id, workspace_id, folder_id, title, content, content_json, icon, cover_image, created_by, last_edited_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	doc.ID = uuid.New()
	now := time.Now()
	doc.CreatedAt = now
	doc.UpdatedAt = now
	if doc.Title == "" {
		doc.Title = "Untitled Document"
	}

	_, err := r.pool.Exec(ctx, query, doc.ID, doc.WorkspaceID, doc.FolderID, doc.Title, doc.Content, doc.ContentJSON, doc.Icon, doc.CoverImage, doc.CreatedBy, doc.LastEditedBy, doc.CreatedAt, doc.UpdatedAt)
	return err
}

func (r *repository) GetWorkspaceDocuments(ctx context.Context, workspaceID uuid.UUID) ([]*models.Document, error) {
	query := `
		SELECT id, workspace_id, folder_id, title, content, icon, cover_image, is_archived, is_published, created_by, last_edited_by, created_at, updated_at
		FROM documents WHERE workspace_id = $1 AND is_archived = false
		ORDER BY updated_at DESC
	`
	rows, err := r.pool.Query(ctx, query, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []*models.Document
	for rows.Next() {
		var d models.Document
		if err := rows.Scan(&d.ID, &d.WorkspaceID, &d.FolderID, &d.Title, &d.Content, &d.Icon, &d.CoverImage, &d.IsArchived, &d.IsPublished, &d.CreatedBy, &d.LastEditedBy, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		docs = append(docs, &d)
	}
	return docs, nil
}

func (r *repository) GetDocumentByID(ctx context.Context, id uuid.UUID) (*models.Document, error) {
	query := `
		SELECT id, workspace_id, folder_id, title, content, content_json, icon, cover_image, is_archived, is_published, created_by, last_edited_by, created_at, updated_at
		FROM documents WHERE id = $1
	`
	var d models.Document
	err := r.pool.QueryRow(ctx, query, id).Scan(&d.ID, &d.WorkspaceID, &d.FolderID, &d.Title, &d.Content, &d.ContentJSON, &d.Icon, &d.CoverImage, &d.IsArchived, &d.IsPublished, &d.CreatedBy, &d.LastEditedBy, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *repository) UpdateDocument(ctx context.Context, doc *models.Document) error {
	query := `
		UPDATE documents
		SET title = $1, content = $2, folder_id = $3, last_edited_by = $4, updated_at = NOW()
		WHERE id = $5
	`
	_, err := r.pool.Exec(ctx, query, doc.Title, doc.Content, doc.FolderID, doc.LastEditedBy, doc.ID)
	return err
}

func (r *repository) CreateFolder(ctx context.Context, folder *models.Folder) error {
	query := `
		INSERT INTO folders (id, workspace_id, parent_id, name, color, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	folder.ID = uuid.New()
	now := time.Now()
	folder.CreatedAt = now
	folder.UpdatedAt = now
	_, err := r.pool.Exec(ctx, query, folder.ID, folder.WorkspaceID, folder.ParentID, folder.Name, folder.Color, folder.CreatedBy, folder.CreatedAt, folder.UpdatedAt)
	return err
}

func (r *repository) GetWorkspaceFolders(ctx context.Context, workspaceID uuid.UUID) ([]*models.Folder, error) {
	query := `
		SELECT id, workspace_id, parent_id, name, color, created_by, created_at, updated_at
		FROM folders WHERE workspace_id = $1 ORDER BY name ASC
	`
	rows, err := r.pool.Query(ctx, query, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var folders []*models.Folder
	for rows.Next() {
		var f models.Folder
		if err := rows.Scan(&f.ID, &f.WorkspaceID, &f.ParentID, &f.Name, &f.Color, &f.CreatedBy, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		folders = append(folders, &f)
	}
	return folders, nil
}
