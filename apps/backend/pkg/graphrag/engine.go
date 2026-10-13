package graphrag

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
)

// GraphRAGEngine orchestrates knowledge graph extraction, community summarization, and hybrid graph searches.
type GraphRAGEngine struct {
	extractor *GraphExtractor
	store     *GraphStore
}

// NewGraphRAGEngine initializes a new GraphRAG engine instance.
func NewGraphRAGEngine() *GraphRAGEngine {
	return &GraphRAGEngine{
		extractor: NewGraphExtractor(),
		store:     NewGraphStore(),
	}
}

// GetStore returns the underlying graph storage engine.
func (e *GraphRAGEngine) GetStore() *GraphStore {
	return e.store
}

// ExtractAndIngest parses document text, stores knowledge triplets, and updates community partitions.
func (e *GraphRAGEngine) ExtractAndIngest(ctx context.Context, req GraphExtractionRequest) (*GraphExtractionResult, error) {
	nodes, edges := e.extractor.ExtractTriplets(req.WorkspaceID, req.DocumentID, req.Content)

	for _, n := range nodes {
		e.store.UpsertNode(n)
	}
	for _, edge := range edges {
		e.store.UpsertEdge(edge)
	}

	// Update community detection partitions
	e.store.DetectCommunities(req.WorkspaceID)

	return &GraphExtractionResult{
		ExtractedNodes: nodes,
		ExtractedEdges: edges,
		ExtractedCount: len(nodes) + len(edges),
	}, nil
}

// LocalSearch executes 1-hop or 2-hop neighborhood exploration centered around query entities.
func (e *GraphRAGEngine) LocalSearch(ctx context.Context, req LocalGraphSearchRequest) (*LocalGraphSearchResult, error) {
	depth := req.MaxDepth
	if depth <= 0 {
		depth = 2
	}

	var seedIDs []string

	if len(req.EntitySeeds) > 0 {
		for _, seedName := range req.EntitySeeds {
			matches := e.store.FindNodesByName(req.WorkspaceID, seedName)
			for _, m := range matches {
				seedIDs = append(seedIDs, m.ID)
			}
		}
	} else {
		// Discover seed entities matching the query
		matches := e.store.FindNodesByName(req.WorkspaceID, req.Query)
		for _, m := range matches {
			seedIDs = append(seedIDs, m.ID)
		}
	}

	if len(seedIDs) == 0 {
		return &LocalGraphSearchResult{
			Query:           req.Query,
			FocusedNodes:    nil,
			ConnectingEdges: nil,
			ContextSummary:  "No localized entity matches found in workspace knowledge graph for query.",
			Confidence:      0.2,
		}, nil
	}

	nodeMap := make(map[string]EntityNode)
	edgeMap := make(map[string]RelationshipEdge)

	for _, sID := range seedIDs {
		nodes, edges := e.store.TraverseNeighbors(req.WorkspaceID, sID, depth)
		for _, n := range nodes {
			nodeMap[n.ID] = n
		}
		for _, ed := range edges {
			edgeMap[ed.ID] = ed
		}
	}

	var focusedNodes []EntityNode
	for _, n := range nodeMap {
		focusedNodes = append(focusedNodes, n)
	}
	sort.Slice(focusedNodes, func(i, j int) bool {
		return focusedNodes[i].Degree > focusedNodes[j].Degree
	})

	var connectingEdges []RelationshipEdge
	for _, ed := range edgeMap {
		connectingEdges = append(connectingEdges, ed)
	}

	// Build context summary
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Local Knowledge Subgraph for '%s':\n", req.Query))
	sb.WriteString("Core Entities:\n")
	for idx, n := range focusedNodes {
		if idx >= 8 {
			break
		}
		sb.WriteString(fmt.Sprintf("- %s (%s, Degree: %d)\n", n.Name, n.Type, n.Degree))
	}
	sb.WriteString("\nRelational Paths:\n")
	for idx, ed := range connectingEdges {
		if idx >= 10 {
			break
		}
		sb.WriteString(fmt.Sprintf("- %s --[%s]--> %s (%s)\n", ed.SourceName, ed.RelationType, ed.TargetName, ed.Description))
	}

	confidence := math.Min(0.95, 0.50+float64(len(focusedNodes))*0.05+float64(len(connectingEdges))*0.03)

	return &LocalGraphSearchResult{
		Query:           req.Query,
		FocusedNodes:    focusedNodes,
		ConnectingEdges: connectingEdges,
		ContextSummary:  sb.String(),
		Confidence:      math.Round(confidence*100) / 100,
	}, nil
}

// GlobalSearch synthesizes holistic answers across hierarchical community cluster summaries.
func (e *GraphRAGEngine) GlobalSearch(ctx context.Context, req GlobalGraphSearchRequest) (*GlobalGraphSearchResult, error) {
	clusters := e.store.GetCommunities(req.WorkspaceID)
	if len(clusters) == 0 {
		clusters = e.store.DetectCommunities(req.WorkspaceID)
	}

	if len(clusters) == 0 {
		return &GlobalGraphSearchResult{
			Query:                 req.Query,
			SynthesizedAnswer:     "No knowledge communities exist in this workspace graph yet. Please ingest architecture documents first.",
			ReferencedCommunities: nil,
			RelevantThemes:        nil,
			Confidence:            0.1,
		}, nil
	}

	// Score clusters based on query terms
	qLower := strings.ToLower(req.Query)
	qWords := strings.Fields(qLower)

	type scoredCluster struct {
		cluster CommunityCluster
		score   float64
	}

	var scored []scoredCluster
	for _, c := range clusters {
		s := 0.0
		cText := strings.ToLower(c.Name + " " + c.Summary + " " + strings.Join(c.KeyThemes, " "))
		for _, w := range qWords {
			if strings.Contains(cText, w) {
				s += 1.0
			}
		}
		if s > 0 || len(qWords) == 0 {
			scored = append(scored, scoredCluster{cluster: c, score: s})
		}
	}

	// Fallback to top clusters if no direct word match
	if len(scored) == 0 {
		for _, c := range clusters {
			scored = append(scored, scoredCluster{cluster: c, score: 0.5})
		}
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	var selectedClusters []CommunityCluster
	themeSet := make(map[string]bool)
	var themes []string

	var answerBuilder strings.Builder
	answerBuilder.WriteString(fmt.Sprintf("Holistic Graph Synthesis for query '%s':\n\n", req.Query))

	for idx, sc := range scored {
		if idx >= 4 {
			break
		}
		selectedClusters = append(selectedClusters, sc.cluster)
		answerBuilder.WriteString(fmt.Sprintf("### %s\n%s\nKey Components: %s\n\n",
			sc.cluster.Name, sc.cluster.Summary, strings.Join(sc.cluster.KeyThemes, ", ")))

		for _, t := range sc.cluster.KeyThemes {
			if !themeSet[t] {
				themeSet[t] = true
				themes = append(themes, t)
			}
		}
	}

	confidence := math.Min(0.92, 0.60+float64(len(selectedClusters))*0.08)

	return &GlobalGraphSearchResult{
		Query:                 req.Query,
		SynthesizedAnswer:     answerBuilder.String(),
		ReferencedCommunities: selectedClusters,
		RelevantThemes:        themes,
		Confidence:            math.Round(confidence*100) / 100,
	}, nil
}

// GetGraphVisual generates visualization topology data for frontend renderers.
func (e *GraphRAGEngine) GetGraphVisual(ctx context.Context, workspaceID string) (*GraphVisualizationData, error) {
	data := e.store.GetVisualization(workspaceID)
	return &data, nil
}
