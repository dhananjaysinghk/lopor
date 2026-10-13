package graphrag

import (
	"time"
)

// EntityType categorizes knowledge graph nodes.
type EntityType string

const (
	EntityTechnology   EntityType = "TECHNOLOGY"
	EntityOrganization EntityType = "ORGANIZATION"
	EntityConcept      EntityType = "CONCEPT"
	EntityPerson       EntityType = "PERSON"
	EntityDocument     EntityType = "DOCUMENT"
	EntityComponent    EntityType = "COMPONENT"
)

// EntityNode represents a vertex in the knowledge graph.
type EntityNode struct {
	ID          string            `json:"id"`
	WorkspaceID string            `json:"workspace_id"`
	Name        string            `json:"name"`
	Type        EntityType        `json:"type"`
	Description string            `json:"description"`
	DocumentID  string            `json:"document_id,omitempty"`
	Degree      int               `json:"degree"`
	CommunityID string            `json:"community_id,omitempty"`
	Properties  map[string]string `json:"properties,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
}

// RelationshipEdge represents a directed relationship between two entities.
type RelationshipEdge struct {
	ID           string    `json:"id"`
	WorkspaceID  string    `json:"workspace_id"`
	SourceID     string    `json:"source_id"`
	SourceName   string    `json:"source_name"`
	TargetID     string    `json:"target_id"`
	TargetName   string    `json:"target_name"`
	RelationType string    `json:"relation_type"` // e.g. "DEPENDS_ON", "IMPLEMENTS", "INTEGRATES_WITH"
	Weight       float64   `json:"weight"`
	Description  string    `json:"description"`
	CreatedAt    time.Time `json:"created_at"`
}

// CommunityCluster represents a partition of densely connected knowledge entities.
type CommunityCluster struct {
	ID            string    `json:"id"`
	WorkspaceID   string    `json:"workspace_id"`
	Level         int       `json:"level"`
	Name          string    `json:"name"`
	Summary       string    `json:"summary"`
	MemberNodeIDs []string  `json:"member_node_ids"`
	KeyThemes     []string  `json:"key_themes"`
	Density       float64   `json:"density"`
	CreatedAt     time.Time `json:"created_at"`
}

// GraphExtractionRequest provides document text to extract knowledge triplets from.
type GraphExtractionRequest struct {
	WorkspaceID string `json:"workspace_id"`
	DocumentID  string `json:"document_id,omitempty"`
	Content     string `json:"content"`
}

// GraphExtractionResult provides newly extracted entities and relations.
type GraphExtractionResult struct {
	ExtractedNodes []EntityNode       `json:"extracted_nodes"`
	ExtractedEdges []RelationshipEdge `json:"extracted_edges"`
	ExtractedCount int                `json:"extracted_count"`
}

// LocalGraphSearchRequest queries an entity's immediate relationship neighborhood.
type LocalGraphSearchRequest struct {
	WorkspaceID string   `json:"workspace_id"`
	Query       string   `json:"query"`
	EntitySeeds []string `json:"entity_seeds,omitempty"`
	MaxDepth    int      `json:"max_depth,omitempty"` // default 1 or 2
	Limit       int      `json:"limit,omitempty"`
}

// LocalGraphSearchResult provides entity context and localized graph traversal paths.
type LocalGraphSearchResult struct {
	Query           string             `json:"query"`
	FocusedNodes    []EntityNode       `json:"focused_nodes"`
	ConnectingEdges []RelationshipEdge `json:"connecting_edges"`
	ContextSummary  string             `json:"context_summary"`
	Confidence      float64            `json:"confidence"`
}

// GlobalGraphSearchRequest asks a broad question answered by community summaries.
type GlobalGraphSearchRequest struct {
	WorkspaceID    string `json:"workspace_id"`
	Query          string `json:"query"`
	CommunityLevel int    `json:"community_level,omitempty"`
}

// GlobalGraphSearchResult synthesizes answers across modular community summaries.
type GlobalGraphSearchResult struct {
	Query                 string             `json:"query"`
	SynthesizedAnswer     string             `json:"synthesized_answer"`
	ReferencedCommunities []CommunityCluster `json:"referenced_communities"`
	RelevantThemes        []string           `json:"relevant_themes"`
	Confidence            float64            `json:"confidence"`
}

// VisualNode is formatted for frontend D3 / Cytoscape force-directed graphs.
type VisualNode struct {
	ID          string     `json:"id"`
	Label       string     `json:"label"`
	Type        EntityType `json:"type"`
	Degree      int        `json:"degree"`
	CommunityID string     `json:"community_id"`
}

// VisualLink formats an edge for UI graph renderers.
type VisualLink struct {
	ID     string  `json:"id"`
	Source string  `json:"source"`
	Target string  `json:"target"`
	Label  string  `json:"label"`
	Weight float64 `json:"weight"`
}

// GraphVisualizationData packages nodes, edges, and clusters for visualization.
type GraphVisualizationData struct {
	WorkspaceID string             `json:"workspace_id"`
	Nodes       []VisualNode       `json:"nodes"`
	Links       []VisualLink       `json:"links"`
	Communities []CommunityCluster `json:"communities"`
	TotalNodes  int                `json:"total_nodes"`
	TotalEdges  int                `json:"total_edges"`
}
