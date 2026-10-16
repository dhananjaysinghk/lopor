package hitl

import (
	"time"
)

// ApprovalStatus represents the lifecycle state of a human intervention ticket.
type ApprovalStatus string

const (
	StatusPending   ApprovalStatus = "PENDING"
	StatusApproved  ApprovalStatus = "APPROVED"
	StatusRejected  ApprovalStatus = "REJECTED"
	StatusExpired   ApprovalStatus = "EXPIRED"
	StatusCancelled ApprovalStatus = "CANCELLED"
)

// ActionRiskLevel categorizes the potential impact of an autonomous action.
type ActionRiskLevel string

const (
	RiskLow      ActionRiskLevel = "LOW"
	RiskMedium   ActionRiskLevel = "MEDIUM"
	RiskHigh     ActionRiskLevel = "HIGH"
	RiskCritical ActionRiskLevel = "CRITICAL"
)

// ApprovalRequirement determines the human intervention quorum needed.
type ApprovalRequirement string

const (
	RequirementAutomaticPass   ApprovalRequirement = "AUTOMATIC_PASS"
	RequirementSingleSignature ApprovalRequirement = "SINGLE_SIGNATURE"
	RequirementDualSignature   ApprovalRequirement = "DUAL_SIGNATURE"
	RequirementAlwaysBlocked   ApprovalRequirement = "ALWAYS_BLOCKED"
)

// ReviewDecision denotes the human reviewer's verdict.
type ReviewDecision string

const (
	DecisionApprove ReviewDecision = "APPROVE"
	DecisionReject  ReviewDecision = "REJECT"
	DecisionCancel  ReviewDecision = "CANCEL"
)

// ReviewEvent preserves an individual human reviewer's audit signature.
type ReviewEvent struct {
	ReviewerID     string                 `json:"reviewer_id"`
	ReviewerName   string                 `json:"reviewer_name"`
	Decision       ReviewDecision         `json:"decision"`
	Feedback       string                 `json:"feedback,omitempty"`
	ModifiedParams map[string]interface{} `json:"modified_params,omitempty"`
	Timestamp      time.Time              `json:"timestamp"`
}

// ApprovalTicket represents an intercepted action awaiting human authorization.
type ApprovalTicket struct {
	ID                string                 `json:"id"`
	WorkspaceID       string                 `json:"workspace_id"`
	SessionID         string                 `json:"session_id"`
	AgentID           string                 `json:"agent_id"`
	ToolName          string                 `json:"tool_name"`
	RiskLevel         ActionRiskLevel        `json:"risk_level"`
	Requirement       ApprovalRequirement    `json:"requirement"`
	Status            ApprovalStatus         `json:"status"`
	Summary           string                 `json:"summary"`
	PayloadPreview    string                 `json:"payload_preview"`
	Parameters        map[string]interface{} `json:"parameters"`
	Reviews           []ReviewEvent          `json:"reviews"`
	RequiredApprovals int                    `json:"required_approvals"`
	CurrentApprovals  int                    `json:"current_approvals"`
	CreatedAt         time.Time              `json:"created_at"`
	ExpiresAt         time.Time              `json:"expires_at"`
}

// ActionPolicy defines matching rules for tool intervention gates.
type ActionPolicy struct {
	ToolPattern    string              `json:"tool_pattern"` // Regex or exact name
	DefaultRisk    ActionRiskLevel     `json:"default_risk"`
	Requirement    ApprovalRequirement `json:"requirement"`
	RequiredRole   string              `json:"required_role"` // "admin", "owner", "member"
	TimeoutMinutes int                 `json:"timeout_minutes"`
}

// WorkspaceHITLConfig contains the policy ensemble for a workspace.
type WorkspaceHITLConfig struct {
	WorkspaceID        string              `json:"workspace_id"`
	Policies           []ActionPolicy      `json:"policies"`
	DefaultRequirement ApprovalRequirement `json:"default_requirement"`
	AutoExpireMinutes  int                 `json:"auto_expire_minutes"`
	UpdatedAt          time.Time           `json:"updated_at"`
}

// InterceptRequest conveys the action parameters to evaluate against HITL gates.
type InterceptRequest struct {
	WorkspaceID    string                 `json:"workspace_id"`
	SessionID      string                 `json:"session_id"`
	AgentID        string                 `json:"agent_id"`
	ToolName       string                 `json:"tool_name"`
	Summary        string                 `json:"summary"`
	PayloadPreview string                 `json:"payload_preview"`
	Parameters     map[string]interface{} `json:"parameters"`
}

// ReviewSubmission represents a reviewer's decision on a pending ticket.
type ReviewSubmission struct {
	TicketID       string                 `json:"ticket_id"`
	ReviewerID     string                 `json:"reviewer_id"`
	ReviewerName   string                 `json:"reviewer_name"`
	Decision       ReviewDecision         `json:"decision"`
	Feedback       string                 `json:"feedback,omitempty"`
	ModifiedParams map[string]interface{} `json:"modified_params,omitempty"`
}
