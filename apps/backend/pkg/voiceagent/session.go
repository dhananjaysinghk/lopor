package voiceagent

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"sync"
	"time"
)

// ActiveSession wraps a VoiceSession with real-time VAD processing and buffer accumulation.
type ActiveSession struct {
	mu           sync.Mutex
	session      VoiceSession
	vad          *VADProcessor
	audioBuffer  []byte
	speechStart  time.Time
}

// NewActiveSession instantiates an active duplex session with configured signal processing.
func NewActiveSession(workspaceID, userID string, config *VADConfig) *ActiveSession {
	cfg := VADConfig{
		SampleRateHz:        16000,
		EnergyThresholdRMS:  0.015,
		SilenceTimeoutMs:    500,
		MinSpeechDurationMs: 150,
	}
	if config != nil {
		if config.SampleRateHz > 0 {
			cfg.SampleRateHz = config.SampleRateHz
		}
		if config.EnergyThresholdRMS > 0.0 {
			cfg.EnergyThresholdRMS = config.EnergyThresholdRMS
		}
		if config.SilenceTimeoutMs > 0 {
			cfg.SilenceTimeoutMs = config.SilenceTimeoutMs
		}
		if config.MinSpeechDurationMs > 0 {
			cfg.MinSpeechDurationMs = config.MinSpeechDurationMs
		}
	}

	sessionID := generateID("vcs")
	now := time.Now().UTC()

	return &ActiveSession{
		session: VoiceSession{
			ID:          sessionID,
			WorkspaceID: workspaceID,
			UserID:      userID,
			State:       StateIdle,
			Config:      cfg,
			Turns:       make([]VoiceTurn, 0),
			CurrentTurn: 0,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		vad:         NewVADProcessor(cfg),
		audioBuffer: make([]byte, 0),
	}
}

// ProcessAudioChunk consumes a streaming PCM chunk, evaluates VAD, handles interruptions, and triggers responses.
func (s *ActiveSession) ProcessAudioChunk(pcmBytes []byte, chunkDurationMs int) AudioProcessingResult {
	s.mu.Lock()
	defer s.mu.Unlock()

	vadResult := s.vad.ProcessFrame(pcmBytes, chunkDurationMs)
	var interruptedPriorTTS bool
	now := time.Now().UTC()

	// 1. Interruption / Barge-in Detection:
	// If the agent is speaking or thinking and the user starts talking, abort current playback
	if (s.session.State == StateSpeaking || s.session.State == StateThinking) && vadResult.SpeechStartDetected {
		s.session.State = StateInterrupted
		s.session.InterruptionCount++
		interruptedPriorTTS = true

		// Mark last agent turn as interrupted
		if len(s.session.Turns) > 0 {
			lastIdx := len(s.session.Turns) - 1
			if s.session.Turns[lastIdx].Speaker == SpeakerAgent {
				s.session.Turns[lastIdx].WasInterrupted = true
			}
		}

		s.session.State = StateListening
		s.audioBuffer = append(s.audioBuffer[:0], pcmBytes...)
		s.speechStart = now
		s.session.UpdatedAt = now

		return AudioProcessingResult{
			SessionID:           s.session.ID,
			State:               StateListening,
			VAD:                 vadResult,
			InterruptedPriorTTS: true,
		}
	}

	// 2. Active User Speech Buffering
	if vadResult.IsSpeech {
		if s.session.State != StateListening {
			s.session.State = StateListening
			s.speechStart = now
		}
		s.audioBuffer = append(s.audioBuffer, pcmBytes...)
		s.session.UpdatedAt = now

		return AudioProcessingResult{
			SessionID: s.session.ID,
			State:     StateListening,
			VAD:       vadResult,
		}
	}

	// 3. User Speech End Detected -> Trigger Turn Completion & Agent Response
	if vadResult.SpeechEndDetected && len(s.audioBuffer) > 0 {
		t0 := time.Now()
		s.session.State = StateThinking

		bytesPerSecond := s.session.Config.SampleRateHz * 2 // 16-bit mono = 2 bytes per sample
		audioDurationMs := int64(len(s.audioBuffer)) * 1000 / int64(bytesPerSecond)
		s.session.TotalAudioMs += audioDurationMs

		// Record User Turn
		s.session.CurrentTurn++
		userTranscript := fmt.Sprintf("User voice command (duration: %dms, energy: %.4f)", audioDurationMs, vadResult.RMSEnergy)
		userTurn := VoiceTurn{
			TurnID:          generateID("turn"),
			Speaker:         SpeakerUser,
			Transcript:      userTranscript,
			AudioDurationMs: audioDurationMs,
			Timestamp:       now,
		}
		s.session.Turns = append(s.session.Turns, userTurn)

		// Generate Agent Response
		s.session.CurrentTurn++
		agentTranscript := fmt.Sprintf("Understood. Executing action based on your voice input (%dms processed).", audioDurationMs)
		latencyMs := time.Since(t0).Milliseconds() + 180 // Simulated TTFAB latency (~180ms)

		// Generate synthesized PCM audio bytes (simulated 440Hz sine wave tone for demonstration)
		responseAudio := generateMockPCMAudio(s.session.Config.SampleRateHz, 600) // 600ms audio

		agentTurn := VoiceTurn{
			TurnID:          generateID("turn"),
			Speaker:         SpeakerAgent,
			Transcript:      agentTranscript,
			AudioDurationMs: 600,
			LatencyMs:       latencyMs,
			WasInterrupted:  false,
			Timestamp:       time.Now().UTC(),
		}
		s.session.Turns = append(s.session.Turns, agentTurn)

		// Reset speech buffer for next cycle
		s.audioBuffer = s.audioBuffer[:0]
		s.session.State = StateSpeaking
		s.session.UpdatedAt = time.Now().UTC()

		return AudioProcessingResult{
			SessionID:           s.session.ID,
			State:               StateSpeaking,
			VAD:                 vadResult,
			InterruptedPriorTTS: interruptedPriorTTS,
			GeneratedTranscript: agentTranscript,
			ResponseAudioBytes:  responseAudio,
			LatencyMs:           latencyMs,
		}
	}

	return AudioProcessingResult{
		SessionID: s.session.ID,
		State:     s.session.State,
		VAD:       vadResult,
	}
}

// GetSnapshot returns a point-in-time copy of the session metadata.
func (s *ActiveSession) GetSnapshot() VoiceSession {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.session
}

// Close marks the session as closed.
func (s *ActiveSession) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.session.State = StateClosed
	s.session.UpdatedAt = time.Now().UTC()
}

func generateMockPCMAudio(sampleRate int, durationMs int) []byte {
	numSamples := sampleRate * durationMs / 1000
	buf := make([]byte, numSamples*2)
	freq := 440.0 // 440Hz A4 tone

	for i := 0; i < numSamples; i++ {
		t := float64(i) / float64(sampleRate)
		sample := int16(math.Sin(2*math.Pi*freq*t) * 12000.0)
		binary.LittleEndian.PutUint16(buf[i*2:i*2+2], uint16(sample))
	}
	return buf
}

func generateID(prefix string) string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%s_%s", prefix, hex.EncodeToString(b))
}
