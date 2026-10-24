package temporalkg

import (
	"errors"
)

// TemporalKGEngine coordinates the ingestion, conflict resolution, time-travel queries, and evolution tracking.
type TemporalKGEngine struct {
	store *TemporalStore
}

// NewEngine initializes a new TemporalKGEngine with an in-memory temporal store.
func NewEngine() *TemporalKGEngine {
	return &TemporalKGEngine{
		store: NewStore(),
	}
}

// RecordFact records a new temporal relation, applying conflict superseding if an active fact exists.
func (e *TemporalKGEngine) RecordFact(req IngestFactRequest) (*TemporalFact, error) {
	if req.WorkspaceID == "" {
		return nil, errors.New("workspace_id is required")
	}
	return e.store.RecordFact(req)
}

// IngestBatch ingests multiple temporal assertions in a single operation.
func (e *TemporalKGEngine) IngestBatch(workspaceID string, facts []IngestFactRequest) ([]*TemporalFact, error) {
	if workspaceID == "" {
		return nil, errors.New("workspace_id is required")
	}

	var recorded []*TemporalFact
	for _, req := range facts {
		req.WorkspaceID = workspaceID
		fact, err := e.store.RecordFact(req)
		if err != nil {
			return nil, err
		}
		recorded = append(recorded, fact)
	}
	return recorded, nil
}

// Query performs point-in-time time-travel or temporal interval filtering with grounded context synthesis.
func (e *TemporalKGEngine) Query(req TemporalQueryRequest) TemporalQueryResponse {
	return ExecuteTemporalQuery(e.store, req)
}

// GetEntity retrieves an entity definition by ID.
func (e *TemporalKGEngine) GetEntity(workspaceID, entityID string) (*TemporalEntity, error) {
	return e.store.GetEntity(workspaceID, entityID)
}

// ListEntities retrieves all entities recorded in the specified workspace.
func (e *TemporalKGEngine) ListEntities(workspaceID string) []*TemporalEntity {
	return e.store.ListEntities(workspaceID)
}

// GetEntityEvolution computes the full chronological mutation history of an entity.
func (e *TemporalKGEngine) GetEntityEvolution(workspaceID, entityID string) (*EntityEvolution, error) {
	return BuildEntityEvolution(e.store, workspaceID, entityID)
}
