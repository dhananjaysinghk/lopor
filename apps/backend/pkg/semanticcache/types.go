package semanticcache

import (
	"time"
)

// CacheEntry represents a cached query-response pair with vector embedding and usage metrics.
type CacheEntry struct {
	ID              string    `json:"id"`
	WorkspaceID     string    `json:"workspace_id"`
	ModelID         string    `json:"model_id"`
	Query           string    `json:"query"`
	NormalizedQuery string    `json:"normalized_query"`
	Embedding       []float32 `json:"embedding,omitempty"`
	Response        string    `json:"response"`
	HitCount        int       `json:"hit_count"`
	TokensCount     int       `json:"tokens_count"`
	CreatedAt       time.Time `json:"created_at"`
	ExpiresAt       time.Time `json:"expires_at"`
	LastHitAt       time.Time `json:"last_hit_at"`
}

// CacheLookupRequest specifies query parameters for approximate vector cache retrieval.
type CacheLookupRequest struct {
	WorkspaceID         string  `json:"workspace_id"`
	ModelID             string  `json:"model_id"`
	Query               string  `json:"query"`
	SimilarityThreshold float64 `json:"similarity_threshold,omitempty"` // Default 0.88
}

// CacheLookupResult provides the lookup decision, matched content, and efficiency savings.
type CacheLookupResult struct {
	Hit             bool        `json:"hit"`
	Entry           *CacheEntry `json:"entry,omitempty"`
	SimilarityScore float64     `json:"similarity_score"`
	LatencySavedMs  int64       `json:"latency_saved_ms"`
	TokensSaved     int         `json:"tokens_saved"`
}

// CacheStoreRequest provides completion payload to cache.
type CacheStoreRequest struct {
	WorkspaceID string        `json:"workspace_id"`
	ModelID     string        `json:"model_id"`
	Query       string        `json:"query"`
	Response    string        `json:"response"`
	TokensCount int           `json:"tokens_count"`
	TTL         time.Duration `json:"ttl,omitempty"` // Default 24h
}

// CacheTelemetry compiles efficiency, hit-rate, and financial savings metrics.
type CacheTelemetry struct {
	WorkspaceID         string  `json:"workspace_id"`
	TotalEntries        int     `json:"total_entries"`
	TotalLookups        int64   `json:"total_lookups"`
	TotalHits           int64   `json:"total_hits"`
	TotalMisses         int64   `json:"total_misses"`
	HitRate             float64 `json:"hit_rate"` // 0.0 to 100.0%
	TotalTokensSaved    int64   `json:"total_tokens_saved"`
	EstimatedCostSaved  float64 `json:"estimated_cost_saved"` // In USD ($)
	AverageHitLatencyMs float64 `json:"average_hit_latency_ms"`
}
