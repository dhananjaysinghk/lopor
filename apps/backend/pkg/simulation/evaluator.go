package simulation

import (
	"fmt"
	"math"
)

// EvaluateConvergence calculates the overall consensus alignment across recent turns.
func EvaluateConvergence(turns []SimulationTurn, threshold float64) (score float64, converged bool, deadlock bool) {
	if len(turns) == 0 {
		return 0.0, false, false
	}

	windowSize := 4
	if len(turns) < windowSize {
		windowSize = len(turns)
	}

	recentTurns := turns[len(turns)-windowSize:]
	var weightedSum float64
	var weightTotal float64

	consecutiveNegativeSentiment := 0

	for i, turn := range recentTurns {
		// More recent turns carry progressively higher weight
		weight := float64(i + 1)
		weightedSum += turn.AgreementScore * weight
		weightTotal += weight

		if turn.SentimentScore < -0.3 {
			consecutiveNegativeSentiment++
		} else {
			consecutiveNegativeSentiment = 0
		}
	}

	score = math.Round((weightedSum/weightTotal)*1000) / 1000

	// Deadlock check: 3+ turns with persistent hostility or disagreement
	if len(turns) >= 5 && consecutiveNegativeSentiment >= 3 && score < 0.45 {
		deadlock = true
		return score, false, true
	}

	if score >= threshold && len(turns) >= 3 {
		converged = true
	}

	return score, converged, false
}

// SynthesizeFinalVerdict generates executive findings, decisions, and action items.
func SynthesizeFinalVerdict(session *SimulationSession) (verdict string, summary string, actionItems []string) {
	if len(session.Turns) == 0 {
		return "INCOMPLETE", "Simulation was created but no interactions occurred.", nil
	}

	switch session.Status {
	case StatusConverged:
		verdict = "CONSENSUS_REACHED"
		summary = fmt.Sprintf("Autonomous agent personas converged on a mutual resolution for scenario '%s' after %d turns with a convergence alignment score of %.2f.", session.Title, len(session.Turns), session.ConvergenceScore)
		actionItems = []string{
			"Formalize negotiated terms into binding operational playbook",
			"Deploy telemetry health probes and canary validation gates",
			"Distribute final debrief and incident timeline to enterprise stakeholders",
		}

	case StatusDeadlock:
		verdict = "DEADLOCK_UNRESOLVED"
		summary = fmt.Sprintf("Personas reached an intractable impasse during scenario '%s' after %d turns due to irreconcilable objectives and negative sentiment.", session.Title, len(session.Turns))
		actionItems = []string{
			"Escalate decision to executive human leadership committee",
			"Re-evaluate constraint boundaries and relax pricing or SLA floor targets",
			"Schedule structured mediation session with amended ground rules",
		}

	case StatusCompleted:
		verdict = "MAX_TURNS_REACHED"
		summary = fmt.Sprintf("Simulation reached its configured maximum turn threshold (%d turns) with an alignment score of %.2f.", len(session.Turns), session.ConvergenceScore)
		actionItems = []string{
			"Review final position delta between opposing personas",
			"Identify unaddressed objections recorded in the turn ledger",
			"Iterate on scenario parameters and re-run simulation",
		}

	default:
		verdict = "IN_PROGRESS"
		summary = fmt.Sprintf("Simulation active at turn %d of %d.", session.CurrentTurn, session.Config.MaxTurns)
		actionItems = []string{
			"Advance remaining turns to evaluate convergence",
		}
	}

	return verdict, summary, actionItems
}

// ExtractSentimentTrajectory returns an ordered slice of turn-by-turn sentiment values.
func ExtractSentimentTrajectory(turns []SimulationTurn) []float64 {
	trajectory := make([]float64, len(turns))
	for i, turn := range turns {
		trajectory[i] = turn.SentimentScore
	}
	return trajectory
}
