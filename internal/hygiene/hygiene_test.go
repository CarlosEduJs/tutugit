package hygiene

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"tutugit/internal/git"
	"tutugit/internal/workspace"
)

func setupTestRepo(t *testing.T) (string, func()) {
	tmpDir, err := os.MkdirTemp("", "tutugit-hygiene-test-*")
	if err != nil {
		t.Fatal(err)
	}

	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = tmpDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("git %v failed: %v", args, err)
		}
	}

	run("init")
	run("config", "user.name", "Test")
	run("config", "user.email", "test@example.com")

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return tmpDir, cleanup
}

func TestAnalyzer_GetReport(t *testing.T) {
	tmpDir, cleanup := setupTestRepo(t)
	defer cleanup()

	g := git.NewRunner(tmpDir)
	w := workspace.NewManager(tmpDir)
	w.Bootstrap()

	// Commit the bootstrap metadata so the repo starts clean
	runGit := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = tmpDir
		cmd.Run()
	}
	runGit("add", ".tutugit")
	runGit("commit", "-m", "chore: bootstrap tutugit")

	analyzer := NewAnalyzer(g, w)
	ctx := context.Background()

	// 1. Test Clean Repo
	report, err := analyzer.GetReport(ctx)
	if err != nil {
		t.Fatalf("GetReport failed: %v", err)
	}
	if report.DirtyFiles {
		t.Error("Expected no dirty files after bootstrap commit")
	}

	// 2. Test Dirty Files
	os.WriteFile(filepath.Join(tmpDir, "dirty.txt"), []byte("data"), 0644)
	report, _ = analyzer.GetReport(ctx)
	if !report.DirtyFiles {
		t.Error("Expected dirty files to be detected")
	}

	// 3. Test WIP Commits
	// Commit some WIP
	runGit("add", ".")
	runGit("commit", "-m", "wip: working on stuff")

	report, _ = analyzer.GetReport(ctx)
	if len(report.WIPCommits) == 0 {
		t.Error("Expected WIP commit to be detected")
	}

	// 4. Test Stale Workspace (Hash doesn't exist anymore)
	w.AddCommitToWorkspace("general", "nonexistenthash")
	report, _ = analyzer.GetReport(ctx)
	foundStale := false
	for _, wsName := range report.StaleWorkspaces {
		if wsName == "General" {
			foundStale = true
			break
		}
	}
	if !foundStale {
		t.Error("Expected workspace 'General' to be marked as stale due to invalid hash")
	}
}
