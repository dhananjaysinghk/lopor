package semanticdiff

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	reNumbers = regexp.MustCompile(`\b\d+(?:\.\d+)?%?\b`)
)

// ClassifySemanticShift evaluates the delta between original and revised segments.
func ClassifySemanticShift(orig, rev string, docType string) (SemanticChangeCategory, ImpactSeverity, bool, string, string) {
	origLower := strings.ToLower(orig)
	revLower := strings.ToLower(rev)

	vecOrig := ComputeSegmentVector(orig)
	vecRev := ComputeSegmentVector(rev)
	similarity := CosineSimilarity(vecOrig, vecRev)

	// Check 1: Numeric threshold change
	origNums := reNumbers.FindAllString(orig, -1)
	revNums := reNumbers.FindAllString(rev, -1)
	if !slicesEqual(origNums, revNums) && (len(origNums) > 0 || len(revNums) > 0) {
		rationale := fmt.Sprintf("Numeric parameters modified: [%s] -> [%s].", strings.Join(origNums, ", "), strings.Join(revNums, ", "))
		return CategoryThresholdChange, SeverityHigh, false, rationale, "Verify metric targets and operational SLA commitments with stakeholders."
	}

	// Check 2: Legal / Liability / Risk Escalation
	riskKeywords := []string{"liability", "indemnif", "warranty", "breach", "damages", "disclaim", "exclusive remedy", "governing law"}
	for _, kw := range riskKeywords {
		if strings.Contains(revLower, kw) && !strings.Contains(origLower, kw) {
			return CategoryRiskEscalation, SeverityCritical, false,
				fmt.Sprintf("Introduction of restrictive legal constraint or liability keyword: '%s'.", kw),
				"Mandatory legal counsel sign-off required before approving amendment."
		}
		if strings.Contains(origLower, kw) && !strings.Contains(revLower, kw) {
			return CategoryRiskEscalation, SeverityCritical, false,
				fmt.Sprintf("Removal of protective covenant or warranty keyword: '%s'.", kw),
				"Review omission of protective legal terms."
		}
	}

	// Check 3: Obligation / Responsibility Shift
	obligationKeywords := []string{"shall", "must", "will", "required to", "sole responsibility", "within 30 days", "immediately"}
	for _, kw := range obligationKeywords {
		if strings.Contains(revLower, kw) && !strings.Contains(origLower, kw) {
			return CategoryObligationShift, SeverityHigh, false,
				fmt.Sprintf("New mandatory obligation added: '%s'.", kw),
				"Ensure operational capacity satisfies newly assigned obligation."
		}
	}

	// Check 4: Condition or Caveat Addition
	conditionKeywords := []string{"provided that", "unless", "except where", "subject to", "in the event of", "only if"}
	for _, kw := range conditionKeywords {
		if strings.Contains(revLower, kw) && !strings.Contains(origLower, kw) {
			return CategoryConditionAddition, SeverityModerate, true,
				fmt.Sprintf("New conditional constraint introduced: '%s'.", kw),
				"Validate operational trigger conditions."
		}
	}

	// Check 5: Code refactor vs Behavioral Drift
	if strings.ToUpper(docType) == "CODE" {
		if similarity >= 0.85 {
			return CategoryCodeRefactor, SeverityCosmetic, true,
				"Structural syntax formatting or variable renaming with preserved semantics.",
				""
		}
		return CategoryBehavioralDrift, SeverityHigh, false,
			"Significant semantic drift in code logic; control flow or side effects altered.",
			"Run regression and integration test harnesses to verify behavior."
	}

	// Check 6: General Clarification vs Cosmetic Rephrasing
	if similarity >= 0.80 {
		return CategoryClarification, SeverityCosmetic, true,
			"Stylistic rewording or tone refinement without contractual or behavioral shift.",
			""
	}

	return CategoryClarification, SeverityModerate, true,
		"Substantive explanatory rephrasing; underlying intent remains closely aligned.",
		"Review phrasing to ensure unintended ambiguity is not introduced."
}

func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

