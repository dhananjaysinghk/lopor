package semanticdiff

import (
	"time"
)

// DiffGranularity defines the resolution level of change detection.
type DiffGranularity string

const (
	GranularityClause    DiffGranularity = "CLAUSE"
	GranularitySentence  DiffGranularity = "SENTENCE"
	GranularityParagraph DiffGranularity = "PARAGRAPH"
	GranularityCodeBlock DiffGranularity = "CODE_BLOCK"
)

// SemanticChangeCategory classifies the qualitative nature and business impact of a change.
type SemanticChangeCategory string

const (
	CategoryClarification     SemanticChangeCategory = "CLARIFICATION"      // Rephrasing, tone adjustment without altering obligations
	CategoryConditionAddition SemanticChangeCategory = "CONDITION_ADDITION" // New constraints, prerequisites, or caveats introduced
	CategoryObligationShift   SemanticChangeCategory = "OBLIGATION_SHIFT"   // Changes who is responsible or timeline constraints
	CategoryThresholdChange   SemanticChangeCategory = "THRESHOLD_CHANGE"   // Numeric adjustments (e.g. SLA 99.9% -> 99.99%, latency 50ms -> 30ms)
	CategoryRiskEscalation    SemanticChangeCategory = "RISK_ESCALATION"    // Broadening liabilities or removing warranties/remedies
	CategoryCodeRefactor      SemanticChangeCategory = "CODE_REFACTOR"      // Structural code modification preserving behavioral contract
	CategoryBehavioralDrift   SemanticChangeCategory = "BEHAVIORAL_DRIFT"   // Breaking logic or altered side-effects in code/prompt
)

// ImpactSeverity rates the significance of semantic divergence.
type ImpactSeverity string

const (
	SeverityCosmetic ImpactSeverity = "COSMETIC" // Trivial wording or style update
	SeverityModerate ImpactSeverity = "MODERATE" // Meaningful nuance without fundamental contract change
	SeverityHigh     ImpactSeverity = "HIGH"     // Direct shift in business, legal, or runtime contract
	SeverityCritical ImpactSeverity = "CRITICAL" // Severe risk escalation or breaking behavioral alteration
)

// SemanticDelta represents an individual localized shift between original and revised text.
type SemanticDelta struct {
	DeltaID            string                 `json:"delta_id"`
	OriginalSegment    string                 `json:"original_segment"`
	RevisedSegment     string                 `json:"revised_segment"`
	Category           SemanticChangeCategory `json:"category"`
	Severity           ImpactSeverity         `json:"severity"`
	SimilarityScore    float64                `json:"similarity_score"` // 0.0 (totally different) to 1.0 (identical)
	IntentPreserved    bool                   `json:"intent_preserved"`
	ExecutiveRationale string                 `json:"executive_rationale"`
	ActionRequired     string                 `json:"action_required,omitempty"`
}

// CompareRequest specifies texts and parameters for semantic diffing.
type CompareRequest struct {
	WorkspaceID  string          `json:"workspace_id"`
	OriginalText string          `json:"original_text"`
	RevisedText  string          `json:"revised_text"`
	Granularity  DiffGranularity `json:"granularity,omitempty"`
	DocType      string          `json:"doc_type,omitempty"` // "CONTRACT", "PROMPT", "CODE", "POLICY", "SPEC"
}

// CompareResponse encapsulates the structured semantic diff analysis.
type CompareResponse struct {
	WorkspaceID          string          `json:"workspace_id"`
	DocType              string          `json:"doc_type"`
	Granularity          DiffGranularity `json:"granularity"`
	OverallSimilarity    float64         `json:"overall_similarity"`     // 0.0 - 100.0%
	ContractualRiskScore float64         `json:"contractual_risk_score"` // 0.0 - 100.0% (higher = higher risk shift)
	TotalDeltas          int             `json:"total_deltas"`
	CriticalDeltasCount  int             `json:"critical_deltas_count"`
	HighDeltasCount      int             `json:"high_deltas_count"`
	RequiresHumanReview  bool            `json:"requires_human_review"`
	ExecutiveSummary     string          `json:"executive_summary"`
	Deltas               []SemanticDelta `json:"deltas"`
	AnalyzedAt           time.Time       `json:"analyzed_at"`
}

