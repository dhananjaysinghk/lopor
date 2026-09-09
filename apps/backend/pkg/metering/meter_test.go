package metering_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/lopor-ai/lopor/pkg/metering"
)

func TestMeterService_FreeTierQuotas(t *testing.T) {
	service := metering.NewMeterService()
	wsID := uuid.New()

	meter := service.GetWorkspaceUsage(context.Background(), wsID)
	if meter.Tier != metering.TierFree {
		t.Errorf("expected default tier to be free, got %s", meter.Tier)
	}

	if meter.Quotas.MaxTokensPerMonth != 100_000 {
		t.Errorf("expected 100k max tokens for free tier, got %d", meter.Quotas.MaxTokensPerMonth)
	}

	// Record token usage
	updated, err := service.RecordUsage(context.Background(), wsID, metering.UsageRecordReq{
		TokenType: "ai_tokens",
		Quantity:  85_000,
	})
	if err != nil {
		t.Fatalf("unexpected error recording usage: %v", err)
	}

	if updated.Usage.TokensUsed != 85_000 {
		t.Errorf("expected 85,000 tokens used, got %d", updated.Usage.TokensUsed)
	}

	if len(updated.Warnings) == 0 {
		t.Errorf("expected warning when usage exceeds 80%% of quota")
	}

	// Upgrade to Pro Tier
	upgraded := service.SetWorkspaceTier(context.Background(), wsID, metering.TierPro)
	if upgraded.Tier != metering.TierPro {
		t.Errorf("expected tier pro, got %s", upgraded.Tier)
	}
	if len(upgraded.Warnings) > 0 {
		t.Errorf("expected 0 warnings after upgrading to pro tier quota")
	}
}
