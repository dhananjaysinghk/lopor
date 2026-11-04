package specsynth

import (
	"strings"
	"testing"
)

func TestParseCodeEndpointsAndSynthesizeOpenAPI(t *testing.T) {
	engine := NewEngine()
	wsID := "ws_alpha"

	codeSample := `
type UserProfile struct {
	ID        string ` + "`" + `json:"id"` + "`" + `
	Email     string ` + "`" + `json:"email"` + "`" + `
	Age       int    ` + "`" + `json:"age"` + "`" + `
	IsActive  bool   ` + "`" + `json:"is_active"` + "`" + `
}

func RegisterRoutes(router *fiber.App) {
	router.Get("/api/v1/users/:userId", GetUserHandler)
	router.Post("/api/v1/users", CreateUserHandler)
}
`

	spec, err := engine.GenerateSpecification(GenerateSpecRequest{
		WorkspaceID:  wsID,
		Title:        "Enterprise User Management API",
		Standard:     StandardOpenAPI30,
		Source:       SourceCodeAnalysis,
		InputPayload: codeSample,
	})
	if err != nil {
		t.Fatalf("failed to generate specification: %v", err)
	}

	if len(spec.Endpoints) != 2 {
		t.Fatalf("expected 2 extracted endpoints, got %d", len(spec.Endpoints))
	}
	if len(spec.Models) != 1 {
		t.Fatalf("expected 1 extracted model, got %d", len(spec.Models))
	}

	raw := spec.RawSpecification
	if !strings.Contains(raw, "openapi: 3.0.3") {
		t.Errorf("expected openapi: 3.0.3 in output: %s", raw)
	}
	if !strings.Contains(raw, "/api/v1/users/:userId:") {
		t.Errorf("expected /api/v1/users/:userId in output paths")
	}
	if !strings.Contains(raw, "UserProfile:") {
		t.Errorf("expected UserProfile in components schema")
	}
}

func TestGenerateGraphQLAndProtobuf(t *testing.T) {
	engine := NewEngine()
	wsID := "ws_alpha"

	codeSample := `
type PaymentOrder struct {
	OrderID string  ` + "`" + `json:"order_id"` + "`" + `
	Amount  float64 ` + "`" + `json:"amount"` + "`" + `
	Paid    bool    ` + "`" + `json:"paid"` + "`" + `
}
`

	// 1. GraphQL SDL
	specGQL, err := engine.GenerateSpecification(GenerateSpecRequest{
		WorkspaceID:  wsID,
		Title:        "Payment GraphQL Schema",
		Standard:     StandardGraphQLSDL,
		InputPayload: codeSample,
	})
	if err != nil {
		t.Fatalf("failed to generate GraphQL spec: %v", err)
	}
	if !strings.Contains(specGQL.RawSpecification, "type PaymentOrder {") {
		t.Errorf("expected type PaymentOrder in GraphQL SDL: %s", specGQL.RawSpecification)
	}

	// 2. Protobuf v3
	specProto, err := engine.GenerateSpecification(GenerateSpecRequest{
		WorkspaceID:  wsID,
		Title:        "payment.service.v1",
		Standard:     StandardProtobufV3,
		InputPayload: codeSample,
	})
	if err != nil {
		t.Fatalf("failed to generate Proto spec: %v", err)
	}
	if !strings.Contains(specProto.RawSpecification, "syntax = \"proto3\";") {
		t.Errorf("expected proto3 syntax declaration: %s", specProto.RawSpecification)
	}
	if !strings.Contains(specProto.RawSpecification, "message PaymentOrder {") {
		t.Errorf("expected message PaymentOrder in Protobuf: %s", specProto.RawSpecification)
	}
}

func TestValidateSpecification(t *testing.T) {
	engine := NewEngine()

	// Valid OpenAPI
	validYAML := "openapi: 3.0.3\ninfo:\n  title: Test\npaths:\n  /test:\n    get:\n      responses:\n        '200':\n"
	valRes := engine.ValidateSpecification(ValidateSpecRequest{
		Standard: StandardOpenAPI30,
		Content:  validYAML,
	})
	if !valRes.IsValid {
		t.Errorf("expected valid OpenAPI validation, got errors: %v", valRes.Errors)
	}

	// Invalid OpenAPI (missing openapi header)
	invalidYAML := "info:\n  title: Broken\n"
	invRes := engine.ValidateSpecification(ValidateSpecRequest{
		Standard: StandardOpenAPI30,
		Content:  invalidYAML,
	})
	if invRes.IsValid {
		t.Errorf("expected invalid result for broken spec")
	}
}

func TestWorkspaceTenancyIsolation(t *testing.T) {
	engine := NewEngine()

	spec, _ := engine.GenerateSpecification(GenerateSpecRequest{
		WorkspaceID:  "ws_alpha",
		Title:        "Alpha Spec",
		InputPayload: "type Alpha struct {}",
	})

	_, err := engine.GetSpecification("ws_beta", spec.SpecID)
	if err != ErrSpecNotFound {
		t.Errorf("expected ErrSpecNotFound for cross-tenant access, got %v", err)
	}
}

