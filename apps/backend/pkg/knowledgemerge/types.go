package knowledgemerge

import (
	"time"
)

// ResolutionStrategy defines how contradictory factual assertions are resolved.
type ResolutionStrategy string

const (
	StrategyRecencyWeighted   ResolutionStrategy = "RECENCY_WEIGHTED"    // Latest timestamp takes precedence
	StrategyConfidenceVote    ResolutionStrategy = "CONFIDENCE_VOTE"     // Highest multi-source confidence wins
	StrategyAuthorityHierarchy ResolutionStrategy = "AUTHORITY_HIERARCHY" // Verified human/admin overrides AI inference
	StrategySemanticSynthesis ResolutionStrategy = "SEMANTIC_SYNTHESIS"  // Nuanced union of overlapping non-mutually exclusive facts
)

// KnowledgeEntity represents a deduplicated canonical entity in the unified graph.
type KnowledgeEntity struct {
	EntityID      string            `json:"entity_id"`
	WorkspaceID   string            `json:"workspace_id"`
	CanonicalName string            `json:"canonical_name"`
	Aliases       []string          `json:"aliases"`
	Type          string            `json:"type"` // e.g. PERSON, ORGANIZATION, CONCEPT, SYSTEM
	Attributes    map[string]string `json:"attributes"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

// KnowledgeFact represents a relational statement connecting a Subject to an Object.
type KnowledgeFact struct {
	FactID         string    `json:"fact_id"`
	WorkspaceID    string    `json:"workspace_id"`
	SourceID       string    `json:"source_id"` // Document ID or Agent ID that asserted this fact
	SubjectID      string    `json:"subject_id"`
	SubjectName    string    `json:"subject_name"`
	Predicate      string    `json:"predicate"`
	ObjectID       string    `json:"object_id"`
	ObjectName     string    `json:"object_name"`
	Confidence     float64   `json:"confidence"` // 0.0 - 1.0
	SourceType     string    `json:"source_type"` // "HUMAN_EXPERT", "RAG_EXTRACTION", "AI_AGENT"
	SourceCitation string    `json:"source_citation,omitempty"`
	ValidFrom      time.Time `json:"valid_from"`
	RecordedAt     time.Time `json:"recorded_at"`
}

// FactConflict represents a contradictory or competing relationship between the same Subject and Predicate.
type FactConflict struct {
	ConflictID       string             `json:"conflict_id"`
	SubjectName      string             `json:"subject_name"`
	Predicate        string             `json:"predicate"`
	ConflictingFacts []KnowledgeFact    `json:"conflicting_facts"`
	ResolvedFact     *KnowledgeFact     `json:"resolved_fact,omitempty"`
	StrategyUsed     ResolutionStrategy `json:"strategy_used"`
	Explanation      string             `json:"explanation"`
	IsResolved       bool               `json:"is_resolved"`
}

// IngestKnowledgeGraphRequest provides external or newly extracted graphs to reconcile.
type IngestKnowledgeGraphRequest struct {
	WorkspaceID string          `json:"workspace_id"`
	Entities    []KnowledgeEntity `json:"entities"`
	Facts       []KnowledgeFact   `json:"facts"`
	DefaultStrategy ResolutionStrategy `json:"default_strategy,omitempty"`
}

// ReconcileResult summarizes the graph unification, entity merges, and conflict resolutions.
type ReconcileResult struct {
	WorkspaceID       string         `json:"workspace_id"`
	TotalEntities     int            `json:"total_entities"`
	TotalFacts        int            `json:"total_facts"`
	MergedEntities    int            `json:"merged_entities"`
	ConflictsDetected int            `json:"conflicts_detected"`
	ConflictsResolved int            `json:"conflicts_resolved"`
	Conflicts         []FactConflict `json:"conflicts"`
	ReconciledAt      time.Time      `json:"reconciled_at"`
}

