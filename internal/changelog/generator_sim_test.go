package changelog

import (
	"context"
	"strings"
	"testing"
	"tutugit/internal/git"
	"tutugit/internal/workspace"
)

func TestGenerator_Simulation(t *testing.T) {
	// setup Mock Git with a complex history
	mock := git.NewMockRunner()

	// create some commits
	mock.Commits = []git.Commit{
		{Hash: "hash1", ShortHash: "h1", Message: "feat: add user login", Author: "Carlos", Date: "1 hour ago"},
		{Hash: "hash2", ShortHash: "h2", Message: "fix: crash on logout", Author: "Carlos", Date: "45 mins ago"},
		{Hash: "hash3", ShortHash: "h3", Message: "refactor: optimize database", Author: "John", Date: "30 mins ago"},
		{Hash: "hash4", ShortHash: "h4", Message: "experiment: test new api", Author: "Carlos", Date: "10 mins ago"},
		{Hash: "hash5", ShortHash: "h5", Message: "chore: update docs", Author: "Carlos", Date: "5 mins ago"},
	}

	// setup Meta with Workspace grouping
	meta := &workspace.Meta{
		Workspaces: []workspace.Workspace{
			{
				ID:      "auth",
				Name:    "Auth System",
				Commits: []string{"hash1", "hash2"},
			},
			{
				ID:      "db",
				Name:    "Database",
				Commits: []string{"hash3"},
			},
		},
		Tags: make(map[string][]string),
	}

	// generate Release
	gen := NewGenerator(mock, meta)
	ctx := context.Background()

	rel, err := gen.GenerateRelease(ctx, "v1.0.0", "base", "head")
	if err != nil {
		t.Fatalf("GenerateRelease failed: %v", err)
	}

	// validate Semantic Logic
	if rel.Version != "v1.0.0" {
		t.Errorf("Expected version v1.0.0, got %s", rel.Version)
	}

	// find the workspace-grouped commits
	foundAuth := false
	foundDB := false
	foundOther := false

	for _, e := range rel.Entries {
		switch e.Hash {
		case "hash1":
			if e.Workspace != "Auth System" {
				t.Errorf("hash1 should be in Auth System, got %s", e.Workspace)
			}
			if e.Tag != "feature" {
				t.Errorf("hash1 should have tag feature, got %s", e.Tag)
			}
			foundAuth = true
		case "hash3":
			if e.Workspace != "Database" {
				t.Errorf("hash3 should be in Database, got %s", e.Workspace)
			}
			foundDB = true
		case "hash5":
			if e.Workspace != "" {
				t.Errorf("hash5 should have no workspace, got %s", e.Workspace)
			}
			foundOther = true
		}
	}

	if !foundAuth || !foundDB || !foundOther {
		t.Error("Did not find all expected commits in semantic mapping")
	}

	// build Final Markdown Summary
	summary := gen.ExportMarkdown([]*Release{rel})

	// check if Markdown reflects the semantic intelligence
	if !strings.Contains(summary, "## v1.0.0") {
		t.Error("Markdown missing version header")
	}
	if !strings.Contains(summary, "1 features, 1 fixes, 1 refactors") {
		t.Errorf("Markdown counts incorrect. Summary:\n%s", summary)
	}
	if !strings.Contains(summary, "Auth System") {
		t.Error("Markdown missing workspace names")
	}
}
