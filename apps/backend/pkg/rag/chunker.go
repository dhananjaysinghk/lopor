package rag

import (
	"strings"
)

type Chunk struct {
	Index      int
	Text       string
	TokenCount int
}

type Chunker struct {
	ChunkSize int // Default 512
	Overlap   int // Default 64
}

func NewChunker(chunkSize, overlap int) *Chunker {
	if chunkSize <= 0 {
		chunkSize = 512
	}
	if overlap < 0 {
		overlap = 64
	}
	return &Chunker{
		ChunkSize: chunkSize,
		Overlap:   overlap,
	}
}

// ChunkText splits raw text into overlapping recursive character chunks
func (c *Chunker) ChunkText(text string) []Chunk {
	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}

	var chunks []Chunk
	chunkIdx := 0

	step := c.ChunkSize - c.Overlap
	if step <= 0 {
		step = c.ChunkSize / 2
	}

	for i := 0; i < len(words); i += step {
		end := i + c.ChunkSize
		if end > len(words) {
			end = len(words)
		}

		chunkText := strings.Join(words[i:end], " ")
		chunks = append(chunks, Chunk{
			Index:      chunkIdx,
			Text:       chunkText,
			TokenCount: end - i,
		})
		chunkIdx++

		if end == len(words) {
			break
		}
	}

	return chunks
}
