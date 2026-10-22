package selfheal

import (
	"context"
	"strings"
	"testing"
)

func TestCoerceJSON(t *testing.T) {
	// Markdown fence + trailing comma
	input := "```json\n{\n  \"status\": \"success\",\n  \"code\": 200,\n}\n```"
	coerced, ok := CoerceJSON(input)
	if !ok {
		t.Fatalf("expected successful JSON coercion, got failure")
	}

	if !strings.Contains(coerced, `"status": "success"`) || !strings.Contains(coerced, `"code": 200`) {
		t.Errorf("coerced output missing expected keys: %s", coerced)
	}

	// Conversational filler surrounding JSON
	conversational := "Here is the result you requested:\n{ \"data\": [1, 2, 3] }\nHope this helps!"
	coerced2, ok2 := CoerceJSON(conversational)
	if !ok2 {
		t.Fatalf("expected successful boundary extraction from conversational text")
	}
	if !strings.Contains(coerced2, `"data"`) {
		t.Errorf("unexpected coerced content: %s", coerced2)
	}
}

func TestRemediateSchemaCoercion(t *testing.T) {
	engine := NewEngine()
	ctx := context.Background()

	fault := ExecutionFault{
		WorkspaceID:     "ws_test",
		Category:        FaultSchemaValidation,
		ErrorMessage:    "invalid character '}' looking for beginning of object key string",
		OffendingOutput: "```json\n{\n  \"action\": \"DEPLOY\",\n  \"target\": \"production\",\n}\n```",
	}

	res, err := engine.Remediate(ctx, fault, nil)
	if err != nil {
		t.Fatalf("unexpected error during remediation: %v", err)
	}

	if res.Status != StatusHealed {
		t.Errorf("expected StatusHealed, got %s", res.Status)
	}
	if res.PrimaryStrategy != StrategyCoercion {
		t.Errorf("expected PrimaryStrategy StrategyCoercion, got %s", res.PrimaryStrategy)
	}
	if !strings.Contains(res.RepairedOutput, `"action": "DEPLOY"`) {
		t.Errorf("repaired output missing valid JSON: %s", res.RepairedOutput)
	}
}

func TestRemediateContextPruning(t *testing.T) {
	engine := NewEngine()
	ctx := context.Background()

	longPrompt := strings.Repeat("Long context preamble. ", 500)
	fault := ExecutionFault{
		WorkspaceID:   "ws_test",
		Category:      FaultContextLimitExceeded,
		ErrorMessage:  "maximum context length is 8192 tokens, however your messages resulted in 9400 tokens",
		OriginalInput: longPrompt,
	}

	res, err := engine.Remediate(ctx, fault, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.Status != StatusHealed {
		t.Errorf("expected StatusHealed, got %s", res.Status)
	}
	if res.PrimaryStrategy != StrategyContextPrune {
		t.Errorf("expected StrategyContextPrune, got %s", res.PrimaryStrategy)
	}
}

func TestRemediateProviderFallback(t *testing.T) {
	engine := NewEngine()
	ctx := context.Background()

	fault := ExecutionFault{
		WorkspaceID:  "ws_test",
		Category:     FaultUpstreamOutage,
		ErrorMessage: "503 service unavailable from anthropic upstream",
	}

	res, err := engine.Remediate(ctx, fault, &RemediationPolicy{
		EnableProviderFallback: true,
		FallbackProviders:      []string{"gpt-4o"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.Status != StatusHealed {
		t.Errorf("expected StatusHealed, got %s", res.Status)
	}
	if res.EffectiveModelID != "gpt-4o" {
		t.Errorf("expected effective model gpt-4o, got %s", res.EffectiveModelID)
	}
}

func TestTelemetryAccounting(t *testing.T) {
	engine := NewEngine()
	ctx := context.Background()

	fault := ExecutionFault{
		WorkspaceID:     "ws_telemetry",
		Category:        FaultSchemaValidation,
		OffendingOutput: "{ \"valid\": true }",
	}

	_, _ = engine.Remediate(ctx, fault, nil)

	telemetry := engine.GetTelemetry("ws_telemetry")
	if telemetry.TotalFaultsDetected != 1 {
		t.Errorf("expected 1 detected fault, got %d", telemetry.TotalFaultsDetected)
	}
	if telemetry.TotalHealed != 1 {
		t.Errorf("expected 1 healed fault, got %d", telemetry.TotalHealed)
	}
	if telemetry.HealingSuccessRate != 100.0 {
		t.Errorf("expected 100%% success rate, got %.2f", telemetry.HealingSuccessRate)
	}
}
