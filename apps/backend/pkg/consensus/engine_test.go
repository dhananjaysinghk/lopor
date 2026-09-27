package consensus_test

import (
	"context"
	"testing"

	"github.com/lopor-ai/lopor/pkg/consensus"
)

func TestConsensusEngine_RunDebate(t *testing.T) {
	engine := consensus.NewConsensusEngine()

	req := consensus.ConsensusRequest{
		Topic:   "Adopt PostgreSQL pgvector for enterprise semantic search and RAG indexing",
		Context: "Evaluating latency, cost, and developer velocity compared to external vector databases.",
		Rounds:  2,
	}

	verdict, err := engine.RunDebate(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error running debate: %v", err)
	}

	if verdict.Topic != req.Topic {
		t.Errorf("expected topic %s, got %s", req.Topic, verdict.Topic)
	}

	if !verdict.ConsensusReached {
		t.Errorf("expected consensus to be reached")
	}

	if verdict.AgreementScore <= 0.0 || verdict.AgreementScore > 1.0 {
		t.Errorf("expected agreement score between 0.0 and 1.0, got %f", verdict.AgreementScore)
	}

	if len(verdict.RoundsSummary) != 2 {
		t.Errorf("expected 2 rounds in summary, got %d", len(verdict.RoundsSummary))
	}

	if len(verdict.ConsensusClaims) == 0 {
		t.Errorf("expected consensus claims to be populated")
	}

	if len(verdict.ActionItems) == 0 {
		t.Errorf("expected action items to be generated")
	}
}

func TestConsensusEngine_RejectionThreshold(t *testing.T) {
	engine := consensus.NewConsensusEngine()

	req := consensus.ConsensusRequest{
		Topic:            "Disable auth checks and allow unsafe unauthenticated access to admin endpoints",
		Context:          "Fast testing in public environments",
		Rounds:           1,
		RequireUnanimity: true,
	}

	verdict, err := engine.RunDebate(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if verdict.ConsensusReached {
		t.Errorf("expected debate to reject unsafe topic with unanimity requirement")
	}

	if verdict.FinalStance != consensus.StanceReject {
		t.Errorf("expected REJECT final stance, got %s", verdict.FinalStance)
	}

	if len(verdict.DissentingPoints) == 0 {
		t.Errorf("expected dissenting security points to be recorded")
	}
}

func TestConsensusEngine_GetSpecialists(t *testing.T) {
	engine := consensus.NewConsensusEngine()
	specialists := engine.GetAvailableSpecialists()

	if len(specialists) < 5 {
		t.Errorf("expected at least 5 default debater specialists, got %d", len(specialists))
	}

	foundSecurity := false
	for _, s := range specialists {
		if s.Role == consensus.RoleSecurity {
			foundSecurity = true
			break
		}
	}

	if !foundSecurity {
		t.Errorf("expected Security Architect specialist in catalog")
	}
}
