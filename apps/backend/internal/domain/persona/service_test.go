package persona_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/lopor-ai/lopor/internal/domain/persona"
)

func TestPersonaService(t *testing.T) {
	repo := persona.NewRepository(nil)
	svc := persona.NewService(repo)

	wsID := uuid.New()
	userID := uuid.New()

	// 1. Test Create Persona
	p := &persona.PersonaRecord{
		WorkspaceID:  wsID,
		Name:         "Senior Code Auditor",
		SystemPrompt: "Review code thoroughly for security vulnerabilities, memory leaks, and performance bottlenecks.",
		Tone:         "authoritative",
		Temperature:  0.2,
		MaxTokens:    8192,
		Guardrails:   `["No hallucinated APIs", "Always cite line numbers"]`,
		IsDefault:    true,
		CreatedBy:    userID,
	}

	created, err := svc.CreatePersona(context.Background(), p)
	if err != nil {
		t.Fatalf("failed to create persona: %v", err)
	}

	if created.ID == uuid.Nil {
		t.Errorf("expected generated UUID for persona, got Nil")
	}

	// 2. Test GetWorkspacePersonas
	personas, err := svc.GetWorkspacePersonas(context.Background(), wsID)
	if err != nil {
		t.Fatalf("failed to fetch workspace personas: %v", err)
	}
	if len(personas) != 1 {
		t.Errorf("expected 1 persona, got %d", len(personas))
	}

	// 3. Test FormatSystemPrompt
	formatted := svc.FormatSystemPrompt(created)
	if !testing.Verbose() {
		t.Logf("Formatted prompt: %s", formatted)
	}
	if formatted == "" {
		t.Errorf("expected non-empty formatted system prompt")
	}

	// 4. Test Set Default Persona
	p2 := &persona.PersonaRecord{
		WorkspaceID:  wsID,
		Name:         "Creative Technical Writer",
		SystemPrompt: "Write documentation with clear code examples.",
		IsDefault:    false,
	}
	created2, err := svc.CreatePersona(context.Background(), p2)
	if err != nil {
		t.Fatalf("failed to create second persona: %v", err)
	}

	err = svc.SetDefaultPersona(context.Background(), wsID, created2.ID)
	if err != nil {
		t.Fatalf("failed to set default persona: %v", err)
	}

	// Verify p2 is now default
	fetched2, err := svc.GetPersonaByID(context.Background(), created2.ID)
	if err != nil {
		t.Fatalf("failed to get persona by ID: %v", err)
	}
	if !fetched2.IsDefault {
		t.Errorf("expected persona2 to be default after SetDefaultPersona")
	}
}
