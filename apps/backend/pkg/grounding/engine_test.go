package grounding

import (
	"context"
	"strings"
	"testing"
)

func TestFullGroundedResponseWithCitations(t *testing.T) {
	engine := NewEngine()
	ctx := context.Background()
	wsID := "ws_grounding_success"

	sources := []SourceEvidence{
		{
			ID:      "src_01",
			Title:   "HNSW Configuration Manual",
			Content: "For optimal vector index recall, PostgreSQL pgvector requires setting ef_search to at least 40 before querying large indexes.",
			URI:     "https://docs.lopor.ai/pgvector-hnsw",
		},
		{
			ID:      "src_02",
			Title:   "Memory Sizing Guide",
			Content: "PostgreSQL shared buffers should be configured to 25 percent of system RAM for production workloads.",
			URI:     "https://docs.lopor.ai/memory-sizing",
		},
	}

	response := "PostgreSQL pgvector requires setting ef_search to at least 40 before querying large indexes. PostgreSQL shared buffers should be configured to 25 percent of system RAM for production workloads."

	req := VerifyGroundingRequest{
		WorkspaceID:       wsID,
		GeneratedResponse: response,
		SourceEvidences:   sources,
		Strictness:        0.60,
	}

	report, err := engine.VerifyGrounding(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.TotalClaims != 2 {
		t.Fatalf("expected 2 atomic claims, got %d", report.TotalClaims)
	}
	if report.SupportedClaims != 2 {
		t.Errorf("expected 2 supported claims, got %d", report.SupportedClaims)
	}
	if report.GroundednessScore != 1.0 {
		t.Errorf("expected groundedness score 1.0, got %.2f", report.GroundednessScore)
	}
	if report.HallucinationRisk != RiskLow {
		t.Errorf("expected RiskLow, got %s", report.HallucinationRisk)
	}
	if len(report.Citations) != 2 {
		t.Errorf("expected 2 citation anchors, got %d", len(report.Citations))
	}
	if !strings.Contains(report.AttributedResponse, "[^1]") {
		t.Errorf("expected attributed response to contain [^1]")
	}
	if !strings.Contains(report.AttributedResponse, "### References & Citations") {
		t.Errorf("expected bibliography in attributed response")
	}
}

func TestContradictionDetection(t *testing.T) {
	engine := NewEngine()
	ctx := context.Background()
	wsID := "ws_grounding_contra"

	sources := []SourceEvidence{
		{
			ID:      "src_gpu",
			Title:   "pgvector GPU Capabilities",
			Content: "PostgreSQL pgvector does not support distributed GPU acceleration natively without specialized extensions.",
		},
	}

	// Generated statement inverts the negation (hallucinated contradiction)
	response := "PostgreSQL pgvector supports distributed GPU acceleration natively without specialized extensions."

	req := VerifyGroundingRequest{
		WorkspaceID:       wsID,
		GeneratedResponse: response,
		SourceEvidences:   sources,
	}

	report, err := engine.VerifyGrounding(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.ContradictedClaims == 0 {
		t.Errorf("expected at least 1 contradicted claim detected")
	}
	if report.HallucinationRisk != RiskCritical {
		t.Errorf("expected RiskCritical due to contradiction, got %s", report.HallucinationRisk)
	}
	if !strings.Contains(report.AttributedResponse, "CONTRADICTION") {
		t.Errorf("expected CONTRADICTION warning tag in attributed response")
	}
}

func TestUngroundedHallucinationDetection(t *testing.T) {
	engine := NewEngine()
	ctx := context.Background()
	wsID := "ws_grounding_hallucination"

	sources := []SourceEvidence{
		{
			ID:      "src_k8s",
			Title:   "Kubernetes Architecture",
			Content: "Kubernetes orchestrates containerized workloads across a cluster of worker nodes managed by a control plane.",
		},
	}

	// Hallucinated cooking statement completely ungrounded in Kubernetes source
	response := "Bake sourdough bread at 450 degrees Fahrenheit inside a preheated Dutch oven for twenty minutes."

	req := VerifyGroundingRequest{
		WorkspaceID:       wsID,
		GeneratedResponse: response,
		SourceEvidences:   sources,
	}

	report, err := engine.VerifyGrounding(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.UngroundedClaims == 0 {
		t.Errorf("expected ungrounded claim count > 0")
	}
	if report.GroundednessScore > 0.0 {
		t.Errorf("expected groundedness score 0.0, got %.2f", report.GroundednessScore)
	}
	if report.HallucinationRisk != RiskHigh {
		t.Errorf("expected RiskHigh, got %s", report.HallucinationRisk)
	}
}

func TestAutoRemediationOutput(t *testing.T) {
	engine := NewEngine()
	ctx := context.Background()
	wsID := "ws_grounding_remediate"

	sources := []SourceEvidence{
		{
			ID:      "src_redis",
			Title:   "Redis Architecture",
			Content: "Redis functions as an in-memory key-value data store supporting pub-sub messaging.",
		},
	}

	response := "Redis functions as an in-memory key-value data store supporting pub-sub messaging. Redis was invented by Isaac Newton during the classical mechanics revolution."

	req := VerifyGroundingRequest{
		WorkspaceID:       wsID,
		GeneratedResponse: response,
		SourceEvidences:   sources,
		AutoRemediate:     true,
	}

	report, err := engine.VerifyGrounding(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(report.RemediatedResponse, "[^1]") {
		t.Errorf("expected remediated response to retain supported claim citation")
	}
	if !strings.Contains(report.RemediatedResponse, "Unverified claim removed") {
		t.Errorf("expected remediated response to flag unverified statement")
	}
}

func TestGroundingMetricsTelemetry(t *testing.T) {
	engine := NewEngine()
	ctx := context.Background()
	wsID := "ws_grounding_metrics"

	sources := []SourceEvidence{
		{ID: "s1", Title: "Source 1", Content: "Docker containers isolate processes using Linux namespaces and cgroups."},
	}

	// Request 1: Supported
	_, _ = engine.VerifyGrounding(ctx, VerifyGroundingRequest{
		WorkspaceID:       wsID,
		GeneratedResponse: "Docker containers isolate processes using Linux namespaces and cgroups.",
		SourceEvidences:   sources,
	})

	// Request 2: Contradicted
	sourcesContra := []SourceEvidence{
		{ID: "s2", Title: "Source 2", Content: "Python is not compiled directly into bare-metal machine code."},
	}
	_, _ = engine.VerifyGrounding(ctx, VerifyGroundingRequest{
		WorkspaceID:       wsID,
		GeneratedResponse: "Python is compiled directly into bare-metal machine code.",
		SourceEvidences:   sourcesContra,
	})

	metrics := engine.GetMetrics(wsID)
	if metrics.TotalVerifications != 2 {
		t.Errorf("expected 2 verifications recorded, got %d", metrics.TotalVerifications)
	}
	if metrics.CriticalRiskIncidentsCount != 1 {
		t.Errorf("expected 1 critical risk incident recorded, got %d", metrics.CriticalRiskIncidentsCount)
	}
	if metrics.AverageGroundednessScore <= 0 {
		t.Errorf("expected positive average groundedness score")
	}
}

