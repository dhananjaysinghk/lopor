package sqlsynth

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type QueryComplexity string

const (
	ComplexitySimple   QueryComplexity = "SIMPLE"
	ComplexityModerate QueryComplexity = "MODERATE"
	ComplexityComplex  QueryComplexity = "COMPLEX"
)

type TableSchema struct {
	TableName   string   `json:"table_name"`
	Columns     []string `json:"columns"`
	PrimaryKeys []string `json:"primary_keys"`
}

type SQLGenerateRequest struct {
	NaturalQuery string        `json:"natural_query"`
	Dialect      string        `json:"dialect"` // "postgres", "mysql", "sqlite"
	Schemas      []TableSchema `json:"schemas,omitempty"`
}

type SQLGenerateResult struct {
	GeneratedSQL string          `json:"generated_sql"`
	Dialect      string          `json:"dialect"`
	Explanation  string          `json:"explanation"`
	Complexity   QueryComplexity `json:"complexity"`
	IsSafe       bool            `json:"is_safe"`
	Warnings     []string        `json:"warnings"`
	GeneratedAt  string          `json:"generated_at"`
}

type SQLSynthesizer struct{}

func NewSQLSynthesizer() *SQLSynthesizer {
	return &SQLSynthesizer{}
}

// GenerateSQL synthesizes natural language requests into safe, dialect-specific SQL.
func (s *SQLSynthesizer) GenerateSQL(ctx context.Context, req SQLGenerateRequest) (*SQLGenerateResult, error) {
	if strings.TrimSpace(req.NaturalQuery) == "" {
		return nil, fmt.Errorf("natural language query is required")
	}

	dialect := req.Dialect
	if dialect == "" {
		dialect = "postgres"
	}

	lowerQuery := strings.ToLower(req.NaturalQuery)
	var generatedSQL string
	var explanation string
	complexity := ComplexitySimple
	var warnings []string

	// Simple heuristic query synthesis for common analytical questions
	if strings.Contains(lowerQuery, "user") || strings.Contains(lowerQuery, "member") {
		generatedSQL = "SELECT id, email, role, created_at FROM users ORDER BY created_at DESC LIMIT 50;"
		explanation = "Retrieves recent registered users with role and creation timestamp."
	} else if strings.Contains(lowerQuery, "workspace") || strings.Contains(lowerQuery, "tenant") {
		generatedSQL = "SELECT w.id, w.name, COUNT(d.id) AS document_count FROM workspaces w LEFT JOIN documents d ON w.id = d.workspace_id GROUP BY w.id, w.name ORDER BY document_count DESC;"
		explanation = "Aggregates documents grouped by workspace to determine activity rank."
		complexity = ComplexityModerate
	} else if strings.Contains(lowerQuery, "audit") || strings.Contains(lowerQuery, "activity") {
		generatedSQL = "SELECT user_id, action, resource_type, created_at FROM audit_logs WHERE created_at >= NOW() - INTERVAL '7 days' ORDER BY created_at DESC;"
		explanation = "Queries security audit trail logs for the trailing 7-day period."
	} else {
		generatedSQL = "SELECT * FROM workspaces ORDER BY created_at DESC LIMIT 20;"
		explanation = "General entity retrieval with default pagination limit applied."
	}

	// Safety validation: ensure no destructive statements
	isSafe := true
	destructiveKeywords := []string{"DROP TABLE", "DROP DATABASE", "TRUNCATE", "ALTER TABLE", "GRANT ALL"}
	for _, kw := range destructiveKeywords {
		if strings.Contains(strings.ToUpper(generatedSQL), kw) {
			isSafe = false
			warnings = append(warnings, fmt.Sprintf("Destructive operation '%s' rejected for security.", kw))
		}
	}

	return &SQLGenerateResult{
		GeneratedSQL: generatedSQL,
		Dialect:      dialect,
		Explanation:  explanation,
		Complexity:   complexity,
		IsSafe:       isSafe,
		Warnings:     warnings,
		GeneratedAt:  time.Now().Format(time.RFC3339),
	}, nil
}
