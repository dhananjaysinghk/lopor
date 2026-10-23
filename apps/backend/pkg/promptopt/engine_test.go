package promptopt

import (
	"strings"
	"testing"
)

func TestStartOptimizationTournament(t *testing.T) {
	engine := NewEngine()

	examples := []PromptExample{
		{
			ID: "ex_1",
			InputVariables: map[string]string{
				"query": "Extract payment total from invoice 4410",
			},
			TargetOutput: "{\"total\": 450.00, \"currency\": \"USD\"}",
			Rationale:    "Exact currency and decimal total extracted without conversational filler",
			Score:        95.0,
		},
		{
			ID: "ex_2",
			InputVariables: map[string]string{
				"query": "Summarize customer complaint and refund eligibility",
			},
			TargetOutput: "Flawed summary with missing policy citation and broken json",
			Rationale:    "Omitted policy citation and hallucinated eligibility threshold",
			Score:        40.0,
			IsNegativeExemplar: true,
		},
	}

	req := StartOptimizationRequest{
		WorkspaceID: "ws_alpha",
		Title:       "Enterprise Invoice Extraction Optimization",
		BasePrompt:  "You are an assistant that extracts data from customer queries into JSON format.",
		Method:      MethodHybridDSPy,
		Metric:      MetricComposite,
		Examples:    examples,
	}

	job, err := engine.StartOptimization(req)
	if err != nil {
		t.Fatalf("failed to start optimization: %v", err)
	}

	if job.ID == "" {
		t.Errorf("expected non-empty job ID")
	}
	if job.Status != StatusCompleted {
		t.Errorf("expected job status COMPLETED, got %s", job.Status)
	}
	if len(job.Candidates) < 5 {
		t.Errorf("expected at least 5 candidates in tournament, got %d", len(job.Candidates))
	}
	if job.BestCandidate == nil {
		t.Fatalf("expected non-nil BestCandidate")
	}
	if job.ScoreImprovement <= 0 {
		t.Errorf("expected positive score improvement over baseline, got %.2f", job.ScoreImprovement)
	}
}

func TestPromptReportGeneration(t *testing.T) {
	engine := NewEngine()

	req := StartOptimizationRequest{
		WorkspaceID: "ws_alpha",
		Title:       "Customer Service Intent Prompt",
		BasePrompt:  "Classify the customer message into billing, tech_support, or sales.",
	}

	job, err := engine.StartOptimization(req)
	if err != nil {
		t.Fatalf("failed to start optimization: %v", err)
	}

	report, err := engine.GetReport("ws_alpha", job.ID)
	if err != nil {
		t.Fatalf("failed to get report: %v", err)
	}

	if report.JobID != job.ID {
		t.Errorf("mismatched job ID in report: %s vs %s", report.JobID, job.ID)
	}
	if report.OptimizationDiff == "" {
		t.Errorf("expected non-empty optimization diff")
	}
	if !strings.Contains(report.OptimizationDiff, "BASELINE PROMPT") {
		t.Errorf("diff missing BASELINE section: %s", report.OptimizationDiff)
	}
	if len(report.Recommendations) == 0 {
		t.Errorf("expected recommendations in report, got 0")
	}
}

func TestWorkspaceTenancyIsolation(t *testing.T) {
	engine := NewEngine()

	req := StartOptimizationRequest{
		WorkspaceID: "ws_alpha",
		Title:       "Alpha Internal Classifier",
		BasePrompt:  "Confidential prompt instructions.",
	}

	job, err := engine.StartOptimization(req)
	if err != nil {
		t.Fatalf("failed to start optimization: %v", err)
	}

	// Access from ws_beta should be rejected
	_, err = engine.GetJob("ws_beta", job.ID)
	if err != ErrJobNotFound {
		t.Errorf("expected ErrJobNotFound for cross-tenant access, got %v", err)
	}

	_, err = engine.GetReport("ws_beta", job.ID)
	if err != ErrJobNotFound {
		t.Errorf("expected ErrJobNotFound for cross-tenant report access, got %v", err)
	}
}
