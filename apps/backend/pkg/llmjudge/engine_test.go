package llmjudge

import (
	"context"
	"testing"
)

func TestPointwiseJudgeCleanOutput(t *testing.T) {
	service := NewLLMJudgeService()
	ctx := context.Background()

	prompt := "How do I implement binary search in Go?"
	contextText := "Binary search is an algorithm with O(log n) time complexity that operates on sorted arrays."
	candidateOutput := "```go\nfunc binarySearch(arr []int, target int) int {\n    left, right := 0, len(arr)-1\n    for left <= right {\n        mid := left + (right-left)/2\n        if arr[mid] == target { return mid }\n        if arr[mid] < target { left = mid + 1 } else { right = mid - 1 }\n    }\n    return -1\n}\n```\nBinary search runs in O(log n) time complexity on sorted slices."

	req := PointwiseJudgeRequest{
		WorkspaceID:      "ws-judge-1",
		Prompt:           prompt,
		CandidateOutput:  candidateOutput,
		ReferenceContext: contextText,
		PassThreshold:    7.0,
	}

	result, err := service.JudgePointwise(ctx, req)
	if err != nil {
		t.Fatalf("JudgePointwise failed: %v", err)
	}

	if !result.Passed {
		t.Errorf("expected clean output to pass evaluation, got overall score: %.2f", result.OverallScore)
	}
	if result.OverallScore < 7.0 {
		t.Errorf("expected score >= 7.0, got %.2f", result.OverallScore)
	}
	if len(result.CriteriaScores) == 0 {
		t.Fatalf("expected criteria scores breakdown")
	}
}

func TestPairwiseBlindComparison(t *testing.T) {
	service := NewLLMJudgeService()
	ctx := context.Background()

	prompt := "Explain idempotency in REST APIs"
	candA := "An idempotent HTTP method can be executed multiple times producing the exact same server resource state. Examples include GET, PUT, and DELETE, while POST is not idempotent."
	candB := "idempotency means same thing."

	req := PairwiseJudgeRequest{
		WorkspaceID:    "ws-judge-1",
		Prompt:         prompt,
		CandidateA:     candA,
		CandidateB:     candB,
		CandidateAName: "Model A (Detailed)",
		CandidateBName: "Model B (Brief)",
	}

	result, err := service.JudgePairwise(ctx, req)
	if err != nil {
		t.Fatalf("JudgePairwise failed: %v", err)
	}

	if result.Verdict != VerdictAWins {
		t.Errorf("expected Model A to win pairwise comparison, got %s", result.Verdict)
	}
	if !result.PositionBiasMitigated {
		t.Errorf("expected position bias mitigation flag to be true")
	}
	if result.ScoreA <= result.ScoreB {
		t.Errorf("expected ScoreA (%.1f) > ScoreB (%.1f)", result.ScoreA, result.ScoreB)
	}
}

func TestEloLeaderboardUpdates(t *testing.T) {
	service := NewLLMJudgeService()
	ctx := context.Background()

	wsID := "ws-elo-test"
	req := PairwiseJudgeRequest{
		WorkspaceID:    wsID,
		Prompt:         "Write a Fibonacci function in Go",
		CandidateA:     "```go\nfunc Fib(n int) int {\n    if n <= 1 { return n }\n    return Fib(n-1) + Fib(n-2)\n}\n```",
		CandidateB:     "fibonacci bad code",
		CandidateAName: "model-alpha",
		CandidateBName: "model-beta",
	}

	res, err := service.JudgePairwise(ctx, req)
	if err != nil {
		t.Fatalf("JudgePairwise failed: %v", err)
	}

	if res.Verdict != VerdictAWins {
		t.Fatalf("expected model-alpha to win")
	}

	board := service.GetLeaderboard(ctx, wsID)
	if len(board) < 2 {
		t.Fatalf("expected at least 2 models on leaderboard, got %d", len(board))
	}

	// First model should be model-alpha with Elo > 1500
	if board[0].ModelID != "model-alpha" {
		t.Errorf("expected model-alpha to rank 1st, got %s", board[0].ModelID)
	}
	if board[0].EloRating <= 1500.0 {
		t.Errorf("expected model-alpha Elo > 1500, got %.1f", board[0].EloRating)
	}
}

func TestTournamentBenchmark(t *testing.T) {
	service := NewLLMJudgeService()
	ctx := context.Background()

	req := TournamentRequest{
		WorkspaceID: "ws-tournament-1",
		Prompt:      "Explain the differences between PostgreSQL and Redis",
		Candidates: map[string]string{
			"model-gpt4":   "PostgreSQL is a relational database optimized for ACID transactions, complex SQL queries, and vector search with pgvector. Redis is an in-memory key-value data store used for sub-millisecond caching and pub/sub queues.",
			"model-claude": "PostgreSQL provides relational storage with tables and indexes. Redis is an in-memory cache supporting data structures like hashes, lists, and sets.",
			"model-weak":   "both are dbs.",
		},
	}

	res, err := service.RunTournament(ctx, req)
	if err != nil {
		t.Fatalf("RunTournament failed: %v", err)
	}

	if res.MatchesExecuted != 3 {
		t.Errorf("expected 3 round-robin matches for 3 models, got %d", res.MatchesExecuted)
	}
	if len(res.Rankings) != 3 {
		t.Errorf("expected 3 models in tournament rankings, got %d", len(res.Rankings))
	}
}
