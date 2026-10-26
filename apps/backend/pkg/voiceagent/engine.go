package voiceagent

import (
	"errors"
	"sync"
)

var (
	ErrSessionNotFound = errors.New("voice session not found")
	ErrSessionClosed   = errors.New("voice session is closed")
)

// VoiceAgentEngine coordinates active duplex voice sessions and tenant-isolated audio processing.
type VoiceAgentEngine struct {
	mu       sync.RWMutex
	sessions map[string]*ActiveSession
}

// NewEngine initializes a new VoiceAgentEngine instance.
func NewEngine() *VoiceAgentEngine {
	return &VoiceAgentEngine{
		sessions: make(map[string]*ActiveSession),
	}
}

// StartSession creates and registers an active voice session for a user in a workspace.
func (e *VoiceAgentEngine) StartSession(req StartSessionRequest) (*VoiceSession, error) {
	if req.WorkspaceID == "" || req.UserID == "" {
		return nil, errors.New("workspace_id and user_id are required")
	}

	active := NewActiveSession(req.WorkspaceID, req.UserID, req.Config)
	snapshot := active.GetSnapshot()

	e.mu.Lock()
	e.sessions[snapshot.ID] = active
	e.mu.Unlock()

	return &snapshot, nil
}

// ProcessChunk routes a streaming audio slice to the corresponding session.
func (e *VoiceAgentEngine) ProcessChunk(workspaceID, sessionID string, pcmBytes []byte, chunkDurationMs int) (*AudioProcessingResult, error) {
	e.mu.RLock()
	active, exists := e.sessions[sessionID]
	e.mu.RUnlock()

	if !exists {
		return nil, ErrSessionNotFound
	}

	snapshot := active.GetSnapshot()
	if snapshot.WorkspaceID != workspaceID {
		return nil, ErrSessionNotFound
	}
	if snapshot.State == StateClosed {
		return nil, ErrSessionClosed
	}

	result := active.ProcessAudioChunk(pcmBytes, chunkDurationMs)
	return &result, nil
}

// GetSession retrieves the latest snapshot of a voice session.
func (e *VoiceAgentEngine) GetSession(workspaceID, sessionID string) (*VoiceSession, error) {
	e.mu.RLock()
	active, exists := e.sessions[sessionID]
	e.mu.RUnlock()

	if !exists {
		return nil, ErrSessionNotFound
	}

	snapshot := active.GetSnapshot()
	if snapshot.WorkspaceID != workspaceID {
		return nil, ErrSessionNotFound
	}

	return &snapshot, nil
}

// ListSessions returns all active and concluded voice sessions in the workspace.
func (e *VoiceAgentEngine) ListSessions(workspaceID string) []*VoiceSession {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var result []*VoiceSession
	for _, active := range e.sessions {
		s := active.GetSnapshot()
		if s.WorkspaceID == workspaceID {
			result = append(result, &s)
		}
	}
	return result
}

// EndSession terminates an active voice session.
func (e *VoiceAgentEngine) EndSession(workspaceID, sessionID string) error {
	e.mu.RLock()
	active, exists := e.sessions[sessionID]
	e.mu.RUnlock()

	if !exists {
		return ErrSessionNotFound
	}

	snapshot := active.GetSnapshot()
	if snapshot.WorkspaceID != workspaceID {
		return ErrSessionNotFound
	}

	active.Close()
	return nil
}

// GetTelemetry aggregates operational latency and interruption metrics for the workspace.
func (e *VoiceAgentEngine) GetTelemetry(workspaceID string) *VoiceSessionTelemetry {
	sessions := e.ListSessions(workspaceID)

	var totalTurns int
	var totalInterruptions int
	var totalLatency int64
	var totalDuration int64
	latencyCount := 0

	for _, s := range sessions {
		totalTurns += len(s.Turns)
		totalInterruptions += s.InterruptionCount
		totalDuration += s.TotalAudioMs

		for _, t := range s.Turns {
			if t.Speaker == SpeakerAgent && t.LatencyMs > 0 {
				totalLatency += t.LatencyMs
				latencyCount++
			}
		}
	}

	var avgTTFAB float64
	if latencyCount > 0 {
		avgTTFAB = float64(totalLatency) / float64(latencyCount)
	}

	var avgTurnDuration float64
	if totalTurns > 0 {
		avgTurnDuration = float64(totalDuration) / float64(totalTurns)
	}

	return &VoiceSessionTelemetry{
		WorkspaceID:            workspaceID,
		TotalSessions:          len(sessions),
		TotalTurnsExchanged:    totalTurns,
		TotalInterruptions:     totalInterruptions,
		AverageTTFABMs:         avgTTFAB,
		AverageTurnDurationMs: avgTurnDuration,
	}
}
