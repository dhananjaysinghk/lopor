package rag

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"
)

type EmbeddingGenerator interface {
	GenerateVector(ctx context.Context, text string) ([]float32, error)
}

type MockEmbedder struct {
	Dimension int // 1536
}

func NewMockEmbedder(dimension int) *MockEmbedder {
	if dimension <= 0 {
		dimension = 1536
	}
	return &MockEmbedder{Dimension: dimension}
}

// GenerateVector produces a normalized 1536-dimensional float vector for pgvector storage
func (m *MockEmbedder) GenerateVector(ctx context.Context, text string) ([]float32, error) {
	vec := make([]float32, m.Dimension)
	seed := int64(0)
	for _, ch := range text {
		seed += int64(ch)
	}

	r := rand.New(rand.NewSource(seed + time.Now().UnixNano()))
	var sumSq float64

	for i := 0; i < m.Dimension; i++ {
		val := r.Float64()*2 - 1
		vec[i] = float32(val)
		sumSq += val * val
	}

	norm := math.Sqrt(sumSq)
	if norm > 0 {
		for i := 0; i < m.Dimension; i++ {
			vec[i] /= float32(norm)
		}
	}

	return vec, nil
}

// FormatPgVector converts []float32 to pgvector string format "[0.1, 0.2, ...]"
func FormatPgVector(vec []float32) string {
	var sb strings.Builder
	sb.WriteString("[")
	for i, v := range vec {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(fmt.Sprintf("%f", v))
	}
	sb.WriteString("]")
	return sb.String()
}
