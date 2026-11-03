package agentnetwork

import (
	"math"
	"time"
)

// EvaluateConsensus determines if an open proposal satisfies its required consensus threshold.
func EvaluateConsensus(proposal *NetworkProposal, totalActiveNodes int) (resolved bool, passed bool, ratio float64) {
	if len(proposal.Votes) == 0 || totalActiveNodes == 0 {
		return false, false, 0.0
	}

	totalWeight := 0.0
	approvedWeight := 0.0

	for _, v := range proposal.Votes {
		weight := v.Weight
		if weight <= 0.0 {
			weight = 1.0
		}
		totalWeight += weight
		if v.Approved {
			approvedWeight += weight
		}
	}

	if totalWeight > 0 {
		ratio = math.Round((approvedWeight/totalWeight)*1000) / 1000
	}

	// Quorum check: all active nodes must have voted or minimum quorum reached
	allNodesVoted := len(proposal.Votes) >= totalActiveNodes

	switch proposal.Strategy {
	case ConsensusUnanimous:
		if allNodesVoted {
			resolved = true
			passed = ratio == 1.0
		} else {
			// Early termination: if any negative vote cast, cannot be unanimous
			for _, v := range proposal.Votes {
				if !v.Approved {
					resolved = true
					passed = false
					break
				}
			}
		}

	case ConsensusSupermajority:
		if allNodesVoted || ratio >= 0.667 {
			resolved = true
			passed = ratio >= 0.667
		}

	case ConsensusMajority:
		fallthrough
	default:
		if allNodesVoted || ratio > 0.50 {
			resolved = true
			passed = ratio > 0.50
		}
	}

	if resolved {
		now := time.Now().UTC()
		proposal.IsResolved = true
		proposal.Passed = passed
		proposal.ApprovalRatio = ratio
		proposal.ResolvedAt = &now
	}

	return resolved, passed, ratio
}

// SelectBestDelegate selects the most suitable active node matching the target role and highest reputation score.
func SelectBestDelegate(nodes map[string]*NetworkNode, targetRole AgentRole, excludeNodeID string) *NetworkNode {
	var bestNode *NetworkNode
	highestScore := -1.0

	for _, node := range nodes {
		if !node.IsActive || node.NodeID == excludeNodeID {
			continue
		}
		if node.Role == targetRole {
			if node.ReputationScore > highestScore {
				highestScore = node.ReputationScore
				bestNode = node
			}
		}
	}

	return bestNode
}

