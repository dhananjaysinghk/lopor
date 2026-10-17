package clarification

import (
	"math"
	"regexp"
	"strings"
	"time"
)

// Analyzer inspects user queries for semantic underspecification and synthesizes interactive clarification questions.
type Analyzer struct{}

// NewAnalyzer initializes the ambiguity analyzer.
func NewAnalyzer() *Analyzer {
	return &Analyzer{}
}

// Analyze evaluates the ambiguity entropy of a query.
func (a *Analyzer) Analyze(query string, context string) *DisambiguationAnalysis {
	qTrimmed := strings.TrimSpace(query)
	qLower := strings.ToLower(qTrimmed)
	words := strings.Fields(qLower)

	var factors []AmbiguityFactor
	var questions []ClarificationQuestion
	var assumptions []string

	// Check 1: Query Brevity
	if len(words) < 4 && len(words) > 0 {
		factors = append(factors, AmbiguityFactor{
			Type:        "QUERY_BREVITY",
			Description: "The query contains fewer than 4 words, leaving execution scope broad.",
			Severity:    0.25,
		})
	}

	// Check 2: Ambiguous Pronouns
	pronounRegex := regexp.MustCompile(`\b(it|them|this|that|the service|the app|the bug)\b`)
	if pronounRegex.MatchString(qLower) && !strings.Contains(strings.ToLower(context), "active_file") {
		factors = append(factors, AmbiguityFactor{
			Type:        "AMBIGUOUS_PRONOUN",
			Description: "The query contains referential pronouns without an explicit antecedent.",
			Severity:    0.35,
		})
		questions = append(questions, ClarificationQuestion{
			ID:           "q_target_entity",
			ParameterKey: "target_component",
			Question:     "Which component or file should be targeted?",
			Options: []ClarificationOption{
				{ID: "opt_active", Label: "Current Active Document", Value: "active_document", IsRecommended: true},
				{ID: "opt_backend", Label: "Backend Service API", Value: "backend_api", IsRecommended: false},
				{ID: "opt_frontend", Label: "Frontend User Interface", Value: "frontend_ui", IsRecommended: false},
			},
			AllowCustomInput: true,
			IsMultiSelect:    false,
		})
	}

	// Check 3: Deployment Verbs missing Environment
	deployRegex := regexp.MustCompile(`\b(deploy|release|ship|publish|rollout)\b`)
	if deployRegex.MatchString(qLower) {
		envRegex := regexp.MustCompile(`\b(staging|production|prod|preview|development|dev|canary)\b`)
		if !envRegex.MatchString(qLower) {
			factors = append(factors, AmbiguityFactor{
				Type:        "MISSING_ENVIRONMENT",
				Description: "Deployment requested without specifying destination target environment.",
				Severity:    0.45,
			})
			questions = append(questions, ClarificationQuestion{
				ID:           "q_target_env",
				ParameterKey: "target_environment",
				Question:     "Which target environment should receive this deployment?",
				Options: []ClarificationOption{
					{ID: "opt_staging", Label: "Staging (Recommended)", Value: "staging", Description: "Safe pre-production validation", IsRecommended: true},
					{ID: "opt_production", Label: "Production", Value: "production", Description: "Live customer-facing environment", IsRecommended: false},
					{ID: "opt_preview", Label: "Ephemeral Preview", Value: "preview", Description: "Temporary PR review branch", IsRecommended: false},
				},
				AllowCustomInput: false,
				IsMultiSelect:    false,
			})
		}
	}

	// Check 4: Export/Download verbs missing Format
	exportRegex := regexp.MustCompile(`\b(export|download|extract|generate report|dump)\b`)
	if exportRegex.MatchString(qLower) {
		formatRegex := regexp.MustCompile(`\b(csv|json|pdf|excel|xlsx|markdown|zip)\b`)
		if !formatRegex.MatchString(qLower) {
			factors = append(factors, AmbiguityFactor{
				Type:        "UNDERSPECIFIED_FORMAT",
				Description: "Data export requested without designating output document format.",
				Severity:    0.35,
			})
			questions = append(questions, ClarificationQuestion{
				ID:           "q_export_format",
				ParameterKey: "export_format",
				Question:     "Which file format would you prefer for the exported data?",
				Options: []ClarificationOption{
					{ID: "opt_csv", Label: "CSV (Recommended)", Value: "CSV", Description: "Standard tabular spreadsheet format", IsRecommended: true},
					{ID: "opt_json", Label: "JSON", Value: "JSON", Description: "Structured machine-readable objects", IsRecommended: false},
					{ID: "opt_pdf", Label: "PDF Report", Value: "PDF", Description: "Formatted printable executive document", IsRecommended: false},
					{ID: "opt_markdown", Label: "Markdown Table", Value: "Markdown", Description: "Plain text documentation table", IsRecommended: false},
				},
				AllowCustomInput: true,
				IsMultiSelect:    false,
			})
		}
	}

	// Check 5: Test Generation missing Test Suite Type
	testRegex := regexp.MustCompile(`\b(write tests|generate tests|create tests|unit tests|test suite)\b`)
	if testRegex.MatchString(qLower) {
		typeRegex := regexp.MustCompile(`\b(unit|integration|e2e|end-to-end|benchmark|fuzz)\b`)
		if !typeRegex.MatchString(qLower) {
			factors = append(factors, AmbiguityFactor{
				Type:        "UNDERSPECIFIED_TEST_TYPE",
				Description: "Testing requested without specifying test scope or harness type.",
				Severity:    0.30,
			})
			questions = append(questions, ClarificationQuestion{
				ID:           "q_test_scope",
				ParameterKey: "test_type",
				Question:     "What category of automated test suite should be generated?",
				Options: []ClarificationOption{
					{ID: "opt_unit", Label: "Unit Tests (Recommended)", Value: "unit", Description: "Fast, isolated component tests", IsRecommended: true},
					{ID: "opt_integration", Label: "Integration Tests", Value: "integration", Description: "Multi-service API and database interactions", IsRecommended: false},
					{ID: "opt_benchmark", Label: "Performance Benchmarks", Value: "benchmark", Description: "Throughput and latency profiling tests", IsRecommended: false},
				},
				AllowCustomInput: true,
				IsMultiSelect:    false,
			})
		}
	}

	// Compute overall ambiguity score
	totalScore := 0.0
	for _, f := range factors {
		totalScore += f.Severity
	}
	if totalScore > 1.0 {
		totalScore = 1.0
	}
	totalScore = math.Round(totalScore*100) / 100

	// Decision threshold evaluation
	decision := DecisionProceed
	if totalScore >= 0.40 {
		decision = DecisionNeedsClarification
	} else if totalScore > 0.15 {
		decision = DecisionProceedWithAssumptions
		for _, q := range questions {
			for _, opt := range q.Options {
				if opt.IsRecommended {
					assumptions = append(assumptions, "Assuming "+q.ParameterKey+" = "+opt.Value)
				}
			}
		}
		questions = nil // Clear questions since we are proceeding with assumptions
	}

	return &DisambiguationAnalysis{
		OriginalQuery:  query,
		AmbiguityScore: totalScore,
		Decision:       decision,
		Factors:        factors,
		Assumptions:    assumptions,
		Questions:      questions,
		AnalyzedAt:     time.Now().UTC(),
	}
}
