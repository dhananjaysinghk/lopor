package rag

import (
	"context"

	"github.com/google/uuid"
	"github.com/lopor-ai/lopor/internal/domain/models"
	pkgRag "github.com/lopor-ai/lopor/pkg/rag"
)

type Service interface {
	IngestDocumentText(ctx context.Context, workspaceID uuid.UUID, docID *uuid.UUID, fileID *uuid.UUID, text string) (int, error)
	SemanticSearch(ctx context.Context, workspaceID uuid.UUID, query string, topK int) ([]SearchResult, error)
	GetFiles(ctx context.Context, workspaceID uuid.UUID) ([]*models.File, error)
	UploadFileRecord(ctx context.Context, file *models.File) error
}

type service struct {
	repo     Repository
	chunker  *pkgRag.Chunker
	embedder *pkgRag.MockEmbedder
}

func NewService(repo Repository) Service {
	return &service{
		repo:     repo,
		chunker:  pkgRag.NewChunker(512, 64),
		embedder: pkgRag.NewMockEmbedder(1536),
	}
}

func (s *service) IngestDocumentText(ctx context.Context, workspaceID uuid.UUID, docID *uuid.UUID, fileID *uuid.UUID, text string) (int, error) {
	chunks := s.chunker.ChunkText(text)
	count := 0

	for _, c := range chunks {
		vec, err := s.embedder.GenerateVector(ctx, c.Text)
		if err != nil {
			return count, err
		}

		vecStr := pkgRag.FormatPgVector(vec)
		emb := &models.Embedding{
			WorkspaceID: workspaceID,
			DocumentID:  docID,
			FileID:      fileID,
			ChunkIndex:  c.Index,
			ChunkText:   c.Text,
		}

		if err := s.repo.InsertEmbedding(ctx, emb, vecStr); err != nil {
			return count, err
		}
		count++
	}

	return count, nil
}

func (s *service) SemanticSearch(ctx context.Context, workspaceID uuid.UUID, query string, topK int) ([]SearchResult, error) {
	queryVec, err := s.embedder.GenerateVector(ctx, query)
	if err != nil {
		return nil, err
	}
	vecStr := pkgRag.FormatPgVector(queryVec)
	return s.repo.VectorSearch(ctx, workspaceID, vecStr, topK)
}

func (s *service) GetFiles(ctx context.Context, workspaceID uuid.UUID) ([]*models.File, error) {
	return s.repo.GetWorkspaceFiles(ctx, workspaceID)
}

func (s *service) UploadFileRecord(ctx context.Context, file *models.File) error {
	return s.repo.InsertFileRecord(ctx, file)
}
