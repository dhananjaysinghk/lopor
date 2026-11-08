package grounding

import (
	"fmt"
	"strings"
	"unicode"
)

var fillerPrefixes = []string{
	"in conclusion", "to summarize", "i hope this helps", "certainly",
	"here is the information", "as an ai", "sure", "of course",
	"let me know if you need", "thank you", "in summary",
}

// DecomposeResponse extracts atomic factual propositions and their byte spans from generated text.
func DecomposeResponse(text string) []AtomicClaim {
	var claims []AtomicClaim
	if strings.TrimSpace(text) == "" {
		return claims
	}

	runes := []rune(text)
	n := len(runes)
	start := 0
	claimSeq := 1

	for i := 0; i < n; i++ {
		r := runes[i]
		isBoundary := false

		if r == '.' || r == '!' || r == '?' || r == '\n' {
			// Avoid splitting on numbers (e.g. 3.14) or common abbreviations (e.g. e.g., i.e.)
			if r == '.' && i+1 < n && unicode.IsDigit(runes[i+1]) {
				continue
			}
			isBoundary = true
		} else if i == n-1 {
			isBoundary = true
			i = n // Include last character
		}

		if isBoundary {
			candidate := strings.TrimSpace(string(runes[start:i]))
			if len(candidate) > 10 && !isFiller(candidate) {
				// Find actual trimmed offsets in original text
				actualStart := start
				for actualStart < i && unicode.IsSpace(runes[actualStart]) {
					actualStart++
				}
				actualEnd := i
				for actualEnd > actualStart && unicode.IsSpace(runes[actualEnd-1]) {
					actualEnd--
				}

				claimID := fmt.Sprintf("claim_%03d", claimSeq)
				claimSeq++

				claims = append(claims, AtomicClaim{
					ID:          claimID,
					ClaimText:   candidate,
					OffsetStart: actualStart,
					OffsetEnd:   actualEnd,
				})
			}
			start = i + 1
		}
	}

	return claims
}

func isFiller(sentence string) bool {
	lower := strings.ToLower(strings.TrimSpace(sentence))
	for _, prefix := range fillerPrefixes {
		if strings.HasPrefix(lower, prefix) {
			return true
		}
	}
	return false
}

