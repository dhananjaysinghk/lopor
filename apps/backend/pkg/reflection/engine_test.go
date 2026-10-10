package reflection_test

import (
	"context"
	"strings"
	"testing"

	"github.com/lopor-ai/lopor/pkg/reflection"
)

func TestReflectionEngine_CritiqueOutput(t *testing.T) {
	engine := reflection.NewReflectionEngine()

	// Defective code with unclosed braces and prohibited comments
	badCode := `func ProcessUser() {
	// A comment that should not exist
	println("Hello")
`
	prompt := "Write a Go function without comments"

	findings, score, err := engine.CritiqueOutput(context.Background(), prompt, badCode, "", "CODE")
	if err != nil {
		t.Fatalf("unexpected error critiquing output: %v", err)
	}

	if score >= 80.0 {
		t.Errorf("expected score < 80 for defective code, got %f", score)
	}

	if len(findings) < 2 {
		t.Errorf("expected at least 2 critique findings, got %d", len(findings))
	}
}

func TestReflectionEngine_SelfCorrectCode(t *testing.T) {
	engine := reflection.NewReflectionEngine()

	buggyCode := `func CalculateTotal() {
	// Calculating total
	res := 100
`
	req := reflection.ReflectionRequest{
		Prompt:           "Write a Go function without comments with error handling",
		CandidateOutput:  buggyCode,
		OutputFormat:     "CODE",
		MaxIterations:    2,
		QualityThreshold: 85.0,
	}

	result, err := engine.SelfCorrect(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error self-correcting: %v", err)
	}

	if result.InitialScore >= result.FinalScore {
		t.Errorf("expected FinalScore (%f) to exceed InitialScore (%f)", result.FinalScore, result.InitialScore)
	}

	if len(result.ImprovementsApplied) == 0 {
		t.Errorf("expected recorded improvements")
	}

	// Verify unclosed brace was automatically resolved
	open := strings.Count(result.RefinedOutput, "{")
	closeB := strings.Count(result.RefinedOutput, "}")
	if open != closeB {
		t.Errorf("expected balanced braces in refined output, got %d open vs %d close", open, closeB)
	}

	// Verify prohibited comment was stripped
	if strings.Contains(result.RefinedOutput, "// Calculating total") {
		t.Errorf("expected comment to be stripped in refined output")
	}
}

func TestReflectionEngine_ApprovedOutput(t *testing.T) {
	engine := reflection.NewReflectionEngine()

	cleanOutput := "SELECT id, email, created_at FROM users WHERE is_active = true;"
	req := reflection.ReflectionRequest{
		Prompt:          "Write a simple SQL query to select active users",
		CandidateOutput: cleanOutput,
		OutputFormat:    "SQL",
	}

	result, err := engine.SelfCorrect(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.IsApproved {
		t.Errorf("expected clean SQL output to be approved, score: %f", result.FinalScore)
	}

	if result.IterationsCount != 0 {
		t.Errorf("expected 0 iterations for already-approved candidate, got %d", result.IterationsCount)
	}
}
