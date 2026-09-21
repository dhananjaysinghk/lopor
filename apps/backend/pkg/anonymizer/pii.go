package anonymizer

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"
)

type PIICategory string

const (
	CategoryEmail      PIICategory = "EMAIL"
	CategoryPhone      PIICategory = "PHONE"
	CategorySSN        PIICategory = "SSN"
	CategoryCreditCard PIICategory = "CREDIT_CARD"
	CategoryIPAddress  PIICategory = "IP_ADDRESS"
)

type RedactedEntity struct {
	Category    PIICategory `json:"category"`
	OriginalVal string      `json:"original_value"`
	Token       string      `json:"token"`
	Position    int         `json:"position"`
}

type AnonymizeRequest struct {
	Text         string `json:"text"`
	MaskChar     string `json:"mask_char"` // e.g. "*" or "[TOKEN]"
	EnableMapping bool  `json:"enable_mapping"` // return reversible map
}

type AnonymizeResult struct {
	RedactedText   string                 `json:"redacted_text"`
	TotalRedactions int                   `json:"total_redactions"`
	EntitiesFound  []RedactedEntity       `json:"entities_found"`
	TokenMap       map[string]string      `json:"token_map,omitempty"` // Token -> Original Value
	RiskLevel      string                 `json:"risk_level"`          // "NONE", "LOW", "MEDIUM", "HIGH"
	ProcessedAt    string                 `json:"processed_at"`
}

type PIIRule struct {
	Category PIICategory
	Pattern  *regexp.Regexp
	Prefix   string
}

type PIIAnonymizer struct {
	rules []PIIRule
}

func NewPIIAnonymizer() *PIIAnonymizer {
	rules := []PIIRule{
		{
			Category: CategoryEmail,
			Pattern:  regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`),
			Prefix:   "EMAIL",
		},
		{
			Category: CategorySSN,
			Pattern:  regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`),
			Prefix:   "SSN",
		},
		{
			Category: CategoryPhone,
			Pattern:  regexp.MustCompile(`(?:\+?1[-.\s]?)?\(?[0-9]{3}\)?[-.\s]?[0-9]{3}[-.\s]?[0-9]{4}\b`),
			Prefix:   "PHONE",
		},
		{
			Category: CategoryCreditCard,
			Pattern:  regexp.MustCompile(`\b(?:\d{4}[-\s]?){3}\d{4}\b`),
			Prefix:   "CREDIT_CARD",
		},
		{
			Category: CategoryIPAddress,
			Pattern:  regexp.MustCompile(`\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\b`),
			Prefix:   "IP_ADDR",
		},
	}

	return &PIIAnonymizer{rules: rules}
}

// AnonymizeText replaces sensitive PII in text with privacy-preserving tokens.
func (a *PIIAnonymizer) AnonymizeText(ctx context.Context, req AnonymizeRequest) (*AnonymizeResult, error) {
	if strings.TrimSpace(req.Text) == "" {
		return nil, fmt.Errorf("text cannot be empty")
	}

	currentText := req.Text
	var entities []RedactedEntity
	tokenMap := make(map[string]string)
	tokenCounter := 1

	for _, rule := range a.rules {
		matches := rule.Pattern.FindAllStringIndex(currentText, -1)
		if len(matches) == 0 {
			continue
		}

		// Replace from end to start to maintain indices
		for i := len(matches) - 1; i >= 0; i-- {
			start := matches[i][0]
			end := matches[i][1]
			origVal := currentText[start:end]

			token := fmt.Sprintf("[%s_%d]", rule.Prefix, tokenCounter)
			tokenCounter++

			entities = append(entities, RedactedEntity{
				Category:    rule.Category,
				OriginalVal: origVal,
				Token:       token,
				Position:    start,
			})

			if req.EnableMapping {
				tokenMap[token] = origVal
			}

			currentText = currentText[:start] + token + currentText[end:]
		}
	}

	total := len(entities)
	riskLevel := "NONE"
	if total > 5 {
		riskLevel = "HIGH"
	} else if total > 2 {
		riskLevel = "MEDIUM"
	} else if total > 0 {
		riskLevel = "LOW"
	}

	return &AnonymizeResult{
		RedactedText:    currentText,
		TotalRedactions: total,
		EntitiesFound:   entities,
		TokenMap:        tokenMap,
		RiskLevel:       riskLevel,
		ProcessedAt:     time.Now().Format(time.RFC3339),
	}, nil
}
