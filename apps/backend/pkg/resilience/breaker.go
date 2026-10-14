package resilience

import (
	"sync"
	"time"
)

// CircuitBreaker manages the failure threshold and recovery state machine for an individual provider.
type CircuitBreaker struct {
	mu                   sync.RWMutex
	config               CircuitBreakerConfig
	state                CircuitState
	consecutiveFailures  int
	canarySuccesses      int
	totalRequests        int64
	totalFailures        int64
	averageLatencyMs     float64
	lastFailureTime      time.Time
	nextCanaryTime       time.Time
	rateLimitCount       int
}

// NewCircuitBreaker initializes a circuit breaker with defaults if unspecified.
func NewCircuitBreaker(cfg CircuitBreakerConfig) *CircuitBreaker {
	if cfg.FailureThreshold <= 0 {
		cfg.FailureThreshold = 3
	}
	if cfg.CooldownDuration <= 0 {
		cfg.CooldownDuration = 30 * time.Second
	}
	if cfg.CanarySuccessCount <= 0 {
		cfg.CanarySuccessCount = 2
	}
	if cfg.RequestTimeout <= 0 {
		cfg.RequestTimeout = 15 * time.Second
	}

	return &CircuitBreaker{
		config: cfg,
		state:  StateClosed,
	}
}

// CanExecute checks if requests are currently permitted to be dispatched to this provider.
func (cb *CircuitBreaker) CanExecute() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()

	switch cb.state {
	case StateClosed:
		return true

	case StateOpen:
		if now.After(cb.nextCanaryTime) {
			// Cooldown expired; transition to HalfOpen for probe verification
			cb.state = StateHalfOpen
			cb.canarySuccesses = 0
			return true
		}
		return false

	case StateHalfOpen:
		// Allow canary trial
		return true

	default:
		return true
	}
}

// RecordSuccess updates provider telemetry following a successful execution.
func (cb *CircuitBreaker) RecordSuccess(latencyMs float64) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.totalRequests++
	cb.consecutiveFailures = 0

	// Exponential Moving Average (EMA) for latency with 0.2 smoothing factor
	if cb.averageLatencyMs == 0 {
		cb.averageLatencyMs = latencyMs
	} else {
		cb.averageLatencyMs = (0.2 * latencyMs) + (0.8 * cb.averageLatencyMs)
	}

	if cb.state == StateHalfOpen {
		cb.canarySuccesses++
		if cb.canarySuccesses >= cb.config.CanarySuccessCount {
			// Breaker successfully recovered; close circuit
			cb.state = StateClosed
			cb.canarySuccesses = 0
		}
	}
}

// RecordFailure registers a failed request or rate limit event, tripping the breaker if threshold is exceeded.
func (cb *CircuitBreaker) RecordFailure(isRateLimit bool) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()
	cb.totalRequests++
	cb.totalFailures++
	cb.consecutiveFailures++
	cb.lastFailureTime = now

	if isRateLimit {
		cb.rateLimitCount++
	}

	if cb.state == StateHalfOpen {
		// Canary request failed; trip straight back to Open
		cb.state = StateOpen
		cb.nextCanaryTime = now.Add(cb.config.CooldownDuration)
		return
	}

	// Trip open if consecutive failures meet threshold or severe rate limiting
	if cb.consecutiveFailures >= cb.config.FailureThreshold || (isRateLimit && cb.consecutiveFailures >= 2) {
		cb.state = StateOpen
		cb.nextCanaryTime = now.Add(cb.config.CooldownDuration)
	}
}

// Snapshot returns current metrics and state for inspection.
func (cb *CircuitBreaker) Snapshot() (state CircuitState, fails int, totalReqs int64, totalFails int64, avgLat float64, lastFail time.Time, nextCanary time.Time, rateLimits int) {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return cb.state, cb.consecutiveFailures, cb.totalRequests, cb.totalFailures, cb.averageLatencyMs, cb.lastFailureTime, cb.nextCanaryTime, cb.rateLimitCount
}

// GetState returns the current operational state.
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}
