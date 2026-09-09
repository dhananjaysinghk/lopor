package metering

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Tier string

const (
	TierFree       Tier = "free"
	TierPro        Tier = "pro"
	TierEnterprise Tier = "enterprise"
)

type Quotas struct {
	MaxTokensPerMonth   int64 `json:"max_tokens_per_month"`
	MaxVectorDocs       int64 `json:"max_vector_docs"`
	MaxSandboxExecs     int64 `json:"max_sandbox_execs"`
	MaxAudioMinutes     int64 `json:"max_audio_minutes"`
}

type UsageStats struct {
	TokensUsed     int64   `json:"tokens_used"`
	VectorDocsUsed int64   `json:"vector_docs_used"`
	SandboxExecs   int64   `json:"sandbox_execs"`
	AudioMinutes   int64   `json:"audio_minutes"`
	TokenPercent   float64 `json:"token_percent"`
	DocsPercent    float64 `json:"docs_percent"`
}

type UsageMeter struct {
	WorkspaceID uuid.UUID   `json:"workspace_id"`
	Tier        Tier        `json:"tier"`
	Quotas      Quotas      `json:"quotas"`
	Usage       UsageStats  `json:"usage"`
	IsExceeded  bool        `json:"is_exceeded"`
	Warnings    []string    `json:"warnings"`
	LastReset   string      `json:"last_reset"`
}

type UsageRecordReq struct {
	TokenType    string `json:"token_type"` // "ai_tokens", "vector_doc", "sandbox_exec", "audio_min"
	Quantity     int64  `json:"quantity"`
}

type MeterService struct {
	mu     sync.RWMutex
	meters map[uuid.UUID]*UsageMeter
}

func NewMeterService() *MeterService {
	return &MeterService{
		meters: make(map[uuid.UUID]*UsageMeter),
	}
}

func GetTierQuotas(tier Tier) Quotas {
	switch tier {
	case TierPro:
		return Quotas{
			MaxTokensPerMonth: 5_000_000,
			MaxVectorDocs:     50_000,
			MaxSandboxExecs:   5_000,
			MaxAudioMinutes:   1_000,
		}
	case TierEnterprise:
		return Quotas{
			MaxTokensPerMonth: 100_000_000,
			MaxVectorDocs:     1_000_000,
			MaxSandboxExecs:   100_000,
			MaxAudioMinutes:   50_000,
		}
	default: // TierFree
		return Quotas{
			MaxTokensPerMonth: 100_000,
			MaxVectorDocs:     500,
			MaxSandboxExecs:   50,
			MaxAudioMinutes:   15,
		}
	}
}

func (ms *MeterService) GetWorkspaceUsage(ctx context.Context, workspaceID uuid.UUID) *UsageMeter {
	ms.mu.RLock()
	m, ok := ms.meters[workspaceID]
	ms.mu.RUnlock()

	if !ok {
		ms.mu.Lock()
		defer ms.mu.Unlock()
		// Double check
		if existing, found := ms.meters[workspaceID]; found {
			return existing
		}
		m = &UsageMeter{
			WorkspaceID: workspaceID,
			Tier:        TierFree,
			Quotas:      GetTierQuotas(TierFree),
			Usage:       UsageStats{},
			LastReset:   time.Now().Format("2006-01-01"),
		}
		ms.meters[workspaceID] = m
	}

	ms.recalculateMeter(m)
	return m
}

func (ms *MeterService) RecordUsage(ctx context.Context, workspaceID uuid.UUID, req UsageRecordReq) (*UsageMeter, error) {
	ms.mu.Lock()
	m, ok := ms.meters[workspaceID]
	if !ok {
		m = &UsageMeter{
			WorkspaceID: workspaceID,
			Tier:        TierFree,
			Quotas:      GetTierQuotas(TierFree),
			Usage:       UsageStats{},
			LastReset:   time.Now().Format("2006-01-01"),
		}
		ms.meters[workspaceID] = m
	}

	qty := req.Quantity
	if qty <= 0 {
		qty = 1
	}

	switch req.TokenType {
	case "ai_tokens":
		m.Usage.TokensUsed += qty
	case "vector_doc":
		m.Usage.VectorDocsUsed += qty
	case "sandbox_exec":
		m.Usage.SandboxExecs += qty
	case "audio_min":
		m.Usage.AudioMinutes += qty
	default:
		m.Usage.TokensUsed += qty
	}

	ms.recalculateMeter(m)
	ms.mu.Unlock()

	return m, nil
}

func (ms *MeterService) SetWorkspaceTier(ctx context.Context, workspaceID uuid.UUID, tier Tier) *UsageMeter {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	m, ok := ms.meters[workspaceID]
	if !ok {
		m = &UsageMeter{
			WorkspaceID: workspaceID,
			Usage:       UsageStats{},
			LastReset:   time.Now().Format("2006-01-01"),
		}
		ms.meters[workspaceID] = m
	}

	m.Tier = tier
	m.Quotas = GetTierQuotas(tier)
	ms.recalculateMeter(m)
	return m
}

func (ms *MeterService) recalculateMeter(m *UsageMeter) {
	var warnings []string
	isExceeded := false

	if m.Quotas.MaxTokensPerMonth > 0 {
		m.Usage.TokenPercent = float64(m.Usage.TokensUsed) / float64(m.Quotas.MaxTokensPerMonth) * 100
		if m.Usage.TokenPercent >= 100 {
			isExceeded = true
			warnings = append(warnings, "Monthly AI Token quota exceeded")
		} else if m.Usage.TokenPercent >= 80 {
			warnings = append(warnings, fmt.Sprintf("AI Token consumption at %.1f%% of quota", m.Usage.TokenPercent))
		}
	}

	if m.Quotas.MaxVectorDocs > 0 {
		m.Usage.DocsPercent = float64(m.Usage.VectorDocsUsed) / float64(m.Quotas.MaxVectorDocs) * 100
		if m.Usage.DocsPercent >= 100 {
			isExceeded = true
			warnings = append(warnings, "RAG Vector document storage quota exceeded")
		}
	}

	m.IsExceeded = isExceeded
	m.Warnings = warnings
}
