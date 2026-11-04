package specsynth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

var (
	ErrSpecNotFound   = errors.New("synthesized specification not found")
	ErrInvalidRequest = errors.New("invalid specification generation request parameters")
)

// SpecSynthEngine coordinates code analysis, telemetry inference, schema compilation, and verification.
type SpecSynthEngine struct {
	mu    sync.RWMutex
	specs map[string]*SynthesizedSpecification
}

// NewEngine initializes a new SpecSynthEngine.
func NewEngine() *SpecSynthEngine {
	return &SpecSynthEngine{
		specs: make(map[string]*SynthesizedSpecification),
	}
}

// GenerateSpecification synthesizes an enterprise API spec from code, JSON payloads, or requirements.
func (e *SpecSynthEngine) GenerateSpecification(req GenerateSpecRequest) (*SynthesizedSpecification, error) {
	if req.WorkspaceID == "" || req.Title == "" || req.InputPayload == "" {
		return nil, fmt.Errorf("%w: workspace_id, title, and input_payload are required", ErrInvalidRequest)
	}

	standard := req.Standard
	if standard == "" {
		standard = StandardOpenAPI30
	}

	version := req.Version
	if version == "" {
		version = "1.0.0"
	}

	var endpoints []EndpointSpec
	var models []ModelDefinition

	switch req.Source {
	case SourceTelemetryTrace:
		// Infer schema model from JSON sample payload
		fields, err := InferSchemaFromJSON(req.InputPayload)
		if err == nil {
			models = append(models, ModelDefinition{
				Name:   "InferredPayload",
				Fields: fields,
			})
			endpoints = append(endpoints, EndpointSpec{
				Method:          "POST",
				Path:            "/api/v1/telemetry",
				Summary:         "Inferred Telemetry Endpoint",
				RequestBodyType: "InferredPayload",
				ResponseType:    "StandardResponse",
				StatusCode:      201,
			})
		}

	case SourceCodeAnalysis:
		fallthrough
	default:
		endpoints, models = ParseCodeEndpoints(req.InputPayload)
		if len(endpoints) == 0 && len(models) == 0 {
			// If input payload is prose/conversational, generate base template
			models = append(models, ModelDefinition{
				Name: "GenericResource",
				Fields: map[string]string{
					"id":         "string",
					"created_at": "string (date-time)",
				},
			})
			endpoints = append(endpoints, EndpointSpec{
				Method:       "GET",
				Path:         "/api/v1/resources",
				Summary:      "List Resources",
				ResponseType: "GenericResource",
				StatusCode:   200,
			})
		}
	}

	var rawSpec string
	switch standard {
	case StandardGraphQLSDL:
		rawSpec = FormatGraphQLSDL(models)
	case StandardProtobufV3:
		rawSpec = FormatProtobufV3(req.Title, models)
	case StandardTypeScript:
		rawSpec = FormatTypeScript(models)
	case StandardOpenAPI30, StandardOpenAPI31:
		fallthrough
	default:
		rawSpec = FormatOpenAPI30(req.Title, version, endpoints, models)
	}

	specID := generateID("spec")
	now := time.Now().UTC()

	spec := &SynthesizedSpecification{
		SpecID:           specID,
		WorkspaceID:      req.WorkspaceID,
		Title:            req.Title,
		Version:          version,
		Standard:         standard,
		Source:           req.Source,
		Endpoints:        endpoints,
		Models:           models,
		RawSpecification: rawSpec,
		IsValid:          true,
		CreatedAt:        now,
	}

	e.mu.Lock()
	e.specs[specID] = spec
	e.mu.Unlock()

	return spec, nil
}

// ValidateSpecification analyzes an API specification for standard compliance and semantic errors.
func (e *SpecSynthEngine) ValidateSpecification(req ValidateSpecRequest) ValidationResult {
	var errorsList []string
	var warnings []string

	trimmed := strings.TrimSpace(req.Content)
	if trimmed == "" {
		errorsList = append(errorsList, "Specification content is empty.")
	}

	switch req.Standard {
	case StandardOpenAPI30, StandardOpenAPI31:
		if !strings.Contains(trimmed, "openapi:") {
			errorsList = append(errorsList, "Missing required 'openapi: 3.x' header.")
		}
		if !strings.Contains(trimmed, "paths:") {
			warnings = append(warnings, "Specification contains no defined 'paths:'.")
		}

	case StandardProtobufV3:
		if !strings.Contains(trimmed, "syntax = \"proto3\";") {
			errorsList = append(errorsList, "Missing 'syntax = \"proto3\";' declaration.")
		}

	case StandardGraphQLSDL:
		if !strings.Contains(trimmed, "type ") {
			warnings = append(warnings, "No GraphQL object types defined.")
		}
	}

	isValid := len(errorsList) == 0
	return ValidationResult{
		IsValid:     isValid,
		Standard:    req.Standard,
		IssueCount:  len(errorsList) + len(warnings),
		Errors:      errorsList,
		Warnings:    warnings,
		ValidatedAt: time.Now().UTC(),
	}
}

// GetSpecification retrieves a synthesized specification by ID with workspace tenancy check.
func (e *SpecSynthEngine) GetSpecification(workspaceID, specID string) (*SynthesizedSpecification, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	spec, exists := e.specs[specID]
	if !exists || spec.WorkspaceID != workspaceID {
		return nil, ErrSpecNotFound
	}

	return spec, nil
}

// ListSpecifications returns all synthesized specs for a workspace.
func (e *SpecSynthEngine) ListSpecifications(workspaceID string) []*SynthesizedSpecification {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var result []*SynthesizedSpecification
	for _, s := range e.specs {
		if s.WorkspaceID == workspaceID {
			result = append(result, s)
		}
	}
	return result
}

func generateID(prefix string) string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%s_%s", prefix, hex.EncodeToString(b))
}

