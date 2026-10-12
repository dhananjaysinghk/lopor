package guardrails

import (
	"context"
	"math"
	"regexp"
	"sync"
	"time"
)

// compiledRule caches the pre-compiled regex for high-throughput evaluation.
type compiledRule struct {
	rule    GuardrailRule
	matcher *regexp.Regexp
}

// GuardrailEngine enforces bi-directional security policies across prompt inputs and model completions.
type GuardrailEngine struct {
	mu            sync.RWMutex
	builtinRules  []compiledRule
	profiles      map[string]GuardrailProfile // workspaceID -> profile
}

// NewGuardrailEngine initializes the firewall engine with enterprise-grade baseline rules.
func NewGuardrailEngine() *GuardrailEngine {
	engine := &GuardrailEngine{
		profiles: make(map[string]GuardrailProfile),
	}
	engine.initBuiltinRules()
	return engine
}

func (e *GuardrailEngine) initBuiltinRules() {
	rawRules := []GuardrailRule{
		// Inbound: Prompt Injection & Adversarial Jailbreaks
		{
			ID:          "rule-inj-01",
			Name:        "Instruction Override Defense",
			Category:    CategoryPromptInjection,
			Direction:   DirectionInboundPrompt,
			Regex:       `(?i)\b(ignore (all )?previous instructions|disregard (all )?prior rules|bypass (safety|security) policy|forget (all )?prior guidelines|override system prompt)\b`,
			Severity:    SeverityCritical,
			Action:      ActionBlock,
			Description: "Detects explicit attempts to override platform instructions and safety boundaries",
			Enabled:     true,
		},
		{
			ID:          "rule-jail-01",
			Name:        "Adversarial Role Reversal Jailbreak",
			Category:    CategoryJailbreak,
			Direction:   DirectionInboundPrompt,
			Regex:       `(?i)\b(you are now (unfiltered|unrestricted|DAN|Developer Mode|evil mode)|do anything now mode|ignore ethical guidelines)\b`,
			Severity:    SeverityCritical,
			Action:      ActionBlock,
			Description: "Detects well-known jailbreak role-play maneuvers designed to subvert moderation",
			Enabled:     true,
		},
		{
			ID:          "rule-leak-01",
			Name:        "System Prompt Extraction Barrier",
			Category:    CategorySystemPromptLeak,
			Direction:   DirectionInboundPrompt,
			Regex:       `(?i)\b(repeat the (text|words|system prompt) above|output your (initial|developer) prompt verbatim|print your hidden prompt)\b`,
			Severity:    SeverityHigh,
			Action:      ActionBlock,
			Description: "Blocks extraction of confidential system instructions or internal persona context",
			Enabled:     true,
		},

		// Bi-Directional: Secret & Credential Exposure
		{
			ID:          "rule-sec-aws",
			Name:        "AWS Access Key Detection",
			Category:    CategorySecretExposure,
			Direction:   DirectionInboundPrompt,
			Regex:       `\bAKIA[0-9A-Z]{16}\b`,
			Severity:    SeverityHigh,
			Action:      ActionSanitize,
			Description: "Identifies exposed Amazon Web Services IAM access keys",
			Enabled:     true,
		},
		{
			ID:          "rule-sec-gh",
			Name:        "GitHub Personal Access Token Detection",
			Category:    CategorySecretExposure,
			Direction:   DirectionInboundPrompt,
			Regex:       `\bghp_[a-zA-Z0-9]{36}\b`,
			Severity:    SeverityHigh,
			Action:      ActionSanitize,
			Description: "Identifies leaked GitHub personal access tokens",
			Enabled:     true,
		},
		{
			ID:          "rule-sec-jwt",
			Name:        "JWT Bearer Token Detection",
			Category:    CategorySecretExposure,
			Direction:   DirectionInboundPrompt,
			Regex:       `\beyJ[a-zA-Z0-9_-]{10,}\.[a-zA-Z0-9_-]{10,}\.[a-zA-Z0-9_-]{10,}\b`,
			Severity:    SeverityHigh,
			Action:      ActionSanitize,
			Description: "Identifies raw JSON Web Tokens in prompt payloads",
			Enabled:     true,
		},
		{
			ID:          "rule-sec-pem",
			Name:        "Private RSA/EC Key Header",
			Category:    CategorySecretExposure,
			Direction:   DirectionInboundPrompt,
			Regex:       `-----BEGIN [A-Z ]*PRIVATE KEY-----`,
			Severity:    SeverityCritical,
			Action:      ActionBlock,
			Description: "Detects transmission of private cryptographic keys",
			Enabled:     true,
		},

		// Outbound: Destructive Execution & Database Erasure
		{
			ID:          "rule-dest-01",
			Name:        "Destructive Shell Command Guard",
			Category:    CategoryDestructiveCommand,
			Direction:   DirectionOutboundResponse,
			Regex:       `(?i)\b(rm -rf \/|mkfs\.[a-z0-9]+|:\(\)\{\s*:\|:&\s*\};:|chmod -R 777 \/)\b`,
			Severity:    SeverityCritical,
			Action:      ActionBlock,
			Description: "Prevents synthesis of commands known to wipe host storage or cause denial of service",
			Enabled:     true,
		},
		{
			ID:          "rule-dest-db",
			Name:        "Destructive Database DDL Defense",
			Category:    CategoryDestructiveCommand,
			Direction:   DirectionOutboundResponse,
			Regex:       `(?i)\b(DROP DATABASE|DROP TABLE [a-zA-Z0-9_]+ CASCADE|TRUNCATE TABLE [a-zA-Z0-9_]+)\b`,
			Severity:    SeverityHigh,
			Action:      ActionBlock,
			Description: "Protects enterprise databases against unintentional or malicious cascading drops",
			Enabled:     true,
		},
		{
			ID:          "rule-sec-out-aws",
			Name:        "Outbound AWS Key Leakage Prevention",
			Category:    CategorySecretExposure,
			Direction:   DirectionOutboundResponse,
			Regex:       `\bAKIA[0-9A-Z]{16}\b`,
			Severity:    SeverityHigh,
			Action:      ActionSanitize,
			Description: "Scrubs generated AWS access keys before leaving server boundary",
			Enabled:     true,
		},
	}

	for _, r := range rawRules {
		compiled, err := regexp.Compile(r.Regex)
		if err == nil {
			e.builtinRules = append(e.builtinRules, compiledRule{
				rule:    r,
				matcher: compiled,
			})
		}
	}
}

// Inspect evaluates content against configured policies, returning allowed status and sanitizations.
func (e *GuardrailEngine) Inspect(ctx context.Context, req InspectionRequest) (*InspectionResult, error) {
	start := time.Now()

	profile := e.getProfile(req.WorkspaceID)
	sanitized := req.Content
	var violations []Violation
	maxSeverity := SeverityNone

	e.mu.RLock()
	rules := append([]compiledRule{}, e.builtinRules...)
	for _, cr := range profile.CustomRules {
		if cr.Enabled {
			re, err := regexp.Compile(cr.Regex)
			if err == nil {
				rules = append(rules, compiledRule{rule: cr, matcher: re})
			}
		}
	}
	e.mu.RUnlock()

	for _, cr := range rules {
		if !cr.rule.Enabled {
			continue
		}
		// Match direction (or apply if rule matches both directions)
		if cr.rule.Direction != req.Direction {
			continue
		}

		matches := cr.matcher.FindAllString(sanitized, -1)
		if len(matches) > 0 {
			for _, m := range matches {
				v := Violation{
					Category:    cr.rule.Category,
					Severity:    cr.rule.Severity,
					RuleName:    cr.rule.Name,
					Description: cr.rule.Description,
					MatchedText: m,
					Remediation: e.getRemediation(cr.rule.Category),
				}
				violations = append(violations, v)
				if isHigherSeverity(cr.rule.Severity, maxSeverity) {
					maxSeverity = cr.rule.Severity
				}
			}

			// Apply secret masking if configured
			if cr.rule.Category == CategorySecretExposure && (cr.rule.Action == ActionSanitize || profile.MaskSecrets) {
				sanitized = cr.matcher.ReplaceAllString(sanitized, "[REDACTED_SECRET]")
			}
		}
	}

	riskScore := calculateRiskScore(violations)
	allowed := true
	actionTaken := ActionPass

	// Check if block threshold is crossed
	if severityRank(maxSeverity) >= severityRank(profile.BlockThreshold) && maxSeverity != SeverityNone {
		allowed = false
		actionTaken = ActionBlock
	} else if sanitized != req.Content {
		allowed = true
		actionTaken = ActionSanitize
	} else if len(violations) > 0 {
		allowed = true
		actionTaken = ActionFlagAndLog
	}

	return &InspectionResult{
		Allowed:          allowed,
		ActionTaken:      actionTaken,
		SanitizedContent: sanitized,
		OverallRiskScore: riskScore,
		Violations:       violations,
		EvaluatedAt:      time.Now().UTC(),
		LatencyMs:        time.Since(start).Milliseconds(),
	}, nil
}

// SetWorkspaceProfile saves or updates custom guardrail policies for a workspace.
func (e *GuardrailEngine) SetWorkspaceProfile(profile GuardrailProfile) {
	e.mu.Lock()
	defer e.mu.Unlock()
	profile.UpdatedAt = time.Now().UTC()
	e.profiles[profile.WorkspaceID] = profile
}

// GetWorkspaceProfile returns the active policy profile for a workspace.
func (e *GuardrailEngine) GetWorkspaceProfile(workspaceID string) GuardrailProfile {
	return e.getProfile(workspaceID)
}

func (e *GuardrailEngine) getProfile(workspaceID string) GuardrailProfile {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if p, ok := e.profiles[workspaceID]; ok {
		return p
	}

	// Default enterprise profile
	return GuardrailProfile{
		WorkspaceID:    workspaceID,
		Name:           "Default Enterprise Guardrail Profile",
		DefaultAction:  ActionPass,
		BlockThreshold: SeverityHigh, // Block HIGH and CRITICAL violations by default
		MaskSecrets:    true,
		UpdatedAt:      time.Now().UTC(),
	}
}

func (e *GuardrailEngine) getRemediation(cat GuardrailCategory) string {
	switch cat {
	case CategoryPromptInjection:
		return "Remove adversarial prefix instructions and rephrase query neutrally."
	case CategoryJailbreak:
		return "Refrain from roleplay directives targeting AI policy bypasses."
	case CategorySystemPromptLeak:
		return "System prompts are proprietary; query platform APIs directly."
	case CategorySecretExposure:
		return "Store secrets in environment variables or KMS; never paste raw credentials."
	case CategoryDestructiveCommand:
		return "Use sandbox isolation or parameterize destructive system operations."
	default:
		return "Ensure query aligns with enterprise safety guidelines."
	}
}

func severityRank(s ViolationSeverity) int {
	switch s {
	case SeverityCritical:
		return 4
	case SeverityHigh:
		return 3
	case SeverityMedium:
		return 2
	case SeverityLow:
		return 1
	default:
		return 0
	}
}

func isHigherSeverity(a, b ViolationSeverity) bool {
	return severityRank(a) > severityRank(b)
}

func calculateRiskScore(violations []Violation) float64 {
	if len(violations) == 0 {
		return 0.0
	}

	score := 0.0
	for _, v := range violations {
		switch v.Severity {
		case SeverityCritical:
			score += 0.50
		case SeverityHigh:
			score += 0.30
		case SeverityMedium:
			score += 0.15
		case SeverityLow:
			score += 0.05
		}
	}

	if score > 1.0 {
		score = 1.0
	}
	return math.Round(score*100) / 100
}
