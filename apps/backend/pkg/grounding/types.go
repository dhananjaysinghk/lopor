package grounding

import (
	"time"
)

// ClaimStatus indicates whether an individual propositional claim is backed by provided evidence.
type ClaimStatus string

const (
	StatusSupported    ClaimStatus = "SUPPORTED"
	StatusContradicted ClaimStatus = "CONTRADICTED"
	StatusUngrounded   ClaimStatus = "UNGROUNDED"
)

// HallucinationRisk characterizes the aggregate hallucination liability of a generated response.
type HallucinationRisk string

const (
	RiskLow      HallucinationRisk = "LOW"
	RiskMedium   HallucinationRisk = "MEDIUM"
	RiskHigh     HallucinationRisk = "HIGH"
	RiskCritical HallucinationRisk = "CRITICAL"
)

// SourceEvidence represents a reference document, chunk, or retrieved source context.
type SourceEvidence struct {
	ID         string            `json:"id"`
	DocumentID string            `json:"document_id,omitempty"`
	Title      string            `json:"title"`
	Content    string            `json:"content"`
	URI        string            `json:"uri,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

// AtomicClaim represents a single factual proposition extracted from the generated output.
type AtomicClaim struct {
	ID                   string      `json:"id"`
	ClaimText            string      `json:"claim_text"`
	OffsetStart          int         `json:"offset_start"`
	OffsetEnd            int         `json:"offset_end"`
	Status               ClaimStatus `json:"status"`
	EntailmentConfidence float64     `json:"entailment_confidence"` // 0.0 - 1.0
	SourceEvidenceID     string      `json:"source_evidence_id,omitempty"`
	SourceTitle          string      `json:"source_title,omitempty"`
	EvidenceSnippet      string      `json:"evidence_snippet,omitempty"`
	CitationIndex        int         `json:"citation_index,omitempty"`
	Explanation          string      `json:"explanation,omitempty"`
}

// CitationAnchor links an in-text claim to an authoritative source evidence block.
type CitationAnchor struct {
	CitationIndex int    `json:"citation_index"`
	ClaimID       string `json:"claim_id"`
	SourceID      string `json:"source_id"`
	DocumentTitle string `json:"document_title"`
	URI           string `json:"uri,omitempty"`
	ExactSnippet  string `json:"exact_snippet"`
}

// VerifyGroundingRequest encapsulates parameters for claim-level fact verification.
type VerifyGroundingRequest struct {
	WorkspaceID       string           `json:"workspace_id,omitempty"`
	GeneratedResponse string           `json:"generated_response"`
	SourceEvidences   []SourceEvidence `json:"source_evidences"`
	Strictness        float64          `json:"strictness,omitempty"`         // Threshold for entailment (default: 0.65)
	AutoRemediate     bool             `json:"auto_remediate,omitempty"`     // Generate sanitized output without hallucinations
}

// GroundingReport represents the comprehensive audit report for a response's faithfulness.
type GroundingReport struct {
	VerificationID       string            `json:"verification_id"`
	WorkspaceID          string            `json:"workspace_id"`
	TotalClaims          int               `json:"total_claims"`
	SupportedClaims      int               `json:"supported_claims"`
	ContradictedClaims   int               `json:"contradicted_claims"`
	UngroundedClaims     int               `json:"ungrounded_claims"`
	GroundednessScore    float64           `json:"groundedness_score"` // 0.0 - 1.0 (faithfulness ratio)
	HallucinationRisk    HallucinationRisk `json:"hallucination_risk"`
	Claims               []AtomicClaim     `json:"claims"`
	Citations            []CitationAnchor  `json:"citations"`
	AttributedResponse   string            `json:"attributed_response"` // Response with embedded [^1] citation anchors
	RemediatedResponse   string            `json:"remediated_response,omitempty"`
	VerificationDuration float64           `json:"verification_duration_ms"`
	VerifiedAt           time.Time         `json:"verified_at"`
}

// GroundingMetrics tracks telemetry on enterprise hallucination detection and audit scores.
type GroundingMetrics struct {
	WorkspaceID                string    `json:"workspace_id"`
	TotalVerifications         int64     `json:"total_verifications"`
	TotalClaimsEvaluated       int64     `json:"total_claims_evaluated"`
	TotalSupportedClaims       int64     `json:"total_supported_claims"`
	TotalContradictedClaims    int64     `json:"total_contradicted_claims"`
	TotalUngroundedClaims      int64     `json:"total_ungrounded_claims"`
	AverageGroundednessScore   float64   `json:"average_groundedness_score"`
	CriticalRiskIncidentsCount int64     `json:"critical_risk_incidents_count"`
	LastVerifiedAt             time.Time `json:"last_verified_at"`
}

