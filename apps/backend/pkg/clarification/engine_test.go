package clarification

import (
	"context"
	"strings"
	"testing"
)

func TestAmbiguousDeploymentQuery(t *testing.T) {
	engine := NewClarificationEngine()
	ctx := context.Background()

	analysis, err := engine.AnalyzeQuery(ctx, "Deploy the application", "")
	if err != nil {
		t.Fatalf("AnalyzeQuery failed: %v", err)
	}

	if analysis.Decision != DecisionNeedsClarification {
		t.Errorf("expected DecisionNeedsClarification for underspecified deployment, got %s", analysis.Decision)
	}
	if len(analysis.Questions) == 0 {
		t.Fatalf("expected clarification questions to be generated")
	}

	foundEnvQ := false
	for _, q := range analysis.Questions {
		if q.ParameterKey == "target_environment" {
			foundEnvQ = true
			if len(q.Options) == 0 {
				t.Errorf("expected options for target_environment")
			}
			foundRecommended := false
			for _, opt := range q.Options {
				if opt.IsRecommended {
					foundRecommended = true
				}
			}
			if !foundRecommended {
				t.Errorf("expected at least one option to be marked as recommended")
			}
		}
	}

	if !foundEnvQ {
		t.Errorf("expected question for target_environment")
	}
}

func TestAmbiguousExportQuery(t *testing.T) {
	engine := NewClarificationEngine()
	ctx := context.Background()

	analysis, err := engine.AnalyzeQuery(ctx, "Export the workspace documents", "")
	if err != nil {
		t.Fatalf("AnalyzeQuery failed: %v", err)
	}

	if analysis.Decision != DecisionNeedsClarification {
		t.Errorf("expected DecisionNeedsClarification for underspecified export, got %s", analysis.Decision)
	}

	foundFormatQ := false
	for _, q := range analysis.Questions {
		if q.ParameterKey == "export_format" {
			foundFormatQ = true
			if len(q.Options) < 3 {
				t.Errorf("expected at least 3 format options, got %d", len(q.Options))
			}
		}
	}

	if !foundFormatQ {
		t.Errorf("expected question for export_format")
	}
}

func TestClearSpecificQueryProceeds(t *testing.T) {
	engine := NewClarificationEngine()
	ctx := context.Background()

	query := "Please export the workspace documents to CSV format for the staging database."
	analysis, err := engine.AnalyzeQuery(ctx, query, "")
	if err != nil {
		t.Fatalf("AnalyzeQuery failed: %v", err)
	}

	if analysis.Decision != DecisionProceed {
		t.Errorf("expected DecisionProceed for clear specific query, got %s", analysis.Decision)
	}
	if len(analysis.Questions) != 0 {
		t.Errorf("expected 0 questions for clear query, got %d", len(analysis.Questions))
	}
}

func TestContextFusion(t *testing.T) {
	engine := NewClarificationEngine()
	ctx := context.Background()

	req := FusionRequest{
		OriginalQuery: "Deploy the service",
		Answers: []ClarificationAnswer{
			{
				QuestionID:    "q_target_env",
				ParameterKey:  "target_environment",
				SelectedValue: "staging",
			},
		},
	}

	result, err := engine.FuseClarification(ctx, req)
	if err != nil {
		t.Fatalf("FuseClarification failed: %v", err)
	}

	if !strings.Contains(result.ResolvedPrompt, "staging") {
		t.Errorf("expected resolved prompt to include 'staging', got: %s", result.ResolvedPrompt)
	}
	if result.ResolvedParameters["target_environment"] != "staging" {
		t.Errorf("expected target_environment parameter to be 'staging'")
	}
}
