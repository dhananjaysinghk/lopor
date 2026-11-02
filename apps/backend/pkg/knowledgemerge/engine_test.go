package knowledgemerge

import (
	"testing"
	"time"
)

func TestEntityDeduplicationAndAliasMerging(t *testing.T) {
	engine := NewEngine()
	wsID := "ws_alpha"

	// Initial entity: International Business Machines
	_, err := engine.IngestAndReconcile(IngestKnowledgeGraphRequest{
		WorkspaceID: wsID,
		Entities: []KnowledgeEntity{
			{
				EntityID:      "ent_ibm",
				CanonicalName: "International Business Machines",
				Aliases:       []string{"IBM Corp"},
				Type:          "ORGANIZATION",
				Attributes:    map[string]string{"ticker": "IBM"},
			},
		},
	})
	if err != nil {
		t.Fatalf("failed initial ingestion: %v", err)
	}

	// Incoming entity with alias: IBM Corp
	res, err := engine.IngestAndReconcile(IngestKnowledgeGraphRequest{
		WorkspaceID: wsID,
		Entities: []KnowledgeEntity{
			{
				EntityID:      "ent_ibm_alias",
				CanonicalName: "IBM Corp",
				Aliases:       []string{"Big Blue"},
				Type:          "ORGANIZATION",
				Attributes:    map[string]string{"industry": "Technology"},
			},
		},
	})
	if err != nil {
		t.Fatalf("failed alias ingestion: %v", err)
	}

	if res.MergedEntities != 1 {
		t.Errorf("expected 1 merged entity, got %d", res.MergedEntities)
	}

	graph, _ := engine.GetGraph(wsID)
	if len(graph.Entities) != 1 {
		t.Errorf("expected 1 unified canonical entity, got %d", len(graph.Entities))
	}

	unified := graph.Entities["ent_ibm"]
	if unified.Attributes["ticker"] != "IBM" || unified.Attributes["industry"] != "Technology" {
		t.Errorf("expected merged attributes, got: %v", unified.Attributes)
	}
}

func TestContradictionResolutionAuthorityHierarchy(t *testing.T) {
	engine := NewEngine()
	wsID := "ws_alpha"

	// Fact 1: AI Agent extracted that CEO is John Doe (Confidence 0.90)
	// Fact 2: Human Expert asserted that CEO is Jane Smith (Confidence 0.85)
	facts := []KnowledgeFact{
		{
			SubjectName: "Acme Corp",
			Predicate:   "HAS_CEO",
			ObjectName:  "John Doe",
			Confidence:  0.90,
			SourceType:  "AI_AGENT",
			SourceID:    "agent_extractor_v1",
		},
		{
			SubjectName: "Acme Corp",
			Predicate:   "HAS_CEO",
			ObjectName:  "Jane Smith",
			Confidence:  0.85,
			SourceType:  "HUMAN_EXPERT",
			SourceID:    "admin_audit_officer",
		},
	}

	res, err := engine.IngestAndReconcile(IngestKnowledgeGraphRequest{
		WorkspaceID:     wsID,
		Facts:           facts,
		DefaultStrategy: StrategyAuthorityHierarchy,
	})
	if err != nil {
		t.Fatalf("unexpected error during reconciliation: %v", err)
	}

	if res.ConflictsDetected != 1 {
		t.Errorf("expected 1 conflict detected, got %d", res.ConflictsDetected)
	}
	if res.ConflictsResolved != 1 {
		t.Errorf("expected 1 conflict resolved, got %d", res.ConflictsResolved)
	}

	conflict := res.Conflicts[0]
	if conflict.ResolvedFact.ObjectName != "Jane Smith" {
		t.Errorf("expected Jane Smith to win under Authority Hierarchy, got %s", conflict.ResolvedFact.ObjectName)
	}
}

func TestContradictionResolutionRecencyWeighted(t *testing.T) {
	engine := NewEngine()
	wsID := "ws_alpha"

	t2022 := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
	t2024 := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)

	facts := []KnowledgeFact{
		{
			SubjectName: "CloudCluster",
			Predicate:   "DEPLOYED_REGION",
			ObjectName:  "us-east-1",
			ValidFrom:   t2022,
			SourceID:    "legacy_doc",
		},
		{
			SubjectName: "CloudCluster",
			Predicate:   "DEPLOYED_REGION",
			ObjectName:  "eu-west-1",
			ValidFrom:   t2024,
			SourceID:    "latest_terraform",
		},
	}

	res, err := engine.IngestAndReconcile(IngestKnowledgeGraphRequest{
		WorkspaceID:     wsID,
		Facts:           facts,
		DefaultStrategy: StrategyRecencyWeighted,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	conflict := res.Conflicts[0]
	if conflict.ResolvedFact.ObjectName != "eu-west-1" {
		t.Errorf("expected eu-west-1 to win under Recency Weighted, got %s", conflict.ResolvedFact.ObjectName)
	}
}

func TestWorkspaceTenancyIsolation(t *testing.T) {
	engine := NewEngine()

	_, _ = engine.IngestAndReconcile(IngestKnowledgeGraphRequest{
		WorkspaceID: "ws_alpha",
		Entities: []KnowledgeEntity{
			{CanonicalName: "Secret Project Alpha"},
		},
	})

	// ws_beta should not see ws_alpha graph
	_, err := engine.GetGraph("ws_beta")
	if err != ErrWorkspaceNotFound {
		t.Errorf("expected ErrWorkspaceNotFound for non-existent workspace graph, got %v", err)
	}
}

