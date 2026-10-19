package chunking

import (
	"context"
	"errors"
	"math"
	"strings"
	"time"
)

// ChunkingEngine coordinates semantic document parsing, section partitioning, and enriched chunk synthesis.
type ChunkingEngine struct {
	parser  *Parser
	chunker *StructuralChunker
}

// NewChunkingEngine initializes a new structural chunking engine.
func NewChunkingEngine() *ChunkingEngine {
	return &ChunkingEngine{
		parser:  NewParser(),
		chunker: NewStructuralChunker(),
	}
}

// ChunkDocument executes content-aware chunking on raw document text.
func (e *ChunkingEngine) ChunkDocument(ctx context.Context, req ChunkingRequest) (*ChunkingResult, error) {
	if strings.TrimSpace(req.RawContent) == "" {
		return nil, errors.New("raw content cannot be empty")
	}

	detectedFormat := e.parser.DetectFormat(req.RawContent)
	strategy := req.Options.Strategy
	if strategy == "" || strategy == StrategyAuto {
		switch detectedFormat {
		case FormatCode:
			strategy = StrategyCodeAware
		case FormatMarkdown:
			strategy = StrategyMarkdownHierarchical
		case FormatTable:
			strategy = StrategyTablePreserving
		default:
			strategy = StrategySemanticSentence
		}
	}

	var sections []DocumentSection
	switch strategy {
	case StrategyMarkdownHierarchical, StrategyTablePreserving:
		sections = e.parser.ParseMarkdownSections(req.Title, req.RawContent)
	case StrategyCodeAware:
		sections = e.parser.ParseCodeSections(req.RawContent)
	default:
		// Prose fallback: single top-level section
		sections = []DocumentSection{
			{
				Title:          req.Title,
				BreadcrumbPath: req.Title,
				Content:        req.RawContent,
				Format:         FormatProse,
			},
		}
	}

	chunks := e.chunker.ChunkSections(req.DocumentID, req.WorkspaceID, sections, req.Options)

	totalTokens := 0
	for _, c := range chunks {
		totalTokens += c.TokenCount
	}

	avgTokens := 0.0
	if len(chunks) > 0 {
		avgTokens = math.Round((float64(totalTokens)/float64(len(chunks)))*10) / 10
	}

	return &ChunkingResult{
		DocumentID:         req.DocumentID,
		FormatDetected:     detectedFormat,
		StrategyUsed:       strategy,
		TotalChunks:        len(chunks),
		TotalTokens:        totalTokens,
		AverageChunkTokens: avgTokens,
		Chunks:             chunks,
		ProcessedAt:        time.Now().UTC(),
	}, nil
}
