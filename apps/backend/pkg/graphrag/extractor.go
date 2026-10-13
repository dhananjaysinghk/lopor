package graphrag

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// GraphExtractor parses document text and synthesizes knowledge graph triplets.
type GraphExtractor struct {
	tripletRules []tripletRule
}

type tripletRule struct {
	regex        *regexp.Regexp
	relationType string
	defaultDesc  string
}

// NewGraphExtractor initializes the triplet extraction engine.
func NewGraphExtractor() *GraphExtractor {
	rules := []tripletRule{
		{
			regex:        regexp.MustCompile(`(?i)\b([A-Z][a-zA-Z0-9_\.\-]+(?:\s+[A-Z][a-zA-Z0-9_\.\-]+)?)\s+(?:depends on|relies on|requires)\s+([A-Z][a-zA-Z0-9_\.\-]+(?:\s+[A-Z][a-zA-Z0-9_\.\-]+)?)\b`),
			relationType: "DEPENDS_ON",
			defaultDesc:  "expresses structural dependency",
		},
		{
			regex:        regexp.MustCompile(`(?i)\b([A-Z][a-zA-Z0-9_\.\-]+(?:\s+[A-Z][a-zA-Z0-9_\.\-]+)?)\s+(?:implements|extends|conforms to)\s+([A-Z][a-zA-Z0-9_\.\-]+(?:\s+[A-Z][a-zA-Z0-9_\.\-]+)?)\b`),
			relationType: "IMPLEMENTS",
			defaultDesc:  "implements contract or abstraction",
		},
		{
			regex:        regexp.MustCompile(`(?i)\b([A-Z][a-zA-Z0-9_\.\-]+(?:\s+[A-Z][a-zA-Z0-9_\.\-]+)?)\s+(?:integrates with|connects to|communicates with|talks to)\s+([A-Z][a-zA-Z0-9_\.\-]+(?:\s+[A-Z][a-zA-Z0-9_\.\-]+)?)\b`),
			relationType: "INTEGRATES_WITH",
			defaultDesc:  "interfaces via protocol or API",
		},
		{
			regex:        regexp.MustCompile(`(?i)\b([A-Z][a-zA-Z0-9_\.\-]+(?:\s+[A-Z][a-zA-Z0-9_\.\-]+)?)\s+(?:stores data in|persists to|saves records in)\s+([A-Z][a-zA-Z0-9_\.\-]+(?:\s+[A-Z][a-zA-Z0-9_\.\-]+)?)\b`),
			relationType: "STORES_DATA_IN",
			defaultDesc:  "persists application state",
		},
		{
			regex:        regexp.MustCompile(`(?i)\b([A-Z][a-zA-Z0-9_\.\-]+(?:\s+[A-Z][a-zA-Z0-9_\.\-]+)?)\s+(?:uses|utilizes|employs)\s+([A-Z][a-zA-Z0-9_\.\-]+(?:\s+[A-Z][a-zA-Z0-9_\.\-]+)?)\b`),
			relationType: "USES",
			defaultDesc:  "utilizes component functionality",
		},
		{
			regex:        regexp.MustCompile(`(?i)\b([A-Z][a-zA-Z0-9_\.\-]+(?:\s+[A-Z][a-zA-Z0-9_\.\-]+)?)\s+(?:is part of|belongs to|resides within)\s+([A-Z][a-zA-Z0-9_\.\-]+(?:\s+[A-Z][a-zA-Z0-9_\.\-]+)?)\b`),
			relationType: "PART_OF",
			defaultDesc:  "compositional membership",
		},
	}

	return &GraphExtractor{tripletRules: rules}
}

// ExtractTriplets processes text content and extracts entity nodes and relational edges.
func (ge *GraphExtractor) ExtractTriplets(workspaceID, docID, content string) ([]EntityNode, []RelationshipEdge) {
	nodeMap := make(map[string]EntityNode)
	edgeMap := make(map[string]RelationshipEdge)
	now := time.Now().UTC()

	// Normalize whitespace
	normalized := strings.ReplaceAll(content, "\r\n", "\n")

	for _, rule := range ge.tripletRules {
		matches := rule.regex.FindAllStringSubmatch(normalized, -1)
		for _, m := range matches {
			if len(m) < 3 {
				continue
			}

			srcRaw := strings.TrimSpace(m[1])
			tgtRaw := strings.TrimSpace(m[2])

			// Skip trivial matches or stop words
			if isStopEntity(srcRaw) || isStopEntity(tgtRaw) || strings.EqualFold(srcRaw, tgtRaw) {
				continue
			}

			srcID := hashID(workspaceID, srcRaw)
			tgtID := hashID(workspaceID, tgtRaw)

			if _, exists := nodeMap[srcID]; !exists {
				nodeMap[srcID] = EntityNode{
					ID:          srcID,
					WorkspaceID: workspaceID,
					Name:        srcRaw,
					Type:        inferEntityType(srcRaw),
					Description: fmt.Sprintf("Entity %s extracted from document knowledge context", srcRaw),
					DocumentID:  docID,
					CreatedAt:   now,
				}
			}

			if _, exists := nodeMap[tgtID]; !exists {
				nodeMap[tgtID] = EntityNode{
					ID:          tgtID,
					WorkspaceID: workspaceID,
					Name:        tgtRaw,
					Type:        inferEntityType(tgtRaw),
					Description: fmt.Sprintf("Entity %s extracted from document knowledge context", tgtRaw),
					DocumentID:  docID,
					CreatedAt:   now,
				}
			}

			edgeKey := fmt.Sprintf("%s-%s-%s", srcID, rule.relationType, tgtID)
			if _, exists := edgeMap[edgeKey]; !exists {
				edgeMap[edgeKey] = RelationshipEdge{
					ID:           edgeKey,
					WorkspaceID:  workspaceID,
					SourceID:     srcID,
					SourceName:   srcRaw,
					TargetID:     tgtID,
					TargetName:   tgtRaw,
					RelationType: rule.relationType,
					Weight:       1.0,
					Description:  rule.defaultDesc,
					CreatedAt:    now,
				}
			}
		}
	}

	nodes := make([]EntityNode, 0, len(nodeMap))
	for _, n := range nodeMap {
		nodes = append(nodes, n)
	}

	edges := make([]RelationshipEdge, 0, len(edgeMap))
	for _, e := range edgeMap {
		edges = append(edges, e)
	}

	return nodes, edges
}

func inferEntityType(name string) EntityType {
	lower := strings.ToLower(name)
	techKeywords := []string{
		"postgres", "postgresql", "pgvector", "redis", "next.js", "react", "golang", "go",
		"fiber", "docker", "kubernetes", "kafka", "elasticsearch", "ollama", "openai",
		"bedrock", "api", "database", "sql", "jwt", "totp", "sdk", "cli",
	}
	for _, kw := range techKeywords {
		if strings.Contains(lower, kw) {
			return EntityTechnology
		}
	}

	componentKeywords := []string{"service", "handler", "router", "engine", "controller", "pipeline", "module", "worker"}
	for _, kw := range componentKeywords {
		if strings.Contains(lower, kw) {
			return EntityComponent
		}
	}

	orgKeywords := []string{"team", "corp", "inc", "foundation", "enterprise", "group", "labs"}
	for _, kw := range orgKeywords {
		if strings.Contains(lower, kw) {
			return EntityOrganization
		}
	}

	return EntityConcept
}

func isStopEntity(val string) bool {
	lower := strings.ToLower(val)
	stops := map[string]bool{
		"the": true, "this": true, "that": true, "each": true, "some": true,
		"all": true, "any": true, "our": true, "your": true, "their": true,
		"it": true, "here": true, "there": true, "what": true, "which": true,
	}
	return stops[lower] || len(val) < 2
}

func hashID(workspaceID, name string) string {
	h := sha256.Sum256([]byte(workspaceID + ":" + strings.ToLower(name)))
	return "ent_" + hex.EncodeToString(h[:8])
}
