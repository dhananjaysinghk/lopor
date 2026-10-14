package resilience

import (
	"errors"
	"math"
	"sort"
	"sync"
	"sync/atomic"
)

// ProviderPool manages a cluster of providers with individual circuit breakers and dynamic balancing.
type ProviderPool struct {
	mu        sync.RWMutex
	config    ProviderPoolConfig
	breakers  map[string]*CircuitBreaker
	rrCounter uint64
}

// NewProviderPool initializes a managed pool of providers with dedicated circuit breakers.
func NewProviderPool(cfg ProviderPoolConfig) *ProviderPool {
	breakers := make(map[string]*CircuitBreaker)
	for _, p := range cfg.Providers {
		breakers[p.ID] = NewCircuitBreaker(cfg.BreakerConfig)
	}

	return &ProviderPool{
		config:   cfg,
		breakers: breakers,
	}
}

// GetOrderedCandidates returns healthy providers ordered according to the configured balancing strategy.
func (p *ProviderPool) GetOrderedCandidates(strategy BalanceStrategy) ([]ProviderNode, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var healthy []ProviderNode

	for _, node := range p.config.Providers {
		cb, exists := p.breakers[node.ID]
		if !exists {
			continue
		}

		if cb.CanExecute() {
			state, fails, reqs, fCount, avgLat, lastFail, nextCan, rl := cb.Snapshot()
			node.State = state
			node.ConsecutiveFailures = fails
			node.TotalRequests = reqs
			node.TotalFailures = fCount
			node.AverageLatencyMs = math.Round(avgLat*100) / 100
			node.LastFailureTime = lastFail
			node.NextCanaryTime = nextCan
			node.RateLimitCount = rl
			healthy = append(healthy, node)
		}
	}

	if len(healthy) == 0 {
		return nil, errors.New("no healthy AI providers currently available in pool (all circuits tripped or cooling down)")
	}

	switch strategy {
	case StrategyLeastLatency:
		sort.Slice(healthy, func(i, j int) bool {
			// Untried nodes with 0 latency get slight priority for benchmarking
			if healthy[i].AverageLatencyMs == 0 {
				return true
			}
			if healthy[j].AverageLatencyMs == 0 {
				return false
			}
			return healthy[i].AverageLatencyMs < healthy[j].AverageLatencyMs
		})

	case StrategyCostOptimized:
		sort.Slice(healthy, func(i, j int) bool {
			return healthy[i].CostPer1kTokens < healthy[j].CostPer1kTokens
		})

	case StrategyWeightedPriority:
		sort.Slice(healthy, func(i, j int) bool {
			return healthy[i].Priority < healthy[j].Priority
		})

	case StrategyRoundRobin:
		idx := atomic.AddUint64(&p.rrCounter, 1)
		offset := int(idx % uint64(len(healthy)))
		rotated := make([]ProviderNode, len(healthy))
		copy(rotated, healthy[offset:])
		copy(rotated[len(healthy)-offset:], healthy[:offset])
		healthy = rotated
	}

	return healthy, nil
}

// RecordSuccess updates provider telemetry following a successful execution.
func (p *ProviderPool) RecordSuccess(providerID string, latencyMs float64) {
	p.mu.RLock()
	cb, exists := p.breakers[providerID]
	p.mu.RUnlock()

	if exists {
		cb.RecordSuccess(latencyMs)
	}
}

// RecordFailure notes a failed request for a provider.
func (p *ProviderPool) RecordFailure(providerID string, isRateLimit bool) {
	p.mu.RLock()
	cb, exists := p.breakers[providerID]
	p.mu.RUnlock()

	if exists {
		cb.RecordFailure(isRateLimit)
	}
}

// GetStatus compiles live telemetry across all pool members.
func (p *ProviderPool) GetStatus() ProviderPoolStatus {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var nodes []ProviderNode
	healthyCount := 0

	for _, node := range p.config.Providers {
		if cb, exists := p.breakers[node.ID]; exists {
			state, fails, reqs, fCount, avgLat, lastFail, nextCan, rl := cb.Snapshot()
			node.State = state
			node.ConsecutiveFailures = fails
			node.TotalRequests = reqs
			node.TotalFailures = fCount
			node.AverageLatencyMs = math.Round(avgLat*100) / 100
			node.LastFailureTime = lastFail
			node.NextCanaryTime = nextCan
			node.RateLimitCount = rl

			if cb.CanExecute() {
				healthyCount++
			}
		}
		nodes = append(nodes, node)
	}

	return ProviderPoolStatus{
		PoolID:       p.config.PoolID,
		WorkspaceID:  p.config.WorkspaceID,
		Strategy:     p.config.Strategy,
		Nodes:        nodes,
		TotalNodes:   len(nodes),
		HealthyNodes: healthyCount,
	}
}
