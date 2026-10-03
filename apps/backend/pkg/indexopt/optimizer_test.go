package indexopt_test

import (
	"context"
	"strings"
	"testing"

	"github.com/lopor-ai/lopor/pkg/indexopt"
)

func TestIndexOptimizer_InspectHealth(t *testing.T) {
	optimizer := indexopt.NewIndexOptimizer()

	// Moderate dataset
	health, err := optimizer.InspectHealth(context.Background(), "ws-idx-test", 15000)
	if err != nil {
		t.Fatalf("unexpected error inspecting health: %v", err)
	}

	if health.VectorDimensions != 1536 {
		t.Errorf("expected 1536 vector dimensions, got %d", health.VectorDimensions)
	}

	if health.IndexType != indexopt.IndexHNSW {
		t.Errorf("expected HNSW index type, got %s", health.IndexType)
	}

	if health.EstimatedRecall <= 0.8 || health.EstimatedRecall > 1.0 {
		t.Errorf("expected valid estimated recall between 0.8 and 1.0, got %f", health.EstimatedRecall)
	}

	// Large fragmented dataset
	largeHealth, err := optimizer.InspectHealth(context.Background(), "ws-idx-large", 200000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !largeHealth.NeedsCompaction {
		t.Errorf("expected large dataset with bloat to need compaction")
	}

	if largeHealth.FragmentationLevel != "SEVERE" {
		t.Errorf("expected SEVERE fragmentation, got %s", largeHealth.FragmentationLevel)
	}
}

func TestIndexOptimizer_CalibrateHNSW(t *testing.T) {
	optimizer := indexopt.NewIndexOptimizer()

	req := indexopt.CalibrationRequest{
		TargetRecall:           0.99,
		MaxAcceptableLatencyMs: 20.0,
		EstimatedDatasetSize:   50000,
		HardwareProfile:        "BALANCED",
	}

	params, err := optimizer.CalibrateHNSW(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error calibrating HNSW: %v", err)
	}

	if params.M < 16 {
		t.Errorf("expected M >= 16 for high recall, got %d", params.M)
	}

	if params.EfConstruction < 64 {
		t.Errorf("expected EfConstruction >= 64, got %d", params.EfConstruction)
	}

	if params.MemoryOverheadMB <= 0 {
		t.Errorf("expected positive memory overhead MB, got %f", params.MemoryOverheadMB)
	}

	if !strings.Contains(params.SQLCommand, "CREATE INDEX CONCURRENTLY") {
		t.Errorf("expected SQLCommand to contain CREATE INDEX CONCURRENTLY, got %s", params.SQLCommand)
	}
}

func TestIndexOptimizer_GenerateOptimizationPlan(t *testing.T) {
	optimizer := indexopt.NewIndexOptimizer()

	report, err := optimizer.GenerateOptimizationPlan(context.Background(), "ws-plan-test", 25000)
	if err != nil {
		t.Fatalf("unexpected error generating plan: %v", err)
	}

	if len(report.Actions) < 2 {
		t.Errorf("expected at least 2 optimization actions, got %d", len(report.Actions))
	}

	foundVacuum := false
	for _, a := range report.Actions {
		if a.ActionType == "VACUUM_ANALYZE" {
			foundVacuum = true
			break
		}
	}

	if !foundVacuum {
		t.Errorf("expected VACUUM_ANALYZE in optimization actions")
	}

	if report.Status != "OPTIMIZATION_PLAN_GENERATED" {
		t.Errorf("expected status OPTIMIZATION_PLAN_GENERATED, got %s", report.Status)
	}
}
