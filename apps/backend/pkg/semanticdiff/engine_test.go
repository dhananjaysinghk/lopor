package semanticdiff

import (
	"testing"
)

func TestSemanticDiffIdenticalTexts(t *testing.T) {
	engine := NewEngine()

	text := "The vendor shall deliver the software within thirty days. All intellectual property remains with the customer."
	res, err := engine.Compare(CompareRequest{
		WorkspaceID:  "ws_legal",
		OriginalText: text,
		RevisedText:  text,
		Granularity:  GranularitySentence,
	})
	if err != nil {
		t.Fatalf("unexpected error comparing identical texts: %v", err)
	}

	if res.OverallSimilarity != 100.0 {
		t.Errorf("expected 100%% similarity, got %.1f%%", res.OverallSimilarity)
	}
	if len(res.Deltas) != 0 {
		t.Errorf("expected 0 deltas for identical texts, got %d", len(res.Deltas))
	}
	if res.RequiresHumanReview {
		t.Errorf("expected RequiresHumanReview to be false")
	}
}

func TestSemanticDiffNumericThresholdShift(t *testing.T) {
	engine := NewEngine()

	orig := "The service uptime SLA shall be 99.9% measured on a monthly basis."
	rev := "The service uptime SLA shall be 99.99% measured on a monthly basis."

	res, err := engine.Compare(CompareRequest{
		WorkspaceID:  "ws_legal",
		OriginalText: orig,
		RevisedText:  rev,
		Granularity:  GranularitySentence,
		DocType:      "CONTRACT",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(res.Deltas) != 1 {
		t.Fatalf("expected 1 delta, got %d", len(res.Deltas))
	}

	d := res.Deltas[0]
	if d.Category != CategoryThresholdChange {
		t.Errorf("expected CategoryThresholdChange, got %s", d.Category)
	}
	if d.Severity != SeverityHigh {
		t.Errorf("expected SeverityHigh for SLA change, got %s", d.Severity)
	}
}

func TestSemanticDiffRiskEscalation(t *testing.T) {
	engine := NewEngine()

	orig := "The vendor warrants that the software will perform substantially in accordance with the documentation."
	rev := "The vendor disclaims all warranties, express or implied, and shall have zero liability for consequential damages."

	res, err := engine.Compare(CompareRequest{
		WorkspaceID:  "ws_legal",
		OriginalText: orig,
		RevisedText:  rev,
		Granularity:  GranularitySentence,
		DocType:      "CONTRACT",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(res.Deltas) != 1 {
		t.Fatalf("expected 1 delta, got %d", len(res.Deltas))
	}

	d := res.Deltas[0]
	if d.Category != CategoryRiskEscalation {
		t.Errorf("expected CategoryRiskEscalation, got %s", d.Category)
	}
	if d.Severity != SeverityCritical {
		t.Errorf("expected SeverityCritical for liability disclaimer, got %s", d.Severity)
	}
	if !res.RequiresHumanReview {
		t.Errorf("expected RequiresHumanReview to be true")
	}
}

func TestSemanticDiffCodeRefactorVsDrift(t *testing.T) {
	engine := NewEngine()

	codeOrig := "func calculateTax(amount float64) float64 {\n    return amount * 0.15\n}"
	codeRefactor := "func calculateTax(subtotal float64) float64 {\n    // Clean variable renaming\n    return subtotal * 0.15\n}"

	resRefactor, err := engine.Compare(CompareRequest{
		WorkspaceID:  "ws_dev",
		OriginalText: codeOrig,
		RevisedText:  codeRefactor,
		Granularity:  GranularityParagraph,
		DocType:      "CODE",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resRefactor.Deltas) > 0 {
		d := resRefactor.Deltas[0]
		if d.Category != CategoryCodeRefactor {
			t.Errorf("expected CategoryCodeRefactor, got %s", d.Category)
		}
	}
}

