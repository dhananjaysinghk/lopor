package agentnetwork

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrNetworkNotFound  = errors.New("agent network session not found")
	ErrNodeNotFound     = errors.New("agent node not found in network")
	ErrProposalNotFound = errors.New("network proposal not found")
	ErrAlreadyVoted     = errors.New("agent node has already voted on this proposal")
	ErrInvalidRequest   = errors.New("invalid agent network request parameters")
)

// AgentNetworkEngine manages multi-agent swarms, message bus routing, and consensus protocols.
type AgentNetworkEngine struct {
	mu       sync.RWMutex
	networks map[string]*AgentNetworkSession
}

// NewEngine initializes a new AgentNetworkEngine.
func NewEngine() *AgentNetworkEngine {
	return &AgentNetworkEngine{
		networks: make(map[string]*AgentNetworkSession),
	}
}

// CreateNetwork initializes an autonomous collaborative agent network session.
func (e *AgentNetworkEngine) CreateNetwork(req CreateNetworkRequest) (*AgentNetworkSession, error) {
	if req.WorkspaceID == "" || req.Title == "" || req.Goal == "" {
		return nil, fmt.Errorf("%w: workspace_id, title, and goal are required", ErrInvalidRequest)
	}

	networkID := generateID("net")
	now := time.Now().UTC()

	nodes := make(map[string]*NetworkNode)
	if len(req.InitialNodes) > 0 {
		for _, n := range req.InitialNodes {
			nodeCopy := n
			if nodeCopy.NodeID == "" {
				nodeCopy.NodeID = generateID("node")
			}
			nodeCopy.WorkspaceID = req.WorkspaceID
			nodeCopy.IsActive = true
			if nodeCopy.ReputationScore <= 0 {
				nodeCopy.ReputationScore = 0.95
			}
			nodeCopy.LastHeartbeatAt = now
			nodes[nodeCopy.NodeID] = &nodeCopy
		}
	} else {
		// Spawn default collaborative swarm triad (Coordinator, Researcher, Reviewer)
		defaultNodes := getDefaultSwarmNodes(req.WorkspaceID)
		for _, n := range defaultNodes {
			nodes[n.NodeID] = n
		}
	}

	session := &AgentNetworkSession{
		NetworkID:   networkID,
		WorkspaceID: req.WorkspaceID,
		Title:       req.Title,
		Goal:        req.Goal,
		Nodes:       nodes,
		Messages:    make([]NetworkMessage, 0),
		Proposals:   make([]NetworkProposal, 0),
		IsCompleted: false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	e.mu.Lock()
	e.networks[networkID] = session
	e.mu.Unlock()

	return session, nil
}

// RegisterNode adds an autonomous agent node to an active network.
func (e *AgentNetworkEngine) RegisterNode(workspaceID, networkID string, node NetworkNode) (*NetworkNode, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	session, exists := e.networks[networkID]
	if !exists || session.WorkspaceID != workspaceID {
		return nil, ErrNetworkNotFound
	}

	if node.NodeID == "" {
		node.NodeID = generateID("node")
	}
	node.WorkspaceID = workspaceID
	node.IsActive = true
	if node.ReputationScore <= 0 {
		node.ReputationScore = 0.95
	}
	node.LastHeartbeatAt = time.Now().UTC()

	session.Nodes[node.NodeID] = &node
	session.UpdatedAt = time.Now().UTC()

	return &node, nil
}

// SendMessage dispatches an asynchronous point-to-point or broadcast message across agents.
func (e *AgentNetworkEngine) SendMessage(workspaceID, networkID string, req SendMessageRequest) (*NetworkMessage, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	session, exists := e.networks[networkID]
	if !exists || session.WorkspaceID != workspaceID {
		return nil, ErrNetworkNotFound
	}

	// Validate sender
	if _, ok := session.Nodes[req.SenderID]; !ok {
		return nil, fmt.Errorf("%w: sender %s", ErrNodeNotFound, req.SenderID)
	}

	// Validate recipient if not broadcast
	if req.RecipientID != "BROADCAST" {
		if _, ok := session.Nodes[req.RecipientID]; !ok {
			return nil, fmt.Errorf("%w: recipient %s", ErrNodeNotFound, req.RecipientID)
		}
	}

	msgID := generateID("msg")
	now := time.Now().UTC()

	msg := NetworkMessage{
		MessageID:     msgID,
		NetworkID:     networkID,
		SenderID:      req.SenderID,
		RecipientID:   req.RecipientID,
		Type:          req.Type,
		Subject:       req.Subject,
		Payload:       req.Payload,
		CorrelationID: req.CorrelationID,
		Timestamp:     now,
	}

	session.Messages = append(session.Messages, msg)
	session.UpdatedAt = now

	return &msg, nil
}

// CreateProposal initiates a formal consensus vote among active network nodes.
func (e *AgentNetworkEngine) CreateProposal(workspaceID, networkID string, req CreateProposalRequest) (*NetworkProposal, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	session, exists := e.networks[networkID]
	if !exists || session.WorkspaceID != workspaceID {
		return nil, ErrNetworkNotFound
	}

	proposer, ok := session.Nodes[req.ProposerID]
	if !ok {
		return nil, fmt.Errorf("%w: proposer %s", ErrNodeNotFound, req.ProposerID)
	}

	strategy := req.Strategy
	if strategy == "" {
		strategy = ConsensusMajority
	}

	proposalID := generateID("prop")
	now := time.Now().UTC()

	proposal := NetworkProposal{
		ProposalID:     proposalID,
		NetworkID:      networkID,
		ProposerID:     proposer.NodeID,
		Title:          req.Title,
		Description:    req.Description,
		ProposedAction: req.ProposedAction,
		Strategy:       strategy,
		Votes:          make([]ProposalVote, 0),
		IsResolved:     false,
		Passed:         false,
		ApprovalRatio:  0.0,
		CreatedAt:      now,
	}

	session.Proposals = append(session.Proposals, proposal)
	session.UpdatedAt = now

	return &proposal, nil
}

// CastVote records an individual node ballot and evaluates whether the proposal resolves.
func (e *AgentNetworkEngine) CastVote(workspaceID, networkID, proposalID string, req CastVoteRequest) (*NetworkProposal, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	session, exists := e.networks[networkID]
	if !exists || session.WorkspaceID != workspaceID {
		return nil, ErrNetworkNotFound
	}

	voter, ok := session.Nodes[req.NodeID]
	if !ok {
		return nil, fmt.Errorf("%w: voter %s", ErrNodeNotFound, req.NodeID)
	}

	var targetProp *NetworkProposal
	for i := range session.Proposals {
		if session.Proposals[i].ProposalID == proposalID {
			targetProp = &session.Proposals[i]
			break
		}
	}

	if targetProp == nil {
		return nil, ErrProposalNotFound
	}
	if targetProp.IsResolved {
		return targetProp, nil
	}

	// Check if already voted
	for _, v := range targetProp.Votes {
		if v.NodeID == req.NodeID {
			return nil, ErrAlreadyVoted
		}
	}

	vote := ProposalVote{
		NodeID:    voter.NodeID,
		VoterRole: voter.Role,
		Approved:  req.Approved,
		Rationale: req.Rationale,
		Weight:    voter.ReputationScore,
		Timestamp: time.Now().UTC(),
	}

	targetProp.Votes = append(targetProp.Votes, vote)

	// Count active nodes for consensus threshold check
	activeCount := 0
	for _, n := range session.Nodes {
		if n.IsActive {
			activeCount++
		}
	}

	EvaluateConsensus(targetProp, activeCount)
	session.UpdatedAt = time.Now().UTC()

	return targetProp, nil
}

// GetNetwork retrieves full session topology, message log, and proposals.
func (e *AgentNetworkEngine) GetNetwork(workspaceID, networkID string) (*AgentNetworkSession, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	session, exists := e.networks[networkID]
	if !exists || session.WorkspaceID != workspaceID {
		return nil, ErrNetworkNotFound
	}

	return session, nil
}

// ListNetworks returns all collaboration networks in the workspace.
func (e *AgentNetworkEngine) ListNetworks(workspaceID string) []*AgentNetworkSession {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var result []*AgentNetworkSession
	for _, s := range e.networks {
		if s.WorkspaceID == workspaceID {
			result = append(result, s)
		}
	}
	return result
}

func getDefaultSwarmNodes(workspaceID string) []*NetworkNode {
	now := time.Now().UTC()
	return []*NetworkNode{
		{
			NodeID:          "node_coordinator",
			WorkspaceID:     workspaceID,
			Name:            "Orchestrator Agent",
			Role:            RoleCoordinator,
			Capabilities:    []string{"task_decomposition", "consensus_arbitration", "monitoring"},
			IsActive:        true,
			ReputationScore: 0.99,
			LastHeartbeatAt: now,
		},
		{
			NodeID:          "node_researcher",
			WorkspaceID:     workspaceID,
			Name:            "RAG Research Agent",
			Role:            RoleResearcher,
			Capabilities:    []string{"hybrid_search", "document_summarization", "provenance_check"},
			IsActive:        true,
			ReputationScore: 0.95,
			LastHeartbeatAt: now,
		},
		{
			NodeID:          "node_reviewer",
			WorkspaceID:     workspaceID,
			Name:            "Quality Reviewer Agent",
			Role:            RoleReviewer,
			Capabilities:    []string{"hallucination_detection", "rubric_scoring", "safety_audit"},
			IsActive:        true,
			ReputationScore: 0.92,
			LastHeartbeatAt: now,
		},
	}
}

func generateID(prefix string) string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%s_%s", prefix, hex.EncodeToString(b))
}

