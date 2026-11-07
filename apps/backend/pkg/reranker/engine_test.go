package reranker

import (
	"context"
	"strings"
	"testing"
)

func TestReRankRelevanceScoringAndOrdering(t *testing.T) {
	engine := NewEngine()
	ctx := context.Background()
	wsID := "ws_test_rerank"

	req := ReRankRequest{
		Query: "PostgreSQL vector index tuning pgvector HNSW",
		Chunks: []DocumentChunk{
			{
				ID:      "c1",
				Title:   "Sourdough Baking Recipe",
				Content: "Mix flour, water, and salt together. Let ferment for 12 hours before baking in a Dutch oven.",
			},
			{
				ID:      "c2",
				Title:   "HNSW pgvector Optimization Guide",
				Content: "For optimal PostgreSQL vector index tuning pgvector HNSW, configure m=16 and ef_construction=64 to balance recall and build time.",
			},
			{
				ID:      "c3",
				Title:   "Relational Index Overview",
				Content: "PostgreSQL supports standard B-tree and GIN indexes for relational query acceleration across transactional tables.",
			},
		},
		MaxTopK:              5,
		RelevanceCutoff:      0.20,
		LostInTheMiddleOrder: false,
	}

	result, err := engine.ReRank(ctx, wsID, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.RetainedChunksCount == 0 {
		t.Fatalf("expected retained chunks, got 0")
	}

	// Chunk c2 should be top ranked
	topChunk := result.ScoredChunks[0]
	if topChunk.ID != "c2" {
		t.Errorf("expected top chunk to be 'c2', got '%s'", topChunk.ID)
	}
	if topChunk.RelevanceGrade != GradeHigh {
		t.Errorf("expected GradeHigh for exact matching chunk, got %s", topChunk.RelevanceGrade)
	}
	if topChunk.CombinedScore < 0.70 {
		t.Errorf("expected high score >= 0.70, got %.3f", topChunk.CombinedScore)
	}

	if !strings.Contains(result.AssembledPromptContext, "HNSW pgvector Optimization Guide") {
		t.Errorf("assembled prompt context missing top chunk content")
	}
}

func TestLostInTheMiddleUPlacement(t *testing.T) {
	engine := NewEngine()
	ctx := context.Background()
	wsID := "ws_test_litm"

	// Create 5 chunks with varying degrees of relevance
	req := ReRankRequest{
		Query: "Distributed Raft consensus state machine replication",
		Chunks: []DocumentChunk{
			{ID: "c1", Title: "Raft Consensus Protocol", Content: "Distributed Raft consensus state machine replication ensures strong consistency across cluster nodes."}, // 1st
			{ID: "c2", Title: "Paxos Agreement", Content: "Paxos consensus provides state machine replication across unreliable networks."},                                     // 2nd
			{ID: "c3", Title: "Gossip Dissemination", Content: "Gossip protocols spread state probabilistically across peer nodes."},                                            // 4th
			{ID: "c4", Title: "Two-Phase Commit", Content: "Two-phase commit coordinates transactions across distributed nodes."},                                                // 3rd
			{ID: "c5", Title: "Unrelated Cooking", Content: "Simmer the sauce slowly over medium heat for thirty minutes."},                                                      // Pruned or 5th
		},
		MaxTopK:              5,
		RelevanceCutoff:      0.10,
		LostInTheMiddleOrder: true,
	}

	result, err := engine.ReRank(ctx, wsID, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.LostInTheMiddleApplied {
		t.Errorf("expected LostInTheMiddleApplied to be true")
	}

	if len(result.ScoredChunks) < 4 {
		t.Fatalf("expected at least 4 retained chunks, got %d", len(result.ScoredChunks))
	}

	// In U-shaped distribution, chunk 0 must have highest or near highest score,
	// and last chunk must have second highest score
	first := result.ScoredChunks[0]
	last := result.ScoredChunks[len(result.ScoredChunks)-1]

	if first.ID != "c1" {
		t.Errorf("expected extreme beginning (index 0) to hold highest scored chunk 'c1', got '%s'", first.ID)
	}
	if last.ID != "c2" {
		t.Errorf("expected extreme end (last index) to hold 2nd highest scored chunk 'c2', got '%s'", last.ID)
	}
}

func TestTokenBudgetConstraint(t *testing.T) {
	engine := NewEngine()
	ctx := context.Background()
	wsID := "ws_test_budget"

	req := ReRankRequest{
		Query: "Machine learning model weights",
		Chunks: []DocumentChunk{
			{ID: "b1", Title: "Weights 1", Content: "Machine learning model weights optimization step one.", TokenEstimate: 100},
			{ID: "b2", Title: "Weights 2", Content: "Machine learning model weights optimization step two.", TokenEstimate: 150},
			{ID: "b3", Title: "Weights 3", Content: "Machine learning model weights optimization step three.", TokenEstimate: 200},
		},
		MaxTokensBudget: 220, // Should only fit b1 (100) and maybe b2 if total <= 220, but not all three
		RelevanceCutoff: 0.10,
	}

	result, err := engine.ReRank(ctx, wsID, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.TotalTokensRetained > 220 {
		t.Errorf("expected tokens <= 220, got %d", result.TotalTokensRetained)
	}
}

func TestRelevanceCutoffPruning(t *testing.T) {
	engine := NewEngine()
	ctx := context.Background()
	wsID := "ws_test_cutoff"

	req := ReRankRequest{
		Query: "Quantum cryptography key distribution",
		Chunks: []DocumentChunk{
			{ID: "q1", Title: "Quantum Crypto", Content: "Quantum cryptography key distribution leverages photon polarization states."},
			{ID: "q2", Title: "Banana Bread", Content: "Bake banana bread with ripe bananas and cinnamon."},
			{ID: "q3", Title: "Garden Plants", Content: "Water the tomato plants every morning during the dry summer."},
		},
		RelevanceCutoff: 0.50, // Strict threshold
	}

	result, err := engine.ReRank(ctx, wsID, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.RetainedChunksCount != 1 {
		t.Errorf("expected exactly 1 retained chunk above 0.50 cutoff, got %d", result.RetainedChunksCount)
	}
	if result.PrunedChunksCount != 2 {
		t.Errorf("expected 2 pruned chunks, got %d", result.PrunedChunksCount)
	}
	if result.PruningRatioPercent < 60.0 {
		t.Errorf("expected pruning ratio >= 60%%, got %.1f%%", result.PruningRatioPercent)
	}
}

func TestRerankerMetricsAccumulation(t *testing.T) {
	engine := NewEngine()
	ctx := context.Background()
	wsID := "ws_test_metrics"

	req := ReRankRequest{
		Query: "Kubernetes pod networking",
		Chunks: []DocumentChunk{
			{ID: "k1", Content: "Kubernetes pod networking utilizes CNI plugins."},
			{ID: "k2", Content: "Unrelated topic content."},
		},
	}

	_, _ = engine.ReRank(ctx, wsID, req)
	_, _ = engine.ReRank(ctx, wsID, req)

	metrics := engine.GetMetrics(wsID)
	if metrics.TotalRequests != 2 {
		t.Errorf("expected 2 requests recorded, got %d", metrics.TotalRequests)
	}
	if metrics.CumulativeCandidates != 4 {
		t.Errorf("expected 4 cumulative candidates, got %d", metrics.CumulativeCandidates)
	}
	if metrics.AverageExecutionDurationMs <= 0 {
		t.Errorf("expected positive average execution duration")
	}
}

