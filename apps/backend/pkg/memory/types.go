package memory

import (
	"time"
)

// MemoryCategory classifies the type of long-term episodic knowledge.
type MemoryCategory string

const (
	CategoryPreference MemoryCategory = "PREFERENCE"
	CategoryTechStack  MemoryCategory = "TECH_STACK"
	CategoryDomainFact MemoryCategory = "DOMAIN_FACT"
	CategoryConstraint MemoryCategory = "CONSTRAINT"
	CategoryGoal       MemoryCategory = "GOAL"
)

// MemoryItem represents an episodic memory entity preserved across sessions.
type MemoryItem struct {
	ID             string         `json:"id"`
	WorkspaceID    string         `json:"workspace_id"`
	UserID         string         `json:"user_id"`
	Category       MemoryCategory `json:"category"`
	Content        string         `json:"content"`
	Keywords       []string       `json:"keywords"`
	Importance     float64        `json:"importance"`       // 0.0 to 1.0 scale
	Confidence     float64        `json:"confidence"`       // 0.0 to 1.0 extraction confidence
	AccessCount    int            `json:"access_count"`     // Frequency of retrieval
	CreatedAt      time.Time      `json:"created_at"`
	LastAccessedAt time.Time      `json:"last_accessed_at"`
	DecayScore     float64        `json:"decay_score"`      // Time-decayed retention weight
	IsActive       bool           `json:"is_active"`
}

// DialogueTurn models a single conversational exchange turn.
type DialogueTurn struct {
	Role      string    `json:"role"` // "user", "assistant", "system"
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// ExtractionRequest captures the dialogue to extract memories from.
type ExtractionRequest struct {
	WorkspaceID     string         `json:"workspace_id"`
	UserID          string         `json:"user_id"`
	DialogueHistory []DialogueTurn `json:"dialogue_history"`
	CurrentMessage  string         `json:"current_message"`
}

// ExtractionResult captures the newly identified and stored memories.
type ExtractionResult struct {
	ExtractedMemories    []MemoryItem     `json:"extracted_memories"`
	IdentifiedCategories []MemoryCategory `json:"identified_categories"`
	ProcessedTurnCount   int              `json:"processed_turn_count"`
}

// RecallRequest defines parameters for retrieving relevant context memories.
type RecallRequest struct {
	WorkspaceID    string           `json:"workspace_id"`
	UserID         string           `json:"user_id"`
	Query          string           `json:"query"`
	CategoryFilter []MemoryCategory `json:"category_filter,omitempty"`
	MaxResults     int              `json:"max_results,omitempty"`     // Defaults to 5
	MinRelevance   float64          `json:"min_relevance,omitempty"`   // Defaults to 0.25
}

// RecalledMemory pairs a memory with ranking metrics.
type RecalledMemory struct {
	Memory         MemoryItem `json:"memory"`
	RelevanceScore float64    `json:"relevance_score"`
	RecencyScore   float64    `json:"recency_score"`
	CombinedScore  float64    `json:"combined_score"`
}

// RecallResult provides retrieved memories ready for LLM prompt augmentation.
type RecallResult struct {
	Items                []RecalledMemory `json:"items"`
	FormattedPromptBlock string           `json:"formatted_prompt_block"`
	TotalMemoriesScanned int              `json:"total_memories_scanned"`
}

// ConsolidationReport summarizes the maintenance and deduplication run.
type ConsolidationReport struct {
	WorkspaceID    string    `json:"workspace_id"`
	UserID         string    `json:"user_id"`
	TotalBefore    int       `json:"total_before"`
	TotalAfter     int       `json:"total_after"`
	MergedCount    int       `json:"merged_count"`
	PrunedCount    int       `json:"pruned_count"`
	ConsolidatedAt time.Time `json:"consolidated_at"`
}
