package translation_test

import (
	"context"
	"strings"
	"testing"

	"github.com/lopor-ai/lopor/pkg/translation"
)

func TestTranslationEngine_SupportedLanguages(t *testing.T) {
	engine := translation.NewTranslationEngine()
	langs := engine.GetSupportedLanguages()

	if len(langs) < 8 {
		t.Errorf("expected at least 8 supported languages, got %d", len(langs))
	}

	foundJa := false
	for _, l := range langs {
		if l.Code == "ja" {
			foundJa = true
			if l.Name != "Japanese" {
				t.Errorf("expected Japanese name, got %s", l.Name)
			}
			break
		}
	}

	if !foundJa {
		t.Errorf("expected Japanese in supported languages")
	}
}

func TestTranslationEngine_DetectLanguage(t *testing.T) {
	engine := translation.NewTranslationEngine()

	// Japanese detection
	resJa, err := engine.DetectLanguage(context.Background(), "これは日本語のテキストです。")
	if err != nil || resJa.DetectedLanguage != "ja" {
		t.Errorf("expected ja detection, got %v (err: %v)", resJa, err)
	}

	// German detection
	resDe, err := engine.DetectLanguage(context.Background(), "Guten Tag, das ist ein Test und Dokumentation.")
	if err != nil || resDe.DetectedLanguage != "de" {
		t.Errorf("expected de detection, got %v (err: %v)", resDe, err)
	}
}

func TestTranslationEngine_TranslateMarkdownPreservingCode(t *testing.T) {
	engine := translation.NewTranslationEngine()

	markdownDoc := `# Architecture Overview
Welcome to the system.
` + "```go\nfunc Main() {\n\tprintln(\"Security\")\n}\n```" + `
Features and Performance details.`

	req := translation.TranslationRequest{
		Text:               markdownDoc,
		TargetLanguage:     "es",
		PreserveCodeBlocks: true,
		Glossary: []translation.TranslationGlossary{
			{
				Term:       "system",
				TargetTerm: "Lopor Platform",
			},
		},
	}

	res, err := engine.Translate(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected translation error: %v", err)
	}

	// Verify code block is intact
	if !strings.Contains(res.TranslatedText, "func Main() {") {
		t.Errorf("expected code block to be preserved unaltered, got:\n%s", res.TranslatedText)
	}

	// Verify translated phrases
	if !strings.Contains(res.TranslatedText, "Arquitectura") {
		t.Errorf("expected 'Architecture' translated to 'Arquitectura', got:\n%s", res.TranslatedText)
	}

	// Verify glossary term applied
	if !strings.Contains(res.TranslatedText, "Lopor Platform") {
		t.Errorf("expected glossary term 'Lopor Platform' to be applied")
	}

	if res.CodeBlocksPreserved != 1 {
		t.Errorf("expected 1 preserved code block, got %d", res.CodeBlocksPreserved)
	}
}
