package auditproof

import (
	"strings"
	"testing"
)

func TestMerkleTreeConstructionAndProof(t *testing.T) {
	leaves := []string{
		HashString("event_1_payload"),
		HashString("event_2_payload"),
		HashString("event_3_payload"),
		HashString("event_4_payload"),
	}

	rootHash, levels := BuildMerkleTree(leaves)
	if rootHash == "" {
		t.Fatalf("expected non-empty root hash")
	}

	// Generate and verify proof for leaf 2 (index 1)
	proof := GenerateMerkleProof(1, levels)
	if len(proof) == 0 {
		t.Fatalf("expected non-empty audit path for leaf 1")
	}

	isValid := VerifyMerkleProof(leaves[1], proof, rootHash)
	if !isValid {
		t.Errorf("expected valid Merkle proof for authentic leaf")
	}

	// Tampered leaf hash must fail verification
	tamperedLeaf := HashString("malicious_tampered_payload")
	isTamperedValid := VerifyMerkleProof(tamperedLeaf, proof, rootHash)
	if isTamperedValid {
		t.Errorf("expected tampered leaf verification to fail")
	}
}

func TestAppendEventAndBlockSealing(t *testing.T) {
	engine := NewEngine()
	wsID := "ws_compliance"

	// 1. Record 3 audit events
	e1, err := engine.RecordEvent(RecordEventRequest{
		WorkspaceID:   wsID,
		EventType:     EventModelCompletion,
		ActorID:       "agent_orchestrator",
		ActionSummary: "Synthesized financial Q3 forecast completion",
		RawPayload:    "{\"prompt\": \"Forecast revenue\", \"response\": \"$4.2M target\"}",
	})
	if err != nil {
		t.Fatalf("failed to record event 1: %v", err)
	}

	e2, err := engine.RecordEvent(RecordEventRequest{
		WorkspaceID:   wsID,
		EventType:     EventToolExecution,
		ActorID:       "agent_worker_sql",
		ActionSummary: "Executed read-only analytics query on transactions",
		RawPayload:    "{\"query\": \"SELECT SUM(amount) FROM payments WHERE q = 3\"}",
	})
	if err != nil {
		t.Fatalf("failed to record event 2: %v", err)
	}

	if e2.PrevEventHash != e1.EventHash {
		t.Errorf("expected event 2 PrevEventHash to match event 1 EventHash")
	}

	// 2. Generate inclusion proof for event 1
	proof, err := engine.GenerateProof(wsID, e1.EventID)
	if err != nil {
		t.Fatalf("failed to generate proof for event 1: %v", err)
	}
	if !proof.IsVerified {
		t.Errorf("expected proof to be self-verifying")
	}

	// 3. Seal the active block
	sealedBlock, err := engine.SealBlock(wsID)
	if err != nil {
		t.Fatalf("failed to seal block: %v", err)
	}
	if !sealedBlock.IsSealed {
		t.Errorf("expected sealed block to have IsSealed=true")
	}
	if sealedBlock.RootHash == "" || sealedBlock.BlockHash == "" {
		t.Errorf("expected non-empty root and block hashes")
	}

	// 4. Verify full ledger integrity
	integrity := engine.VerifyIntegrity(wsID)
	if !integrity.IsValid {
		t.Errorf("expected ledger integrity to be valid, got: %s", integrity.IntegrityReport)
	}
	if integrity.TotalEventsVerified != 2 {
		t.Errorf("expected 2 verified events, got %d", integrity.TotalEventsVerified)
	}
}

func TestTamperDetection(t *testing.T) {
	engine := NewEngine()
	wsID := "ws_tamper_test"

	_, _ = engine.RecordEvent(RecordEventRequest{
		WorkspaceID:   wsID,
		EventType:     EventRAGRetrieval,
		ActorID:       "rag_engine",
		ActionSummary: "Retrieved confidential patient telemetry",
		RawPayload:    "{\"doc_id\": \"doc_pat_102\", \"sensitive\": true}",
	})

	sealedBlock, err := engine.SealBlock(wsID)
	if err != nil {
		t.Fatalf("failed to seal block: %v", err)
	}

	// Verify before tamper
	validIntegrity := engine.VerifyIntegrity(wsID)
	if !validIntegrity.IsValid {
		t.Fatalf("expected ledger to be valid prior to tamper")
	}

	// Deliberately tamper with the event hash inside the sealed block
	sealedBlock.Events[0].EventHash = strings.Repeat("f", 64)

	// Verify after tamper - must detect corruption!
	tamperResult := engine.VerifyIntegrity(wsID)
	if tamperResult.IsValid {
		t.Errorf("expected tamper detection to flag corrupted event hash")
	}
	if !strings.Contains(tamperResult.IntegrityReport, "Merkle root mismatch") {
		t.Errorf("expected integrity report to cite root mismatch: %s", tamperResult.IntegrityReport)
	}
}

func TestExportComplianceBundle(t *testing.T) {
	engine := NewEngine()
	wsID := "ws_audit_bundle"

	_, _ = engine.RecordEvent(RecordEventRequest{
		WorkspaceID:   wsID,
		EventType:     EventGuardrailInterception,
		ActorID:       "security_firewall",
		ActionSummary: "Intercepted prompt injection attempt",
		RawPayload:    "{\"blocked\": true, \"threat\": \"PROMPT_INJECTION\"}",
	})

	bundle := engine.ExportBundle(wsID, ComplianceEUAIAct)
	if bundle.WorkspaceID != wsID {
		t.Errorf("mismatched workspace in bundle: %s vs %s", bundle.WorkspaceID, wsID)
	}
	if bundle.Standard != ComplianceEUAIAct {
		t.Errorf("mismatched compliance standard: %s", bundle.Standard)
	}
	if !bundle.VerificationStatus {
		t.Errorf("expected verification status true in exported bundle")
	}
	if bundle.TotalEvents != 1 {
		t.Errorf("expected 1 event in bundle, got %d", bundle.TotalEvents)
	}
}

