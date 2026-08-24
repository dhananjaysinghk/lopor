package rag

import "testing"

func TestComputeRRF(t *testing.T) {
	vecResults := []RankedItem{
		{ID: "doc_a", Text: "Doc A Content", Score: 0.95},
		{ID: "doc_b", Text: "Doc B Content", Score: 0.88},
	}

	textResults := []RankedItem{
		{ID: "doc_b", Text: "Doc B Content", Score: 12.4},
		{ID: "doc_c", Text: "Doc C Content", Score: 8.1},
	}

	fused := ComputeRRF(vecResults, textResults, 60.0)

	if len(fused) == 0 {
		t.Fatal("Expected RRF fusion results, got 0")
	}

	// doc_b appears in both rankings, so it should rank highest after fusion
	if fused[0].ID != "doc_b" {
		t.Errorf("Expected top fused result to be 'doc_b', got '%s'", fused[0].ID)
	}
}
