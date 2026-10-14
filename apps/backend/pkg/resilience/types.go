package resilience

import (
	"time"
)

// CircuitState represents the operational health of an AI provider endpoint.
type CircuitState string

const (
	StateClosed   CircuitState = "CLOSED"    // Normal healthy operation
	StateOpen     CircuitState = "OPEN"      // Tripped; temporarily rejecting requests
	StateHalfOpen CircuitState = "HALF_OPEN" // Trial recovery mode
)

// BalanceStrategy defines the provider selection policy.
type BalanceStrategy string

const (
	StrategyLeastLatency     BalanceStrategy = "LEAST_LATENCY"
	StrategyRoundRobin       BalanceStrategy = "ROUND_ROBIN"
	StrategyWeightedPriority BalanceStrategy = "WEIGHTED_PRIORITY"
	StrategyCostOptimized    BalanceStrategy = "COST_OPTIMIZED"
)

// ProviderNode encapsulates an AI provider's configuration and telemetry.
type ProviderNode struct {
	ID                  string       `json:"id"`
	Name                string       `json:"name"`
	Model               string       `json:"model"`
	CostPer1kTokens     float64      `json:"cost_per_1k_tokens"`
	Priority            int          `json:"priority"` // 1 = highest
	State               CircuitState `json:"state"`
	ConsecutiveFailures int          `json:"consecutive_failures"`
	TotalRequests       int64        `json:"total_requests"`
	TotalFailures       int64        `json:"total_failures"`
	AverageLatencyMs    float64      `json:"average_latency_ms"`
	LastFailureTime     time.Time    `json:"last_failure_time,omitempty"`
	NextCanaryTime      time.Time    `json:"next_canary_time,omitempty"`
	RateLimitCount      int          `json:"rate_limit_count"`
}

// CircuitBreakerConfig governs failure trip thresholds and cooldown durations.
type CircuitBreakerConfig struct {
	FailureThreshold   int           `json:"failure_threshold"`
	CooldownDuration   time.Duration `json:"cooldown_duration"`
	CanarySuccessCount int           `json:"canary_success_count"`
	RequestTimeout     time.Duration `json:"request_timeout"`
}

// ProviderPoolConfig defines an ensemble of AI providers configured for a workspace.
type ProviderPoolConfig struct {
	PoolID        string               `json:"pool_id"`
	WorkspaceID   string               `json:"workspace_id"`
	Name          string               `json:"name"`
	Strategy      BalanceStrategy      `json:"strategy"`
	Providers     []ProviderNode       `json:"providers"`
	BreakerConfig CircuitBreakerConfig `json:"breaker_config"`
	UpdatedAt     time.Time            `json:"updated_at"`
}

// ExecutionRequest represents a call to be dispatched with automated resilience.
type ExecutionRequest struct {
	PoolID      string `json:"pool_id"`
	WorkspaceID string `json:"workspace_id"`
	Prompt      string `json:"prompt"`
	MaxTokens   int    `json:"max_tokens,omitempty"`
}

// ExecutionResult captures the outcome, including failovers and latency.
type ExecutionResult struct {
	ProviderUsed     string   `json:"provider_used"`
	ModelUsed        string   `json:"model_used"`
	Attempts         int      `json:"attempts"`
	Output           string   `json:"output"`
	LatencyMs        int64    `json:"latency_ms"`
	FallbackOccurred bool     `json:"fallback_occurred"`
	FallbackChain    []string `json:"fallback_chain"`
}

// ProviderPoolStatus provides telemetry across all nodes in a pool.
type ProviderPoolStatus struct {
	PoolID      string         `json:"pool_id"`
	WorkspaceID string         `json:"workspace_id"`
	Strategy    BalanceStrategy `json:"strategy"`
	Nodes       []ProviderNode `json:"nodes"`
	TotalNodes  int            `json:"total_nodes"`
	HealthyNodes int           `json:"healthy_nodes"`
}
