package simulation

import (
	"time"
)

// PersonaRole defines the structural function of an agent within a simulation scenario.
type PersonaRole string

const (
	RoleCoordinator PersonaRole = "COORDINATOR"
	RoleExpert      PersonaRole = "EXPERT"
	RoleAdversary   PersonaRole = "ADVERSARY"
	RoleCritic      PersonaRole = "CRITIC"
	RoleObserver    PersonaRole = "OBSERVER"
	RoleStakeholder PersonaRole = "STAKEHOLDER"
)

// Temperament defines the behavioral orientation and decision posture of an agent persona.
type Temperament string

const (
	TemperamentAnalytic     Temperament = "ANALYTIC"
	TemperamentConciliatory Temperament = "CONCILIATORY"
	TemperamentSkeptical    Temperament = "SKEPTICAL"
	TemperamentAggressive   Temperament = "AGGRESSIVE"
	TemperamentPragmatic    Temperament = "PRAGMATIC"
)

// SimulationStatus represents the lifecycle state of a simulation run.
type SimulationStatus string

const (
	StatusDraft     SimulationStatus = "DRAFT"
	StatusRunning   SimulationStatus = "RUNNING"
	StatusPaused    SimulationStatus = "PAUSED"
	StatusConverged SimulationStatus = "CONVERGED"
	StatusDeadlock  SimulationStatus = "DEADLOCK"
	StatusCompleted SimulationStatus = "COMPLETED"
)

// AgentPersona models an autonomous participant with custom instructions and temperament.
type AgentPersona struct {
	ID                 string      `json:"id"`
	Name               string      `json:"name"`
	Role               PersonaRole `json:"role"`
	Temperament        Temperament `json:"temperament"`
	SystemInstructions string      `json:"system_instructions"`
	PrimaryObjective   string      `json:"primary_objective"`
	KnowledgeBoundary  string      `json:"knowledge_boundary,omitempty"`
}

// SimulationTurn captures an individual utterance, decision, or action in the simulation.
type SimulationTurn struct {
	TurnNumber      int         `json:"turn_number"`
	SpeakerID       string      `json:"speaker_id"`
	SpeakerName     string      `json:"speaker_name"`
	Role            PersonaRole `json:"role"`
	Content         string      `json:"content"`
	Rationale       string      `json:"rationale,omitempty"`
	SentimentScore  float64     `json:"sentiment_score"` // -1.0 (hostile/opposing) to +1.0 (collaborative/aligned)
	AgreementScore  float64     `json:"agreement_score"` // 0.0 to 1.0 alignment with consensus
	Timestamp       time.Time   `json:"timestamp"`
}

// SimulationConfig specifies operational limits and parameters for scenario execution.
type SimulationConfig struct {
	MaxTurns                 int     `json:"max_turns"`
	AutoConvergenceDetection bool    `json:"auto_convergence_detection"`
	ConvergenceThreshold     float64 `json:"convergence_threshold"` // e.g. 0.85
	EnvironmentContext       string  `json:"environment_context"`
}

// SimulationSession encapsulates an active or concluded multi-agent simulation scenario.
type SimulationSession struct {
	ID               string             `json:"id"`
	WorkspaceID      string             `json:"workspace_id"`
	Title            string             `json:"title"`
	Scenario         string             `json:"scenario"`
	Status           SimulationStatus   `json:"status"`
	Personas         []AgentPersona     `json:"personas"`
	Turns            []SimulationTurn   `json:"turns"`
	CurrentTurn      int                `json:"current_turn"`
	Config           SimulationConfig   `json:"config"`
	ConvergenceScore float64            `json:"convergence_score"`
	FinalVerdict     string             `json:"final_verdict,omitempty"`
	Summary          string             `json:"summary,omitempty"`
	ActionItems      []string           `json:"action_items,omitempty"`
	CreatedAt        time.Time          `json:"created_at"`
	UpdatedAt        time.Time          `json:"updated_at"`
}

// StartSimulationRequest supplies initial parameters to spawn a simulation scenario.
type StartSimulationRequest struct {
	WorkspaceID    string            `json:"workspace_id"`
	Title          string            `json:"title"`
	Scenario       string            `json:"scenario"`
	Preset         string            `json:"preset,omitempty"` // e.g. "INCIDENT_DRILL", "NEGOTIATION_WAR_ROOM", "EXECUTIVE_REVIEW"
	CustomPersonas []AgentPersona    `json:"custom_personas,omitempty"`
	Config         *SimulationConfig `json:"config,omitempty"`
}

// AdvanceStepRequest allows manual step execution or human user injection.
type AdvanceStepRequest struct {
	HumanInput  string `json:"human_input,omitempty"`
	HumanName   string `json:"human_name,omitempty"`
	TargetSteps int    `json:"target_steps,omitempty"` // Number of turns to advance (default 1)
}

// SimulationStepResult provides the outcome of advancing simulation turns.
type SimulationStepResult struct {
	SessionID        string           `json:"session_id"`
	Status           SimulationStatus `json:"status"`
	CurrentTurn      int              `json:"current_turn"`
	AddedTurns       []SimulationTurn `json:"added_turns"`
	ConvergenceScore float64          `json:"convergence_score"`
	HasConverged     bool             `json:"has_converged"`
	HasDeadlock      bool             `json:"has_deadlock"`
}

// SimulationReport summarizes the final outcomes, consensus, and metrics of a completed session.
type SimulationReport struct {
	SessionID             string           `json:"session_id"`
	Title                 string           `json:"title"`
	Scenario              string           `json:"scenario"`
	Status                SimulationStatus `json:"status"`
	TotalTurns            int              `json:"total_turns"`
	ParticipatingPersonas int              `json:"participating_personas"`
	ConvergenceScore      float64          `json:"convergence_score"`
	FinalVerdict          string           `json:"final_verdict"`
	Summary               string           `json:"summary"`
	ActionItems           []string         `json:"action_items"`
	SentimentTrajectory   []float64        `json:"sentiment_trajectory"`
	Turns                 []SimulationTurn `json:"turns"`
}
