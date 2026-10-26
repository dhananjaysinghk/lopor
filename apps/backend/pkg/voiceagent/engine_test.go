package voiceagent

import (
	"encoding/binary"
	"math"
	"testing"
)

// Helper: generates PCM16 audio bytes representing silence or tone
func generateTestPCM16(sampleRate int, durationMs int, amplitude float64, freq float64) []byte {
	numSamples := sampleRate * durationMs / 1000
	buf := make([]byte, numSamples*2)

	for i := 0; i < numSamples; i++ {
		t := float64(i) / float64(sampleRate)
		sample := int16(math.Sin(2*math.Pi*freq*t) * amplitude * 32767.0)
		binary.LittleEndian.PutUint16(buf[i*2:i*2+2], uint16(sample))
	}
	return buf
}

func TestVADRMSCalculation(t *testing.T) {
	sampleRate := 16000

	// 1. Absolute silence
	silence := make([]byte, sampleRate*2/10) // 100ms silence
	silenceRMS := CalculatePCM16RMS(silence)
	if silenceRMS != 0.0 {
		t.Errorf("expected 0.0 RMS for absolute silence, got %f", silenceRMS)
	}

	// 2. Active speech-level tone (50% amplitude)
	tone := generateTestPCM16(sampleRate, 100, 0.5, 300)
	toneRMS := CalculatePCM16RMS(tone)
	// Pure sine wave RMS should be amplitude / sqrt(2) ~= 0.5 / 1.414 ~= 0.353
	if toneRMS < 0.3 || toneRMS > 0.4 {
		t.Errorf("expected tone RMS around ~0.35, got %f", toneRMS)
	}
}

func TestVoiceTurnTakingAndCompletion(t *testing.T) {
	engine := NewEngine()
	wsID := "ws_alpha"
	userID := "usr_1"

	session, err := engine.StartSession(StartSessionRequest{
		WorkspaceID: wsID,
		UserID:      userID,
		Config: &VADConfig{
			EnergyThresholdRMS:  0.02,
			SilenceTimeoutMs:    200, // 200ms silence to commit
			MinSpeechDurationMs: 100, // 100ms speech to trigger start
		},
	})
	if err != nil {
		t.Fatalf("failed to start voice session: %v", err)
	}

	sampleRate := 16000
	speechChunk := generateTestPCM16(sampleRate, 100, 0.4, 400) // 100ms speech
	silenceChunk := make([]byte, sampleRate*2*100/1000)          // 100ms silence

	// 1. Send 2 speech chunks (200ms speech total) -> speech start should trigger
	res1, _ := engine.ProcessChunk(wsID, session.ID, speechChunk, 100)
	if res1.State != StateListening && !res1.VAD.IsSpeech {
		t.Errorf("expected state to transition to Listening")
	}

	res2, _ := engine.ProcessChunk(wsID, session.ID, speechChunk, 100)
	if !res2.VAD.IsSpeech {
		t.Errorf("expected VAD.IsSpeech to be true")
	}

	// 2. Send 2 silence chunks (200ms silence) -> silence timeout should trigger turn commit
	_, _ = engine.ProcessChunk(wsID, session.ID, silenceChunk, 100)
	resFinal, err := engine.ProcessChunk(wsID, session.ID, silenceChunk, 100)
	if err != nil {
		t.Fatalf("unexpected error processing chunk: %v", err)
	}

	if resFinal.State != StateSpeaking {
		t.Errorf("expected state to transition to Speaking, got %s", resFinal.State)
	}
	if len(resFinal.ResponseAudioBytes) == 0 {
		t.Errorf("expected non-empty response audio bytes from synthesized TTS")
	}

	// Verify session snapshot has turns recorded
	snap, _ := engine.GetSession(wsID, session.ID)
	if len(snap.Turns) != 2 {
		t.Errorf("expected 2 turns (1 user + 1 agent), got %d", len(snap.Turns))
	}
}

func TestBargeInInterruption(t *testing.T) {
	engine := NewEngine()
	wsID := "ws_alpha"
	userID := "usr_1"

	session, err := engine.StartSession(StartSessionRequest{
		WorkspaceID: wsID,
		UserID:      userID,
		Config: &VADConfig{
			EnergyThresholdRMS:  0.02,
			SilenceTimeoutMs:    100,
			MinSpeechDurationMs: 50,
		},
	})
	if err != nil {
		t.Fatalf("failed to start session: %v", err)
	}

	sampleRate := 16000
	speechChunk := generateTestPCM16(sampleRate, 50, 0.4, 400)
	silenceChunk := make([]byte, sampleRate*2*50/1000)

	// Step 1: Complete first user turn to enter StateSpeaking
	_, _ = engine.ProcessChunk(wsID, session.ID, speechChunk, 50)
	_, _ = engine.ProcessChunk(wsID, session.ID, speechChunk, 50)
	_, _ = engine.ProcessChunk(wsID, session.ID, silenceChunk, 50)
	resCommit, _ := engine.ProcessChunk(wsID, session.ID, silenceChunk, 50)

	if resCommit.State != StateSpeaking {
		t.Fatalf("expected session to be in StateSpeaking, got %s", resCommit.State)
	}

	// Step 2: User speaks while agent is speaking (Barge-in)
	resBargeIn, _ := engine.ProcessChunk(wsID, session.ID, speechChunk, 50)
	if !resBargeIn.InterruptedPriorTTS {
		t.Errorf("expected InterruptedPriorTTS to be true on barge-in")
	}
	if resBargeIn.State != StateListening {
		t.Errorf("expected session state to abort to Listening, got %s", resBargeIn.State)
	}

	snap, _ := engine.GetSession(wsID, session.ID)
	if snap.InterruptionCount != 1 {
		t.Errorf("expected InterruptionCount 1, got %d", snap.InterruptionCount)
	}
	if len(snap.Turns) > 1 && !snap.Turns[1].WasInterrupted {
		t.Errorf("expected prior agent turn to be flagged as WasInterrupted")
	}
}

func TestWorkspaceTenancyIsolation(t *testing.T) {
	engine := NewEngine()

	session, _ := engine.StartSession(StartSessionRequest{
		WorkspaceID: "ws_alpha",
		UserID:      "usr_1",
	})

	// Access from ws_beta should fail with ErrSessionNotFound
	_, err := engine.GetSession("ws_beta", session.ID)
	if err != ErrSessionNotFound {
		t.Errorf("expected ErrSessionNotFound for unauthorized workspace, got %v", err)
	}

	silenceChunk := make([]byte, 160)
	_, err = engine.ProcessChunk("ws_beta", session.ID, silenceChunk, 10)
	if err != ErrSessionNotFound {
		t.Errorf("expected ErrSessionNotFound for cross-tenant process chunk, got %v", err)
	}
}
