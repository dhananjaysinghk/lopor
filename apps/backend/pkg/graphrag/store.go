package graphrag

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"
)

// GraphStore maintains knowledge graph state and provides topological traversal.
type GraphStore struct {
	mu          sync.RWMutex
	nodes       map[string]EntityNode
	edges       map[string]RelationshipEdge
	outEdges    map[string][]string           // nodeID -> list of edgeIDs
	inEdges     map[string][]string           // nodeID -> list of edgeIDs
	wsNodes     map[string][]string           // wsID -> list of nodeIDs
	wsEdges     map[string][]string           // wsID -> list of edgeIDs
	communities map[string][]CommunityCluster // wsID -> list of clusters
}

// NewGraphStore initializes the graph storage engine.
func NewGraphStore() *GraphStore {
	return &GraphStore{
		nodes:       make(map[string]EntityNode),
		edges:       make(map[string]RelationshipEdge),
		outEdges:    make(map[string][]string),
		inEdges:     make(map[string][]string),
		wsNodes:     make(map[string][]string),
		wsEdges:     make(map[string][]string),
		communities: make(map[string][]CommunityCluster),
	}
}

// UpsertNode saves or updates an entity node.
func (s *GraphStore) UpsertNode(node EntityNode) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.nodes[node.ID]; !exists {
		s.wsNodes[node.WorkspaceID] = append(s.wsNodes[node.WorkspaceID], node.ID)
	}
	s.nodes[node.ID] = node
}

// UpsertEdge saves or updates a relationship edge.
func (s *GraphStore) UpsertEdge(edge RelationshipEdge) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.edges[edge.ID]; !exists {
		s.wsEdges[edge.WorkspaceID] = append(s.wsEdges[edge.WorkspaceID], edge.ID)
		s.outEdges[edge.SourceID] = append(s.outEdges[edge.SourceID], edge.ID)
		s.inEdges[edge.TargetID] = append(s.inEdges[edge.TargetID], edge.ID)

		// Increment node degrees
		if src, ok := s.nodes[edge.SourceID]; ok {
			src.Degree++
			s.nodes[edge.SourceID] = src
		}
		if tgt, ok := s.nodes[edge.TargetID]; ok {
			tgt.Degree++
			s.nodes[edge.TargetID] = tgt
		}
	}
	s.edges[edge.ID] = edge
}

// GetNode retrieves an entity node by its ID.
func (s *GraphStore) GetNode(id string) (EntityNode, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	n, ok := s.nodes[id]
	return n, ok
}

// FindNodesByName searches for nodes matching a string pattern in a workspace.
func (s *GraphStore) FindNodesByName(workspaceID, query string) []EntityNode {
	s.mu.RLock()
	defer s.mu.RUnlock()

	q := strings.ToLower(query)
	var matches []EntityNode
	for _, nID := range s.wsNodes[workspaceID] {
		node := s.nodes[nID]
		if strings.Contains(strings.ToLower(node.Name), q) || strings.Contains(strings.ToLower(node.Description), q) {
			matches = append(matches, node)
		}
	}
	return matches
}

// TraverseNeighbors conducts BFS neighborhood exploration starting from a seed node.
func (s *GraphStore) TraverseNeighbors(workspaceID, seedID string, maxDepth int) ([]EntityNode, []RelationshipEdge) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if maxDepth <= 0 {
		maxDepth = 1
	}

	visitedNodes := make(map[string]bool)
	visitedEdges := make(map[string]bool)

	queue := []string{seedID}
	visitedNodes[seedID] = true
	currentDepth := 0

	for len(queue) > 0 && currentDepth < maxDepth {
		levelSize := len(queue)
		for i := 0; i < levelSize; i++ {
			currNodeID := queue[0]
			queue = queue[1:]

			// Traverse outgoing edges
			for _, eID := range s.outEdges[currNodeID] {
				edge := s.edges[eID]
				visitedEdges[eID] = true
				if !visitedNodes[edge.TargetID] {
					visitedNodes[edge.TargetID] = true
					queue = append(queue, edge.TargetID)
				}
			}

			// Traverse incoming edges
			for _, eID := range s.inEdges[currNodeID] {
				edge := s.edges[eID]
				visitedEdges[eID] = true
				if !visitedNodes[edge.SourceID] {
					visitedNodes[edge.SourceID] = true
					queue = append(queue, edge.SourceID)
				}
			}
		}
		currentDepth++
	}

	var resultNodes []EntityNode
	for nID := range visitedNodes {
		if node, ok := s.nodes[nID]; ok {
			resultNodes = append(resultNodes, node)
		}
	}

	var resultEdges []RelationshipEdge
	for eID := range visitedEdges {
		if edge, ok := s.edges[eID]; ok {
			resultEdges = append(resultEdges, edge)
		}
	}

	return resultNodes, resultEdges
}

// DetectCommunities clusters connected graph partitions and generates thematic summaries.
func (s *GraphStore) DetectCommunities(workspaceID string) []CommunityCluster {
	s.mu.Lock()
	defer s.mu.Unlock()

	nodeIDs := s.wsNodes[workspaceID]
	if len(nodeIDs) == 0 {
		return nil
	}

	// Connected components clustering algorithm
	visited := make(map[string]bool)
	var components [][]string

	for _, nID := range nodeIDs {
		if visited[nID] {
			continue
		}

		var comp []string
		q := []string{nID}
		visited[nID] = true

		for len(q) > 0 {
			curr := q[0]
			q = q[1:]
			comp = append(comp, curr)

			// Adjacent nodes
			for _, eID := range s.outEdges[curr] {
				edge := s.edges[eID]
				if !visited[edge.TargetID] {
					visited[edge.TargetID] = true
					q = append(q, edge.TargetID)
				}
			}
			for _, eID := range s.inEdges[curr] {
				edge := s.edges[eID]
				if !visited[edge.SourceID] {
					visited[edge.SourceID] = true
					q = append(q, edge.SourceID)
				}
			}
		}
		components = append(components, comp)
	}

	var clusters []CommunityCluster
	for idx, comp := range components {
		cID := fmt.Sprintf("comm_%s_%d", workspaceID, idx+1)
		var names []string
		for _, nID := range comp {
			node := s.nodes[nID]
			node.CommunityID = cID
			s.nodes[nID] = node
			names = append(names, node.Name)
		}

		sort.Strings(names)
		keyThemes := names
		if len(keyThemes) > 5 {
			keyThemes = keyThemes[:5]
		}

		summary := fmt.Sprintf("Community %d synthesizes knowledge regarding %s, comprising %d interconnected architecture entities.",
			idx+1, strings.Join(keyThemes, ", "), len(comp))

		density := 1.0
		if len(comp) > 1 {
			possibleEdges := float64(len(comp) * (len(comp) - 1))
			density = math.Min(1.0, float64(len(comp))/possibleEdges)
		}

		cluster := CommunityCluster{
			ID:            cID,
			WorkspaceID:   workspaceID,
			Level:         1,
			Name:          fmt.Sprintf("Cluster %d: %s", idx+1, strings.Join(keyThemes, "-")),
			Summary:       summary,
			MemberNodeIDs: comp,
			KeyThemes:     keyThemes,
			Density:       math.Round(density*100) / 100,
			CreatedAt:     time.Now().UTC(),
		}
		clusters = append(clusters, cluster)
	}

	s.communities[workspaceID] = clusters
	return clusters
}

// GetCommunities returns previously calculated community clusters.
func (s *GraphStore) GetCommunities(workspaceID string) []CommunityCluster {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.communities[workspaceID]
}

// GetVisualization formats graph topology for visual renderers.
func (s *GraphStore) GetVisualization(workspaceID string) GraphVisualizationData {
	s.mu.RLock()
	defer s.mu.RUnlock()

	nodeIDs := s.wsNodes[workspaceID]
	edgeIDs := s.wsEdges[workspaceID]

	var vNodes []VisualNode
	for _, nID := range nodeIDs {
		n := s.nodes[nID]
		vNodes = append(vNodes, VisualNode{
			ID:          n.ID,
			Label:       n.Name,
			Type:        n.Type,
			Degree:      n.Degree,
			CommunityID: n.CommunityID,
		})
	}

	var vLinks []VisualLink
	for _, eID := range edgeIDs {
		e := s.edges[eID]
		vLinks = append(vLinks, VisualLink{
			ID:     e.ID,
			Source: e.SourceID,
			Target: e.TargetID,
			Label:  e.RelationType,
			Weight: e.Weight,
		})
	}

	return GraphVisualizationData{
		WorkspaceID: workspaceID,
		Nodes:       vNodes,
		Links:       vLinks,
		Communities: s.communities[workspaceID],
		TotalNodes:  len(vNodes),
		TotalEdges:  len(vLinks),
	}
}
