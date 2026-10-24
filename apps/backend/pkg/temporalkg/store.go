package temporalkg

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
	ErrEntityNotFound = errors.New("temporal entity not found")
	ErrFactNotFound   = errors.New("temporal fact not found")
)

// TemporalStore manages the thread-safe persistence and indexing of entities and temporal facts.
type TemporalStore struct {
	mu          sync.RWMutex
	entities    map[string]*TemporalEntity // keyed by "workspace:canonical_name"
	entityByID  map[string]*TemporalEntity // keyed by entity ID
	facts       map[string]*TemporalFact   // keyed by fact ID
	entityFacts map[string][]string        // entityID -> []factID
}

// NewStore initializes a new TemporalStore.
func NewStore() *TemporalStore {
	return &TemporalStore{
		entities:    make(map[string]*TemporalEntity),
		entityByID:  make(map[string]*TemporalEntity),
		facts:       make(map[string]*TemporalFact),
		entityFacts: make(map[string][]string),
	}
}

// UpsertEntity retrieves or creates a temporal entity within a given workspace.
func (s *TemporalStore) UpsertEntity(workspaceID, name string, entityType EntityType) *TemporalEntity {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := fmt.Sprintf("%s:%s", workspaceID, strings.ToLower(strings.TrimSpace(name)))
	if entity, exists := s.entities[key]; exists {
		return entity
	}

	entityID := generateID("ent")
	entity := &TemporalEntity{
		ID:            entityID,
		WorkspaceID:   workspaceID,
		CanonicalName: strings.TrimSpace(name),
		Type:          entityType,
		Attributes:    make(map[string]string),
		CreatedAt:     time.Now().UTC(),
	}

	s.entities[key] = entity
	s.entityByID[entityID] = entity
	return entity
}

// RecordFact stores a temporal assertion, superseding conflicting current facts for the subject.
func (s *TemporalStore) RecordFact(req IngestFactRequest) (*TemporalFact, error) {
	if req.WorkspaceID == "" || req.SubjectName == "" || req.Predicate == "" || req.ObjectName == "" {
		return nil, errors.New("workspace_id, subject_name, predicate, and object_name are required")
	}

	subj := s.UpsertEntity(req.WorkspaceID, req.SubjectName, req.SubjectType)
	obj := s.UpsertEntity(req.WorkspaceID, req.ObjectName, req.ObjectType)

	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	validFrom := req.ValidFrom
	if validFrom.IsZero() {
		validFrom = now
	}

	// 1. Conflict Resolution & Superseding:
	// If an existing fact for the same Subject and Predicate is active and current, supersede it
	existingFactIDs := s.entityFacts[subj.ID]
	for _, fID := range existingFactIDs {
		f := s.facts[fID]
		if f != nil && f.Predicate == req.Predicate && f.IsCurrent {
			// Supersede existing fact
			f.IsCurrent = false
			f.SupersededAt = &now
			if f.ValidTo == nil || f.ValidTo.After(validFrom) {
				f.ValidTo = &validFrom
			}
		}
	}

	// 2. Insert new fact
	isCurrent := req.ValidTo == nil || req.ValidTo.After(now)
	confidence := req.Confidence
	if confidence <= 0.0 {
		confidence = 0.95
	}

	factID := generateID("fct")
	newFact := &TemporalFact{
		ID:             factID,
		WorkspaceID:    req.WorkspaceID,
		SubjectID:      subj.ID,
		SubjectName:    subj.CanonicalName,
		Predicate:      req.Predicate,
		ObjectID:       obj.ID,
		ObjectName:     obj.CanonicalName,
		ValidFrom:      validFrom,
		ValidTo:        req.ValidTo,
		IsCurrent:      isCurrent,
		Confidence:     confidence,
		SourceCitation: req.SourceCitation,
		RecordedAt:     now,
	}

	s.facts[factID] = newFact
	s.entityFacts[subj.ID] = append(s.entityFacts[subj.ID], factID)
	s.entityFacts[obj.ID] = append(s.entityFacts[obj.ID], factID)

	return newFact, nil
}

// GetEntity retrieves an entity by its ID and workspace.
func (s *TemporalStore) GetEntity(workspaceID, entityID string) (*TemporalEntity, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entity, exists := s.entityByID[entityID]
	if !exists || entity.WorkspaceID != workspaceID {
		return nil, ErrEntityNotFound
	}
	return entity, nil
}

// ListEntities returns all entities registered within the workspace.
func (s *TemporalStore) ListEntities(workspaceID string) []*TemporalEntity {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*TemporalEntity
	for _, e := range s.entityByID {
		if e.WorkspaceID == workspaceID {
			result = append(result, e)
		}
	}
	return result
}

// GetAllFactsForWorkspace returns all facts recorded for a workspace.
func (s *TemporalStore) GetAllFactsForWorkspace(workspaceID string) []*TemporalFact {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*TemporalFact
	for _, f := range s.facts {
		if f.WorkspaceID == workspaceID {
			result = append(result, f)
		}
	}
	return result
}

func generateID(prefix string) string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%s_%s", prefix, hex.EncodeToString(b))
}
