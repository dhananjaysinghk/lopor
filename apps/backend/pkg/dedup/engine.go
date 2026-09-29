package dedup

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math"
	"strings"
	"time"
)

// MatchCategory defines the degree of semantic and syntactic overlap between documents.
type MatchCategory string

const (
	CategoryExact      MatchCategory = "EXACT_DUPLICATE"
	CategoryNear       MatchCategory = "NEAR_DUPLICATE"
	CategoryDerivative MatchCategory = "REVISION_DERIVATIVE"
	CategoryUnique     MatchCategory = "UNIQUE"
)

// DocumentItem represents a document or text chunk to be evaluated for duplicates.
type DocumentItem struct {
	ID        string                 `json:"id"`
	Title     string                 `json:"title"`
	Content   string                 `json:"content"`
	CreatedAt string                 `json:"created_at,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// DuplicatePair represents a matched pair of similar documents.
type DuplicatePair struct {
	SourceID             string        `json:"source_id"`
	SourceTitle          string        `json:"source_title"`
	TargetID             string        `json:"target_id"`
	TargetTitle          string        `json:"target_title"`
	SimilarityScore      float64       `json:"similarity_score"` // 0.0 to 1.0
	Category             MatchCategory `json:"category"`
	SharedShingleCount   int           `json:"shared_shingle_count"`
	DifferenceHighlights []string      `json:"difference_highlights,omitempty"`
	Recommendation       string        `json:"recommendation"`
}

// ClusterGroup represents a canonical document and its affiliated duplicate variants.
type ClusterGroup struct {
	ClusterID            string         `json:"cluster_id"`
	CanonicalDocumentID  string         `json:"canonical_document_id"`
	CanonicalTitle       string         `json:"canonical_title"`
	DuplicateDocuments   []DocumentItem `json:"duplicate_documents"`
	AverageSimilarity    float64        `json:"average_similarity"`
	TotalSavingsBytes    int64          `json:"total_savings_bytes"`
	RecommendedAction    string         `json:"recommended_action"`
}

// DedupAnalysisRequest parameters for identifying and clustering duplicate content.
type DedupAnalysisRequest struct {
	Documents           []DocumentItem `json:"documents"`
	SimilarityThreshold float64        `json:"similarity_threshold,omitempty"` // Default: 0.80
	ShingleSize         int            `json:"shingle_size,omitempty"`          // Default: 3
}

// DedupReport summarizes duplicate findings, clusters, and savings.
type DedupReport struct {
	TotalAnalyzed         int             `json:"total_analyzed"`
	UniqueCount           int             `json:"unique_count"`
	DuplicateCount        int             `json:"duplicate_count"`
	DuplicateRatio        float64         `json:"duplicate_ratio"`
	Clusters              []ClusterGroup  `json:"clusters"`
	DuplicatePairs        []DuplicatePair `json:"duplicate_pairs"`
	EstimatedTokenSavings int             `json:"estimated_token_savings"`
	DurationMs            int64           `json:"duration_ms"`
	GeneratedAt           string          `json:"generated_at"`
}

// DedupEngine coordinates semantic shingling, Jaccard similarity, and document clustering.
type DedupEngine struct {
	defaultThreshold float64
	defaultShingle   int
}

// NewDedupEngine initializes the document deduplication engine.
func NewDedupEngine() *DedupEngine {
	return &DedupEngine{
		defaultThreshold: 0.80,
		defaultShingle:   3,
	}
}

// AnalyzeDuplicates processes a slice of documents and produces a comprehensive deduplication report.
func (e *DedupEngine) AnalyzeDuplicates(ctx context.Context, req DedupAnalysisRequest) (*DedupReport, error) {
	start := time.Now()

	docs := req.Documents
	if len(docs) == 0 {
		return nil, fmt.Errorf("at least one document is required for deduplication analysis")
	}

	threshold := req.SimilarityThreshold
	if threshold <= 0.0 || threshold > 1.0 {
		threshold = e.defaultThreshold
	}

	shingleSize := req.ShingleSize
	if shingleSize <= 0 {
		shingleSize = e.defaultShingle
	}

	// Step 1: Pre-process documents and extract normalized shingles & MD5 hashes
	type docProfile struct {
		doc       DocumentItem
		md5Hash   string
		shingles  map[string]bool
		wordCount int
		byteCount int64
	}

	profiles := make([]docProfile, len(docs))
	for i, d := range docs {
		words := tokenizeText(d.Content)
		shingles := buildShingles(words, shingleSize)
		hash := md5String(d.Content)
		profiles[i] = docProfile{
			doc:       d,
			md5Hash:   hash,
			shingles:  shingles,
			wordCount: len(words),
			byteCount: int64(len(d.Content)),
		}
	}

	var duplicatePairs []DuplicatePair
	clusterMap := make(map[int][]int) // canonicalIdx -> []duplicateIdx
	assignedToCluster := make(map[int]bool)

	// Step 2: Pairwise similarity computation
	for i := 0; i < len(profiles); i++ {
		for j := i + 1; j < len(profiles); j++ {
			p1 := profiles[i]
			p2 := profiles[j]

			var sim float64
			var category MatchCategory
			var sharedCount int

			if p1.md5Hash == p2.md5Hash {
				sim = 1.0
				category = CategoryExact
				sharedCount = len(p1.shingles)
			} else {
				sim, sharedCount = calculateJaccard(p1.shingles, p2.shingles)
				if sim >= threshold {
					category = CategoryNear
				} else if sim >= 0.50 {
					category = CategoryDerivative
				} else {
					category = CategoryUnique
				}
			}

			if sim >= 0.50 {
				rec := generateRecommendation(category, sim)
				duplicatePairs = append(duplicatePairs, DuplicatePair{
					SourceID:           p1.doc.ID,
					SourceTitle:        p1.doc.Title,
					TargetID:           p2.doc.ID,
					TargetTitle:        p2.doc.Title,
					SimilarityScore:    math.Round(sim*100) / 100,
					Category:           category,
					SharedShingleCount: sharedCount,
					Recommendation:     rec,
				})

				// Group into clusters if similarity >= threshold
				if sim >= threshold {
					canonical := i
					dup := j
					// If j was earlier chosen as canonical or i is already assigned
					if !assignedToCluster[canonical] && !assignedToCluster[dup] {
						clusterMap[canonical] = append(clusterMap[canonical], dup)
						assignedToCluster[dup] = true
					} else if assignedToCluster[canonical] {
						// Find existing canonical group
						for c, dups := range clusterMap {
							for _, dIdx := range dups {
								if dIdx == canonical {
									clusterMap[c] = append(clusterMap[c], dup)
									assignedToCluster[dup] = true
									break
								}
							}
						}
					}
				}
			}
		}
	}

	// Step 3: Construct Cluster Groups
	var clusters []ClusterGroup
	var totalSavingsBytes int64
	var totalDuplicateDocs int

	clusterCounter := 1
	for canonicalIdx, dupIndices := range clusterMap {
		canonical := profiles[canonicalIdx]
		var dups []DocumentItem
		var clusterSavings int64
		var simSum float64

		for _, dIdx := range dupIndices {
			dupDoc := profiles[dIdx]
			dups = append(dups, dupDoc.doc)
			clusterSavings += dupDoc.byteCount
			totalDuplicateDocs++

			sim, _ := calculateJaccard(canonical.shingles, dupDoc.shingles)
			simSum += sim
		}

		avgSim := 1.0
		if len(dups) > 0 {
			avgSim = math.Round((simSum/float64(len(dups)))*100) / 100
		}

		totalSavingsBytes += clusterSavings

		clusters = append(clusters, ClusterGroup{
			ClusterID:           fmt.Sprintf("cluster-%03d", clusterCounter),
			CanonicalDocumentID: canonical.doc.ID,
			CanonicalTitle:      canonical.doc.Title,
			DuplicateDocuments:  dups,
			AverageSimilarity:   avgSim,
			TotalSavingsBytes:   clusterSavings,
			RecommendedAction:   "Retain canonical document; suppress duplicates from RAG retrieval and vector indexing",
		})
		clusterCounter++
	}

	uniqueDocsCount := len(docs) - totalDuplicateDocs
	dupRatio := 0.0
	if len(docs) > 0 {
		dupRatio = math.Round((float64(totalDuplicateDocs)/float64(len(docs)))*100) / 100
	}

	// Approximate token savings: ~4 bytes per token
	estimatedTokenSavings := int(totalSavingsBytes / 4)

	return &DedupReport{
		TotalAnalyzed:         len(docs),
		UniqueCount:           uniqueDocsCount,
		DuplicateCount:        totalDuplicateDocs,
		DuplicateRatio:        dupRatio,
		Clusters:              clusters,
		DuplicatePairs:        duplicatePairs,
		EstimatedTokenSavings: estimatedTokenSavings,
		DurationMs:            time.Since(start).Milliseconds(),
		GeneratedAt:           time.Now().Format(time.RFC3339),
	}, nil
}

func tokenizeText(text string) []string {
	clean := strings.ToLower(text)
	// Replace punctuations with spaces
	replacer := strings.NewReplacer(".", " ", ",", " ", "!", " ", "?", " ", ";", " ", ":", " ", "\n", " ", "\t", " ", "-", " ")
	clean = replacer.Replace(clean)
	return strings.Fields(clean)
}

func buildShingles(words []string, n int) map[string]bool {
	shingles := make(map[string]bool)
	if len(words) < n {
		if len(words) > 0 {
			shingles[strings.Join(words, " ")] = true
		}
		return shingles
	}

	for i := 0; i <= len(words)-n; i++ {
		shingle := strings.Join(words[i:i+n], " ")
		shingles[shingle] = true
	}
	return shingles
}

func calculateJaccard(s1, s2 map[string]bool) (float64, int) {
	if len(s1) == 0 && len(s2) == 0 {
		return 1.0, 0
	}
	if len(s1) == 0 || len(s2) == 0 {
		return 0.0, 0
	}

	intersection := 0
	for k := range s1 {
		if s2[k] {
			intersection++
		}
	}

	union := len(s1) + len(s2) - intersection
	if union == 0 {
		return 0.0, 0
	}

	return float64(intersection) / float64(union), intersection
}

func md5String(text string) string {
	hash := md5.Sum([]byte(strings.TrimSpace(text)))
	return hex.EncodeToString(hash[:])
}

func generateRecommendation(category MatchCategory, sim float64) string {
	switch category {
	case CategoryExact:
		return "Exact byte-for-byte duplicate: purge or auto-archive target document."
	case CategoryNear:
		return fmt.Sprintf("Near-duplicate (%.0f%% overlap): consolidate into canonical version and update citations.", sim*100)
	case CategoryDerivative:
		return fmt.Sprintf("Derivative revision (%.0f%% overlap): mark as sequential revision or branch.", sim*100)
	default:
		return "Distinct document: retain in primary knowledge index."
	}
}
