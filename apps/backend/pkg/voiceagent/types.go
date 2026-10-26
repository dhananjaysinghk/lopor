package voiceagent

import (
	"time"
)

// AgentState represents the real-time operational state of the conversational voice session.
type AgentState string

const (
	StateIdle        AgentState = "IDLE"
	StateListening   AgentState = "LISTENING"
	StateThinking    AgentState = "THINKING"
	StateSpeaking    AgentState = "SPEAKING"
	StateInterrupted AgentState = "INTERRUPTED"
	StateClosed      AgentState = "CLOSED"
)

// SpeakerType identifies who produced the turn.
type SpeakerType string

const (
	SpeakerUser  SpeakerType = "USER"
	SpeakerAgent SpeakerType = "AGENT"
)

// VADConfig specifies signal processing parameters for voice activity detection.
type VADConfig struct {
	SampleRateHz        int     `json:"sample_rate_hz"`         // Default 16000
	EnergyThresholdRMS  float64 `json:"energy_threshold_rms"`  // Silence vs speech RMS cutoff (e.g., 0.015)
	SilenceTimeoutMs    int     `json:"silence_timeout_ms"`     // Duration of silence before committing turn (e.g., 500ms)
	MinSpeechDurationMs int     `json:"min_speech_duration_ms"` // Duration of speech to trigger barge-in (e.g., 150ms)
}

// VADResult provides frame-level signal classification.
type VADResult struct {
	IsSpeech            bool    `json:"is_speech"`
	RMSEnergy           float64 `json:"rms_energy"`
	SpeechStartDetected bool    `json:"speech_start_detected"`
	SpeechEndDetected   bool    `json:"speech_end_detected"`
}

// VoiceTurn captures an individual spoken utterance in the duplex session.
type VoiceTurn struct {
	TurnID          string      `json:"turn_id"`
	Speaker         SpeakerType `json:"speaker"`
	Transcript      string      `json:"transcript"`
	AudioDurationMs int64       `json:"audio_duration_ms"`
	LatencyMs       int64       `json:"latency_ms"` // Time from user speech end to agent audio output
	WasInterrupted  bool        `json:"was_interrupted"`
	Timestamp       time.Time   `json:"timestamp"`
}

// VoiceSession encapsulates the duplex state, audio buffers, and turn history.
type VoiceSession struct {
	ID                string      `json:"id"`
	WorkspaceID       string      `json:"workspace_id"`
	UserID            string      `json:"user_id"`
	State             AgentState  `json:"state"`
	Config            VADConfig   `json:"config"`
	Turns             []VoiceTurn `json:"turns"`
	CurrentTurn       int         `json:"current_turn"`
	TotalAudioMs      int64       `json:"total_audio_ms"`
	InterruptionCount int         `json:"interruption_count"`
	CreatedAt         time.Time   `json:"created_at"`
	UpdatedAt         time.Time   `json:"updated_at"`
}

// AudioChunkRequest delivers a streaming slice of raw PCM audio bytes.
type AudioChunkRequest struct {
	SessionID   string `json:"session_id"`
	ChunkIndex  int    `json:"chunk_index"`
	PCMBytes    []byte `json:"pcm_bytes"`
	IsFinal     bool   `json:"is_final"`
}

// AudioProcessingResult contains VAD signals, state transitions, and generated audio response.
type AudioProcessingResult struct {
	SessionID           string      `json:"session_id"`
	State               AgentState  `json:"state"`
	VAD                 VADResult   `json:"vad"`
	InterruptedPriorTTS bool        `json:"interrupted_prior_tts"`
	GeneratedTranscript string      `json:"generated_transcript,omitempty"`
	ResponseAudioBytes  []byte      `json:"response_audio_bytes,omitempty"`
	LatencyMs           int64       `json:"latency_ms,omitempty"`
}

// StartSessionRequest defines parameters to spawn a voice agent session.
type StartSessionRequest struct {
	WorkspaceID string     `json:"workspace_id"`
	UserID      string     `json:"user_id"`
	Config      *VADConfig `json:"config,omitempty"`
}

// VoiceSessionTelemetry compiles performance metrics across voice sessions.
type VoiceSessionTelemetry struct {
	WorkspaceID            string  `json:"workspace_id"`
	TotalSessions          int     `json:"total_sessions"`
	TotalTurnsExchanged    int     `json:"total_turns_exchanged"`
	TotalInterruptions     int     `json:"total_interruptions"`
	AverageTTFABMs         float64 `json:"average_ttfab_ms"` // Time to First Audio Byte
	AverageTurnDurationMs float64 `json:"average_turn_duration_ms"`
}
