package testgen_test

import (
	"context"
	"testing"

	"github.com/lopor-ai/lopor/pkg/testgen"
)

func TestGenerateTestSuite_Go(t *testing.T) {
	gen := testgen.NewGenerator()
	req := testgen.TestGenRequest{
		SourceCode:      "package main\n\nfunc Add(a, b int) int { return a + b }",
		SourceFileName:  "calculator.go",
		Language:        testgen.LangGo,
		TestType:        testgen.TestTypeUnit,
		IncludeMocks:    true,
		IncludeEdgeCase: true,
	}

	res, err := gen.GenerateTestSuite(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.TestFileName != "calculator_test.go" {
		t.Errorf("expected test file calculator_test.go, got %s", res.TestFileName)
	}

	if res.Framework != "testing" {
		t.Errorf("expected framework testing, got %s", res.Framework)
	}

	if len(res.TestCases) == 0 {
		t.Errorf("expected generated test cases, got 0")
	}

	if res.EstimatedCoverage < 50.0 {
		t.Errorf("expected estimated coverage >= 50, got %f", res.EstimatedCoverage)
	}
}

func TestGenerateTestSuite_TypeScript(t *testing.T) {
	gen := testgen.NewGenerator()
	req := testgen.TestGenRequest{
		SourceCode:     "export function multiply(a: number, b: number): number { return a * b; }",
		SourceFileName: "math.ts",
		Language:       testgen.LangTypeScript,
		Framework:      "vitest",
		IncludeMocks:   true,
	}

	res, err := gen.GenerateTestSuite(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.TestFileName != "math.test.ts" {
		t.Errorf("expected math.test.ts, got %s", res.TestFileName)
	}

	if res.Framework != "vitest" {
		t.Errorf("expected vitest, got %s", res.Framework)
	}
}

func TestGenerateTestSuite_EmptySource(t *testing.T) {
	gen := testgen.NewGenerator()
	req := testgen.TestGenRequest{
		SourceCode: "",
	}

	_, err := gen.GenerateTestSuite(context.Background(), req)
	if err == nil {
		t.Fatalf("expected error for empty source code, got nil")
	}
}
