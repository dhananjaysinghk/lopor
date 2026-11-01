package docsync

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrRevisionNotFound = errors.New("document revision not found")
	ErrDocumentNotFound = errors.New("document history not found")
	ErrInvalidRequest  = errors.New("invalid request parameters")
)

// DocumentHistory tracks the full acyclic commit DAG of a single document.
type DocumentHistory struct {
	DocumentID  string
	WorkspaceID string
	Revisions   map[string]*DocumentRevision
	HeadRevID   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// DocSyncEngine manages decentralized multi-branch document synchronization and conflict arbitration.
type DocSyncEngine struct {
	mu        sync.RWMutex
	histories map[string]*DocumentHistory // Key: "workspace:docID"
}

// NewEngine initializes a new DocSyncEngine.
func NewEngine() *DocSyncEngine {
	return &DocSyncEngine{
		histories: make(map[string]*DocumentHistory),
	}
}

func (e *DocSyncEngine) getOrCreateHistory(workspaceID, docID string) *DocumentHistory {
	key := fmt.Sprintf("%s:%s", workspaceID, docID)
	h, exists := e.histories[key]
	if !exists {
		now := time.Now().UTC()
		h = &DocumentHistory{
			DocumentID:  docID,
			WorkspaceID: workspaceID,
			Revisions:   make(map[string]*DocumentRevision),
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		e.histories[key] = h
	}
	return h
}

// CommitRevision appends a new revision node to the document DAG.
func (e *DocSyncEngine) CommitRevision(req CommitRevisionRequest) (*DocumentRevision, error) {
	if req.WorkspaceID == "" || req.DocumentID == "" || req.AuthorID == "" {
		return nil, fmt.Errorf("%w: workspace_id, document_id, and author_id are required", ErrInvalidRequest)
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	h := e.getOrCreateHistory(req.WorkspaceID, req.DocumentID)
	parentID := req.ParentRevID
	if parentID == "" {
		parentID = h.HeadRevID
	}

	revID := generateID("rev")
	now := time.Now().UTC()
	hash := HashContent(req.Content)

	msg := req.CommitMessage
	if msg == "" {
		msg = fmt.Sprintf("Revision committed by %s", req.AuthorName)
	}

	rev := &DocumentRevision{
		RevisionID:    revID,
		DocumentID:    req.DocumentID,
		WorkspaceID:   req.WorkspaceID,
		ParentRevID:   parentID,
		AuthorID:      req.AuthorID,
		AuthorName:    req.AuthorName,
		IsAIAgent:     req.IsAIAgent,
		Content:       req.Content,
		ContentHash:   hash,
		CommitMessage: msg,
		Timestamp:     now,
	}

	h.Revisions[revID] = rev
	h.HeadRevID = revID
	h.UpdatedAt = now

	return rev, nil
}

// MergeRevisions executes a 3-way synchronization merge between two diverging branch revisions.
func (e *DocSyncEngine) MergeRevisions(req MergeRevisionsRequest) (*MergeResult, error) {
	if req.WorkspaceID == "" || req.DocumentID == "" || req.LocalRevID == "" || req.RemoteRevID == "" {
		return nil, fmt.Errorf("%w: workspace_id, document_id, local_rev_id, and remote_rev_id are required", ErrInvalidRequest)
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	key := fmt.Sprintf("%s:%s", req.WorkspaceID, req.DocumentID)
	h, exists := e.histories[key]
	if !exists {
		return nil, ErrDocumentNotFound
	}

	localRev, exists := h.Revisions[req.LocalRevID]
	if !exists {
		return nil, fmt.Errorf("%w: local revision %s", ErrRevisionNotFound, req.LocalRevID)
	}

	remoteRev, exists := h.Revisions[req.RemoteRevID]
	if !exists {
		return nil, fmt.Errorf("%w: remote revision %s", ErrRevisionNotFound, req.RemoteRevID)
	}

	baseContent := ""
	if req.BaseRevID != "" {
		if baseRev, exists := h.Revisions[req.BaseRevID]; exists {
			baseContent = baseRev.Content
		}
	} else if localRev.ParentRevID != "" {
		if parentRev, exists := h.Revisions[localRev.ParentRevID]; exists {
			baseContent = parentRev.Content
		}
	}

	strategy := req.Strategy
	if strategy == "" {
		strategy = StrategyThreeWayMerge
	}

	var mergedContent string
	var conflicts []ConflictHunk
	var hasConflicts bool

	switch strategy {
	case StrategyLastWriteWins:
		if localRev.Timestamp.After(remoteRev.Timestamp) {
			mergedContent = localRev.Content
		} else {
			mergedContent = remoteRev.Content
		}

	case StrategySemanticAI:
		merged, rawConflicts, conflictFound := ThreeWayMergeText(baseContent, localRev.Content, remoteRev.Content)
		if conflictFound {
			mergedContent = ArbitrateSemanticMerge(baseContent, localRev.Content, remoteRev.Content, rawConflicts)
			hasConflicts = false
			conflicts = rawConflicts
		} else {
			mergedContent = merged
		}

	case StrategyThreeWayMerge:
		fallthrough
	default:
		mergedContent, conflicts, hasConflicts = ThreeWayMergeText(baseContent, localRev.Content, remoteRev.Content)
	}

	var mergedRevID string
	now := time.Now().UTC()

	// If clean merge (no conflicts), automatically commit the merged revision
	if !hasConflicts {
		mergedRevID = generateID("rev_merge")
		mergedRev := &DocumentRevision{
			RevisionID:    mergedRevID,
			DocumentID:    req.DocumentID,
			WorkspaceID:   req.WorkspaceID,
			ParentRevID:   req.LocalRevID,
			AuthorID:      "system_merger",
			AuthorName:    "Sync Engine",
			IsAIAgent:     true,
			Content:       mergedContent,
			ContentHash:   HashContent(mergedContent),
			CommitMessage: fmt.Sprintf("Merged revision %s and %s via %s", req.LocalRevID, req.RemoteRevID, strategy),
			Timestamp:     now,
		}
		h.Revisions[mergedRevID] = mergedRev
		h.HeadRevID = mergedRevID
		h.UpdatedAt = now
	}

	return &MergeResult{
		DocumentID:     req.DocumentID,
		BaseRevID:      req.BaseRevID,
		LocalRevID:     req.LocalRevID,
		RemoteRevID:    req.RemoteRevID,
		HasConflicts:   hasConflicts,
		ConflictsCount: len(conflicts),
		StrategyUsed:   strategy,
		MergedContent:  mergedContent,
		MergedRevID:    mergedRevID,
		Conflicts:      conflicts,
		MergedAt:       now,
	}, nil
}

// GetRevision fetches a specific historical snapshot.
func (e *DocSyncEngine) GetRevision(workspaceID, docID, revID string) (*DocumentRevision, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	key := fmt.Sprintf("%s:%s", workspaceID, docID)
	h, exists := e.histories[key]
	if !exists {
		return nil, ErrDocumentNotFound
	}

	rev, exists := h.Revisions[revID]
	if !exists {
		return nil, ErrRevisionNotFound
	}
	return rev, nil
}

// ListRevisions returns the complete chronological revision DAG for a document.
func (e *DocSyncEngine) ListRevisions(workspaceID, docID string) ([]*DocumentRevision, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	key := fmt.Sprintf("%s:%s", workspaceID, docID)
	h, exists := e.histories[key]
	if !exists {
		return nil, ErrDocumentNotFound
	}

	var revs []*DocumentRevision
	for _, r := range h.Revisions {
		revs = append(revs, r)
	}
	return revs, nil
}

func generateID(prefix string) string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%s_%s", prefix, hex.EncodeToString(b))
}

