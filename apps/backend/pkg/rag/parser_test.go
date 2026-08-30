package rag

import "testing"

func TestSmartParser(t *testing.T) {
	parser := NewSmartParser()
	sampleMarkdown := "# Architecture Overview\nLopor AI Workspace architecture.\n\n## Database Schema\nPostgreSQL pgvector HNSW index."

	parsed := parser.ParseDocumentText(sampleMarkdown, "architecture.md")

	if parsed.Format != "markdown" {
		t.Errorf("Expected format 'markdown', got '%s'", parsed.Format)
	}

	if len(parsed.Sections) < 2 {
		t.Errorf("Expected at least 2 sections, got %d", len(parsed.Sections))
	}
}
