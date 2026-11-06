package intentengine

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

var (
	ErrEmptyPrompt = errors.New("prompt cannot be empty")
)

// IntentRouterEngine dynamically scores prompt complexity and routes queries to optimal LLMs.
type IntentRouterEngine struct {
	mu         sync.RWMutex
	candidates []ModelCandidate
	telemetry  map[string]*RouterTelemetry
}

// NewEngine initializes a new IntentRouterEngine populated with industry standard model candidates.
func NewEngine() *IntentRouterEngine {
	e := &IntentRouterEngine{
		candidates: getDefaultModelCandidates(),
		telemetry:  make(map[string]*RouterTelemetry),
	}
	return e
}

// Route analyzes the prompt, evaluates complexity, selects optimal candidate, and tracks cost avoidance.
func (e *IntentRouterEngine) Route(req RoutePromptRequest) (*RoutingDecision, error) {
	start := time.Now()
	trimmed := strings.TrimSpace(req.Prompt)
	if trimmed == "" {
		return nil, ErrEmptyPrompt
	}

	opt := req.OptimizeFor
	if opt == "" {
		opt = "BALANCED"
	}

	factors, domain := AnalyzeComplexity(trimmed)
	targetTier := SelectTier(factors, domain, opt, req.ForceTier)

	selected, fallback := e.selectModelForTier(targetTier, domain)

	// Approximate token counts
	words := len(strings.Fields(trimmed))
	estInTokens := int64(float64(words) * 1.33)
	estOutTokens := int64(float64(estInTokens) * 1.5)
	if estOutTokens < 100 {
		estOutTokens = 100
	}

	estCost := (float64(estInTokens)/1000000.0)*selected.CostPer1MInTokens +
		(float64(estOutTokens)/1000000.0)*selected.CostPer1MOutTokens

	// Flagship baseline cost (e.g. $5.00 in / $15.00 out)
	baselineCost := (float64(estInTokens)/1000000.0)*5.00 + (float64(estOutTokens)/1000000.0)*15.00
	potentialSavings := baselineCost - estCost
	if potentialSavings < 0 {
		potentialSavings = 0
	}

	decisionID := generateID("dec")
	routingLatencyMicros := time.Since(start).Microseconds()
	if routingLatencyMicros <= 0 {
		routingLatencyMicros = 1
	}

	explanation := fmt.Sprintf("Classified domain as '%s' with %.1f%% complexity score. Routed to '%s' (%s tier, ~%dms latency) optimizing for '%s'. Estimated savings vs flagship: $%.5f.",
		domain, factors.OverallComplexity, selected.ModelID, selected.Tier, selected.TypicalLatencyMs, opt, potentialSavings)

	decision := &RoutingDecision{
		DecisionID:           decisionID,
		WorkspaceID:          req.WorkspaceID,
		PrimaryDomain:        domain,
		Complexity:           factors,
		SelectedModel:        selected,
		FallbackModel:        fallback,
		ReasoningRequired:    targetTier == TierReasoning,
		EstimatedLatencyMs:   selected.TypicalLatencyMs,
		EstimatedCostUSD:     estCost,
		PotentialSavingsUSD:  potentialSavings,
		RoutingExplanation:   explanation,
		RoutingLatencyMicros: routingLatencyMicros,
		DecidedAt:            time.Now().UTC(),
	}

	// Record Telemetry
	e.recordTelemetry(req.WorkspaceID, targetTier, domain, potentialSavings, float64(routingLatencyMicros)/1000.0)

	return decision, nil
}

// GetCandidates returns the list of active model candidates available for routing.
func (e *IntentRouterEngine) GetCandidates() []ModelCandidate {
	e.mu.RLock()
	defer e.mu.RUnlock()
	res := make([]ModelCandidate, len(e.candidates))
	copy(res, e.candidates)
	return res
}

// GetTelemetry retrieves routing distribution and cumulative financial savings for a workspace.
func (e *IntentRouterEngine) GetTelemetry(workspaceID string) *RouterTelemetry {
	e.mu.RLock()
	defer e.mu.RUnlock()

	t, exists := e.telemetry[workspaceID]
	if !exists {
		return &RouterTelemetry{
			WorkspaceID:      workspaceID,
			TierDistribution: make(map[TargetTier]int64),
			DomainDistribution: make(map[IntentDomain]int64),
		}
	}
	return t
}

func (e *IntentRouterEngine) selectModelForTier(tier TargetTier, domain IntentDomain) (ModelCandidate, *ModelCandidate) {
	var matches []ModelCandidate
	for _, m := range e.candidates {
		if m.Tier == tier {
			matches = append(matches, m)
		}
	}

	if len(matches) == 0 {
		// Fallback to standard
		return e.candidates[1], nil
	}

	primary := matches[0]
	var fallback *ModelCandidate
	if len(matches) > 1 {
		fallback = &matches[1]
	} else if len(e.candidates) > 0 {
		fallback = &e.candidates[1] // Default standard fallback
	}

	return primary, fallback
}

func (e *IntentRouterEngine) recordTelemetry(workspaceID string, tier TargetTier, domain IntentDomain, savingsUSD float64, latencyMs float64) {
	if workspaceID == "" {
		return
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	t, exists := e.telemetry[workspaceID]
	if !exists {
		t = &RouterTelemetry{
			WorkspaceID:        workspaceID,
			TierDistribution:   make(map[TargetTier]int64),
			DomainDistribution: make(map[IntentDomain]int64),
		}
		e.telemetry[workspaceID] = t
	}

	t.TotalRoutedRequests++
	t.TierDistribution[tier]++
	t.DomainDistribution[domain]++
	t.CumulativeSavedUSD += savingsUSD

	if t.TotalRoutedRequests > 0 {
		t.AverageRouterLatencyMs = (t.AverageRouterLatencyMs*float64(t.TotalRoutedRequests-1) + latencyMs) / float64(t.TotalRoutedRequests)
	}
}

func getDefaultModelCandidates() []ModelCandidate {
	return []ModelCandidate{
		{
			ModelID:            "gemini-1.5-flash-8b",
			Provider:           "google",
			Tier:               TierUltraFast,
			CostPer1MInTokens:  0.0375,
			CostPer1MOutTokens: 0.15,
			TypicalLatencyMs:   95,
			SupportedDomains:   []IntentDomain{DomainFastFactual, DomainCreativeWriting},
			MaxContextWindow:   1000000,
		},
		{
			ModelID:            "gpt-4o",
			Provider:           "openai",
			Tier:               TierStandard,
			CostPer1MInTokens:  2.50,
			CostPer1MOutTokens: 10.00,
			TypicalLatencyMs:   450,
			SupportedDomains:   []IntentDomain{DomainCodeEngineering, DomainDataAnalytics, DomainFastFactual},
			MaxContextWindow:   128000,
		},
		{
			ModelID:            "deepseek-reasoner-r1",
			Provider:           "deepseek",
			Tier:               TierReasoning,
			CostPer1MInTokens:  0.55,
			CostPer1MOutTokens: 2.19,
			TypicalLatencyMs:   1800,
			SupportedDomains:   []IntentDomain{DomainComplexReasoning, DomainSecurityAudit, DomainCodeEngineering},
			MaxContextWindow:   64000,
		},
	}
}

func generateID(prefix string) string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%s_%s", prefix, hex.EncodeToString(b))
}
