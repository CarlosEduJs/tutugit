package main

import (
	"context"
	"tutugit/internal/config"
	"tutugit/internal/git"
	"tutugit/internal/hygiene"
	"tutugit/internal/workspace"

	"github.com/charmbracelet/bubbles/textinput"
)

// initialDemoModel creates a playground state for TUI testing.
func initialDemoModel() model {
	mock := git.NewMockRunner()

	// mock History
	mock.Commits = []git.Commit{
		{Hash: "sha7", ShortHash: "abc777", Message: "docs: update API documentation", Author: "Alice", Date: "2 mins ago", Email: "alice@example.com"},
		{Hash: "sha6", ShortHash: "abc666", Message: "fix: prevent memory leak in workspace manager", Author: "Bob", Date: "5 mins ago", Email: "bob@example.com"},
		{Hash: "sha5", ShortHash: "abc555", Message: "feat: add demo mode toggle", Author: "Carlos", Date: "10 mins ago", Email: "carlos@example.com"},
		{Hash: "sha4", ShortHash: "abc444", Message: "wip: incomplete work on styles", Author: "Dev", Date: "1 hour ago", Email: "dev@example.com"},
		{Hash: "sha3", ShortHash: "abc333", Message: "refactor: clean up viewport logic", Author: "Carlos", Date: "5 hours ago", Email: "carlos@example.com"},
		{Hash: "sha2", ShortHash: "abc222", Message: "fix: resolve scary panic in update.go", Author: "John", Date: "1 day ago", Email: "john@example.com"},
		{Hash: "sha1", ShortHash: "abc111", Message: "feat: add super cool workspace grouping", Author: "Carlos", Date: "2 days ago", Email: "carlos@example.com"},
		{Hash: "sha0", ShortHash: "abc000", Message: "chore: initial commit", Author: "Carlos", Date: "1 week ago", Email: "carlos@example.com"},
	}
	mock.ValidHashes = map[string]bool{
		"sha7": true, "sha6": true, "sha5": true, "sha4": true, "sha3": true, "sha2": true, "sha1": true, "sha0": true,
	}

	// mock Status
	mock.Files = []git.FileStatus{
		{Path: "cmd/tutugit/main.go", Staged: true, Modified: true},
		{Path: "cmd/tutugit/demo.go", Staged: false, Modified: true},
		{Path: "internal/git/mock.go", Staged: true, Modified: false},
		{Path: "internal/hygiene/checker.go", Staged: false, Modified: true},
		{Path: "config.yml", Staged: false, Modified: true}, // Unstaged mod
		{Path: "README.md", New: true, Staged: true},
		{Path: "DELETED_FILE.txt", Deleted: true, Staged: false},
		{Path: "untracked_config.json", New: true, Staged: false},
	}
	// Initial status generation
	mock.Status = "M  cmd/tutugit/main.go\n M cmd/tutugit/demo.go\nM  internal/git/mock.go\n M internal/hygiene/checker.go\n M config.yml\nA  README.md\n D DELETED_FILE.txt\n?? untracked_config.json\n"
	mock.CurrentBranch = "feature/demo-mode"

	// mock Diffs
	mock.GetDiffFunc = func(ctx context.Context, path string, staged bool) (string, error) {
		switch path {
		case "cmd/tutugit/main.go":
			return "--- a/cmd/tutugit/main.go\n+++ b/cmd/tutugit/main.go\n@@ -10,6 +10,7 @@\n-import \"fmt\"\n+import (\n+    \"fmt\"\n+    \"os\"\n+)", nil
		case "cmd/tutugit/demo.go":
			return "--- a/cmd/tutugit/demo.go\n+++ b/cmd/tutugit/demo.go\n@@ -1,3 +1,10 @@\n package main\n \n-func demo() {}\n+func initialDemoModel() model {\n+    // magic simulation\n+}\n@@ -20,2 +27,5 @@\n func Process() {\n-   return\n+   fmt.Println(\"Processing demo...\")\n+   fmt.Println(\"Done.\")\n }", nil
		case "config.yml":
			return "--- a/config.yml\n+++ b/config.yml\n@@ -2,3 +2,4 @@\n project:\n-  name: \"tutugit\"\n+  name: \"tutugit plus\"\n+  theme: \"dark\"", nil
		default:
			return "diff --git a/" + path + " b/" + path + "\nindex 1234567..890abcd 100644\n--- a/" + path + "\n+++ b/" + path + "\n@@ -1,1 +1,2 @@\n-old content\n+new awesome content\n+another line", nil
		}
	}

	// mock Reflog
	mock.Reflog = []git.ReflogEntry{
		{Hash: "sha5", Selector: "HEAD@{0}", Action: "commit", Message: "feat: add demo mode toggle"},
		{Hash: "sha4", Selector: "HEAD@{1}", Action: "rebase (finish)", Message: "returning to status"},
		{Hash: "sha1", Selector: "HEAD@{2}", Action: "checkout", Message: "from main to feature/demo-mode"},
		{Hash: "sha2", Selector: "HEAD@{3}", Action: "reset", Message: "moving to sha2"},
	}

	// mock Worktrees
	mock.Worktrees = []git.Worktree{
		{Path: "/home/user/tutugit", Branch: "feature/demo-mode", Hash: "sha5", IsMain: true},
		{Path: "/tmp/tutugit-fix", Branch: "hotfix/emergency", Hash: "sha2", IsMain: false},
	}

	// mock Rebase state
	mock.IsRebasingVal = true
	mock.RebaseTodo = []git.RebaseStep{
		{Action: git.ActionPick, Hash: "sha3", Message: "refactor: clean up viewport logic"},
		{Action: git.ActionEdit, Hash: "sha4", Message: "wip: incomplete work on styles"},
		{Action: git.ActionPick, Hash: "sha5", Message: "feat: add demo mode toggle"},
	}

	// mock Workspaces
	meta := &workspace.Meta{
		Version: 1,
		Workspaces: []workspace.Workspace{
			{ID: "ws1", Name: "UI Refactor", Commits: []string{"sha1", "sha3", "sha5"}},
			{ID: "ws2", Name: "Bug Hunting", Commits: []string{"sha2"}},
			{ID: "ws3", Name: "Experiments", Commits: []string{"sha4"}},
		},
		ActiveWorkspace: "ws1",
		Tags: map[string][]string{
			"sha1": {"feature"},
			"sha2": {"fix"},
			"sha3": {"refactor"},
			"sha5": {"feature"},
		},
		Impacts: map[string]string{
			"sha5": "minor",
			"sha4": "patch",
			"sha1": "major",
		},
	}

	// initialize Inputs and Viewports (reusing styles and defaults)
	// this use a real manager but it won't find files if we point to a temp dir
	w := workspace.NewManager("/tmp/tutugit-demo")

	ti := textinput.New()
	ti.Placeholder = "Demo commit..."
	wn := textinput.New()
	wd := textinput.New()

	vp := newStyledViewport("62")
	hp := newStyledViewport("63")
	rp := newStyledViewport("64")
	sp := newStyledViewport("39")

	return model{
		git:             mock,
		wsManager:       w,
		cfgManager:      nil,
		cfg:             config.DefaultConfig(),
		hygiene:         hygiene.NewAnalyzer(mock, w),
		meta:            meta,
		branch:          "demo-branch",
		files:           mock.Files,
		commits:         mock.Commits,
		reflogEntries:   mock.Reflog,
		worktrees:       mock.Worktrees,
		rebaseSteps:     mock.RebaseTodo,
		isRebasing:      true,
		state:           stateMain,
		commitMsg:       ti,
		newWsName:       wn,
		newWsDesc:       wd,
		expandedFile:    -1,
		diffViewport:    vp,
		historyViewport: hp,
		reflogViewport:  rp,
		summaryViewport: sp,
		expandedHistory: make(map[string]bool),
		isUpdating:      false,
		decidedImpact:   "patch",
	}
}
