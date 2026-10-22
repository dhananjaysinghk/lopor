package selfheal

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"
)

// SelfHealingEngine coordinates runtime fault diagnosis, autonomous repairs, and resiliency telemetry.
type SelfHealingEngine struct {
	mu        sync.RWMutex
	telemetry map[string]*SelfHealingTelemetry
}

// NewEngine initializes a new instance of the SelfHealingEngine.
func NewEngine() *SelfHealingEngine {
	return &SelfHealingEngine{
		telemetry: make(map[string]*SelfHealingTelemetry),
	}
}

// Diagnose evaluates an execution fault to determine criticality and the optimal remediation strategy.
func (e *SelfHealingEngine) Diagnose(fault ExecutionFault) (RemediationStrategy, FaultSeverity) {
	category := fault.Category
	errLower := strings.ToLower(fault.ErrorMessage)

	if category == FaultContextLimitExceeded || strings.Contains(errLower, "context_length_exceeded") || strings.Contains(errLower, "maximum context length") {
		return StrategyContextPrune, SeverityCritical
	}

	if category == FaultUpstreamOutage || strings.Contains(errLower, "rate limit") || strings.Contains(errLower, "503") || strings.Contains(errLower, "timeout") {
		return StrategyFallbackProvider, SeverityCritical
	}

	if category == FaultSchemaValidation || strings.Contains(errLower, "json") || strings.Contains(errLower, "unmarshal") {
		// If it's a minor formatting or code fence issue, coercion is first line of defense
		return StrategyCoercion, SeverityModerate
	}

	if category == FaultCodeSyntaxError || strings.Contains(errLower, "syntaxerror") || strings.Contains(errLower, "parse error") {
		return StrategyDiagnosticPrompt, SeverityModerate
	}

	return StrategyDiagnosticPrompt, SeverityMinor
}

// Remediate executes an autonomous multi-phase recovery pipeline against an execution fault.
func (e *SelfHealingEngine) Remediate(ctx context.Context, fault ExecutionFault, policy *RemediationPolicy) (*RemediationResult, error) {
	start := time.Now()
	faultID := generateID("flt")

	if policy == nil {
		policy = &RemediationPolicy{
			MaxAttempts:            3,
			EnableCoercion:         true,
			EnableDiagnosticPrompt: true,
			EnableProviderFallback: true,
			FallbackProviders:      []string{"claude-3-5-sonnet", "gpt-4o", "deepseek-coder"},
			AllowDegradedResult:    true,
		}
	}

	primaryStrategy, severity := e.Diagnose(fault)
	fault.Severity = severity

	var attempts []RemediationAttempt
	var finalRepairedOutput string
	var finalEffectiveModel string
	var finalStatus RemediationStatus = StatusFailed
	var diagnosticNotes string

	// Phase 1: Automated Fast Schema Coercion (Sub-millisecond)
	if policy.EnableCoercion && (fault.Category == FaultSchemaValidation || primaryStrategy == StrategyCoercion) {
		t0 := time.Now()
		coerced, ok := CoerceJSON(fault.OffendingOutput)
		lat := time.Since(t0).Milliseconds()

		if ok {
			attempts = append(attempts, RemediationAttempt{
				AttemptNumber:  len(attempts) + 1,
				Strategy:       StrategyCoercion,
				Diagnosis:      "Malformed JSON repaired via regex boundary extraction and trailing comma normalization.",
				RepairedOutput: coerced,
				Success:        true,
				LatencyMs:      lat,
			})
			finalRepairedOutput = coerced
			finalStatus = StatusHealed
			diagnosticNotes = "Successfully healed invalid JSON structure without re-prompting model."
		} else {
			attempts = append(attempts, RemediationAttempt{
				AttemptNumber: len(attempts) + 1,
				Strategy:      StrategyCoercion,
				Diagnosis:     "Direct JSON coercion unsuccessful; output lacks valid JSON boundaries.",
				Success:       false,
				LatencyMs:     lat,
			})
		}
	}

	// Phase 2: Context Pruning for Token Limit Errors
	if finalStatus != StatusHealed && (fault.Category == FaultContextLimitExceeded || primaryStrategy == StrategyContextPrune) {
		t0 := time.Now()
		pruned := PruneContextForTokenLimit(fault.OriginalInput, 4000)
		lat := time.Since(t0).Milliseconds()

		attempts = append(attempts, RemediationAttempt{
			AttemptNumber:   len(attempts) + 1,
			Strategy:        StrategyContextPrune,
			Diagnosis:       "Input token limit breached. Compressed conversational context retaining critical directives.",
			RemediatedInput: pruned,
			Success:         true,
			LatencyMs:       lat,
		})
		finalStatus = StatusHealed
		diagnosticNotes = "Successfully pruned input context below target token budget."
	}

	// Phase 3: Provider Failover for Upstream Outages / Rate Limits
	if finalStatus != StatusHealed && policy.EnableProviderFallback && (fault.Category == FaultUpstreamOutage || primaryStrategy == StrategyFallbackProvider) {
		t0 := time.Now()
		fallbackModel := "claude-3-5-sonnet"
		if len(policy.FallbackProviders) > 0 {
			fallbackModel = policy.FallbackProviders[0]
		}
		lat := time.Since(t0).Milliseconds()

		attempts = append(attempts, RemediationAttempt{
			AttemptNumber: len(attempts) + 1,
			Strategy:      StrategyFallbackProvider,
			Diagnosis:     fmt.Sprintf("Upstream model outage detected. Rerouting execution pipeline to fallback model: %s.", fallbackModel),
			Success:       true,
			LatencyMs:     lat,
		})
		finalEffectiveModel = fallbackModel
		finalStatus = StatusHealed
		diagnosticNotes = fmt.Sprintf("Failed over to redundant model '%s'.", fallbackModel)
	}

	// Phase 4: Diagnostic Reprompt Formulation
	if finalStatus != StatusHealed && policy.EnableDiagnosticPrompt {
		t0 := time.Now()
		reprompt := BuildDiagnosticReprompt(fault)
		lat := time.Since(t0).Milliseconds()

		attempts = append(attempts, RemediationAttempt{
			AttemptNumber:   len(attempts) + 1,
			Strategy:        StrategyDiagnosticPrompt,
			Diagnosis:       "Synthesized compiler error diagnostic directive for model self-repair.",
			RemediatedInput: reprompt,
			Success:         true,
			LatencyMs:       lat,
		})
		finalStatus = StatusHealed
		diagnosticNotes = "Formulated precision diagnostic reprompt with schema constraints."
	}

	// Phase 5: Graceful Degradation Fallback
	if finalStatus != StatusHealed {
		if policy.AllowDegradedResult {
			finalStatus = StatusDegraded
			finalRepairedOutput = fault.OffendingOutput
			diagnosticNotes = "All automated healing strategies exhausted; returning degraded output with warning flags."
		} else {
			finalStatus = StatusFailed
			diagnosticNotes = "Remediation failed. Manual intervention required."
		}
	}

	totalLatency := time.Since(start).Milliseconds()

	result := &RemediationResult{
		FaultID:          faultID,
		Status:           finalStatus,
		PrimaryStrategy:  primaryStrategy,
		Attempts:         attempts,
		TotalLatencyMs:   totalLatency,
		RepairedOutput:   finalRepairedOutput,
		EffectiveModelID: finalEffectiveModel,
		DiagnosticNotes:  diagnosticNotes,
	}

	// Record Telemetry
	e.recordTelemetry(fault.WorkspaceID, fault.Category, primaryStrategy, finalStatus, totalLatency)

	return result, nil
}

// GetTelemetry retrieves aggregate self-healing performance metrics for the specified workspace.
func (e *SelfHealingEngine) GetTelemetry(workspaceID string) *SelfHealingTelemetry {
	e.mu.RLock()
	defer e.mu.RUnlock()

	telemetry, exists := e.telemetry[workspaceID]
	if !exists {
		return &SelfHealingTelemetry{
			WorkspaceID:          workspaceID,
			StrategyDistribution: make(map[RemediationStrategy]int),
			CategoryDistribution: make(map[FaultCategory]int),
		}
	}

	return telemetry
}

func (e *SelfHealingEngine) recordTelemetry(workspaceID string, category FaultCategory, strategy RemediationStrategy, status RemediationStatus, latency int64) {
	if workspaceID == "" {
		return
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	t, exists := e.telemetry[workspaceID]
	if !exists {
		t = &SelfHealingTelemetry{
			WorkspaceID:          workspaceID,
			StrategyDistribution: make(map[RemediationStrategy]int),
			CategoryDistribution: make(map[FaultCategory]int),
		}
		e.telemetry[workspaceID] = t
	}

	t.TotalFaultsDetected++
	t.CategoryDistribution[category]++
	t.StrategyDistribution[strategy]++

	switch status {
	case StatusHealed:
		t.TotalHealed++
	case StatusDegraded:
		t.TotalDegraded++
	case StatusFailed:
		t.TotalFailed++
	}

	if t.TotalFaultsDetected > 0 {
		t.HealingSuccessRate = float64(t.TotalHealed) / float64(t.TotalFaultsDetected) * 100.0
		// Running average latency
		t.AverageRecoveryMs = (t.AverageRecoveryMs*float64(t.TotalFaultsDetected-1) + float64(latency)) / float64(t.TotalFaultsDetected)
	}
}

func generateID(prefix string) string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%s_%s", prefix, hex.EncodeToString(b))
}
