package knowledgemerge

import (
	"fmt"
	"strings"
)

// ResolveContradiction evaluates two or more conflicting facts for the same Subject and Predicate.
func ResolveContradiction(facts []KnowledgeFact, strategy ResolutionStrategy) (*KnowledgeFact, string) {
	if len(facts) == 0 {
		return nil, "No facts provided."
	}
	if len(facts) == 1 {
		return &facts[0], "Single fact without contradiction."
	}

	switch strategy {
	case StrategyAuthorityHierarchy:
		// Priority: HUMAN_EXPERT > AI_AGENT > RAG_EXTRACTION
		var bestFact *KnowledgeFact
		highestRank := -1

		for i := range facts {
			f := &facts[i]
			rank := 0
			switch strings.ToUpper(f.SourceType) {
			case "HUMAN_EXPERT":
				rank = 3
			case "AI_AGENT":
				rank = 2
			case "RAG_EXTRACTION":
				rank = 1
			}

			if rank > highestRank || (rank == highestRank && f.Confidence > bestFact.Confidence) {
				highestRank = rank
				bestFact = f
			}
		}
		explanation := fmt.Sprintf("Selected authoritative source '%s' (type: %s, confidence: %.2f) under Authority Hierarchy.",
			bestFact.SourceID, bestFact.SourceType, bestFact.Confidence)
		return bestFact, explanation

	case StrategyConfidenceVote:
		// Highest numerical confidence wins
		var bestFact *KnowledgeFact
		highestConf := -1.0

		for i := range facts {
			f := &facts[i]
			if f.Confidence > highestConf {
				highestConf = f.Confidence
				bestFact = f
			}
		}
		explanation := fmt.Sprintf("Selected fact with highest confidence score (%.2f) from source '%s'.",
			bestFact.Confidence, bestFact.SourceID)
		return bestFact, explanation

	case StrategySemanticSynthesis:
		// If objects can be logically combined
		var objNames []string
		for _, f := range facts {
			objNames = append(objNames, f.ObjectName)
		}
		synthesized := facts[0]
		synthesized.FactID = synthesized.FactID + "_synth"
		synthesized.ObjectName = strings.Join(objNames, " & ")
		explanation := fmt.Sprintf("Synthesized non-exclusive relation combining values: [%s].", strings.Join(objNames, ", "))
		return &synthesized, explanation

	case StrategyRecencyWeighted:
		fallthrough
	default:
		// Latest ValidFrom or RecordedAt timestamp takes precedence
		var latestFact *KnowledgeFact
		for i := range facts {
			f := &facts[i]
			if latestFact == nil || f.ValidFrom.After(latestFact.ValidFrom) {
				latestFact = f
			}
		}
		explanation := fmt.Sprintf("Selected most recent fact (ValidFrom: %s) under Recency Weighted strategy.",
			latestFact.ValidFrom.Format("2006-01-02"))
		return latestFact, explanation
	}
}

