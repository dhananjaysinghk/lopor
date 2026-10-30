package semanticdiff

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"time"
)

// SemanticDiffEngine coordinates structural segmentation, semantic vector comparison, and risk scoring.
type SemanticDiffEngine struct{}

// NewEngine initializes a new SemanticDiffEngine.
func NewEngine() *SemanticDiffEngine {
	return &SemanticDiffEngine{}
}

// Compare analyzes the semantic divergence between original and revised document versions.
func (e *SemanticDiffEngine) Compare(req CompareRequest) (*CompareResponse, error) {
	if req.OriginalText == "" && req.RevisedText == "" {
		return nil, errors.New("original_text or revised_text must be non-empty")
	}

	granularity := req.Granularity
	if granularity == "" {
		granularity = GranularitySentence
	}

	docType := req.DocType
	if docType == "" {
		docType = "GENERAL"
	}

	origSegments := SegmentText(req.OriginalText, granularity)
	revSegments := SegmentText(req.RevisedText, granularity)

	maxLen := len(origSegments)
	if len(revSegments) > maxLen {
		maxLen = len(revSegments)
	}

	var deltas []SemanticDelta
	var totalSim float64
	var criticalCount, highCount int

	for i := 0; i < maxLen; i++ {
		orig := ""
		rev := ""
		if i < len(origSegments) {
			orig = origSegments[i]
		}
		if i < len(revSegments) {
			rev = revSegments[i]
		}

		if orig == rev {
			totalSim += 1.0
			continue
		}

		vecA := ComputeSegmentVector(orig)
		vecB := ComputeSegmentVector(rev)
		sim := CosineSimilarity(vecA, vecB)
		totalSim += sim

		category, severity, intentPreserved, rationale, action := ClassifySemanticShift(orig, rev, docType)

		if severity == SeverityCritical {
			criticalCount++
		} else if severity == SeverityHigh {
			highCount++
		}

		deltaID := generateID("dlt")
		deltas = append(deltas, SemanticDelta{
			DeltaID:            deltaID,
			OriginalSegment:    orig,
			RevisedSegment:     rev,
			Category:           category,
			Severity:           severity,
			SimilarityScore:    sim,
			IntentPreserved:    intentPreserved,
			ExecutiveRationale: rationale,
			ActionRequired:     action,
		})
	}

	var overallSimilarity float64
	if maxLen > 0 {
		overallSimilarity = math.Round((totalSim/float64(maxLen))*1000) / 10
	} else {
		overallSimilarity = 100.0
	}

	// Compute composite contractual risk score (0 - 100%)
	riskScore := float64(criticalCount)*35.0 + float64(highCount)*15.0 + (100.0-overallSimilarity)*0.2
	if riskScore > 100.0 {
		riskScore = 100.0
	}
	riskScore = math.Round(riskScore*10) / 10

	requiresReview := criticalCount > 0 || highCount >= 2 || riskScore >= 40.0

	// Formulate executive summary
	var summary string
	if len(deltas) == 0 {
		summary = "Zero semantic deltas detected. Original and revised versions are functionally identical."
	} else if requiresReview {
		summary = fmt.Sprintf("High-impact semantic divergence detected (%d critical, %d high-severity shifts; risk score: %.1f%%). Immediate human review required prior to acceptance.", criticalCount, highCount, riskScore)
	} else {
		summary = fmt.Sprintf("Minor semantic modifications identified across %d segments (similarity: %.1f%%). Underlying obligations and intent preserved.", len(deltas), overallSimilarity)
	}

	return &CompareResponse{
		WorkspaceID:          req.WorkspaceID,
		DocType:              docType,
		Granularity:          granularity,
		OverallSimilarity:    overallSimilarity,
		ContractualRiskScore: riskScore,
		TotalDeltas:          len(deltas),
		CriticalDeltasCount:  criticalCount,
		HighDeltasCount:      highCount,
		RequiresHumanReview:  requiresReview,
		ExecutiveSummary:     summary,
		Deltas:               deltas,
		AnalyzedAt:           time.Now().UTC(),
	}, nil
}

func generateID(prefix string) string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%s_%s", prefix, hex.EncodeToString(b))
}

