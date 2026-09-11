package secscan

import (
	"context"
	"regexp"
	"strings"
	"time"
)

type Severity string

const (
	SeverityCritical Severity = "CRITICAL"
	SeverityHigh     Severity = "HIGH"
	SeverityMedium   Severity = "MEDIUM"
	SeverityLow      Severity = "LOW"
)

type VulnerabilityItem struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Severity    Severity `json:"severity"`
	LineNumber  int      `json:"line_number"`
	Snippet     string   `json:"snippet"`
	Remediation string   `json:"remediation"`
}

type ScanRequest struct {
	FileName string `json:"file_name"`
	Content  string `json:"content"`
}

type ScanResult struct {
	FileName        string              `json:"file_name"`
	TotalIssues     int                 `json:"total_issues"`
	CriticalCount   int                 `json:"critical_count"`
	HighCount       int                 `json:"high_count"`
	MediumCount     int                 `json:"medium_count"`
	LowCount        int                 `json:"low_count"`
	Passed          bool                `json:"passed"`
	Vulnerabilities []VulnerabilityItem `json:"vulnerabilities"`
	Timestamp       string              `json:"timestamp"`
}

type SecurityRule struct {
	ID          string
	Title       string
	Description string
	Severity    Severity
	Pattern     *regexp.Regexp
	Remediation string
}

type Scanner struct {
	rules []SecurityRule
}

func NewScanner() *Scanner {
	rules := []SecurityRule{
		{
			ID:          "SEC-001",
			Title:       "Exposed OpenAI API Key",
			Description: "Hardcoded OpenAI API key detected in source code.",
			Severity:    SeverityCritical,
			Pattern:     regexp.MustCompile(`sk-[a-zA-Z0-9T3BlbkFJ]{20,}`),
			Remediation: "Store API keys in environment variables (.env) or secret managers.",
		},
		{
			ID:          "SEC-002",
			Title:       "Exposed AWS Access Key ID",
			Description: "Hardcoded AWS Access Key ID detected.",
			Severity:    SeverityCritical,
			Pattern:     regexp.MustCompile(`AKIA[0-9A-Z]{16}`),
			Remediation: "Use IAM Roles or load credentials via AWS SDK default credential provider.",
		},
		{
			ID:          "SEC-003",
			Title:       "Exposed GitHub Personal Access Token",
			Description: "GitHub personal access token or fine-grained token detected.",
			Severity:    SeverityCritical,
			Pattern:     regexp.MustCompile(`gh[pousr]_[A-Za-z0-9_]{36,}`),
			Remediation: "Revoke token immediately and use GitHub Actions Secrets or OAuth2.",
		},
		{
			ID:          "SEC-004",
			Title:       "RSA / SSH Private Key Block",
			Description: "Unencrypted private cryptographic key found in file.",
			Severity:    SeverityCritical,
			Pattern:     regexp.MustCompile(`-----BEGIN (RSA|EC|OPENSSH|PRIVATE) KEY-----`),
			Remediation: "Do not commit private key files to source control. Use SSH agent or vault.",
		},
		{
			ID:          "SEC-005",
			Title:       "Potential SQL Injection (Raw Concatenation)",
			Description: "Raw query string concatenation detected.",
			Severity:    SeverityHigh,
			Pattern:     regexp.MustCompile(`(?i)(SELECT|INSERT|UPDATE|DELETE).*\+.*\b(req|params|input|id|user)\b`),
			Remediation: "Use parameterized queries ($1, $2 or ?) to prevent SQL injection.",
		},
		{
			ID:          "SEC-006",
			Title:       "Hardcoded Secret / Password String",
			Description: "Hardcoded password or secret string variable assignment.",
			Severity:    SeverityMedium,
			Pattern:     regexp.MustCompile(`(?i)(password|secret|jwt_secret)\s*[:=]\s*["'][^"']{6,}["']`),
			Remediation: "Load secrets dynamically via environment variables.",
		},
	}

	return &Scanner{rules: rules}
}

// ScanCode inspects source content for security vulnerabilities and secret leaks.
func (s *Scanner) ScanCode(ctx context.Context, req ScanRequest) (*ScanResult, error) {
	lines := strings.Split(req.Content, "\n")
	var items []VulnerabilityItem

	critCount := 0
	highCount := 0
	medCount := 0
	lowCount := 0

	for lineIdx, line := range lines {
		lineNum := lineIdx + 1
		for _, rule := range s.rules {
			if rule.Pattern.MatchString(line) {
				snippet := strings.TrimSpace(line)
				if len(snippet) > 80 {
					snippet = snippet[:80] + "..."
				}

				item := VulnerabilityItem{
					ID:          rule.ID,
					Title:       rule.Title,
					Description: rule.Description,
					Severity:    rule.Severity,
					LineNumber:  lineNum,
					Snippet:     snippet,
					Remediation: rule.Remediation,
				}
				items = append(items, item)

				switch rule.Severity {
				case SeverityCritical:
					critCount++
				case SeverityHigh:
					highCount++
				case SeverityMedium:
					medCount++
				case SeverityLow:
					lowCount++
				}
			}
		}
	}

	passed := (critCount == 0 && highCount == 0)

	return &ScanResult{
		FileName:        req.FileName,
		TotalIssues:     len(items),
		CriticalCount:   critCount,
		HighCount:       highCount,
		MediumCount:     medCount,
		LowCount:        lowCount,
		Passed:          passed,
		Vulnerabilities: items,
		Timestamp:       time.Now().Format(time.RFC3339),
	}, nil
}
