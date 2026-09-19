package codereview_test

import (
	"context"
	"testing"

	"github.com/lopor-ai/lopor/pkg/codereview"
)

func TestCodeReviewer_WithIssues(t *testing.T) {
	reviewer := codereview.NewCodeReviewer()

	code := `
package service

func ProcessData() {
	result, _ := executeQuery()
	// TODO: Add caching
	if result == nil {
		panic("database record not found")
	}
}
`

	req := codereview.ReviewRequest{
		FileName:    "service.go",
		Language:    "go",
		CodeContent: code,
	}

	report, err := reviewer.ReviewCode(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error during code review: %v", err)
	}

	if report.Verdict != codereview.VerdictRequestChanges {
		t.Errorf("expected REQUEST_CHANGES due to panic(), got %s", report.Verdict)
	}

	if report.CodeHealthScore >= 90 {
		t.Errorf("expected reduced health score, got %d", report.CodeHealthScore)
	}

	if len(report.Comments) < 3 {
		t.Errorf("expected at least 3 comments (blank error, TODO, panic), got %d", len(report.Comments))
	}
}

func TestCodeReviewer_CleanCode(t *testing.T) {
	reviewer := codereview.NewCodeReviewer()

	cleanCode := `
package service

import "fmt"

func ProcessData() error {
	result, err := executeQuery()
	if err != nil {
		return fmt.Errorf("failed query: %w", err)
	}
	return nil
}
`

	report, err := reviewer.ReviewCode(context.Background(), codereview.ReviewRequest{
		FileName:    "clean.go",
		Language:    "go",
		CodeContent: cleanCode,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.Verdict != codereview.VerdictApprove {
		t.Errorf("expected APPROVE on clean code, got %s", report.Verdict)
	}

	if report.CodeHealthScore != 100 {
		t.Errorf("expected 100 health score on clean code, got %d", report.CodeHealthScore)
	}
}
