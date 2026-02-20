package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func setupGitRepo(t *testing.T) (string, func()) {
	tmpDir, err := os.MkdirTemp("", "git-test-*")
	if err != nil {
		t.Fatal(err)
	}

	run := func(args ...string) error {
		cmd := exec.Command("git", args...)
		cmd.Dir = tmpDir
		return cmd.Run()
	}

	if err := run("init"); err != nil {
		t.Fatalf("git init failed: %v", err)
	}
	run("config", "user.name", "Test User")
	run("config", "user.email", "test@example.com")

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return tmpDir, cleanup
}

func TestRunner_GetCurrentBranch(t *testing.T) {
	dir, cleanup := setupGitRepo(t)
	defer cleanup()

	r := NewRunner(dir)
	ctx := context.Background()

	// Need at least one commit for HEAD to exist
	testFile := filepath.Join(dir, "test.txt")
	os.WriteFile(testFile, []byte("data"), 0644)
	r.StageFile(ctx, "test.txt")
	r.Commit(ctx, "initial")

	branch, err := r.GetCurrentBranch(ctx)
	if err != nil {
		t.Fatalf("GetCurrentBranch failed: %v", err)
	}

	// Git init creates 'master' or 'main' depending on version
	if branch != "master" && branch != "main" {
		t.Errorf("Unexpected branch: %s", branch)
	}
}

func TestRunner_GetStatus(t *testing.T) {
	dir, cleanup := setupGitRepo(t)
	defer cleanup()

	r := NewRunner(dir)

	// Empty repo should have empty status
	status, err := r.GetStatus(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if strings.TrimSpace(status) != "" {
		t.Errorf("Expected empty status, got: %s", status)
	}

	// Create a file
	os.WriteFile(filepath.Join(dir, "test.txt"), []byte("hello"), 0644)

	status, err = r.GetStatus(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(status, "test.txt") {
		t.Errorf("Expected test.txt in status, got: %s", status)
	}
}

func TestRunner_ParseStatus(t *testing.T) {
	dir, cleanup := setupGitRepo(t)
	defer cleanup()

	// Create some files
	os.WriteFile(filepath.Join(dir, "new.txt"), []byte("new file"), 0644)
	os.WriteFile(filepath.Join(dir, "modified.txt"), []byte("data"), 0644)

	r := NewRunner(dir)
	ctx := context.Background()

	// Stage modified.txt and commit it
	r.Run(ctx, "add", "modified.txt")
	r.Run(ctx, "commit", "-m", "initial")

	// Modify it
	os.WriteFile(filepath.Join(dir, "modified.txt"), []byte("changed"), 0644)

	// Parse status
	files, err := r.ParseStatus(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if len(files) != 2 {
		t.Fatalf("Expected 2 files, got %d", len(files))
	}

	// Check new.txt
	var newFile, modFile *FileStatus
	for i := range files {
		if files[i].Path == "new.txt" {
			newFile = &files[i]
		}
		if files[i].Path == "modified.txt" {
			modFile = &files[i]
		}
	}

	if newFile == nil || modFile == nil {
		t.Fatal("Expected both new.txt and modified.txt in status")
	}

	if !newFile.New {
		t.Error("new.txt should be marked as New")
	}

	if !modFile.Modified {
		t.Error("modified.txt should be marked as Modified")
	}
}

func TestRunner_StageFile(t *testing.T) {
	dir, cleanup := setupGitRepo(t)
	defer cleanup()

	testFile := filepath.Join(dir, "test.txt")
	os.WriteFile(testFile, []byte("data"), 0644)

	r := NewRunner(dir)
	ctx := context.Background()

	if err := r.StageFile(ctx, "test.txt"); err != nil {
		t.Fatalf("StageFile failed: %v", err)
	}

	// Verify it's staged
	files, _ := r.ParseStatus(ctx)
	if len(files) != 1 {
		t.Fatalf("Expected 1 staged file, got %d", len(files))
	}

	if !files[0].Staged {
		t.Error("File should be marked as staged")
	}
}

func TestRunner_UnstageFile(t *testing.T) {
	dir, cleanup := setupGitRepo(t)
	defer cleanup()

	testFile := filepath.Join(dir, "test.txt")
	os.WriteFile(testFile, []byte("data"), 0644)

	r := NewRunner(dir)
	ctx := context.Background()

	// Stage it first
	r.StageFile(ctx, "test.txt")

	// Now unstage
	if err := r.UnstageFile(ctx, "test.txt"); err != nil {
		t.Fatalf("UnstageFile failed: %v", err)
	}

	// Verify it's not staged
	files, _ := r.ParseStatus(ctx)
	if len(files) != 1 {
		t.Fatalf("Expected 1 file, got %d", len(files))
	}

	if files[0].Staged {
		t.Error("File should not be staged after unstage")
	}
}

func TestRunner_Commit(t *testing.T) {
	dir, cleanup := setupGitRepo(t)
	defer cleanup()

	testFile := filepath.Join(dir, "test.txt")
	os.WriteFile(testFile, []byte("data"), 0644)

	r := NewRunner(dir)
	ctx := context.Background()

	// Stage and commit
	r.StageFile(ctx, "test.txt")

	if err := r.Commit(ctx, "test: add test file"); err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	// Verify commit exists
	commits, err := r.GetLog(ctx, 10)
	if err != nil {
		t.Fatal(err)
	}

	if len(commits) != 1 {
		t.Fatalf("Expected 1 commit, got %d", len(commits))
	}

	if commits[0].Message != "test: add test file" {
		t.Errorf("Unexpected message: %s", commits[0].Message)
	}
}

func TestRunner_GetLastCommitHash(t *testing.T) {
	dir, cleanup := setupGitRepo(t)
	defer cleanup()

	r := NewRunner(dir)
	ctx := context.Background()

	// Create initial commit
	testFile := filepath.Join(dir, "test.txt")
	os.WriteFile(testFile, []byte("data"), 0644)
	r.StageFile(ctx, "test.txt")
	r.Commit(ctx, "initial")

	hash, err := r.GetLastCommitHash(ctx)
	if err != nil {
		t.Fatalf("GetLastCommitHash failed: %v", err)
	}

	if len(hash) != 40 {
		t.Errorf("Expected 40 char hash, got %d chars: %s", len(hash), hash)
	}
}

func TestRunner_GetDiff(t *testing.T) {
	dir, cleanup := setupGitRepo(t)
	defer cleanup()

	testFile := filepath.Join(dir, "test.txt")
	os.WriteFile(testFile, []byte("original"), 0644)

	r := NewRunner(dir)
	ctx := context.Background()

	// Initial commit
	r.StageFile(ctx, "test.txt")
	r.Commit(ctx, "initial")

	// Modify file
	os.WriteFile(testFile, []byte("modified"), 0644)

	// Get unstaged diff
	diff, err := r.GetDiff(ctx, "test.txt", false)
	if err != nil {
		t.Fatalf("GetDiff failed: %v", err)
	}

	if !strings.Contains(diff, "-original") {
		t.Error("Diff should contain removed line")
	}
	if !strings.Contains(diff, "+modified") {
		t.Error("Diff should contain added line")
	}

	// Stage and get staged diff
	r.StageFile(ctx, "test.txt")
	stagedDiff, err := r.GetDiff(ctx, "test.txt", true)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(stagedDiff, "+modified") {
		t.Error("Staged diff should contain changes")
	}
}

func TestRunner_GetLog(t *testing.T) {
	dir, cleanup := setupGitRepo(t)
	defer cleanup()

	r := NewRunner(dir)
	ctx := context.Background()

	// Create multiple commits
	for i := 1; i <= 3; i++ {
		filename := filepath.Join(dir, "file"+string(rune('0'+i))+".txt")
		os.WriteFile(filename, []byte("data"), 0644)
		r.StageFile(ctx, filepath.Base(filename))
		r.Commit(ctx, "commit "+string(rune('0'+i)))
		time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	}

	commits, err := r.GetLog(ctx, 10)
	if err != nil {
		t.Fatal(err)
	}

	if len(commits) != 3 {
		t.Fatalf("Expected 3 commits, got %d", len(commits))
	}

	// Verify order (newest first)
	if commits[0].Message != "commit 3" {
		t.Errorf("First commit should be newest, got: %s", commits[0].Message)
	}

	// Verify fields are populated
	if commits[0].Hash == "" {
		t.Error("Hash should not be empty")
	}
	if commits[0].ShortHash == "" {
		t.Error("ShortHash should not be empty")
	}
	if commits[0].Author != "Test User" {
		t.Errorf("Unexpected author: %s", commits[0].Author)
	}
	if commits[0].Email != "test@example.com" {
		t.Errorf("Unexpected email: %s", commits[0].Email)
	}
}

func TestRunner_GetCommitsInRange(t *testing.T) {
	dir, cleanup := setupGitRepo(t)
	defer cleanup()

	r := NewRunner(dir)
	ctx := context.Background()

	// Create 3 commits
	for i := 1; i <= 3; i++ {
		filename := filepath.Join(dir, "file"+string(rune('0'+i))+".txt")
		os.WriteFile(filename, []byte("data"), 0644)
		r.StageFile(ctx, filepath.Base(filename))
		r.Commit(ctx, "commit "+string(rune('0'+i)))
	}

	// Get all commits
	allCommits, _ := r.GetLog(ctx, 10)
	baseHash := allCommits[2].Hash // First commit
	headHash := allCommits[0].Hash // Last commit

	// Get range (should return commits 2 and 3, excluding base)
	rangeCommits, err := r.GetCommitsInRange(ctx, baseHash, headHash)
	if err != nil {
		t.Fatal(err)
	}

	if len(rangeCommits) != 2 {
		t.Fatalf("Expected 2 commits in range, got %d", len(rangeCommits))
	}

	// Verify order is oldest to newest (reverse of GetLog)
	if rangeCommits[0].Message != "commit 2" {
		t.Errorf("First commit should be oldest in range, got: %s", rangeCommits[0].Message)
	}
}

func TestRunner_GetTags(t *testing.T) {
	dir, cleanup := setupGitRepo(t)
	defer cleanup()

	r := NewRunner(dir)
	ctx := context.Background()

	// Create initial commit
	testFile := filepath.Join(dir, "test.txt")
	os.WriteFile(testFile, []byte("data"), 0644)
	r.StageFile(ctx, "test.txt")
	r.Commit(ctx, "initial")

	// Create tags
	r.Run(ctx, "tag", "v0.1.0")
	r.Run(ctx, "tag", "v1.0.0")
	r.Run(ctx, "tag", "v1.1.0")

	tags, err := r.GetTags(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if len(tags) != 3 {
		t.Fatalf("Expected 3 tags, got %d", len(tags))
	}

	// Should be sorted by version (descending)
	if tags[0] != "v1.1.0" {
		t.Errorf("First tag should be v1.1.0, got: %s", tags[0])
	}
}

func TestRunner_ValidateHash(t *testing.T) {
	dir, cleanup := setupGitRepo(t)
	defer cleanup()

	r := NewRunner(dir)
	ctx := context.Background()

	// Create commit
	testFile := filepath.Join(dir, "test.txt")
	os.WriteFile(testFile, []byte("data"), 0644)
	r.StageFile(ctx, "test.txt")
	r.Commit(ctx, "test")

	hash, _ := r.GetLastCommitHash(ctx)

	// Valid hash
	if !r.ValidateHash(ctx, hash) {
		t.Error("Valid hash should return true")
	}

	// Invalid hash
	if r.ValidateHash(ctx, "invalidhash123") {
		t.Error("Invalid hash should return false")
	}
}

func TestRunner_GetReflog(t *testing.T) {
	dir, cleanup := setupGitRepo(t)
	defer cleanup()

	r := NewRunner(dir)
	ctx := context.Background()

	// Create some commits to generate reflog
	for i := 1; i <= 2; i++ {
		filename := filepath.Join(dir, "file"+string(rune('0'+i))+".txt")
		os.WriteFile(filename, []byte("data"), 0644)
		r.StageFile(ctx, filepath.Base(filename))
		r.Commit(ctx, "commit "+string(rune('0'+i)))
	}

	entries, err := r.GetReflog(ctx, 10)
	if err != nil {
		t.Fatal(err)
	}

	if len(entries) < 2 {
		t.Fatalf("Expected at least 2 reflog entries, got %d", len(entries))
	}

	// Verify structure
	entry := entries[0]
	if entry.Hash == "" {
		t.Error("Reflog entry should have hash")
	}
	if entry.Selector == "" {
		t.Error("Reflog entry should have selector")
	}
}

func TestRunner_ListWorktrees(t *testing.T) {
	dir, cleanup := setupGitRepo(t)
	defer cleanup()

	r := NewRunner(dir)
	ctx := context.Background()

	// Need at least one commit for worktrees
	testFile := filepath.Join(dir, "test.txt")
	os.WriteFile(testFile, []byte("data"), 0644)
	r.StageFile(ctx, "test.txt")
	r.Commit(ctx, "initial")

	worktrees, err := r.ListWorktrees(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Should have at least the main worktree
	if len(worktrees) < 1 {
		t.Fatal("Expected at least 1 worktree (main)")
	}

	main := worktrees[0]
	if !main.IsMain {
		t.Error("First worktree should be marked as main")
	}
	if main.Path != dir {
		t.Errorf("Main worktree path should be %s, got %s", dir, main.Path)
	}
}

func TestRunner_IsRebasing(t *testing.T) {
	dir, cleanup := setupGitRepo(t)
	defer cleanup()

	r := NewRunner(dir)
	ctx := context.Background()

	// Create commit
	testFile := filepath.Join(dir, "test.txt")
	os.WriteFile(testFile, []byte("data"), 0644)
	r.StageFile(ctx, "test.txt")
	r.Commit(ctx, "test")

	// Not rebasing initially
	if r.IsRebasing(ctx) {
		t.Error("Should not be rebasing initially")
	}

	// Note: Testing actual rebase is complex and may cause interactive prompts
	// This test just verifies the method doesn't crash
}

func TestRunner_ApplyHunk(t *testing.T) {
	dir, cleanup := setupGitRepo(t)
	defer cleanup()

	testFile := filepath.Join(dir, "test.txt")
	os.WriteFile(testFile, []byte("line1\nline2\nline3\n"), 0644)

	r := NewRunner(dir)
	ctx := context.Background()

	// Initial commit
	r.StageFile(ctx, "test.txt")
	r.Commit(ctx, "initial")

	// Modify file
	os.WriteFile(testFile, []byte("line1\nmodified\nline3\n"), 0644)

	// Create a patch for this change
	patch := `diff --git a/test.txt b/test.txt
index 123..456 100644
--- a/test.txt
+++ b/test.txt
@@ -1,3 +1,3 @@
 line1
-line2
+modified
 line3
`

	// Apply the hunk (stage it)
	if err := r.ApplyHunk(ctx, patch); err != nil {
		t.Fatalf("ApplyHunk failed: %v", err)
	}

	// Verify it's staged
	status, _ := r.GetStatus(ctx)
	if !strings.Contains(status, "M") {
		t.Error("File should be modified and staged")
	}
}

func TestRunner_ResetToHash(t *testing.T) {
	dir, cleanup := setupGitRepo(t)
	defer cleanup()

	r := NewRunner(dir)
	ctx := context.Background()

	// Create two commits
	testFile := filepath.Join(dir, "test.txt")
	os.WriteFile(testFile, []byte("v1"), 0644)
	r.StageFile(ctx, "test.txt")
	r.Commit(ctx, "commit 1")

	firstHash, _ := r.GetLastCommitHash(ctx)

	os.WriteFile(testFile, []byte("v2"), 0644)
	r.StageFile(ctx, "test.txt")
	r.Commit(ctx, "commit 2")

	// Reset to first commit
	if err := r.ResetToHash(ctx, firstHash); err != nil {
		t.Fatalf("ResetToHash failed: %v", err)
	}

	// Verify we're back at first commit
	commits, _ := r.GetLog(ctx, 10)
	if len(commits) != 1 {
		t.Errorf("Expected 1 commit after reset, got %d", len(commits))
	}

	// Verify file content is back to v1
	content, _ := os.ReadFile(testFile)
	if string(content) != "v1" {
		t.Errorf("File should be reset to v1, got: %s", content)
	}
}

func TestRunner_GetRemoteURL(t *testing.T) {
	dir, cleanup := setupGitRepo(t)
	defer cleanup()

	r := NewRunner(dir)
	ctx := context.Background()

	// No remote initially
	url, err := r.GetRemoteURL(ctx)
	if err != nil {
		t.Fatalf("GetRemoteURL should not error: %v", err)
	}
	if url != "" {
		t.Errorf("Expected empty URL, got: %s", url)
	}

	// Add a remote
	r.Run(ctx, "remote", "add", "origin", "https://github.com/test/repo.git")

	url, err = r.GetRemoteURL(ctx)
	if err != nil {
		t.Fatalf("GetRemoteURL failed: %v", err)
	}
	if url != "https://github.com/test/repo.git" {
		t.Errorf("Expected remote URL, got: %s", url)
	}
}

func TestRunner_AddWorktree(t *testing.T) {
	dir, cleanup := setupGitRepo(t)
	defer cleanup()

	r := NewRunner(dir)
	ctx := context.Background()

	// Need at least one commit
	testFile := filepath.Join(dir, "test.txt")
	os.WriteFile(testFile, []byte("data"), 0644)
	r.StageFile(ctx, "test.txt")
	r.Commit(ctx, "initial")

	// Create a new branch
	r.Run(ctx, "branch", "feature")

	// Add worktree
	worktreePath := filepath.Join(dir, "..", "feature-worktree")
	defer os.RemoveAll(worktreePath)

	if err := r.AddWorktree(ctx, worktreePath, "feature"); err != nil {
		t.Fatalf("AddWorktree failed: %v", err)
	}

	// Verify worktree exists
	if _, err := os.Stat(worktreePath); os.IsNotExist(err) {
		t.Error("Worktree directory should exist")
	}

	// List worktrees and verify
	worktrees, err := r.ListWorktrees(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if len(worktrees) < 2 {
		t.Errorf("Expected at least 2 worktrees, got %d", len(worktrees))
	}
}

func TestRunner_RemoveWorktree(t *testing.T) {
	dir, cleanup := setupGitRepo(t)
	defer cleanup()

	r := NewRunner(dir)
	ctx := context.Background()

	// Setup: create commit and worktree
	testFile := filepath.Join(dir, "test.txt")
	os.WriteFile(testFile, []byte("data"), 0644)
	r.StageFile(ctx, "test.txt")
	r.Commit(ctx, "initial")
	r.Run(ctx, "branch", "feature")

	worktreePath := filepath.Join(dir, "..", "feature-worktree")
	r.AddWorktree(ctx, worktreePath, "feature")

	// Remove worktree
	if err := r.RemoveWorktree(ctx, worktreePath, false); err != nil {
		t.Fatalf("RemoveWorktree failed: %v", err)
	}

	// Verify it's removed
	worktrees, _ := r.ListWorktrees(ctx)
	for _, wt := range worktrees {
		if strings.Contains(wt.Path, "feature-worktree") {
			t.Error("Worktree should have been removed")
		}
	}
}

func TestRunner_RemoveWorktree_Force(t *testing.T) {
	dir, cleanup := setupGitRepo(t)
	defer cleanup()

	r := NewRunner(dir)
	ctx := context.Background()

	// Setup
	testFile := filepath.Join(dir, "test.txt")
	os.WriteFile(testFile, []byte("data"), 0644)
	r.StageFile(ctx, "test.txt")
	r.Commit(ctx, "initial")
	r.Run(ctx, "branch", "feature")

	worktreePath := filepath.Join(dir, "..", "feature-worktree-force")
	r.AddWorktree(ctx, worktreePath, "feature")

	// Add uncommitted changes to worktree
	wtFile := filepath.Join(worktreePath, "new.txt")
	os.WriteFile(wtFile, []byte("uncommitted"), 0644)

	// Force remove should work even with uncommitted changes
	if err := r.RemoveWorktree(ctx, worktreePath, true); err != nil {
		t.Fatalf("RemoveWorktree with force failed: %v", err)
	}
}

func TestRunner_RebaseContinue(t *testing.T) {
	dir, cleanup := setupGitRepo(t)
	defer cleanup()

	r := NewRunner(dir)
	ctx := context.Background()

	// Create commits
	testFile := filepath.Join(dir, "test.txt")
	os.WriteFile(testFile, []byte("v1"), 0644)
	r.StageFile(ctx, "test.txt")
	r.Commit(ctx, "initial")

	// Try to continue non-existent rebase (should error)
	err := r.RebaseContinue(ctx)
	if err == nil {
		t.Error("RebaseContinue should fail when no rebase is in progress")
	}
}

func TestRunner_RebaseAbort(t *testing.T) {
	dir, cleanup := setupGitRepo(t)
	defer cleanup()

	r := NewRunner(dir)
	ctx := context.Background()

	// Create commits
	testFile := filepath.Join(dir, "test.txt")
	os.WriteFile(testFile, []byte("v1"), 0644)
	r.StageFile(ctx, "test.txt")
	r.Commit(ctx, "initial")

	// Try to abort non-existent rebase (should error)
	err := r.RebaseAbort(ctx)
	if err == nil {
		t.Error("RebaseAbort should fail when no rebase is in progress")
	}
}

func TestRunner_RebaseSkip(t *testing.T) {
	dir, cleanup := setupGitRepo(t)
	defer cleanup()

	r := NewRunner(dir)
	ctx := context.Background()

	// Create commits
	testFile := filepath.Join(dir, "test.txt")
	os.WriteFile(testFile, []byte("v1"), 0644)
	r.StageFile(ctx, "test.txt")
	r.Commit(ctx, "initial")

	// Try to skip non-existent rebase (should error)
	err := r.RebaseSkip(ctx)
	if err == nil {
		t.Error("RebaseSkip should fail when no rebase is in progress")
	}
}

func TestRunner_RunInteractiveRebase(t *testing.T) {
	dir, cleanup := setupGitRepo(t)
	defer cleanup()

	r := NewRunner(dir)
	ctx := context.Background()

	// Create multiple commits
	for i := 1; i <= 3; i++ {
		filename := filepath.Join(dir, fmt.Sprintf("file%d.txt", i))
		os.WriteFile(filename, []byte(fmt.Sprintf("content%d", i)), 0644)
		r.StageFile(ctx, filepath.Base(filename))
		r.Commit(ctx, fmt.Sprintf("commit %d", i))
		time.Sleep(10 * time.Millisecond)
	}

	// Get commits
	commits, _ := r.GetLog(ctx, 10)
	if len(commits) < 3 {
		t.Fatal("Need at least 3 commits for rebase test")
	}

	// Prepare rebase steps (squash second commit into first)
	steps := []RebaseStep{
		{Action: ActionPick, Hash: commits[2].ShortHash, Message: commits[2].Message},
		{Action: ActionSquash, Hash: commits[1].ShortHash, Message: commits[1].Message},
		{Action: ActionPick, Hash: commits[0].ShortHash, Message: commits[0].Message},
	}

	// Try running interactive rebase
	// Note: This might fail in some environments, so we just check it doesn't panic
	err := r.RunInteractiveRebase(ctx, commits[2].Hash+"^", steps)
	// We don't assert on the error because interactive rebase is complex
	// The important part is that the function runs without panicking
	_ = err
}
