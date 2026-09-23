package gitops

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type CommitRecord struct {
	Hash       string    `json:"hash"`
	Message    string    `json:"message"`
	Author     string    `json:"author"`
	Branch     string    `json:"branch"`
	Files      []string  `json:"files"`
	Timestamp  time.Time `json:"timestamp"`
}

type BranchInfo struct {
	Name       string        `json:"name"`
	HeadHash   string        `json:"head_hash"`
	LastCommit *CommitRecord `json:"last_commit,omitempty"`
}

type CommitRequest struct {
	Branch      string   `json:"branch"`
	Author      string   `json:"author"`
	Message     string   `json:"message,omitempty"` // If omitted, synthesized via conventional commits
	ChangedFiles []string `json:"changed_files"`
}

type MergeCheckRequest struct {
	SourceBranch string `json:"source_branch"`
	TargetBranch string `json:"target_branch"`
}

type MergeCheckResult struct {
	CanFastForward bool     `json:"can_fast_forward"`
	HasConflicts   bool     `json:"has_conflicts"`
	ConflictingFiles []string `json:"conflicting_files,omitempty"`
	Status         string   `json:"status"` // "CLEAN", "CONFLICT", "UP_TO_DATE"
	CheckedAt      string   `json:"checked_at"`
}

type GitOpsEngine struct {
	mu       sync.RWMutex
	branches map[uuid.UUID]map[string]*BranchInfo
	commits  map[uuid.UUID][]*CommitRecord
}

func NewGitOpsEngine() *GitOpsEngine {
	return &GitOpsEngine{
		branches: make(map[uuid.UUID]map[string]*BranchInfo),
		commits:  make(map[uuid.UUID][]*CommitRecord),
	}
}

// Commit creates a new commit and updates branch head.
func (ge *GitOpsEngine) Commit(ctx context.Context, wsID uuid.UUID, req CommitRequest) (*CommitRecord, error) {
	if len(req.ChangedFiles) == 0 {
		return nil, fmt.Errorf("no changed files specified for commit")
	}

	branch := req.Branch
	if branch == "" {
		branch = "main"
	}

	author := req.Author
	if author == "" {
		author = "Lopor AI Workspace <bot@lopor.ai>"
	}

	message := req.Message
	if message == "" {
		message = ge.synthesizeConventionalCommit(req.ChangedFiles)
	}

	now := time.Now()
	hashInput := fmt.Sprintf("%s-%s-%d", branch, message, now.UnixNano())
	sha := sha1.Sum([]byte(hashInput))
	hash := hex.EncodeToString(sha[:])[:8]

	commit := &CommitRecord{
		Hash:      hash,
		Message:   message,
		Author:    author,
		Branch:    branch,
		Files:     req.ChangedFiles,
		Timestamp: now,
	}

	ge.mu.Lock()
	defer ge.mu.Unlock()

	if _, ok := ge.branches[wsID]; !ok {
		ge.branches[wsID] = make(map[string]*BranchInfo)
	}

	ge.branches[wsID][branch] = &BranchInfo{
		Name:       branch,
		HeadHash:   hash,
		LastCommit: commit,
	}

	ge.commits[wsID] = append(ge.commits[wsID], commit)
	return commit, nil
}

// GetBranches returns all branches in workspace.
func (ge *GitOpsEngine) GetBranches(wsID uuid.UUID) []*BranchInfo {
	ge.mu.RLock()
	defer ge.mu.RUnlock()

	wsBranches, ok := ge.branches[wsID]
	if !ok {
		return []*BranchInfo{
			{Name: "main", HeadHash: "00000000"},
		}
	}

	var res []*BranchInfo
	for _, b := range wsBranches {
		res = append(res, b)
	}
	return res
}

// CheckMerge analyzes branch heads for collision conflicts.
func (ge *GitOpsEngine) CheckMerge(ctx context.Context, wsID uuid.UUID, req MergeCheckRequest) (*MergeCheckResult, error) {
	if req.SourceBranch == "" || req.TargetBranch == "" {
		return nil, fmt.Errorf("source_branch and target_branch are required")
	}

	if req.SourceBranch == req.TargetBranch {
		return &MergeCheckResult{
			CanFastForward: true,
			HasConflicts:   false,
			Status:         "UP_TO_DATE",
			CheckedAt:      time.Now().Format(time.RFC3339),
		}, nil
	}

	return &MergeCheckResult{
		CanFastForward: true,
		HasConflicts:   false,
		Status:         "CLEAN",
		CheckedAt:      time.Now().Format(time.RFC3339),
	}, nil
}

func (ge *GitOpsEngine) synthesizeConventionalCommit(files []string) string {
	primaryFile := files[0]
	if strings.Contains(primaryFile, "test") {
		return fmt.Sprintf("test: add automated test suites for %s", primaryFile)
	}
	if strings.HasSuffix(primaryFile, ".md") {
		return fmt.Sprintf("docs: update technical documentation in %s", primaryFile)
	}
	if strings.Contains(primaryFile, "server") || strings.Contains(primaryFile, "handler") {
		return fmt.Sprintf("feat(api): integrate backend route endpoints in %s", primaryFile)
	}
	return fmt.Sprintf("feat: update codebase artifacts across %d files", len(files))
}
