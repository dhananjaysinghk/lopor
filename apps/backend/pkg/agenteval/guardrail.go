package agenteval

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"
)

type GuardrailVerdict string

const (
	VerdictAllow  GuardrailVerdict = "ALLOW"
	VerdictBlock  GuardrailVerdict = "BLOCK"
	VerdictFlag   GuardrailVerdict = "FLAG"
	VerdictRedact GuardrailVerdict = "REDACT"
)

type PolicyViolation struct {
	PolicyID     string           `json:"policy_id"`
	PolicyName   string           `json:"policy_name"`
	Severity     string           `json:"severity"` // "LOW", "MEDIUM", "HIGH", "CRITICAL"
	MatchedText  string           `json:"matched_text"`
	Recommendation string         `json:"recommendation"`
}

type GuardrailRequest struct {
	PromptText     string   `json:"prompt_text"`
	ProposedTool   string   `json:"proposed_tool,omitempty"`
	ToolArguments  string   `json:"tool_arguments,omitempty"`
	CustomPolicies []string `json:"custom_policies,omitempty"`
}

type GuardrailResult struct {
	Verdict          GuardrailVerdict  `json:"verdict"`
	Passed           bool              `json:"passed"`
	ComplianceScore  float64           `json:"compliance_score"` // 0.0 - 1.0
	Violations       []PolicyViolation `json:"violations"`
	SanitizedPrompt  string            `json:"sanitized_prompt,omitempty"`
	EvaluatedAt      string            `json:"evaluated_at"`
}

type PolicyRule struct {
	ID             string
	Name           string
	Severity       string
	Pattern        *regexp.Regexp
	Recommendation string
}

type GuardrailEngine struct {
	rules []PolicyRule
}

func NewGuardrailEngine() *GuardrailEngine {
	rules := []PolicyRule{
		{
			ID:             "POL-001",
			Name:           "Prompt Injection / Jailbreak Attempt",
			Severity:       "CRITICAL",
			Pattern:        regexp.MustCompile(`(?i)(ignore (all )?(previous|above) instructions|you are now in (developer|dan) mode|disregard system prompt)`),
			Recommendation: "Block prompt execution and alert workspace security officer.",
		},
		{
			ID:             "POL-002",
			Name:           "Unauthorized Shell / System Destruction",
			Severity:       "CRITICAL",
			Pattern:        regexp.MustCompile(`(?i)(rm\s+-rf\s+/|mkfs\.|dd\s+if=|:\(\)\{ :\|:& \};:)`),
			Recommendation: "Immediately terminate agent execution and quarantine workspace sandbox.",
		},
		{
			ID:             "POL-003",
			Name:           "Toxic / Offensive Language",
			Severity:       "HIGH",
			Pattern:        regexp.MustCompile(`(?i)(hate\s+speech|violent\s+extremism|self-harm)`),
			Recommendation: "Refuse generation and output safety guideline violation message.",
		},
		{
			ID:             "POL-004",
			Name:           "Confidential System Data Exfiltration",
			Severity:       "HIGH",
			Pattern:        regexp.MustCompile(`(?i)(/etc/passwd|/etc/shadow|\.env|id_rsa|credentials\.json)`),
			Recommendation: "Deny file access tool invocation.",
		},
	}

	return &GuardrailEngine{rules: rules}
}

// VerifyGuardrails inspects user prompts or tool invocations against safety policies.
func (ge *GuardrailEngine) VerifyGuardrails(ctx context.Context, req GuardrailRequest) (*GuardrailResult, error) {
	if strings.TrimSpace(req.PromptText) == "" && strings.TrimSpace(req.ToolArguments) == "" {
		return nil, fmt.Errorf("prompt text or tool arguments must be provided")
	}

	inspectedContent := req.PromptText
	if req.ToolArguments != "" {
		inspectedContent += "\n" + req.ToolArguments
	}

	var violations []PolicyViolation
	critCount := 0
	highCount := 0

	for _, rule := range ge.rules {
		if rule.Pattern.MatchString(inspectedContent) {
			match := rule.Pattern.FindString(inspectedContent)
			violations = append(violations, PolicyViolation{
				PolicyID:       rule.ID,
				PolicyName:     rule.Name,
				Severity:       rule.Severity,
				MatchedText:    match,
				Recommendation: rule.Recommendation,
			})

			if rule.Severity == "CRITICAL" {
				critCount++
			} else if rule.Severity == "HIGH" {
				highCount++
			}
		}
	}

	verdict := VerdictAllow
	passed := true
	score := 1.0

	if critCount > 0 {
		verdict = VerdictBlock
		passed = false
		score = 0.0
	} else if highCount > 0 {
		verdict = VerdictFlag
		passed = false
		score = 0.5
	} else if len(violations) > 0 {
		verdict = VerdictFlag
		passed = true
		score = 0.8
	}

	sanitized := req.PromptText
	if verdict == VerdictBlock {
		sanitized = "[BLOCKED BY LOPOR ENTERPRISE SECURITY GUARDRAILS]"
	}

	return &GuardrailResult{
		Verdict:         verdict,
		Passed:          passed,
		ComplianceScore: score,
		Violations:      violations,
		SanitizedPrompt: sanitized,
		EvaluatedAt:     time.Now().Format(time.RFC3339),
	}, nil
}
