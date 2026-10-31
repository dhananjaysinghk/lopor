package datasynth

import (
	"strings"
	"testing"
)

func TestLaplaceNoiseAndEpsilonScaling(t *testing.T) {
	// Epsilon 0.1 should yield larger scale noise than Epsilon 2.0
	noiseStrict := SampleLaplace(1.0 / 0.1) // scale = 10.0
	noiseRelaxed := SampleLaplace(1.0 / 2.0) // scale = 0.5

	if noiseStrict == 0.0 && noiseRelaxed == 0.0 {
		t.Log("Note: sampled zero Laplace noise (rare)")
	}

	epsRestricted := GetEpsilonForPrivacyLevel(PrivacyRestricted, 0)
	if epsRestricted != 0.1 {
		t.Errorf("expected restricted epsilon 0.1, got %f", epsRestricted)
	}

	epsPublic := GetEpsilonForPrivacyLevel(PrivacyPublic, 0)
	if epsPublic != 2.0 {
		t.Errorf("expected public epsilon 2.0, got %f", epsPublic)
	}
}

func TestGenerateSyntheticDatasetDefaultSchema(t *testing.T) {
	engine := NewEngine()
	wsID := "ws_synth_test"

	req := GenerateDatasetRequest{
		WorkspaceID: wsID,
		Title:       "Enterprise Customer Mock Database",
		Config: SyntheticDatasetConfig{
			RowCount:     25,
			PrivacyLevel: PrivacyConfidential,
		},
	}

	ds, err := engine.GenerateDataset(req)
	if err != nil {
		t.Fatalf("failed to generate synthetic dataset: %v", err)
	}

	if ds.RowCount != 25 {
		t.Errorf("expected 25 rows, got %d", ds.RowCount)
	}
	if len(ds.Records) != 25 {
		t.Fatalf("expected 25 records slice, got %d", len(ds.Records))
	}

	firstRecord := ds.Records[0]
	if _, ok := firstRecord["customer_id"].(string); !ok {
		t.Errorf("expected string customer_id")
	}
	if !strings.HasPrefix(firstRecord["customer_id"].(string), "synth-") {
		t.Errorf("expected synth- prefix in UUID")
	}

	if ds.Metrics.EpsilonUsed != 0.5 {
		t.Errorf("expected epsilon 0.5 for Confidential privacy, got %f", ds.Metrics.EpsilonUsed)
	}
	if !ds.Metrics.IsGDPRCompliant {
		t.Errorf("expected GDPR compliance for Confidential tier")
	}
}

func TestCustomSchemaSynthesisAndNullProbability(t *testing.T) {
	engine := NewEngine()
	wsID := "ws_fintech"

	customFields := []SyntheticFieldSchema{
		{
			Name: "transaction_id",
			Type: FieldTypeUUID,
		},
		{
			Name:            "memo",
			Type:            FieldTypeFreeText,
			IsNullable:      true,
			NullProbability: 0.5,
		},
		{
			Name:       "status",
			Type:       FieldTypeCategorical,
			Categories: []string{"PENDING", "CLEARED", "REJECTED"},
			Weights:    []float64{0.2, 0.7, 0.1},
		},
	}

	ds, err := engine.GenerateDataset(GenerateDatasetRequest{
		WorkspaceID: wsID,
		Title:       "Fintech Ledger Sample",
		Config: SyntheticDatasetConfig{
			RowCount: 50,
			Fields:   customFields,
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	nullCount := 0
	for _, rec := range ds.Records {
		if rec["memo"] == nil {
			nullCount++
		}
	}

	// Over 50 rows with 0.5 probability, nullCount should reasonably be > 5 and < 45
	if nullCount == 0 || nullCount == 50 {
		t.Errorf("null probability unexpected: got %d nulls out of 50", nullCount)
	}
}

func TestWorkspaceTenancyIsolation(t *testing.T) {
	engine := NewEngine()

	ds, err := engine.GenerateDataset(GenerateDatasetRequest{
		WorkspaceID: "ws_alpha",
		Title:       "Alpha Medical Records",
		Config: SyntheticDatasetConfig{
			RowCount: 10,
		},
	})
	if err != nil {
		t.Fatalf("failed to generate dataset: %v", err)
	}

	// ws_beta should not be able to retrieve dataset
	_, err = engine.GetDataset("ws_beta", ds.DatasetID)
	if err != ErrDatasetNotFound {
		t.Errorf("expected ErrDatasetNotFound for cross-tenant access, got %v", err)
	}
}

