package rag

import (
	"strings"
	"testing"
)

func TestChunker(t *testing.T) {
	chunker := NewChunker(10, 2)
	sampleText := "Word1 Word2 Word3 Word4 Word5 Word6 Word7 Word8 Word9 Word10 Word11 Word12 Word13 Word14 Word15"

	chunks := chunker.ChunkText(sampleText)

	if len(chunks) == 0 {
		t.Fatal("Expected chunker to produce chunks, got 0")
	}

	if chunks[0].Index != 0 {
		t.Errorf("Expected first chunk index 0, got %d", chunks[0].Index)
	}

	wordsInFirstChunk := strings.Fields(chunks[0].Text)
	if len(wordsInFirstChunk) > 10 {
		t.Errorf("Expected max 10 words per chunk, got %d", len(wordsInFirstChunk))
	}
}
