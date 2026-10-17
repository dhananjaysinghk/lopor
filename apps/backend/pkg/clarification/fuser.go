package clarification

import (
	"fmt"
	"strings"
)

// Fuser merges user clarification answers into a canonical disambiguated prompt.
type Fuser struct{}

// NewFuser initializes the prompt fusion engine.
func NewFuser() *Fuser {
	return &Fuser{}
}

// Fuse combines the original query with clarification answers into an execution-ready prompt.
func (f *Fuser) Fuse(req FusionRequest) *FusionResult {
	params := make(map[string]string)
	var additions []string

	for _, ans := range req.Answers {
		val := strings.TrimSpace(ans.SelectedValue)
		if ans.CustomInput != "" {
			val = strings.TrimSpace(ans.CustomInput)
		}
		if val == "" {
			continue
		}

		params[ans.ParameterKey] = val

		switch ans.ParameterKey {
		case "target_environment":
			additions = append(additions, fmt.Sprintf("target environment: %s", val))
		case "export_format":
			additions = append(additions, fmt.Sprintf("format: %s", val))
		case "test_type":
			additions = append(additions, fmt.Sprintf("test type: %s", val))
		case "target_component":
			additions = append(additions, fmt.Sprintf("target component: %s", val))
		default:
			additions = append(additions, fmt.Sprintf("%s: %s", ans.ParameterKey, val))
		}
	}

	trimmedOriginal := strings.TrimRight(strings.TrimSpace(req.OriginalQuery), ".!?;,")
	var resolved string
	if len(additions) > 0 {
		resolved = fmt.Sprintf("%s (Clarifications: %s).", trimmedOriginal, strings.Join(additions, ", "))
	} else {
		resolved = trimmedOriginal + "."
	}

	summary := fmt.Sprintf("Resolved %d query parameter(s) through clarification.", len(params))

	return &FusionResult{
		ResolvedPrompt:     resolved,
		ResolvedParameters: params,
		FusedSummary:       summary,
	}
}
