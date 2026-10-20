package semanticcache

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestExactMatchLookup(t *testing.T) {
	engine := NewSemanticCacheEngine(100)
	ctx := context.Background()

	wsID := "ws-cache-1"
	modelID := "gpt-4o"
	query := "What are the platform rate limits for API tokens?"
	resp := "The platform enforces 100 requests per minute on free tier and 5000 requests per minute on enterprise."

	err := engine.Store(ctx, CacheStoreRequest{
		WorkspaceID: wsID,
		ModelID:     modelID,
		Query:       query,
		Response:    resp,
		TokensCount: 25,
		TTL:         1 * time.Hour,
	})
	if err != nil {
		t.Fatalf("Store failed: %v", err)
	}

	lookupRes, err := engine.Lookup(ctx, CacheLookupRequest{
		WorkspaceID: wsID,
		ModelID:     modelID,
		Query:       query,
	})
	if err != nil {
		t.Fatalf("Lookup failed: %v", err)
	}

	if !lookupRes.Hit {
		t.Errorf("expected exact match to hit cache")
	}
	if lookupRes.Entry == nil || lookupRes.Entry.Response != resp {
		t.Errorf("expected cached response '%s', got '%v'", resp, lookupRes.Entry)
	}
	if lookupRes.SimilarityScore < 0.95 {
		t.Errorf("expected similarity >= 0.95 for exact match, got %.3f", lookupRes.SimilarityScore)
	}
}

func TestApproximateSemanticMatch(t *testing.T) {
	engine := NewSemanticCacheEngine(100)
	ctx := context.Background()

	wsID := "ws-cache-1"
	modelID := "gpt-4o"
	storedQuery := "How do I configure postgres database connection pool settings?"
	storedResp := "Configure pgxpool using MaxConns and MinConns in postgres:// connection URI."

	_ = engine.Store(ctx, CacheStoreRequest{
		WorkspaceID: wsID,
		ModelID:     modelID,
		Query:       storedQuery,
		Response:    storedResp,
		TokensCount: 50,
	})

	// Rephrased query with slight word variance
	rephrasedQuery := "How can I configure the postgres database connection pool settings?"
	lookupRes, err := engine.Lookup(ctx, CacheLookupRequest{
		WorkspaceID:         wsID,
		ModelID:             modelID,
		Query:               rephrasedQuery,
		SimilarityThreshold: 0.85,
	})
	if err != nil {
		t.Fatalf("Lookup failed: %v", err)
	}

	if !lookupRes.Hit {
		t.Errorf("expected rephrased query to hit semantic cache with similarity >= 0.85, got similarity %.3f", lookupRes.SimilarityScore)
	}
	if !strings.Contains(lookupRes.Entry.Response, "pgxpool") {
		t.Errorf("expected matched response")
	}
}

func TestDissimilarQueryMiss(t *testing.T) {
	engine := NewSemanticCacheEngine(100)
	ctx := context.Background()

	wsID := "ws-cache-1"
	modelID := "gpt-4o"

	_ = engine.Store(ctx, CacheStoreRequest{
		WorkspaceID: wsID,
		ModelID:     modelID,
		Query:       "Explain the difference between TCP and UDP protocols",
		Response:    "TCP is connection-oriented while UDP is connectionless.",
		TokensCount: 30,
	})

	unrelatedQuery := "What is the capital of France?"
	lookupRes, err := engine.Lookup(ctx, CacheLookupRequest{
		WorkspaceID: wsID,
		ModelID:     modelID,
		Query:       unrelatedQuery,
	})
	if err != nil {
		t.Fatalf("Lookup failed: %v", err)
	}

	if lookupRes.Hit {
		t.Errorf("expected unrelated query to miss cache, got hit with similarity %.3f", lookupRes.SimilarityScore)
	}
}

func TestWorkspaceIsolation(t *testing.T) {
	engine := NewSemanticCacheEngine(100)
	ctx := context.Background()

	_ = engine.Store(ctx, CacheStoreRequest{
		WorkspaceID: "ws-secret-alpha",
		ModelID:     "gpt-4o",
		Query:       "Secret project codename",
		Response:    "Project Titan",
		TokensCount: 10,
	})

	// Lookup from different workspace
	res, err := engine.Lookup(ctx, CacheLookupRequest{
		WorkspaceID: "ws-other-beta",
		ModelID:     "gpt-4o",
		Query:       "Secret project codename",
	})
	if err != nil {
		t.Fatalf("Lookup failed: %v", err)
	}

	if res.Hit {
		t.Errorf("expected cross-workspace lookup to be isolated and miss, but hit")
	}
}

func TestTelemetryAccounting(t *testing.T) {
	engine := NewSemanticCacheEngine(100)
	ctx := context.Background()

	wsID := "ws-telemetry-test"
	modelID := "gpt-4o"

	_ = engine.Store(ctx, CacheStoreRequest{
		WorkspaceID: wsID,
		ModelID:     modelID,
		Query:       "Benchmark query test",
		Response:    "Benchmark answer",
		TokensCount: 100,
	})

	// 1 hit
	_, _ = engine.Lookup(ctx, CacheLookupRequest{
		WorkspaceID: wsID,
		ModelID:     modelID,
		Query:       "Benchmark query test",
	})

	// 1 miss
	_, _ = engine.Lookup(ctx, CacheLookupRequest{
		WorkspaceID: wsID,
		ModelID:     modelID,
		Query:       "Uncached different query",
	})

	telemetry := engine.GetTelemetry(ctx, wsID)
	if telemetry.TotalLookups != 2 {
		t.Errorf("expected 2 lookups, got %d", telemetry.TotalLookups)
	}
	if telemetry.TotalHits != 1 {
		t.Errorf("expected 1 hit, got %d", telemetry.TotalHits)
	}
	if telemetry.TotalMisses != 1 {
		t.Errorf("expected 1 miss, got %d", telemetry.TotalMisses)
	}
	if telemetry.HitRate != 50.0 {
		t.Errorf("expected 50.0%% hit rate, got %.1f%%", telemetry.HitRate)
	}
	if telemetry.TotalTokensSaved != 100 {
		t.Errorf("expected 100 tokens saved, got %d", telemetry.TotalTokensSaved)
	}
}
