package promptopt

import (
	"time"
)

// OptimizationMethod defines the algorithmic technique used to improve the prompt.
type OptimizationMethod string

const (
	MethodBootstrapFewShot     OptimizationMethod = "BOOTSTRAP_FEWSHOT"
	MethodInstructionMutation  OptimizationMethod = "INSTRUCTION_MUTATION"
	MethodHardNegativeInject   OptimizationMethod = "HARD_NEGATIVE_INJECTION"
	MethodHybridDSPy           OptimizationMethod = "HYBRID_DSPY"
)

// OptimizationMetric specifies the primary objective function for scoring prompt candidates.
type OptimizationMetric string

const (
	MetricAccuracy         OptimizationMetric = "ACCURACY"
	MetricFormatCompliance OptimizationMetric = "FORMAT_COMPLIANCE"
	MetricTokenEfficiency  OptimizationMetric = "TOKEN_EFFICIENCY"
	MetricComposite        OptimizationMetric = "COMPOSITE"
)

// JobStatus tracks the state of an optimization tournament.
type JobStatus string

const (
	StatusPending   JobStatus = "PENDING"
	StatusRunning   JobStatus = "RUNNING"
	StatusCompleted JobStatus = "COMPLETED"
	StatusFailed    JobStatus = "FAILED"
)

// PromptExample represents an input-output demonstration or test case.
type PromptExample struct {
	ID                 string            `json:"id"`
	InputVariables     map[string]string `json:"input_variables"`
	TargetOutput       string            `json:"target_output"`
	Rationale          string            `json:"rationale,omitempty"`
	Score              float64           `json:"score,omitempty"`
	IsNegativeExemplar bool              `json:"is_negative_exemplar,omitempty"` // Example of what NOT to do
}

// OptimizationCandidate represents a mutated or augmented prompt variant.
type OptimizationCandidate struct {
	CandidateID        string             `json:"candidate_id"`
	Iteration          int                `json:"iteration"`
	MutationStrategy   OptimizationMethod `json:"mutation_strategy"`
	SystemInstructions string             `json:"system_instructions"`
	FewShotExemplars   []PromptExample    `json:"few_shot_exemplars,omitempty"`
	AntiPatterns       []string           `json:"anti_patterns,omitempty"`
	CompiledPrompt     string             `json:"compiled_prompt"`
	BenchmarkScore     float64            `json:"benchmark_score"` // 0.0 - 100.0%
	EstimatedTokens    int                `json:"estimated_tokens"`
	ScoreDelta         float64            `json:"score_delta"` // Relative to baseline
}

// OptimizationJob encapsulates an end-to-end prompt auto-tuning experiment.
type OptimizationJob struct {
	ID               string                  `json:"id"`
	WorkspaceID      string                  `json:"workspace_id"`
	Title            string                  `json:"title"`
	BasePrompt       string                  `json:"base_prompt"`
	Method           OptimizationMethod      `json:"method"`
	Metric           OptimizationMetric      `json:"metric"`
	Status           JobStatus               `json:"status"`
	MaxIterations    int                     `json:"max_iterations"`
	IterationsRun    int                     `json:"iterations_run"`
	BaselineScore    float64                 `json:"baseline_score"`
	BestScore        float64                 `json:"best_score"`
	ScoreImprovement float64                 `json:"score_improvement"`
	Candidates       []OptimizationCandidate `json:"candidates"`
	BestCandidate    *OptimizationCandidate  `json:"best_candidate,omitempty"`
	CreatedAt        time.Time               `json:"created_at"`
	UpdatedAt        time.Time               `json:"updated_at"`
}

// StartOptimizationRequest initiates a new prompt tuning tournament.
type StartOptimizationRequest struct {
	WorkspaceID   string             `json:"workspace_id"`
	Title         string             `json:"title"`
	BasePrompt    string             `json:"base_prompt"`
	Method        OptimizationMethod `json:"method,omitempty"`
	Metric        OptimizationMetric `json:"metric,omitempty"`
	Examples      []PromptExample    `json:"examples"`
	MaxIterations int                `json:"max_iterations,omitempty"` // Default 3
}

// OptimizationReport provides an executive overview of prompt auto-tuning results.
type OptimizationReport struct {
	JobID            string                 `json:"job_id"`
	Title            string                 `json:"title"`
	Method           OptimizationMethod     `json:"method"`
	BaselineScore    float64                `json:"baseline_score"`
	OptimizedScore   float64                `json:"optimized_score"`
	ScoreImprovement float64                `json:"score_improvement"`
	TotalCandidates  int                    `json:"total_candidates"`
	WinningCandidate *OptimizationCandidate `json:"winning_candidate"`
	OptimizationDiff string                 `json:"optimization_diff"`
	Recommendations  []string               `json:"recommendations"`
}
