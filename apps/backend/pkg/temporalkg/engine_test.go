package temporalkg

import (
	"strings"
	"testing"
	"time"
)

func TestTemporalSupersedingAndPointInTimeQuery(t *testing.T) {
	engine := NewEngine()
	wsID := "ws_alpha"

	// 1. Fact 1: Bob was CTO from 2021-01-01 to 2023-12-31
	t2021 := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	t2024 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	_, err := engine.RecordFact(IngestFactRequest{
		WorkspaceID:    wsID,
		SubjectName:    "Acme Corp",
		SubjectType:    EntityOrganization,
		Predicate:      "HAS_CTO",
		ObjectName:     "Bob Vance",
		ObjectType:     EntityPerson,
		ValidFrom:      t2021,
		SourceCitation: "SEC 10-K 2021 filing",
	})
	if err != nil {
		t.Fatalf("failed to record initial fact: %v", err)
	}

	// 2. Fact 2: Alice appointed CTO on 2024-01-01 (should supersede Bob)
	_, err = engine.RecordFact(IngestFactRequest{
		WorkspaceID:    wsID,
		SubjectName:    "Acme Corp",
		SubjectType:    EntityOrganization,
		Predicate:      "HAS_CTO",
		ObjectName:     "Alice Smith",
		ObjectType:     EntityPerson,
		ValidFrom:      t2024,
		SourceCitation: "Press Release 2024-01-01",
	})
	if err != nil {
		t.Fatalf("failed to record superseding fact: %v", err)
	}

	// 3. Query AsOf 2022 (Should return Bob Vance)
	asOf2022 := time.Date(2022, 6, 1, 0, 0, 0, 0, time.UTC)
	res2022 := engine.Query(TemporalQueryRequest{
		WorkspaceID: wsID,
		Predicate:   "HAS_CTO",
		AsOf:        &asOf2022,
	})

	if res2022.TotalMatches != 1 {
		t.Fatalf("expected 1 match for AsOf 2022, got %d", res2022.TotalMatches)
	}
	if res2022.MatchedFacts[0].ObjectName != "Bob Vance" {
		t.Errorf("expected CTO in 2022 to be Bob Vance, got %s", res2022.MatchedFacts[0].ObjectName)
	}

	// 4. Query AsOf Current / 2024-06 (Should return Alice Smith)
	asOf2024 := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	res2024 := engine.Query(TemporalQueryRequest{
		WorkspaceID: wsID,
		Predicate:   "HAS_CTO",
		AsOf:        &asOf2024,
	})

	if res2024.TotalMatches != 1 {
		t.Fatalf("expected 1 match for AsOf 2024, got %d", res2024.TotalMatches)
	}
	if res2024.MatchedFacts[0].ObjectName != "Alice Smith" {
		t.Errorf("expected CTO in 2024 to be Alice Smith, got %s", res2024.MatchedFacts[0].ObjectName)
	}

	// 5. Query Only Current
	resCurrent := engine.Query(TemporalQueryRequest{
		WorkspaceID: wsID,
		Predicate:   "HAS_CTO",
		OnlyCurrent: true,
	})
	if resCurrent.TotalMatches != 1 || resCurrent.MatchedFacts[0].ObjectName != "Alice Smith" {
		t.Errorf("expected only Alice Smith to be current CTO")
	}
}

func TestEntityEvolutionAndContextSynthesis(t *testing.T) {
	engine := NewEngine()
	wsID := "ws_alpha"

	t2020 := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	t2023 := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

	_, _ = engine.RecordFact(IngestFactRequest{
		WorkspaceID: wsID,
		SubjectName: "Project Titan",
		SubjectType: EntityProject,
		Predicate:   "MANAGED_BY",
		ObjectName:  "Lead Engineer Dave",
		ObjectType:  EntityPerson,
		ValidFrom:   t2020,
	})

	_, _ = engine.RecordFact(IngestFactRequest{
		WorkspaceID: wsID,
		SubjectName: "Project Titan",
		SubjectType: EntityProject,
		Predicate:   "MANAGED_BY",
		ObjectName:  "Director Sarah",
		ObjectType:  EntityPerson,
		ValidFrom:   t2023,
	})

	entities := engine.ListEntities(wsID)
	var titanID string
	for _, e := range entities {
		if e.CanonicalName == "Project Titan" {
			titanID = e.ID
			break
		}
	}

	if titanID == "" {
		t.Fatalf("Project Titan entity not found")
	}

	evolution, err := engine.GetEntityEvolution(wsID, titanID)
	if err != nil {
		t.Fatalf("failed to get entity evolution: %v", err)
	}

	if len(evolution.Events) < 2 {
		t.Errorf("expected at least 2 mutation events, got %d", len(evolution.Events))
	}

	// Test Grounded Context Synthesis
	queryRes := engine.Query(TemporalQueryRequest{
		WorkspaceID: wsID,
		EntityID:    titanID,
	})

	if !strings.Contains(queryRes.GroundedContext, "TEMPORAL KNOWLEDGE GRAPH EVIDENCE") {
		t.Errorf("grounded context missing header: %s", queryRes.GroundedContext)
	}
	if !strings.Contains(queryRes.GroundedContext, "Director Sarah") {
		t.Errorf("grounded context missing active fact: %s", queryRes.GroundedContext)
	}
}

func TestWorkspaceIsolation(t *testing.T) {
	engine := NewEngine()

	_, _ = engine.RecordFact(IngestFactRequest{
		WorkspaceID: "ws_alpha",
		SubjectName: "Secret Org",
		SubjectType: EntityOrganization,
		Predicate:   "LOCATED_IN",
		ObjectName:  "Geneva",
		ObjectType:  EntityOrganization,
	})

	// ws_beta should see 0 facts and 0 entities
	res := engine.Query(TemporalQueryRequest{
		WorkspaceID: "ws_beta",
	})
	if res.TotalMatches != 0 {
		t.Errorf("expected 0 matches for isolated workspace, got %d", res.TotalMatches)
	}

	entities := engine.ListEntities("ws_beta")
	if len(entities) != 0 {
		t.Errorf("expected 0 entities for isolated workspace, got %d", len(entities))
	}
}
