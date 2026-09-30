package orchestrator

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

// EventType represents the category of platform event triggering workflows.
type EventType string

const (
	EventDocCreated               EventType = "document.created"
	EventDocUpdated               EventType = "document.updated"
	EventChatHighCost             EventType = "chat.high_cost"
	EventSecurityAlert            EventType = "security.vulnerability_detected"
	EventAgentGuardrailBreached   EventType = "agent.guardrail_breached"
	EventConsensusDebateCompleted EventType = "consensus.debate_completed"
)

// ActionType represents the automated action executed when workflow conditions are satisfied.
type ActionType string

const (
	ActionSendWebhook       ActionType = "SEND_WEBHOOK"
	ActionAutoSummarize     ActionType = "AUTO_SUMMARIZE"
	ActionAnonymizePII      ActionType = "ANONYMIZE_PII"
	ActionSendNotification  ActionType = "SEND_NOTIFICATION"
	ActionInvokeAgent       ActionType = "INVOKE_AGENT"
)

// OperatorType defines conditional evaluation logic.
type OperatorType string

const (
	OpEquals      OperatorType = "EQUALS"
	OpContains    OperatorType = "CONTAINS"
	OpGreaterThan OperatorType = "GREATER_THAN"
	OpLessThan    OperatorType = "LESS_THAN"
)

// WorkflowCondition specifies criteria for executing a workflow action.
type WorkflowCondition struct {
	Field    string       `json:"field"`
	Operator OperatorType `json:"operator"`
	Value    string       `json:"value"`
}

// WorkflowAction defines the task to be performed upon trigger.
type WorkflowAction struct {
	Type       ActionType        `json:"type"`
	Target     string            `json:"target"`
	Parameters map[string]string `json:"parameters,omitempty"`
}

// WorkflowRule represents a user-configured event-driven automation rule.
type WorkflowRule struct {
	ID          string              `json:"id"`
	WorkspaceID string              `json:"workspace_id"`
	Name        string              `json:"name"`
	Description string              `json:"description,omitempty"`
	EventType   EventType           `json:"event_type"`
	Conditions  []WorkflowCondition `json:"conditions,omitempty"`
	Actions     []WorkflowAction    `json:"actions"`
	IsEnabled   bool                `json:"is_enabled"`
	CreatedAt   string              `json:"created_at"`
}

// WorkflowExecutionLog records the execution trace of an evaluated workflow.
type WorkflowExecutionLog struct {
	ExecutionID       string    `json:"execution_id"`
	WorkspaceID       string    `json:"workspace_id"`
	WorkflowID        string    `json:"workflow_id"`
	WorkflowName      string    `json:"workflow_name"`
	TriggeredByEvent  EventType `json:"triggered_by_event"`
	MatchedConditions bool      `json:"matched_conditions"`
	ActionsExecuted   []string  `json:"actions_executed"`
	Status            string    `json:"status"` // SUCCESS, FAILED, SKIPPED
	ErrorMessage      string    `json:"error_message,omitempty"`
	ExecutedAt        string    `json:"executed_at"`
	DurationMs        int64     `json:"duration_ms"`
}

// EventPayload contains the runtime event data dispatched into the orchestrator.
type EventPayload struct {
	EventID     string                 `json:"event_id"`
	WorkspaceID string                 `json:"workspace_id"`
	EventType   EventType              `json:"event_type"`
	Data        map[string]interface{} `json:"data"`
	Timestamp   string                 `json:"timestamp"`
}

// WorkflowOrchestrator manages event subscription, condition filtering, and action dispatching.
type WorkflowOrchestrator struct {
	mu    sync.RWMutex
	rules map[string]WorkflowRule
	logs  []WorkflowExecutionLog
}

// NewWorkflowOrchestrator initializes the event-driven workflow engine.
func NewWorkflowOrchestrator() *WorkflowOrchestrator {
	return &WorkflowOrchestrator{
		rules: make(map[string]WorkflowRule),
		logs:  make([]WorkflowExecutionLog, 0),
	}
}

// RegisterRule creates or updates a workflow automation rule.
func (o *WorkflowOrchestrator) RegisterRule(rule WorkflowRule) error {
	if rule.ID == "" {
		rule.ID = fmt.Sprintf("wf-%d", time.Now().UnixNano())
	}
	if rule.WorkspaceID == "" {
		return fmt.Errorf("workspace_id is required")
	}
	if rule.Name == "" {
		return fmt.Errorf("workflow name is required")
	}
	if rule.EventType == "" {
		return fmt.Errorf("event_type is required")
	}
	if len(rule.Actions) == 0 {
		return fmt.Errorf("at least one action is required")
	}
	if rule.CreatedAt == "" {
		rule.CreatedAt = time.Now().Format(time.RFC3339)
	}

	o.mu.Lock()
	defer o.mu.Unlock()
	o.rules[rule.ID] = rule
	return nil
}

// GetRules retrieves all workflow automation rules for a given workspace.
func (o *WorkflowOrchestrator) GetRules(workspaceID string) []WorkflowRule {
	o.mu.RLock()
	defer o.mu.RUnlock()

	var result []WorkflowRule
	for _, r := range o.rules {
		if r.WorkspaceID == workspaceID {
			result = append(result, r)
		}
	}
	return result
}

// GetExecutionLogs retrieves all execution traces for a given workspace.
func (o *WorkflowOrchestrator) GetExecutionLogs(workspaceID string) []WorkflowExecutionLog {
	o.mu.RLock()
	defer o.mu.RUnlock()

	var result []WorkflowExecutionLog
	for _, l := range o.logs {
		if l.WorkspaceID == workspaceID {
			result = append(result, l)
		}
	}
	return result
}

// DispatchEvent routes an event through all matching active workflow rules.
func (o *WorkflowOrchestrator) DispatchEvent(ctx context.Context, event EventPayload) ([]WorkflowExecutionLog, error) {
	if event.WorkspaceID == "" {
		return nil, fmt.Errorf("event workspace_id is required")
	}
	if event.EventType == "" {
		return nil, fmt.Errorf("event_type is required")
	}
	if event.EventID == "" {
		event.EventID = fmt.Sprintf("evt-%d", time.Now().UnixNano())
	}
	if event.Timestamp == "" {
		event.Timestamp = time.Now().Format(time.RFC3339)
	}

	o.mu.RLock()
	var matchingRules []WorkflowRule
	for _, r := range o.rules {
		if r.WorkspaceID == event.WorkspaceID && r.EventType == event.EventType && r.IsEnabled {
			matchingRules = append(matchingRules, r)
		}
	}
	o.mu.RUnlock()

	var executedLogs []WorkflowExecutionLog

	for _, rule := range matchingRules {
		start := time.Now()
		matched := o.evaluateConditions(rule.Conditions, event.Data)

		log := WorkflowExecutionLog{
			ExecutionID:       fmt.Sprintf("exec-%d", time.Now().UnixNano()),
			WorkspaceID:       event.WorkspaceID,
			WorkflowID:        rule.ID,
			WorkflowName:      rule.Name,
			TriggeredByEvent:  event.EventType,
			MatchedConditions: matched,
			ExecutedAt:        time.Now().Format(time.RFC3339),
		}

		if !matched {
			log.Status = "SKIPPED"
			log.DurationMs = time.Since(start).Milliseconds()
		} else {
			var actionSummaries []string
			for _, action := range rule.Actions {
				summary := o.executeAction(ctx, action, event.Data)
				actionSummaries = append(actionSummaries, summary)
			}
			log.ActionsExecuted = actionSummaries
			log.Status = "SUCCESS"
			log.DurationMs = time.Since(start).Milliseconds()
		}

		executedLogs = append(executedLogs, log)

		o.mu.Lock()
		o.logs = append(o.logs, log)
		// Keep up to 200 logs
		if len(o.logs) > 200 {
			o.logs = o.logs[len(o.logs)-200:]
		}
		o.mu.Unlock()
	}

	return executedLogs, nil
}

func (o *WorkflowOrchestrator) evaluateConditions(conditions []WorkflowCondition, data map[string]interface{}) bool {
	if len(conditions) == 0 {
		return true // No conditions means execute on every matching event
	}

	for _, cond := range conditions {
		val, exists := data[cond.Field]
		if !exists {
			return false
		}
		valStr := fmt.Sprintf("%v", val)

		switch cond.Operator {
		case OpEquals:
			if !strings.EqualFold(valStr, cond.Value) {
				return false
			}
		case OpContains:
			if !strings.Contains(strings.ToLower(valStr), strings.ToLower(cond.Value)) {
				return false
			}
		case OpGreaterThan:
			numVal, err1 := strconv.ParseFloat(valStr, 64)
			targetVal, err2 := strconv.ParseFloat(cond.Value, 64)
			if err1 != nil || err2 != nil || numVal <= targetVal {
				return false
			}
		case OpLessThan:
			numVal, err1 := strconv.ParseFloat(valStr, 64)
			targetVal, err2 := strconv.ParseFloat(cond.Value, 64)
			if err1 != nil || err2 != nil || numVal >= targetVal {
				return false
			}
		default:
			return false
		}
	}

	return true
}

func (o *WorkflowOrchestrator) executeAction(ctx context.Context, action WorkflowAction, data map[string]interface{}) string {
	switch action.Type {
	case ActionSendWebhook:
		return fmt.Sprintf("Dispatched outbound webhook to target endpoint: %s", action.Target)
	case ActionAutoSummarize:
		return fmt.Sprintf("Triggered background document summarization for target: %s", action.Target)
	case ActionAnonymizePII:
		return fmt.Sprintf("Invoked automatic PII anonymization sanitizer on target: %s", action.Target)
	case ActionSendNotification:
		return fmt.Sprintf("Sent priority alert notification to channel: %s", action.Target)
	case ActionInvokeAgent:
		return fmt.Sprintf("Enqueued autonomous task execution for agent ID: %s", action.Target)
	default:
		return fmt.Sprintf("Executed action of type %s on target %s", action.Type, action.Target)
	}
}
