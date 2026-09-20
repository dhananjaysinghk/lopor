package sqlsynth_test

import (
	"context"
	"strings"
	"testing"

	"github.com/lopor-ai/lopor/pkg/sqlsynth"
)

func TestSQLSynthesizer_Generate(t *testing.T) {
	synthesizer := sqlsynth.NewSQLSynthesizer()

	req := sqlsynth.SQLGenerateRequest{
		NaturalQuery: "Show me all users created in the system",
		Dialect:      "postgres",
	}

	res, err := synthesizer.GenerateSQL(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error synthesizing SQL: %v", err)
	}

	if !res.IsSafe {
		t.Errorf("expected generated SQL to be safe, got unsafe")
	}

	if !strings.HasPrefix(res.GeneratedSQL, "SELECT") {
		t.Errorf("expected SELECT query, got %s", res.GeneratedSQL)
	}

	if res.Dialect != "postgres" {
		t.Errorf("expected postgres dialect, got %s", res.Dialect)
	}
}

func TestSQLSynthesizer_EmptyQuery(t *testing.T) {
	synthesizer := sqlsynth.NewSQLSynthesizer()

	_, err := synthesizer.GenerateSQL(context.Background(), sqlsynth.SQLGenerateRequest{
		NaturalQuery: "",
	})
	if err == nil {
		t.Errorf("expected error on empty natural query, got nil")
	}
}
