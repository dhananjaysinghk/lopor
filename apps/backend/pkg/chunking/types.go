package chunking

import (
	"time"
)

// ChunkStrategy defines the structural splitting methodology.
type ChunkStrategy string

const (
	StrategyAuto                 ChunkStrategy = "AUTO"
	StrategyMarkdownHierarchical ChunkStrategy = "MARKDOWN_HIERARCHICAL"
	StrategyCodeAware            ChunkStrategy = "CODE_AWARE"
	StrategyTablePreserving      ChunkStrategy = "TABLE_PRESERVING"
	StrategySemanticSentence     ChunkStrategy = "SEMANTIC_SENTENCE"
)

// ContentFormat classifies the detected structure of the source text.
type ContentFormat string

const (
	FormatMarkdown ContentFormat = "MARKDOWN"
	FormatCode     ContentFormat = "CODE"
	FormatTable    ContentFormat = "TABLE"
	FormatProse    ContentFormat = "PROSE"
)

// EnrichedChunk represents an enriched, structurally bounded text unit for RAG embeddings.
type EnrichedChunk struct {
	ID             string        `json:"id"`
	DocumentID     string        `json:"document_id"`
	WorkspaceID    string        `json:"workspace_id"`
	ChunkIndex     int           `json:"chunk_index"`
	Content        string        `json:"content"`
	BreadcrumbPath string        `json:"breadcrumb_path,omitempty"` // e.g. "Architecture > Database > PostgreSQL"
	Format         ContentFormat `json:"format"`
	TokenCount     int           `json:"token_count"`
	HasCodeBlock   bool          `json:"has_code_block"`
	HasTable       bool          `json:"has_table"`
	KeyPhrases     []string      `json:"key_phrases,omitempty"`
	StartOffset    int           `json:"start_offset"`
	EndOffset      int           `json:"end_offset"`
	CreatedAt      time.Time     `json:"created_at"`
}

// ChunkingOptions governs budget boundaries, overlap, and enrichment flags.
type ChunkingOptions struct {
	TargetTokens      int           `json:"target_tokens"`       // default 400
	MaxTokens         int           `json:"max_tokens"`          // default 600
	OverlapTokens     int           `json:"overlap_tokens"`      // default 50
	MinOrphanTokens   int           `json:"min_orphan_tokens"`    // default 50
	Strategy          ChunkStrategy `json:"strategy"`            // default StrategyAuto
	EnrichBreadcrumbs bool          `json:"enrich_breadcrumbs"`  // default true
}

// ChunkingRequest provides document text and target configuration for chunking.
type ChunkingRequest struct {
	WorkspaceID string          `json:"workspace_id"`
	DocumentID  string          `json:"document_id"`
	Title       string          `json:"title"`
	RawContent  string          `json:"raw_content"`
	Options     ChunkingOptions `json:"options"`
}

// ChunkingResult encapsulates the structured chunks and statistical telemetry.
type ChunkingResult struct {
	DocumentID         string          `json:"document_id"`
	FormatDetected     ContentFormat   `json:"format_detected"`
	StrategyUsed       ChunkStrategy   `json:"strategy_used"`
	TotalChunks        int             `json:"total_chunks"`
	TotalTokens        int             `json:"total_tokens"`
	AverageChunkTokens float64         `json:"average_chunk_tokens"`
	Chunks             []EnrichedChunk `json:"chunks"`
	ProcessedAt        time.Time       `json:"processed_at"`
}
