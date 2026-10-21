package simulation

import (
	"testing"
)

func TestCreateSession(t *testing.T) {
	engine := NewEngine()

	// Preset test
	req := StartSimulationRequest{
		WorkspaceID: "ws_alpha",
		Title:       "Critical Database Failover Drill",
		Scenario:    "Primary Postgres instance experiencing high I/O wait and replication lag during peak hours.",
		Preset:      PresetIncidentDrill,
	}

	session, err := engine.CreateSession(req)
	if err != nil {
		t.Fatalf("unexpected error creating session: %v", err)
	}

	if session.ID == "" {
		t.Errorf("expected session ID to be non-empty")
	}
	if len(session.Personas) != 4 {
		t.Errorf("expected 4 personas for INCIDENT_DRILL preset, got %d", len(session.Personas))
	}
	if session.Status != StatusRunning {
		t.Errorf("expected status RUNNING, got %s", session.Status)
	}

	// Invalid request test
	_, err = engine.CreateSession(StartSimulationRequest{WorkspaceID: ""})
	if err == nil {
		t.Errorf("expected error on empty request, got nil")
	}
}

func TestAdvanceStepWithHumanInjection(t *testing.T) {
	engine := NewEngine()

	req := StartSimulationRequest{
		WorkspaceID: "ws_alpha",
		Title:       "Vendor Negotiation",
		Scenario:    "Enterprise SaaS pricing renewal and SLA commitments.",
		Preset:      PresetNegotiationWarRoom,
	}

	session, err := engine.CreateSession(req)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// 1. Advance single step
	stepRes, err := engine.AdvanceStep("ws_alpha", session.ID, AdvanceStepRequest{TargetSteps: 1})
	if err != nil {
		t.Fatalf("failed to advance step: %v", err)
	}

	if len(stepRes.AddedTurns) != 1 {
		t.Fatalf("expected 1 added turn, got %d", len(stepRes.AddedTurns))
	}
	if stepRes.CurrentTurn != 1 {
		t.Errorf("expected current turn 1, got %d", stepRes.CurrentTurn)
	}

	// 2. Advance with human intervention
	humanRes, err := engine.AdvanceStep("ws_alpha", session.ID, AdvanceStepRequest{
		HumanInput:  "We can offer a 15% discount if the client signs a multi-year commit.",
		HumanName:   "VP of Sales",
		TargetSteps: 1,
	})
	if err != nil {
		t.Fatalf("failed to advance step with human input: %v", err)
	}

	// Should have added human turn + 1 persona turn = 2 turns
	if len(humanRes.AddedTurns) != 2 {
		t.Errorf("expected 2 added turns (human + persona), got %d", len(humanRes.AddedTurns))
	}
	if humanRes.CurrentTurn != 3 {
		t.Errorf("expected current turn 3, got %d", humanRes.CurrentTurn)
	}
}

func TestRunToCompletionAndReport(t *testing.T) {
	engine := NewEngine()

	req := StartSimulationRequest{
		WorkspaceID: "ws_alpha",
		Title:       "Executive AI Strategy Review",
		Scenario:    "Budget allocation for generative AI tooling versus security remediation.",
		Preset:      PresetExecutiveReview,
		Config: &SimulationConfig{
			MaxTurns:                 6,
			AutoConvergenceDetection: true,
			ConvergenceThreshold:     0.80,
		},
	}

	session, err := engine.CreateSession(req)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	report, err := engine.RunToCompletion("ws_alpha", session.ID)
	if err != nil {
		t.Fatalf("failed to run simulation to completion: %v", err)
	}

	if report.SessionID != session.ID {
		t.Errorf("report session ID mismatch: %s vs %s", report.SessionID, session.ID)
	}
	if len(report.Turns) == 0 {
		t.Errorf("expected turns in report, got 0")
	}
	if report.FinalVerdict == "" {
		t.Errorf("expected non-empty final verdict")
	}
	if len(report.ActionItems) == 0 {
		t.Errorf("expected actionable items in report, got 0")
	}
	if len(report.SentimentTrajectory) != len(report.Turns) {
		t.Errorf("expected sentiment trajectory count %d to match turns %d", len(report.SentimentTrajectory), len(report.Turns))
	}
}

func TestWorkspaceTenancyIsolation(t *testing.T) {
	engine := NewEngine()

	req := StartSimulationRequest{
		WorkspaceID: "ws_alpha",
		Title:       "Tenant Alpha War Room",
		Scenario:    "Confidential security drill.",
		Preset:      PresetIncidentDrill,
	}

	session, err := engine.CreateSession(req)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Accessing from another workspace should fail with ErrSessionNotFound
	_, err = engine.GetSession("ws_beta", session.ID)
	if err != ErrSessionNotFound {
		t.Errorf("expected ErrSessionNotFound for unauthorized workspace, got %v", err)
	}

	_, err = engine.AdvanceStep("ws_beta", session.ID, AdvanceStepRequest{TargetSteps: 1})
	if err != ErrSessionNotFound {
		t.Errorf("expected ErrSessionNotFound for unauthorized workspace step, got %v", err)
	}
}
