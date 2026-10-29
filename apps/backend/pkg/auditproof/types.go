package auditproof

import (
	"time"
)

// AuditEventType categorizes critical operational actions recorded in the compliance ledger.
type AuditEventType string

const (
	EventModelCompletion      AuditEventType = "MODEL_COMPLETION"
	EventToolExecution        AuditEventType = "TOOL_EXECUTION"
	EventRAGRetrieval          AuditEventType = "RAG_RETRIEVAL"
	EventHITLIntervention      AuditEventType = "HITL_INTERVENTION"
	EventGuardrailInterception AuditEventType = "GUARDRAIL_INTERCEPTION"
	EventPolicyMutation        AuditEventType = "POLICY_MUTATION"
)

// ComplianceStandard identifies regulatory frameworks governing audit requirements.
type ComplianceStandard string

const (
	ComplianceSOC2Type2  ComplianceStandard = "SOC2_TYPE_II"
	ComplianceEUAIAct    ComplianceStandard = "EU_AI_ACT"
	ComplianceHIPAA      ComplianceStandard = "HIPAA_AUDIT"
	ComplianceISO27001   ComplianceStandard = "ISO_27001"
)

// AuditEvent represents a tamper-evident atomic record of an AI interaction or governance event.
type AuditEvent struct {
	EventID        string             `json:"event_id"`
	WorkspaceID    string             `json:"workspace_id"`
	EventType      AuditEventType     `json:"event_type"`
	Standard       ComplianceStandard `json:"standard"`
	ActorID        string             `json:"actor_id"`
	ActorRole      string             `json:"actor_role"`
	ActionSummary  string             `json:"action_summary"`
	PayloadSummary string             `json:"payload_summary"`
	PayloadHash    string             `json:"payload_hash"` // SHA-256 of raw canonical payload
	PrevEventHash  string             `json:"prev_event_hash"`
	EventHash      string             `json:"event_hash"` // SHA-256 of (EventID + PayloadHash + PrevEventHash + Timestamp)
	BlockIndex     int64              `json:"block_index"`
	Timestamp      time.Time          `json:"timestamp"`
}

// MerkleProof provides a zero-knowledge cryptographic proof of event existence in a block.
type MerkleProof struct {
	EventID        string   `json:"event_id"`
	LeafHash       string   `json:"leaf_hash"`
	BlockIndex     int64    `json:"block_index"`
	AuditPath      []string `json:"audit_path"` // Sequence of sibling hashes
	RootHash       string   `json:"root_hash"`
	IsVerified     bool     `json:"is_verified"`
	GeneratedAt    time.Time `json:"generated_at"`
}

// LedgerBlock aggregates a batch of audit events sealed with a cryptographic Merkle root.
type LedgerBlock struct {
	BlockIndex    int64        `json:"block_index"`
	PreviousHash  string       `json:"previous_hash"`
	RootHash      string       `json:"root_hash"`
	BlockHash     string       `json:"block_hash"`
	EventCount    int          `json:"event_count"`
	Events        []AuditEvent `json:"events"`
	IsSealed      bool         `json:"is_sealed"`
	StartedAt     time.Time    `json:"started_at"`
	SealedAt      *time.Time   `json:"sealed_at,omitempty"`
}

// VerificationResult summarizes the cryptographic integrity status of the ledger.
type VerificationResult struct {
	WorkspaceID         string   `json:"workspace_id"`
	IsValid             bool     `json:"is_valid"`
	TotalBlocksChecked  int      `json:"total_blocks_checked"`
	TotalEventsVerified int      `json:"total_events_verified"`
	TamperedBlockIndex  int64    `json:"tampered_block_index,omitempty"`
	TamperedEventID     string   `json:"tampered_event_id,omitempty"`
	IntegrityReport     string   `json:"integrity_report"`
	CheckedAt           time.Time `json:"checked_at"`
}

// RecordEventRequest specifies parameters for logging a compliant audit event.
type RecordEventRequest struct {
	WorkspaceID    string             `json:"workspace_id"`
	EventType      AuditEventType     `json:"event_type"`
	Standard       ComplianceStandard `json:"standard,omitempty"`
	ActorID        string             `json:"actor_id"`
	ActorRole      string             `json:"actor_role,omitempty"`
	ActionSummary  string             `json:"action_summary"`
	RawPayload     string             `json:"raw_payload"`
}

// AuditBundle provides a portable, tamper-evident compliance export package.
type AuditBundle struct {
	WorkspaceID        string             `json:"workspace_id"`
	Standard           ComplianceStandard `json:"standard"`
	ExportTimestamp    time.Time          `json:"export_timestamp"`
	TotalBlocks        int                `json:"total_blocks"`
	TotalEvents        int                `json:"total_events"`
	LatestRootHash     string             `json:"latest_root_hash"`
	Blocks             []LedgerBlock      `json:"blocks"`
	VerificationStatus bool               `json:"verification_status"`
}

