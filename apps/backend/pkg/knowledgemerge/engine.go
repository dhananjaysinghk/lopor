package knowledgemerge

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

var (
	ErrWorkspaceNotFound = errors.New("workspace graph not found")
	ErrInvalidRequest   = errors.New("invalid reconciliation request parameters")
)

// WorkspaceKnowledgeGraph maintains unified, deduplicated entities and reconciled facts.
type WorkspaceKnowledgeGraph struct {
	WorkspaceID string
	Entities    map[string]*KnowledgeEntity // Key: EntityID
	Facts       map[string]*KnowledgeFact   // Key: FactID
	Conflicts   []FactConflict
	UpdatedAt   time.Time
}

// KnowledgeMergeEngine coordinates cross-source graph ingestion, deduplication, and resolution.
type KnowledgeMergeEngine struct {
	mu     sync.RWMutex
	graphs map[string]*WorkspaceKnowledgeGraph
}

// NewEngine initializes a new KnowledgeMergeEngine.
func NewEngine() *KnowledgeMergeEngine {
	return &KnowledgeMergeEngine{
		graphs: make(map[string]*WorkspaceKnowledgeGraph),
	}
}

func (e *KnowledgeMergeEngine) getOrCreateGraph(workspaceID string) *WorkspaceKnowledgeGraph {
	g, exists := e.graphs[workspaceID]
	if !exists {
		g = &WorkspaceKnowledgeGraph{
			WorkspaceID: workspaceID,
			Entities:    make(map[string]*KnowledgeEntity),
			Facts:       make(map[string]*KnowledgeFact),
			Conflicts:   make([]FactConflict, 0),
			UpdatedAt:   time.Now().UTC(),
		}
		e.graphs[workspaceID] = g
	}
	return g
}

// IngestAndReconcile processes incoming graph entities and relational facts, unifying canonical state.
func (e *KnowledgeMergeEngine) IngestAndReconcile(req IngestKnowledgeGraphRequest) (*ReconcileResult, error) {
	if req.WorkspaceID == "" {
		return nil, fmt.Errorf("%w: workspace_id is required", ErrInvalidRequest)
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	g := e.getOrCreateGraph(req.WorkspaceID)
	strategy := req.DefaultStrategy
	if strategy == "" {
		strategy = StrategyRecencyWeighted
	}

	mergedEntitiesCount := 0
	now := time.Now().UTC()

	// 1. Entity Disambiguation and Deduplication
	entityIDMap := make(map[string]string) // incoming EntityID/Name -> canonical EntityID

	for _, incoming := range req.Entities {
		existing := FindMatchingEntity(g.Entities, incoming.CanonicalName, 0.75)
		if existing != nil {
			// Merge attributes & aliases
			MergeEntityAttributes(existing, incoming)
			existing.UpdatedAt = now
			entityIDMap[incoming.EntityID] = existing.EntityID
			entityIDMap[incoming.CanonicalName] = existing.EntityID
			mergedEntitiesCount++
		} else {
			newID := incoming.EntityID
			if newID == "" {
				newID = generateID("ent")
			}
			ent := &KnowledgeEntity{
				EntityID:      newID,
				WorkspaceID:   req.WorkspaceID,
				CanonicalName: incoming.CanonicalName,
				Aliases:       incoming.Aliases,
				Type:          incoming.Type,
				Attributes:    incoming.Attributes,
				CreatedAt:     now,
				UpdatedAt:     now,
			}
			g.Entities[newID] = ent
			entityIDMap[incoming.EntityID] = newID
			entityIDMap[incoming.CanonicalName] = newID
		}
	}

	// 2. Fact Grouping by Subject and Predicate
	type factKey struct {
		Subject string
		Pred    string
	}
	factGroups := make(map[factKey][]KnowledgeFact)

	for _, f := range req.Facts {
		// Map subject to canonical ID/name if recognized
		subKey := f.SubjectName
		if canonID, ok := entityIDMap[f.SubjectName]; ok {
			if ent, ok := g.Entities[canonID]; ok {
				f.SubjectID = ent.EntityID
				f.SubjectName = ent.CanonicalName
				subKey = ent.CanonicalName
			}
		}

		// Map object to canonical ID/name if recognized
		if canonID, ok := entityIDMap[f.ObjectName]; ok {
			if ent, ok := g.Entities[canonID]; ok {
				f.ObjectID = ent.EntityID
				f.ObjectName = ent.CanonicalName
			}
		}

		if f.FactID == "" {
			f.FactID = generateID("fct")
		}
		if f.RecordedAt.IsZero() {
			f.RecordedAt = now
		}
		if f.ValidFrom.IsZero() {
			f.ValidFrom = now
		}

		key := factKey{
			Subject: strings.ToLower(subKey),
			Pred:    strings.ToLower(f.Predicate),
		}
		factGroups[key] = append(factGroups[key], f)
	}

	// Include existing active facts into grouping to test cross-source contradiction
	for _, f := range g.Facts {
		key := factKey{
			Subject: strings.ToLower(f.SubjectName),
			Pred:    strings.ToLower(f.Predicate),
		}
		// If incoming facts touched this key, add existing fact to conflict check
		if _, exists := factGroups[key]; exists {
			factGroups[key] = append(factGroups[key], *f)
		}
	}

	// 3. Conflict Detection and Arbitration
	conflictsDetected := 0
	conflictsResolved := 0
	var newConflicts []FactConflict

	for _, group := range factGroups {
		// Deduplicate identical object assertions
		uniqueObjects := make(map[string]bool)
		for _, f := range group {
			uniqueObjects[strings.ToLower(f.ObjectName)] = true
		}

		if len(uniqueObjects) > 1 {
			// Contradiction detected!
			conflictsDetected++
			resolvedFact, explanation := ResolveContradiction(group, strategy)
			conflictsResolved++

			conflictID := generateID("cnf")
			conflict := FactConflict{
				ConflictID:       conflictID,
				SubjectName:      group[0].SubjectName,
				Predicate:        group[0].Predicate,
				ConflictingFacts: group,
				ResolvedFact:     resolvedFact,
				StrategyUsed:     strategy,
				Explanation:      explanation,
				IsResolved:       true,
			}
			newConflicts = append(newConflicts, conflict)

			// Store resolved fact into canonical graph
			g.Facts[resolvedFact.FactID] = resolvedFact
		} else {
			// No conflict, store latest fact
			latest := group[0]
			g.Facts[latest.FactID] = &latest
		}
	}

	g.Conflicts = append(g.Conflicts, newConflicts...)
	g.UpdatedAt = now

	return &ReconcileResult{
		WorkspaceID:       req.WorkspaceID,
		TotalEntities:     len(g.Entities),
		TotalFacts:        len(g.Facts),
		MergedEntities:    mergedEntitiesCount,
		ConflictsDetected: conflictsDetected,
		ConflictsResolved: conflictsResolved,
		Conflicts:         newConflicts,
		ReconciledAt:      now,
	}, nil
}

// GetGraph retrieves the canonical entities and facts for a workspace.
func (e *KnowledgeMergeEngine) GetGraph(workspaceID string) (*WorkspaceKnowledgeGraph, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	g, exists := e.graphs[workspaceID]
	if !exists {
		return nil, ErrWorkspaceNotFound
	}
	return g, nil
}

// ListConflicts returns all detected factual conflicts in the workspace.
func (e *KnowledgeMergeEngine) ListConflicts(workspaceID string) ([]FactConflict, error) {
	g, err := e.GetGraph(workspaceID)
	if err != nil {
		return nil, err
	}
	return g.Conflicts, nil
}

func generateID(prefix string) string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%s_%s", prefix, hex.EncodeToString(b))
}
