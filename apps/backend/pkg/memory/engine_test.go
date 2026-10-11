package memory

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestMemoryExtractionAndStorage(t *testing.T) {
	engine := NewMemoryEngine()
	ctx := context.Background()

	dialogue := []DialogueTurn{
		{
			Role:      "user",
			Content:   "Hi! I prefer concise code without verbose comments in all answers.",
			Timestamp: time.Now().Add(-10 * time.Minute),
		},
		{
			Role:      "assistant",
			Content:   "Understood, I will keep all code concise.",
			Timestamp: time.Now().Add(-9 * time.Minute),
		},
		{
			Role:      "user",
			Content:   "We use Postgres 16 and Redis 7 for our primary workspace storage.",
			Timestamp: time.Now().Add(-5 * time.Minute),
		},
		{
			Role:      "user",
			Content:   "Never use external CGO dependencies in any backend packages.",
			Timestamp: time.Now().Add(-2 * time.Minute),
		},
		{
			Role:      "user",
			Content:   "My goal is to achieve sub-50ms API latency across all endpoints.",
			Timestamp: time.Now().Add(-1 * time.Minute),
		},
	}

	req := ExtractionRequest{
		WorkspaceID:     "ws-test-1",
		UserID:          "usr-alice",
		DialogueHistory: dialogue,
		CurrentMessage:  "Our internal API is versioned with v1 prefix.",
	}

	result, err := engine.ExtractAndStore(ctx, req)
	if err != nil {
		t.Fatalf("ExtractAndStore failed: %v", err)
	}

	if len(result.ExtractedMemories) == 0 {
		t.Fatalf("expected extracted memories, got 0")
	}

	// Verify categories identified
	foundPreference := false
	foundConstraint := false
	foundTechStack := false
	foundGoal := false
	foundDomainFact := false

	for _, mem := range result.ExtractedMemories {
		switch mem.Category {
		case CategoryPreference:
			foundPreference = true
		case CategoryConstraint:
			foundConstraint = true
		case CategoryTechStack:
			foundTechStack = true
		case CategoryGoal:
			foundGoal = true
		case CategoryDomainFact:
			foundDomainFact = true
		}
	}

	if !foundPreference {
		t.Errorf("expected CategoryPreference to be extracted")
	}
	if !foundConstraint {
		t.Errorf("expected CategoryConstraint to be extracted")
	}
	if !foundTechStack {
		t.Errorf("expected CategoryTechStack to be extracted")
	}
	if !foundGoal {
		t.Errorf("expected CategoryGoal to be extracted")
	}
	if !foundDomainFact {
		t.Errorf("expected CategoryDomainFact to be extracted")
	}

	total := engine.GetStore().Count("ws-test-1", "usr-alice")
	if total < 5 {
		t.Errorf("expected at least 5 active memories in store, got %d", total)
	}
}

func TestMemoryRecallAndPromptInjection(t *testing.T) {
	engine := NewMemoryEngine()
	ctx := context.Background()

	// Seed memories
	wsID := "ws-recall-1"
	uID := "usr-bob"

	engine.GetStore().Save(MemoryItem{
		ID:          "mem-1",
		WorkspaceID: wsID,
		UserID:      uID,
		Category:    CategoryConstraint,
		Content:     "Never use external CGO dependencies in backend modules",
		Keywords:    []string{"external", "cgo", "dependencies", "backend", "modules"},
		Importance:  0.95,
		Confidence:  0.90,
		DecayScore:  0.95,
		IsActive:    true,
	})

	engine.GetStore().Save(MemoryItem{
		ID:          "mem-2",
		WorkspaceID: wsID,
		UserID:      uID,
		Category:    CategoryTechStack,
		Content:     "We use Postgres 16 with pgvector extension",
		Keywords:    []string{"postgres", "pgvector", "extension"},
		Importance:  0.80,
		Confidence:  0.88,
		DecayScore:  0.80,
		IsActive:    true,
	})

	recallReq := RecallRequest{
		WorkspaceID: wsID,
		UserID:      uID,
		Query:       "How should I configure the database pgvector tables?",
		MaxResults:  5,
	}

	recallRes, err := engine.RecallMemories(ctx, recallReq)
	if err != nil {
		t.Fatalf("RecallMemories failed: %v", err)
	}

	if len(recallRes.Items) == 0 {
		t.Fatalf("expected recalled items, got 0")
	}

	// Verify prompt block
	if !strings.Contains(recallRes.FormattedPromptBlock, "<user_long_term_memory>") {
		t.Errorf("expected prompt block to contain <user_long_term_memory> tag")
	}
	if !strings.Contains(recallRes.FormattedPromptBlock, "pgvector") {
		t.Errorf("expected prompt block to mention recalled content 'pgvector'")
	}
}

func TestMemoryConsolidation(t *testing.T) {
	engine := NewMemoryEngine()
	ctx := context.Background()

	wsID := "ws-cons-1"
	uID := "usr-charlie"

	// Insert duplicate preference items
	engine.GetStore().Save(MemoryItem{
		ID:          "dup-1",
		WorkspaceID: wsID,
		UserID:      uID,
		Category:    CategoryPreference,
		Content:     "I prefer concise Go code with clean architecture",
		Keywords:    []string{"concise", "code", "clean", "architecture"},
		Importance:  0.78,
		CreatedAt:   time.Now().Add(-2 * time.Hour),
		IsActive:    true,
	})

	engine.GetStore().Save(MemoryItem{
		ID:          "dup-2",
		WorkspaceID: wsID,
		UserID:      uID,
		Category:    CategoryPreference,
		Content:     "I prefer concise code with clean architecture design",
		Keywords:    []string{"concise", "code", "clean", "architecture", "design"},
		Importance:  0.80,
		CreatedAt:   time.Now().Add(-1 * time.Hour),
		IsActive:    true,
	})

	report, err := engine.Consolidate(ctx, wsID, uID)
	if err != nil {
		t.Fatalf("Consolidate failed: %v", err)
	}

	if report.MergedCount == 0 {
		t.Errorf("expected near-duplicate memories to be merged, got 0 merged")
	}

	remaining := engine.GetStore().Count(wsID, uID)
	if remaining != 1 {
		t.Errorf("expected exactly 1 consolidated memory remaining, got %d", remaining)
	}
}

func TestMemoryGDPRPurge(t *testing.T) {
	engine := NewMemoryEngine()
	wsID := "ws-gdpr-1"
	uID := "usr-david"

	engine.GetStore().Save(MemoryItem{
		ID:          "gdpr-mem-1",
		WorkspaceID: wsID,
		UserID:      uID,
		Category:    CategoryPreference,
		Content:     "I like light mode themes",
		Keywords:    []string{"light", "mode", "themes"},
		Importance:  0.5,
		IsActive:    true,
	})

	countBefore := engine.GetStore().Count(wsID, uID)
	if countBefore != 1 {
		t.Fatalf("expected 1 item, got %d", countBefore)
	}

	purged := engine.GetStore().PurgeUser(wsID, uID)
	if purged != 1 {
		t.Errorf("expected 1 purged item, got %d", purged)
	}

	countAfter := engine.GetStore().Count(wsID, uID)
	if countAfter != 0 {
		t.Errorf("expected 0 items after purge, got %d", countAfter)
	}
}
