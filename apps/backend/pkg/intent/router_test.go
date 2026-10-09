package intent_test

import (
	"context"
	"strings"
	"testing"

	"github.com/lopor-ai/lopor/pkg/intent"
)

func TestSemanticRouter_GetRoutes(t *testing.T) {
	router := intent.NewSemanticRouter()
	routes := router.GetRoutes()

	if len(routes) < 6 {
		t.Errorf("expected at least 6 core capability routes, got %d", len(routes))
	}
}

func TestSemanticRouter_ClassifyCodePrompt(t *testing.T) {
	router := intent.NewSemanticRouter()

	req := intent.ClassificationRequest{
		Prompt: "Please refactor function and fix bug in code snippet",
	}

	res, err := router.ClassifyPrompt(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error classifying prompt: %v", err)
	}

	if res.PrimaryIntent.Category != intent.IntentCodeDev {
		t.Errorf("expected CODE_ENGINEERING intent, got %s", res.PrimaryIntent.Category)
	}

	if !strings.Contains(res.PrimaryIntent.DestinationEndpoint, "sandbox") {
		t.Errorf("expected sandbox destination endpoint, got %s", res.PrimaryIntent.DestinationEndpoint)
	}

	if res.PrimaryIntent.Confidence <= 0.0 {
		t.Errorf("expected positive confidence score, got %f", res.PrimaryIntent.Confidence)
	}
}

func TestSemanticRouter_ClassifyRAGPrompt(t *testing.T) {
	router := intent.NewSemanticRouter()

	req := intent.ClassificationRequest{
		Prompt: "Search knowledge base to find in documents what our cloud policy says",
	}

	res, err := router.ClassifyPrompt(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.PrimaryIntent.Category != intent.IntentRAGSearch {
		t.Errorf("expected KNOWLEDGE_SEARCH intent, got %s", res.PrimaryIntent.Category)
	}
}

func TestSemanticRouter_ClassifySQLPrompt(t *testing.T) {
	router := intent.NewSemanticRouter()

	req := intent.ClassificationRequest{
		Prompt: "Generate sql query to calculate how many users signed up last month",
	}

	res, err := router.ClassifyPrompt(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.PrimaryIntent.Category != intent.IntentSQLSynth {
		t.Errorf("expected DATABASE_QUERY intent, got %s", res.PrimaryIntent.Category)
	}
}

func TestSemanticRouter_RegisterCustomRoute(t *testing.T) {
	router := intent.NewSemanticRouter()

	custom := intent.IntentRoute{
		ID:                  "route-custom-crm",
		Category:            "CRM_SYNC",
		Name:                "HubSpot CRM Synchronization",
		Description:         "Sync leads, update sales deals, check contact records in CRM.",
		Exemplars:           []string{"sync leads in hubspot", "update sales deal stage", "lookup contact in crm"},
		DestinationEndpoint: "/api/v1/workspaces/:wsId/integrations/crm/sync",
	}

	if err := router.RegisterCustomRoute(custom); err != nil {
		t.Fatalf("unexpected error registering custom route: %v", err)
	}

	res, err := router.ClassifyPrompt(context.Background(), intent.ClassificationRequest{
		Prompt: "Please sync leads in hubspot for current quarter",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.PrimaryIntent.IntentID != "route-custom-crm" {
		t.Errorf("expected route-custom-crm match, got %s", res.PrimaryIntent.IntentID)
	}
}
