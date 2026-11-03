package agentnetwork

import (
	"time"
)

// AgentRole defines the structural function of an autonomous node in the collaboration network.
type AgentRole string

const (
	RolePlanner     AgentRole = "PLANNER"
	RoleResearcher  AgentRole = "RESEARCHER"
	RoleCoder       AgentRole = "CODER"
	RoleReviewer    AgentRole = "REVIEWER"
	RoleValidator   AgentRole = "VALIDATOR"
	RoleCoordinator AgentRole = "COORDINATOR"
)

// MessageType classifies the intent of an inter-agent network packet.
type MessageType string

const (
	MsgTaskDelegation    MessageType = "TASK_DELEGATION"    // Upstream assigns goal to downstream
	MsgTaskResult        MessageType = "TASK_RESULT"        // Downstream delivers output to upstream
	MsgConsensusVote     MessageType = "CONSENSUS_VOTE"     // Voting on a proposal or artifact
	MsgConflictAlert     MessageType = "CONFLICT_ALERT"     // Notifying of contradiction or deadlock
	MsgStatusHeartbeat   MessageType = "STATUS_HEARTBEAT"   // Node liveness and workload telemetry
)

// ConsensusStrategy dictates voting thresholds for multi-agent agreements.
type ConsensusStrategy string

const (
	ConsensusMajority   ConsensusStrategy = "MAJORITY"    // > 50%
	ConsensusSupermajority ConsensusStrategy = "SUPERMAJORITY" // >= 66.7%
	ConsensusUnanimous  ConsensusStrategy = "UNANIMOUS"   // 100%
)

// NetworkNode represents an individual registered autonomous agent participant.
type NetworkNode struct {
	NodeID          string            `json:"node_id"`
	WorkspaceID     string            `json:"workspace_id"`
	Name            string            `json:"name"`
	Role            AgentRole         `json:"role"`
	Capabilities    []string          `json:"capabilities"`
	IsActive        bool              `json:"is_active"`
	ReputationScore float64           `json:"reputation_score"` // 0.0 - 1.0 (Dynamic weighting)
	Metadata        map[string]string `json:"metadata,omitempty"`
	LastHeartbeatAt time.Time         `json:"last_heartbeat_at"`
}

// NetworkMessage models an asynchronous communication packet routed across agents.
type NetworkMessage struct {
	MessageID   string      `json:"message_id"`
	NetworkID   string      `json:"network_id"`
	SenderID    string      `json:"sender_id"`
	RecipientID string      `json:"recipient_id"` // NodeID or "BROADCAST"
	Type        MessageType `json:"type"`
	Subject     string      `json:"subject"`
	Payload     string      `json:"payload"`
	CorrelationID string    `json:"correlation_id,omitempty"`
	Timestamp   time.Time   `json:"timestamp"`
}

// ProposalVote captures an individual agent node's ballot on a network proposal.
type ProposalVote struct {
	NodeID    string    `json:"node_id"`
	VoterRole AgentRole `json:"voter_role"`
	Approved  bool      `json:"approved"`
	Rationale string    `json:"rationale"`
	Weight    float64   `json:"weight"`
	Timestamp time.Time `json:"timestamp"`
}

// NetworkProposal represents a formal proposition requiring multi-agent consensus.
type NetworkProposal struct {
	ProposalID      string            `json:"proposal_id"`
	NetworkID       string            `json:"network_id"`
	ProposerID      string            `json:"proposer_id"`
	Title           string            `json:"title"`
	Description     string            `json:"description"`
	ProposedAction  string            `json:"proposed_action"`
	Strategy        ConsensusStrategy `json:"strategy"`
	Votes           []ProposalVote    `json:"votes"`
	IsResolved      bool              `json:"is_resolved"`
	Passed          bool              `json:"passed"`
	ApprovalRatio   float64           `json:"approval_ratio"`
	CreatedAt       time.Time         `json:"created_at"`
	ResolvedAt      *time.Time        `json:"resolved_at,omitempty"`
}

// AgentNetworkSession encapsulates an active collaborative network of autonomous nodes.
type AgentNetworkSession struct {
	NetworkID       string             `json:"network_id"`
	WorkspaceID     string             `json:"workspace_id"`
	Title           string             `json:"title"`
	Goal            string             `json:"goal"`
	Nodes           map[string]*NetworkNode `json:"nodes"`
	Messages        []NetworkMessage   `json:"messages"`
	Proposals       []NetworkProposal  `json:"proposals"`
	IsCompleted     bool               `json:"is_completed"`
	CreatedAt       time.Time          `json:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at"`
}

// CreateNetworkRequest defines parameters to spawn a new agent collaboration network.
type CreateNetworkRequest struct {
	WorkspaceID string        `json:"workspace_id"`
	Title       string        `json:"title"`
	Goal        string        `json:"goal"`
	InitialNodes []NetworkNode `json:"initial_nodes,omitempty"`
}

// SendMessageRequest defines parameters for dispatching a routed inter-agent packet.
type SendMessageRequest struct {
	SenderID      string      `json:"sender_id"`
	RecipientID   string      `json:"recipient_id"`
	Type          MessageType `json:"type"`
	Subject       string      `json:"subject"`
	Payload       string      `json:"payload"`
	CorrelationID string      `json:"correlation_id,omitempty"`
}

// CreateProposalRequest specifies parameters to initiate a consensus vote.
type CreateProposalRequest struct {
	ProposerID     string            `json:"proposer_id"`
	Title          string            `json:"title"`
	Description    string            `json:"description"`
	ProposedAction string            `json:"proposed_action"`
	Strategy       ConsensusStrategy `json:"strategy,omitempty"`
}

// CastVoteRequest captures a node's ballot on a proposal.
type CastVoteRequest struct {
	NodeID    string `json:"node_id"`
	Approved  bool   `json:"approved"`
	Rationale string `json:"rationale"`
}

