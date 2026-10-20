package semanticcache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"
)

// SemanticCacheEngine coordinates vector similarity caching, query normalization, and cost elimination telemetry.
type SemanticCacheEngine struct {
	normalizer *Normalizer
	store      *MemoryStore
}

// NewSemanticCacheEngine initializes the semantic cache engine.
func NewSemanticCacheEngine(maxCapacity int) *SemanticCacheEngine {
	return &SemanticCacheEngine{
		normalizer: NewNormalizer(),
		store:      NewMemoryStore(maxCapacity),
	}
}

// Lookup queries the cache for semantically equivalent completions within threshold.
func (e *SemanticCacheEngine) Lookup(ctx context.Context, req CacheLookupRequest) (*CacheLookupResult, error) {
	if strings.TrimSpace(req.Query) == "" {
		return nil, errors.New("lookup query cannot be empty")
	}

	threshold := req.SimilarityThreshold
	if threshold <= 0.0 {
		threshold = 0.88
	}

	normQuery := e.normalizer.Normalize(req.Query)
	queryVec := e.normalizer.ComputeEmbedding(normQuery)

	start := time.Now()
	entry, sim := e.store.FindSimilar(req.WorkspaceID, req.ModelID, queryVec, threshold)

	if entry != nil && sim >= threshold {
		e.store.RecordHit(req.WorkspaceID, entry.ID, entry.TokensCount)
		// Approx 1500ms typical LLM round-trip saved vs sub-5ms in-memory cache return
		savedMs := int64(1500) - time.Since(start).Milliseconds()
		if savedMs < 500 {
			savedMs = 500
		}

		return &CacheLookupResult{
			Hit:             true,
			Entry:           entry,
			SimilarityScore: sim,
			LatencySavedMs:  savedMs,
			TokensSaved:     entry.TokensCount,
		}, nil
	}

	e.store.RecordMiss(req.WorkspaceID)
	return &CacheLookupResult{
		Hit:             false,
		Entry:           nil,
		SimilarityScore: sim,
		LatencySavedMs:  0,
		TokensSaved:     0,
	}, nil
}

// Store caches a query-response pair with normalized semantic embedding.
func (e *SemanticCacheEngine) Store(ctx context.Context, req CacheStoreRequest) error {
	if strings.TrimSpace(req.Query) == "" || strings.TrimSpace(req.Response) == "" {
		return errors.New("query and response cannot be empty")
	}

	normQuery := e.normalizer.Normalize(req.Query)
	queryVec := e.normalizer.ComputeEmbedding(normQuery)

	ttl := req.TTL
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}

	tokens := req.TokensCount
	if tokens <= 0 {
		// Heuristic token estimation if omitted
		tokens = len(strings.Fields(req.Response)) * 4 / 3
	}

	now := time.Now().UTC()
	hash := sha256.Sum256([]byte(req.WorkspaceID + ":" + req.ModelID + ":" + normQuery))
	entryID := "cache_" + hex.EncodeToString(hash[:8])

	entry := &CacheEntry{
		ID:              entryID,
		WorkspaceID:     req.WorkspaceID,
		ModelID:         req.ModelID,
		Query:           req.Query,
		NormalizedQuery: normQuery,
		Embedding:       queryVec,
		Response:        req.Response,
		HitCount:        0,
		TokensCount:     tokens,
		CreatedAt:       now,
		ExpiresAt:       now.Add(ttl),
		LastHitAt:       now,
	}

	e.store.Put(entry)
	return nil
}

// Invalidate purges cached entries for a workspace and model.
func (e *SemanticCacheEngine) Invalidate(ctx context.Context, wsID, modelID string) int {
	return e.store.Invalidate(wsID, modelID)
}

// GetTelemetry compiles metrics for a workspace.
func (e *SemanticCacheEngine) GetTelemetry(ctx context.Context, wsID string) *CacheTelemetry {
	t := e.store.GetTelemetry(wsID)
	return &t
}
