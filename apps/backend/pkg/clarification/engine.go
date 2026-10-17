package clarification

import (
	"context"
	"errors"
	"strings"
)

// ClarificationEngine coordinates conversational query disambiguation, entropy analysis, and prompt context fusion.
type ClarificationEngine struct {
	analyzer *Analyzer
	fuser    *Fuser
}

// NewClarificationEngine initializes the query clarification engine.
func NewClarificationEngine() *ClarificationEngine {
	return &ClarificationEngine{
		analyzer: NewAnalyzer(),
		fuser:    NewFuser(),
	}
}

// AnalyzeQuery evaluates a user query for ambiguity and synthesizes clarification questions if needed.
func (e *ClarificationEngine) AnalyzeQuery(ctx context.Context, query string, workspaceContext string) (*DisambiguationAnalysis, error) {
	if strings.TrimSpace(query) == "" {
		return nil, errors.New("query string cannot be empty")
	}

	analysis := e.analyzer.Analyze(query, workspaceContext)
	return analysis, nil
}

// FuseClarification merges clarification answers into an enriched, disambiguated prompt.
func (e *ClarificationEngine) FuseClarification(ctx context.Context, req FusionRequest) (*FusionResult, error) {
	if strings.TrimSpace(req.OriginalQuery) == "" {
		return nil, errors.New("original query cannot be empty")
	}

	result := e.fuser.Fuse(req)
	return result, nil
}
