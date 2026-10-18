package llmjudge

import (
	"context"
	"errors"
	"time"
)

// LLMJudgeService coordinates automated LLM-as-a-Judge evaluations and competitive Elo tournaments.
type LLMJudgeService struct {
	judge *JudgeEngine
}

// NewLLMJudgeService initializes a new judge service instance.
func NewLLMJudgeService() *LLMJudgeService {
	return &LLMJudgeService{
		judge: NewJudgeEngine(),
	}
}

// JudgePointwise evaluates a single completion against configured rubric criteria.
func (s *LLMJudgeService) JudgePointwise(ctx context.Context, req PointwiseJudgeRequest) (*PointwiseJudgeResult, error) {
	if req.CandidateOutput == "" {
		return nil, errors.New("candidate output cannot be empty")
	}

	res := s.judge.PointwiseScore(req.Prompt, req.CandidateOutput, req.ReferenceContext, req.Criteria, req.PassThreshold)
	return &res, nil
}

// JudgePairwise compares two completions with position-bias mitigation and updates competitive ratings.
func (s *LLMJudgeService) JudgePairwise(ctx context.Context, req PairwiseJudgeRequest) (*PairwiseJudgeResult, error) {
	if req.CandidateA == "" || req.CandidateB == "" {
		return nil, errors.New("both candidate completions are required for pairwise comparison")
	}

	nameA := req.CandidateAName
	if nameA == "" {
		nameA = "Model A"
	}
	nameB := req.CandidateBName
	if nameB == "" {
		nameB = "Model B"
	}

	res := s.judge.PairwiseCompare(req.Prompt, req.CandidateA, req.CandidateB, nameA, nameB, req.ReferenceContext, req.Criteria)

	// Update workspace Elo standings
	if req.WorkspaceID != "" {
		s.judge.UpdateElo(req.WorkspaceID, nameA, nameB, res.Verdict)
	}

	return &res, nil
}

// RunTournament runs a round-robin tournament across multiple model candidates on a benchmark query.
func (s *LLMJudgeService) RunTournament(ctx context.Context, req TournamentRequest) (*TournamentResult, error) {
	if len(req.Candidates) < 2 {
		return nil, errors.New("at least 2 candidate completions required for tournament evaluation")
	}

	models := make([]string, 0, len(req.Candidates))
	for m := range req.Candidates {
		models = append(models, m)
	}

	matches := 0
	for i := 0; i < len(models); i++ {
		for j := i + 1; j < len(models); j++ {
			mA := models[i]
			mB := models[j]
			outA := req.Candidates[mA]
			outB := req.Candidates[mB]

			res := s.judge.PairwiseCompare(req.Prompt, outA, outB, mA, mB, req.ReferenceContext, nil)
			s.judge.UpdateElo(req.WorkspaceID, mA, mB, res.Verdict)
			matches++
		}
	}

	rankings := s.judge.GetLeaderboard(req.WorkspaceID)

	return &TournamentResult{
		WorkspaceID:     req.WorkspaceID,
		Rankings:        rankings,
		MatchesExecuted: matches,
		ExecutedAt:      time.Now().UTC(),
	}, nil
}

// GetLeaderboard returns the current competitive model leaderboard for a workspace.
func (s *LLMJudgeService) GetLeaderboard(ctx context.Context, workspaceID string) []EloLeaderboardEntry {
	if workspaceID == "" {
		workspaceID = "global"
	}
	list := s.judge.GetLeaderboard(workspaceID)
	if len(list) == 0 {
		// Seed baseline frontier models if uninitialized
		s.judge.UpdateElo(workspaceID, "gpt-4o", "claude-3-5-sonnet", VerdictTie)
		s.judge.UpdateElo(workspaceID, "claude-3-5-sonnet", "deepseek-coder", VerdictAWins)
		list = s.judge.GetLeaderboard(workspaceID)
	}
	return list
}
