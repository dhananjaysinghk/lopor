package docsync

import (
	"strings"
	"testing"
)

func TestCleanThreeWayMerge(t *testing.T) {
	engine := NewEngine()
	wsID := "ws_alpha"
	docID := "doc_spec_1"

	// 1. Commit Base Revision
	baseContent := "Line 1: System Overview\nLine 2: Architecture Details\nLine 3: Appendix"
	baseRev, err := engine.CommitRevision(CommitRevisionRequest{
		WorkspaceID:   wsID,
		DocumentID:    docID,
		AuthorID:      "human_eng_1",
		AuthorName:    "Alice",
		Content:       baseContent,
		CommitMessage: "Initial spec baseline",
	})
	if err != nil {
		t.Fatalf("failed to commit base revision: %v", err)
	}

	// 2. Commit Local Branch (Alice modifies Line 1)
	localContent := "Line 1: System Overview (Updated)\nLine 2: Architecture Details\nLine 3: Appendix"
	localRev, err := engine.CommitRevision(CommitRevisionRequest{
		WorkspaceID:   wsID,
		DocumentID:    docID,
		ParentRevID:   baseRev.RevisionID,
		AuthorID:      "human_eng_1",
		AuthorName:    "Alice",
		Content:       localContent,
		CommitMessage: "Updated Line 1 overview",
	})
	if err != nil {
		t.Fatalf("failed to commit local revision: %v", err)
	}

	// 3. Commit Remote Branch (AI Agent modifies Line 3)
	remoteContent := "Line 1: System Overview\nLine 2: Architecture Details\nLine 3: Appendix (Enriched with Benchmarks)"
	remoteRev, err := engine.CommitRevision(CommitRevisionRequest{
		WorkspaceID:   wsID,
		DocumentID:    docID,
		ParentRevID:   baseRev.RevisionID,
		AuthorID:      "agent_copilot",
		AuthorName:    "AI Assistant",
		IsAIAgent:     true,
		Content:       remoteContent,
		CommitMessage: "Appended benchmark notes to appendix",
	})
	if err != nil {
		t.Fatalf("failed to commit remote revision: %v", err)
	}

	// 4. Execute 3-way merge
	mergeRes, err := engine.MergeRevisions(MergeRevisionsRequest{
		WorkspaceID: wsID,
		DocumentID:  docID,
		BaseRevID:   baseRev.RevisionID,
		LocalRevID:  localRev.RevisionID,
		RemoteRevID: remoteRev.RevisionID,
		Strategy:    StrategyThreeWayMerge,
	})
	if err != nil {
		t.Fatalf("failed to execute merge: %v", err)
	}

	if mergeRes.HasConflicts {
		t.Errorf("expected clean non-conflicting merge, got conflicts: %v", mergeRes.Conflicts)
	}
	if !strings.Contains(mergeRes.MergedContent, "Line 1: System Overview (Updated)") {
		t.Errorf("merged content missing local edit: %s", mergeRes.MergedContent)
	}
	if !strings.Contains(mergeRes.MergedContent, "Line 3: Appendix (Enriched with Benchmarks)") {
		t.Errorf("merged content missing remote edit: %s", mergeRes.MergedContent)
	}
	if mergeRes.MergedRevID == "" {
		t.Errorf("expected newly generated MergedRevID for clean merge")
	}
}

func TestConflictingThreeWayMerge(t *testing.T) {
	engine := NewEngine()
	wsID := "ws_alpha"
	docID := "doc_contract_2"

	baseContent := "Clause A: Target SLA is 99.9%"
	baseRev, _ := engine.CommitRevision(CommitRevisionRequest{
		WorkspaceID: wsID,
		DocumentID:  docID,
		AuthorID:    "admin",
		Content:     baseContent,
	})

	// Local edit
	localContent := "Clause A: Target SLA is 99.95% (Guaranteed)"
	localRev, _ := engine.CommitRevision(CommitRevisionRequest{
		WorkspaceID: wsID,
		DocumentID:  docID,
		ParentRevID: baseRev.RevisionID,
		AuthorID:    "human_lawyer",
		Content:     localContent,
	})

	// Conflicting Remote edit on same line
	remoteContent := "Clause A: Target SLA is 99.99% (Best Effort)"
	remoteRev, _ := engine.CommitRevision(CommitRevisionRequest{
		WorkspaceID: wsID,
		DocumentID:  docID,
		ParentRevID: baseRev.RevisionID,
		AuthorID:    "agent_contract_bot",
		IsAIAgent:   true,
		Content:     remoteContent,
	})

	// Merge under standard 3-way merge
	res, err := engine.MergeRevisions(MergeRevisionsRequest{
		WorkspaceID: wsID,
		DocumentID:  docID,
		BaseRevID:   baseRev.RevisionID,
		LocalRevID:  localRev.RevisionID,
		RemoteRevID: remoteRev.RevisionID,
		Strategy:    StrategyThreeWayMerge,
	})
	if err != nil {
		t.Fatalf("unexpected error during merge: %v", err)
	}

	if !res.HasConflicts {
		t.Errorf("expected conflict flag to be true")
	}
	if len(res.Conflicts) != 1 {
		t.Fatalf("expected 1 conflict hunk, got %d", len(res.Conflicts))
	}
	if !strings.Contains(res.MergedContent, "<<<<<<< LOCAL") {
		t.Errorf("expected conflict markers in merged content: %s", res.MergedContent)
	}
}

func TestSemanticAIArbitrationMerge(t *testing.T) {
	engine := NewEngine()
	wsID := "ws_alpha"
	docID := "doc_spec_ai"

	baseContent := "Requirement 1: Storage latency must be low."
	baseRev, _ := engine.CommitRevision(CommitRevisionRequest{
		WorkspaceID: wsID,
		DocumentID:  docID,
		AuthorID:    "tech_lead",
		Content:     baseContent,
	})

	localRev, _ := engine.CommitRevision(CommitRevisionRequest{
		WorkspaceID: wsID,
		DocumentID:  docID,
		ParentRevID: baseRev.RevisionID,
		AuthorID:    "human_dev",
		Content:     "Requirement 1: Storage latency must be sub-10ms for p95.",
	})

	remoteRev, _ := engine.CommitRevision(CommitRevisionRequest{
		WorkspaceID: wsID,
		DocumentID:  docID,
		ParentRevID: baseRev.RevisionID,
		AuthorID:    "agent_infra",
		IsAIAgent:   true,
		Content:     "Requirement 1: Redis cache layer required for p99 latency reduction.",
	})

	res, err := engine.MergeRevisions(MergeRevisionsRequest{
		WorkspaceID: wsID,
		DocumentID:  docID,
		BaseRevID:   baseRev.RevisionID,
		LocalRevID:  localRev.RevisionID,
		RemoteRevID: remoteRev.RevisionID,
		Strategy:    StrategySemanticAI,
	})
	if err != nil {
		t.Fatalf("unexpected error during semantic merge: %v", err)
	}

	if res.HasConflicts {
		t.Errorf("expected SemanticAI strategy to resolve conflicts automatically without error")
	}
	if !strings.Contains(res.MergedContent, "Integrated") {
		t.Errorf("expected integrated synthesis in semantic merge: %s", res.MergedContent)
	}
}

func TestWorkspaceTenancyIsolation(t *testing.T) {
	engine := NewEngine()

	rev, _ := engine.CommitRevision(CommitRevisionRequest{
		WorkspaceID: "ws_alpha",
		DocumentID:  "doc_secret",
		AuthorID:    "author_1",
		Content:     "Confidential project blueprint",
	})

	// Query from ws_beta should fail with ErrDocumentNotFound
	_, err := engine.GetRevision("ws_beta", "doc_secret", rev.RevisionID)
	if err != ErrDocumentNotFound {
		t.Errorf("expected ErrDocumentNotFound for cross-tenant access, got %v", err)
	}
}

