package anonymizer_test

import (
	"context"
	"strings"
	"testing"

	"github.com/lopor-ai/lopor/pkg/anonymizer"
)

func TestPIIAnonymizer_Redaction(t *testing.T) {
	anonymizerInstance := anonymizer.NewPIIAnonymizer()

	sensitiveText := `
Client John Doe with email john.doe@enterprise.corp contacted support.
His social security number is 123-45-6789 and his phone is 555-019-2831.
Payment was submitted via card 4111-2222-3333-4444 from host 192.168.1.100.
`

	req := anonymizer.AnonymizeRequest{
		Text:          sensitiveText,
		EnableMapping: true,
	}

	res, err := anonymizerInstance.AnonymizeText(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error anonymizing text: %v", err)
	}

	if res.TotalRedactions < 4 {
		t.Errorf("expected at least 4 redactions (email, ssn, phone, card), got %d", res.TotalRedactions)
	}

	// Verify original sensitive strings are completely absent from RedactedText
	if strings.Contains(res.RedactedText, "john.doe@enterprise.corp") {
		t.Errorf("email was not redacted from output text")
	}
	if strings.Contains(res.RedactedText, "123-45-6789") {
		t.Errorf("SSN was not redacted from output text")
	}
	if strings.Contains(res.RedactedText, "4111-2222-3333-4444") {
		t.Errorf("credit card was not redacted from output text")
	}

	// Verify token map for reversible de-anonymization
	if len(res.TokenMap) == 0 {
		t.Errorf("expected token map to be populated when EnableMapping is true")
	}
}

func TestPIIAnonymizer_EmptyText(t *testing.T) {
	anonymizerInstance := anonymizer.NewPIIAnonymizer()
	_, err := anonymizerInstance.AnonymizeText(context.Background(), anonymizer.AnonymizeRequest{
		Text: "",
	})
	if err == nil {
		t.Errorf("expected error on empty text input, got nil")
	}
}
