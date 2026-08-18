package rag

import (
	"sort"
)

type RankedItem struct {
	ID        string  `json:"id"`
	Text      string  `json:"text"`
	Score     float64 `json:"score"`
	SourceDoc string  `json:"source_doc"`
}

// ComputeRRF calculates Reciprocal Rank Fusion scores from dense vector results and sparse BM25 text results
func ComputeRRF(vectorResults []RankedItem, textResults []RankedItem, k float64) []RankedItem {
	if k <= 0 {
		k = 60.0 // Standard RRF constant
	}

	rrfMap := make(map[string]*RankedItem)
	scores := make(map[string]float64)

	// Process Vector Rankings
	for rank, item := range vectorResults {
		id := item.ID
		if _, exists := rrfMap[id]; !exists {
			itemCopy := item
			rrfMap[id] = &itemCopy
		}
		scores[id] += 1.0 / (k + float64(rank+1))
	}

	// Process Sparse Text Rankings
	for rank, item := range textResults {
		id := item.ID
		if _, exists := rrfMap[id]; !exists {
			itemCopy := item
			rrfMap[id] = &itemCopy
		}
		scores[id] += 1.0 / (k + float64(rank+1))
	}

	// Build Final Ranked List
	var finalResults []RankedItem
	for id, item := range rrfMap {
		item.Score = scores[id]
		finalResults = append(finalResults, *item)
	}

	// Sort Descending by RRF Score
	sort.Slice(finalResults, func(i, j int) bool {
		return finalResults[i].Score > finalResults[j].Score
	})

	return finalResults
}
