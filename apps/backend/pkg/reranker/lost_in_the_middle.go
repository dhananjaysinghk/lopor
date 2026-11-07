package reranker

import (
	"fmt"
	"strings"
)

// ApplyLostInTheMiddle reorders a descending-sorted list of scored chunks into a U-shaped distribution.
// The most critical documents are positioned at the extreme beginning (primacy effect) and
// extreme end (recency effect) of the context window, insulating against mid-context attention degradation.
func ApplyLostInTheMiddle(chunks []ScoredChunk) []ScoredChunk {
	n := len(chunks)
	if n <= 2 {
		return chunks
	}

	reordered := make([]ScoredChunk, n)
	left := 0
	right := n - 1

	for i, chunk := range chunks {
		if i%2 == 0 {
			reordered[left] = chunk
			left++
		} else {
			reordered[right] = chunk
			right--
		}
	}

	// Update final ranks post-reordering
	for idx := range reordered {
		reordered[idx].FinalRank = idx + 1
		reordered[idx].PositionDelta = reordered[idx].InitialRank - reordered[idx].FinalRank
	}

	return reordered
}

// AssemblePromptContext formats the selected and reordered chunks into a clean, markdown-structured prompt context block.
func AssemblePromptContext(query string, chunks []ScoredChunk) string {
	if len(chunks) == 0 {
		return "No relevant context found."
	}

	var sb strings.Builder
	sb.WriteString("=== BEGIN RE-RANKED RELEVANT CONTEXT ===\n")
	if query != "" {
		sb.WriteString(fmt.Sprintf("Target Query: %s\n\n", query))
	}

	for idx, chunk := range chunks {
		title := chunk.Title
		if title == "" {
			title = fmt.Sprintf("Document_%s", chunk.ID)
		}

		sb.WriteString(fmt.Sprintf("--- [Source %d | Grade: %s | Score: %.3f | ID: %s] ---\n",
			idx+1, chunk.RelevanceGrade, chunk.CombinedScore, chunk.ID))
		sb.WriteString(fmt.Sprintf("Title: %s\n", title))
		sb.WriteString(strings.TrimSpace(chunk.Content))
		sb.WriteString("\n\n")
	}

	sb.WriteString("=== END RE-RANKED RELEVANT CONTEXT ===")
	return sb.String()
}

