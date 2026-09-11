package secscan_test

import (
	"context"
	"testing"

	"github.com/lopor-ai/lopor/pkg/secscan"
)

func TestScanner_DetectSecrets(t *testing.T) {
	scanner := secscan.NewScanner()

	sampleCode := `
package main

const awsKey = "AKIA1234567890ABCDEF"
const openAI = "sk-1234567890abcdef1234567890abcdef12"

func queryUser(id string) string {
	return "SELECT * FROM users WHERE id = " + id
}
`

	res, err := scanner.ScanCode(context.Background(), secscan.ScanRequest{
		FileName: "config.go",
		Content:  sampleCode,
	})
	if err != nil {
		t.Fatalf("unexpected error scanning code: %v", err)
	}

	if res.Passed {
		t.Errorf("expected scan to fail due to critical secrets, got passed=true")
	}

	if res.CriticalCount < 2 {
		t.Errorf("expected at least 2 critical secrets (AWS + OpenAI), got %d", res.CriticalCount)
	}

	if res.HighCount < 1 {
		t.Errorf("expected at least 1 high issue (SQL injection risk), got %d", res.HighCount)
	}

	if len(res.Vulnerabilities) < 3 {
		t.Errorf("expected at least 3 vulnerabilities, got %d", len(res.Vulnerabilities))
	}
}

func TestScanner_CleanCode(t *testing.T) {
	scanner := secscan.NewScanner()

	cleanCode := `
package main

import "os"

func getAPIKey() string {
	return os.Getenv("OPENAI_API_KEY")
}
`

	res, err := scanner.ScanCode(context.Background(), secscan.ScanRequest{
		FileName: "main.go",
		Content:  cleanCode,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !res.Passed {
		t.Errorf("expected clean code to pass, got passed=false")
	}

	if res.TotalIssues != 0 {
		t.Errorf("expected 0 issues in clean code, got %d", res.TotalIssues)
	}
}
