package toolsim

import (
	"time"
)

// ToolCategory groups tools by functional domain.
type ToolCategory string

const (
	CategoryDatabase       ToolCategory = "DATABASE"
	CategoryPayment        ToolCategory = "PAYMENT"
	CategoryInfrastructure ToolCategory = "INFRASTRUCTURE"
	CategoryCommunication  ToolCategory = "COMMUNICATION"
	CategoryFilesystem     ToolCategory = "FILESYSTEM"
)

// RiskLevel rates the operational danger and irreversibility of a tool action.
type RiskLevel string

const (
	RiskLow      RiskLevel = "LOW"      // Read-only queries, data inspection
	RiskMedium   RiskLevel = "MEDIUM"   // Non-destructive writes, draft creation
	RiskHigh     RiskLevel = "HIGH"     // Financial charges, data mutations
	RiskCritical RiskLevel = "CRITICAL" // Resource teardown, table drops, production merges
)

// SideEffectType classifies the real-world consequence of a tool call.
type SideEffectType string

const (
	EffectCreate       SideEffectType = "RESOURCE_CREATE"
	EffectUpdate       SideEffectType = "RESOURCE_UPDATE"
	EffectDelete       SideEffectType = "RESOURCE_DELETE"
	EffectFinancial    SideEffectType = "FINANCIAL_CHARGE"
	EffectNotification SideEffectType = "EXTERNAL_DISPATCH"
)

// SideEffect describes a concrete simulated impact on external systems.
type SideEffect struct {
	Type           SideEffectType `json:"type"`
	TargetResource string         `json:"target_resource"`
	Description    string         `json:"description"`
	EstimatedCost  float64        `json:"estimated_cost,omitempty"` // In USD ($)
	IsIrreversible bool           `json:"is_reversible"`
}

// MockTool specifies a registered mock tool schema, risk level, and behavior.
type MockTool struct {
	Name                string                 `json:"name"`
	WorkspaceID         string                 `json:"workspace_id,omitempty"` // empty = system global tool
	Category            ToolCategory           `json:"category"`
	RiskLevel           RiskLevel              `json:"risk_level"`
	Description         string                 `json:"description"`
	ParametersSchema    map[string]string      `json:"parameters_schema"`
	DefaultMockResponse map[string]interface{} `json:"default_mock_response"`
	SimulatedLatencyMs  int64                  `json:"simulated_latency_ms"`
	FailureRatePercent  float64                `json:"failure_rate_percent,omitempty"` // Simulated chaos testing
}

// ToolInvocation logs an individual simulated execution step in the sandbox.
type ToolInvocation struct {
	InvocationID        string                 `json:"invocation_id"`
	ToolName            string                 `json:"tool_name"`
	Category            ToolCategory           `json:"category"`
	RiskLevel           RiskLevel              `json:"risk_level"`
	InputParameters     map[string]interface{} `json:"input_parameters"`
	SimulatedOutput     map[string]interface{} `json:"simulated_output"`
	SideEffects         []SideEffect           `json:"side_effects"`
	LatencyMs           int64                  `json:"latency_ms"`
	Status              string                 `json:"status"` // "SUCCESS" or "SIMULATED_FAILURE"
	RequiresHumanReview bool                   `json:"requires_human_review"`
	Timestamp           time.Time              `json:"timestamp"`
}

// SandboxSession tracks an active dry-run environment with virtual state.
type SandboxSession struct {
	ID                     string                 `json:"id"`
	WorkspaceID            string                 `json:"workspace_id"`
	Title                  string                 `json:"title"`
	Invocations            []ToolInvocation       `json:"invocations"`
	VirtualState           map[string]interface{} `json:"virtual_state"`
	BlastRadiusScore       float64                `json:"blast_radius_score"` // 0.0 - 100.0%
	TotalEstimatedCostUSD  float64                `json:"total_estimated_cost_usd"`
	HasCriticalSideEffects bool                   `json:"has_critical_side_effects"`
	CreatedAt              time.Time              `json:"created_at"`
	UpdatedAt              time.Time              `json:"updated_at"`
}

// PlanStep defines an individual tool call within a multi-step agent plan.
type PlanStep struct {
	ToolName   string                 `json:"tool_name"`
	Parameters map[string]interface{} `json:"parameters"`
}

// ExecutePlanRequest provides a sequence of planned tool calls for dry-run simulation.
type ExecutePlanRequest struct {
	WorkspaceID string     `json:"workspace_id"`
	Title       string     `json:"title"`
	Steps       []PlanStep `json:"steps"`
}

// SandboxReport provides a comprehensive risk assessment of simulated plan execution.
type SandboxReport struct {
	SessionID              string           `json:"session_id"`
	Title                  string           `json:"title"`
	TotalSteps             int              `json:"total_steps"`
	BlastRadiusScore       float64          `json:"blast_radius_score"`
	OverallRisk            RiskLevel        `json:"overall_risk"`
	RequiresHumanApproval  bool             `json:"requires_human_approval"`
	TotalEstimatedCostUSD  float64          `json:"total_estimated_cost_usd"`
	CreatedResources       []string         `json:"created_resources"`
	ModifiedResources      []string         `json:"modified_resources"`
	DeletedResources       []string         `json:"deleted_resources"`
	SideEffects            []SideEffect     `json:"side_effects"`
	Invocations            []ToolInvocation `json:"invocations"`
	SafetyRecommendations  []string         `json:"safety_recommendations"`
}
