package chunking

import (
	"context"
	"strings"
	"testing"
)

func TestMarkdownHierarchicalChunking(t *testing.T) {
	engine := NewChunkingEngine()
	ctx := context.Background()

	markdownContent := `# Architecture Overview
This is the system architecture for Lopor platform.

## Database Layer
The database layer relies on PostgreSQL 16 with pgvector extension.

### Connection Pool Configuration
Here is the connection pool configuration:
` + "```go\nfunc InitPool() *pgxpool.Pool {\n    cfg, _ := pgxpool.ParseConfig(\"postgres://...\")\n    return cfg\n}\n```\n" + `
Ensure maximum connection limits are observed.
`

	req := ChunkingRequest{
		WorkspaceID: "ws-chunk-1",
		DocumentID:  "doc-arch-01",
		Title:       "Architecture",
		RawContent:  markdownContent,
		Options: ChunkingOptions{
			TargetTokens:      200,
			MaxTokens:         400,
			EnrichBreadcrumbs: true,
		},
	}

	result, err := engine.ChunkDocument(ctx, req)
	if err != nil {
		t.Fatalf("ChunkDocument failed: %v", err)
	}

	if result.FormatDetected != FormatMarkdown {
		t.Errorf("expected FormatMarkdown, got %s", result.FormatDetected)
	}
	if len(result.Chunks) == 0 {
		t.Fatalf("expected chunks to be created, got 0")
	}

	// Verify breadcrumbs and code fence preservation
	foundCodeBlock := false
	foundBreadcrumb := false

	for _, c := range result.Chunks {
		if c.HasCodeBlock {
			foundCodeBlock = true
			if !strings.Contains(c.Content, "func InitPool()") {
				t.Errorf("expected chunk to preserve code block contents")
			}
		}
		if strings.Contains(c.BreadcrumbPath, "Database Layer") {
			foundBreadcrumb = true
		}
	}

	if !foundCodeBlock {
		t.Errorf("expected at least one chunk with HasCodeBlock = true")
	}
	if !foundBreadcrumb {
		t.Errorf("expected breadcrumb containing 'Database Layer'")
	}
}

func TestCodeAwareChunking(t *testing.T) {
	engine := NewChunkingEngine()
	ctx := context.Background()

	goCode := `package main

import "fmt"

func HandlerA() {
	fmt.Println("Handler A execution")
}

func HandlerB() {
	fmt.Println("Handler B execution")
}
`

	req := ChunkingRequest{
		WorkspaceID: "ws-chunk-1",
		DocumentID:  "doc-code-01",
		Title:       "handlers.go",
		RawContent:  goCode,
	}

	result, err := engine.ChunkDocument(ctx, req)
	if err != nil {
		t.Fatalf("ChunkDocument failed: %v", err)
	}

	if result.FormatDetected != FormatCode {
		t.Errorf("expected FormatCode, got %s", result.FormatDetected)
	}
	if result.StrategyUsed != StrategyCodeAware {
		t.Errorf("expected StrategyCodeAware, got %s", result.StrategyUsed)
	}
	if len(result.Chunks) < 2 {
		t.Errorf("expected at least 2 code chunks, got %d", len(result.Chunks))
	}
}

func TestTablePreservation(t *testing.T) {
	engine := NewChunkingEngine()
	ctx := context.Background()

	tableContent := `## Performance Benchmarks
Here are the throughput metrics:

| Metric | Target | Actual |
| :--- | :--- | :--- |
| Latency | < 50ms | 38ms |
| QPS | 5000 | 7200 |
`

	req := ChunkingRequest{
		WorkspaceID: "ws-chunk-1",
		DocumentID:  "doc-table-01",
		Title:       "benchmarks.md",
		RawContent:  tableContent,
	}

	result, err := engine.ChunkDocument(ctx, req)
	if err != nil {
		t.Fatalf("ChunkDocument failed: %v", err)
	}

	foundTable := false
	for _, c := range result.Chunks {
		if c.HasTable {
			foundTable = true
			if !strings.Contains(c.Content, "| Latency |") {
				t.Errorf("expected table row intact inside chunk")
			}
		}
	}

	if !foundTable {
		t.Errorf("expected HasTable = true on chunk containing markdown table")
	}
}
