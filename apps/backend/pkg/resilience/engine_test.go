package resilience

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestCircuitBreakerTrippingAndRecovery(t *testing.T) {
	cfg := CircuitBreakerConfig{
		FailureThreshold:   2,
		CooldownDuration:   50 * time.Millisecond,
		CanarySuccessCount: 2,
		RequestTimeout:     1 * time.Second,
	}

	cb := NewCircuitBreaker(cfg)

	if !cb.CanExecute() {
		t.Fatalf("expected initial state to permit execution")
	}
	if cb.GetState() != StateClosed {
		t.Errorf("expected StateClosed, got %s", cb.GetState())
	}

	// First failure
	cb.RecordFailure(false)
	if cb.GetState() != StateClosed {
		t.Errorf("expected state to remain closed after 1 failure, got %s", cb.GetState())
	}

	// Second failure -> should trip to Open
	cb.RecordFailure(false)
	if cb.GetState() != StateOpen {
		t.Fatalf("expected state to trip to StateOpen after 2 failures, got %s", cb.GetState())
	}

	if cb.CanExecute() {
		t.Errorf("expected open circuit to reject execution")
	}

	// Sleep past cooldown duration
	time.Sleep(60 * time.Millisecond)

	// Now CanExecute should permit canary and transition to HalfOpen
	if !cb.CanExecute() {
		t.Fatalf("expected canary execution to be allowed after cooldown")
	}
	if cb.GetState() != StateHalfOpen {
		t.Errorf("expected StateHalfOpen, got %s", cb.GetState())
	}

	// 1st canary success
	cb.RecordSuccess(120.0)
	if cb.GetState() != StateHalfOpen {
		t.Errorf("expected StateHalfOpen until canary count satisfied, got %s", cb.GetState())
	}

	// 2nd canary success -> should recover to Closed
	cb.RecordSuccess(110.0)
	if cb.GetState() != StateClosed {
		t.Errorf("expected circuit to recover to StateClosed, got %s", cb.GetState())
	}
}

func TestMultiProviderAutomaticFailover(t *testing.T) {
	engine := NewResilienceEngine()
	ctx := context.Background()

	customPool := ProviderPoolConfig{
		PoolID:      "failover-test-pool",
		WorkspaceID: "ws-res-1",
		Strategy:    StrategyWeightedPriority,
		BreakerConfig: CircuitBreakerConfig{
			FailureThreshold: 2,
			CooldownDuration: 1 * time.Second,
		},
		Providers: []ProviderNode{
			{ID: "primary-p1", Name: "Primary Provider", Model: "model-alpha", Priority: 1},
			{ID: "secondary-p2", Name: "Secondary Provider", Model: "model-beta", Priority: 2},
			{ID: "tertiary-p3", Name: "Tertiary Provider", Model: "model-gamma", Priority: 3},
		},
	}
	engine.RegisterPool(customPool)

	// Mock execFn: Primary fails with 429, Secondary succeeds
	mockExec := func(ctx context.Context, providerID, model string) (string, error) {
		if providerID == "primary-p1" {
			return "", errors.New("HTTP 429: Rate limit quota exceeded")
		}
		if providerID == "secondary-p2" {
			return "Successful synthesized response from secondary provider", nil
		}
		return "", errors.New("unexpected provider called")
	}

	req := ExecutionRequest{
		PoolID:      "failover-test-pool",
		WorkspaceID: "ws-res-1",
		Prompt:      "Analyze system metrics",
	}

	res, err := engine.ExecuteWithFallback(ctx, req, mockExec)
	if err != nil {
		t.Fatalf("ExecuteWithFallback failed: %v", err)
	}

	if res.ProviderUsed != "secondary-p2" {
		t.Errorf("expected failover to secondary-p2, got %s", res.ProviderUsed)
	}
	if !res.FallbackOccurred {
		t.Errorf("expected FallbackOccurred to be true")
	}
	if res.Attempts != 2 {
		t.Errorf("expected 2 attempts, got %d", res.Attempts)
	}
	if len(res.FallbackChain) != 2 {
		t.Errorf("expected 2 entries in fallback chain, got %d", len(res.FallbackChain))
	}
}

func TestCostOptimizedSelection(t *testing.T) {
	poolCfg := ProviderPoolConfig{
		PoolID:      "cost-pool",
		WorkspaceID: "ws-cost-1",
		Strategy:    StrategyCostOptimized,
		Providers: []ProviderNode{
			{ID: "expensive-p1", Name: "Expensive", CostPer1kTokens: 0.03, Priority: 1},
			{ID: "budget-p2", Name: "Budget", CostPer1kTokens: 0.0005, Priority: 2},
			{ID: "mid-p3", Name: "Mid Tier", CostPer1kTokens: 0.005, Priority: 3},
		},
	}

	pool := NewProviderPool(poolCfg)
	candidates, err := pool.GetOrderedCandidates(StrategyCostOptimized)
	if err != nil {
		t.Fatalf("GetOrderedCandidates failed: %v", err)
	}

	if len(candidates) != 3 {
		t.Fatalf("expected 3 candidates, got %d", len(candidates))
	}
	if candidates[0].ID != "budget-p2" {
		t.Errorf("expected cheapest candidate 'budget-p2' first, got '%s'", candidates[0].ID)
	}
}

func TestPoolStatusTelemetry(t *testing.T) {
	engine := NewResilienceEngine()

	status, err := engine.GetPoolStatus("global", "default-pool")
	if err != nil {
		t.Fatalf("GetPoolStatus failed: %v", err)
	}

	if status.TotalNodes == 0 {
		t.Errorf("expected default pool to have nodes, got 0")
	}
	if status.HealthyNodes != status.TotalNodes {
		t.Errorf("expected all initial nodes to be healthy, got %d/%d", status.HealthyNodes, status.TotalNodes)
	}
}
