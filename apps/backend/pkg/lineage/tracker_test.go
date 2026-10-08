package lineage_test

import (
	"testing"

	"github.com/lopor-ai/lopor/pkg/lineage"
)

func TestLineageTracker_RecordAndTrace(t *testing.T) {
	tracker := lineage.NewTracker()
	wsID := "ws-lineage-test"

	// 1. Source Document
	docNode := lineage.LineageNode{
		ID:   "doc-sec-policy",
		Type: lineage.NodeSourceDocument,
		Name: "Enterprise Cloud Security Policy 2026",
	}
	if err := tracker.RecordNode(wsID, docNode); err != nil {
		t.Fatalf("unexpected error recording doc node: %v", err)
	}

	// 2. Text Chunk
	chunkNode := lineage.LineageNode{
		ID:   "chunk-sec-01",
		Type: lineage.NodeTextChunk,
		Name: "Chunk 1: Encryption at Rest",
	}
	_ = tracker.RecordNode(wsID, chunkNode)

	// 3. Vector Embedding
	vectorNode := lineage.LineageNode{
		ID:   "vec-sec-01",
		Type: lineage.NodeVectorEmbedding,
		Name: "1536-dim Embedding Vector",
	}
	_ = tracker.RecordNode(wsID, vectorNode)

	// 4. Model Response
	respNode := lineage.LineageNode{
		ID:   "resp-agent-99",
		Type: lineage.NodeModelResponse,
		Name: "Generated Answer on Cloud Encryption Standards",
	}
	_ = tracker.RecordNode(wsID, respNode)

	// Connect causal edges: doc -> chunk -> vector -> response
	_ = tracker.RecordEdge(wsID, lineage.LineageEdge{
		SourceNodeID: "doc-sec-policy",
		TargetNodeID: "chunk-sec-01",
		Relation:     lineage.RelExtractedFrom,
	})
	_ = tracker.RecordEdge(wsID, lineage.LineageEdge{
		SourceNodeID: "chunk-sec-01",
		TargetNodeID: "vec-sec-01",
		Relation:     lineage.RelEmbeddedAs,
	})
	_ = tracker.RecordEdge(wsID, lineage.LineageEdge{
		SourceNodeID: "vec-sec-01",
		TargetNodeID: "resp-agent-99",
		Relation:     lineage.RelSynthesizedInto,
	})

	// Trace backwards from response
	graph, err := tracker.TraceLineage(wsID, "resp-agent-99")
	if err != nil {
		t.Fatalf("unexpected error tracing lineage: %v", err)
	}

	if graph.RootResponseID != "resp-agent-99" {
		t.Errorf("expected root response ID resp-agent-99, got %s", graph.RootResponseID)
	}

	if len(graph.Nodes) != 4 {
		t.Errorf("expected 4 connected lineage nodes, got %d", len(graph.Nodes))
	}

	if len(graph.Edges) != 3 {
		t.Errorf("expected 3 lineage edges, got %d", len(graph.Edges))
	}

	if graph.SourcesCount != 1 {
		t.Errorf("expected 1 primary source document, got %d", graph.SourcesCount)
	}
}

func TestLineageTracker_CertificateGeneration(t *testing.T) {
	tracker := lineage.NewTracker()
	wsID := "ws-cert-test"

	_ = tracker.RecordNode(wsID, lineage.LineageNode{
		ID:   "doc-fin-audit",
		Type: lineage.NodeSourceDocument,
		Name: "SOX Financial Compliance Audit",
	})
	_ = tracker.RecordNode(wsID, lineage.LineageNode{
		ID:   "resp-fin-summary",
		Type: lineage.NodeModelResponse,
		Name: "Financial Summary Output",
	})
	_ = tracker.RecordEdge(wsID, lineage.LineageEdge{
		SourceNodeID: "doc-fin-audit",
		TargetNodeID: "resp-fin-summary",
		Relation:     lineage.RelSynthesizedInto,
	})

	cert, err := tracker.GenerateProvenanceCertificate(wsID, "resp-fin-summary")
	if err != nil {
		t.Fatalf("unexpected error generating certificate: %v", err)
	}

	if cert.CertificateID == "" {
		t.Errorf("expected non-empty certificate ID")
	}

	if cert.AttestationHash == "" {
		t.Errorf("expected cryptographic attestation hash")
	}

	if len(cert.ComplianceStandards) == 0 {
		t.Errorf("expected compliance standards listed on certificate")
	}

	if len(cert.PrimarySources) != 1 || cert.PrimarySources[0] != "SOX Financial Compliance Audit" {
		t.Errorf("expected primary source recorded, got %v", cert.PrimarySources)
	}
}
