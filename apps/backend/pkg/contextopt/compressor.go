package contextopt

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"
)

type CompressionStrategy string

const (
	StrategyAggressive CompressionStrategy = "aggressive"
	StrategyModerate   CompressionStrategy = "moderate"
	StrategyCodeSafe   CompressionStrategy = "code_safe"
)

type CompressionRequest struct {
	Text           string              `json:"text"`
	MaxTargetTokens int                `json:"max_target_tokens"`
	Strategy       CompressionStrategy `json:"strategy"`
	PreserveCode   bool                `json:"preserve_code"`
}

type CompressionResult struct {
	OriginalTokens   int                 `json:"original_tokens"`
	CompressedTokens int                 `json:"compressed_tokens"`
	SavedTokens      int                 `json:"saved_tokens"`
	SavingsPercent   float64             `json:"savings_percent"`
	CompressedText   string              `json:"compressed_text"`
	Strategy         CompressionStrategy `json:"strategy"`
	Timestamp        string              `json:"timestamp"`
}

type ContextCompressor struct {
	whitespaceRegex *regexp.Regexp
	codeBlockRegex  *regexp.Regexp
}

func NewContextCompressor() *ContextCompressor {
	return &ContextCompressor{
		whitespaceRegex: regexp.MustCompile(`\n{3,}`),
		codeBlockRegex:  regexp.MustCompile("(?s)```.*?```"),
	}
}

// CompressContext reduces redundant tokens while preserving essential facts and code blocks.
func (cc *ContextCompressor) CompressContext(ctx context.Context, req CompressionRequest) (*CompressionResult, error) {
	if strings.TrimSpace(req.Text) == "" {
		return nil, fmt.Errorf("input text cannot be empty")
	}

	strategy := req.Strategy
	if strategy == "" {
		strategy = StrategyModerate
	}

	origTokens := cc.estimateTokens(req.Text)

	var compressed string
	if req.PreserveCode || strategy == StrategyCodeSafe {
		compressed = cc.compressWithCodePreservation(req.Text, strategy)
	} else {
		compressed = cc.compressText(req.Text, strategy)
	}

	compTokens := cc.estimateTokens(compressed)
	saved := origTokens - compTokens
	if saved < 0 {
		saved = 0
		compTokens = origTokens
		compressed = req.Text
	}

	savingsPercent := 0.0
	if origTokens > 0 {
		savingsPercent = float64(saved) / float64(origTokens) * 100.0
	}

	return &CompressionResult{
		OriginalTokens:   origTokens,
		CompressedTokens: compTokens,
		SavedTokens:      saved,
		SavingsPercent:   savingsPercent,
		CompressedText:   compressed,
		Strategy:         strategy,
		Timestamp:        time.Now().Format(time.RFC3339),
	}, nil
}

func (cc *ContextCompressor) compressWithCodePreservation(text string, strategy CompressionStrategy) string {
	// Extract code blocks to prevent altering syntax
	codeBlocks := cc.codeBlockRegex.FindAllString(text, -1)
	placeholder := "___CODE_BLOCK_PLACEHOLDER___"
	textWithoutCode := cc.codeBlockRegex.ReplaceAllString(text, placeholder)

	compressedProse := cc.compressText(textWithoutCode, strategy)

	for _, cb := range codeBlocks {
		compressedProse = strings.Replace(compressedProse, placeholder, cb, 1)
	}

	return compressedProse
}

func (cc *ContextCompressor) compressText(text string, strategy CompressionStrategy) string {
	lines := strings.Split(text, "\n")
	var prunedLines []string

	fillerPhrases := []string{
		"in order to", "as a matter of fact", "for the purpose of",
		"it is important to note that", "please be advised that",
		"at the end of the day", "needless to say",
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		cleanLine := trimmed
		for _, filler := range fillerPhrases {
			cleanLine = strings.ReplaceAll(cleanLine, filler, "")
		}

		if strategy == StrategyAggressive {
			// Remove common introductory fillers
			cleanLine = strings.TrimPrefix(cleanLine, "Basically, ")
			cleanLine = strings.TrimPrefix(cleanLine, "Essentially, ")
		}

		prunedLines = append(prunedLines, cleanLine)
	}

	res := strings.Join(prunedLines, "\n")
	res = cc.whitespaceRegex.ReplaceAllString(res, "\n\n")
	return res
}

func (cc *ContextCompressor) estimateTokens(text string) int {
	words := len(strings.Fields(text))
	return words * 4 / 3
}
