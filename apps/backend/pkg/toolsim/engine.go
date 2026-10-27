package toolsim

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrToolNotFound    = errors.New("mock tool not found in registry")
	ErrSessionNotFound = errors.New("sandbox session not found")
)

// ToolSimEngine coordinates mock tool definitions, dry-run sandboxes, and blast radius auditing.
type ToolSimEngine struct {
	mu       sync.RWMutex
	registry *ToolRegistry
	sessions map[string]*SandboxSession
}

// NewEngine initializes a new ToolSimEngine instance.
func NewEngine() *ToolSimEngine {
	return &ToolSimEngine{
		registry: NewToolRegistry(),
		sessions: make(map[string]*SandboxSession),
	}
}

// RegisterTool registers a custom mock tool definition.
func (e *ToolSimEngine) RegisterTool(tool MockTool) {
	e.registry.Register(tool)
}

// ListTools returns all mock tools available for a workspace.
func (e *ToolSimEngine) ListTools(workspaceID string) []MockTool {
	return e.registry.List(workspaceID)
}

// CreateSandbox spawns a new dry-run session with virtual state tracking.
func (e *ToolSimEngine) CreateSandbox(workspaceID, title string) (*SandboxSession, error) {
	if workspaceID == "" || title == "" {
		return nil, errors.New("workspace_id and title are required")
	}

	sessionID := generateID("sbx")
	now := time.Now().UTC()

	session := &SandboxSession{
		ID:           sessionID,
		WorkspaceID:  workspaceID,
		Title:        title,
		Invocations:  make([]ToolInvocation, 0),
		VirtualState: make(map[string]interface{}),
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	e.mu.Lock()
	e.sessions[sessionID] = session
	e.mu.Unlock()

	return session, nil
}

// ExecuteToolDryRun executes a simulated tool call within an existing sandbox session.
func (e *ToolSimEngine) ExecuteToolDryRun(workspaceID, sessionID, toolName string, params map[string]interface{}) (*ToolInvocation, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	session, exists := e.sessions[sessionID]
	if !exists || session.WorkspaceID != workspaceID {
		return nil, ErrSessionNotFound
	}

	tool, found := e.registry.Get(workspaceID, toolName)
	if !found {
		return nil, fmt.Errorf("%w: %s", ErrToolNotFound, toolName)
	}

	sideEffects := AnalyzeSideEffects(*tool, params)
	invocationID := generateID("inv")
	now := time.Now().UTC()

	// Update virtual state based on tool name
	stateKey := fmt.Sprintf("%s:%s", tool.Category, tool.Name)
	session.VirtualState[stateKey] = params

	inv := ToolInvocation{
		InvocationID:        invocationID,
		ToolName:            tool.Name,
		Category:            tool.Category,
		RiskLevel:           tool.RiskLevel,
		InputParameters:     params,
		SimulatedOutput:     tool.DefaultMockResponse,
		SideEffects:         sideEffects,
		LatencyMs:           tool.SimulatedLatencyMs,
		Status:              "SUCCESS",
		RequiresHumanReview: tool.RiskLevel == RiskHigh || tool.RiskLevel == RiskCritical,
		Timestamp:           now,
	}

	session.Invocations = append(session.Invocations, inv)

	// Recompute session metrics
	score, cost, hasCrit, _ := CalculateBlastRadius(session.Invocations)
	session.BlastRadiusScore = score
	session.TotalEstimatedCostUSD = cost
	session.HasCriticalSideEffects = hasCrit
	session.UpdatedAt = now

	return &inv, nil
}

// ExecutePlanDryRun runs a multi-step sequence of tool actions in a temporary sandbox.
func (e *ToolSimEngine) ExecutePlanDryRun(req ExecutePlanRequest) (*SandboxReport, error) {
	if req.WorkspaceID == "" || req.Title == "" {
		return nil, errors.New("workspace_id and title are required")
	}

	session, err := e.CreateSandbox(req.WorkspaceID, req.Title)
	if err != nil {
		return nil, err
	}

	for _, step := range req.Steps {
		_, err := e.ExecuteToolDryRun(req.WorkspaceID, session.ID, step.ToolName, step.Parameters)
		if err != nil {
			return nil, err
		}
	}

	e.mu.RLock()
	latestSession := e.sessions[session.ID]
	e.mu.RUnlock()

	report := GenerateSandboxReport(latestSession)
	return &report, nil
}

// GetReport compiles the final risk report for a sandbox session.
func (e *ToolSimEngine) GetReport(workspaceID, sessionID string) (*SandboxReport, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	session, exists := e.sessions[sessionID]
	if !exists || session.WorkspaceID != workspaceID {
		return nil, ErrSessionNotFound
	}

	report := GenerateSandboxReport(session)
	return &report, nil
}

// GetSession returns the active session snapshot.
func (e *ToolSimEngine) GetSession(workspaceID, sessionID string) (*SandboxSession, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	session, exists := e.sessions[sessionID]
	if !exists || session.WorkspaceID != workspaceID {
		return nil, ErrSessionNotFound
	}

	return session, nil
}

func generateID(prefix string) string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%s_%s", prefix, hex.EncodeToString(b))
}
