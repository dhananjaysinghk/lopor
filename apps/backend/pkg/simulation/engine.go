package simulation

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrSessionNotFound = errors.New("simulation session not found")
	ErrSessionClosed   = errors.New("simulation session is already closed or completed")
	ErrInvalidRequest  = errors.New("invalid simulation request parameters")
)

// SimulationEngine manages the lifecycle, execution, and analytics of multi-agent scenarios.
type SimulationEngine struct {
	mu       sync.RWMutex
	sessions map[string]*SimulationSession
}

// NewEngine creates an initialized instance of the SimulationEngine.
func NewEngine() *SimulationEngine {
	return &SimulationEngine{
		sessions: make(map[string]*SimulationSession),
	}
}

// CreateSession initializes a new multi-agent simulation arena with personas and scenario parameters.
func (e *SimulationEngine) CreateSession(req StartSimulationRequest) (*SimulationSession, error) {
	if req.WorkspaceID == "" || req.Title == "" || req.Scenario == "" {
		return nil, fmt.Errorf("%w: workspace_id, title, and scenario are required", ErrInvalidRequest)
	}

	personas := req.CustomPersonas
	if len(personas) == 0 {
		preset := req.Preset
		if preset == "" {
			preset = PresetIncidentDrill
		}
		personas = GetPresetPersonas(preset)
	}

	config := req.Config
	if config == nil {
		config = &SimulationConfig{
			MaxTurns:                 12,
			AutoConvergenceDetection: true,
			ConvergenceThreshold:     0.82,
			EnvironmentContext:       "Production cloud environment under active enterprise governance.",
		}
	} else {
		if config.MaxTurns <= 0 {
			config.MaxTurns = 12
		}
		if config.ConvergenceThreshold <= 0.0 {
			config.ConvergenceThreshold = 0.82
		}
	}

	sessionID := generateID("sim")
	now := time.Now().UTC()

	session := &SimulationSession{
		ID:               sessionID,
		WorkspaceID:      req.WorkspaceID,
		Title:            req.Title,
		Scenario:         req.Scenario,
		Status:           StatusRunning,
		Personas:         personas,
		Turns:            make([]SimulationTurn, 0),
		CurrentTurn:      0,
		Config:           *config,
		ConvergenceScore: 0.0,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	e.mu.Lock()
	e.sessions[sessionID] = session
	e.mu.Unlock()

	return session, nil
}

// GetSession retrieves an active or past simulation session by ID with workspace tenancy check.
func (e *SimulationEngine) GetSession(workspaceID, sessionID string) (*SimulationSession, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	session, exists := e.sessions[sessionID]
	if !exists || session.WorkspaceID != workspaceID {
		return nil, ErrSessionNotFound
	}

	return session, nil
}

// ListSessions returns all sessions registered under the given workspace.
func (e *SimulationEngine) ListSessions(workspaceID string) []*SimulationSession {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var result []*SimulationSession
	for _, s := range e.sessions {
		if s.WorkspaceID == workspaceID {
			result = append(result, s)
		}
	}
	return result
}

// AdvanceStep progresses the simulation by one or more turns, optionally accepting a human injection.
func (e *SimulationEngine) AdvanceStep(workspaceID, sessionID string, req AdvanceStepRequest) (*SimulationStepResult, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	session, exists := e.sessions[sessionID]
	if !exists || session.WorkspaceID != workspaceID {
		return nil, ErrSessionNotFound
	}

	if session.Status != StatusRunning && session.Status != StatusDraft {
		return nil, ErrSessionClosed
	}

	var addedTurns []SimulationTurn

	// 1. Process human-in-the-loop intervention if provided
	if req.HumanInput != "" {
		humanName := req.HumanName
		if humanName == "" {
			humanName = "Human Moderator"
		}
		session.CurrentTurn++
		humanTurn := SimulationTurn{
			TurnNumber:     session.CurrentTurn,
			SpeakerID:      "human_participant",
			SpeakerName:    humanName,
			Role:           RoleStakeholder,
			Content:        req.HumanInput,
			Rationale:      "Human participant directive guiding the simulation.",
			SentimentScore: 0.2,
			AgreementScore: 0.7,
			Timestamp:      time.Now().UTC(),
		}
		session.Turns = append(session.Turns, humanTurn)
		addedTurns = append(addedTurns, humanTurn)
	}

	stepsToRun := req.TargetSteps
	if stepsToRun <= 0 {
		stepsToRun = 1
	}

	// 2. Advance autonomous persona turns
	for i := 0; i < stepsToRun; i++ {
		if session.CurrentTurn >= session.Config.MaxTurns {
			session.Status = StatusCompleted
			break
		}

		personaIndex := session.CurrentTurn % len(session.Personas)
		activePersona := session.Personas[personaIndex]

		session.CurrentTurn++
		turn := GenerateTurnUtterance(
			activePersona,
			session.Scenario,
			session.Config.EnvironmentContext,
			session.Turns,
			session.CurrentTurn-1,
		)

		session.Turns = append(session.Turns, turn)
		addedTurns = append(addedTurns, turn)

		// Convergence and deadlock assessment
		if session.Config.AutoConvergenceDetection {
			score, converged, deadlock := EvaluateConvergence(session.Turns, session.Config.ConvergenceThreshold)
			session.ConvergenceScore = score

			if converged {
				session.Status = StatusConverged
				break
			}
			if deadlock {
				session.Status = StatusDeadlock
				break
			}
		}

		if session.CurrentTurn >= session.Config.MaxTurns {
			session.Status = StatusCompleted
			break
		}
	}

	// 3. Finalize verdict if terminal state reached
	if session.Status == StatusConverged || session.Status == StatusDeadlock || session.Status == StatusCompleted {
		verdict, summary, actions := SynthesizeFinalVerdict(session)
		session.FinalVerdict = verdict
		session.Summary = summary
		session.ActionItems = actions
	}

	session.UpdatedAt = time.Now().UTC()

	return &SimulationStepResult{
		SessionID:        session.ID,
		Status:           session.Status,
		CurrentTurn:      session.CurrentTurn,
		AddedTurns:       addedTurns,
		ConvergenceScore: session.ConvergenceScore,
		HasConverged:     session.Status == StatusConverged,
		HasDeadlock:      session.Status == StatusDeadlock,
	}, nil
}

// RunToCompletion advances an active simulation until consensus, deadlock, or max turns are reached.
func (e *SimulationEngine) RunToCompletion(workspaceID, sessionID string) (*SimulationReport, error) {
	session, err := e.GetSession(workspaceID, sessionID)
	if err != nil {
		return nil, err
	}

	for session.Status == StatusRunning || session.Status == StatusDraft {
		_, err := e.AdvanceStep(workspaceID, sessionID, AdvanceStepRequest{TargetSteps: 2})
		if err != nil {
			if errors.Is(err, ErrSessionClosed) {
				break
			}
			return nil, err
		}

		// Refresh state check
		session, err = e.GetSession(workspaceID, sessionID)
		if err != nil {
			return nil, err
		}
	}

	return e.GetReport(workspaceID, sessionID)
}

// GetReport compiles the final analytical debrief and sentiment trajectory.
func (e *SimulationEngine) GetReport(workspaceID, sessionID string) (*SimulationReport, error) {
	session, err := e.GetSession(workspaceID, sessionID)
	if err != nil {
		return nil, err
	}

	verdict := session.FinalVerdict
	summary := session.Summary
	actions := session.ActionItems

	if verdict == "" {
		verdict, summary, actions = SynthesizeFinalVerdict(session)
	}

	return &SimulationReport{
		SessionID:             session.ID,
		Title:                 session.Title,
		Scenario:              session.Scenario,
		Status:                session.Status,
		TotalTurns:            len(session.Turns),
		ParticipatingPersonas: len(session.Personas),
		ConvergenceScore:      session.ConvergenceScore,
		FinalVerdict:          verdict,
		Summary:               summary,
		ActionItems:           actions,
		SentimentTrajectory:   ExtractSentimentTrajectory(session.Turns),
		Turns:                 session.Turns,
	}, nil
}

func generateID(prefix string) string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%s_%s", prefix, hex.EncodeToString(b))
}
