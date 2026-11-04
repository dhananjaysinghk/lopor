package specsynth

import (
	"time"
)

// TargetStandard identifies the target enterprise specification framework.
type TargetStandard string

const (
	StandardOpenAPI30   TargetStandard = "OPENAPI_3_0"
	StandardOpenAPI31   TargetStandard = "OPENAPI_3_1"
	StandardJSONSchema  TargetStandard = "JSON_SCHEMA"
	StandardGraphQLSDL  TargetStandard = "GRAPHQL_SDL"
	StandardProtobufV3  TargetStandard = "PROTOBUF_V3"
	StandardTypeScript  TargetStandard = "TYPESCRIPT_DTS"
)

// GenerationSource specifies the origin material parsed for schema synthesis.
type GenerationSource string

const (
	SourceCodeAnalysis     GenerationSource = "CODE_ANALYSIS"      // Extracted from handlers, structs, routes
	SourceTelemetryTrace   GenerationSource = "TELEMETRY_TRACE"    // Inferred from production JSON payloads
	SourceConversationalReq GenerationSource = "CONVERSATIONAL_REQ" // Synthesized from business directives
)

// EndpointSpec captures an individual synthesized route, parameters, and responses.
type EndpointSpec struct {
	Method          string            `json:"method"`
	Path            string            `json:"path"`
	Summary         string            `json:"summary"`
	Description     string            `json:"description,omitempty"`
	Tags            []string          `json:"tags,omitempty"`
	PathParams      map[string]string `json:"path_params,omitempty"`
	QueryParams     map[string]string `json:"query_params,omitempty"`
	RequestBodyType string            `json:"request_body_type,omitempty"`
	ResponseType    string            `json:"response_type"`
	StatusCode      int               `json:"status_code"`
	IsDeprecated    bool              `json:"is_deprecated"`
}

// ModelDefinition represents a structured data model or type definition.
type ModelDefinition struct {
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Fields      map[string]string `json:"fields"` // FieldName -> DataType
	Required    []string          `json:"required,omitempty"`
}

// SynthesizedSpecification encapsulates the compiled schema, validation metrics, and raw artifact.
type SynthesizedSpecification struct {
	SpecID           string            `json:"spec_id"`
	WorkspaceID      string            `json:"workspace_id"`
	Title            string            `json:"title"`
	Version          string            `json:"version"`
	Standard         TargetStandard    `json:"standard"`
	Source           GenerationSource  `json:"source"`
	Endpoints        []EndpointSpec    `json:"endpoints"`
	Models           []ModelDefinition `json:"models"`
	RawSpecification string            `json:"raw_specification"` // Formatted YAML, JSON, GraphQL, or Proto
	ValidationErrors []string          `json:"validation_errors,omitempty"`
	IsValid          bool              `json:"is_valid"`
	CreatedAt        time.Time         `json:"created_at"`
}

// GenerateSpecRequest specifies parameters to synthesize an API specification.
type GenerateSpecRequest struct {
	WorkspaceID  string           `json:"workspace_id"`
	Title        string           `json:"title"`
	Version      string           `json:"version,omitempty"`
	Standard     TargetStandard   `json:"standard"`
	Source       GenerationSource `json:"source"`
	InputPayload string           `json:"input_payload"` // Source code snippet, JSON samples, or prompt requirements
}

// ValidateSpecRequest defines parameters to validate an existing or synthesized specification.
type ValidateSpecRequest struct {
	Standard TargetStandard `json:"standard"`
	Content  string         `json:"content"`
}

// ValidationResult summarizes semantic consistency and standard compliance.
type ValidationResult struct {
	IsValid     bool     `json:"is_valid"`
	Standard    TargetStandard `json:"standard"`
	IssueCount  int      `json:"issue_count"`
	Errors      []string `json:"errors,omitempty"`
	Warnings    []string `json:"warnings,omitempty"`
	ValidatedAt time.Time `json:"validated_at"`
}

