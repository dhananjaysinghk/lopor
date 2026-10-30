package semanticdiff

import (
	"crypto/sha256"
	"math"
	"regexp"
	"strings"
)

var (
	reSentenceSplit  = regexp.MustCompile(`(?m)(?<=[.!?])\s+`)
	reParagraphSplit = regexp.MustCompile(`\n\s*\n`)
	reCodeBlockSplit = regexp.MustCompile("(?s)```[a-zA-Z0-9_-]*\\s*\\n(.+?)\\n```")
)

// SegmentText splits text according to the requested granularity.
func SegmentText(text string, granularity DiffGranularity) []string {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return nil
	}

	switch granularity {
	case GranularityParagraph:
		parts := reParagraphSplit.Split(trimmed, -1)
		var res []string
		for _, p := range parts {
			clean := strings.TrimSpace(p)
			if clean != "" {
				res = append(res, clean)
			}
		}
		return res

	case GranularityCodeBlock:
		matches := reCodeBlockSplit.FindAllStringSubmatch(trimmed, -1)
		if len(matches) > 0 {
			var blocks []string
			for _, m := range matches {
				if len(m) > 1 {
					blocks = append(blocks, strings.TrimSpace(m[1]))
				}
			}
			return blocks
		}
		// Fallback to lines if no code fences found
		return strings.Split(trimmed, "\n")

	case GranularitySentence, GranularityClause:
		fallthrough
	default:
		parts := reSentenceSplit.Split(trimmed, -1)
		var res []string
		for _, p := range parts {
			clean := strings.TrimSpace(p)
			if clean != "" {
				res = append(res, clean)
			}
		}
		if len(res) == 0 {
			res = []string{trimmed}
		}
		return res
	}
}

// ComputeSegmentVector computes a deterministic 32-dimensional unit vector from text n-grams.
func ComputeSegmentVector(text string) []float64 {
	vec := make([]float64, 32)
	words := strings.Fields(strings.ToLower(text))
	if len(words) == 0 {
		return vec
	}

	for _, w := range words {
		// Clean punctuation
		cleaned := strings.Trim(w, ",.!?;:\"'()[]{}")
		if cleaned == "" {
			continue
		}
		h := sha256.Sum256([]byte(cleaned))
		for i := 0; i < 32; i++ {
			val := float64(int8(h[i]))
			vec[i] += val
		}
	}

	// Normalize vector to unit length
	var norm float64
	for _, v := range vec {
		norm += v * v
	}
	norm = math.Sqrt(norm)

	if norm > 0 {
		for i := range vec {
			vec[i] /= norm
		}
	}

	return vec
}

// CosineSimilarity computes the dot product of two normalized unit vectors.
func CosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0.0
	}

	var dot float64
	for i := range a {
		dot += a[i] * b[i]
	}

	// Clamp to [0.0, 1.0] for non-negative similarity interpretation
	if dot < 0 {
		dot = 0
	}
	if dot > 1.0 {
		dot = 1.0
	}

	return math.Round(dot*1000) / 1000
}

