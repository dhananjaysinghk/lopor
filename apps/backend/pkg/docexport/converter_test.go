package docexport_test

import (
	"context"
	"strings"
	"testing"

	"github.com/lopor-ai/lopor/pkg/docexport"
)

func TestDocumentConverter_ConvertHTML(t *testing.T) {
	converter := docexport.NewDocumentConverter()

	req := docexport.ExportRequest{
		DocumentTitle: "Q3 Engineering Roadmap",
		Content:       "# Overview\n\n- Build high-scale RAG\n- Multi-LLM Routing",
		Format:        docexport.FormatHTML,
		Author:        "Lead Architect",
		IncludeTOC:    true,
		Theme:         "dark",
	}

	res, err := converter.ConvertDocument(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error converting document: %v", err)
	}

	if res.FileName != "q3_engineering_roadmap.html" {
		t.Errorf("expected filename q3_engineering_roadmap.html, got %s", res.FileName)
	}

	if !strings.Contains(string(res.Data), "<!DOCTYPE html>") {
		t.Errorf("expected DOCTYPE html in converted data")
	}

	if !strings.Contains(string(res.Data), "Lead Architect") {
		t.Errorf("expected author name in rendered HTML")
	}
}

func TestDocumentConverter_ConvertMarkdown(t *testing.T) {
	converter := docexport.NewDocumentConverter()

	req := docexport.ExportRequest{
		DocumentTitle: "Release Notes",
		Content:       "## Features\n\n- Real-time collaboration",
		Format:        docexport.FormatMarkdown,
	}

	res, err := converter.ConvertDocument(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error converting to markdown: %v", err)
	}

	if res.FileName != "release_notes.md" {
		t.Errorf("expected filename release_notes.md, got %s", res.FileName)
	}

	if string(res.Data) != req.Content {
		t.Errorf("expected exact markdown content match")
	}
}

func TestDocumentConverter_EmptyContent(t *testing.T) {
	converter := docexport.NewDocumentConverter()

	_, err := converter.ConvertDocument(context.Background(), docexport.ExportRequest{
		Content: "",
	})
	if err == nil {
		t.Errorf("expected error on empty content, got nil")
	}
}
