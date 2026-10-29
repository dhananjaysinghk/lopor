package auditproof

import (
	"errors"
	"sync"
)

// AuditProofEngine coordinates multi-tenant cryptographic ledgers, Merkle proofs, and compliance exports.
type AuditProofEngine struct {
	mu      sync.RWMutex
	ledgers map[string]*WorkspaceLedger
}

// NewEngine initializes a new AuditProofEngine.
func NewEngine() *AuditProofEngine {
	return &AuditProofEngine{
		ledgers: make(map[string]*WorkspaceLedger),
	}
}

func (e *AuditProofEngine) getOrCreateLedger(workspaceID string) *WorkspaceLedger {
	e.mu.Lock()
	defer e.mu.Unlock()

	l, exists := e.ledgers[workspaceID]
	if !exists {
		l = NewWorkspaceLedger(workspaceID)
		e.ledgers[workspaceID] = l
	}
	return l
}

// RecordEvent records an immutable operational audit event into the workspace ledger.
func (e *AuditProofEngine) RecordEvent(req RecordEventRequest) (*AuditEvent, error) {
	if req.WorkspaceID == "" || req.EventType == "" || req.ActionSummary == "" {
		return nil, errors.New("workspace_id, event_type, and action_summary are required")
	}

	ledger := e.getOrCreateLedger(req.WorkspaceID)
	return ledger.AppendEvent(req)
}

// SealBlock manually closes and anchors the current active block with a Merkle root.
func (e *AuditProofEngine) SealBlock(workspaceID string) (*LedgerBlock, error) {
	if workspaceID == "" {
		return nil, errors.New("workspace_id is required")
	}

	ledger := e.getOrCreateLedger(workspaceID)
	return ledger.SealBlock()
}

// GenerateProof extracts a verifiable Merkle inclusion proof for any recorded event.
func (e *AuditProofEngine) GenerateProof(workspaceID, eventID string) (*MerkleProof, error) {
	if workspaceID == "" || eventID == "" {
		return nil, errors.New("workspace_id and event_id are required")
	}

	ledger := e.getOrCreateLedger(workspaceID)
	return ledger.GenerateProof(eventID)
}

// VerifyProof verifies an external or internal Merkle inclusion proof against a root hash.
func (e *AuditProofEngine) VerifyProof(leafHash string, auditPath []string, rootHash string) bool {
	return VerifyMerkleProof(leafHash, auditPath, rootHash)
}

// VerifyIntegrity traverses and mathematically checks the hash chain and Merkle roots of the workspace ledger.
func (e *AuditProofEngine) VerifyIntegrity(workspaceID string) VerificationResult {
	ledger := e.getOrCreateLedger(workspaceID)
	return ledger.VerifyIntegrity()
}

// ExportBundle produces an exportable audit archive for regulatory certification.
func (e *AuditProofEngine) ExportBundle(workspaceID string, standard ComplianceStandard) AuditBundle {
	ledger := e.getOrCreateLedger(workspaceID)
	return ledger.ExportBundle(standard)
}

