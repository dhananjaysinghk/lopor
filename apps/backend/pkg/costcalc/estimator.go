package costcalc

import (
	"context"
	"strings"
	"time"
)

type ModelPrice struct {
	ModelID        string  `json:"model_id"`
	Provider       string  `json:"provider"`
	PromptPer1M    float64 `json:"prompt_per_1m"`    // USD per 1M input tokens
	CompletionPer1M float64 `json:"completion_per_1m"`// USD per 1M output tokens
	CacheReadPer1M  float64 `json:"cache_read_per_1m"` // USD per 1M cached input tokens
}

type CostEstimateRequest struct {
	ModelID        string `json:"model_id"`
	PromptText     string `json:"prompt_text,omitempty"`
	PromptTokens   int    `json:"prompt_tokens,omitempty"`
	ExpectedOutput int    `json:"expected_output_tokens,omitempty"`
	CachedTokens   int    `json:"cached_tokens,omitempty"`
}

type CostBreakdown struct {
	ModelID         string  `json:"model_id"`
	Provider        string  `json:"provider"`
	PromptTokens    int     `json:"prompt_tokens"`
	CompletionTokens int    `json:"completion_tokens"`
	CachedTokens    int     `json:"cached_tokens"`
	PromptCostUSD   float64 `json:"prompt_cost_usd"`
	CompletionCostUSD float64 `json:"completion_cost_usd"`
	CacheSavingsUSD float64 `json:"cache_savings_usd"`
	TotalCostUSD    float64 `json:"total_cost_usd"`
	IsHighCostQuery bool    `json:"is_high_cost_query"`
	EstimatedAt     string  `json:"estimated_at"`
}

type CostEstimator struct {
	pricingCatalog map[string]ModelPrice
}

func NewCostEstimator() *CostEstimator {
	catalog := map[string]ModelPrice{
		"gpt-4o": {
			ModelID:         "gpt-4o",
			Provider:        "OpenAI",
			PromptPer1M:     2.50,
			CompletionPer1M: 10.00,
			CacheReadPer1M:  1.25,
		},
		"gpt-4o-mini": {
			ModelID:         "gpt-4o-mini",
			Provider:        "OpenAI",
			PromptPer1M:     0.15,
			CompletionPer1M: 0.60,
			CacheReadPer1M:  0.075,
		},
		"claude-3-5-sonnet": {
			ModelID:         "claude-3-5-sonnet",
			Provider:        "Anthropic",
			PromptPer1M:     3.00,
			CompletionPer1M: 15.00,
			CacheReadPer1M:  0.30,
		},
		"gemini-1-5-pro": {
			ModelID:         "gemini-1-5-pro",
			Provider:        "Google",
			PromptPer1M:     1.25,
			CompletionPer1M: 5.00,
			CacheReadPer1M:  0.3125,
		},
		"gemini-1-5-flash": {
			ModelID:         "gemini-1-5-flash",
			Provider:        "Google",
			PromptPer1M:     0.075,
			CompletionPer1M: 0.30,
			CacheReadPer1M:  0.01875,
		},
		"llama-3-70b": {
			ModelID:         "llama-3-70b",
			Provider:        "Groq / Meta",
			PromptPer1M:     0.59,
			CompletionPer1M: 0.79,
			CacheReadPer1M:  0.20,
		},
	}

	return &CostEstimator{pricingCatalog: catalog}
}

// GetCatalog returns all available model pricing matrices.
func (ce *CostEstimator) GetCatalog() []ModelPrice {
	var list []ModelPrice
	for _, p := range ce.pricingCatalog {
		list = append(list, p)
	}
	return list
}

// EstimateCost calculates detailed monetary cost for input prompt and estimated output.
func (ce *CostEstimator) EstimateCost(ctx context.Context, req CostEstimateRequest) (*CostBreakdown, error) {
	modelKey := req.ModelID
	if modelKey == "" {
		modelKey = "gpt-4o-mini"
	}

	price, ok := ce.pricingCatalog[modelKey]
	if !ok {
		// Default to gpt-4o-mini rates if model not explicitly cataloged
		price = ce.pricingCatalog["gpt-4o-mini"]
		price.ModelID = modelKey
		price.Provider = "Custom / Compatible"
	}

	promptTokens := req.PromptTokens
	if promptTokens == 0 && req.PromptText != "" {
		promptTokens = len(strings.Fields(req.PromptText)) * 4 / 3
	}

	completionTokens := req.ExpectedOutput
	if completionTokens <= 0 {
		completionTokens = 800 // default expected response length
	}

	cachedTokens := req.CachedTokens
	if cachedTokens > promptTokens {
		cachedTokens = promptTokens
	}

	nonCachedTokens := promptTokens - cachedTokens

	promptCost := (float64(nonCachedTokens) / 1_000_000.0) * price.PromptPer1M
	cachedCost := (float64(cachedTokens) / 1_000_000.0) * price.CacheReadPer1M
	completionCost := (float64(completionTokens) / 1_000_000.0) * price.CompletionPer1M

	totalCost := promptCost + cachedCost + completionCost
	fullPromptWithoutCacheCost := (float64(promptTokens) / 1_000_000.0) * price.PromptPer1M
	savings := fullPromptWithoutCacheCost - (promptCost + cachedCost)
	if savings < 0 {
		savings = 0
	}

	isHighCost := totalCost >= 0.25 // flag queries costing >= $0.25

	return &CostBreakdown{
		ModelID:           price.ModelID,
		Provider:          price.Provider,
		PromptTokens:      promptTokens,
		CompletionTokens:  completionTokens,
		CachedTokens:      cachedTokens,
		PromptCostUSD:     promptCost + cachedCost,
		CompletionCostUSD: completionCost,
		CacheSavingsUSD:   savings,
		TotalCostUSD:      totalCost,
		IsHighCostQuery:   isHighCost,
		EstimatedAt:       time.Now().Format(time.RFC3339),
	}, nil
}
