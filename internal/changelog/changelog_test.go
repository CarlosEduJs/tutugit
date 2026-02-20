package changelog

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"tutugit/internal/git"
	"tutugit/internal/workspace"
)

func TestFormatSummary(t *testing.T) {
	rel := &Release{
		Version: "v1.1.0",
		Entries: []ChangeEntry{
			{
				Subject:   "add new logic",
				Author:    "Carlos Edu Js",
				ShortHash: "abc1234",
				Tag:       "feature",
				Impact:    "minor",
				Workspace: "Core",
			},
			{
				Subject:   "fix nasty bug",
				Author:    "John Doe",
				ShortHash: "def5678",
				Tag:       "fix",
				Impact:    "patch",
			},
		},
	}

	g := &Generator{}
	output := g.FormatSummary([]*Release{rel})

	if !strings.Contains(output, "Release v1.1.0") {
		t.Errorf("Missing version header. Output:\n%s", output)
	}
	if !strings.Contains(output, "Impact: minor") {
		t.Errorf("Missing impact. Output:\n%s", output)
	}
	if !strings.Contains(output, "1 features") {
		t.Errorf("Missing feature count. Output:\n%s", output)
	}
	if !strings.Contains(output, "1 fixes") {
		t.Errorf("Missing fix count. Output:\n%s", output)
	}
	if !strings.Contains(output, "Workspace: Core") {
		t.Errorf("Missing workspace. Output:\n%s", output)
	}
	if !strings.Contains(output, "abc1234") {
		t.Errorf("Missing hash. Output:\n%s", output)
	}
}

func TestFormatSummary_NoEntries(t *testing.T) {
	g := &Generator{}
	output := g.FormatSummary([]*Release{})

	if !strings.Contains(output, "No releases found") {
		t.Errorf("Expected 'No releases found'. Output:\n%s", output)
	}
}

func TestExportJSON(t *testing.T) {
	rel := &Release{
		Version: "v2.0.0",
		Entries: []ChangeEntry{
			{Hash: "aaa111", ShortHash: "aaa111", Subject: "big change", Tag: "feature", Impact: "major"},
		},
	}

	g := &Generator{}
	data, err := g.ExportJSON([]*Release{rel})
	if err != nil {
		t.Fatalf("ExportJSON failed: %v", err)
	}

	var parsed []Release
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Invalid JSON output: %v", err)
	}

	if len(parsed) != 1 {
		t.Fatalf("Expected 1 release, got %d", len(parsed))
	}
	if parsed[0].Version != "v2.0.0" {
		t.Errorf("Expected version v2.0.0, got %s", parsed[0].Version)
	}
	if len(parsed[0].Entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(parsed[0].Entries))
	}
	if parsed[0].Entries[0].Impact != "major" {
		t.Errorf("Expected major impact, got %s", parsed[0].Entries[0].Impact)
	}
}
func TestExportMarkdown(t *testing.T) {
	rel := &Release{
		Version: "v1.5.0",
		Entries: []ChangeEntry{
			{Subject: "cool feature", Tag: "feature", Impact: "minor"},
		},
	}

	g := &Generator{}
	output := g.ExportMarkdown([]*Release{rel})

	if !strings.HasPrefix(output, "# Release Summary") {
		t.Errorf("Missing Markdown main title. Output:\n%s", output)
	}
	if !strings.Contains(output, "## v1.5.0") {
		t.Errorf("Missing release header. Output:\n%s", output)
	}
	if !strings.Contains(output, "**Impact:** minor") {
		t.Errorf("Missing bold impact label. Output:\n%s", output)
	}
	if !strings.Contains(output, "**feature:** cool feature") {
		t.Errorf("Missing entry format. Output:\n%s", output)
	}
}

func TestGenerateFull_NoTags(t *testing.T) {
	mockGit := &git.MockRunner{
		Tags: []string{},
		Commits: []git.Commit{
			{
				Hash:      "abc123",
				ShortHash: "abc123",
				Author:    "Test User",
				Message:   "feat: add feature",
			},
		},
	}

	mockMeta := &workspace.Meta{
		Tags: map[string][]string{
			"abc123": {"feature"},
		},
		Impacts: map[string]string{
			"abc123": "minor",
		},
	}

	g := NewGenerator(mockGit, mockMeta)
	releases, err := g.GenerateFull(context.Background())
	if err != nil {
		t.Fatalf("GenerateFull failed: %v", err)
	}

	if len(releases) != 1 {
		t.Errorf("Expected 1 release (Unreleased), got %d", len(releases))
	}

	if len(releases) > 0 && releases[0].Version != "Unreleased" {
		t.Errorf("Expected Unreleased, got %s", releases[0].Version)
	}
}

func TestGenerateFull_WithTags(t *testing.T) {
	mockGit := &git.MockRunner{
		Tags: []string{"v2.0.0", "v1.0.0"},
		Commits: []git.Commit{
			{Hash: "new123", ShortHash: "new123", Message: "feat: new feature"},
			{Hash: "mid123", ShortHash: "mid123", Message: "fix: bug fix"},
			{Hash: "old123", ShortHash: "old123", Message: "feat: initial"},
		},
	}

	mockMeta := &workspace.Meta{
		Tags:    map[string][]string{},
		Impacts: map[string]string{},
	}

	g := NewGenerator(mockGit, mockMeta)
	releases, err := g.GenerateFull(context.Background())
	if err != nil {
		t.Fatalf("GenerateFull failed: %v", err)
	}

	// Should have: Unreleased, v2.0.0, v1.0.0
	if len(releases) < 2 {
		t.Errorf("Expected at least 2 releases, got %d", len(releases))
	}

	// First should be Unreleased
	if len(releases) > 0 && releases[0].Version != "Unreleased" {
		t.Errorf("First release should be Unreleased, got %s", releases[0].Version)
	}
}

func TestGenerateFull_EmptyCommits(t *testing.T) {
	mockGit := &git.MockRunner{
		Tags:    []string{},
		Commits: []git.Commit{}, // No commits
	}

	mockMeta := &workspace.Meta{
		Tags:    map[string][]string{},
		Impacts: map[string]string{},
	}

	g := NewGenerator(mockGit, mockMeta)
	releases, err := g.GenerateFull(context.Background())
	if err != nil {
		t.Fatalf("GenerateFull failed: %v", err)
	}

	// Should have 0 releases since there are no commits
	if len(releases) != 0 {
		t.Errorf("Expected 0 releases when no commits, got %d", len(releases))
	}
}
