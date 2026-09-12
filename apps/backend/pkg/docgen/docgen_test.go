package docgen_test

import (
	"context"
	"strings"
	"testing"

	"github.com/lopor-ai/lopor/pkg/docgen"
)

func TestDocGenerator_GenerateDocs(t *testing.T) {
	gen := docgen.NewGenerator()

	sampleSource := `
package server

func RegisterRoutes(app *fiber.App) {
	app.Get("/api/v1/users", GetUsers)
	app.Post("/api/v1/users", CreateUser)
}
`

	req := docgen.DocRequest{
		Title:       "Users API Specification",
		SourceCode:  sampleSource,
		Language:    "go",
		Format:      docgen.FormatMarkdown,
		WithDiagram: true,
	}

	res, err := gen.GenerateDocs(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error generating docs: %v", err)
	}

	if res.Title != "Users API Specification" {
		t.Errorf("expected title 'Users API Specification', got %s", res.Title)
	}

	if !strings.Contains(res.MarkdownBody, "# Users API Specification") {
		t.Errorf("expected title header in markdown body")
	}

	if !strings.Contains(res.MarkdownBody, "```mermaid") {
		t.Errorf("expected mermaid diagram in markdown body")
	}

	if len(res.Endpoints) < 2 {
		t.Errorf("expected at least 2 endpoints extracted, got %d", len(res.Endpoints))
	}
}

func TestDocGenerator_EmptySource(t *testing.T) {
	gen := docgen.NewGenerator()
	_, err := gen.GenerateDocs(context.Background(), docgen.DocRequest{
		SourceCode: "",
	})
	if err == nil {
		t.Errorf("expected error on empty source code, got nil")
	}
}
