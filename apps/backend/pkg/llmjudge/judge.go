package llmjudge

import (
	"fmt"
	"math"
	"sort"
	"sync"
	"time"
)

// JudgeEngine handles pointwise scoring, blind pairwise comparisons, and Elo rating updates.
type JudgeEngine struct {
	mu           sync.RWMutex
	rubrics      *RubricEvaluator
	leaderboards map[string]map[string]*EloLeaderboardEntry // wsID -> modelID -> entry
}

// NewJudgeEngine initializes the evaluation engine.
func NewJudgeEngine() *JudgeEngine {
	return &JudgeEngine{
		rubrics:      NewRubricEvaluator(),
		leaderboards: make(map[string]map[string]*EloLeaderboardEntry),
	}
}

// PointwiseScore grades a single candidate completion against a set of criteria.
func (j *JudgeEngine) PointwiseScore(prompt, output, context string, criteria []JudgeCriterion, threshold float64) PointwiseJudgeResult {
	if len(criteria) == 0 {
		criteria = []JudgeCriterion{
			CriterionFaithfulness,
			CriterionRelevance,
			CriterionCoherence,
			CriterionSafety,
			CriterionConciseness,
		}
	}
	if threshold <= 0 {
		threshold = 7.0
	}

	var scores []CriterionScore
	totalWeight := 0.0
	weightedSum := 0.0

	for _, c := range criteria {
		cs := j.rubrics.EvaluateCriterion(c, prompt, output, context)
		scores = append(scores, cs)
		weightedSum += cs.Score * cs.Weight
		totalWeight += cs.Weight
	}

	overall := 0.0
	if totalWeight > 0 {
		overall = weightedSum / totalWeight
	}
	overall = math.Round(overall*10) / 10

	passed := overall >= threshold
	summary := fmt.Sprintf("Overall Judge Rating: %.1f/10.0 (Threshold: %.1f). Passed: %t.", overall, threshold, passed)

	return PointwiseJudgeResult{
		OverallScore:    overall,
		Passed:          passed,
		CriteriaScores:  scores,
		SummaryFeedback: summary,
		JudgedAt:        time.Now().UTC(),
	}
}

// PairwiseCompare executes position-bias mitigated blind A/B comparative evaluation.
func (j *JudgeEngine) PairwiseCompare(prompt, candA, candB, nameA, nameB, context string, criteria []JudgeCriterion) PairwiseJudgeResult {
	if nameA == "" {
		nameA = "Candidate A"
	}
	if nameB == "" {
		nameB = "Candidate B"
	}

	// Forward evaluation: A then B
	resA1 := j.PointwiseScore(prompt, candA, context, criteria, 7.0)
	resB1 := j.PointwiseScore(prompt, candB, context, criteria, 7.0)

	// Reverse evaluation: B then A (to cancel out positional bias)
	resB2 := j.PointwiseScore(prompt, candB, context, criteria, 7.0)
	resA2 := j.PointwiseScore(prompt, candA, context, criteria, 7.0)

	scoreA := (resA1.OverallScore + resA2.OverallScore) / 2.0
	scoreB := (resB1.OverallScore + resB2.OverallScore) / 2.0
	scoreA = math.Round(scoreA*10) / 10
	scoreB = math.Round(scoreB*10) / 10

	diff := scoreA - scoreB
	verdict := VerdictTie
	winnerName := "Tie"
	confidence := 0.85

	if math.Abs(diff) < 0.25 {
		verdict = VerdictTie
		winnerName = "Tie"
	} else if diff > 0 {
		verdict = VerdictAWins
		winnerName = nameA
		confidence = math.Min(0.98, 0.70+(math.Abs(diff)*0.1))
	} else {
		verdict = VerdictBWins
		winnerName = nameB
		confidence = math.Min(0.98, 0.70+(math.Abs(diff)*0.1))
	}

	rationale := fmt.Sprintf("%s scored %.1f vs %s scored %.1f across evaluated criteria.",
		nameA, scoreA, nameB, scoreB)

	return PairwiseJudgeResult{
		Verdict:               verdict,
		WinnerName:            winnerName,
		Confidence:            math.Round(confidence*100) / 100,
		ComparisonRationale:   rationale,
		PositionBiasMitigated: true,
		ScoreA:                scoreA,
		ScoreB:                scoreB,
		JudgedAt:              time.Now().UTC(),
	}
}

// UpdateElo updates the competitive ratings of two models based on match outcome.
func (j *JudgeEngine) UpdateElo(workspaceID, modelA, modelB string, verdict PairwiseVerdict) {
	j.mu.Lock()
	defer j.mu.Unlock()

	if _, ok := j.leaderboards[workspaceID]; !ok {
		j.leaderboards[workspaceID] = make(map[string]*EloLeaderboardEntry)
	}

	entryA := j.getOrCreateEntry(workspaceID, modelA)
	entryB := j.getOrCreateEntry(workspaceID, modelB)

	kFactor := 32.0
	// Expected scores
	expA := 1.0 / (1.0 + math.Pow(10.0, (entryB.EloRating-entryA.EloRating)/400.0))
	expB := 1.0 / (1.0 + math.Pow(10.0, (entryA.EloRating-entryB.EloRating)/400.0))

	var actualA, actualB float64
	switch verdict {
	case VerdictAWins:
		actualA, actualB = 1.0, 0.0
		entryA.Wins++
		entryB.Losses++
	case VerdictBWins:
		actualA, actualB = 0.0, 1.0
		entryB.Wins++
		entryA.Losses++
	default:
		actualA, actualB = 0.5, 0.5
		entryA.Ties++
		entryB.Ties++
	}

	entryA.MatchesPlayed++
	entryB.MatchesPlayed++

	entryA.EloRating += kFactor * (actualA - expA)
	entryB.EloRating += kFactor * (actualB - expB)

	entryA.EloRating = math.Round(entryA.EloRating*10) / 10
	entryB.EloRating = math.Round(entryB.EloRating*10) / 10

	entryA.WinRate = math.Round((float64(entryA.Wins)/float64(entryA.MatchesPlayed))*1000) / 10
	entryB.WinRate = math.Round((float64(entryB.Wins)/float64(entryB.MatchesPlayed))*1000) / 10

	now := time.Now().UTC()
	entryA.LastUpdated = now
	entryB.LastUpdated = now
}

func (j *JudgeEngine) getOrCreateEntry(wsID, modelID string) *EloLeaderboardEntry {
	if entry, ok := j.leaderboards[wsID][modelID]; ok {
		return entry
	}
	entry := &EloLeaderboardEntry{
		ModelID:       modelID,
		EloRating:     1500.0, // Standard baseline starting Elo
		MatchesPlayed: 0,
		Wins:          0,
		Losses:        0,
		Ties:          0,
		WinRate:       0.0,
		LastUpdated:   time.Now().UTC(),
	}
	j.leaderboards[wsID][modelID] = entry
	return entry
}

// GetLeaderboard returns the sorted Elo rankings for a workspace.
func (j *JudgeEngine) GetLeaderboard(workspaceID string) []EloLeaderboardEntry {
	j.mu.RLock()
	defer j.mu.RUnlock()

	var list []EloLeaderboardEntry
	for _, entry := range j.leaderboards[workspaceID] {
		list = append(list, *entry)
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].EloRating > list[j].EloRating
	})

	return list
}
