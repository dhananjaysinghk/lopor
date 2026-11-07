package reranker

import (
	"time"
)

// RelevanceGrade characterizes the semantic utility of a retrieved chunk.
type RelevanceGrade string

const (
	GradeHigh       RelevanceGrade = "HIGH"
	GradeMedium     RelevanceGrade = "MEDIUM"
	GradeLow        RelevanceGrade = "LOW"
	GradeIrrelevant RelevanceGrade = "IRRELEVANT"
)

// RerankStrategy represents the scoring methodology.
type RerankStrategy string

const (
	StrategyCrossEncoder   RerankStrategy = "CROSS_ENCODER"
	StrategyHybridLexical  RerankStrategy = "HYBRID_LEXICAL_DENSE"
	StrategyReciprocalRank RerankStrategy = "RECIPROCAL_RANK"
)

// DocumentChunk represents a candidate chunk retrieved from initial vector search or full-text index.
type DocumentChunk struct {
	ID            string            `json:"id"`
	DocumentID    string            `json:"document_id,omitempty"`
	Title         string            `json:"title,omitempty"`
	Content       string            `json:"content"`
	InitialScore  float64           `json:"initial_score,omitempty"` // Vector cosine similarity or BM25 score
	TokenEstimate int               `json:"token_estimate,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

// ReRankRequest encapsulates input parameters for adaptive re-ranking and context pruning.
type ReRankRequest struct {
	Query               string          `json:"query"`
	Chunks              []DocumentChunk `json:"chunks"`
	MaxTopK             int             `json:"max_top_k,omitempty"`              // Default: 5
	RelevanceCutoff     float64         `json:"relevance_cutoff,omitempty"`       // Minimum score threshold (default: 0.35)
	MaxTokensBudget     int             `json:"max_tokens_budget,omitempty"`      // Maximum token capacity (default: 4000)
	Strategy            RerankStrategy  `json:"strategy,omitempty"`                // Default: CROSS_ENCODER
	LostInTheMiddleOrder bool            `json:"lost_in_the_middle_order,omitempty"` // U-shaped context ordering
}

// ScoredChunk represents a chunk post cross-encoder scoring and positional adjustment.
type ScoredChunk struct {
	DocumentChunk
	CrossEncoderScore float64        `json:"cross_encoder_score"`
	CombinedScore     float64        `json:"combined_score"`
	RelevanceGrade    RelevanceGrade `json:"relevance_grade"`
	InitialRank       int            `json:"initial_rank"`
	FinalRank         int            `json:"final_rank"`
	PositionDelta     int            `json:"position_delta"` // Difference between initial and final rank
}

// ReRankResult encapsulates the final re-ranked context and operational statistics.
type ReRankResult struct {
	RequestID              string        `json:"request_id"`
	WorkspaceID            string        `json:"workspace_id"`
	Query                  string        `json:"query"`
	StrategyUsed           RerankStrategy `json:"strategy_used"`
	TotalCandidateChunks   int           `json:"total_candidate_chunks"`
	RetainedChunksCount    int           `json:"retained_chunks_count"`
	PrunedChunksCount      int           `json:"pruned_chunks_count"`
	PruningRatioPercent    float64       `json:"pruning_ratio_percent"`
	TotalTokensRetained    int           `json:"total_tokens_retained"`
	ExecutionDurationMs    float64       `json:"execution_duration_ms"`
	LostInTheMiddleApplied bool          `json:"lost_in_the_middle_applied"`
	ScoredChunks           []ScoredChunk `json:"scored_chunks"`
	AssembledPromptContext string        `json:"assembled_prompt_context"`
	ProcessedAt            time.Time     `json:"processed_at"`
}

// RerankerMetrics tracks workspace-level context optimization and savings telemetry.
type RerankerMetrics struct {
	WorkspaceID                 string    `json:"workspace_id"`
	TotalRequests               int64     `json:"total_requests"`
	CumulativeCandidates        int64     `json:"cumulative_candidates"`
	CumulativeRetained          int64     `json:"cumulative_retained"`
	CumulativePruned            int64     `json:"cumulative_pruned"`
	AveragePruningRatioPercent  float64   `json:"average_pruning_ratio_percent"`
	AverageExecutionDurationMs  float64   `json:"average_execution_duration_ms"`
	LastUpdated                 time.Time `json:"last_updated"`
}

