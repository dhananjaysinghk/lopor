package dedup_test

import (
	"context"
	"testing"

	"github.com/lopor-ai/lopor/pkg/dedup"
)

func TestDedupEngine_ExactAndNearDuplicates(t *testing.T) {
	engine := dedup.NewDedupEngine()

	docs := []dedup.DocumentItem{
		{
			ID:      "doc-1",
			Title:   "PostgreSQL Vector Extension RFC",
			Content: "This document outlines the migration to PostgreSQL 16 pgvector for enterprise semantic similarity search and hybrid RAG retrieval.",
		},
		{
			ID:      "doc-2",
			Title:   "PostgreSQL Vector Extension RFC (Duplicate Copy)",
			Content: "This document outlines the migration to PostgreSQL 16 pgvector for enterprise semantic similarity search and hybrid RAG retrieval.",
		},
		{
			ID:      "doc-3",
			Title:   "PostgreSQL Vector Extension RFC v2 (Near Duplicate)",
			Content: "This document outlines the migration to PostgreSQL 16 pgvector for enterprise semantic similarity search and hybrid RAG retrieval with Redis caching.",
		},
		{
			ID:      "doc-4",
			Title:   "Corporate Travel Policy 2026",
			Content: "All employees traveling internationally must submit expense receipts within 30 business days via the finance portal.",
		},
	}

	req := dedup.DedupAnalysisRequest{
		Documents:           docs,
		SimilarityThreshold: 0.70,
	}

	report, err := engine.AnalyzeDuplicates(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error analyzing duplicates: %v", err)
	}

	if report.TotalAnalyzed != 4 {
		t.Errorf("expected 4 analyzed documents, got %d", report.TotalAnalyzed)
	}

	if report.DuplicateCount < 1 {
		t.Errorf("expected at least 1 duplicate detected, got %d", report.DuplicateCount)
	}

	if len(report.Clusters) == 0 {
		t.Errorf("expected at least 1 cluster formed")
	}

	foundExact := false
	for _, p := range report.DuplicatePairs {
		if p.Category == dedup.CategoryExact {
			foundExact = true
			if p.SimilarityScore != 1.0 {
				t.Errorf("expected 1.0 similarity for exact match, got %f", p.SimilarityScore)
			}
			break
		}
	}

	if !foundExact {
		t.Errorf("expected to find exact duplicate pair between doc-1 and doc-2")
	}

	if report.EstimatedTokenSavings <= 0 {
		t.Errorf("expected positive token savings, got %d", report.EstimatedTokenSavings)
	}
}

func TestDedupEngine_DistinctDocuments(t *testing.T) {
	engine := dedup.NewDedupEngine()

	docs := []dedup.DocumentItem{
		{
			ID:      "alpha",
			Title:   "Frontend React Guidelines",
			Content: "Use functional components with TypeScript and Tailwind CSS for rapid UI development.",
		},
		{
			ID:      "beta",
			Title:   "Kubernetes Deployment Manifests",
			Content: "Configure Helm charts with horizontal pod autoscaling and ingress controllers.",
		},
	}

	report, err := engine.AnalyzeDuplicates(context.Background(), dedup.DedupAnalysisRequest{
		Documents:           docs,
		SimilarityThreshold: 0.80,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.DuplicateCount != 0 {
		t.Errorf("expected 0 duplicates for distinct documents, got %d", report.DuplicateCount)
	}

	if len(report.Clusters) != 0 {
		t.Errorf("expected 0 clusters for distinct documents, got %d", len(report.Clusters))
	}
}
