package guardrails

import (
	"time"
)

// PolicyAction defines the enforcement action taken upon a guardrail violation.
type PolicyAction string

const (
	ActionPass      PolicyAction = "PASS"
	ActionBlock     PolicyAction = "BLOCK"
	ActionSanitize  PolicyAction = "SANITIZE"
	ActionFlagAndLog PolicyAction = "FLAG_AND_LOG"
)

// ViolationSeverity classifies the risk level of a detected violation.
type ViolationSeverity string

const (
	SeverityCritical ViolationSeverity = "CRITICAL"
	SeverityHigh     ViolationSeverity = "HIGH"
	SeverityMedium   ViolationSeverity = "MEDIUM"
	SeverityLow      ViolationSeverity = "LOW"
	SeverityNone     ViolationSeverity = "NONE"
)

// GuardrailCategory classifies the security and safety domain.
type GuardrailCategory string

const (
	CategoryPromptInjection    GuardrailCategory = "PROMPT_INJECTION"
	CategoryJailbreak          GuardrailCategory = "JAILBREAK"
	CategorySystemPromptLeak   GuardrailCategory = "SYSTEM_PROMPT_LEAK"
	CategorySecretExposure     GuardrailCategory = "SECRET_EXPOSURE"
	CategoryDestructiveCommand GuardrailCategory = "DESTRUCTIVE_COMMAND"
	CategoryToxicity           GuardrailCategory = "TOXICITY"
	CategoryTopicPolicy        GuardrailCategory = "TOPIC_POLICY"
)

// InspectionDirection specifies whether prompt input or LLM output is being evaluated.
type InspectionDirection string

const (
	DirectionInboundPrompt    InspectionDirection = "INBOUND_PROMPT"
	DirectionOutboundResponse InspectionDirection = "OUTBOUND_RESPONSE"
)

// Violation details a specific rule match or security hazard detected.
type Violation struct {
	Category    GuardrailCategory `json:"category"`
	Severity    ViolationSeverity `json:"severity"`
	RuleName    string            `json:"rule_name"`
	Description string            `json:"description"`
	MatchedText string            `json:"matched_text,omitempty"`
	Remediation string            `json:"remediation"`
}

// GuardrailRule defines an individual pattern or heuristic policy.
type GuardrailRule struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Category    GuardrailCategory   `json:"category"`
	Direction   InspectionDirection `json:"direction"`
	Regex       string              `json:"regex"`
	Severity    ViolationSeverity   `json:"severity"`
	Action      PolicyAction        `json:"action"`
	Description string              `json:"description"`
	Enabled     bool                `json:"enabled"`
}

// GuardrailProfile configures policy enforcement rules for a workspace.
type GuardrailProfile struct {
	WorkspaceID       string              `json:"workspace_id"`
	Name              string              `json:"name"`
	DefaultAction     PolicyAction        `json:"default_action"`
	BlockThreshold    ViolationSeverity   `json:"block_threshold"` // Violations at or above this severity block the request
	MaskSecrets       bool                `json:"mask_secrets"`
	EnabledCategories []GuardrailCategory `json:"enabled_categories"`
	CustomRules       []GuardrailRule     `json:"custom_rules,omitempty"`
	UpdatedAt         time.Time           `json:"updated_at"`
}

// InspectionRequest provides the payload to be verified by the firewall.
type InspectionRequest struct {
	WorkspaceID string              `json:"workspace_id"`
	Direction   InspectionDirection `json:"direction"`
	Content     string              `json:"content"`
	Context     string              `json:"context,omitempty"`
}

// InspectionResult conveys the firewall decision, sanitized content, and violation details.
type InspectionResult struct {
	Allowed          bool          `json:"allowed"`
	ActionTaken      PolicyAction  `json:"action_taken"`
	SanitizedContent string        `json:"sanitized_content"`
	OverallRiskScore float64       `json:"overall_risk_score"` // 0.0 to 1.0 scale
	Violations       []Violation   `json:"violations"`
	EvaluatedAt      time.Time     `json:"evaluated_at"`
	LatencyMs        int64         `json:"latency_ms"`
}
