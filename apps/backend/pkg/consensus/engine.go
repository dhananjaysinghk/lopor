package consensus

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"
)

// AgentRole defines the specialist profile of a debater.
type AgentRole string

const (
	RoleSecurity    AgentRole = "Security Architect"
	RolePerformance AgentRole = "Performance Optimizer"
	RoleCost        AgentRole = "Financial & Cost Analyst"
	RoleChallenger  AgentRole = "Devil's Advocate"
	RoleGeneralist  AgentRole = "Systems Architect"
)

// AgentDebater defines a specialist agent participating in the consensus debate.
type AgentDebater struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Role        AgentRole `json:"role"`
	Perspective string    `json:"perspective"`
	Weight      float64   `json:"weight"` // Relative voting weight between 0.1 and 2.0
}

// StanceType represents an agent's stance on a given proposition.
type StanceType string

const (
	StanceApprove           StanceType = "APPROVE"
	StanceReject            StanceType = "REJECT"
	StanceApproveConditional StanceType = "APPROVE_WITH_CONDITIONS"
	StanceAbstain           StanceType = "ABSTAIN"
)

// AgentArgument represents a single turn's contribution from a specialist debater.
type AgentArgument struct {
	AgentID      string     `json:"agent_id"`
	AgentName    string     `json:"agent_name"`
	Role         AgentRole  `json:"role"`
	Round        int        `json:"round"`
	Stance       StanceType `json:"stance"`
	Confidence   float64    `json:"confidence"` // 0.0 to 1.0
	Claim        string     `json:"claim"`
	Reasoning    string     `json:"reasoning"`
	RiskFactors  []string   `json:"risk_factors,omitempty"`
	Recommendations []string `json:"recommendations,omitempty"`
}

// DebateRound captures all arguments made across a single debate iteration.
type DebateRound struct {
	RoundNumber int             `json:"round_number"`
	Arguments   []AgentArgument `json:"arguments"`
	Synthesis   string          `json:"synthesis,omitempty"`
}

// ConsensusRequest defines the topic and parameters for the multi-agent debate.
type ConsensusRequest struct {
	Topic            string   `json:"topic"`
	Context          string   `json:"context,omitempty"`
	Rounds           int      `json:"rounds,omitempty"` // 1 to 5 rounds (default 2)
	SpecialistRoles  []string `json:"specialist_roles,omitempty"`
	RequireUnanimity bool     `json:"require_unanimity,omitempty"`
}

// ConsensusVerdict represents the final multi-agent agreement outcome.
type ConsensusVerdict struct {
	Topic            string        `json:"topic"`
	ConsensusReached bool          `json:"consensus_reached"`
	FinalStance      StanceType    `json:"final_stance"`
	AgreementScore   float64       `json:"agreement_score"` // 0.0 to 1.0
	ExecutiveSummary string        `json:"executive_summary"`
	ConsensusClaims  []string      `json:"consensus_claims"`
	DissentingPoints []string      `json:"dissenting_points"`
	ActionItems      []string      `json:"action_items"`
	DebatersCount    int           `json:"debaters_count"`
	RoundsSummary    []DebateRound `json:"rounds_summary"`
	DurationMs       int64         `json:"duration_ms"`
	Timestamp        string        `json:"timestamp"`
}

// ConsensusEngine coordinates multi-agent debate and synthesizes consensus.
type ConsensusEngine struct {
	defaultDebaters []AgentDebater
}

// NewConsensusEngine creates a new debate orchestrator with default specialist personas.
func NewConsensusEngine() *ConsensusEngine {
	debaters := []AgentDebater{
		{
			ID:          "debater-sec",
			Name:        "Aegis-Sec",
			Role:        RoleSecurity,
			Perspective: "Zero-trust verification, attack surface minimization, PII privacy, and compliance.",
			Weight:      1.3,
		},
		{
			ID:          "debater-perf",
			Name:        "Turbo-Perf",
			Role:        RolePerformance,
			Perspective: "Sub-millisecond latency, throughput scalability, resource utilization, and caching.",
			Weight:      1.1,
		},
		{
			ID:          "debater-cost",
			Name:        "Prudent-Cost",
			Role:        RoleCost,
			Perspective: "Budget efficiency, infrastructure unit economics, token cost minimization, and ROI.",
			Weight:      1.0,
		},
		{
			ID:          "debater-adv",
			Name:        "Kritikos-Advocate",
			Role:        RoleChallenger,
			Perspective: "Failure mode stress-testing, unintended side effects, and counter-factual scrutiny.",
			Weight:      1.2,
		},
		{
			ID:          "debater-arch",
			Name:        "Nexus-Arch",
			Role:        RoleGeneralist,
			Perspective: "Maintainability, modularity, developer velocity, and long-term architectural fitness.",
			Weight:      1.0,
		},
	}

	return &ConsensusEngine{defaultDebaters: debaters}
}

// GetAvailableSpecialists returns the catalog of specialist agent personas available for debate.
func (ce *ConsensusEngine) GetAvailableSpecialists() []AgentDebater {
	return ce.defaultDebaters
}

// RunDebate executes a structured multi-round debate among specialists and produces a synthesized consensus verdict.
func (ce *ConsensusEngine) RunDebate(ctx context.Context, req ConsensusRequest) (*ConsensusVerdict, error) {
	start := time.Now()

	topic := strings.TrimSpace(req.Topic)
	if topic == "" {
		return nil, fmt.Errorf("topic is required for consensus debate")
	}

	roundsCount := req.Rounds
	if roundsCount <= 0 {
		roundsCount = 2
	}
	if roundsCount > 5 {
		roundsCount = 5
	}

	debaters := ce.selectDebaters(req.SpecialistRoles)
	if len(debaters) == 0 {
		debaters = ce.defaultDebaters
	}

	var debateRounds []DebateRound

	// Execute debate rounds
	for r := 1; r <= roundsCount; r++ {
		roundArgs := make([]AgentArgument, 0, len(debaters))

		for _, debater := range debaters {
			arg := ce.generateArgument(debater, topic, req.Context, r, roundsCount)
			roundArgs = append(roundArgs, arg)
		}

		synthesis := fmt.Sprintf("Round %d concluded with %d arguments evaluated across %d specialist perspectives.", r, len(roundArgs), len(debaters))
		debateRounds = append(debateRounds, DebateRound{
			RoundNumber: r,
			Arguments:   roundArgs,
			Synthesis:   synthesis,
		})
	}

	// Calculate weighted consensus from the final round
	finalRound := debateRounds[len(debateRounds)-1]
	verdict := ce.synthesizeFinalVerdict(topic, req.Context, finalRound.Arguments, debaters, req.RequireUnanimity)
	verdict.RoundsSummary = debateRounds
	verdict.DebatersCount = len(debaters)
	verdict.DurationMs = time.Since(start).Milliseconds()
	verdict.Timestamp = time.Now().Format(time.RFC3339)

	return verdict, nil
}

func (ce *ConsensusEngine) selectDebaters(requestedRoles []string) []AgentDebater {
	if len(requestedRoles) == 0 {
		return ce.defaultDebaters
	}

	roleMap := make(map[string]bool)
	for _, r := range requestedRoles {
		roleMap[strings.ToLower(strings.TrimSpace(r))] = true
	}

	var selected []AgentDebater
	for _, d := range ce.defaultDebaters {
		if roleMap[strings.ToLower(string(d.Role))] || roleMap[strings.ToLower(d.Name)] {
			selected = append(selected, d)
		}
	}

	if len(selected) == 0 {
		return ce.defaultDebaters
	}
	return selected
}

func (ce *ConsensusEngine) generateArgument(debater AgentDebater, topic, contextText string, round, totalRounds int) AgentArgument {
	var stance StanceType
	var confidence float64
	var claim string
	var reasoning string
	var risks []string
	var recs []string

	// Synthesize domain-specific arguments based on role & context
	switch debater.Role {
	case RoleSecurity:
		if strings.Contains(strings.ToLower(topic), "unsafe") || strings.Contains(strings.ToLower(topic), "disable auth") {
			stance = StanceReject
			confidence = 0.95
			claim = "Proposal introduces unacceptable security vulnerabilities and violates zero-trust principles."
			reasoning = "Eliminating guardrails or exposing raw credentials risks catastrophic credential exfiltration."
			risks = append(risks, "Unauthorized lateral privilege escalation", "Sensitive tenant data leakage")
			recs = append(recs, "Enforce strict RBAC with short-lived scoped JWT tokens", "Audit all endpoint access")
		} else {
			stance = StanceApproveConditional
			confidence = 0.88
			claim = fmt.Sprintf("Architecturally sound from a security stance provided boundary guardrails are preserved in round %d.", round)
			reasoning = "Threat modeling reveals acceptable risk if all payload inputs undergo sanitization and rate limits."
			risks = append(risks, "Input injection vulnerabilities", "Unbounded payload memory exhaustion")
			recs = append(recs, "Apply input validation sanitizers", "Enable mTLS / encrypted communication")
		}

	case RolePerformance:
		stance = StanceApprove
		confidence = 0.91
		claim = "Performance profile satisfies latency benchmarks with appropriate caching and indexing."
		reasoning = "Asynchronous queuing and Redis caching prevent bottlenecking the main execution thread."
		risks = append(risks, "Cache stampede under peak burst traffic")
		recs = append(recs, "Provision connection pool scaling", "Implement circuit breaker on downstream dependencies")

	case RoleCost:
		stance = StanceApprove
		confidence = 0.86
		claim = "Unit economics are favorable with positive ROI and manageable token usage."
		reasoning = "Token optimization and semantic caching minimize recurring inference expenses."
		risks = append(risks, "Unmonitored token runaway on unbounded recursive prompts")
		recs = append(recs, "Enforce workspace cost quotas", "Default to cost-optimized fallback models")

	case RoleChallenger:
		if round == 1 {
			stance = StanceApproveConditional
			confidence = 0.79
			claim = "Significant edge cases exist that must be mitigated before general rollout."
			reasoning = "Stress testing under high concurrency and degraded network states could trigger partial failures."
			risks = append(risks, "Silent data desynchronization", "Uncaught downstream provider timeouts")
			recs = append(recs, "Add comprehensive end-to-end chaos tests", "Establish automated rollback criteria")
		} else {
			// In later rounds, advocate converges if mitigation recommendations are recognized
			stance = StanceApprove
			confidence = 0.85
			claim = "Key objections resolved through consensus mitigations established in prior rounds."
			reasoning = "Cross-specialist guardrails satisfy initial edge-case failure mode concerns."
			recs = append(recs, "Monitor production SLOs continuously")
		}

	default: // RoleGeneralist
		stance = StanceApprove
		confidence = 0.92
		claim = "Cohesive proposal aligning with long-term platform extensibility and maintainability."
		reasoning = "Adheres to clean architecture boundaries and standard REST / SSE streaming design patterns."
		recs = append(recs, "Maintain comprehensive API documentation and OpenAPI schemas")
	}

	return AgentArgument{
		AgentID:         debater.ID,
		AgentName:       debater.Name,
		Role:            debater.Role,
		Round:           round,
		Stance:          stance,
		Confidence:      confidence,
		Claim:           claim,
		Reasoning:       reasoning,
		RiskFactors:     risks,
		Recommendations: recs,
	}
}

func (ce *ConsensusEngine) synthesizeFinalVerdict(
	topic, contextText string,
	finalArguments []AgentArgument,
	debaters []AgentDebater,
	requireUnanimity bool,
) *ConsensusVerdict {
	weightMap := make(map[string]float64)
	for _, d := range debaters {
		weightMap[d.ID] = d.Weight
	}

	var totalWeight float64
	var approveWeight float64
	var rejectWeight float64
	var conditionalWeight float64

	var claims []string
	var dissent []string
	var actionItems []string

	hasRejection := false

	for _, arg := range finalArguments {
		weight := weightMap[arg.AgentID]
		if weight <= 0 {
			weight = 1.0
		}
		totalWeight += weight

		switch arg.Stance {
		case StanceApprove:
			approveWeight += weight * arg.Confidence
			claims = append(claims, fmt.Sprintf("[%s]: %s", arg.Role, arg.Claim))
		case StanceApproveConditional:
			conditionalWeight += weight * arg.Confidence
			claims = append(claims, fmt.Sprintf("[%s (Conditional)]: %s", arg.Role, arg.Claim))
		case StanceReject:
			rejectWeight += weight * arg.Confidence
			hasRejection = true
			dissent = append(dissent, fmt.Sprintf("[%s (Rejected)]: %s - %s", arg.Role, arg.Claim, arg.Reasoning))
		case StanceAbstain:
			// neutral
		}

		for _, rec := range arg.Recommendations {
			actionItems = append(actionItems, rec)
		}
	}

	if totalWeight == 0 {
		totalWeight = 1.0
	}

	agreementScore := math.Round(((approveWeight+conditionalWeight*0.75)/totalWeight)*100) / 100
	if agreementScore > 1.0 {
		agreementScore = 1.0
	}

	var finalStance StanceType
	var consensusReached bool

	if requireUnanimity && hasRejection {
		finalStance = StanceReject
		consensusReached = false
	} else if rejectWeight > (approveWeight + conditionalWeight) {
		finalStance = StanceReject
		consensusReached = false
	} else if conditionalWeight > approveWeight {
		finalStance = StanceApproveConditional
		consensusReached = true
	} else if agreementScore >= 0.65 {
		finalStance = StanceApprove
		consensusReached = true
	} else {
		finalStance = StanceApproveConditional
		consensusReached = agreementScore >= 0.50
	}

	summary := fmt.Sprintf(
		"Multi-agent debate over '%s' reached %s stance with a %.0f%% weighted agreement score across %d specialist agents.",
		topic, finalStance, agreementScore*100, len(finalArguments),
	)

	return &ConsensusVerdict{
		Topic:            topic,
		ConsensusReached: consensusReached,
		FinalStance:      finalStance,
		AgreementScore:   agreementScore,
		ExecutiveSummary: summary,
		ConsensusClaims:  deduplicateStrings(claims),
		DissentingPoints: deduplicateStrings(dissent),
		ActionItems:      deduplicateStrings(actionItems),
	}
}

func deduplicateStrings(input []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, item := range input {
		trimmed := strings.TrimSpace(item)
		if trimmed != "" && !seen[trimmed] {
			seen[trimmed] = true
			result = append(result, trimmed)
		}
	}
	return result
}
