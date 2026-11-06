package intentengine

import (
	"math"
	"regexp"
	"strings"
)

var (
	reCodeTokens = regexp.MustCompile(`(?m)(?:func\s+|def\s+|class\s+|SELECT\s+|WHERE\s+|import\s+|export\s+|const\s+|var\s+|let\s+|public\s+|private\s+|return\s+|=>|[{}\[\];<>])`)
	reMathLogic  = regexp.MustCompile(`(?i)(?:prove\b|theorem\b|formula\b|derive\b|calculate\b|algorithm\b|complexity\b|optimize\b|differential\b|integral\b|equation\b|step-by-step\b|tradeoff\b)`)
	reSecurity   = regexp.MustCompile(`(?i)(?:vulnerability\b|cve-\d+|exploit\b|sql injection\b|xss\b|csrf\b|firewall\b|sanitize\b|secret leak\b|penetration\b|zero-day\b)`)
	reDataSQL    = regexp.MustCompile(`(?i)(?:select\b|insert\b|update\b|delete\b|join\b|group by\b|database\b|dataset\b|dataframe\b|pandas\b|sql\b|postgres\b|csv\b|parquet\b)`)
	reCreative   = regexp.MustCompile(`(?i)(?:poem\b|story\b|blog post\b|creative\b|metaphor\b|screenplay\b|marketing copy\b|essay\b|rewrite\b)`)
)

// AnalyzeComplexity evaluates prompt hardness across code, reasoning, length, and depth.
func AnalyzeComplexity(prompt string) (ComplexityFactors, IntentDomain) {
	trimmed := strings.TrimSpace(prompt)
	words := strings.Fields(trimmed)
	wordCount := len(words)

	// 1. Length Factor
	lengthScore := math.Min(1.0, float64(wordCount)/400.0)

	// 2. Code Density Factor
	codeMatches := reCodeTokens.FindAllString(trimmed, -1)
	codeDensity := math.Min(1.0, float64(len(codeMatches))/float64(math.Max(10, float64(wordCount)*0.15)))

	// 3. Reasoning Depth Factor
	reasoningMatches := reMathLogic.FindAllString(trimmed, -1)
	reasoningDepth := math.Min(1.0, float64(len(reasoningMatches))/3.0)

	// 4. Determine Primary Domain
	var domain IntentDomain
	if reSecurity.MatchString(trimmed) {
		domain = DomainSecurityAudit
	} else if codeDensity >= 0.40 {
		domain = DomainCodeEngineering
	} else if reDataSQL.MatchString(trimmed) {
		domain = DomainDataAnalytics
	} else if reasoningDepth >= 0.50 {
		domain = DomainComplexReasoning
	} else if reCreative.MatchString(trimmed) {
		domain = DomainCreativeWriting
	} else {
		domain = DomainFastFactual
	}

	// 5. Composite Complexity Score (0 - 100%)
	rawComplexity := (lengthScore*25.0) + (codeDensity*35.0) + (reasoningDepth*40.0)
	if domain == DomainComplexReasoning || domain == DomainSecurityAudit {
		rawComplexity += 15.0
	}
	if rawComplexity > 100.0 {
		rawComplexity = 100.0
	}
	rawComplexity = math.Round(rawComplexity*10) / 10

	factors := ComplexityFactors{
		TokenLengthScore:    math.Round(lengthScore*100) / 100,
		ReasoningDepthScore: math.Round(reasoningDepth*100) / 100,
		CodeDensityScore:    math.Round(codeDensity*100) / 100,
		ContextDependency:   math.Round(lengthScore*0.5*100) / 100,
		OverallComplexity:   rawComplexity,
	}

	return factors, domain
}

// SelectTier determines the optimal target model tier from complexity and routing preferences.
func SelectTier(factors ComplexityFactors, domain IntentDomain, optimizeFor string, forceTier TargetTier) TargetTier {
	if forceTier != "" {
		return forceTier
	}

	switch strings.ToUpper(optimizeFor) {
	case "COST":
		if factors.OverallComplexity >= 75.0 {
			return TierStandard
		}
		return TierUltraFast

	case "LATENCY":
		if factors.OverallComplexity >= 85.0 {
			return TierStandard
		}
		return TierUltraFast

	case "QUALITY":
		if factors.OverallComplexity >= 50.0 || domain == DomainComplexReasoning || domain == DomainSecurityAudit {
			return TierReasoning
		}
		return TierStandard

	case "BALANCED":
		fallthrough
	default:
		if factors.OverallComplexity >= 70.0 || domain == DomainComplexReasoning {
			return TierReasoning
		}
		if factors.OverallComplexity <= 30.0 && domain == DomainFastFactual {
			return TierUltraFast
		}
		return TierStandard
	}
}

