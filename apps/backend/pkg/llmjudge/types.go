package llmjudge

import (
	"time"
)

// JudgeCriterion defines evaluation dimensions.
type JudgeCriterion string

const (
	CriterionFaithfulness JudgeCriterion = "FAITHFULNESS"
	CriterionRelevance    JudgeCriterion = "RELEVANCE"
	CriterionCoherence    JudgeCriterion = "COHERENCE"
	CriterionSafety       JudgeCriterion = "SAFETY"
	CriterionConciseness  JudgeCriterion = "CONCISENESS"
	CriterionCustom       JudgeCriterion = "CUSTOM"
)

// CriterionScore records the score and rationale for a specific evaluation dimension.
type CriterionScore struct {
	Criterion  JudgeCriterion `json:"criterion"`
	Score      float64        `json:"score"`      // 1.0 to 10.0 scale
	Weight     float64        `json:"weight"`     // relative weighting
	Rationale  string         `json:"rationale"`
	Strengths  []string       `json:"strengths,omitempty"`
	Weaknesses []string       `json:"weaknesses,omitempty"`
}

// PointwiseJudgeRequest provides an output to be scored against criteria.
type PointwiseJudgeRequest struct {
	WorkspaceID      string           `json:"workspace_id"`
	Prompt           string           `json:"prompt"`
	CandidateOutput  string           `json:"candidate_output"`
	ReferenceContext string           `json:"reference_context,omitempty"`
	Criteria         []JudgeCriterion `json:"criteria,omitempty"`
	CustomRubric     string           `json:"custom_rubric,omitempty"`
	PassThreshold    float64          `json:"pass_threshold,omitempty"` // default 7.0
}

// PointwiseJudgeResult conveys the graded results with breakdown and commentary.
type PointwiseJudgeResult struct {
	OverallScore    float64          `json:"overall_score"` // 1.0 to 10.0 scale
	Passed          bool             `json:"passed"`
	CriteriaScores  []CriterionScore `json:"criteria_scores"`
	SummaryFeedback string           `json:"summary_feedback"`
	JudgedAt        time.Time        `json:"judged_at"`
}

// PairwiseVerdict denotes the comparative outcome between two candidates.
type PairwiseVerdict string

const (
	VerdictAWins PairwiseVerdict = "A_WINS"
	VerdictBWins PairwiseVerdict = "B_WINS"
	VerdictTie   PairwiseVerdict = "TIE"
)

// PairwiseJudgeRequest provides two candidate completions for blind A/B grading.
type PairwiseJudgeRequest struct {
	WorkspaceID      string           `json:"workspace_id"`
	Prompt           string           `json:"prompt"`
	CandidateA       string           `json:"candidate_a"`
	CandidateB       string           `json:"candidate_b"`
	CandidateAName   string           `json:"candidate_a_name,omitempty"`
	CandidateBName   string           `json:"candidate_b_name,omitempty"`
	ReferenceContext string           `json:"reference_context,omitempty"`
	Criteria         []JudgeCriterion `json:"criteria,omitempty"`
}

// PairwiseJudgeResult provides the comparative evaluation outcome and bias analysis.
type PairwiseJudgeResult struct {
	Verdict               PairwiseVerdict `json:"verdict"`
	WinnerName            string          `json:"winner_name,omitempty"`
	Confidence            float64         `json:"confidence"`
	ComparisonRationale   string          `json:"comparison_rationale"`
	PositionBiasMitigated bool            `json:"position_bias_mitigated"`
	ScoreA                float64         `json:"score_a"`
	ScoreB                float64         `json:"score_b"`
	JudgedAt              time.Time       `json:"judged_at"`
}

// EloLeaderboardEntry represents a model's competitive ranking in a workspace.
type EloLeaderboardEntry struct {
	ModelID       string    `json:"model_id"`
	EloRating     float64   `json:"elo_rating"`
	MatchesPlayed int       `json:"matches_played"`
	Wins          int       `json:"wins"`
	Losses        int       `json:"losses"`
	Ties          int       `json:"ties"`
	WinRate       float64   `json:"win_rate"`
	LastUpdated   time.Time `json:"last_updated"`
}

// TournamentRequest runs a round-robin benchmark among multiple candidates.
type TournamentRequest struct {
	WorkspaceID      string            `json:"workspace_id"`
	Prompt           string            `json:"prompt"`
	Candidates       map[string]string `json:"candidates"` // modelID -> output
	ReferenceContext string            `json:"reference_context,omitempty"`
}

// TournamentResult summarizes the tournament outcomes and updated standings.
type TournamentResult struct {
	WorkspaceID     string                `json:"workspace_id"`
	Rankings        []EloLeaderboardEntry `json:"rankings"`
	MatchesExecuted int                   `json:"matches_executed"`
	ExecutedAt      time.Time             `json:"executed_at"`
}
