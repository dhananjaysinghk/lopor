package selfheal

import (
	"fmt"
	"strings"
)

// BuildDiagnosticReprompt constructs an error-feedback prompt instructing the model to self-correct.
func BuildDiagnosticReprompt(fault ExecutionFault) string {
	var b strings.Builder

	b.WriteString("### SYSTEM DIAGNOSTIC ERROR CORRECTION DIRECTIVE\n")
	b.WriteString("Your previous execution resulted in a runtime validation failure. Correct ONLY the offending segment while preserving the original intent.\n\n")

	b.WriteString("#### FAULT CLASSIFICATION\n")
	b.WriteString(fmt.Sprintf("- **Category**: %s\n", fault.Category))
	b.WriteString(fmt.Sprintf("- **Severity**: %s\n", fault.Severity))
	b.WriteString(fmt.Sprintf("- **Error Diagnostic**: %s\n\n", fault.ErrorMessage))

	if fault.TargetSchema != "" {
		b.WriteString("#### REQUIRED TARGET SCHEMA / SPECIFICATION\n")
		b.WriteString("```json\n")
		b.WriteString(fault.TargetSchema)
		b.WriteString("\n```\n\n")
	}

	if fault.OffendingOutput != "" {
		b.WriteString("#### OFFENDING ERRONEOUS OUTPUT\n")
		b.WriteString("```\n")
		// Limit to first 1200 characters to prevent prompt bloat
		offending := fault.OffendingOutput
		if len(offending) > 1200 {
			offending = offending[:1200] + "... [TRUNCATED]"
		}
		b.WriteString(offending)
		b.WriteString("\n```\n\n")
	}

	b.WriteString("#### REPAIR INSTRUCTIONS\n")
	switch fault.Category {
	case FaultSchemaValidation:
		b.WriteString("1. Output MUST be strictly valid, parsable JSON matching the target schema.\n")
		b.WriteString("2. Do NOT include markdown code fences, conversational preambles, or postscript explanations.\n")
		b.WriteString("3. Ensure all keys and string values use standard double quotes and no trailing commas exist.\n")

	case FaultCodeSyntaxError:
		b.WriteString("1. Inspect syntax errors indicated in the diagnostic log above.\n")
		b.WriteString("2. Fix all syntax, missing closing braces, invalid imports, and scope issues.\n")
		b.WriteString("3. Provide complete, executable, production-grade code.\n")

	case FaultHallucinatedTool:
		b.WriteString("1. The tool or function signature called does not exist or parameters are mismatched.\n")
		b.WriteString("2. Re-check the available tool specifications in your system prompt.\n")
		b.WriteString("3. Invoke only registered tools with valid parameter keys.\n")

	default:
		b.WriteString("1. Rectify the diagnostic error highlighted above.\n")
		b.WriteString("2. Adhere strictly to the requested constraints.\n")
	}

	return b.String()
}

// PruneContextForTokenLimit compresses or slices conversational input to resolve context limit errors.
func PruneContextForTokenLimit(originalInput string, maxChars int) string {
	if len(originalInput) <= maxChars {
		return originalInput
	}

	// Keep instructions at the start and recent query context at the end, pruning the middle
	keepHead := maxChars / 3
	keepTail := maxChars - keepHead - 100

	head := originalInput[:keepHead]
	tail := originalInput[len(originalInput)-keepTail:]

	return fmt.Sprintf("%s\n\n[... %d characters of intermediate context pruned to fit model context window ...]\n\n%s", head, len(originalInput)-maxChars, tail)
}
