package federation_test

import (
	"context"
	"strings"
	"testing"

	"github.com/lopor-ai/lopor/pkg/federation"
)

func TestFederationEngine_SetAndGetPolicy(t *testing.T) {
	engine := federation.NewFederationEngine()

	policy := federation.FederationPolicy{
		ID:                  "policy-eng-sales",
		SourceWorkspaceID:   "ws-engineering",
		TargetWorkspaceID:   "ws-sales",
		TargetWorkspaceName: "Enterprise Sales",
		TrustLevel:          federation.TrustReadOnly,
		AllowedTags:         []string{"public", "product-roadmap"},
		IsEnabled:           true,
	}

	if err := engine.SetPolicy(policy); err != nil {
		t.Fatalf("unexpected error setting policy: %v", err)
	}

	policies := engine.GetPolicies("ws-engineering")
	if len(policies) != 1 {
		t.Fatalf("expected 1 policy for ws-engineering, got %d", len(policies))
	}

	if policies[0].TargetWorkspaceID != "ws-sales" {
		t.Errorf("expected target workspace ws-sales, got %s", policies[0].TargetWorkspaceID)
	}
}

func TestFederationEngine_FederatedSearch(t *testing.T) {
	engine := federation.NewFederationEngine()

	// Policy 1: Read-Only
	_ = engine.SetPolicy(federation.FederationPolicy{
		ID:                  "pol-1",
		SourceWorkspaceID:   "ws-main",
		TargetWorkspaceID:   "ws-security",
		TargetWorkspaceName: "InfoSec & Compliance",
		TrustLevel:          federation.TrustReadOnly,
		IsEnabled:           true,
	})

	// Policy 2: Aggregate-Only
	_ = engine.SetPolicy(federation.FederationPolicy{
		ID:                  "pol-2",
		SourceWorkspaceID:   "ws-main",
		TargetWorkspaceID:   "ws-legal",
		TargetWorkspaceName: "Corporate Legal",
		TrustLevel:          federation.TrustAggregateOnly,
		IsEnabled:           true,
	})

	res, err := engine.FederatedSearch(context.Background(), "ws-main", "SOC 2 Type II audit", 5)
	if err != nil {
		t.Fatalf("unexpected error executing federated search: %v", err)
	}

	if res.FederatedWorkspacesQueried != 2 {
		t.Errorf("expected 2 federated workspaces queried, got %d", res.FederatedWorkspacesQueried)
	}

	if len(res.Results) != 2 {
		t.Errorf("expected 2 federated results, got %d", len(res.Results))
	}

	// Verify aggregate-only redaction
	foundRedacted := false
	for _, r := range res.Results {
		if r.WorkspaceID == "ws-legal" && r.IsRedacted {
			foundRedacted = true
			if !strings.Contains(r.Snippet, "Confidential Excerpt") {
				t.Errorf("expected redacted excerpt marker in legal snippet")
			}
			break
		}
	}

	if !foundRedacted {
		t.Errorf("expected legal workspace results to be marked as redacted")
	}
}

func TestFederationEngine_DiscoverEntities(t *testing.T) {
	engine := federation.NewFederationEngine()

	entities := []string{"AuthGateway", "PostgreSQL", "Redis"}
	links, err := engine.DiscoverCrossWorkspaceEntities(context.Background(), "ws-eng", "ws-ops", entities)
	if err != nil {
		t.Fatalf("unexpected error discovering entities: %v", err)
	}

	if len(links) != 3 {
		t.Errorf("expected 3 entity links, got %d", len(links))
	}

	for _, l := range links {
		if l.ConfidenceScore <= 0.0 {
			t.Errorf("expected positive confidence score, got %f", l.ConfidenceScore)
		}
	}
}
