package resilience

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// ResilienceEngine coordinates provider pools, automated failovers, and circuit breaker telemetry.
type ResilienceEngine struct {
	mu    sync.RWMutex
	pools map[string]*ProviderPool // key: wsID + ":" + poolID
}

// NewResilienceEngine initializes the resilience engine with an enterprise default fallback pool.
func NewResilienceEngine() *ResilienceEngine {
	engine := &ResilienceEngine{
		pools: make(map[string]*ProviderPool),
	}
	engine.initDefaultPool()
	return engine
}

func (e *ResilienceEngine) initDefaultPool() {
	defaultConfig := ProviderPoolConfig{
		PoolID:      "default-pool",
		WorkspaceID: "global",
		Name:        "Global Multi-Provider Resilience Ensemble",
		Strategy:    StrategyWeightedPriority,
		BreakerConfig: CircuitBreakerConfig{
			FailureThreshold:   3,
			CooldownDuration:   30 * time.Second,
			CanarySuccessCount: 2,
			RequestTimeout:     10 * time.Second,
		},
		Providers: []ProviderNode{
			{ID: "openai-gpt4o", Name: "OpenAI GPT-4o", Model: "gpt-4o", CostPer1kTokens: 0.005, Priority: 1},
			{ID: "anthropic-claude35", Name: "Anthropic Claude 3.5 Sonnet", Model: "claude-3-5-sonnet", CostPer1kTokens: 0.003, Priority: 2},
			{ID: "deepseek-coder", Name: "DeepSeek Coder V2", Model: "deepseek-coder", CostPer1kTokens: 0.0005, Priority: 3},
			{ID: "ollama-llama3", Name: "Ollama Llama 3 (Local)", Model: "llama3-8b", CostPer1kTokens: 0.0, Priority: 4},
		},
		UpdatedAt: time.Now().UTC(),
	}

	e.RegisterPool(defaultConfig)
}

func poolKey(wsID, pID string) string {
	if wsID == "" {
		wsID = "global"
	}
	if pID == "" {
		pID = "default-pool"
	}
	return wsID + ":" + pID
}

// RegisterPool adds or updates a provider pool configuration.
func (e *ResilienceEngine) RegisterPool(cfg ProviderPoolConfig) {
	e.mu.Lock()
	defer e.mu.Unlock()

	cfg.UpdatedAt = time.Now().UTC()
	k := poolKey(cfg.WorkspaceID, cfg.PoolID)
	e.pools[k] = NewProviderPool(cfg)
}

// GetPoolStatus returns live telemetry for a workspace provider pool.
func (e *ResilienceEngine) GetPoolStatus(workspaceID, poolID string) (*ProviderPoolStatus, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	k := poolKey(workspaceID, poolID)
	pool, exists := e.pools[k]
	if !exists {
		// Fallback to global default pool
		pool, exists = e.pools[poolKey("global", "default-pool")]
	}

	if !exists {
		return nil, fmt.Errorf("provider pool '%s' not found for workspace '%s'", poolID, workspaceID)
	}

	status := pool.GetStatus()
	return &status, nil
}

// ExecuteWithFallback executes an operation with dynamic load balancing, health monitoring, and cascading failover.
func (e *ResilienceEngine) ExecuteWithFallback(
	ctx context.Context,
	req ExecutionRequest,
	execFn func(ctx context.Context, providerID, model string) (string, error),
) (*ExecutionResult, error) {
	e.mu.RLock()
	k := poolKey(req.WorkspaceID, req.PoolID)
	pool, exists := e.pools[k]
	if !exists {
		pool = e.pools[poolKey("global", "default-pool")]
	}
	e.mu.RUnlock()

	if pool == nil {
		return nil, fmt.Errorf("no operational provider pool configured")
	}

	candidates, err := pool.GetOrderedCandidates(pool.config.Strategy)
	if err != nil {
		return nil, err
	}

	var fallbackChain []string
	attempts := 0
	overallStart := time.Now()

	for _, candidate := range candidates {
		attempts++
		fallbackChain = append(fallbackChain, fmt.Sprintf("%s (%s)", candidate.ID, candidate.Model))

		callStart := time.Now()
		out, callErr := execFn(ctx, candidate.ID, candidate.Model)
		lat := float64(time.Since(callStart).Milliseconds())

		if callErr == nil {
			// Success! Record latency and clear failures
			pool.RecordSuccess(candidate.ID, lat)

			return &ExecutionResult{
				ProviderUsed:     candidate.ID,
				ModelUsed:        candidate.Model,
				Attempts:         attempts,
				Output:           out,
				LatencyMs:        time.Since(overallStart).Milliseconds(),
				FallbackOccurred: attempts > 1,
				FallbackChain:    fallbackChain,
			}, nil
		}

		// Failure handling
		isRateLimit := strings.Contains(strings.ToLower(callErr.Error()), "429") ||
			strings.Contains(strings.ToLower(callErr.Error()), "rate limit")

		pool.RecordFailure(candidate.ID, isRateLimit)
	}

	return nil, fmt.Errorf("all %d provider candidates in pool exhausted. Fallback chain: %s",
		attempts, strings.Join(fallbackChain, " -> "))
}
