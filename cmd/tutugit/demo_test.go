package main

import (
	"context"
	"testing"
)

func TestMockRunner_StageUnstageCommit(t *testing.T) {
	m := initialDemoModel()

    // demo.go it's an interface, let's verify through the interface
    ctx := context.Background()

    // initial state
    status, _ := m.git.GetStatus(ctx)
    if status == "" {
        t.Fatal("Initial status should not be empty")
    }

    // findd unstaged file to stage
    targetFile := "config.yml"
    
    // stage the file
    err := m.git.StageFile(ctx, targetFile)
    if err != nil {
        t.Fatalf("Failed to stage file: %v", err)
    }

    // verify it's staged
    files, _ := m.git.ParseStatus(ctx)
    staged := false
    for _, f := range files {
        if f.Path == targetFile && f.Staged {
            staged = true
            break
        }
    }
    if !staged {
        t.Errorf("File %s should be staged", targetFile)
    }

    // unstage the file
    err = m.git.UnstageFile(ctx, targetFile)
    if err != nil {
        t.Fatalf("Failed to unstage file: %v", err)
    }

    // verify it's unstaged
    files, _ = m.git.ParseStatus(ctx)
    for _, f := range files {
        if f.Path == targetFile && f.Staged {
            t.Errorf("File %s should not be staged", targetFile)
            break
        }
    }

    // stage again and commit
    m.git.StageFile(ctx, targetFile)
    initialCommits, _ := m.git.GetLog(ctx, 100)
    initialCommitCount := len(initialCommits)

    err = m.git.Commit(ctx, "feat: awesome new feature")
    if err != nil {
        t.Fatalf("Failed to commit: %v", err)
    }

    // verify commit was added
    finalCommits, _ := m.git.GetLog(ctx, 100)
    if len(finalCommits) != initialCommitCount+1 {
        t.Errorf("Expected %d commits, got %d", initialCommitCount+1, len(finalCommits))
    }

    // verify commit message
    if finalCommits[0].Message != "feat: awesome new feature" {
        t.Errorf("Expected commit message 'feat: awesome new feature', got '%s'", finalCommits[0].Message)
    }

    // verify file is no longer in status (it was fully staged and committed)
    files, _ = m.git.ParseStatus(ctx)
    for _, f := range files {
        if f.Path == targetFile {
            t.Errorf("File %s should no longer be in status after commit", targetFile)
        }
    }
}

func TestMockRunner_ReflogReset(t *testing.T) {
	m := initialDemoModel()
	ctx := context.Background()

	// check initial history length
	initialCommits, _ := m.git.GetLog(ctx, 100)
	if len(initialCommits) < 3 {
		t.Fatal("Need at least 3 commits for this test")
	}

	// target a commit in the past (e.g. the 3rd commit)
	targetCommit := initialCommits[2]

	// execute reset
	err := m.git.ResetToHash(ctx, targetCommit.Hash)
	if err != nil {
		t.Fatalf("Failed to reset to hash: %v", err)
	}

	// verify history was truncated
	finalCommits, _ := m.git.GetLog(ctx, 100)
	if len(finalCommits) != len(initialCommits)-2 {
		t.Errorf("Expected %d commits after reset, got %d", len(initialCommits)-2, len(finalCommits))
	}

	// verify the new HEAD is the target commit
	if finalCommits[0].Hash != targetCommit.Hash {
		t.Errorf("Expected new HEAD to be %s, got %s", targetCommit.Hash, finalCommits[0].Hash)
	}

	// verify reflog was updated
	reflog, _ := m.git.GetReflog(ctx, 100)
	if len(reflog) == 0 {
		t.Fatal("Reflog should not be empty")
	}
	
	latestReflog := reflog[0]
	if latestReflog.Action != "reset" {
		t.Errorf("Expected reflog action to be 'reset', got '%s'", latestReflog.Action)
	}
	if latestReflog.Hash != targetCommit.Hash {
		t.Errorf("Expected reflog hash to match target commit, got '%s'", latestReflog.Hash)
	}
}
