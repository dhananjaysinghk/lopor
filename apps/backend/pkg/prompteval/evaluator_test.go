package prompteval_test

import (
	"context"
	"testing"

	"github.com/lopor-ai/lopor/pkg/prompteval"
)

func TestPromptEvaluator_Evaluate(t *testing.T) {
	evaluator := prompteval.NewEvaluator()

	template := "You are a senior analyst. Summarize the following quarterly earnings data accurately: {{earnings_report}}"

	req := prompteval.EvalRequest{
		PromptTemplate: template,
		ModelTarget:    "gpt-4o",
		ExpectedTone:   "analytical",
		TestCases: []prompteval.TestCase{
			{
				ID:             "case-q3-report",
				Variables:      map[string]string{"earnings_report": "Q3 Revenue grew 18% YoY to $42M."},
				ExpectedOutput: "Revenue grew 18% YoY to $42M",
				TargetKeywords: []string{"Revenue", "YoY", "earnings"},
			},
		},
	}

	res, err := evaluator.EvaluatePrompt(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error running prompt evaluation: %v", err)
	}

	if res.OverallScore < 80.0 {
		t.Errorf("expected overall score >= 80, got %f", res.OverallScore)
	}

	if res.PassedRate != 100.0 {
		t.Errorf("expected 100%% pass rate, got %f", res.PassedRate)
	}

	if res.Status != "PASSED" {
		t.Errorf("expected status PASSED, got %s", res.Status)
	}

	if len(res.CaseScores) != 1 {
		t.Errorf("expected 1 test case score, got %d", len(res.CaseScores))
	}

	if len(res.CaseScores[0].MatchedKeywords) == 0 {
		t.Errorf("expected matched keywords in case score")
	}
}

func TestPromptEvaluator_EmptyTemplate(t *testing.T) {
	evaluator := prompteval.NewEvaluator()
	_, err := evaluator.EvaluatePrompt(context.Background(), prompteval.EvalRequest{
		PromptTemplate: "",
	})
	if err == nil {
		t.Errorf("expected error on empty prompt template, got nil")
	}
}
