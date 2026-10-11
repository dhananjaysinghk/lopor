package memory

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

// MemoryEngine coordinates long-term episodic memory extraction, storage, and contextual recall.
type MemoryEngine struct {
	extractor *Extractor
	store     *MemoryStore
}

// NewMemoryEngine initializes the long-term memory engine with default configurations.
func NewMemoryEngine() *MemoryEngine {
	return &MemoryEngine{
		extractor: NewExtractor(),
		store:     NewMemoryStore(14.0), // 14-day recency half-life
	}
}

// GetStore returns the underlying thread-safe memory store.
func (e *MemoryEngine) GetStore() *MemoryStore {
	return e.store
}

// ExtractAndStore parses incoming dialogue turns, identifies salient memories, and persists them.
func (e *MemoryEngine) ExtractAndStore(ctx context.Context, req ExtractionRequest) (*ExtractionResult, error) {
	turns := make([]DialogueTurn, len(req.DialogueHistory))
	copy(turns, req.DialogueHistory)

	if req.CurrentMessage != "" {
		turns = append(turns, DialogueTurn{
			Role:      "user",
			Content:   req.CurrentMessage,
			Timestamp: time.Now().UTC(),
		})
	}

	extracted := e.extractor.ExtractFromTurns(req.WorkspaceID, req.UserID, turns)
	categoryMap := make(map[MemoryCategory]bool)
	var saved []MemoryItem

	for _, item := range extracted {
		stored := e.store.Save(item)
		saved = append(saved, stored)
		categoryMap[item.Category] = true
	}

	categories := make([]MemoryCategory, 0, len(categoryMap))
	for cat := range categoryMap {
		categories = append(categories, cat)
	}

	return &ExtractionResult{
		ExtractedMemories:    saved,
		IdentifiedCategories: categories,
		ProcessedTurnCount:   len(turns),
	}, nil
}

// RecallMemories retrieves prioritized memories relevant to a query.
func (e *MemoryEngine) RecallMemories(ctx context.Context, req RecallRequest) (*RecallResult, error) {
	maxResults := req.MaxResults
	if maxResults <= 0 {
		maxResults = 5
	}
	minRelevance := req.MinRelevance
	if minRelevance <= 0.0 {
		minRelevance = 0.20
	}

	// Update decay dynamics
	e.store.ApplyDecay()

	candidates := e.store.List(req.WorkspaceID, req.UserID, req.CategoryFilter)
	queryTokens := e.extractor.extractKeywords(req.Query)
	querySet := make(map[string]bool)
	for _, t := range queryTokens {
		querySet[t] = true
	}

	var scored []RecalledMemory
	for _, mem := range candidates {
		// Calculate keyword overlap
		overlap := 0
		for _, kw := range mem.Keywords {
			if querySet[kw] {
				overlap++
			}
		}

		// Also check query tokens directly inside memory content
		contentLower := strings.ToLower(mem.Content)
		textMatches := 0
		for _, qt := range queryTokens {
			if strings.Contains(contentLower, qt) {
				textMatches++
			}
		}

		totalTokens := len(queryTokens)
		if totalTokens == 0 {
			totalTokens = 1
		}

		relevanceScore := 0.0
		if len(mem.Keywords) > 0 {
			jaccard := float64(overlap) / float64(len(queryTokens)+len(mem.Keywords)-overlap)
			textRatio := float64(textMatches) / float64(totalTokens)
			relevanceScore = (0.6 * jaccard) + (0.4 * textRatio)
		} else {
			relevanceScore = float64(textMatches) / float64(totalTokens)
		}

		// Constraints always receive a high baseline relevance bonus so instructions aren't missed
		if mem.Category == CategoryConstraint {
			relevanceScore = math.Max(relevanceScore, 0.45)
		}

		if relevanceScore < minRelevance {
			continue
		}

		recencyScore := mem.DecayScore
		// Combined formula: 0.5 * Relevance + 0.3 * Importance + 0.2 * Recency
		combined := (0.50 * relevanceScore) + (0.30 * mem.Importance) + (0.20 * recencyScore)

		scored = append(scored, RecalledMemory{
			Memory:         mem,
			RelevanceScore: math.Round(relevanceScore*100) / 100,
			RecencyScore:   math.Round(recencyScore*100) / 100,
			CombinedScore:  math.Round(combined*100) / 100,
		})
	}

	// Sort descending by combined score
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].CombinedScore > scored[j].CombinedScore
	})

	if len(scored) > maxResults {
		scored = scored[:maxResults]
	}

	// Increment access count for the recalled memories
	for _, s := range scored {
		e.store.IncrementAccess(s.Memory.ID)
	}

	promptBlock := e.FormatPromptInjection(scored)

	return &RecallResult{
		Items:                scored,
		FormattedPromptBlock: promptBlock,
		TotalMemoriesScanned: len(candidates),
	}, nil
}

// Consolidate merges overlapping memories and prunes expired knowledge.
func (e *MemoryEngine) Consolidate(ctx context.Context, workspaceID, userID string) (*ConsolidationReport, error) {
	candidates := e.store.List(workspaceID, userID, nil)
	totalBefore := len(candidates)

	merged := 0
	pruned := 0
	toDelete := make(map[string]bool)

	// Detect near-duplicate memories in the same category
	for i := 0; i < len(candidates); i++ {
		if toDelete[candidates[i].ID] {
			continue
		}
		for j := i + 1; j < len(candidates); j++ {
			if toDelete[candidates[j].ID] {
				continue
			}
			if candidates[i].Category != candidates[j].Category {
				continue
			}

			// Calculate token overlap
			sim := jaccardSimilarity(candidates[i].Keywords, candidates[j].Keywords)
			if sim >= 0.60 {
				// Keep the more recent or higher importance memory, merge access counts
				primary := &candidates[i]
				secondary := &candidates[j]
				if secondary.CreatedAt.After(primary.CreatedAt) || secondary.Importance > primary.Importance {
					primary, secondary = secondary, primary
				}

				primary.AccessCount += secondary.AccessCount
				e.store.Save(*primary)
				toDelete[secondary.ID] = true
				merged++
			}
		}
	}

	for id := range toDelete {
		if e.store.Delete(id) {
			pruned++
		}
	}

	totalAfter := e.store.Count(workspaceID, userID)

	return &ConsolidationReport{
		WorkspaceID:    workspaceID,
		UserID:         userID,
		TotalBefore:    totalBefore,
		TotalAfter:     totalAfter,
		MergedCount:    merged,
		PrunedCount:    pruned,
		ConsolidatedAt: time.Now().UTC(),
	}, nil
}

// FormatPromptInjection formats recalled memories into a structured system prompt injection string.
func (e *MemoryEngine) FormatPromptInjection(items []RecalledMemory) string {
	if len(items) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("<user_long_term_memory>\n")
	sb.WriteString("Retrieved user and workspace preferences, tech stack conventions, and critical constraints:\n")

	for _, item := range items {
		sb.WriteString(fmt.Sprintf("- [%s] %s (Importance: %.2f, Relevance: %.2f)\n",
			item.Memory.Category,
			item.Memory.Content,
			item.Memory.Importance,
			item.RelevanceScore,
		))
	}

	sb.WriteString("</user_long_term_memory>")
	return sb.String()
}

func jaccardSimilarity(a, b []string) float64 {
	if len(a) == 0 && len(b) == 0 {
		return 1.0
	}
	if len(a) == 0 || len(b) == 0 {
		return 0.0
	}

	setA := make(map[string]bool)
	for _, val := range a {
		setA[val] = true
	}

	intersection := 0
	unionMap := make(map[string]bool)
	for _, val := range a {
		unionMap[val] = true
	}
	for _, val := range b {
		if setA[val] {
			intersection++
		}
		unionMap[val] = true
	}

	if len(unionMap) == 0 {
		return 0.0
	}
	return float64(intersection) / float64(len(unionMap))
}
