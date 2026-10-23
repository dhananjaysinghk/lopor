package promptopt

import (
	"fmt"
	"sort"
	"strings"
)

// CompileCandidatePrompt synthesizes instructions, few-shot demonstrations, and anti-patterns into a single prompt.
func CompileCandidatePrompt(instructions string, exemplars []PromptExample, antiPatterns []string) string {
	var b strings.Builder

	b.WriteString(strings.TrimSpace(instructions))
	b.WriteString("\n\n")

	if len(antiPatterns) > 0 {
		b.WriteString("### CRITICAL ANTI-PATTERNS (DO NOT DO THE FOLLOWING):\n")
		for i, ap := range antiPatterns {
			b.WriteString(fmt.Sprintf("%d. %s\n", i+1, ap))
		}
		b.WriteString("\n")
	}

	if len(exemplars) > 0 {
		b.WriteString("### DEMONSTRATION EXEMPLARS (FEW-SHOT GUIDANCE):\n")
		for i, ex := range exemplars {
			b.WriteString(fmt.Sprintf("--- Example %d ---\n", i+1))
			b.WriteString("Input:\n")
			for k, v := range ex.InputVariables {
				b.WriteString(fmt.Sprintf("  %s: %s\n", k, v))
			}
			if ex.Rationale != "" {
				b.WriteString(fmt.Sprintf("Reasoning: %s\n", ex.Rationale))
			}
			b.WriteString(fmt.Sprintf("Target Output:\n%s\n\n", strings.TrimSpace(ex.TargetOutput)))
		}
	}

	return strings.TrimSpace(b.String())
}

// MutateInstructionClarification sharpens baseline instructions with explicit structural constraints.
func MutateInstructionClarification(basePrompt string) string {
	var b strings.Builder
	b.WriteString(strings.TrimSpace(basePrompt))
	b.WriteString("\n\n### STRUCTURAL & OPERATIONAL EXECUTION CONSTRAINTS:\n")
	b.WriteString("1. Maintain absolute faithfulness to provided input data; do NOT extrapolate unverified facts.\n")
	b.WriteString("2. Adhere strictly to requested formats (e.g., valid JSON, clean Markdown) without unnecessary preambles.\n")
	b.WriteString("3. Ensure all edge cases and boundary conditions are handled deterministically.\n")
	return b.String()
}

// MutateReasoningScaffold injects a structured step-by-step reasoning protocol into the prompt.
func MutateReasoningScaffold(basePrompt string) string {
	var b strings.Builder
	b.WriteString(strings.TrimSpace(basePrompt))
	b.WriteString("\n\n### REASONING PROTOCOL:\n")
	b.WriteString("Before formulating the final response, systematically execute these steps:\n")
	b.WriteString("1. Deconstruct input parameters and verify semantic constraints.\n")
	b.WriteString("2. Cross-reference edge cases, missing data, and structural boundaries.\n")
	b.WriteString("3. Produce the final concise, validated output adhering strictly to the objective.\n")
	return b.String()
}

// BootstrapFewShots extracts the top-scoring positive exemplars from training data.
func BootstrapFewShots(examples []PromptExample, maxExemplars int) []PromptExample {
	if len(examples) == 0 {
		return nil
	}

	var positives []PromptExample
	for _, ex := range examples {
		if !ex.IsNegativeExemplar {
			positives = append(positives, ex)
		}
	}

	// Sort descending by score
	sort.Slice(positives, func(i, j int) bool {
		return positives[i].Score > positives[j].Score
	})

	if len(positives) > maxExemplars {
		positives = positives[:maxExemplars]
	}

	return positives
}

// ExtractHardNegatives synthesizes explicit negative anti-pattern rules from failed/negative exemplars.
func ExtractHardNegatives(examples []PromptExample) []string {
	var antiPatterns []string
	for _, ex := range examples {
		if ex.IsNegativeExemplar {
			if ex.Rationale != "" {
				antiPatterns = append(antiPatterns, fmt.Sprintf("Avoid pattern: %s (Failed on input: %v)", ex.Rationale, ex.InputVariables))
			} else {
				antiPatterns = append(antiPatterns, fmt.Sprintf("Avoid producing flawed response: '%s'", truncate(ex.TargetOutput, 80)))
			}
		}
	}

	if len(antiPatterns) == 0 {
		antiPatterns = append(antiPatterns, "Avoid generic conversational fluff or apologetic filler phrases.")
		antiPatterns = append(antiPatterns, "Never omit required attributes or leave fields with ambiguous placeholders.")
	}

	return antiPatterns
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
