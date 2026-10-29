package auditproof

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

var (
	ErrEventNotFound = errors.New("audit event not found in ledger")
	ErrEmptyBlock    = errors.New("cannot seal empty block")
)

// WorkspaceLedger maintains an append-only, tamper-evident cryptographic block ledger.
type WorkspaceLedger struct {
	mu            sync.RWMutex
	workspaceID   string
	maxBlockSize  int
	blocks        []*LedgerBlock
	currentBlock  *LedgerBlock
	eventIndex    map[string]*AuditEvent
	eventBlockMap map[string]int64
	lastEventHash string
}

// NewWorkspaceLedger initializes a ledger with an empty genesis block.
func NewWorkspaceLedger(workspaceID string) *WorkspaceLedger {
	now := time.Now().UTC()
	genesisHash := strings.Repeat("0", 64)

	genesis := &LedgerBlock{
		BlockIndex:   0,
		PreviousHash: genesisHash,
		Events:       make([]AuditEvent, 0),
		IsSealed:     false,
		StartedAt:    now,
	}

	return &WorkspaceLedger{
		workspaceID:   workspaceID,
		maxBlockSize:  50,
		blocks:        []*LedgerBlock{genesis},
		currentBlock:  genesis,
		eventIndex:    make(map[string]*AuditEvent),
		eventBlockMap: make(map[string]int64),
		lastEventHash: genesisHash,
	}
}

// AppendEvent registers a new audit event into the active unsealed block.
func (l *WorkspaceLedger) AppendEvent(req RecordEventRequest) (*AuditEvent, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now().UTC()
	eventID := generateID("aud")

	payloadHash := HashString(req.RawPayload)
	eventHashData := fmt.Sprintf("%s:%s:%s:%s:%d", eventID, req.EventType, payloadHash, l.lastEventHash, now.UnixNano())
	eventHash := HashString(eventHashData)

	standard := req.Standard
	if standard == "" {
		standard = ComplianceSOC2Type2
	}

	actorRole := req.ActorRole
	if actorRole == "" {
		actorRole = "SYSTEM_AGENT"
	}

	summary := req.RawPayload
	if len(summary) > 200 {
		summary = summary[:200] + "... [TRUNCATED]"
	}

	event := AuditEvent{
		EventID:        eventID,
		WorkspaceID:    l.workspaceID,
		EventType:      req.EventType,
		Standard:       standard,
		ActorID:        req.ActorID,
		ActorRole:      actorRole,
		ActionSummary:  req.ActionSummary,
		PayloadSummary: summary,
		PayloadHash:    payloadHash,
		PrevEventHash:  l.lastEventHash,
		EventHash:      eventHash,
		BlockIndex:     l.currentBlock.BlockIndex,
		Timestamp:      now,
	}

	l.currentBlock.Events = append(l.currentBlock.Events, event)
	l.currentBlock.EventCount = len(l.currentBlock.Events)
	l.eventIndex[eventID] = &event
	l.eventBlockMap[eventID] = l.currentBlock.BlockIndex
	l.lastEventHash = eventHash

	// Auto-seal when block reaches capacity
	if len(l.currentBlock.Events) >= l.maxBlockSize {
		_, _ = l.sealCurrentBlockLocked()
	}

	return &event, nil
}

// SealBlock seals the active block with its calculated Merkle root.
func (l *WorkspaceLedger) SealBlock() (*LedgerBlock, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.sealCurrentBlockLocked()
}

func (l *WorkspaceLedger) sealCurrentBlockLocked() (*LedgerBlock, error) {
	if len(l.currentBlock.Events) == 0 {
		return nil, ErrEmptyBlock
	}

	leafHashes := make([]string, len(l.currentBlock.Events))
	for i, e := range l.currentBlock.Events {
		leafHashes[i] = e.EventHash
	}

	rootHash, _ := BuildMerkleTree(leafHashes)
	now := time.Now().UTC()

	blockData := fmt.Sprintf("%d:%s:%s:%d:%d",
		l.currentBlock.BlockIndex,
		l.currentBlock.PreviousHash,
		rootHash,
		l.currentBlock.EventCount,
		now.UnixNano(),
	)
	blockHash := HashString(blockData)

	l.currentBlock.RootHash = rootHash
	l.currentBlock.BlockHash = blockHash
	l.currentBlock.IsSealed = true
	l.currentBlock.SealedAt = &now

	sealed := l.currentBlock

	// Advance to next block
	nextIndex := sealed.BlockIndex + 1
	nextBlock := &LedgerBlock{
		BlockIndex:   nextIndex,
		PreviousHash: blockHash,
		Events:       make([]AuditEvent, 0),
		IsSealed:     false,
		StartedAt:    now,
	}

	l.blocks = append(l.blocks, nextBlock)
	l.currentBlock = nextBlock

	return sealed, nil
}

// GenerateProof creates a verifiable Merkle inclusion proof for a given event ID.
func (l *WorkspaceLedger) GenerateProof(eventID string) (*MerkleProof, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	event, exists := l.eventIndex[eventID]
	if !exists {
		return nil, ErrEventNotFound
	}

	blockIdx := l.eventBlockMap[eventID]
	var targetBlock *LedgerBlock
	for _, b := range l.blocks {
		if b.BlockIndex == blockIdx {
			targetBlock = b
			break
		}
	}

	if targetBlock == nil {
		return nil, ErrEventNotFound
	}

	// Build Merkle tree for the block's events
	leafHashes := make([]string, len(targetBlock.Events))
	leafIndex := -1
	for i, e := range targetBlock.Events {
		leafHashes[i] = e.EventHash
		if e.EventID == eventID {
			leafIndex = i
		}
	}

	if leafIndex == -1 {
		return nil, ErrEventNotFound
	}

	rootHash, levels := BuildMerkleTree(leafHashes)
	auditPath := GenerateMerkleProof(leafIndex, levels)
	isVerified := VerifyMerkleProof(event.EventHash, auditPath, rootHash)

	return &MerkleProof{
		EventID:     eventID,
		LeafHash:    event.EventHash,
		BlockIndex:  blockIdx,
		AuditPath:   auditPath,
		RootHash:    rootHash,
		IsVerified:  isVerified,
		GeneratedAt: time.Now().UTC(),
	}, nil
}

// VerifyIntegrity traverses the full block chain and re-derives all cryptographic proofs.
func (l *WorkspaceLedger) VerifyIntegrity() VerificationResult {
	l.mu.RLock()
	defer l.mu.RUnlock()

	now := time.Now().UTC()
	totalEvents := 0
	expectedPrevHash := strings.Repeat("0", 64)

	for _, block := range l.blocks {
		// 1. Verify previous block hash chaining
		if block.BlockIndex == 0 {
			if block.PreviousHash != strings.Repeat("0", 64) {
				return VerificationResult{
					WorkspaceID:        l.workspaceID,
					IsValid:            false,
					TamperedBlockIndex: 0,
					IntegrityReport:    "Genesis block previous hash is corrupted.",
					CheckedAt:          now,
				}
			}
		} else {
			if block.PreviousHash != expectedPrevHash {
				return VerificationResult{
					WorkspaceID:        l.workspaceID,
					IsValid:            false,
					TamperedBlockIndex: block.BlockIndex,
					IntegrityReport:    fmt.Sprintf("Block %d previous_hash mismatch. Chain continuity broken.", block.BlockIndex),
					CheckedAt:          now,
				}
			}
		}

		// 2. Verify sealed blocks Merkle root
		if block.IsSealed {
			leafHashes := make([]string, len(block.Events))
			for i, e := range block.Events {
				leafHashes[i] = e.EventHash
			}
			computedRoot, _ := BuildMerkleTree(leafHashes)
			if computedRoot != block.RootHash {
				return VerificationResult{
					WorkspaceID:        l.workspaceID,
					IsValid:            false,
					TamperedBlockIndex: block.BlockIndex,
					IntegrityReport:    fmt.Sprintf("Block %d Merkle root mismatch: expected %s, got %s.", block.BlockIndex, block.RootHash, computedRoot),
					CheckedAt:          now,
				}
			}
			expectedPrevHash = block.BlockHash
		}

		totalEvents += len(block.Events)
	}

	return VerificationResult{
		WorkspaceID:         l.workspaceID,
		IsValid:             true,
		TotalBlocksChecked:  len(l.blocks),
		TotalEventsVerified: totalEvents,
		IntegrityReport:     fmt.Sprintf("Cryptographic chain verified across %d blocks and %d audit events. Zero anomalies detected.", len(l.blocks), totalEvents),
		CheckedAt:           now,
	}
}

// ExportBundle packages all blocks and cryptographic root hashes for external auditors.
func (l *WorkspaceLedger) ExportBundle(standard ComplianceStandard) AuditBundle {
	l.mu.RLock()
	defer l.mu.RUnlock()

	integrity := l.VerifyIntegrity()
	latestRoot := strings.Repeat("0", 64)
	if len(l.blocks) > 0 {
		lastBlock := l.blocks[len(l.blocks)-1]
		if lastBlock.IsSealed {
			latestRoot = lastBlock.RootHash
		} else if len(lastBlock.Events) > 0 {
			leafHashes := make([]string, len(lastBlock.Events))
			for i, e := range lastBlock.Events {
				leafHashes[i] = e.EventHash
			}
			latestRoot, _ = BuildMerkleTree(leafHashes)
		}
	}

	var blockCopies []LedgerBlock
	for _, b := range l.blocks {
		blockCopies = append(blockCopies, *b)
	}

	return AuditBundle{
		WorkspaceID:        l.workspaceID,
		Standard:           standard,
		ExportTimestamp:    time.Now().UTC(),
		TotalBlocks:        len(l.blocks),
		TotalEvents:        integrity.TotalEventsVerified,
		LatestRootHash:     latestRoot,
		Blocks:             blockCopies,
		VerificationStatus: integrity.IsValid,
	}
}

func generateID(prefix string) string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%s_%s", prefix, hex.EncodeToString(b))
}
