package chunking

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"
)

// StructuralChunker converts document sections into enriched chunks respecting token budgets.
type StructuralChunker struct {
	parser *Parser
}

// NewStructuralChunker initializes the chunker.
func NewStructuralChunker() *StructuralChunker {
	return &StructuralChunker{
		parser: NewParser(),
	}
}

// ChunkSections processes sections into a final slice of EnrichedChunks.
func (sc *StructuralChunker) ChunkSections(
	docID, wsID string,
	sections []DocumentSection,
	opts ChunkingOptions,
) []EnrichedChunk {
	if opts.TargetTokens <= 0 {
		opts.TargetTokens = 400
	}
	if opts.MaxTokens <= 0 {
		opts.MaxTokens = 600
	}
	if opts.MinOrphanTokens <= 0 {
		opts.MinOrphanTokens = 50
	}

	var rawChunks []EnrichedChunk
	now := time.Now().UTC()

	for _, sec := range sections {
		tokenCount := estimateTokens(sec.Content)

		if tokenCount <= opts.MaxTokens {
			// Section comfortably fits within token budget
			rawChunks = append(rawChunks, EnrichedChunk{
				DocumentID:     docID,
				WorkspaceID:    wsID,
				Content:        sec.Content,
				BreadcrumbPath: sec.BreadcrumbPath,
				Format:         sec.Format,
				TokenCount:     tokenCount,
				HasCodeBlock:   sec.HasCodeBlock,
				HasTable:       sec.HasTable,
				KeyPhrases:     extractKeyPhrases(sec.Content),
				CreatedAt:      now,
			})
			continue
		}

		// Oversized section: split along paragraphs or sentences
		subChunks := sc.splitOversized(docID, wsID, sec, opts, now)
		rawChunks = append(rawChunks, subChunks...)
	}

	// Post-processing: Merge tiny orphan chunks into preceding chunks
	merged := sc.mergeOrphans(rawChunks, opts.MinOrphanTokens, opts.MaxTokens)

	// Finalize IDs, indices, and offsets
	var finalized []EnrichedChunk
	currentOffset := 0

	for idx, chk := range merged {
		h := sha256.Sum256([]byte(fmt.Sprintf("%s:%d:%s", docID, idx, chk.Content)))
		chk.ID = "chk_" + hex.EncodeToString(h[:8])
		chk.ChunkIndex = idx
		chk.StartOffset = currentOffset
		chk.EndOffset = currentOffset + len(chk.Content)
		chk.TokenCount = estimateTokens(chk.Content)
		currentOffset = chk.EndOffset + 1
		finalized = append(finalized, chk)
	}

	return finalized
}

func (sc *StructuralChunker) splitOversized(
	docID, wsID string,
	sec DocumentSection,
	opts ChunkingOptions,
	now time.Time,
) []EnrichedChunk {
	paragraphs := strings.Split(sec.Content, "\n\n")
	var chunks []EnrichedChunk

	var current strings.Builder
	currentTokens := 0

	flush := func() {
		text := strings.TrimSpace(current.String())
		if text != "" {
			chunks = append(chunks, EnrichedChunk{
				DocumentID:     docID,
				WorkspaceID:    wsID,
				Content:        text,
				BreadcrumbPath: sec.BreadcrumbPath,
				Format:         sec.Format,
				TokenCount:     estimateTokens(text),
				HasCodeBlock:   sec.HasCodeBlock,
				HasTable:       sec.HasTable,
				KeyPhrases:     extractKeyPhrases(text),
				CreatedAt:      now,
			})
		}
		current.Reset()
		currentTokens = 0
	}

	for _, p := range paragraphs {
		pTokens := estimateTokens(p)

		if currentTokens+pTokens > opts.TargetTokens && currentTokens > 0 {
			flush()
		}

		if current.Len() > 0 {
			current.WriteString("\n\n")
		}
		current.WriteString(p)
		currentTokens += pTokens
	}

	if current.Len() > 0 {
		flush()
	}

	return chunks
}

func (sc *StructuralChunker) mergeOrphans(chunks []EnrichedChunk, minOrphanTokens, maxTokens int) []EnrichedChunk {
	if len(chunks) <= 1 {
		return chunks
	}

	var merged []EnrichedChunk

	for i := 0; i < len(chunks); i++ {
		curr := chunks[i]

		// If current chunk is an orphan and we have a preceding chunk to merge into
		if curr.TokenCount < minOrphanTokens && len(merged) > 0 {
			lastIdx := len(merged) - 1
			combinedTokens := merged[lastIdx].TokenCount + curr.TokenCount
			if combinedTokens <= maxTokens {
				merged[lastIdx].Content += "\n\n" + curr.Content
				merged[lastIdx].TokenCount = combinedTokens
				merged[lastIdx].HasCodeBlock = merged[lastIdx].HasCodeBlock || curr.HasCodeBlock
				merged[lastIdx].HasTable = merged[lastIdx].HasTable || curr.HasTable
				continue
			}
		}

		merged = append(merged, curr)
	}

	return merged
}

func estimateTokens(text string) int {
	words := strings.Fields(text)
	if len(words) == 0 {
		return 0
	}
	// Heuristic: standard tokenizers (cl100k) average ~1.3 tokens per word in English
	return int(math.Ceil(float64(len(words)) * 1.3))
}

func extractKeyPhrases(text string) []string {
	reg := regexp.MustCompile(`[^a-zA-Z0-9_\-\s]`)
	cleaned := reg.ReplaceAllString(strings.ToLower(text), " ")

	stopWords := map[string]bool{
		"the": true, "and": true, "for": true, "with": true, "that": true,
		"this": true, "from": true, "have": true, "will": true, "your": true,
		"they": true, "been": true, "were": true, "what": true, "when": true,
		"where": true, "which": true, "about": true, "into": true, "some": true,
	}

	words := strings.Fields(cleaned)
	counts := make(map[string]int)
	for _, w := range words {
		if len(w) > 3 && !stopWords[w] {
			counts[w]++
		}
	}

	type kv struct {
		k string
		v int
	}
	var pairs []kv
	for k, v := range counts {
		pairs = append(pairs, kv{k, v})
	}

	// Pick top 5 most frequent terms
	var phrases []string
	for i := 0; i < len(pairs) && i < 5; i++ {
		phrases = append(phrases, pairs[i].k)
	}
	return phrases
}
