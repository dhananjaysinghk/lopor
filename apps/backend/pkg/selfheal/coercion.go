package selfheal

import (
	"encoding/json"
	"regexp"
	"strings"
)

var (
	reMarkdownFence = regexp.MustCompile("(?s)^\\s*```(?:json|JSON)?\\s*(.+?)\\s*```\\s*$")
	reTrailingComma = regexp.MustCompile(`,(\s*[}\]])`)
)

// CoerceJSON attempts to extract and repair malformed or conversational JSON payloads.
func CoerceJSON(raw string) (string, bool) {
	cleaned := strings.TrimSpace(raw)
	if cleaned == "" {
		return "", false
	}

	// 1. Strip markdown fences if present
	if match := reMarkdownFence.FindStringSubmatch(cleaned); len(match) > 1 {
		cleaned = strings.TrimSpace(match[1])
	}

	// 2. Normalize smart quotes to standard ASCII quotes
	cleaned = strings.ReplaceAll(cleaned, "“", "\"")
	cleaned = strings.ReplaceAll(cleaned, "”", "\"")
	cleaned = strings.ReplaceAll(cleaned, "‘", "'")
	cleaned = strings.ReplaceAll(cleaned, "’", "'")

	// 3. Find boundaries of JSON object or array if wrapped in conversational filler
	startObj := strings.Index(cleaned, "{")
	startArr := strings.Index(cleaned, "[")

	if startObj != -1 || startArr != -1 {
		var startIdx, endIdx int
		if startObj != -1 && (startArr == -1 || startObj < startArr) {
			startIdx = startObj
			endIdx = strings.LastIndex(cleaned, "}")
		} else if startArr != -1 {
			startIdx = startArr
			endIdx = strings.LastIndex(cleaned, "]")
		}

		if startIdx != -1 && endIdx > startIdx {
			cleaned = cleaned[startIdx : endIdx+1]
		}
	}

	// 4. Remove trailing commas in objects or arrays: { "a": 1, } -> { "a": 1 }
	cleaned = reTrailingComma.ReplaceAllString(cleaned, "$1")

	// 5. Check if it is now valid JSON
	if json.Valid([]byte(cleaned)) {
		// Pretty-indent for canonical structure
		var indented interface{}
		if err := json.Unmarshal([]byte(cleaned), &indented); err == nil {
			if formatted, err := json.MarshalIndent(indented, "", "  "); err == nil {
				return string(formatted), true
			}
		}
		return cleaned, true
	}

	// 6. Attempt trailing brace completion if closed bracket was truncated
	if strings.HasPrefix(cleaned, "{") && !strings.HasSuffix(cleaned, "}") {
		candidate := cleaned + "}"
		candidate = reTrailingComma.ReplaceAllString(candidate, "$1")
		if json.Valid([]byte(candidate)) {
			return candidate, true
		}
	}
	if strings.HasPrefix(cleaned, "[") && !strings.HasSuffix(cleaned, "]") {
		candidate := cleaned + "]"
		candidate = reTrailingComma.ReplaceAllString(candidate, "$1")
		if json.Valid([]byte(candidate)) {
			return candidate, true
		}
	}

	return raw, false
}

// CoerceCodeExtract retrieves raw executable source code by stripping formatting commentary and fences.
func CoerceCodeExtract(raw string) (string, bool) {
	trimmed := strings.TrimSpace(raw)
	reCodeBlock := regexp.MustCompile("(?s)```[a-zA-Z0-9_-]*\\s*\\n(.+?)\\n```")
	matches := reCodeBlock.FindAllStringSubmatch(trimmed, -1)

	if len(matches) > 0 {
		var blocks []string
		for _, m := range matches {
			if len(m) > 1 {
				blocks = append(blocks, strings.TrimSpace(m[1]))
			}
		}
		return strings.Join(blocks, "\n\n"), true
	}

	return trimmed, false
}
