package agentnetwork

import (
	"testing"
)

func TestCreateNetworkWithDefaultSwarm(t *testing.T) {
	engine := NewEngine()
	wsID := "ws_alpha"

	req := CreateNetworkRequest{
		WorkspaceID: wsID,
		Title:       "Autonomous Code Architecture Swarm",
		Goal:        "Design, review, and evaluate distributed ledger microservices.",
	}

	net, err := engine.CreateNetwork(req)
	if err != nil {
		t.Fatalf("failed to create agent network: %v", err)
	}

	if net.NetworkID == "" {
		t.Errorf("expected non-empty network ID")
	}
	if len(net.Nodes) != 3 {
		t.Fatalf("expected 3 default swarm nodes, got %d", len(net.Nodes))
	}
	if _, exists := net.Nodes["node_coordinator"]; !exists {
		t.Errorf("expected node_coordinator in swarm")
	}
}

func TestInterAgentMessageBusRouting(t *testing.T) {
	engine := NewEngine()
	wsID := "ws_alpha"

	net, _ := engine.CreateNetwork(CreateNetworkRequest{
		WorkspaceID: wsID,
		Title:       "Test Bus Network",
		Goal:        "Verify communication pipeline",
	})

	// Dispatch message from coordinator to researcher
	msg, err := engine.SendMessage(wsID, net.NetworkID, SendMessageRequest{
		SenderID:    "node_coordinator",
		RecipientID: "node_researcher",
		Type:        MsgTaskDelegation,
		Subject:     "Execute semantic query on cloud docs",
		Payload:     "{\"query\": \"Kubernetes HPA autoscaling thresholds\"}",
	})
	if err != nil {
		t.Fatalf("failed to send message: %v", err)
	}

	if msg.MessageID == "" {
		t.Errorf("expected generated message ID")
	}

	// Verify message in network history
	snapshot, _ := engine.GetNetwork(wsID, net.NetworkID)
	if len(snapshot.Messages) != 1 {
		t.Fatalf("expected 1 message in network ledger, got %d", len(snapshot.Messages))
	}
	if snapshot.Messages[0].RecipientID != "node_researcher" {
		t.Errorf("expected recipient to be node_researcher")
	}
}

func TestConsensusProposalMajorityVote(t *testing.T) {
	engine := NewEngine()
	wsID := "ws_alpha"

	net, _ := engine.CreateNetwork(CreateNetworkRequest{
		WorkspaceID: wsID,
		Title:       "Consensus Testing Network",
		Goal:        "Verify voting thresholds",
	})

	// 1. Propose action: Deploy Postgres v16 upgrade
	prop, err := engine.CreateProposal(wsID, net.NetworkID, CreateProposalRequest{
		ProposerID:     "node_coordinator",
		Title:          "Approve Database Partitioning Plan",
		Description:    "Partition vector collections by workspace ID.",
		ProposedAction: "EXECUTE_MIGRATION_PG16",
		Strategy:       ConsensusMajority,
	})
	if err != nil {
		t.Fatalf("failed to create proposal: %v", err)
	}

	// 2. Cast vote from node_coordinator (Approve)
	p1, err := engine.CastVote(wsID, net.NetworkID, prop.ProposalID, CastVoteRequest{
		NodeID:    "node_coordinator",
		Approved:  true,
		Rationale: "Architecture verified against benchmarks",
	})
	if err != nil {
		t.Fatalf("failed to cast vote: %v", err)
	}
	if p1.IsResolved {
		t.Errorf("expected proposal to be unresolved after 1 of 3 votes")
	}

	// 3. Cast vote from node_researcher (Approve) -> Reaches 2 of 3 votes (> 50% majority)
	p2, err := engine.CastVote(wsID, net.NetworkID, prop.ProposalID, CastVoteRequest{
		NodeID:    "node_researcher",
		Approved:  true,
		Rationale: "Empirical latency improvements confirmed",
	})
	if err != nil {
		t.Fatalf("failed to cast second vote: %v", err)
	}

	if !p2.IsResolved {
		t.Errorf("expected proposal to resolve upon achieving majority")
	}
	if !p2.Passed {
		t.Errorf("expected proposal to have passed")
	}
}

func TestWorkspaceTenancyIsolation(t *testing.T) {
	engine := NewEngine()

	net, _ := engine.CreateNetwork(CreateNetworkRequest{
		WorkspaceID: "ws_alpha",
		Title:       "Alpha Swarm",
		Goal:        "Confidential research",
	})

	// Access from ws_beta should be rejected
	_, err := engine.GetNetwork("ws_beta", net.NetworkID)
	if err != ErrNetworkNotFound {
		t.Errorf("expected ErrNetworkNotFound for unauthorized workspace, got %v", err)
	}

	_, err = engine.SendMessage("ws_beta", net.NetworkID, SendMessageRequest{
		SenderID:    "node_coordinator",
		RecipientID: "node_researcher",
	})
	if err != ErrNetworkNotFound {
		t.Errorf("expected ErrNetworkNotFound for cross-tenant message dispatch, got %v", err)
	}
}

