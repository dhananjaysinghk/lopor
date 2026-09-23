package gitops_test

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/lopor-ai/lopor/pkg/gitops"
)

func TestGitOpsEngine_CommitAndBranches(t *testing.T) {
	engine := gitops.NewGitOpsEngine()
	wsID := uuid.New()

	// 1. Create a commit with auto-generated message
	commit, err := engine.Commit(context.Background(), wsID, gitops.CommitRequest{
		Branch:       "feature/login",
		ChangedFiles: []string{"apps/backend/internal/server/server.go"},
	})
	if err != nil {
		t.Fatalf("unexpected error committing: %v", err)
	}

	if len(commit.Hash) == 0 {
		t.Errorf("expected non-empty commit hash")
	}

	if !strings.HasPrefix(commit.Message, "feat(api):") {
		t.Errorf("expected conventional commit message feat(api):, got %s", commit.Message)
	}

	// 2. Fetch branches
	branches := engine.GetBranches(wsID)
	if len(branches) != 1 {
		t.Errorf("expected 1 branch, got %d", len(branches))
	}

	// 3. Merge Check
	mergeRes, err := engine.CheckMerge(context.Background(), wsID, gitops.MergeCheckRequest{
		SourceBranch: "feature/login",
		TargetBranch: "main",
	})
	if err != nil {
		t.Fatalf("unexpected error checking merge: %v", err)
	}

	if !mergeRes.CanFastForward {
		t.Errorf("expected can fast forward, got false")
	}
	if mergeRes.HasConflicts {
		t.Errorf("expected no conflicts, got true")
	}
}

func TestGitOpsEngine_EmptyFiles(t *testing.T) {
	engine := gitops.NewGitOpsEngine()
	_, err := engine.Commit(context.Background(), uuid.New(), gitops.CommitRequest{
		ChangedFiles: []string{},
	})
	if err == nil {
		t.Errorf("expected error on empty changed files, got nil")
	}
}
