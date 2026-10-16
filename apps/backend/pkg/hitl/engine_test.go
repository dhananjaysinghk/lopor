package hitl

import (
	"context"
	"testing"
)

func TestSafeActionAutoPass(t *testing.T) {
	engine := NewHITLEngine()
	ctx := context.Background()

	req := InterceptRequest{
		WorkspaceID: "ws-hitl-1",
		ToolName:    "search_semantic",
		Summary:     "Search knowledge base for Q3 metrics",
	}

	ticket, requiresApproval, err := engine.EvaluateAndIntercept(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if requiresApproval {
		t.Errorf("expected safe action 'search_semantic' to auto-pass without approval")
	}
	if ticket != nil {
		t.Errorf("expected nil ticket for auto-pass action")
	}
}

func TestModerateActionSingleSignature(t *testing.T) {
	engine := NewHITLEngine()
	ctx := context.Background()

	req := InterceptRequest{
		WorkspaceID:    "ws-hitl-1",
		ToolName:       "run_sandbox",
		Summary:        "Execute Python script to plot revenue distribution",
		PayloadPreview: "python main.py",
		Parameters:     map[string]interface{}{"timeout": 10},
	}

	ticket, requiresApproval, err := engine.EvaluateAndIntercept(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !requiresApproval {
		t.Fatalf("expected 'run_sandbox' to require human approval")
	}
	if ticket.Status != StatusPending {
		t.Errorf("expected StatusPending, got %s", ticket.Status)
	}
	if ticket.RequiredApprovals != 1 {
		t.Errorf("expected 1 required approval, got %d", ticket.RequiredApprovals)
	}

	// Approve ticket
	sub := ReviewSubmission{
		TicketID:     ticket.ID,
		ReviewerID:   "usr-reviewer-1",
		ReviewerName: "Alice Admin",
		Decision:     DecisionApprove,
		Feedback:     "Looks clean, execution permitted",
	}

	reviewed, err := engine.SubmitReview(ctx, sub)
	if err != nil {
		t.Fatalf("SubmitReview failed: %v", err)
	}

	if reviewed.Status != StatusApproved {
		t.Errorf("expected StatusApproved, got %s", reviewed.Status)
	}
	if reviewed.CurrentApprovals != 1 {
		t.Errorf("expected 1 current approval, got %d", reviewed.CurrentApprovals)
	}
}

func TestCriticalActionDualSignature(t *testing.T) {
	engine := NewHITLEngine()
	ctx := context.Background()

	req := InterceptRequest{
		WorkspaceID:    "ws-hitl-1",
		ToolName:       "execute_sql",
		Summary:        "Execute schema migration index rebuild",
		PayloadPreview: "REINDEX TABLE document_embeddings;",
	}

	ticket, requiresApproval, err := engine.EvaluateAndIntercept(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !requiresApproval {
		t.Fatalf("expected 'execute_sql' to require approval")
	}
	if ticket.RequiredApprovals != 2 {
		t.Errorf("expected 2 required approvals for critical action, got %d", ticket.RequiredApprovals)
	}

	// First reviewer approves
	sub1 := ReviewSubmission{
		TicketID:     ticket.ID,
		ReviewerID:   "usr-eng-lead",
		ReviewerName: "Bob Lead",
		Decision:     DecisionApprove,
	}
	t1, err := engine.SubmitReview(ctx, sub1)
	if err != nil {
		t.Fatalf("first review failed: %v", err)
	}

	if t1.Status != StatusPending {
		t.Errorf("expected ticket to remain StatusPending after only 1 of 2 approvals, got %s", t1.Status)
	}
	if t1.CurrentApprovals != 1 {
		t.Errorf("expected 1 approval count, got %d", t1.CurrentApprovals)
	}

	// Second reviewer approves
	sub2 := ReviewSubmission{
		TicketID:     ticket.ID,
		ReviewerID:   "usr-sre-manager",
		ReviewerName: "Charlie SRE",
		Decision:     DecisionApprove,
	}
	t2, err := engine.SubmitReview(ctx, sub2)
	if err != nil {
		t.Fatalf("second review failed: %v", err)
	}

	if t2.Status != StatusApproved {
		t.Errorf("expected ticket to be StatusApproved after 2 approvals, got %s", t2.Status)
	}
	if t2.CurrentApprovals != 2 {
		t.Errorf("expected 2 approval counts, got %d", t2.CurrentApprovals)
	}
}

func TestRejectionWithFeedback(t *testing.T) {
	engine := NewHITLEngine()
	ctx := context.Background()

	req := InterceptRequest{
		WorkspaceID: "ws-hitl-1",
		ToolName:    "delete_workspace_records",
		Summary:     "Purge records older than 90 days",
	}

	ticket, _, _ := engine.EvaluateAndIntercept(ctx, req)

	sub := ReviewSubmission{
		TicketID:     ticket.ID,
		ReviewerID:   "usr-compliance-officer",
		ReviewerName: "Dana Compliance",
		Decision:     DecisionReject,
		Feedback:     "Retention policy requires 180 days minimum before archival purge.",
	}

	reviewed, err := engine.SubmitReview(ctx, sub)
	if err != nil {
		t.Fatalf("SubmitReview failed: %v", err)
	}

	if reviewed.Status != StatusRejected {
		t.Errorf("expected StatusRejected, got %s", reviewed.Status)
	}
	if len(reviewed.Reviews) != 1 || reviewed.Reviews[0].Feedback == "" {
		t.Errorf("expected review feedback to be saved")
	}
}

func TestAlwaysBlockedAction(t *testing.T) {
	engine := NewHITLEngine()
	ctx := context.Background()

	req := InterceptRequest{
		WorkspaceID: "ws-hitl-1",
		ToolName:    "drop_database_all",
		Summary:     "Drop primary production database",
	}

	ticket, requiresApproval, err := engine.EvaluateAndIntercept(ctx, req)
	if err == nil {
		t.Fatalf("expected error for strictly blocked action, got nil")
	}

	if !requiresApproval {
		t.Errorf("expected requiresApproval to be true for blocked action")
	}
	if ticket.Status != StatusRejected {
		t.Errorf("expected StatusRejected on blocked ticket, got %s", ticket.Status)
	}
}
