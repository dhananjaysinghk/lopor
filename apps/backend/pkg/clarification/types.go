package clarification

import (
	"time"
)

// ClarificationDecision defines whether the platform should proceed, assume defaults, or ask questions.
type ClarificationDecision string

const (
	DecisionProceed                ClarificationDecision = "PROCEED"
	DecisionProceedWithAssumptions ClarificationDecision = "PROCEED_WITH_ASSUMPTIONS"
	DecisionNeedsClarification     ClarificationDecision = "NEEDS_CLARIFICATION"
)

// AmbiguityFactor represents a specific underspecified dimension in a query.
type AmbiguityFactor struct {
	Type        string  `json:"type"`        // e.g. "MISSING_ENVIRONMENT", "UNDERSPECIFIED_FORMAT", "AMBIGUOUS_PRONOUN"
	Description string  `json:"description"`
	Severity    float64 `json:"severity"`    // 0.0 to 1.0 weight
}

// ClarificationOption represents a selectable choice presented to the user.
type ClarificationOption struct {
	ID            string `json:"id"`
	Label         string `json:"label"`
	Value         string `json:"value"`
	Description   string `json:"description,omitempty"`
	IsRecommended bool   `json:"is_recommended"`
}

// ClarificationQuestion defines an interactive question to resolve query ambiguity.
type ClarificationQuestion struct {
	ID               string                `json:"id"`
	ParameterKey     string                `json:"parameter_key"` // e.g. "target_environment", "export_format"
	Question         string                `json:"question"`
	Options          []ClarificationOption `json:"options"`
	AllowCustomInput bool                  `json:"allow_custom_input"`
	IsMultiSelect    bool                  `json:"is_multi_select"`
}

// DisambiguationAnalysis details the ambiguity assessment and synthesized clarification prompts.
type DisambiguationAnalysis struct {
	OriginalQuery string                  `json:"original_query"`
	AmbiguityScore float64                 `json:"ambiguity_score"` // 0.0 (fully clear) to 1.0 (highly ambiguous)
	Decision      ClarificationDecision   `json:"decision"`
	Factors       []AmbiguityFactor       `json:"factors"`
	Assumptions   []string                `json:"assumptions,omitempty"`
	Questions     []ClarificationQuestion `json:"questions,omitempty"`
	AnalyzedAt    time.Time               `json:"analyzed_at"`
}

// ClarificationAnswer represents a user's selection or write-in for a specific question.
type ClarificationAnswer struct {
	QuestionID    string `json:"question_id"`
	ParameterKey  string `json:"parameter_key"`
	SelectedValue string `json:"selected_value"`
	CustomInput   string `json:"custom_input,omitempty"`
}

// FusionRequest captures the user's answers to be integrated into the original query.
type FusionRequest struct {
	OriginalQuery string                `json:"original_query"`
	Answers       []ClarificationAnswer `json:"answers"`
	Context       string                `json:"context,omitempty"`
}

// FusionResult provides the synthesized, fully disambiguated prompt ready for execution.
type FusionResult struct {
	ResolvedPrompt     string            `json:"resolved_prompt"`
	ResolvedParameters map[string]string `json:"resolved_parameters"`
	FusedSummary       string            `json:"fused_summary"`
}
