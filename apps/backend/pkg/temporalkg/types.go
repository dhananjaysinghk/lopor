package temporalkg

import (
	"time"
)

// EntityType categorizes an entity within the enterprise domain.
type EntityType string

const (
	EntityOrganization   EntityType = "ORGANIZATION"
	EntityPerson         EntityType = "PERSON"
	EntityProject        EntityType = "PROJECT"
	EntityPolicy         EntityType = "POLICY"
	EntityInfrastructure EntityType = "INFRASTRUCTURE"
)

// MutationType represents a state change in the temporal lifecycle of a relation or entity.
type MutationType string

const (
	MutationCreated    MutationType = "CREATED"
	MutationUpdated    MutationType = "UPDATED"
	MutationSuperseded MutationType = "SUPERSEDED"
	MutationRevoked    MutationType = "REVOKED"
)

// TemporalEntity models a unique enterprise entity whose relationships evolve over time.
type TemporalEntity struct {
	ID            string            `json:"id"`
	WorkspaceID   string            `json:"workspace_id"`
	CanonicalName string            `json:"canonical_name"`
	Type          EntityType        `json:"type"`
	Attributes    map[string]string `json:"attributes,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
}

// TemporalFact represents a bi-temporal relational assertion (Subject -[Predicate]-> Object).
type TemporalFact struct {
	ID             string     `json:"id"`
	WorkspaceID    string     `json:"workspace_id"`
	SubjectID      string     `json:"subject_id"`
	SubjectName    string     `json:"subject_name"`
	Predicate      string     `json:"predicate"`
	ObjectID       string     `json:"object_id"`
	ObjectName     string     `json:"object_name"`
	ValidFrom      time.Time  `json:"valid_from"`
	ValidTo        *time.Time `json:"valid_to,omitempty"` // Nil indicates fact is currently active
	IsCurrent      bool       `json:"is_current"`
	Confidence     float64    `json:"confidence"`
	SourceCitation string     `json:"source_citation,omitempty"`
	RecordedAt     time.Time  `json:"recorded_at"`
	SupersededAt   *time.Time `json:"superseded_at,omitempty"`
}

// FactEvent records a timestamped mutation event along an entity's historical trajectory.
type FactEvent struct {
	Timestamp      time.Time    `json:"timestamp"`
	Mutation       MutationType `json:"mutation"`
	Predicate      string       `json:"predicate"`
	TargetEntity   string       `json:"target_entity"`
	ValidInterval  string       `json:"valid_interval"`
	SourceCitation string       `json:"source_citation,omitempty"`
}

// EntityEvolution encapsulates the chronological lifecycle and relationship changes of an entity.
type EntityEvolution struct {
	EntityID      string      `json:"entity_id"`
	CanonicalName string      `json:"canonical_name"`
	Type          EntityType  `json:"type"`
	Events        []FactEvent `json:"events"`
	CurrentFacts  []TemporalFact `json:"current_facts"`
}

// TemporalQueryRequest specifies parameters for historical time-travel or interval retrieval.
type TemporalQueryRequest struct {
	WorkspaceID string     `json:"workspace_id"`
	EntityID    string     `json:"entity_id,omitempty"`
	Predicate   string     `json:"predicate,omitempty"`
	AsOf        *time.Time `json:"as_of,omitempty"`      // Point-in-time snapshot
	FromTime    *time.Time `json:"from_time,omitempty"`  // Range query start
	ToTime      *time.Time `json:"to_time,omitempty"`    // Range query end
	OnlyCurrent bool       `json:"only_current,omitempty"`
}

// TemporalQueryResponse contains the filtered facts and contextual synthesis for RAG grounding.
type TemporalQueryResponse struct {
	WorkspaceID     string         `json:"workspace_id"`
	MatchedFacts    []TemporalFact `json:"matched_facts"`
	AsOfTimestamp   *time.Time     `json:"as_of_timestamp,omitempty"`
	GroundedContext string         `json:"grounded_context"`
	TotalMatches    int            `json:"total_matches"`
}

// IngestFactRequest specifies parameters to record a new or updated temporal assertion.
type IngestFactRequest struct {
	WorkspaceID    string    `json:"workspace_id"`
	SubjectName    string    `json:"subject_name"`
	SubjectType    EntityType `json:"subject_type"`
	Predicate      string    `json:"predicate"`
	ObjectName     string    `json:"object_name"`
	ObjectType     EntityType `json:"object_type"`
	ValidFrom      time.Time `json:"valid_from"`
	ValidTo        *time.Time `json:"valid_to,omitempty"`
	SourceCitation string    `json:"source_citation,omitempty"`
	Confidence     float64   `json:"confidence,omitempty"`
}
