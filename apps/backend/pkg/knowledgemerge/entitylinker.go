package knowledgemerge

import (
	"math"
	"strings"
)

// StringSimilarity computes an n-gram token overlap similarity (0.0 to 1.0).
func StringSimilarity(a, b string) float64 {
	cleanA := strings.ToLower(strings.TrimSpace(a))
	cleanB := strings.ToLower(strings.TrimSpace(b))

	if cleanA == cleanB {
		return 1.0
	}
	if cleanA == "" || cleanB == "" {
		return 0.0
	}

	wordsA := strings.Fields(cleanA)
	wordsB := strings.Fields(cleanB)

	setA := make(map[string]bool)
	for _, w := range wordsA {
		setA[w] = true
	}

	intersection := 0
	for _, w := range wordsB {
		if setA[w] {
			intersection++
		}
	}

	union := len(wordsA) + len(wordsB) - intersection
	if union <= 0 {
		return 0.0
	}

	return math.Round((float64(intersection)/float64(union))*1000) / 1000
}

// FindMatchingEntity determines if an entity name matches an existing canonical entity or alias.
func FindMatchingEntity(entities map[string]*KnowledgeEntity, name string, threshold float64) *KnowledgeEntity {
	cleanName := strings.ToLower(strings.TrimSpace(name))

	// Direct match check
	for _, ent := range entities {
		if strings.ToLower(ent.CanonicalName) == cleanName {
			return ent
		}
		for _, alias := range ent.Aliases {
			if strings.ToLower(alias) == cleanName {
				return ent
			}
		}
	}

	// Approximate token overlap check
	var bestMatch *KnowledgeEntity
	highestSim := 0.0

	for _, ent := range entities {
		sim := StringSimilarity(cleanName, ent.CanonicalName)
		if sim > highestSim && sim >= threshold {
			highestSim = sim
			bestMatch = ent
		}
	}

	return bestMatch
}

// MergeEntityAttributes combines attributes and aliases non-destructively.
func MergeEntityAttributes(target *KnowledgeEntity, incoming KnowledgeEntity) {
	aliasMap := make(map[string]bool)
	for _, a := range target.Aliases {
		aliasMap[strings.ToLower(a)] = true
	}
	aliasMap[strings.ToLower(target.CanonicalName)] = true

	if !aliasMap[strings.ToLower(incoming.CanonicalName)] {
		target.Aliases = append(target.Aliases, incoming.CanonicalName)
		aliasMap[strings.ToLower(incoming.CanonicalName)] = true
	}

	for _, a := range incoming.Aliases {
		if !aliasMap[strings.ToLower(a)] {
			target.Aliases = append(target.Aliases, a)
			aliasMap[strings.ToLower(a)] = true
		}
	}

	if target.Attributes == nil {
		target.Attributes = make(map[string]string)
	}
	for k, v := range incoming.Attributes {
		if _, exists := target.Attributes[k]; !exists {
			target.Attributes[k] = v
		}
	}
}

