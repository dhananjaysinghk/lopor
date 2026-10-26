package voiceagent

import (
	"encoding/binary"
	"math"
)

// CalculatePCM16RMS computes the root mean square energy over 16-bit PCM mono audio bytes.
func CalculatePCM16RMS(pcmBytes []byte) float64 {
	numSamples := len(pcmBytes) / 2
	if numSamples == 0 {
		return 0.0
	}

	var sumSquares float64
	for i := 0; i < len(pcmBytes)-1; i += 2 {
		sample := int16(binary.LittleEndian.Uint16(pcmBytes[i : i+2]))
		// Normalize sample to [-1.0, 1.0]
		normalized := float64(sample) / 32768.0
		sumSquares += normalized * normalized
	}

	rms := math.Sqrt(sumSquares / float64(numSamples))
	return math.Round(rms*10000) / 10000
}

// VADProcessor maintains stateful frame tracking for speech onset and offset boundaries.
type VADProcessor struct {
	config             VADConfig
	isSpeaking         bool
	speechDurationMs   int
	silenceDurationMs  int
}

// NewVADProcessor initializes a voice activity detector with calibrated thresholds.
func NewVADProcessor(config VADConfig) *VADProcessor {
	if config.SampleRateHz <= 0 {
		config.SampleRateHz = 16000
	}
	if config.EnergyThresholdRMS <= 0.0 {
		config.EnergyThresholdRMS = 0.015
	}
	if config.SilenceTimeoutMs <= 0 {
		config.SilenceTimeoutMs = 500
	}
	if config.MinSpeechDurationMs <= 0 {
		config.MinSpeechDurationMs = 150
	}

	return &VADProcessor{
		config: config,
	}
}

// ProcessFrame evaluates an incoming PCM audio frame and updates speech state.
func (p *VADProcessor) ProcessFrame(pcmBytes []byte, frameDurationMs int) VADResult {
	if frameDurationMs <= 0 {
		frameDurationMs = 50 // default 50ms frame
	}

	rms := CalculatePCM16RMS(pcmBytes)
	isFrameSpeech := rms >= p.config.EnergyThresholdRMS

	var speechStartDetected bool
	var speechEndDetected bool

	if isFrameSpeech {
		p.speechDurationMs += frameDurationMs
		p.silenceDurationMs = 0

		if !p.isSpeaking && p.speechDurationMs >= p.config.MinSpeechDurationMs {
			p.isSpeaking = true
			speechStartDetected = true
		}
	} else {
		p.silenceDurationMs += frameDurationMs
		p.speechDurationMs = 0

		if p.isSpeaking && p.silenceDurationMs >= p.config.SilenceTimeoutMs {
			p.isSpeaking = false
			speechEndDetected = true
		}
	}

	return VADResult{
		IsSpeech:            p.isSpeaking,
		RMSEnergy:           rms,
		SpeechStartDetected: speechStartDetected,
		SpeechEndDetected:   speechEndDetected,
	}
}

// Reset resets stateful speech tracking counters.
func (p *VADProcessor) Reset() {
	p.isSpeaking = false
	p.speechDurationMs = 0
	p.silenceDurationMs = 0
}
