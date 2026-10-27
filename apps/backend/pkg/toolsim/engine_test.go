package toolsim

import (
	"testing"
)

func TestBuiltInToolsRegistration(t *testing.T) {
	engine := NewEngine()

	tools := engine.ListTools("ws_default")
	if len(tools) < 5 {
		t.Fatalf("expected at least 5 built-in tools, got %d", len(tools))
	}

	foundStripe := false
	for _, tool := range tools {
		if tool.Name == "stripe_charge" {
			foundStripe = true
			if tool.Category != CategoryPayment || tool.RiskLevel != RiskHigh {
				t.Errorf("stripe_charge category/risk mismatch: %s / %s", tool.Category, tool.RiskLevel)
			}
		}
	}

	if !foundStripe {
		t.Errorf("stripe_charge tool not found in registry")
	}
}

func TestExecuteToolDryRunFinancial(t *testing.T) {
	engine := NewEngine()
	wsID := "ws_alpha"

	session, err := engine.CreateSandbox(wsID, "Payment Plan Simulation")
	if err != nil {
		t.Fatalf("failed to create sandbox: %v", err)
	}

	params := map[string]interface{}{
		"amount":      250.00,
		"currency":    "usd",
		"customer_id": "cus_enterprise_991",
	}

	inv, err := engine.ExecuteToolDryRun(wsID, session.ID, "stripe_charge", params)
	if err != nil {
		t.Fatalf("failed to execute tool dry-run: %v", err)
	}

	if inv.Status != "SUCCESS" {
		t.Errorf("expected status SUCCESS, got %s", inv.Status)
	}
	if len(inv.SideEffects) != 1 {
		t.Fatalf("expected 1 side effect, got %d", len(inv.SideEffects))
	}
	if inv.SideEffects[0].EstimatedCost != 250.00 {
		t.Errorf("expected estimated cost 250.00, got %f", inv.SideEffects[0].EstimatedCost)
	}

	report, err := engine.GetReport(wsID, session.ID)
	if err != nil {
		t.Fatalf("failed to get report: %v", err)
	}

	if report.TotalEstimatedCostUSD != 250.00 {
		t.Errorf("expected report cost 250.00, got %f", report.TotalEstimatedCostUSD)
	}
}

func TestExecutePlanDryRunCriticalDestruction(t *testing.T) {
	engine := NewEngine()
	wsID := "ws_alpha"

	req := ExecutePlanRequest{
		WorkspaceID: wsID,
		Title:       "Infrastructure Migration & Cleanup Plan",
		Steps: []PlanStep{
			{
				ToolName: "postgres_exec_query",
				Parameters: map[string]interface{}{
					"query": "SELECT * FROM users WHERE active = true",
				},
			},
			{
				ToolName: "aws_s3_delete_bucket",
				Parameters: map[string]interface{}{
					"bucket_name": "legacy-production-backups-2023",
					"force":       true,
				},
			},
			{
				ToolName: "slack_post_message",
				Parameters: map[string]interface{}{
					"channel": "devops-alerts",
					"text":    "Cleaned up legacy S3 backups.",
				},
			},
		},
	}

	report, err := engine.ExecutePlanDryRun(req)
	if err != nil {
		t.Fatalf("failed to execute plan dry-run: %v", err)
	}

	if report.TotalSteps != 3 {
		t.Errorf("expected 3 total steps, got %d", report.TotalSteps)
	}
	if report.OverallRisk != RiskCritical {
		t.Errorf("expected overall risk CRITICAL, got %s", report.OverallRisk)
	}
	if !report.RequiresHumanApproval {
		t.Errorf("expected RequiresHumanApproval to be true for plan with bucket deletion")
	}
	if len(report.DeletedResources) == 0 {
		t.Errorf("expected deleted resources in report, got 0")
	}
	if report.BlastRadiusScore < 40.0 {
		t.Errorf("expected blast radius >= 40 for critical plan, got %f", report.BlastRadiusScore)
	}
}

func TestWorkspaceTenancyIsolation(t *testing.T) {
	engine := NewEngine()

	session, _ := engine.CreateSandbox("ws_alpha", "Confidential Sandbox")

	// ws_beta should not access ws_alpha sandbox
	_, err := engine.GetSession("ws_beta", session.ID)
	if err != ErrSessionNotFound {
		t.Errorf("expected ErrSessionNotFound for unauthorized workspace, got %v", err)
	}

	_, err = engine.ExecuteToolDryRun("ws_beta", session.ID, "slack_post_message", map[string]interface{}{})
	if err != ErrSessionNotFound {
		t.Errorf("expected ErrSessionNotFound for unauthorized dry-run, got %v", err)
	}
}
