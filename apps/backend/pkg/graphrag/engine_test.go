package graphrag

import (
	"context"
	"strings"
	"testing"
)

func TestTripletExtractionAndIngestion(t *testing.T) {
	engine := NewGraphRAGEngine()
	ctx := context.Background()

	docText := `
	Next.js integrates with Go Fiber.
	Go Fiber stores data in PostgreSQL.
	Go Fiber uses Redis.
	PostgreSQL implements pgvector.
	`

	req := GraphExtractionRequest{
		WorkspaceID: "ws-arch-1",
		DocumentID:  "doc-arch-01",
		Content:     docText,
	}

	res, err := engine.ExtractAndIngest(ctx, req)
	if err != nil {
		t.Fatalf("ExtractAndIngest failed: %v", err)
	}

	if len(res.ExtractedNodes) == 0 {
		t.Fatalf("expected extracted nodes, got 0")
	}
	if len(res.ExtractedEdges) == 0 {
		t.Fatalf("expected extracted edges, got 0")
	}

	// Verify specific extracted entities
	foundFiber := false
	foundPostgres := false
	for _, n := range res.ExtractedNodes {
		if strings.Contains(n.Name, "Go Fiber") {
			foundFiber = true
		}
		if strings.Contains(n.Name, "PostgreSQL") {
			foundPostgres = true
		}
	}

	if !foundFiber {
		t.Errorf("expected 'Go Fiber' node to be extracted")
	}
	if !foundPostgres {
		t.Errorf("expected 'PostgreSQL' node to be extracted")
	}

	// Verify degrees
	visual, err := engine.GetGraphVisual(ctx, "ws-arch-1")
	if err != nil {
		t.Fatalf("GetGraphVisual failed: %v", err)
	}
	if visual.TotalNodes == 0 {
		t.Errorf("expected total nodes > 0, got 0")
	}
	if visual.TotalEdges == 0 {
		t.Errorf("expected total edges > 0, got 0")
	}
}

func TestLocalGraphSearch(t *testing.T) {
	engine := NewGraphRAGEngine()
	ctx := context.Background()

	docText := `
	Kafka integrates with Flink.
	Flink stores data in Cassandra.
	Cassandra depends on RocksDB.
	`

	_, _ = engine.ExtractAndIngest(ctx, GraphExtractionRequest{
		WorkspaceID: "ws-stream-1",
		DocumentID:  "doc-stream-01",
		Content:     docText,
	})

	localReq := LocalGraphSearchRequest{
		WorkspaceID: "ws-stream-1",
		Query:       "Kafka",
		MaxDepth:    2,
	}

	localRes, err := engine.LocalSearch(ctx, localReq)
	if err != nil {
		t.Fatalf("LocalSearch failed: %v", err)
	}

	if len(localRes.FocusedNodes) == 0 {
		t.Fatalf("expected focused nodes, got 0")
	}
	if !strings.Contains(localRes.ContextSummary, "Local Knowledge Subgraph") {
		t.Errorf("expected context summary to contain graph header")
	}
	if localRes.Confidence <= 0.0 {
		t.Errorf("expected confidence > 0, got %.2f", localRes.Confidence)
	}
}

func TestGlobalGraphSearch(t *testing.T) {
	engine := NewGraphRAGEngine()
	ctx := context.Background()

	docText := `
	Gateway integrates with AuthServer.
	AuthServer stores data in Vault.
	Gateway integrates with BillingService.
	BillingService stores data in StripeAPI.
	`

	_, _ = engine.ExtractAndIngest(ctx, GraphExtractionRequest{
		WorkspaceID: "ws-global-1",
		DocumentID:  "doc-global-01",
		Content:     docText,
	})

	globalReq := GlobalGraphSearchRequest{
		WorkspaceID: "ws-global-1",
		Query:       "What are the main security and payment architectural services?",
	}

	globalRes, err := engine.GlobalSearch(ctx, globalReq)
	if err != nil {
		t.Fatalf("GlobalSearch failed: %v", err)
	}

	if len(globalRes.ReferencedCommunities) == 0 {
		t.Fatalf("expected referenced communities, got 0")
	}
	if !strings.Contains(globalRes.SynthesizedAnswer, "Holistic Graph Synthesis") {
		t.Errorf("expected synthesized answer header")
	}
	if len(globalRes.RelevantThemes) == 0 {
		t.Errorf("expected relevant themes, got 0")
	}
}
