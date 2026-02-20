package git

import (
	"context"
	"fmt"
)

// MockRunner is a mock implementation of GitProvider for testing.
type MockRunner struct {
	CurrentBranch string
	Status        string
	Files         []FileStatus
	Commits       []Commit
	Reflog        []ReflogEntry
	Worktrees     []Worktree
	Tags          []string
	RemoteURL     string
	IsRebasingVal bool
	RebaseTodo    []RebaseStep
	ValidHashes   map[string]bool
	RunFunc       func(ctx context.Context, args ...string) (string, error)
	GetDiffFunc   func(ctx context.Context, path string, staged bool) (string, error)
}

func NewMockRunner() *MockRunner {
	return &MockRunner{
		ValidHashes: make(map[string]bool),
	}
}

func (m *MockRunner) Run(ctx context.Context, args ...string) (string, error) {
	if m.RunFunc != nil {
		return m.RunFunc(ctx, args...)
	}
	return "", nil
}

func (m *MockRunner) GetCurrentBranch(ctx context.Context) (string, error) {
	return m.CurrentBranch, nil
}

func (m *MockRunner) GetStatus(ctx context.Context) (string, error) {
	return m.Status, nil
}

func (m *MockRunner) StageFile(ctx context.Context, path string) error {
	for i, f := range m.Files {
		if f.Path == path {
			m.Files[i].Staged = true
			break
		}
	}
	m.updateStatusString()
	return nil
}

func (m *MockRunner) UnstageFile(ctx context.Context, path string) error {
	for i, f := range m.Files {
		if f.Path == path {
			m.Files[i].Staged = false
			break
		}
	}
	m.updateStatusString()
	return nil
}

func (m *MockRunner) Commit(ctx context.Context, message string) error {
	// Create a new commit
	hashStr := fmt.Sprintf("mock%d", len(m.Commits)+1)
	newCommit := Commit{
		Hash:      hashStr,
		ShortHash: hashStr[:4], // "mock" is 4 chars, so this is safe
		Message:   message,
		Author:    "Demo User",
		Date:      "Just now",
		Email:     "demo@tutugit.local",
	}
	
	// Add to beginning of commits slice (HEAD)
	m.Commits = append([]Commit{newCommit}, m.Commits...)
	m.ValidHashes[newCommit.Hash] = true

	// Clear staged files
	var newFiles []FileStatus
	for _, f := range m.Files {
		if !f.Staged {
			newFiles = append(newFiles, f)
		}
	}
	m.Files = newFiles
	m.updateStatusString()

	return nil
}

// updateStatusString regenerates the m.Status text based on m.Files
func (m *MockRunner) updateStatusString() {
	var status string
	for _, f := range m.Files {
		stageChar := " "
		if f.Staged {
			if f.New {
				stageChar = "A"
			} else if f.Deleted {
				stageChar = "D"
			} else {
				stageChar = "M"
			}
		} else if f.New && !f.Staged && !f.Modified {
			stageChar = "?"
		}

		modChar := " "
		if f.Modified && !f.Staged { // Simplified for demo
			modChar = "M"
		} else if f.Deleted && !f.Staged {
			modChar = "D"
		} else if f.New && !f.Staged {
			modChar = "?"
		}

		status += fmt.Sprintf("%s%s %s\n", stageChar, modChar, f.Path)
	}
	m.Status = status
}

func (m *MockRunner) GetLastCommitHash(ctx context.Context) (string, error) {
	if len(m.Commits) > 0 {
		return m.Commits[0].Hash, nil
	}
	return "abc1234", nil
}

func (m *MockRunner) GetRemoteURL(ctx context.Context) (string, error) {
	return m.RemoteURL, nil
}

func (m *MockRunner) GetDiff(ctx context.Context, path string, staged bool) (string, error) {
	if m.GetDiffFunc != nil {
		return m.GetDiffFunc(ctx, path, staged)
	}
	return "", nil
}

func (m *MockRunner) ApplyHunk(ctx context.Context, patch string) error {
	// Let's simulate applying a hunk by staging the file if it exists in our mock list
	for i, f := range m.Files {
		if !f.Staged {
			m.Files[i].Staged = true
			m.updateStatusString()
			break
		}
	}
	return nil
}

func (m *MockRunner) GetLog(ctx context.Context, n int) ([]Commit, error) {
	if n < len(m.Commits) {
		return m.Commits[:n], nil
	}
	return m.Commits, nil
}

func (m *MockRunner) GetReflog(ctx context.Context, n int) ([]ReflogEntry, error) {
	if n < len(m.Reflog) {
		return m.Reflog[:n], nil
	}
	return m.Reflog, nil
}

func (m *MockRunner) ListWorktrees(ctx context.Context) ([]Worktree, error) {
	return m.Worktrees, nil
}

func (m *MockRunner) AddWorktree(ctx context.Context, path, branch string) error        { return nil }
func (m *MockRunner) RemoveWorktree(ctx context.Context, path string, force bool) error { return nil }

func (m *MockRunner) ResetToHash(ctx context.Context, target string) error {
	// Find the target commit
	targetIdx := -1
	for i, c := range m.Commits {
		if c.Hash == target || c.ShortHash == target {
			targetIdx = i
			break
		}
	}

	if targetIdx != -1 {
		// Slice the commits to remove everything before the target
		m.Commits = m.Commits[targetIdx:]
		
		// Add entry to reflog
		newEntry := ReflogEntry{
			Hash:     m.Commits[0].Hash, // The new HEAD
			Selector: fmt.Sprintf("HEAD@{%d}", len(m.Reflog)),
			Action:   "reset",
			Message:  "moving to " + target,
		}
		
		// Reflog is typically newest first
		m.Reflog = append([]ReflogEntry{newEntry}, m.Reflog...)
	}
	
	return nil
}

func (m *MockRunner) IsRebasing(ctx context.Context) bool {
	return m.IsRebasingVal
}

func (m *MockRunner) RebaseContinue(ctx context.Context) error { return nil }
func (m *MockRunner) RebaseAbort(ctx context.Context) error    { return nil }
func (m *MockRunner) RebaseSkip(ctx context.Context) error     { return nil }

func (m *MockRunner) GetCommitsInRange(ctx context.Context, base, head string) ([]Commit, error) {
	// Simple simulation: return all commits for now, or could filter by hash
	return m.Commits, nil
}

func (m *MockRunner) ParseStatus(ctx context.Context) ([]FileStatus, error) {
	return m.Files, nil
}

func (m *MockRunner) GetTags(ctx context.Context) ([]string, error) {
	return m.Tags, nil
}

func (m *MockRunner) ValidateHash(ctx context.Context, hash string) bool {
	return m.ValidHashes[hash]
}

func (m *MockRunner) RunInteractiveRebase(ctx context.Context, base string, steps []RebaseStep) error {
	return nil
}
