package auditproof

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// HashString returns the lowercase hexadecimal SHA-256 hash of a string.
func HashString(data string) string {
	sum := sha256.Sum256([]byte(data))
	return hex.EncodeToString(sum[:])
}

// HashBytes returns the lowercase hexadecimal SHA-256 hash of raw bytes.
func HashBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// HashPair combines and hashes two sibling node hashes in canonical sorted order.
func HashPair(left, right string) string {
	var combined string
	if left < right {
		combined = left + ":" + right
	} else {
		combined = right + ":" + left
	}
	return HashString(combined)
}

// BuildMerkleTree constructs a binary Merkle tree from an ordered slice of leaf hashes.
// Returns the root hash and the complete tree levels for proof generation.
func BuildMerkleTree(leafHashes []string) (string, [][]string) {
	if len(leafHashes) == 0 {
		emptyRoot := strings.Repeat("0", 64)
		return emptyRoot, [][]string{{emptyRoot}}
	}

	var levels [][]string
	currentLevel := make([]string, len(leafHashes))
	copy(currentLevel, leafHashes)
	levels = append(levels, currentLevel)

	for len(currentLevel) > 1 {
		var nextLevel []string

		// If odd number of elements, duplicate last element to balance the binary tree
		if len(currentLevel)%2 != 0 {
			currentLevel = append(currentLevel, currentLevel[len(currentLevel)-1])
			levels[len(levels)-1] = currentLevel
		}

		for i := 0; i < len(currentLevel); i += 2 {
			parent := HashPair(currentLevel[i], currentLevel[i+1])
			nextLevel = append(nextLevel, parent)
		}

		levels = append(levels, nextLevel)
		currentLevel = nextLevel
	}

	return levels[len(levels)-1][0], levels
}

// GenerateMerkleProof extracts the sibling audit path hashes for a specific leaf index.
func GenerateMerkleProof(leafIndex int, levels [][]string) []string {
	if len(levels) <= 1 || leafIndex < 0 || leafIndex >= len(levels[0]) {
		return nil
	}

	var auditPath []string
	idx := leafIndex

	for levelIdx := 0; levelIdx < len(levels)-1; levelIdx++ {
		level := levels[levelIdx]

		var siblingIdx int
		if idx%2 == 0 {
			siblingIdx = idx + 1
		} else {
			siblingIdx = idx - 1
		}

		if siblingIdx < len(level) {
			auditPath = append(auditPath, level[siblingIdx])
		} else {
			auditPath = append(auditPath, level[idx])
		}

		idx /= 2
	}

	return auditPath
}

// VerifyMerkleProof cryptographically checks if a leaf hash and audit path resolve to the root hash.
func VerifyMerkleProof(leafHash string, auditPath []string, rootHash string) bool {
	if leafHash == "" || rootHash == "" {
		return false
	}
	if len(auditPath) == 0 {
		return leafHash == rootHash
	}

	current := leafHash
	for _, sibling := range auditPath {
		current = HashPair(current, sibling)
	}

	return current == rootHash
}

