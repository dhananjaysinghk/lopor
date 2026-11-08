package grounding

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"
)

var (
	ErrEmptyResponse = errors.New("generated response cannot be empty")
	ErrEmptySources  = errors.New("source evidences cannot be empty")
)

// GroundingEngine provides claim-level fact verification, hallucination risk assessment, and citation attribution.
type GroundingEngine struct {
	mu      sync.RWMutex
	metrics map[string]*GroundingMetrics
}

// NewEngine initializes a new GroundingEngine instance.
func NewEngine() *GroundingEngine {
	return &GroundingEngine{
		metrics: make(map[string]*GroundingMetrics),
	}
}

// VerifyGrounding executes end-to-end claim decomposition, evidence entailment verification, and citation synthesis.
func (e *GroundingEngine) VerifyGrounding(ctx context.Context, req VerifyGroundingRequest) (*GroundingReport, error) {
	start := time.Now()

	trimmedResponse := strings.TrimSpace(req.GeneratedResponse)
	if trimmedResponse == "" {
		return nil, ErrEmptyResponse
	}
	if len(req.SourceEvidences) == 0 {
		return nil, ErrEmptySources
	}

	strictness := req.Strictness
	if strictness <= 0 {
		strictness = 0.65
	}

	// Step 1: Decompose into atomic claims
	rawClaims := DecomposeResponse(trimmedResponse)
	if len(rawClaims) == 0 {
		// Single sentence / short statement fallback
		rawClaims = []AtomicClaim{
			{
				ID:          "claim_001",
				ClaimText:   trimmedResponse,
				OffsetStart: 0,
				OffsetEnd:   len([]rune(trimmedResponse)),
			},
		}
	}

	// Step 2: Verify each claim against source evidences
	verifiedClaims := make([]AtomicClaim, len(rawClaims))
	supportedCount := 0
	contradictedCount := 0
	ungroundedCount := 0

	for i, claim := range rawClaims {
		verified := VerifyClaim(claim, req.SourceEvidences, strictness)
		verifiedClaims[i] = verified

		switch verified.Status {
		case StatusSupported:
			supportedCount++
		case StatusContradicted:
			contradictedCount++
		case StatusUngrounded:
			ungroundedCount++
		}
	}

	totalClaims := len(verifiedClaims)
	var groundednessScore float64
	if totalClaims > 0 {
		groundednessScore = float64(supportedCount) / float64(totalClaims)
	}
	groundednessScore = math.Round(groundednessScore*1000) / 1000

	// Step 3: Determine aggregate hallucination risk
	var risk HallucinationRisk
	switch {
	case contradictedCount > 0:
		risk = RiskCritical
	case groundednessScore >= 0.85:
		risk = RiskLow
	case groundednessScore >= 0.55:
		risk = RiskMedium
	default:
		risk = RiskHigh
	}

	// Step 4: Attribute citations and build remediated response
	attributedResp, citations, remediatedResp := BuildAttributedResponse(trimmedResponse, verifiedClaims, req.SourceEvidences)

	durationMs := float64(time.Since(start).Microseconds()) / 1000.0
	if durationMs < 0.05 {
		durationMs = 0.05
	}

	verificationID := generateID("grd")

	report := &GroundingReport{
		VerificationID:       verificationID,
		WorkspaceID:          req.WorkspaceID,
		TotalClaims:          totalClaims,
		SupportedClaims:      supportedCount,
		ContradictedClaims:   contradictedCount,
		UngroundedClaims:     ungroundedCount,
		GroundednessScore:    groundednessScore,
		HallucinationRisk:    risk,
		Claims:               verifiedClaims,
		Citations:            citations,
		AttributedResponse:   attributedResp,
		RemediatedResponse:   remediatedResp,
		VerificationDuration: durationMs,
		VerifiedAt:           time.Now().UTC(),
	}

	// Step 5: Update workspace metrics
	e.recordMetrics(req.WorkspaceID, totalClaims, supportedCount, contradictedCount, ungroundedCount, groundednessScore, risk)

	return report, nil
}

// GetMetrics returns workspace grounding and hallucination metrics.
func (e *GroundingEngine) GetMetrics(workspaceID string) *GroundingMetrics {
	e.mu.RLock()
	defer e.mu.RUnlock()

	m, exists := e.metrics[workspaceID]
	if !exists {
		return &GroundingMetrics{
			WorkspaceID:    workspaceID,
			LastVerifiedAt: time.Now().UTC(),
		}
	}
	return m
}

func (e *GroundingEngine) recordMetrics(wsID string, total, supported, contradicted, ungrounded int, score float64, risk HallucinationRisk) {
	if wsID == "" {
		return
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	m, exists := e.metrics[wsID]
	if !exists {
		m = &GroundingMetrics{
			WorkspaceID:    wsID,
			LastVerifiedAt: time.Now().UTC(),
		}
		e.metrics[wsID] = m
	}

	m.TotalVerifications++
	m.TotalClaimsEvaluated += int64(total)
	m.TotalSupportedClaims += int64(supported)
	m.TotalContradictedClaims += int64(contradicted)
	m.TotalUngroundedClaims += int64(ungrounded)
	m.LastVerifiedAt = time.Now().UTC()

	if risk == RiskCritical {
		m.CriticalRiskIncidentsCount++
	}

	if m.TotalVerifications > 0 {
		m.AverageGroundednessScore = (m.AverageGroundednessScore*float64(m.TotalVerifications-1) + score) / float64(m.TotalVerifications)
		m.AverageGroundednessScore = math.Round(m.AverageGroundednessScore*1000) / 1000
	}
}

func generateID(prefix string) string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%s_%s", prefix, hex.EncodeToString(b))
}

