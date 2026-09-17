package docvision_test

import (
	"context"
	"strings"
	"testing"

	"github.com/lopor-ai/lopor/pkg/docvision"
)

func TestVisionExtractor_Invoice(t *testing.T) {
	extractor := docvision.NewVisionExtractor()

	req := docvision.VisionRequest{
		FileName:    "vendor_invoice_q3.png",
		ImageBase64: "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==",
		MimeType:    "image/png",
		TargetType:  docvision.DocTypeInvoice,
	}

	res, err := extractor.ExtractFromImage(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error during vision extraction: %v", err)
	}

	if res.DocumentType != docvision.DocTypeInvoice {
		t.Errorf("expected invoice document type, got %s", res.DocumentType)
	}

	if res.ConfidenceScore < 0.90 {
		t.Errorf("expected confidence score >= 0.90, got %f", res.ConfidenceScore)
	}

	if len(res.Entities) == 0 {
		t.Errorf("expected extracted entities, got 0")
	}

	if !strings.Contains(res.FullText, "INV-2026-0981") {
		t.Errorf("expected invoice number in full text output")
	}
}

func TestVisionExtractor_InferType(t *testing.T) {
	extractor := docvision.NewVisionExtractor()

	req := docvision.VisionRequest{
		FileName:    "system_arch_diagram.png",
		ImageBase64: "dGVzdF9pbWFnZV9kYXRh",
	}

	res, err := extractor.ExtractFromImage(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.DocumentType != docvision.DocTypeDiagram {
		t.Errorf("expected inferred type diagram, got %s", res.DocumentType)
	}
}

func TestVisionExtractor_EmptyRequest(t *testing.T) {
	extractor := docvision.NewVisionExtractor()
	_, err := extractor.ExtractFromImage(context.Background(), docvision.VisionRequest{
		ImageBase64: "",
		FileName:    "",
	})
	if err == nil {
		t.Errorf("expected error on empty vision request, got nil")
	}
}
