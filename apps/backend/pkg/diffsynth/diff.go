package diffsynth

import (
	"fmt"
	"strings"
	"time"
)

type RefactorCategory string

const (
	CategoryPerformance RefactorCategory = "performance"
	CategorySecurity    RefactorCategory = "security"
	CategoryTypeSafety  RefactorCategory = "type_safety"
	CategoryFormatting  RefactorCategory = "formatting"
	CategoryGeneral     RefactorCategory = "general"
)

type DiffRequest struct {
	FileName        string           `json:"file_name"`
	OriginalCode    string           `json:"original_code"`
	ModifiedCode    string           `json:"modified_code"`
	RefactorIntent  RefactorCategory `json:"refactor_intent"`
}

type DiffResult struct {
	FileName         string           `json:"file_name"`
	UnifiedDiff      string           `json:"unified_diff"`
	LinesAdded       int              `json:"lines_added"`
	LinesRemoved     int              `json:"lines_removed"`
	TotalChanges     int              `json:"total_changes"`
	RefactorCategory RefactorCategory `json:"refactor_category"`
	ImpactScore      string           `json:"impact_score"` // "LOW", "MEDIUM", "HIGH"
	Timestamp        string           `json:"timestamp"`
}

type Synthesizer struct{}

func NewSynthesizer() *Synthesizer {
	return &Synthesizer{}
}

// SynthesizeDiff produces a unified diff string and metrics from original and modified code.
func (s *Synthesizer) SynthesizeDiff(req DiffRequest) (*DiffResult, error) {
	if req.FileName == "" {
		req.FileName = "source_file"
	}

	origLines := strings.Split(req.OriginalCode, "\n")
	modLines := strings.Split(req.ModifiedCode, "\n")

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("--- a/%s\n", req.FileName))
	sb.WriteString(fmt.Sprintf("+++ b/%s\n", req.FileName))
	sb.WriteString("@@ -1," + fmt.Sprintf("%d +1,%d @@\n", len(origLines), len(modLines)))

	added := 0
	removed := 0

	// Simple line-by-line diff synthesis
	maxLen := len(origLines)
	if len(modLines) > maxLen {
		maxLen = len(modLines)
	}

	for i := 0; i < maxLen; i++ {
		var orig, mod string
		if i < len(origLines) {
			orig = origLines[i]
		}
		if i < len(modLines) {
			mod = modLines[i]
		}

		if orig == mod {
			sb.WriteString(fmt.Sprintf(" %s\n", orig))
		} else {
			if i < len(origLines) && orig != "" {
				sb.WriteString(fmt.Sprintf("-%s\n", orig))
				removed++
			}
			if i < len(modLines) && mod != "" {
				sb.WriteString(fmt.Sprintf("+%s\n", mod))
				added++
			}
		}
	}

	totalChanges := added + removed
	impact := "LOW"
	if totalChanges > 20 {
		impact = "HIGH"
	} else if totalChanges > 5 {
		impact = "MEDIUM"
	}

	category := req.RefactorIntent
	if category == "" {
		category = CategoryGeneral
	}

	return &DiffResult{
		FileName:         req.FileName,
		UnifiedDiff:      sb.String(),
		LinesAdded:       added,
		LinesRemoved:     removed,
		TotalChanges:     totalChanges,
		RefactorCategory: category,
		ImpactScore:      impact,
		Timestamp:        time.Now().Format(time.RFC3339),
	}, nil
}
