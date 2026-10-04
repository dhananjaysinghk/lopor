package postmortem_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/lopor-ai/lopor/pkg/postmortem"
)

func TestPostmortemSynthesizer_GetTemplates(t *testing.T) {
	synthesizer := postmortem.NewPostmortemSynthesizer()
	templates := synthesizer.GetTemplates()

	if len(templates) < 3 {
		t.Errorf("expected at least 3 postmortem templates, got %d", len(templates))
	}

	foundSRE := false
	for _, tpl := range templates {
		if tpl.ID == "sre_standard" {
			foundSRE = true
			if len(tpl.Sections) == 0 {
				t.Errorf("expected SRE template to contain sections")
			}
			break
		}
	}

	if !foundSRE {
		t.Errorf("expected sre_standard template in catalog")
	}
}

func TestPostmortemSynthesizer_AnalyzeRCA(t *testing.T) {
	synthesizer := postmortem.NewPostmortemSynthesizer()

	logs := "FATAL: remaining connection slots are reserved for non-replication superuser connections (pgx pool exhausted)"
	summary := "PostgreSQL database became unresponsive during high-concurrency batch embeddings."

	rootCause, whys, err := synthesizer.AnalyzeRCA(context.Background(), logs, summary)
	if err != nil {
		t.Fatalf("unexpected error analyzing RCA: %v", err)
	}

	if !strings.Contains(strings.ToLower(rootCause), "connection pool") {
		t.Errorf("expected root cause to mention connection pool, got %s", rootCause)
	}

	if len(whys) != 5 {
		t.Errorf("expected 5 Whys steps, got %d", len(whys))
	}
}

func TestPostmortemSynthesizer_GeneratePostmortem(t *testing.T) {
	synthesizer := postmortem.NewPostmortemSynthesizer()

	start := time.Now().Add(-60 * time.Minute).Format(time.RFC3339)
	end := time.Now().Format(time.RFC3339)

	req := postmortem.PostmortemRequest{
		IncidentTitle: "Database Connection Pool Saturation Outage",
		Severity:      postmortem.Sev1,
		Service:       "apps/backend/pgvector",
		StartTime:     start,
		EndTime:       end,
		Summary:       "Spike in hybrid search queries caused total connection pool exhaustion.",
		LogsSnippet:   "pgxpool: failed to acquire connection: context deadline exceeded",
	}

	report, err := synthesizer.GeneratePostmortem(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error generating postmortem: %v", err)
	}

	if report.IncidentID == "" {
		t.Errorf("expected non-empty incident ID")
	}

	if report.DurationMinutes <= 0 {
		t.Errorf("expected positive duration minutes, got %d", report.DurationMinutes)
	}

	if len(report.ChronologicalTimeline) == 0 {
		t.Errorf("expected non-empty chronological timeline")
	}

	if len(report.ActionItems) < 3 {
		t.Errorf("expected at least 3 action items, got %d", len(report.ActionItems))
	}

	if !strings.Contains(report.MarkdownDocument, "# Postmortem:") {
		t.Errorf("expected markdown document to start with header")
	}
}
