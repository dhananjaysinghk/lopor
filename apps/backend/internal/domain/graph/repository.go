package graph

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Node struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Type  string `json:"type"` // "document", "agent", "vector_chunk", "user"
	Color string `json:"color"`
}

type Edge struct {
	Source string  `json:"source"`
	Target string  `json:"target"`
	Label  string  `json:"label"`
	Weight float64 `json:"weight"`
}

type GraphData struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

type Repository interface {
	GetWorkspaceGraph(ctx context.Context, workspaceID uuid.UUID) (*GraphData, error)
}

type repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{pool: pool}
}

func (r *repository) GetWorkspaceGraph(ctx context.Context, workspaceID uuid.UUID) (*GraphData, error) {
	nodes := []Node{
		{ID: "doc_1", Label: "System Architecture PRD", Type: "document", Color: "#6366f1"},
		{ID: "doc_2", Label: "PostgreSQL pgvector Migration Guide", Type: "document", Color: "#10b981"},
		{ID: "agent_1", Label: "Code Security Reviewer Agent", Type: "agent", Color: "#a855f7"},
		{ID: "vec_1", Label: "pgvector HNSW Cosine Vector", Type: "vector_chunk", Color: "#f59e0b"},
		{ID: "usr_1", Label: "Sarah Connor (Workspace Owner)", Type: "user", Color: "#ec4899"},
	}

	edges := []Edge{
		{Source: "usr_1", Target: "doc_1", Label: "created", Weight: 1.0},
		{Source: "usr_1", Target: "doc_2", Label: "edited", Weight: 0.9},
		{Source: "agent_1", Target: "doc_1", Label: "audited", Weight: 0.85},
		{Source: "doc_2", Target: "vec_1", Label: "vectorized", Weight: 0.98},
		{Source: "doc_1", Target: "vec_1", Label: "semantic_match", Weight: 0.92},
	}

	return &GraphData{
		Nodes: nodes,
		Edges: edges,
	}, nil
}
