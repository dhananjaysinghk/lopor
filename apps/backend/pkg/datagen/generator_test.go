package datagen_test

import (
	"context"
	"strings"
	"testing"

	"github.com/lopor-ai/lopor/pkg/datagen"
)

func TestDatasetGenerator_GenerateDataset(t *testing.T) {
	generator := datagen.NewDatasetGenerator()

	docContent := `# PostgreSQL 16 pgvector Architecture
Lopor utilizes PostgreSQL 16 with pgvector for low-latency similarity search.
Documents are split into semantic chunks with 10% overlap using recursive text splitting.
Hybrid search merges BM25 full-text keyword ranking with vector distance using Reciprocal Rank Fusion.`

	req := datagen.DatasetGenerationRequest{
		DocumentID:       "doc-rag-arch",
		DocumentTitle:    "pgvector Architecture RFC",
		Content:          docContent,
		Count:            6,
		IncludeNegatives: true,
	}

	dataset, err := generator.GenerateDataset(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error generating dataset: %v", err)
	}

	if dataset.TotalItems != 6 {
		t.Errorf("expected 6 total items, got %d", dataset.TotalItems)
	}

	if dataset.DiversityScore <= 0.0 || dataset.DiversityScore > 1.0 {
		t.Errorf("expected diversity score between 0.0 and 1.0, got %f", dataset.DiversityScore)
	}

	// Verify presence of negative refusal test question
	if dataset.ItemDistribution[datagen.TypeNegative] == 0 {
		t.Errorf("expected at least 1 negative refusal question")
	}

	// Verify JSONL export format
	if !strings.Contains(dataset.JSONLExport, "pgvector") {
		t.Errorf("expected JSONL export to contain pgvector data")
	}
}

func TestDatasetGenerator_Formats(t *testing.T) {
	generator := datagen.NewDatasetGenerator()

	items := []datagen.DatasetItem{
		{
			ID:            "eval-01",
			Question:      "What is the maximum token limit?",
			Type:          datagen.TypeFactoid,
			ContextSource: "Tokens are limited to 8192.",
			GroundTruth:   "8192 tokens.",
			Difficulty:    "EASY",
			Confidence:    0.95,
		},
	}

	jsonl := generator.FormatAsJSONL(items)
	if !strings.Contains(jsonl, "\"id\":\"eval-01\"") {
		t.Errorf("expected jsonl output with id eval-01, got %s", jsonl)
	}

	csv := generator.FormatAsCSV(items)
	if !strings.Contains(csv, "\"eval-01\"") {
		t.Errorf("expected csv output with eval-01, got %s", csv)
	}
}
