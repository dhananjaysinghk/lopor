package docsync

import (
	"time"
)

// DocumentFormat defines the serialization format of the document.
type DocumentFormat string

const (
	FormatMarkdown DocumentFormat = "MARKDOWN"
	FormatJSON     DocumentFormat = "JSON"
	FormatYAML     DocumentFormat = "YAML"
	FormatProse    DocumentFormat = "PROSE"
)

// SyncStrategy specifies the convergence algorithm applied during conflicting edits.
type SyncStrategy string

const (
	StrategyThreeWayMerge SyncStrategy = "THREE_WAY_MERGE"
	StrategyLastWriteWins SyncStrategy = "LAST_WRITE_WINS"
	StrategySemanticAI    SyncStrategy = "SEMANTIC_AI_ARBITRATE"
)

// TextOperationType categorizes the character or block level mutation.
type TextOperationType string

const (
	OpInsert TextOperationType = "INSERT"
	OpDelete TextOperationType = "DELETE"
	OpRetain TextOperationType = "RETAIN"
)

// TextOperation represents an atomic operational transformation mutation.
type TextOperation struct {
	Type   TextOperationType `json:"type"`
	Length int               `json:"length"`
	Text   string            `json:"text,omitempty"`
}

// DocumentRevision captures an immutable snapshot along the document version DAG.
type DocumentRevision struct {
	RevisionID    string            `json:"revision_id"`
	DocumentID    string            `json:"document_id"`
	WorkspaceID   string            `json:"workspace_id"`
	ParentRevID   string            `json:"parent_rev_id,omitempty"`
	AuthorID      string            `json:"author_id"`
	AuthorName    string            `json:"author_name"`
	IsAIAgent     bool              `json:"is_ai_agent"`
	Content       string            `json:"content"`
	ContentHash   string            `json:"content_hash"`
	Operations    []TextOperation   `json:"operations,omitempty"`
	CommitMessage string            `json:"commit_message"`
	Timestamp     time.Time         `json:"timestamp"`
}

// ConflictHunk identifies a localized diverging segment between two concurrent revisions.
type ConflictHunk struct {
	HunkID          string `json:"hunk_id"`
	BaseContent     string `json:"base_content"`
	LocalContent    string `json:"local_content"`
	RemoteContent   string `json:"remote_content"`
	ResolvedContent string `json:"resolved_content,omitempty"`
	IsResolved      bool   `json:"is_resolved"`
	ResolutionNote  string `json:"resolution_note,omitempty"`
}

// MergeResult details the outcome of synchronizing two diverging branch revisions.
type MergeResult struct {
	DocumentID       string         `json:"document_id"`
	BaseRevID        string         `json:"base_rev_id"`
	LocalRevID       string         `json:"local_rev_id"`
	RemoteRevID      string         `json:"remote_rev_id"`
	HasConflicts     bool           `json:"has_conflicts"`
	ConflictsCount   int            `json:"conflicts_count"`
	StrategyUsed     SyncStrategy   `json:"strategy_used"`
	MergedContent    string         `json:"merged_content"`
	MergedRevID      string         `json:"merged_rev_id,omitempty"`
	Conflicts        []ConflictHunk `json:"conflicts,omitempty"`
	MergedAt         time.Time      `json:"merged_at"`
}

// CommitRevisionRequest defines parameters to append a new revision.
type CommitRevisionRequest struct {
	WorkspaceID   string         `json:"workspace_id"`
	DocumentID    string         `json:"document_id"`
	ParentRevID   string         `json:"parent_rev_id,omitempty"`
	AuthorID      string         `json:"author_id"`
	AuthorName    string         `json:"author_name"`
	IsAIAgent     bool           `json:"is_ai_agent"`
	Content       string         `json:"content"`
	CommitMessage string         `json:"commit_message"`
	Format        DocumentFormat `json:"format,omitempty"`
}

// MergeRevisionsRequest parameters to synchronize diverging revisions.
type MergeRevisionsRequest struct {
	WorkspaceID string       `json:"workspace_id"`
	DocumentID  string       `json:"document_id"`
	BaseRevID   string       `json:"base_rev_id"`
	LocalRevID  string       `json:"local_rev_id"`
	RemoteRevID string       `json:"remote_rev_id"`
	Strategy    SyncStrategy `json:"strategy,omitempty"`
}

// ResolveConflictRequest parameterizes human or AI resolution of individual conflict hunks.
type ResolveConflictRequest struct {
	ResolutionText string `json:"resolution_text"`
	ResolutionNote string `json:"resolution_note,omitempty"`
}

