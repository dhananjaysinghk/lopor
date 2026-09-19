package codereview

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type ReviewVerdict string

const (
	VerdictApprove        ReviewVerdict = "APPROVE"
	VerdictComment        ReviewVerdict = "COMMENT"
	VerdictRequestChanges ReviewVerdict = "REQUEST_CHANGES"
)

type IssueSeverity string

const (
	SeverityBlocker    IssueSeverity = "BLOCKER"
	SeverityWarning    IssueSeverity = "WARNING"
	SeveritySuggestion IssueSeverity = "SUGGESTION"
)

type ReviewComment struct {
	LineNumber  int           `json:"line_number"`
	Severity    IssueSeverity `json:"severity"`
	Category    string        `json:"category"` // "performance", "error_handling", "concurrency", "security", "style"
	Message     string        `json:"message"`
	Suggestion  string        `json:"suggestion,omitempty"`
}

type ReviewRequest struct {
	FileName    string `json:"file_name"`
	Language    string `json:"language"`
	CodeContent string `json:"code_content"`
	ContextGoal string `json:"context_goal"`
}

type ReviewReport struct {
	FileName        string          `json:"file_name"`
	Language        string          `json:"language"`
	Verdict         ReviewVerdict   `json:"verdict"`
	CodeHealthScore int             `json:"code_health_score"` // 0 - 100
	ExecutiveSummary string         `json:"executive_summary"`
	Comments        []ReviewComment `json:"comments"`
	TotalIssues     int             `json:"total_issues"`
	ReviewedAt      string          `json:"reviewed_at"`
}

type CodeReviewer struct{}

func NewCodeReviewer() *CodeReviewer {
	return &CodeReviewer{}
}

// ReviewCode performs automated static and heuristic code review.
func (cr *CodeReviewer) ReviewCode(ctx context.Context, req ReviewRequest) (*ReviewReport, error) {
	if strings.TrimSpace(req.CodeContent) == "" {
		return nil, fmt.Errorf("code content cannot be empty")
	}

	if req.FileName == "" {
		req.FileName = "source_code"
	}

	lines := strings.Split(req.CodeContent, "\n")
	var comments []ReviewComment

	blockerCount := 0
	warningCount := 0

	for idx, line := range lines {
		lineNum := idx + 1
		trimmed := strings.TrimSpace(line)

		// 1. Check for unhandled errors / discarded errors in Go
		if strings.Contains(trimmed, ", _ =") || strings.Contains(trimmed, ", _ :=") {
			comments = append(comments, ReviewComment{
				LineNumber: lineNum,
				Severity:   SeverityWarning,
				Category:   "error_handling",
				Message:    "Blank identifier error suppression detected. Errors should be inspected and handled.",
				Suggestion: "if err != nil { return err }",
			})
			warningCount++
		}

		// 2. Check for potential Goroutine leaks or unbounded routines
		if strings.HasPrefix(trimmed, "go func(") && !strings.Contains(trimmed, "context") {
			comments = append(comments, ReviewComment{
				LineNumber: lineNum,
				Severity:   SeverityWarning,
				Category:   "concurrency",
				Message:    "Spawning unmonitored goroutine without context cancellation or lifecycle management.",
				Suggestion: "Pass ctx context.Context to ensure goroutine terminates upon cancellation.",
			})
			warningCount++
		}

		// 3. Check for TODO or FIXME comments
		if strings.Contains(trimmed, "TODO") || strings.Contains(trimmed, "FIXME") {
			comments = append(comments, ReviewComment{
				LineNumber: lineNum,
				Severity:   SeveritySuggestion,
				Category:   "style",
				Message:    "Unresolved TODO or FIXME marker found.",
				Suggestion: "Resolve debt or link to a tracking issue prior to merging.",
			})
		}

		// 4. Check for panic calls
		if strings.Contains(trimmed, "panic(") {
			comments = append(comments, ReviewComment{
				LineNumber: lineNum,
				Severity:   SeverityBlocker,
				Category:   "error_handling",
				Message:    "Explicit panic invocation found in production code.",
				Suggestion: "Return a structured error rather than terminating process via panic.",
			})
			blockerCount++
		}
	}

	// Calculate score & verdict
	healthScore := 100 - (blockerCount * 25) - (warningCount * 10) - (len(comments) * 2)
	if healthScore < 0 {
		healthScore = 0
	}

	verdict := VerdictApprove
	summary := "Code looks clean, robust, and ready to merge."

	if blockerCount > 0 {
		verdict = VerdictRequestChanges
		summary = fmt.Sprintf("Changes requested: Found %d blocker(s) that must be addressed before merging.", blockerCount)
	} else if warningCount > 0 {
		verdict = VerdictComment
		summary = fmt.Sprintf("Review passed with %d advisory recommendation(s) for improved resilience.", warningCount)
	}

	return &ReviewReport{
		FileName:         req.FileName,
		Language:         req.Language,
		Verdict:          verdict,
		CodeHealthScore:  healthScore,
		ExecutiveSummary: summary,
		Comments:         comments,
		TotalIssues:      len(comments),
		ReviewedAt:       time.Now().Format(time.RFC3339),
	}, nil
}
