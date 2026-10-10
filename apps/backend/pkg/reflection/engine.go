package reflection

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"
)

// CritiqueDimension categorizes the evaluation axis for model reflection.
type CritiqueDimension string

const (
	DimFaithfulness       CritiqueDimension = "FAITHFULNESS_HALLUCINATION"
	DimCodeCorrectness    CritiqueDimension = "SYNTAX_LOGICAL_CORRECTNESS"
	DimConstraintAdherence CritiqueDimension = "CONSTRAINT_ADHERENCE"
	DimCompleteness       CritiqueDimension = "EDGE_CASE_COMPLETENESS"
)

// CritiqueSeverity rates the criticality of a detected defect.
type CritiqueSeverity string

const (
	SevInfo     CritiqueSeverity = "INFO"
	SevWarning  CritiqueSeverity = "WARNING"
	SevCritical CritiqueSeverity = "CRITICAL"
)

// CritiqueFinding represents an isolated defect identified during self-reflection.
type CritiqueFinding struct {
	Dimension    CritiqueDimension `json:"dimension"`
	Severity     CritiqueSeverity  `json:"severity"`
	Issue        string            `json:"issue"`
	Explanation  string            `json:"explanation"`
	SuggestedFix string            `json:"suggested_fix"`
}

// ReflectionRequest defines parameters for reflective critique and self-correction.
type ReflectionRequest struct {
	Prompt           string  `json:"prompt"`
	CandidateOutput  string  `json:"candidate_output"`
	Context          string  `json:"context,omitempty"`
	OutputFormat     string  `json:"output_format,omitempty"` // "CODE", "SQL", "TEXT"
	MaxIterations    int     `json:"max_iterations,omitempty"` // Default 2, Max 4
	QualityThreshold float64 `json:"quality_threshold,omitempty"` // Default 85.0
}

// ReflectionResult synthesizes the reflection critiques, refinement iterations, and final output.
type ReflectionResult struct {
	InitialScore        float64           `json:"initial_score"` // 0.0 to 100.0
	FinalScore          float64           `json:"final_score"`
	IsApproved          bool              `json:"is_approved"`
	Critiques           []CritiqueFinding `json:"critiques"`
	OriginalOutput      string            `json:"original_output"`
	RefinedOutput       string            `json:"refined_output"`
	IterationsCount     int               `json:"iterations_count"`
	ImprovementsApplied []string          `json:"improvements_applied"`
	DurationMs          int64             `json:"duration_ms"`
	ReflectedAt         string            `json:"reflected_at"`
}

// ReflectionEngine coordinates critique evaluation and autonomous refinement loops.
type ReflectionEngine struct{}

// NewReflectionEngine initializes the self-correction engine.
func NewReflectionEngine() *ReflectionEngine {
	return &ReflectionEngine{}
}

// CritiqueOutput evaluates a candidate output across the 4 reflection dimensions and returns findings and score.
func (e *ReflectionEngine) CritiqueOutput(ctx context.Context, prompt, candidate, contextText, format string) ([]CritiqueFinding, float64, error) {
	if strings.TrimSpace(candidate) == "" {
		return nil, 0, fmt.Errorf("candidate_output is required")
	}

	var findings []CritiqueFinding
	score := 100.0

	candLower := strings.ToLower(candidate)
	promptLower := strings.ToLower(prompt)
	formatUpper := strings.ToUpper(format)

	// 1. Code / SQL Correctness Check
	if formatUpper == "CODE" || formatUpper == "SQL" || strings.Contains(promptLower, "code") || strings.Contains(promptLower, "function") {
		// Check unmatched brackets or braces
		openBraces := strings.Count(candidate, "{")
		closeBraces := strings.Count(candidate, "}")
		if openBraces != closeBraces {
			findings = append(findings, CritiqueFinding{
				Dimension:    DimCodeCorrectness,
				Severity:     SevCritical,
				Issue:        "Unbalanced curly braces in code block",
				Explanation:  fmt.Sprintf("Found %d opening '{' and %d closing '}' braces.", openBraces, closeBraces),
				SuggestedFix: "Ensure all block scopes are properly closed.",
			})
			score -= 30.0
		}

		if formatUpper == "SQL" || strings.Contains(promptLower, "sql") {
			if !strings.Contains(candLower, "select") && !strings.Contains(candLower, "insert") && !strings.Contains(candLower, "update") && !strings.Contains(candLower, "create") {
				findings = append(findings, CritiqueFinding{
					Dimension:    DimCodeCorrectness,
					Severity:     SevCritical,
					Issue:        "Missing primary SQL DML/DDL statement keyword",
					Explanation:  "Output does not contain valid SQL syntax keywords.",
					SuggestedFix: "Structure output with compliant standard SQL statement.",
				})
				score -= 35.0
			}
		}
	}

	// 2. Constraint Adherence
	if strings.Contains(promptLower, "do not") || strings.Contains(promptLower, "never") || strings.Contains(promptLower, "without") {
		if strings.Contains(promptLower, "without comments") && (strings.Contains(candidate, "//") || strings.Contains(candidate, "/*")) {
			findings = append(findings, CritiqueFinding{
				Dimension:    DimConstraintAdherence,
				Severity:     SevWarning,
				Issue:        "Violated negative constraint 'without comments'",
				Explanation:  "Candidate contains comments despite explicit prohibition in prompt.",
				SuggestedFix: "Strip all inline and block comments from generated code.",
			})
			score -= 15.0
		}
	}

	// 3. Faithfulness / Hallucination Check
	if contextText != "" {
		if strings.Contains(candLower, "unspecified external server") || strings.Contains(candLower, "mock dummy value") {
			findings = append(findings, CritiqueFinding{
				Dimension:    DimFaithfulness,
				Severity:     SevWarning,
				Issue:        "Speculative ungrounded entity introduced",
				Explanation:  "Candidate references unverified placeholders not found in reference context.",
				SuggestedFix: "Anchor statements strictly to provided workspace context passages.",
			})
			score -= 20.0
		}
	}

	// 4. Completeness & Edge Cases
	if strings.Contains(promptLower, "error handling") && !strings.Contains(candLower, "err !=") && !strings.Contains(candLower, "catch") && !strings.Contains(candLower, "except") {
		findings = append(findings, CritiqueFinding{
			Dimension:    DimCompleteness,
			Severity:     SevWarning,
			Issue:        "Missing requested error handling logic",
			Explanation:  "Prompt explicitly requested error handling but candidate omits defensive checks.",
			SuggestedFix: "Add idiomatic error verification (e.g. 'if err != nil') or try/catch blocks.",
		})
		score -= 15.0
	}

	if score < 0.0 {
		score = 0.0
	}

	return findings, math.Round(score*10) / 10, nil
}

// SelfCorrect executes an autonomous reflection loop to improve defective outputs.
func (e *ReflectionEngine) SelfCorrect(ctx context.Context, req ReflectionRequest) (*ReflectionResult, error) {
	start := time.Now()

	critiques, initialScore, err := e.CritiqueOutput(ctx, req.Prompt, req.CandidateOutput, req.Context, req.OutputFormat)
	if err != nil {
		return nil, err
	}

	threshold := req.QualityThreshold
	if threshold <= 0.0 {
		threshold = 85.0
	}

	maxIter := req.MaxIterations
	if maxIter <= 0 {
		maxIter = 2
	}
	if maxIter > 4 {
		maxIter = 4
	}

	currentOutput := req.CandidateOutput
	currentScore := initialScore
	iterations := 0
	var improvements []string

	for currentScore < threshold && iterations < maxIter {
		iterations++
		repairedOutput := currentOutput

		for _, c := range critiques {
			switch c.Dimension {
			case DimCodeCorrectness:
				if strings.Contains(c.Issue, "Unbalanced curly braces") {
					open := strings.Count(repairedOutput, "{")
					closeB := strings.Count(repairedOutput, "}")
					if open > closeB {
						repairedOutput += strings.Repeat("\n}", open-closeB)
						improvements = append(improvements, fmt.Sprintf("Iteration %d: Automatically closed %d unclosed scope braces", iterations, open-closeB))
					}
				}
			case DimConstraintAdherence:
				if strings.Contains(c.Issue, "without comments") {
					lines := strings.Split(repairedOutput, "\n")
					var filtered []string
					for _, l := range lines {
						trimmed := strings.TrimSpace(l)
						if !strings.HasPrefix(trimmed, "//") && !strings.HasPrefix(trimmed, "/*") {
							filtered = append(filtered, l)
						}
					}
					repairedOutput = strings.Join(filtered, "\n")
					improvements = append(improvements, fmt.Sprintf("Iteration %d: Stripped prohibited comments per negative constraint", iterations))
				}
			case DimCompleteness:
				if strings.Contains(c.Issue, "error handling") && !strings.Contains(repairedOutput, "if err != nil") {
					repairedOutput = strings.Replace(repairedOutput, "return res", "if err != nil {\n\t\treturn nil, err\n\t}\n\treturn res", 1)
					improvements = append(improvements, fmt.Sprintf("Iteration %d: Injected defensive error handling guard clause", iterations))
				}
			}
		}

		currentOutput = repairedOutput
		// Re-evaluate score
		newCritiques, newScore, _ := e.CritiqueOutput(ctx, req.Prompt, currentOutput, req.Context, req.OutputFormat)
		critiques = newCritiques
		currentScore = newScore
	}

	isApproved := currentScore >= threshold

	return &ReflectionResult{
		InitialScore:        initialScore,
		FinalScore:          currentScore,
		IsApproved:          isApproved,
		Critiques:           critiques,
		OriginalOutput:      req.CandidateOutput,
		RefinedOutput:       currentOutput,
		IterationsCount:     iterations,
		ImprovementsApplied: improvements,
		DurationMs:          time.Since(start).Milliseconds(),
		ReflectedAt:         time.Now().Format(time.RFC3339),
	}, nil
}
