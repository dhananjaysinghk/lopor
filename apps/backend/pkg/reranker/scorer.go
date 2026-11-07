package reranker

import (
	"math"
	"strings"
	"unicode"
)

var commonStopwords = map[string]struct{}{
	"a": {}, "an": {}, "the": {}, "and": {}, "or": {}, "in": {}, "on": {},
	"at": {}, "to": {}, "for": {}, "of": {}, "with": {}, "by": {}, "is": {},
	"are": {}, "was": {}, "were": {}, "be": {}, "this": {}, "that": {},
	"it": {}, "as": {}, "from": {}, "what": {}, "how": {}, "which": {},
	"who": {}, "why": {}, "when": {}, "where": {},
}

// ComputeRelevanceScore calculates an advanced cross-attention score between a query and a document chunk.
func ComputeRelevanceScore(query string, chunk DocumentChunk, strategy RerankStrategy) (float64, RelevanceGrade) {
	qTokens := tokenize(query)
	cTokens := tokenize(chunk.Content)
	titleTokens := tokenize(chunk.Title)

	if len(qTokens) == 0 || len(cTokens) == 0 {
		return 0.0, GradeIrrelevant
	}

	// 1. Lexical Term Overlap & Frequency
	termCounts := make(map[string]int)
	for _, t := range cTokens {
		termCounts[t]++
	}

	uniqueQueryTerms := make(map[string]struct{})
	for _, t := range qTokens {
		if _, isStop := commonStopwords[t]; !isStop {
			uniqueQueryTerms[t] = struct{}{}
		}
	}
	if len(uniqueQueryTerms) == 0 {
		for _, t := range qTokens {
			uniqueQueryTerms[t] = struct{}{}
		}
	}

	matchedTerms := 0
	tfScore := 0.0
	for term := range uniqueQueryTerms {
		if count, found := termCounts[term]; found {
			matchedTerms++
			// Sub-linear term frequency scaling
			tfScore += 1.0 + math.Log(float64(count))
		}
	}

	termCoverage := float64(matchedTerms) / float64(len(uniqueQueryTerms))

	// 2. Exact Phrase Match
	normQuery := strings.ToLower(strings.TrimSpace(query))
	normContent := strings.ToLower(chunk.Content)
	phraseBonus := 0.0
	if len(normQuery) > 5 && strings.Contains(normContent, normQuery) {
		phraseBonus = 0.25
	}

	// 3. Bi-Gram Sequence Matching
	bigramBonus := computeBigramOverlap(qTokens, cTokens) * 0.15

	// 4. Structural & Title Match Bonus
	titleBonus := 0.0
	for term := range uniqueQueryTerms {
		for _, tt := range titleTokens {
			if term == tt {
				titleBonus += 0.05
				break
			}
		}
	}
	if titleBonus > 0.15 {
		titleBonus = 0.15
	}

	// 5. Length penalty / density normalization
	docLen := float64(len(cTokens))
	avgDocLen := 120.0
	lengthFactor := (docLen + 50.0) / (avgDocLen + 50.0)
	normalizedTf := (tfScore / lengthFactor) * 0.20
	if normalizedTf > 0.35 {
		normalizedTf = 0.35
	}

	// Composite Cross-Encoder Approximation Score
	rawScore := (termCoverage * 0.40) + normalizedTf + phraseBonus + bigramBonus + titleBonus

	// Combine with initial score if available according to strategy
	var finalScore float64
	switch strategy {
	case StrategyHybridLexical:
		if chunk.InitialScore > 0 {
			finalScore = (rawScore * 0.60) + (chunk.InitialScore * 0.40)
		} else {
			finalScore = rawScore
		}
	case StrategyReciprocalRank:
		if chunk.InitialScore > 0 {
			finalScore = (rawScore * 0.70) + (chunk.InitialScore * 0.30)
		} else {
			finalScore = rawScore
		}
	case StrategyCrossEncoder:
		fallthrough
	default:
		if chunk.InitialScore > 0 {
			finalScore = (rawScore * 0.75) + (chunk.InitialScore * 0.25)
		} else {
			finalScore = rawScore
		}
	}

	if finalScore > 1.0 {
		finalScore = 1.0
	}
	if finalScore < 0.0 {
		finalScore = 0.0
	}

	// Assign Grade
	var grade RelevanceGrade
	switch {
	case finalScore >= 0.70:
		grade = GradeHigh
	case finalScore >= 0.45:
		grade = GradeMedium
	case finalScore >= 0.25:
		grade = GradeLow
	default:
		grade = GradeIrrelevant
	}

	return math.Round(finalScore*1000) / 1000, grade
}

func tokenize(text string) []string {
	var tokens []string
	var current strings.Builder

	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			current.WriteRune(unicode.ToLower(r))
		} else if current.Len() > 0 {
			tokens = append(tokens, current.String())
			current.Reset()
		}
	}
	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}
	return tokens
}

func computeBigramOverlap(qTokens, cTokens []string) float64 {
	if len(qTokens) < 2 || len(cTokens) < 2 {
		return 0.0
	}

	qBigrams := make(map[string]struct{})
	for i := 0; i < len(qTokens)-1; i++ {
		bg := qTokens[i] + "_" + qTokens[i+1]
		qBigrams[bg] = struct{}{}
	}

	cBigrams := make(map[string]struct{})
	for i := 0; i < len(cTokens)-1; i++ {
		bg := cTokens[i] + "_" + cTokens[i+1]
		cBigrams[bg] = struct{}{}
	}

	matches := 0
	for bg := range qBigrams {
		if _, found := cBigrams[bg]; found {
			matches++
		}
	}

	return float64(matches) / float64(len(qBigrams))
}

