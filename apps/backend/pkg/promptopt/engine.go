package promptopt

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrJobNotFound    = errors.New("prompt optimization job not found")
	ErrInvalidRequest = errors.New("invalid prompt optimization request parameters")
)

// PromptOptEngine manages continuous prompt auto-tuning, few-shot bootstrapping, and candidate tournaments.
type PromptOptEngine struct {
	mu   sync.RWMutex
	jobs map[string]*OptimizationJob
}

// NewEngine initializes a new instance of the PromptOptEngine.
func NewEngine() *PromptOptEngine {
	return &PromptOptEngine{
		jobs: make(map[string]*OptimizationJob),
	}
}

// StartOptimization executes a multi-strategy prompt optimization tournament.
func (e *PromptOptEngine) StartOptimization(req StartOptimizationRequest) (*OptimizationJob, error) {
	if req.WorkspaceID == "" || req.Title == "" || req.BasePrompt == "" {
		return nil, fmt.Errorf("%w: workspace_id, title, and base_prompt are required", ErrInvalidRequest)
	}

	method := req.Method
	if method == "" {
		method = MethodHybridDSPy
	}

	metric := req.Metric
	if metric == "" {
		metric = MetricComposite
	}

	maxIters := req.MaxIterations
	if maxIters <= 0 {
		maxIters = 3
	}

	jobID := generateID("opt")
	now := time.Now().UTC()

	// 1. Calculate baseline candidate
	baselineCandidate := &OptimizationCandidate{
		CandidateID:        generateID("cand_base"),
		Iteration:          0,
		MutationStrategy:   "BASELINE",
		SystemInstructions: req.BasePrompt,
		CompiledPrompt:     req.BasePrompt,
		EstimatedTokens:    EstimateTokenCount(req.BasePrompt),
	}
	baselineScore := EvaluateCandidate(baselineCandidate, req.Examples, metric, 0.0)
	baselineCandidate.BenchmarkScore = baselineScore
	baselineCandidate.ScoreDelta = 0.0

	candidates := []OptimizationCandidate{*baselineCandidate}

	// 2. Execute Tournament Mutations
	// Candidate A: Instruction Clarification & Constraints
	clarified := MutateInstructionClarification(req.BasePrompt)
	candA := OptimizationCandidate{
		CandidateID:        generateID("cand_mut"),
		Iteration:          1,
		MutationStrategy:   MethodInstructionMutation,
		SystemInstructions: clarified,
		CompiledPrompt:     clarified,
		EstimatedTokens:    EstimateTokenCount(clarified),
	}
	candA.BenchmarkScore = EvaluateCandidate(&candA, req.Examples, metric, baselineScore)
	candA.ScoreDelta = candA.BenchmarkScore - baselineScore
	candidates = append(candidates, candA)

	// Candidate B: Few-Shot Bootstrap
	fewShots := BootstrapFewShots(req.Examples, 2)
	compiledB := CompileCandidatePrompt(req.BasePrompt, fewShots, nil)
	candB := OptimizationCandidate{
		CandidateID:        generateID("cand_fewshot"),
		Iteration:          2,
		MutationStrategy:   MethodBootstrapFewShot,
		SystemInstructions: req.BasePrompt,
		FewShotExemplars:   fewShots,
		CompiledPrompt:     compiledB,
		EstimatedTokens:    EstimateTokenCount(compiledB),
	}
	candB.BenchmarkScore = EvaluateCandidate(&candB, req.Examples, metric, baselineScore)
	candB.ScoreDelta = candB.BenchmarkScore - baselineScore
	candidates = append(candidates, candB)

	// Candidate C: Hard Negative Injection
	antiPatterns := ExtractHardNegatives(req.Examples)
	compiledC := CompileCandidatePrompt(req.BasePrompt, nil, antiPatterns)
	candC := OptimizationCandidate{
		CandidateID:        generateID("cand_neg"),
		Iteration:          3,
		MutationStrategy:   MethodHardNegativeInject,
		SystemInstructions: req.BasePrompt,
		AntiPatterns:       antiPatterns,
		CompiledPrompt:     compiledC,
		EstimatedTokens:    EstimateTokenCount(compiledC),
	}
	candC.BenchmarkScore = EvaluateCandidate(&candC, req.Examples, metric, baselineScore)
	candC.ScoreDelta = candC.BenchmarkScore - baselineScore
	candidates = append(candidates, candC)

	// Candidate D: Hybrid DSPy Composite (Best Instructions + Few-Shots + Anti-Patterns)
	reasoningInstructions := MutateReasoningScaffold(req.BasePrompt)
	compiledD := CompileCandidatePrompt(reasoningInstructions, fewShots, antiPatterns)
	candD := OptimizationCandidate{
		CandidateID:        generateID("cand_hybrid"),
		Iteration:          4,
		MutationStrategy:   MethodHybridDSPy,
		SystemInstructions: reasoningInstructions,
		FewShotExemplars:   fewShots,
		AntiPatterns:       antiPatterns,
		CompiledPrompt:     compiledD,
		EstimatedTokens:    EstimateTokenCount(compiledD),
	}
	candD.BenchmarkScore = EvaluateCandidate(&candD, req.Examples, metric, baselineScore)
	candD.ScoreDelta = candD.BenchmarkScore - baselineScore
	candidates = append(candidates, candD)

	// 3. Select Winning Candidate
	var bestCandidate *OptimizationCandidate
	bestScore := baselineScore

	for i := range candidates {
		c := &candidates[i]
		if c.BenchmarkScore > bestScore {
			bestScore = c.BenchmarkScore
			bestCandidate = c
		}
	}

	if bestCandidate == nil {
		bestCandidate = baselineCandidate
	}

	improvement := bestScore - baselineScore

	job := &OptimizationJob{
		ID:               jobID,
		WorkspaceID:      req.WorkspaceID,
		Title:            req.Title,
		BasePrompt:       req.BasePrompt,
		Method:           method,
		Metric:           metric,
		Status:           StatusCompleted,
		MaxIterations:    maxIters,
		IterationsRun:    len(candidates) - 1,
		BaselineScore:    baselineScore,
		BestScore:        bestScore,
		ScoreImprovement: improvement,
		Candidates:       candidates,
		BestCandidate:    bestCandidate,
		CreatedAt:        now,
		UpdatedAt:        time.Now().UTC(),
	}

	e.mu.Lock()
	e.jobs[jobID] = job
	e.mu.Unlock()

	return job, nil
}

// GetJob retrieves an optimization job by ID with workspace tenancy check.
func (e *PromptOptEngine) GetJob(workspaceID, jobID string) (*OptimizationJob, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	job, exists := e.jobs[jobID]
	if !exists || job.WorkspaceID != workspaceID {
		return nil, ErrJobNotFound
	}

	return job, nil
}

// ListJobs returns all prompt optimization jobs in the specified workspace.
func (e *PromptOptEngine) ListJobs(workspaceID string) []*OptimizationJob {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var result []*OptimizationJob
	for _, j := range e.jobs {
		if j.WorkspaceID == workspaceID {
			result = append(result, j)
		}
	}
	return result
}

// GetReport compiles the final tournament report, diff, and recommendations.
func (e *PromptOptEngine) GetReport(workspaceID, jobID string) (*OptimizationReport, error) {
	job, err := e.GetJob(workspaceID, jobID)
	if err != nil {
		return nil, err
	}

	if job.BestCandidate == nil {
		return nil, errors.New("job does not contain a winning candidate")
	}

	diff := GeneratePromptDiff(job.BasePrompt, job.BestCandidate)

	recommendations := []string{
		fmt.Sprintf("Promote candidate '%s' (Strategy: %s) to production registry with a +%.1f%% expected benchmark gain.", job.BestCandidate.CandidateID, job.BestCandidate.MutationStrategy, job.ScoreImprovement),
		"Maintain few-shot exemplars as live fixtures; refresh when dataset distribution shifts.",
		"Integrate anti-pattern guardrails into system prompt to enforce hard negative boundaries.",
	}

	return &OptimizationReport{
		JobID:            job.ID,
		Title:            job.Title,
		Method:           job.Method,
		BaselineScore:    job.BaselineScore,
		OptimizedScore:   job.BestScore,
		ScoreImprovement: job.ScoreImprovement,
		TotalCandidates:  len(job.Candidates),
		WinningCandidate: job.BestCandidate,
		OptimizationDiff: diff,
		Recommendations:  recommendations,
	}, nil
}

func generateID(prefix string) string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%s_%s", prefix, hex.EncodeToString(b))
}
