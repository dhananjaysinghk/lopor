package datagen

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"
)

// QuestionType classifies the cognitive complexity and evaluation intent of an eval item.
type QuestionType string

const (
	TypeFactoid  QuestionType = "FACTUAL_RECALL"
	TypeMultiHop QuestionType = "MULTI_HOP_REASONING"
	TypeSummary  QuestionType = "SUMMARIZATION"
	TypeNegative QuestionType = "NEGATIVE_REFUSAL" // Evaluates whether the model hallucinates or properly refuses
)

// DatasetItem represents an individual ground-truth benchmark evaluation unit.
type DatasetItem struct {
	ID            string       `json:"id"`
	Question      string       `json:"question"`
	Type          QuestionType `json:"type"`
	ContextSource string       `json:"context_source"`
	GroundTruth   string       `json:"ground_truth"`
	Difficulty    string       `json:"difficulty"` // EASY, MEDIUM, HARD
	Confidence    float64      `json:"confidence"`
}

// DatasetGenerationRequest defines parameters for synthesizing a benchmark dataset.
type DatasetGenerationRequest struct {
	DocumentID       string         `json:"document_id,omitempty"`
	DocumentTitle    string         `json:"document_title"`
	Content          string         `json:"content"`
	Count            int            `json:"count,omitempty"` // Default: 5, Max: 20
	IncludeNegatives bool           `json:"include_negatives"`
	TargetTypes      []QuestionType `json:"target_types,omitempty"`
}

// EvaluationDataset bundles synthesized ground-truth items with statistical distributions.
type EvaluationDataset struct {
	DatasetID        string               `json:"dataset_id"`
	Title            string               `json:"title"`
	DocumentTitle    string               `json:"document_title"`
	TotalItems       int                  `json:"total_items"`
	ItemDistribution map[QuestionType]int `json:"item_distribution"`
	Items            []DatasetItem        `json:"items"`
	DiversityScore   float64              `json:"diversity_score"` // 0.0 to 1.0
	JSONLExport      string               `json:"jsonl_export"`
	GeneratedAt      string               `json:"generated_at"`
}

// DatasetGenerator coordinates synthetic eval dataset generation from enterprise documents.
type DatasetGenerator struct{}

// NewDatasetGenerator initializes the synthetic benchmark dataset generator.
func NewDatasetGenerator() *DatasetGenerator {
	return &DatasetGenerator{}
}

// GenerateDataset synthesizes ground-truth Q&A evaluation units from document text.
func (g *DatasetGenerator) GenerateDataset(ctx context.Context, req DatasetGenerationRequest) (*EvaluationDataset, error) {
	content := strings.TrimSpace(req.Content)
	if content == "" {
		return nil, fmt.Errorf("content is required for dataset generation")
	}

	docTitle := strings.TrimSpace(req.DocumentTitle)
	if docTitle == "" {
		docTitle = "Knowledge Document"
	}

	count := req.Count
	if count <= 0 {
		count = 5
	}
	if count > 20 {
		count = 20
	}

	sentences := splitSentences(content)
	if len(sentences) == 0 {
		sentences = []string{content}
	}

	items := make([]DatasetItem, 0, count)
	dist := make(map[QuestionType]int)

	typesOrder := []QuestionType{TypeFactoid, TypeMultiHop, TypeSummary}
	if req.IncludeNegatives {
		typesOrder = append(typesOrder, TypeNegative)
	}

	for i := 0; i < count; i++ {
		qType := typesOrder[i%len(typesOrder)]
		anchorSentence := sentences[i%len(sentences)]

		var q, gt, diff string
		conf := 0.92

		switch qType {
		case TypeFactoid:
			q = fmt.Sprintf("What are the primary specifications described regarding '%s'?", docTitle)
			gt = anchorSentence
			diff = "EASY"
		case TypeMultiHop:
			q = fmt.Sprintf("How does '%s' interact with dependent systems based on the documentation?", docTitle)
			gt = fmt.Sprintf("According to %s: %s", docTitle, anchorSentence)
			diff = "MEDIUM"
		case TypeSummary:
			q = fmt.Sprintf("Summarize the architectural purpose and key takeaways of %s.", docTitle)
			gt = fmt.Sprintf("The document outlines %s with core focus: %s", docTitle, truncateSentence(anchorSentence, 120))
			diff = "MEDIUM"
		case TypeNegative:
			q = fmt.Sprintf("What is the exact pricing model and credit card billing policy for %s?", docTitle)
			gt = "The provided document does not contain billing, credit card, or commercial pricing terms."
			diff = "HARD"
			conf = 0.98
		}

		item := DatasetItem{
			ID:            fmt.Sprintf("eval-%s-%03d", req.DocumentID, i+1),
			Question:      q,
			Type:          qType,
			ContextSource: anchorSentence,
			GroundTruth:   gt,
			Difficulty:    diff,
			Confidence:    conf,
		}

		items = append(items, item)
		dist[qType]++
	}

	// Calculate diversity score based on unique types distribution
	distinctTypes := len(dist)
	diversity := math.Round((float64(distinctTypes)/float64(len(typesOrder)))*100) / 100

	jsonl := g.FormatAsJSONL(items)

	return &EvaluationDataset{
		DatasetID:        fmt.Sprintf("ds-%d", time.Now().UnixNano()%1000000),
		Title:            fmt.Sprintf("Benchmark Golden Eval: %s", docTitle),
		DocumentTitle:    docTitle,
		TotalItems:       len(items),
		ItemDistribution: dist,
		Items:            items,
		DiversityScore:   diversity,
		JSONLExport:      jsonl,
		GeneratedAt:      time.Now().Format(time.RFC3339),
	}, nil
}

// FormatAsJSONL serializes dataset items into newline-delimited JSON (JSONL).
func (g *DatasetGenerator) FormatAsJSONL(items []DatasetItem) string {
	var sb strings.Builder
	for _, item := range items {
		record := map[string]interface{}{
			"id":             item.ID,
			"question":       item.Question,
			"type":           item.Type,
			"context":        item.ContextSource,
			"ground_truth":   item.GroundTruth,
			"difficulty":     item.Difficulty,
			"expected_score": item.Confidence,
		}
		data, err := json.Marshal(record)
		if err == nil {
			sb.WriteString(string(data))
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

// FormatAsCSV serializes dataset items into standard comma-separated values.
func (g *DatasetGenerator) FormatAsCSV(items []DatasetItem) string {
	var sb strings.Builder
	sb.WriteString("id,type,difficulty,question,ground_truth\n")
	for _, item := range items {
		cleanQ := strings.ReplaceAll(item.Question, "\"", "\"\"")
		cleanGT := strings.ReplaceAll(item.GroundTruth, "\"", "\"\"")
		sb.WriteString(fmt.Sprintf("\"%s\",\"%s\",\"%s\",\"%s\",\"%s\"\n", item.ID, item.Type, item.Difficulty, cleanQ, cleanGT))
	}
	return sb.String()
}

func splitSentences(text string) []string {
	var sentences []string
	raw := strings.Split(text, "\n")
	for _, line := range raw {
		trimmed := strings.TrimSpace(line)
		if len(trimmed) > 15 && !strings.HasPrefix(trimmed, "#") && !strings.HasPrefix(trimmed, "```") {
			sentences = append(sentences, trimmed)
		}
	}
	return sentences
}

func truncateSentence(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
