package grounding

import (
	"fmt"
	"strings"
)

// BuildAttributedResponse generates an annotated response containing inline citations and bibliography.
func BuildAttributedResponse(original string, claims []AtomicClaim, sources []SourceEvidence) (string, []CitationAnchor, string) {
	if len(claims) == 0 {
		return original, nil, original
	}

	sourceMap := make(map[string]SourceEvidence)
	for _, s := range sources {
		sourceMap[s.ID] = s
	}

	var citations []CitationAnchor
	citationCounter := 1

	var attributedBuilder strings.Builder
	var remediatedBuilder strings.Builder

	for idx, claim := range claims {
		claimText := claim.ClaimText

		if claim.Status == StatusSupported {
			// Supported: inject citation
			src := sourceMap[claim.SourceEvidenceID]
			cit := CitationAnchor{
				CitationIndex: citationCounter,
				ClaimID:       claim.ID,
				SourceID:      claim.SourceEvidenceID,
				DocumentTitle: claim.SourceTitle,
				URI:           src.URI,
				ExactSnippet:  claim.EvidenceSnippet,
			}
			citations = append(citations, cit)

			marker := fmt.Sprintf(" [^%d]", citationCounter)
			attributedBuilder.WriteString(claimText + marker)
			remediatedBuilder.WriteString(claimText + marker)
			citationCounter++
		} else if claim.Status == StatusContradicted {
			// Contradicted: flag critical hallucination
			attributedBuilder.WriteString(claimText + " [⚠️ CONTRADICTION]")
			remediatedBuilder.WriteString(fmt.Sprintf("~~[Retracted: %s]~~", claimText))
		} else {
			// Ungrounded: flag unverified statement
			attributedBuilder.WriteString(claimText + " [⚠️ UNVERIFIED]")
			remediatedBuilder.WriteString(fmt.Sprintf("[Unverified claim removed]"))
		}

		if idx < len(claims)-1 {
			attributedBuilder.WriteString(". ")
			remediatedBuilder.WriteString(". ")
		} else {
			attributedBuilder.WriteString(".")
			remediatedBuilder.WriteString(".")
		}
	}

	// Append Bibliography to AttributedResponse if citations exist
	if len(citations) > 0 {
		attributedBuilder.WriteString("\n\n### References & Citations\n")
		for _, c := range citations {
			uri := c.URI
			if uri == "" {
				uri = fmt.Sprintf("doc://%s", c.SourceID)
			}
			attributedBuilder.WriteString(fmt.Sprintf("[^%d]: **%s** — \"%s\" ([Source](%s))\n",
				c.CitationIndex, c.DocumentTitle, c.ExactSnippet, uri))
		}
	}

	return attributedBuilder.String(), citations, remediatedBuilder.String()
}

