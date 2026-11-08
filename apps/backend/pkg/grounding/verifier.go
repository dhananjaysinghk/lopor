package grounding

import (
	"fmt"
	"math"
	"regexp"
	"strings"
	"unicode"
)

var (
	negationWords = map[string]struct{}{
		"not": {}, "never": {}, "no": {}, "neither": {}, "nor": {},
		"cannot": {}, "isn't": {}, "aren't": {}, "wasn't": {}, "weren't": {},
		"won't": {}, "don't": {}, "doesn't": {}, "didn't": {},
		"fails": {}, "disallows": {}, "unsupported": {}, "prohibits": {}, "lacks": {},
	}
	stopwords = map[string]struct{}{
		"the": {}, "a": {}, "an": {}, "and": {}, "or": {}, "in": {}, "on": {},
		"at": {}, "to": {}, "for": {}, "of": {}, "with": {}, "by": {}, "is": {},
		"are": {}, "was": {}, "were": {}, "be": {}, "this": {}, "that": {},
		"it": {}, "as": {}, "from": {}, "which": {},
	}
	numberPattern = regexp.MustCompile(`\b\d+(?:\.\d+)?\b`)
)

// VerifyClaim evaluates an atomic claim against all available source evidence.
func VerifyClaim(claim AtomicClaim, sources []SourceEvidence, strictness float64) AtomicClaim {
	if strictness <= 0 {
		strictness = 0.65
	}

	bestStatus := StatusUngrounded
	bestConfidence := 0.0
	var bestSource SourceEvidence
	var bestSnippet string
	var bestExplanation string

	claimTokens := tokenizeText(claim.ClaimText)
	claimNegated := hasNegation(claimTokens)
	claimNumbers := extractNumbers(claim.ClaimText)

	for _, src := range sources {
		srcTokens := tokenizeText(src.Content)
		srcNegated := hasNegation(srcTokens)
		srcNumbers := extractNumbers(src.Content)

		// Term overlap
		overlap, matched := computeOverlap(claimTokens, srcTokens)

		// Number consistency check: if claim contains numbers, do they appear in the source?
		numberMismatch := false
		if len(claimNumbers) > 0 {
			for num := range claimNumbers {
				if _, ok := srcNumbers[num]; !ok {
					numberMismatch = true
					break
				}
			}
		}

		// Polarity inversion check
		polarityContradiction := (claimNegated != srcNegated) && (overlap >= 0.50)

		if polarityContradiction || (numberMismatch && overlap >= 0.60) {
			// Found evidence of contradiction
			bestStatus = StatusContradicted
			bestConfidence = math.Round(overlap*1000) / 1000
			bestSource = src
			bestSnippet = findBestSnippet(claim.ClaimText, src.Content)
			if polarityContradiction {
				bestExplanation = fmt.Sprintf("Polarity contradiction detected against source '%s': statement inverts negation state.", src.Title)
			} else {
				bestExplanation = fmt.Sprintf("Numerical mismatch detected against source '%s': numerical entities in claim do not match reference data.", src.Title)
			}
			break
		}

		if overlap > bestConfidence {
			bestConfidence = overlap
			bestSource = src
			bestSnippet = findBestSnippet(claim.ClaimText, src.Content)
		}

		if overlap >= strictness && !numberMismatch {
			bestStatus = StatusSupported
			bestExplanation = fmt.Sprintf("Corroborated by source '%s' with %d matching salient terms.", src.Title, matched)
		}
	}

	if bestStatus == StatusUngrounded && bestConfidence < strictness {
		bestExplanation = fmt.Sprintf("Insufficient evidence in source documents (highest overlap: %.1f%%, required: %.1f%%).",
			bestConfidence*100.0, strictness*100.0)
	}

	claim.Status = bestStatus
	claim.EntailmentConfidence = math.Round(bestConfidence*1000) / 1000
	claim.SourceEvidenceID = bestSource.ID
	claim.SourceTitle = bestSource.Title
	claim.EvidenceSnippet = bestSnippet
	claim.Explanation = bestExplanation

	return claim
}

func tokenizeText(text string) []string {
	var tokens []string
	var sb strings.Builder

	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			sb.WriteRune(unicode.ToLower(r))
		} else if sb.Len() > 0 {
			tokens = append(tokens, sb.String())
			sb.Reset()
		}
	}
	if sb.Len() > 0 {
		tokens = append(tokens, sb.String())
	}
	return tokens
}

func computeOverlap(claimTokens, srcTokens []string) (float64, int) {
	srcSet := make(map[string]struct{}, len(srcTokens))
	for _, t := range srcTokens {
		srcSet[t] = struct{}{}
	}

	salientClaimTerms := make(map[string]struct{})
	for _, t := range claimTokens {
		if _, isStop := stopwords[t]; !isStop {
			salientClaimTerms[t] = struct{}{}
		}
	}

	if len(salientClaimTerms) == 0 {
		for _, t := range claimTokens {
			salientClaimTerms[t] = struct{}{}
		}
	}

	matched := 0
	for term := range salientClaimTerms {
		if _, found := srcSet[term]; found {
			matched++
		}
	}

	overlap := float64(matched) / float64(len(salientClaimTerms))
	return overlap, matched
}

func hasNegation(tokens []string) bool {
	for _, t := range tokens {
		if _, isNeg := negationWords[t]; isNeg {
			return true
		}
	}
	return false
}

func extractNumbers(text string) map[string]struct{} {
	nums := make(map[string]struct{})
	matches := numberPattern.FindAllString(text, -1)
	for _, m := range matches {
		nums[m] = struct{}{}
	}
	return nums
}

func findBestSnippet(claimText, sourceContent string) string {
	claimWords := tokenizeText(claimText)
	sentences := strings.Split(sourceContent, ".")
	bestSent := ""
	bestMatch := 0

	claimWordSet := make(map[string]struct{})
	for _, w := range claimWords {
		if _, stop := stopwords[w]; !stop {
			claimWordSet[w] = struct{}{}
		}
	}

	for _, s := range sentences {
		sClean := strings.TrimSpace(s)
		if len(sClean) < 10 {
			continue
		}
		sTokens := tokenizeText(sClean)
		matches := 0
		for _, st := range sTokens {
			if _, ok := claimWordSet[st]; ok {
				matches++
			}
		}
		if matches > bestMatch {
			bestMatch = matches
			bestSent = sClean
		}
	}

	if bestSent == "" && len(sourceContent) > 120 {
		return strings.TrimSpace(sourceContent[:120]) + "..."
	} else if bestSent == "" {
		return strings.TrimSpace(sourceContent)
	}

	return bestSent + "."
}
