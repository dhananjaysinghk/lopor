package contextopt_test

import (
	"context"
	"strings"
	"testing"

	"github.com/lopor-ai/lopor/pkg/contextopt"
)

func TestContextCompressor_CodeSafe(t *testing.T) {
	compressor := contextopt.NewContextCompressor()

	rawText := `
It is important to note that in order to initialize the server,
please be advised that you must call NewServer().

` + "```go\nfunc Start() error {\n\treturn nil\n}\n```" + `

Needless to say, at the end of the day, this ensures high availability.
`

	req := contextopt.CompressionRequest{
		Text:         rawText,
		Strategy:     contextopt.StrategyCodeSafe,
		PreserveCode: true,
	}

	res, err := compressor.CompressContext(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error during compression: %v", err)
	}

	if res.CompressedTokens >= res.OriginalTokens {
		t.Errorf("expected token reduction, got orig=%d, comp=%d", res.OriginalTokens, res.CompressedTokens)
	}

	if res.SavingsPercent <= 0 {
		t.Errorf("expected positive savings percentage, got %f", res.SavingsPercent)
	}

	// Code block must remain completely intact
	if !strings.Contains(res.CompressedText, "func Start() error {") {
		t.Errorf("code block was unexpectedly altered during compression")
	}
}

func TestContextCompressor_EmptyText(t *testing.T) {
	compressor := contextopt.NewContextCompressor()
	_, err := compressor.CompressContext(context.Background(), contextopt.CompressionRequest{
		Text: "",
	})
	if err == nil {
		t.Errorf("expected error on empty text input, got nil")
	}
}
