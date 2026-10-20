package semanticcache

import (
	"crypto/sha256"
	"math"
	"regexp"
	"strings"
)

// Normalizer prepares query strings and calculates normalized semantic feature vectors.
type Normalizer struct {
	punctRegex *regexp.Regexp
}

// NewNormalizer creates a new query normalizer.
func NewNormalizer() *Normalizer {
	return &Normalizer{
		punctRegex: regexp.MustCompile(`[^a-zA-Z0-9_\-\s]`),
	}
}

// Normalize cleans and canonicalizes a query string.
func (n *Normalizer) Normalize(query string) string {
	cleaned := n.punctRegex.ReplaceAllString(strings.ToLower(query), " ")
	words := strings.Fields(cleaned)
	return strings.Join(words, " ")
}

// ComputeEmbedding generates a 64-dimensional normalized semantic feature vector.
func (n *Normalizer) ComputeEmbedding(text string) []float32 {
	const dims = 64
	vec := make([]float32, dims)

	words := strings.Fields(strings.ToLower(text))
	if len(words) == 0 {
		return vec
	}

	for _, w := range words {
		h := sha256.Sum256([]byte(w))
		// Distribute word hash entropy across dimensions
		for i := 0; i < dims; i++ {
			byteVal := float32(h[i%len(h)])
			weight := (byteVal / 255.0) - 0.5
			vec[i] += weight
		}

		// Bigram character feature capture
		for j := 0; j < len(w)-1; j++ {
			bg := w[j : j+2]
			idx := (int(bg[0])*31 + int(bg[1])) % dims
			vec[idx] += 0.25
		}
	}

	// L2 normalization to unit vector
	var norm float64
	for i := 0; i < dims; i++ {
		norm += float64(vec[i] * vec[i])
	}
	norm = math.Sqrt(norm)

	if norm > 0 {
		for i := 0; i < dims; i++ {
			vec[i] = float32(float64(vec[i]) / norm)
		}
	}

	return vec
}

// CosineSimilarity calculates the dot product of two normalized unit vectors.
func CosineSimilarity(a, b []float32) float64 {
	if len(a) == 0 || len(b) == 0 || len(a) != len(b) {
		return 0.0
	}

	var dot float64
	for i := 0; i < len(a); i++ {
		dot += float64(a[i] * b[i])
	}

	// Clamp between -1.0 and 1.0
	if dot > 1.0 {
		dot = 1.0
	} else if dot < -1.0 {
		dot = -1.0
	}

	return dot
}
