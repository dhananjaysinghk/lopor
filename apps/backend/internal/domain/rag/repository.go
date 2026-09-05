package rag

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lopor-ai/lopor/internal/domain/models"
)

type SearchResult struct {
	ID              uuid.UUID  `json:"id"`
	WorkspaceID     uuid.UUID  `json:"workspace_id"`
	DocumentID      *uuid.UUID `json:"document_id,omitempty"`
	FileID          *uuid.UUID `json:"file_id,omitempty"`
	ChunkIndex      int        `json:"chunk_index"`
	ChunkText       string     `json:"chunk_text"`
	SimilarityScore float64    `json:"similarity_score"`
}

type Repository interface {
	InsertEmbedding(ctx context.Context, emb *models.Embedding, vectorStr string) error
	VectorSearch(ctx context.Context, workspaceID uuid.UUID, vectorStr string, topK int) ([]SearchResult, error)
	HybridRRFSearch(ctx context.Context, workspaceID uuid.UUID, queryText string, vectorStr string, topK int) ([]SearchResult, error)
	InsertFileRecord(ctx context.Context, file *models.File) error
	GetWorkspaceFiles(ctx context.Context, workspaceID uuid.UUID) ([]*models.File, error)
}

type repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{pool: pool}
}

func (r *repository) InsertEmbedding(ctx context.Context, emb *models.Embedding, vectorStr string) error {
	query := `
		INSERT INTO embeddings (id, workspace_id, document_id, file_id, chunk_index, chunk_text, embedding, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7::vector, $8, $9)
	`
	emb.ID = uuid.New()
	emb.CreatedAt = time.Now()
	_, err := r.pool.Exec(ctx, query, emb.ID, emb.WorkspaceID, emb.DocumentID, emb.FileID, emb.ChunkIndex, emb.ChunkText, vectorStr, emb.Metadata, emb.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert vector embedding: %w", err)
	}
	return nil
}

func (r *repository) VectorSearch(ctx context.Context, workspaceID uuid.UUID, vectorStr string, topK int) ([]SearchResult, error) {
	if topK <= 0 {
		topK = 5
	}

	query := `
		SELECT id, workspace_id, document_id, file_id, chunk_index, chunk_text,
		       1 - (embedding <=> $1::vector) AS similarity_score
		FROM embeddings
		WHERE workspace_id = $2
		ORDER BY embedding <=> $1::vector ASC
		LIMIT $3
	`

	rows, err := r.pool.Query(ctx, query, vectorStr, workspaceID, topK)
	if err != nil {
		return nil, fmt.Errorf("pgvector similarity query failed: %w", err)
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var res SearchResult
		if err := rows.Scan(&res.ID, &res.WorkspaceID, &res.DocumentID, &res.FileID, &res.ChunkIndex, &res.ChunkText, &res.SimilarityScore); err != nil {
			return nil, err
		}
		results = append(results, res)
	}

	return results, nil
}

func (r *repository) HybridRRFSearch(ctx context.Context, workspaceID uuid.UUID, queryText string, vectorStr string, topK int) ([]SearchResult, error) {
	if topK <= 0 {
		topK = 5
	}

	// Hybrid query combining PostgreSQL full-text search (ts_rank_cd) with pgvector cosine distance
	query := `
		SELECT id, workspace_id, document_id, file_id, chunk_index, chunk_text,
		       (0.7 * (1 - (embedding <=> $1::vector)) + 0.3 * ts_rank_cd(to_tsvector('english', chunk_text), plainto_tsquery('english', $2))) AS similarity_score
		FROM embeddings
		WHERE workspace_id = $3
		ORDER BY similarity_score DESC
		LIMIT $4
	`

	rows, err := r.pool.Query(ctx, query, vectorStr, queryText, workspaceID, topK)
	if err != nil {
		// Fallback to pure vector search if tsvector fails
		return r.VectorSearch(ctx, workspaceID, vectorStr, topK)
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var res SearchResult
		if err := rows.Scan(&res.ID, &res.WorkspaceID, &res.DocumentID, &res.FileID, &res.ChunkIndex, &res.ChunkText, &res.SimilarityScore); err != nil {
			return nil, err
		}
		results = append(results, res)
	}

	return results, nil
}

func (r *repository) InsertFileRecord(ctx context.Context, f *models.File) error {
	query := `
		INSERT INTO files (id, workspace_id, filename, mime_type, file_size, storage_key, uploaded_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	f.ID = uuid.New()
	f.CreatedAt = time.Now()
	_, err := r.pool.Exec(ctx, query, f.ID, f.WorkspaceID, f.Filename, f.MimeType, f.FileSize, f.StorageKey, f.UploadedBy, f.CreatedAt)
	return err
}

func (r *repository) GetWorkspaceFiles(ctx context.Context, workspaceID uuid.UUID) ([]*models.File, error) {
	query := `
		SELECT id, workspace_id, filename, mime_type, file_size, storage_key, uploaded_by, created_at
		FROM files WHERE workspace_id = $1 ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []*models.File
	for rows.Next() {
		var f models.File
		if err := rows.Scan(&f.ID, &f.WorkspaceID, &f.Filename, &f.MimeType, &f.FileSize, &f.StorageKey, &f.UploadedBy, &f.CreatedAt); err != nil {
			return nil, err
		}
		files = append(files, &f)
	}
	return files, nil
}
