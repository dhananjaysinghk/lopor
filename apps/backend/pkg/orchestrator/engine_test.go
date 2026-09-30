package orchestrator_test

import (
	"context"
	"testing"

	"github.com/lopor-ai/lopor/pkg/orchestrator"
)

func TestWorkflowOrchestrator_RuleExecution(t *testing.T) {
	engine := orchestrator.NewWorkflowOrchestrator()
	wsID := "ws-test-123"

	rule := orchestrator.WorkflowRule{
		ID:          "rule-sec-alert",
		WorkspaceID: wsID,
		Name:        "Critical Security Alert Dispatcher",
		EventType:   orchestrator.EventSecurityAlert,
		IsEnabled:   true,
		Conditions: []orchestrator.WorkflowCondition{
			{
				Field:    "severity",
				Operator: orchestrator.OpEquals,
				Value:    "CRITICAL",
			},
			{
				Field:    "vulnerabilities_count",
				Operator: orchestrator.OpGreaterThan,
				Value:    "0",
			},
		},
		Actions: []orchestrator.WorkflowAction{
			{
				Type:   orchestrator.ActionSendWebhook,
				Target: "https://security.example.com/alerts",
			},
			{
				Type:   orchestrator.ActionSendNotification,
				Target: "admin@lopor.ai",
			},
		},
	}

	if err := engine.RegisterRule(rule); err != nil {
		t.Fatalf("unexpected error registering rule: %v", err)
	}

	rules := engine.GetRules(wsID)
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule registered, got %d", len(rules))
	}

	// Dispatch matching event
	event := orchestrator.EventPayload{
		WorkspaceID: wsID,
		EventType:   orchestrator.EventSecurityAlert,
		Data: map[string]interface{}{
			"severity":              "CRITICAL",
			"vulnerabilities_count": 3,
			"source_file":           "apps/backend/auth.go",
		},
	}

	logs, err := engine.DispatchEvent(context.Background(), event)
	if err != nil {
		t.Fatalf("unexpected error dispatching event: %v", err)
	}

	if len(logs) != 1 {
		t.Fatalf("expected 1 execution log, got %d", len(logs))
	}

	if logs[0].Status != "SUCCESS" {
		t.Errorf("expected SUCCESS status, got %s", logs[0].Status)
	}

	if !logs[0].MatchedConditions {
		t.Errorf("expected matched conditions to be true")
	}

	if len(logs[0].ActionsExecuted) != 2 {
		t.Errorf("expected 2 actions executed, got %d", len(logs[0].ActionsExecuted))
	}
}

func TestWorkflowOrchestrator_ConditionMismatchSkipped(t *testing.T) {
	engine := orchestrator.NewWorkflowOrchestrator()
	wsID := "ws-test-456"

	rule := orchestrator.WorkflowRule{
		ID:          "rule-doc-auto-summary",
		WorkspaceID: wsID,
		Name:        "Large Document Auto Summarizer",
		EventType:   orchestrator.EventDocCreated,
		IsEnabled:   true,
		Conditions: []orchestrator.WorkflowCondition{
			{
				Field:    "word_count",
				Operator: orchestrator.OpGreaterThan,
				Value:    "1000",
			},
		},
		Actions: []orchestrator.WorkflowAction{
			{
				Type:   orchestrator.ActionAutoSummarize,
				Target: "document-summarizer-agent",
			},
		},
	}

	if err := engine.RegisterRule(rule); err != nil {
		t.Fatalf("unexpected error registering rule: %v", err)
	}

	// Dispatch event with word_count below threshold (500 <= 1000)
	event := orchestrator.EventPayload{
		WorkspaceID: wsID,
		EventType:   orchestrator.EventDocCreated,
		Data: map[string]interface{}{
			"document_id": "doc-99",
			"word_count":  500,
		},
	}

	logs, err := engine.DispatchEvent(context.Background(), event)
	if err != nil {
		t.Fatalf("unexpected error dispatching event: %v", err)
	}

	if len(logs) != 1 {
		t.Fatalf("expected 1 execution log, got %d", len(logs))
	}

	if logs[0].Status != "SKIPPED" {
		t.Errorf("expected SKIPPED status, got %s", logs[0].Status)
	}

	if logs[0].MatchedConditions {
		t.Errorf("expected matched conditions to be false")
	}
}
