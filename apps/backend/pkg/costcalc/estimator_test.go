package costcalc_test

import (
	"context"
	"testing"

	"github.com/lopor-ai/lopor/pkg/costcalc"
)

func TestCostEstimator_GPT4o(t *testing.T) {
	estimator := costcalc.NewCostEstimator()

	req := costcalc.CostEstimateRequest{
		ModelID:        "gpt-4o",
		PromptTokens:   10_000,
		ExpectedOutput: 2_000,
		CachedTokens:   4_000,
	}

	breakdown, err := estimator.EstimateCost(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error estimating cost: %v", err)
	}

	if breakdown.ModelID != "gpt-4o" {
		t.Errorf("expected model gpt-4o, got %s", breakdown.ModelID)
	}

	if breakdown.TotalCostUSD <= 0 {
		t.Errorf("expected positive cost, got %f", breakdown.TotalCostUSD)
	}

	if breakdown.CacheSavingsUSD <= 0 {
		t.Errorf("expected cache savings > 0 when cached tokens are present")
	}
}

func TestCostEstimator_PromptTextEstimation(t *testing.T) {
	estimator := costcalc.NewCostEstimator()

	req := costcalc.CostEstimateRequest{
		ModelID:    "claude-3-5-sonnet",
		PromptText: "Analyze this 50 page annual report and highlight key risks.",
	}

	breakdown, err := estimator.EstimateCost(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if breakdown.PromptTokens <= 0 {
		t.Errorf("expected prompt tokens > 0 calculated from text")
	}

	catalog := estimator.GetCatalog()
	if len(catalog) < 5 {
		t.Errorf("expected at least 5 cataloged models, got %d", len(catalog))
	}
}
