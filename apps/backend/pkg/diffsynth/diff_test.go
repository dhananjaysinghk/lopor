package diffsynth_test

import (
	"strings"
	"testing"

	"github.com/lopor-ai/lopor/pkg/diffsynth"
)

func TestSynthesizeDiff(t *testing.T) {
	synth := diffsynth.NewSynthesizer()

	orig := "func compute(a int) int {\n\treturn a * 2\n}"
	mod := "func compute(a int) int {\n\t// Optimized calculation\n\treturn a << 1\n}"

	req := diffsynth.DiffRequest{
		FileName:       "math.go",
		OriginalCode:   orig,
		ModifiedCode:   mod,
		RefactorIntent: diffsynth.CategoryPerformance,
	}

	res, err := synth.SynthesizeDiff(req)
	if err != nil {
		t.Fatalf("unexpected error synthesizing diff: %v", err)
	}

	if res.FileName != "math.go" {
		t.Errorf("expected math.go, got %s", res.FileName)
	}

	if res.LinesAdded == 0 || res.LinesRemoved == 0 {
		t.Errorf("expected added and removed lines, got added=%d, removed=%d", res.LinesAdded, res.LinesRemoved)
	}

	if !strings.Contains(res.UnifiedDiff, "--- a/math.go") || !strings.Contains(res.UnifiedDiff, "+++ b/math.go") {
		t.Errorf("expected unified git diff headers in result")
	}

	if res.RefactorCategory != diffsynth.CategoryPerformance {
		t.Errorf("expected category performance, got %s", res.RefactorCategory)
	}
}
