package observability_test

import (
	"context"
	"testing"
	"time"

	"github.com/lopor-ai/lopor/pkg/observability"
)

func TestMetricsCollector(t *testing.T) {
	collector := observability.NewMetricsCollector(nil, nil)

	collector.IncRequests()
	collector.IncRequests()
	collector.IncErrors()
	collector.IncWSConns()

	metrics := collector.CollectMetrics()
	if metrics.TotalRequests != 2 {
		t.Errorf("expected 2 requests, got %d", metrics.TotalRequests)
	}
	if metrics.TotalErrors != 1 {
		t.Errorf("expected 1 error, got %d", metrics.TotalErrors)
	}
	if metrics.ActiveWSConns != 1 {
		t.Errorf("expected 1 active WS connection, got %d", metrics.ActiveWSConns)
	}
	if metrics.Goroutines <= 0 {
		t.Errorf("expected goroutines > 0, got %d", metrics.Goroutines)
	}

	collector.DecWSConns()
	metrics2 := collector.CollectMetrics()
	if metrics2.ActiveWSConns != 0 {
		t.Errorf("expected 0 active WS connections after decrement, got %d", metrics2.ActiveWSConns)
	}
}

func TestDeepHealthCheck(t *testing.T) {
	collector := observability.NewMetricsCollector(nil, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	health := collector.PerformDeepHealthCheck(ctx)
	if health.Status != "HEALTHY" {
		t.Errorf("expected status HEALTHY, got %s", health.Status)
	}
	if health.Components["database"] != "in-memory (mock)" {
		t.Errorf("expected in-memory (mock) db status, got %s", health.Components["database"])
	}
	if health.Components["redis"] != "disabled" {
		t.Errorf("expected disabled redis status, got %s", health.Components["redis"])
	}
}
