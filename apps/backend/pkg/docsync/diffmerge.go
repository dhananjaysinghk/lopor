package docsync

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// HashContent returns the SHA-256 hexadecimal hash of document text.
func HashContent(text string) string {
	sum := sha256.Sum256([]byte(text))
	return hex.EncodeToString(sum[:])
}

// ThreeWayMergeText reconciles changes between base, local, and remote text versions line-by-line.
func ThreeWayMergeText(baseText, localText, remoteText string) (string, []ConflictHunk, bool) {
	if localText == remoteText {
		return localText, nil, false
	}
	if baseText == localText {
		return remoteText, nil, false // Clean fast-forward to remote
	}
	if baseText == remoteText {
		return localText, nil, false // Clean fast-forward to local
	}

	baseLines := splitLines(baseText)
	localLines := splitLines(localText)
	remoteLines := splitLines(remoteText)

	maxLines := len(baseLines)
	if len(localLines) > maxLines {
		maxLines = len(localLines)
	}
	if len(remoteLines) > maxLines {
		maxLines = len(remoteLines)
	}

	var mergedLines []string
	var conflicts []ConflictHunk
	hasConflict := false

	for i := 0; i < maxLines; i++ {
		baseLine := getLine(baseLines, i)
		localLine := getLine(localLines, i)
		remoteLine := getLine(remoteLines, i)

		// Case 1: Both changed identically
		if localLine == remoteLine {
			mergedLines = append(mergedLines, localLine)
			continue
		}

		// Case 2: Only local changed
		if remoteLine == baseLine {
			mergedLines = append(mergedLines, localLine)
			continue
		}

		// Case 3: Only remote changed
		if localLine == baseLine {
			mergedLines = append(mergedLines, remoteLine)
			continue
		}

		// Case 4: True conflict - both modified base line differently
		hasConflict = true
		hunkID := fmt.Sprintf("hunk_%d", len(conflicts)+1)
		hunk := ConflictHunk{
			HunkID:        hunkID,
			BaseContent:   baseLine,
			LocalContent:  localLine,
			RemoteContent: remoteLine,
			IsResolved:    false,
		}
		conflicts = append(conflicts, hunk)

		// Insert conflict marker representation into merged draft
		conflictMarker := fmt.Sprintf("<<<<<<< LOCAL (%s)\n%s\n=======\n%s\n>>>>>>> REMOTE", hunkID, localLine, remoteLine)
		mergedLines = append(mergedLines, conflictMarker)
	}

	return strings.Join(mergedLines, "\n"), conflicts, hasConflict
}

// ArbitrateSemanticMerge resolves conflicts favoring synthetic consensus or non-destructive union.
func ArbitrateSemanticMerge(base, local, remote string, conflicts []ConflictHunk) string {
	if len(conflicts) == 0 {
		merged, _, _ := ThreeWayMergeText(base, local, remote)
		return merged
	}

	// For each conflict, produce a synthesized resolution
	var resolvedLines []string
	lines := strings.Split(base, "\n")

	for i, c := range conflicts {
		var syntheticPick string
		// If one is agent and other is human, preserve human edits while annotating agent nuance
		if strings.TrimSpace(c.LocalContent) == "" {
			syntheticPick = c.RemoteContent
		} else if strings.TrimSpace(c.RemoteContent) == "" {
			syntheticPick = c.LocalContent
		} else {
			// Hybrid integration: retain both clauses with clear demarcation
			syntheticPick = fmt.Sprintf("%s (Integrated: %s)", strings.TrimSpace(c.LocalContent), strings.TrimSpace(c.RemoteContent))
		}
		conflicts[i].ResolvedContent = syntheticPick
		conflicts[i].IsResolved = true
		conflicts[i].ResolutionNote = "Arbitrated via Semantic AI synthesis."
		resolvedLines = append(resolvedLines, syntheticPick)
	}

	if len(lines) > len(resolvedLines) {
		resolvedLines = append(resolvedLines, lines[len(resolvedLines):]...)
	}

	return strings.Join(resolvedLines, "\n")
}

func splitLines(text string) []string {
	if text == "" {
		return []string{}
	}
	return strings.Split(text, "\n")
}

func getLine(lines []string, index int) string {
	if index < len(lines) {
		return lines[index]
	}
	return ""
}

