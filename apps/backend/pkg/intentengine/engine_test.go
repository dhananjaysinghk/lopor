package intentengine

import (
	"testing"
)

func TestFastFactualLowComplexityRouting(t *testing.T) {
	engine := NewEngine()
	wsID := "ws_alpha"

	req := RoutePromptRequest{
		WorkspaceID: wsID,
		Prompt:      "What is the capital of France and what is its primary currency?",
		OptimizeFor: "BALANCED",
	}

	decision, err := engine.Route(req)
	if err != nil {
		t.Fatalf("failed to route prompt: %v", err)
	}

	if decision.PrimaryDomain != DomainFastFactual {
		t.Errorf("expected DomainFastFactual, got %s", decision.PrimaryDomain)
	}
	if decision.SelectedModel.Tier != TierUltraFast {
		t.Errorf("expected TierUltraFast for simple factual question, got %s", decision.SelectedModel.Tier)
	}
	if decision.PotentialSavingsUSD <= 0 {
		t.Errorf("expected positive potential cost savings vs flagship")
	}
	if decision.RoutingLatencyMicros <= 0 {
		t.Errorf("expected recorded microsecond router latency")
	}
}

func TestComplexReasoningAndMathRouting(t *testing.T) {
	engine := NewEngine()
	wsID := "ws_alpha"

	prompt := "Derive the mathematical proof for distributed Byzantine consensus under partial synchrony. Calculate the step-by-step bounds on network delay tradeoff delta."

	req := RoutePromptRequest{
		WorkspaceID: wsID,
		Prompt:      prompt,
		OptimizeFor: "BALANCED",
	}

	decision, err := engine.Route(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if decision.PrimaryDomain != DomainComplexReasoning {
		t.Errorf("expected DomainComplexReasoning, got %s", decision.PrimaryDomain)
	}
	if decision.SelectedModel.Tier != TierReasoning {
		t.Errorf("expected TierReasoning for formal mathematical proof, got %s", decision.SelectedModel.Tier)
	}
	if !decision.ReasoningRequired {
		t.Errorf("expected ReasoningRequired to be true")
	}
}

func TestCodeEngineeringRouting(t *testing.T) {
	engine := NewEngine()
	wsID := "ws_alpha"

	codePrompt := `
func ProcessBatch(items []string) error {
	for _, it := range items {
		if err := validate(it); err != nil {
			return err
		}
	}
	return nil
}
Refactor this function to process items concurrently using a worker pool with errgroup.
`

	req := RoutePromptRequest{
		WorkspaceID: wsID,
		Prompt:      codePrompt,
		OptimizeFor: "QUALITY",
	}

	decision, err := engine.Route(req)
	if err != nil {
		t.Fatalf("failed to route code prompt: %v", err)
	}

	if decision.PrimaryDomain != DomainCodeEngineering {
		t.Errorf("expected DomainCodeEngineering, got %s", decision.PrimaryDomain)
	}
	if decision.Complexity.CodeDensityScore < 0.20 {
		t.Errorf("expected high code density score, got %f", decision.Complexity.CodeDensityScore)
	}
}

func TestRouterTelemetryAccounting(t *testing.T) {
	engine := NewEngine()
	wsID := "ws_telemetry"

	_, _ = engine.Route(RoutePromptRequest{
		WorkspaceID: wsID,
		Prompt:      "Brief summary of AI history",
		OptimizeFor: "COST",
	})
	_, _ = engine.Route(RoutePromptRequest{
		WorkspaceID: wsID,
		Prompt:      "Prove Fermat's Last Theorem using elliptic curves",
		OptimizeFor: "QUALITY",
	})

	telemetry := engine.GetTelemetry(wsID)
	if telemetry.TotalRoutedRequests != 2 {
		t.Errorf("expected 2 total requests, got %d", telemetry.TotalRoutedRequests)
	}
	if telemetry.CumulativeSavedUSD <= 0 {
		t.Errorf("expected positive cumulative cost avoidance")
	}
}

