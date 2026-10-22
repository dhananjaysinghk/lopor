package selfheal

import (
	"time"
)

// FaultCategory identifies the structural or operational root cause of an AI execution error.
type FaultCategory string

const (
	FaultSchemaValidation     FaultCategory = "SCHEMA_VALIDATION_ERROR"
	FaultToolExecutionFail    FaultCategory = "TOOL_EXECUTION_FAILURE"
	FaultCodeSyntaxError      FaultCategory = "CODE_SYNTAX_ERROR"
	FaultContextLimitExceeded FaultCategory = "CONTEXT_LIMIT_EXCEEDED"
	FaultHallucinatedTool     FaultCategory = "HALLUCINATED_TOOL_SIGNATURE"
	FaultUpstreamOutage       FaultCategory = "UPSTREAM_PROVIDER_OUTAGE"
)

// FaultSeverity rates the urgency and operational impact of the execution failure.
type FaultSeverity string

const (
	SeverityMinor    FaultSeverity = "MINOR"
	SeverityModerate FaultSeverity = "MODERATE"
	SeverityCritical FaultSeverity = "CRITICAL"
)

// RemediationStrategy represents the recovery tactic applied to heal the execution.
type RemediationStrategy string

const (
	StrategyCoercion         RemediationStrategy = "SCHEMA_COERCION"
	StrategyDiagnosticPrompt RemediationStrategy = "DIAGNOSTIC_REPROMPT"
	StrategyFallbackProvider RemediationStrategy = "FALLBACK_PROVIDER_SWITCH"
	StrategyContextPrune     RemediationStrategy = "CONTEXT_PRUNING"
	StrategyDegradedGraceful RemediationStrategy = "GRACEFUL_DEGRADATION"
)

// RemediationStatus captures the final outcome of an attempted recovery.
type RemediationStatus string

const (
	StatusHealed   RemediationStatus = "HEALED"
	StatusDegraded RemediationStatus = "DEGRADED"
	StatusFailed   RemediationStatus = "FAILED"
)

// ExecutionFault captures runtime failure diagnostics from an agent or pipeline step.
type ExecutionFault struct {
	WorkspaceID     string            `json:"workspace_id"`
	WorkflowID      string            `json:"workflow_id,omitempty"`
	StepID          string            `json:"step_id,omitempty"`
	Category        FaultCategory     `json:"category"`
	Severity        FaultSeverity     `json:"severity"`
	OriginalInput   string            `json:"original_input"`
	OffendingOutput string            `json:"offending_output"`
	ErrorMessage    string            `json:"error_message"`
	TargetSchema    string            `json:"target_schema,omitempty"`
	Metadata        map[string]string `json:"metadata,omitempty"`
	Timestamp       time.Time         `json:"timestamp"`
}

// RemediationPolicy configures operational boundaries and retry budgets for self-healing.
type RemediationPolicy struct {
	MaxAttempts            int      `json:"max_attempts"`
	EnableCoercion         bool     `json:"enable_coercion"`
	EnableDiagnosticPrompt bool     `json:"enable_diagnostic_prompt"`
	EnableProviderFallback bool     `json:"enable_provider_fallback"`
	FallbackProviders      []string `json:"fallback_providers,omitempty"`
	AllowDegradedResult    bool     `json:"allow_degraded_result"`
}

// RemediationAttempt logs a single intervention step within a healing cycle.
type RemediationAttempt struct {
	AttemptNumber   int                 `json:"attempt_number"`
	Strategy        RemediationStrategy `json:"strategy"`
	Diagnosis       string              `json:"diagnosis"`
	RemediatedInput string              `json:"remediated_input,omitempty"`
	RepairedOutput  string              `json:"repaired_output,omitempty"`
	Success         bool                `json:"success"`
	LatencyMs       int64               `json:"latency_ms"`
}

// RemediationResult details the cumulative outcome of self-healing.
type RemediationResult struct {
	FaultID          string               `json:"fault_id"`
	Status           RemediationStatus    `json:"status"`
	PrimaryStrategy  RemediationStrategy  `json:"primary_strategy"`
	Attempts         []RemediationAttempt `json:"attempts"`
	TotalLatencyMs   int64                `json:"total_latency_ms"`
	RepairedOutput   string               `json:"repaired_output,omitempty"`
	EffectiveModelID string               `json:"effective_model_id,omitempty"`
	DiagnosticNotes  string               `json:"diagnostic_notes"`
}

// SelfHealingTelemetry aggregates workspace-wide resilience and recovery metrics.
type SelfHealingTelemetry struct {
	WorkspaceID          string                     `json:"workspace_id"`
	TotalFaultsDetected  int64                      `json:"total_faults_detected"`
	TotalHealed          int64                      `json:"total_healed"`
	TotalDegraded        int64                      `json:"total_degraded"`
	TotalFailed          int64                      `json:"total_failed"`
	HealingSuccessRate   float64                    `json:"healing_success_rate"` // 0.0 - 100.0%
	StrategyDistribution map[RemediationStrategy]int `json:"strategy_distribution"`
	CategoryDistribution map[FaultCategory]int      `json:"category_distribution"`
	AverageRecoveryMs    float64                    `json:"average_recovery_ms"`
}
