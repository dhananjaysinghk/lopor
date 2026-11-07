package reranker

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

var (
	ErrEmptyQuery  = errors.New("query cannot be empty")
	ErrEmptyChunks = errors.New("candidate chunks list cannot be empty")
)

// RerankerEngine coordinates cross-encoder relevance rescoring, semantic pruning, and lost-in-the-middle context placement.
type RerankerEngine struct {
	mu      sync.RWMutex
	metrics map[string]*RerankerMetrics
}

// NewEngine initializes a new RerankerEngine instance.
func NewEngine() *RerankerEngine {
	return &RerankerEngine{
		metrics: make(map[string]*RerankerMetrics),
	}
}

// ReRank scores, filters, budgets, and re-orders candidate document chunks.
func (e *RerankerEngine) ReRank(ctx context.Context, wsID string, req ReRankRequest) (*ReRankResult, error) {
	start := time.Now()

	trimmedQuery := strings.TrimSpace(req.Query)
	if trimmedQuery == "" {
		return nil, ErrEmptyQuery
	}
	if len(req.Chunks) == 0 {
		return nil, ErrEmptyChunks
	}

	// Apply sensible defaults
	maxTopK := req.MaxTopK
	if maxTopK <= 0 {
		maxTopK = 5
	}
	cutoff := req.RelevanceCutoff
	if cutoff <= 0 {
		cutoff = 0.35
	}
	maxTokens := req.MaxTokensBudget
	if maxTokens <= 0 {
		maxTokens = 4000
	}
	strategy := req.Strategy
	if strategy == "" {
		strategy = StrategyCrossEncoder
	}

	totalCandidates := len(req.Chunks)
	scoredList := make([]ScoredChunk, 0, totalCandidates)

	// Step 1: Score all candidates
	for idx, chunk := range req.Chunks {
		if chunk.TokenEstimate <= 0 {
			words := len(strings.Fields(chunk.Content))
			chunk.TokenEstimate = int(float64(words) * 1.33)
			if chunk.TokenEstimate < 10 {
				chunk.TokenEstimate = 10
			}
		}

		score, grade := ComputeRelevanceScore(trimmedQuery, chunk, strategy)

		scoredList = append(scoredList, ScoredChunk{
			DocumentChunk:     chunk,
			CrossEncoderScore: score,
			CombinedScore:     score,
			RelevanceGrade:    grade,
			InitialRank:       idx + 1,
		})
	}

	// Step 2: Sort descending by CombinedScore
	sort.Slice(scoredList, func(i, j int) bool {
		return scoredList[i].CombinedScore > scoredList[j].CombinedScore
	})

	// Step 3: Prune based on RelevanceCutoff, MaxTopK, and MaxTokensBudget
	var retained []ScoredChunk
	tokensAccumulated := 0

	for _, item := range scoredList {
		// Cutoff check
		if item.CombinedScore < cutoff {
			continue
		}

		// Token budget check
		if tokensAccumulated+item.TokenEstimate > maxTokens && len(retained) > 0 {
			continue
		}

		retained = append(retained, item)
		tokensAccumulated += item.TokenEstimate

		// Max top-k check
		if len(retained) >= maxTopK {
			break
		}
	}

	// Step 4: Apply Lost-In-The-Middle distribution if requested
	lostInMiddleApplied := false
	if req.LostInTheMiddleOrder && len(retained) > 2 {
		retained = ApplyLostInTheMiddle(retained)
		lostInMiddleApplied = true
	} else {
		for idx := range retained {
			retained[idx].FinalRank = idx + 1
			retained[idx].PositionDelta = retained[idx].InitialRank - retained[idx].FinalRank
		}
	}

	// Step 5: Assemble prompt context
	assembledContext := AssemblePromptContext(trimmedQuery, retained)

	retainedCount := len(retained)
	prunedCount := totalCandidates - retainedCount
	pruningRatio := (float64(prunedCount) / float64(totalCandidates)) * 100.0

	elapsed := float64(time.Since(start).Microseconds()) / 1000.0
	if elapsed < 0.05 {
		elapsed = 0.05
	}

	reqID := generateID("rrk")

	result := &ReRankResult{
		RequestID:              reqID,
		WorkspaceID:            wsID,
		Query:                  trimmedQuery,
		StrategyUsed:           strategy,
		TotalCandidateChunks:   totalCandidates,
		RetainedChunksCount:    retainedCount,
		PrunedChunksCount:      prunedCount,
		PruningRatioPercent:    pruningRatio,
		TotalTokensRetained:    tokensAccumulated,
		ExecutionDurationMs:    elapsed,
		LostInTheMiddleApplied: lostInMiddleApplied,
		ScoredChunks:           retained,
		AssembledPromptContext: assembledContext,
		ProcessedAt:            time.Now().UTC(),
	}

	// Step 6: Record telemetry metrics
	e.recordMetrics(wsID, totalCandidates, retainedCount, prunedCount, elapsed)

	return result, nil
}

// GetMetrics returns workspace re-ranking metrics.
func (e *RerankerEngine) GetMetrics(wsID string) *RerankerMetrics {
	e.mu.RLock()
	defer e.mu.RUnlock()

	m, exists := e.metrics[wsID]
	if !exists {
		return &RerankerMetrics{
			WorkspaceID: wsID,
			LastUpdated: time.Now().UTC(),
		}
	}
	return m
}

func (e *RerankerEngine) recordMetrics(wsID string, candidates, retained, pruned int, durationMs float64) {
	if wsID == "" {
		return
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	m, exists := e.metrics[wsID]
	if !exists {
		m = &RerankerMetrics{
			WorkspaceID: wsID,
			LastUpdated: time.Now().UTC(),
		}
		e.metrics[wsID] = m
	}

	m.TotalRequests++
	m.CumulativeCandidates += int64(candidates)
	m.CumulativeRetained += int64(retained)
	m.CumulativePruned += int64(pruned)
	m.LastUpdated = time.Now().UTC()

	if m.CumulativeCandidates > 0 {
		m.AveragePruningRatioPercent = (float64(m.CumulativePruned) / float64(m.CumulativeCandidates)) * 100.0
	}
	if m.TotalRequests > 0 {
		m.AverageExecutionDurationMs = (m.AverageExecutionDurationMs*float64(m.TotalRequests-1) + durationMs) / float64(m.TotalRequests)
	}
}

func generateID(prefix string) string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%s_%s", prefix, hex.EncodeToString(b))
}

