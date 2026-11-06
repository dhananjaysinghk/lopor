package intentengine

import (
	"time"
)

// IntentDomain categorizes the operational sphere of an incoming user prompt.
type IntentDomain string

const (
	DomainCodeEngineering   IntentDomain = "CODE_ENGINEERING"   // Syntax, debugging, refactoring, algorithms
	DomainDataAnalytics     IntentDomain = "DATA_ANALYTICS"     // SQL queries, statistical analysis, datasets
	DomainCreativeWriting   IntentDomain = "CREATIVE_WRITING"   // Marketing copy, essays, brain-storming
	DomainComplexReasoning  IntentDomain = "COMPLEX_REASONING"  // Multi-step math, formal logic, architecture design
	DomainFastFactual       IntentDomain = "FAST_FACTUAL"       // General Q&A, definitions, summaries
	DomainSecurityAudit     IntentDomain = "SECURITY_AUDIT"     // Vulnerability scanning, CVE triage, redteaming
)

// TargetTier identifies the hardware and economic profile of the target model.
type TargetTier string

const (
	TierUltraFast TargetTier = "ULTRA_FAST" // Sub-100ms lightweight SLMs (e.g. Flash-Lite, Haiku)
	TierStandard  TargetTier = "STANDARD"   // Balanced flagship models (e.g. GPT-4o, Sonnet 3.5)
	TierReasoning TargetTier = "REASONING"  // Extended CoT thinking models (e.g. o1, o3-mini, DeepSeek-R1)
)

// ModelCandidate defines an individual candidate LLM in the routing pool.
type ModelCandidate struct {
	ModelID            string       `json:"model_id"`
	Provider           string       `json:"provider"` // "openai", "anthropic", "google", "deepseek"
	Tier               TargetTier   `json:"tier"`
	CostPer1MInTokens  float64      `json:"cost_per_1m_in_tokens"`
	CostPer1MOutTokens float64      `json:"cost_per_1m_out_tokens"`
	TypicalLatencyMs   int64        `json:"typical_latency_ms"`
	SupportedDomains   []IntentDomain `json:"supported_domains"`
	MaxContextWindow   int          `json:"max_context_window"`
}

// ComplexityFactors evaluates multiple dimensions of prompt hardness.
type ComplexityFactors struct {
	TokenLengthScore     float64 `json:"token_length_score"`     // 0.0 - 1.0
	ReasoningDepthScore  float64 `json:"reasoning_depth_score"`  // 0.0 - 1.0 (math, logic keywords)
	CodeDensityScore     float64 `json:"code_density_score"`     // 0.0 - 1.0 (syntax, symbols, functions)
	ContextDependency    float64 `json:"context_dependency"`     // 0.0 - 1.0 (RAG documents, conversation history)
	OverallComplexity    float64 `json:"overall_complexity"`     // 0.0 - 100.0%
}

// RoutingDecision encapsulates the selected model, economic savings, and explanation.
type RoutingDecision struct {
	DecisionID            string            `json:"decision_id"`
	WorkspaceID           string            `json:"workspace_id"`
	PrimaryDomain         IntentDomain      `json:"primary_domain"`
	Complexity            ComplexityFactors `json:"complexity"`
	SelectedModel         ModelCandidate    `json:"selected_model"`
	FallbackModel         *ModelCandidate   `json:"fallback_model,omitempty"`
	ReasoningRequired     bool              `json:"reasoning_required"`
	EstimatedLatencyMs    int64             `json:"estimated_latency_ms"`
	EstimatedCostUSD      float64           `json:"estimated_cost_usd"`
	PotentialSavingsUSD   float64           `json:"potential_savings_usd"` // Compared to always routing to flagship
	RoutingExplanation    string            `json:"routing_explanation"`
	RoutingLatencyMicros  int64             `json:"routing_latency_micros"` // Overhead of router in microseconds
	DecidedAt             time.Time         `json:"decided_at"`
}

// RoutePromptRequest specifies the incoming prompt and optimization priorities.
type RoutePromptRequest struct {
	WorkspaceID        string       `json:"workspace_id"`
	Prompt             string       `json:"prompt"`
	OptimizeFor        string       `json:"optimize_for,omitempty"` // "COST", "LATENCY", "QUALITY", "BALANCED"
	MaxAllowedLatencyMs int64        `json:"max_allowed_latency_ms,omitempty"`
	ForceTier          TargetTier   `json:"force_tier,omitempty"`
}

// RouterTelemetry tracks workspace-wide routing efficiency and cumulative cost avoidance.
type RouterTelemetry struct {
	WorkspaceID           string                 `json:"workspace_id"`
	TotalRoutedRequests   int64                  `json:"total_routed_requests"`
	TierDistribution      map[TargetTier]int64   `json:"tier_distribution"`
	DomainDistribution    map[IntentDomain]int64 `json:"domain_distribution"`
	CumulativeSavedUSD    float64                `json:"cumulative_saved_usd"`
	AverageRouterLatencyMs float64               `json:"average_router_latency_ms"`
}

