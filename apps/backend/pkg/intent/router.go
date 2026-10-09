package intent

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"
)

// IntentCategory represents the functional domain capability targeted by a prompt.
type IntentCategory string

const (
	IntentCodeDev        IntentCategory = "CODE_ENGINEERING"
	IntentRAGSearch      IntentCategory = "KNOWLEDGE_SEARCH"
	IntentSQLSynth       IntentCategory = "DATABASE_QUERY"
	IntentConsensus      IntentCategory = "STRATEGIC_DEBATE"
	IntentVoiceAudio     IntentCategory = "VOICE_NARRATION"
	IntentDocumentExport IntentCategory = "DOCUMENT_WORKFLOW"
	IntentIncidentRCA    IntentCategory = "INCIDENT_POSTMORTEM"
	IntentGeneralChat    IntentCategory = "CONVERSATIONAL_CHAT"
)

// IntentRoute defines a registered destination with exemplar utterances for classification.
type IntentRoute struct {
	ID                  string         `json:"id"`
	Category            IntentCategory `json:"category"`
	Name                string         `json:"name"`
	Description         string         `json:"description"`
	Exemplars           []string       `json:"exemplars"`
	DestinationEndpoint string         `json:"destination_endpoint"`
	ConfidenceThreshold float64        `json:"confidence_threshold"`
}

// ClassificationRequest defines the query input to be semantically routed.
type ClassificationRequest struct {
	Prompt      string `json:"prompt"`
	WorkspaceID string `json:"workspace_id,omitempty"`
	TopK        int    `json:"top_k,omitempty"` // Default: 2
}

// CandidateMatch represents a candidate intent with computed confidence.
type CandidateMatch struct {
	IntentID            string         `json:"intent_id"`
	Category            IntentCategory `json:"category"`
	Name                string         `json:"name"`
	Confidence          float64        `json:"confidence"` // 0.0 to 1.0
	MatchedExemplar     string         `json:"matched_exemplar,omitempty"`
	DestinationEndpoint string         `json:"destination_endpoint"`
}

// ClassificationResult details the final routing decision and candidate intents.
type ClassificationResult struct {
	Prompt            string           `json:"prompt"`
	PrimaryIntent     CandidateMatch   `json:"primary_intent"`
	SecondaryIntents  []CandidateMatch `json:"secondary_intents"`
	IsConfident       bool             `json:"is_confident"`
	ExtractedEntities []string         `json:"extracted_entities,omitempty"`
	RecommendedAction string           `json:"recommended_action"`
	DurationMs        int64            `json:"duration_ms"`
	ClassifiedAt      string           `json:"classified_at"`
}

// SemanticRouter coordinates zero-shot and exemplar-based intent classification.
type SemanticRouter struct {
	mu     sync.RWMutex
	routes map[string]IntentRoute
}

// NewSemanticRouter initializes the semantic router with core platform capability routes.
func NewSemanticRouter() *SemanticRouter {
	routes := map[string]IntentRoute{
		"route-code": {
			ID:                  "route-code",
			Category:            IntentCodeDev,
			Name:                "Code Engineering & Sandbox",
			Description:         "Software development, bug fixes, refactoring, code execution, git diffs, unit testing.",
			Exemplars:           []string{"fix bug in code", "refactor function", "execute script in sandbox", "create git diff patch", "generate unit tests", "scan code for vulnerabilities"},
			DestinationEndpoint: "/api/v1/workspaces/:wsId/sandbox/execute",
			ConfidenceThreshold: 0.60,
		},
		"route-rag": {
			ID:                  "route-rag",
			Category:            IntentRAGSearch,
			Name:                "Knowledge Search & RAG",
			Description:         "Document retrieval, semantic search across workspace knowledge base, web grounding.",
			Exemplars:           []string{"find in documents", "search knowledge base", "what does the rfc say", "ground query with web", "retrieve policy guidelines", "search vector index"},
			DestinationEndpoint: "/api/v1/workspaces/:wsId/search/hybrid",
			ConfidenceThreshold: 0.60,
		},
		"route-sql": {
			ID:                  "route-sql",
			Category:            IntentSQLSynth,
			Name:                "Text-to-SQL Query Synthesizer",
			Description:         "Natural language database queries, schema inspections, postgres analytics.",
			Exemplars:           []string{"generate sql query", "how many users signed up", "select rows where active", "table schema for orders", "database analytics aggregate"},
			DestinationEndpoint: "/api/v1/workspaces/:wsId/sql/synthesize",
			ConfidenceThreshold: 0.65,
		},
		"route-consensus": {
			ID:                  "route-consensus",
			Category:            IntentConsensus,
			Name:                "Multi-Agent Consensus Debate",
			Description:         "High-stakes architectural decisions, trade-off debates, devil's advocate stress tests.",
			Exemplars:           []string{"debate this architecture", "multi-agent consensus review", "evaluate trade-offs security vs performance", "devils advocate challenge", "reach team consensus"},
			DestinationEndpoint: "/api/v1/workspaces/:wsId/agents/consensus/debate",
			ConfidenceThreshold: 0.65,
		},
		"route-voice": {
			ID:                  "route-voice",
			Category:            IntentVoiceAudio,
			Name:                "Neural Voice & TTS Narration",
			Description:         "Read document aloud, synthesize speech audio, audio podcast brief.",
			Exemplars:           []string{"read this document aloud", "synthesize text to speech", "generate audio podcast brief", "narrate with voice", "audio dictation"},
			DestinationEndpoint: "/api/v1/workspaces/:wsId/voice/tts/synthesize",
			ConfidenceThreshold: 0.65,
		},
		"route-doc": {
			ID:                  "route-doc",
			Category:            IntentDocumentExport,
			Name:                "Document Workflow & PDF Export",
			Description:         "Export document to PDF or HTML, auto-summarize long text, translate document.",
			Exemplars:           []string{"export to pdf", "convert document to html", "summarize document", "translate to spanish", "deduplicate documents"},
			DestinationEndpoint: "/api/v1/workspaces/:wsId/documents/:docId/export",
			ConfidenceThreshold: 0.60,
		},
		"route-rca": {
			ID:                  "route-rca",
			Category:            IntentIncidentRCA,
			Name:                "Incident Postmortem & RCA",
			Description:         "Generate blameless postmortem, root cause analysis from logs, five whys.",
			Exemplars:           []string{"generate postmortem report", "analyze incident root cause", "five whys investigation", "outage error trace rca", "production incident timeline"},
			DestinationEndpoint: "/api/v1/workspaces/:wsId/incidents/generate-postmortem",
			ConfidenceThreshold: 0.65,
		},
	}

	return &SemanticRouter{routes: routes}
}

// GetRoutes returns all active intent routes.
func (r *SemanticRouter) GetRoutes() []IntentRoute {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var list []IntentRoute
	for _, route := range r.routes {
		list = append(list, route)
	}
	return list
}

// RegisterCustomRoute adds or overrides an intent route.
func (r *SemanticRouter) RegisterCustomRoute(route IntentRoute) error {
	if route.ID == "" || route.Name == "" || len(route.Exemplars) == 0 {
		return fmt.Errorf("id, name, and at least one exemplar are required")
	}
	if route.ConfidenceThreshold <= 0.0 {
		route.ConfidenceThreshold = 0.60
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.routes[route.ID] = route
	return nil
}

// ClassifyPrompt evaluates a prompt and returns the ranked matching intent routes.
func (r *SemanticRouter) ClassifyPrompt(ctx context.Context, req ClassificationRequest) (*ClassificationResult, error) {
	start := time.Now()

	prompt := strings.TrimSpace(req.Prompt)
	if prompt == "" {
		return nil, fmt.Errorf("prompt is required for intent classification")
	}

	topK := req.TopK
	if topK <= 0 {
		topK = 2
	}

	r.mu.RLock()
	var matches []CandidateMatch

	promptTokens := tokenize(prompt)

	for _, route := range r.routes {
		bestSim := 0.0
		var bestExemplar string

		for _, ex := range route.Exemplars {
			exTokens := tokenize(ex)
			sim := computeTokenJaccard(promptTokens, exTokens)
			if sim > bestSim {
				bestSim = sim
				bestExemplar = ex
			}
		}

		// Also check description keyword overlaps
		descTokens := tokenize(route.Description)
		descSim := computeTokenJaccard(promptTokens, descTokens) * 0.75
		if descSim > bestSim {
			bestSim = descSim
		}

		// Scale non-zero matches with confidence baseline
		if bestSim > 0.05 {
			conf := math.Round(math.Min(0.99, bestSim*2.5)*100) / 100
			matches = append(matches, CandidateMatch{
				IntentID:            route.ID,
				Category:            route.Category,
				Name:                route.Name,
				Confidence:          conf,
				MatchedExemplar:     bestExemplar,
				DestinationEndpoint: route.DestinationEndpoint,
			})
		}
	}
	r.mu.RUnlock()

	// Sort matches descending by confidence
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Confidence > matches[j].Confidence
	})

	var primary CandidateMatch
	var secondaries []CandidateMatch
	isConfident := false

	if len(matches) > 0 {
		primary = matches[0]
		isConfident = primary.Confidence >= 0.50
		for i := 1; i < len(matches) && i < topK+1; i++ {
			secondaries = append(secondaries, matches[i])
		}
	} else {
		primary = CandidateMatch{
			IntentID:            "route-general",
			Category:            IntentGeneralChat,
			Name:                "General Assistant & Multi-Turn Chat",
			Confidence:          0.50,
			DestinationEndpoint: "/api/v1/workspaces/:wsId/chats/:chatId/stream",
		}
		isConfident = true
	}

	entities := extractSimpleEntities(prompt)
	recAction := fmt.Sprintf("Route execution to %s (%s)", primary.Name, primary.DestinationEndpoint)

	return &ClassificationResult{
		Prompt:            prompt,
		PrimaryIntent:     primary,
		SecondaryIntents:  secondaries,
		IsConfident:       isConfident,
		ExtractedEntities: entities,
		RecommendedAction: recAction,
		DurationMs:        time.Since(start).Milliseconds(),
		ClassifiedAt:      time.Now().Format(time.RFC3339),
	}, nil
}

func tokenize(text string) map[string]bool {
	tokens := make(map[string]bool)
	clean := strings.ToLower(text)
	replacer := strings.NewReplacer(".", " ", ",", " ", "!", " ", "?", " ", ";", " ", ":", " ", "(", " ", ")", " ", "-", " ")
	clean = replacer.Replace(clean)
	for _, word := range strings.Fields(clean) {
		if len(word) > 2 {
			tokens[word] = true
		}
	}
	return tokens
}

func computeTokenJaccard(t1, t2 map[string]bool) float64 {
	if len(t1) == 0 || len(t2) == 0 {
		return 0.0
	}
	intersection := 0
	for k := range t1 {
		if t2[k] {
			intersection++
		}
	}
	union := len(t1) + len(t2) - intersection
	if union == 0 {
		return 0.0
	}
	return float64(intersection) / float64(union)
}

func extractSimpleEntities(text string) []string {
	var entities []string
	words := strings.Fields(text)
	for _, w := range words {
		clean := strings.Trim(w, ",.?!():\"'")
		if len(clean) > 3 && strings.ToUpper(clean[:1]) == clean[:1] && strings.ToLower(clean[1:]) == clean[1:] {
			entities = append(entities, clean)
		}
	}
	return entities
}
