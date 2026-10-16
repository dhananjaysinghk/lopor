package hitl

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sync"
	"time"

	"github.com/google/uuid"
)

// HITLEngine orchestrates human authorization gates and intervention workflows for autonomous agent actions.
type HITLEngine struct {
	mu        sync.RWMutex
	configs   map[string]WorkspaceHITLConfig
	tickets   map[string]*ApprovalTicket
	wsTickets map[string][]string // workspaceID -> list of ticket IDs
}

// NewHITLEngine initializes the human-in-the-loop intervention engine with enterprise defaults.
func NewHITLEngine() *HITLEngine {
	engine := &HITLEngine{
		configs:   make(map[string]WorkspaceHITLConfig),
		tickets:   make(map[string]*ApprovalTicket),
		wsTickets: make(map[string][]string),
	}
	engine.initDefaultGlobalConfig()
	return engine
}

func (e *HITLEngine) initDefaultGlobalConfig() {
	defaultPolicies := []ActionPolicy{
		// Safe Read-only Actions (Auto Pass)
		{
			ToolPattern:    `(?i)^(search_|read_|get_|list_|view_|summarize_|translate_)`,
			DefaultRisk:    RiskLow,
			Requirement:    RequirementAutomaticPass,
			RequiredRole:   "member",
			TimeoutMinutes: 0,
		},
		// Moderate Actions (Single Signature)
		{
			ToolPattern:    `(?i)^(code_execute|run_sandbox|create_pr|generate_diff|export_)`,
			DefaultRisk:    RiskMedium,
			Requirement:    RequirementSingleSignature,
			RequiredRole:   "member",
			TimeoutMinutes: 30,
		},
		// High-Stakes Side Effects (Dual Signature)
		{
			ToolPattern:    `(?i)^(execute_sql|deploy_production|delete_|purge_|send_email|modify_role)`,
			DefaultRisk:    RiskCritical,
			Requirement:    RequirementDualSignature,
			RequiredRole:   "admin",
			TimeoutMinutes: 60,
		},
		// Catastrophic Actions (Always Blocked)
		{
			ToolPattern:    `(?i)^(drop_database|truncate_database|wipe_host|destroy_cluster)`,
			DefaultRisk:    RiskCritical,
			Requirement:    RequirementAlwaysBlocked,
			RequiredRole:   "owner",
			TimeoutMinutes: 0,
		},
	}

	globalConfig := WorkspaceHITLConfig{
		WorkspaceID:        "global",
		Policies:           defaultPolicies,
		DefaultRequirement: RequirementSingleSignature,
		AutoExpireMinutes:  30,
		UpdatedAt:          time.Now().UTC(),
	}

	e.ConfigureWorkspace(globalConfig)
}

// ConfigureWorkspace sets or updates the HITL policy configuration for a workspace.
func (e *HITLEngine) ConfigureWorkspace(cfg WorkspaceHITLConfig) {
	e.mu.Lock()
	defer e.mu.Unlock()

	cfg.UpdatedAt = time.Now().UTC()
	if cfg.AutoExpireMinutes <= 0 {
		cfg.AutoExpireMinutes = 30
	}
	e.configs[cfg.WorkspaceID] = cfg
}

// GetWorkspaceConfig returns the active policy configuration for a workspace.
func (e *HITLEngine) GetWorkspaceConfig(workspaceID string) WorkspaceHITLConfig {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if cfg, exists := e.configs[workspaceID]; exists {
		return cfg
	}
	return e.configs["global"]
}

// EvaluateAndIntercept checks if a proposed tool execution requires human intervention.
func (e *HITLEngine) EvaluateAndIntercept(ctx context.Context, req InterceptRequest) (*ApprovalTicket, bool, error) {
	cfg := e.GetWorkspaceConfig(req.WorkspaceID)

	matchedPolicy := ActionPolicy{
		ToolPattern:    req.ToolName,
		DefaultRisk:    RiskMedium,
		Requirement:    cfg.DefaultRequirement,
		TimeoutMinutes: cfg.AutoExpireMinutes,
	}

	for _, p := range cfg.Policies {
		if matched, _ := regexp.MatchString(p.ToolPattern, req.ToolName); matched {
			matchedPolicy = p
			break
		}
	}

	if matchedPolicy.Requirement == RequirementAutomaticPass {
		return nil, false, nil
	}

	now := time.Now().UTC()
	timeoutMins := matchedPolicy.TimeoutMinutes
	if timeoutMins <= 0 {
		timeoutMins = cfg.AutoExpireMinutes
	}
	expiresAt := now.Add(time.Duration(timeoutMins) * time.Minute)

	if matchedPolicy.Requirement == RequirementAlwaysBlocked {
		ticket := &ApprovalTicket{
			ID:                "hitl_" + uuid.New().String()[:8],
			WorkspaceID:       req.WorkspaceID,
			SessionID:         req.SessionID,
			AgentID:           req.AgentID,
			ToolName:          req.ToolName,
			RiskLevel:         RiskCritical,
			Requirement:       RequirementAlwaysBlocked,
			Status:            StatusRejected,
			Summary:           fmt.Sprintf("Action '%s' was permanently blocked by enterprise security policy.", req.ToolName),
			PayloadPreview:    req.PayloadPreview,
			Parameters:        req.Parameters,
			CreatedAt:         now,
			ExpiresAt:         expiresAt,
			RequiredApprovals: 0,
		}
		e.saveTicket(ticket)
		return ticket, true, fmt.Errorf("action '%s' is strictly forbidden by enterprise HITL policy", req.ToolName)
	}

	reqApprovals := 1
	if matchedPolicy.Requirement == RequirementDualSignature {
		reqApprovals = 2
	}

	ticket := &ApprovalTicket{
		ID:                "hitl_" + uuid.New().String()[:8],
		WorkspaceID:       req.WorkspaceID,
		SessionID:         req.SessionID,
		AgentID:           req.AgentID,
		ToolName:          req.ToolName,
		RiskLevel:         matchedPolicy.DefaultRisk,
		Requirement:       matchedPolicy.Requirement,
		Status:            StatusPending,
		Summary:           req.Summary,
		PayloadPreview:    req.PayloadPreview,
		Parameters:        req.Parameters,
		Reviews:           nil,
		RequiredApprovals: reqApprovals,
		CurrentApprovals:  0,
		CreatedAt:         now,
		ExpiresAt:         expiresAt,
	}

	e.saveTicket(ticket)
	return ticket, true, nil
}

func (e *HITLEngine) saveTicket(ticket *ApprovalTicket) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.tickets[ticket.ID] = ticket
	wsID := ticket.WorkspaceID
	e.wsTickets[wsID] = append(e.wsTickets[wsID], ticket.ID)
}

// SubmitReview processes an authorized reviewer's decision on a pending ticket.
func (e *HITLEngine) SubmitReview(ctx context.Context, sub ReviewSubmission) (*ApprovalTicket, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	ticket, exists := e.tickets[sub.TicketID]
	if !exists {
		return nil, errors.New("approval ticket not found")
	}

	now := time.Now().UTC()
	if now.After(ticket.ExpiresAt) && ticket.Status == StatusPending {
		ticket.Status = StatusExpired
		return ticket, errors.New("approval ticket has expired")
	}

	if ticket.Status != StatusPending {
		return ticket, fmt.Errorf("ticket cannot be reviewed; current status is '%s'", ticket.Status)
	}

	// Prevent duplicate reviews by the same reviewer
	for _, rev := range ticket.Reviews {
		if rev.ReviewerID == sub.ReviewerID {
			return ticket, errors.New("reviewer has already cast a verdict on this ticket")
		}
	}

	event := ReviewEvent{
		ReviewerID:     sub.ReviewerID,
		ReviewerName:   sub.ReviewerName,
		Decision:       sub.Decision,
		Feedback:       sub.Feedback,
		ModifiedParams: sub.ModifiedParams,
		Timestamp:      now,
	}
	ticket.Reviews = append(ticket.Reviews, event)

	switch sub.Decision {
	case DecisionReject:
		ticket.Status = StatusRejected
		return ticket, nil

	case DecisionCancel:
		ticket.Status = StatusCancelled
		return ticket, nil

	case DecisionApprove:
		if len(sub.ModifiedParams) > 0 {
			ticket.Parameters = sub.ModifiedParams
		}
		ticket.CurrentApprovals++
		if ticket.CurrentApprovals >= ticket.RequiredApprovals {
			ticket.Status = StatusApproved
		}
		return ticket, nil

	default:
		return ticket, fmt.Errorf("invalid review decision '%s'", sub.Decision)
	}
}

// GetTicket retrieves an individual ticket by ID.
func (e *HITLEngine) GetTicket(ticketID string) (*ApprovalTicket, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	t, exists := e.tickets[ticketID]
	if !exists {
		return nil, errors.New("approval ticket not found")
	}

	// Check if pending ticket has passed expiry
	if t.Status == StatusPending && time.Now().UTC().After(t.ExpiresAt) {
		t.Status = StatusExpired
	}

	return t, nil
}

// ListTickets returns tickets for a workspace, optionally filtered by status.
func (e *HITLEngine) ListTickets(workspaceID string, statusFilter []ApprovalStatus) []ApprovalTicket {
	e.mu.RLock()
	defer e.mu.RUnlock()

	now := time.Now().UTC()
	ids := e.wsTickets[workspaceID]
	statusMap := make(map[ApprovalStatus]bool)
	for _, s := range statusFilter {
		statusMap[s] = true
	}

	var results []ApprovalTicket
	for _, id := range ids {
		t, ok := e.tickets[id]
		if !ok {
			continue
		}

		if t.Status == StatusPending && now.After(t.ExpiresAt) {
			t.Status = StatusExpired
		}

		if len(statusFilter) > 0 && !statusMap[t.Status] {
			continue
		}
		results = append(results, *t)
	}

	return results
}
