package memory

import (
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Extractor identifies episodic facts and preferences from dialogue.
type Extractor struct {
	stopWords map[string]struct{}
}

// NewExtractor initializes the semantic extractor.
func NewExtractor() *Extractor {
	stops := []string{
		"a", "about", "above", "after", "again", "against", "all", "am", "an", "and", "any", "are", "aren't",
		"as", "at", "be", "because", "been", "before", "being", "below", "between", "both", "but", "by", "can",
		"cannot", "could", "did", "do", "does", "doing", "don't", "down", "during", "each", "few", "for", "from",
		"further", "had", "has", "have", "having", "he", "her", "here", "hers", "herself", "him", "himself", "his",
		"how", "i", "if", "in", "into", "is", "isn't", "it", "its", "itself", "let's", "me", "more", "most", "my",
		"myself", "no", "nor", "not", "of", "off", "on", "once", "only", "or", "other", "ought", "our", "ours",
		"ourselves", "out", "over", "own", "same", "she", "should", "so", "some", "such", "than", "that", "the",
		"their", "theirs", "them", "themselves", "then", "there", "these", "they", "this", "those", "through", "to",
		"too", "under", "until", "up", "very", "was", "we", "were", "what", "when", "where", "which", "while", "who",
		"whom", "why", "with", "won't", "would", "you", "your", "yours", "yourself", "yourselves",
	}

	stopMap := make(map[string]struct{}, len(stops))
	for _, s := range stops {
		stopMap[s] = struct{}{}
	}

	return &Extractor{stopWords: stopMap}
}

// memoryPattern maps semantic cue phrases to categories and base importance.
type memoryPattern struct {
	category       MemoryCategory
	cueRegex       *regexp.Regexp
	baseImportance float64
	baseConfidence float64
}

var extractionRules = []memoryPattern{
	{
		category:       CategoryConstraint,
		cueRegex:       regexp.MustCompile(`(?i)\b(never use|do not use|don't use|must not|avoid using|ensure you avoid|cannot exceed|strict constraint|forbidden to)\b\s*(.+)`),
		baseImportance: 0.95,
		baseConfidence: 0.92,
	},
	{
		category:       CategoryGoal,
		cueRegex:       regexp.MustCompile(`(?i)\b(my goal is|our goal is|objective is|target deadline is|aiming to|we want to build|deliverable is)\b\s*(.+)`),
		baseImportance: 0.88,
		baseConfidence: 0.88,
	},
	{
		category:       CategoryTechStack,
		cueRegex:       regexp.MustCompile(`(?i)\b(we use|our stack is|our database is|backend is|frontend is|deployed on|running on|using postgres|using redis|built with)\b\s*(.+)`),
		baseImportance: 0.82,
		baseConfidence: 0.89,
	},
	{
		category:       CategoryPreference,
		cueRegex:       regexp.MustCompile(`(?i)\b(i prefer|my preference is|i like|prefer to use|always write|always use|keep answers|format with|favor)\b\s*(.+)`),
		baseImportance: 0.78,
		baseConfidence: 0.85,
	},
	{
		category:       CategoryDomainFact,
		cueRegex:       regexp.MustCompile(`(?i)\b(our company is|the project is|internal api is|our sla is|business rule requires|our system is called)\b\s*(.+)`),
		baseImportance: 0.75,
		baseConfidence: 0.82,
	},
}

// ExtractFromTurns analyzes dialogue and identifies salient memories.
func (e *Extractor) ExtractFromTurns(workspaceID, userID string, turns []DialogueTurn) []MemoryItem {
	var results []MemoryItem
	seenContent := make(map[string]bool)

	for _, turn := range turns {
		// Long-term personal memories primarily stem from user messages and confirmed system facts
		if turn.Role == "assistant" && len(turns) > 1 {
			continue
		}

		sentences := splitIntoSentences(turn.Content)
		for _, s := range sentences {
			trimmed := strings.TrimSpace(s)
			if len(trimmed) < 10 {
				continue
			}

			for _, rule := range extractionRules {
				if matches := rule.cueRegex.FindStringSubmatch(trimmed); len(matches) > 0 {
					cleaned := cleanSentence(trimmed)
					if seenContent[strings.ToLower(cleaned)] {
						continue
					}
					seenContent[strings.ToLower(cleaned)] = true

					keywords := e.extractKeywords(cleaned)
					if len(keywords) == 0 {
						continue
					}

					now := time.Now().UTC()
					item := MemoryItem{
						ID:             uuid.New().String(),
						WorkspaceID:    workspaceID,
						UserID:         userID,
						Category:       rule.category,
						Content:        cleaned,
						Keywords:       keywords,
						Importance:     rule.baseImportance,
						Confidence:     rule.baseConfidence,
						AccessCount:    1,
						CreatedAt:      now,
						LastAccessedAt: now,
						DecayScore:     rule.baseImportance,
						IsActive:       true,
					}
					results = append(results, item)
					break // avoid multiple category matches on the exact same sentence
				}
			}
		}
	}

	return results
}

// extractKeywords extracts unique, relevant non-stop words.
func (e *Extractor) extractKeywords(text string) []string {
	// Strip punctuation
	reg := regexp.MustCompile(`[^a-zA-Z0-9_\-\s]`)
	cleaned := reg.ReplaceAllString(strings.ToLower(text), " ")

	words := strings.Fields(cleaned)
	seen := make(map[string]bool)
	var keywords []string

	for _, w := range words {
		w = strings.TrimSpace(w)
		if len(w) < 3 {
			continue
		}
		if _, isStop := e.stopWords[w]; isStop {
			continue
		}
		if !seen[w] {
			seen[w] = true
			keywords = append(keywords, w)
		}
	}

	return keywords
}

func splitIntoSentences(text string) []string {
	reg := regexp.MustCompile(`[.!?\n]+`)
	raw := reg.Split(text, -1)
	var out []string
	for _, r := range raw {
		r = strings.TrimSpace(r)
		if len(r) > 0 {
			out = append(out, r)
		}
	}
	return out
}

func cleanSentence(text string) string {
	text = strings.TrimSpace(text)
	text = strings.TrimRight(text, ".!?;,")
	return text
}
