package tts_test

import (
	"context"
	"testing"

	"github.com/lopor-ai/lopor/pkg/tts"
)

func TestTTSEngine_AvailableVoices(t *testing.T) {
	engine := tts.NewTTSEngine()
	voices := engine.GetAvailableVoices()

	if len(voices) < 6 {
		t.Errorf("expected at least 6 neural voices in catalog, got %d", len(voices))
	}

	foundAlloy := false
	for _, v := range voices {
		if v.ID == "alloy" {
			foundAlloy = true
			if len(v.SupportedFormats) == 0 {
				t.Errorf("expected alloy to support audio formats")
			}
			break
		}
	}

	if !foundAlloy {
		t.Errorf("expected alloy voice in catalog")
	}
}

func TestTTSEngine_SynthesizeSpeech(t *testing.T) {
	engine := tts.NewTTSEngine()

	req := tts.SynthesisRequest{
		Text:    "Antigravity and Lopor provide enterprise autonomous agent capabilities with clean architecture.",
		VoiceID: "nova",
		Speed:   1.25,
		Format:  tts.FormatMP3,
		Quality: "hd",
	}

	res, err := engine.SynthesizeSpeech(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error synthesizing speech: %v", err)
	}

	if res.VoiceID != "nova" {
		t.Errorf("expected voice nova, got %s", res.VoiceID)
	}

	if res.DurationSeconds <= 0 {
		t.Errorf("expected positive duration, got %f", res.DurationSeconds)
	}

	if res.WordCount <= 0 {
		t.Errorf("expected word count > 0, got %d", res.WordCount)
	}

	if res.AudioBase64 == "" {
		t.Errorf("expected audio base64 payload to be non-empty")
	}

	if res.SampleRateHz != 48000 {
		t.Errorf("expected 48000 sample rate for hd quality, got %d", res.SampleRateHz)
	}
}

func TestTTSEngine_GenerateAudioBrief(t *testing.T) {
	engine := tts.NewTTSEngine()

	req := tts.AudioBriefRequest{
		Title:        "Quarterly Financial Architecture Plan",
		Content:      "# Overview\n- Migrated to pgvector\n- Deployed multi-agent consensus\n- Reduced token spend by 40%",
		VoiceID:      "onyx",
		Format:       tts.FormatWAV,
		IncludeIntro: true,
	}

	res, err := engine.GenerateAudioBrief(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error generating audio brief: %v", err)
	}

	if res.VoiceID != "onyx" {
		t.Errorf("expected onyx voice, got %s", res.VoiceID)
	}

	if res.Format != tts.FormatWAV {
		t.Errorf("expected wav format, got %s", res.Format)
	}

	if res.DurationSeconds <= 0 {
		t.Errorf("expected duration > 0, got %f", res.DurationSeconds)
	}
}
