package changelog

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"tutugit/internal/git"
	"tutugit/internal/workspace"
)

// ChangeEntry -> represents a single normalized change in the history.
type ChangeEntry struct {
	Hash      string `json:"hash"`
	ShortHash string `json:"short_hash"`
	Author    string `json:"author"`
	Subject   string `json:"subject"`
	Tag       string `json:"tag"`
	Impact    string `json:"impact"`
	Workspace string `json:"workspace,omitempty"`
	Date      string `json:"date"`
}

// Release -> represents a versioned collection of changes.
type Release struct {
	Version string        `json:"version"`
	Date    string        `json:"date"`
	Entries []ChangeEntry `json:"entries"`
}

// Generator -> orchestrates the creation of release data from git and tutugit metadata.
type Generator struct {
	Git  git.GitProvider
	Meta *workspace.Meta
}

// NewGenerator -> creates a new generator.
func NewGenerator(g git.GitProvider, m *workspace.Meta) *Generator {
	return &Generator{Git: g, Meta: m}
}

// GenerateRelease -> collects commits for a specific range and returns a Release.
func (g *Generator) GenerateRelease(ctx context.Context, version, base, head string) (*Release, error) {
	commits, err := g.Git.GetCommitsInRange(ctx, base, head)
	if err != nil {
		return nil, err
	}

	var entries []ChangeEntry
	for _, c := range commits {
		entry := ChangeEntry{
			Hash:      c.Hash,
			ShortHash: c.ShortHash,
			Author:    c.Author,
			Subject:   c.Message,
			Date:      c.Date,
		}

		// associate with tag (nil-safe)
		if g.Meta != nil && g.Meta.Tags != nil {
			if tags, ok := g.Meta.Tags[c.Hash]; ok && len(tags) > 0 {
				entry.Tag = tags[0]
			} else {
				entry.Tag = workspace.DetectTag(c.Message)
			}
		} else {
			entry.Tag = workspace.DetectTag(c.Message)
		}

		// associate with workspace (nil-safe)
		if g.Meta != nil {
			for _, ws := range g.Meta.Workspaces {
				for _, h := range ws.Commits {
					if h == c.Hash {
						entry.Workspace = ws.Name
						break
					}
				}
				if entry.Workspace != "" {
					break
				}
			}
		}

		// associate with impact (nil-safe)
		if g.Meta != nil && g.Meta.Impacts != nil {
			if level, ok := g.Meta.Impacts[c.Hash]; ok {
				entry.Impact = level
			} else {
				entry.Impact = workspace.DetectImpact(c.Message)
			}
		} else {
			entry.Impact = workspace.DetectImpact(c.Message)
		}

		entries = append(entries, entry)
	}

	date := ""
	if len(commits) > 0 {
		date = commits[len(commits)-1].Date
	}

	return &Release{
		Version: version,
		Date:    date,
		Entries: entries,
	}, nil
}

// GenerateFull -> iterates through all tags to produce a complete release history.
func (g *Generator) GenerateFull(ctx context.Context) ([]*Release, error) {
	tags, err := g.Git.GetTags(ctx)
	if err != nil {
		return nil, err
	}

	var releases []*Release

	if len(tags) == 0 {
		rel, err := g.GenerateRelease(ctx, "Unreleased", "", "HEAD")
		if err == nil && len(rel.Entries) > 0 {
			releases = append(releases, rel)
		}
		return releases, nil
	}

	headRel, err := g.GenerateRelease(ctx, "Unreleased", tags[0], "HEAD")
	if err == nil && len(headRel.Entries) > 0 {
		releases = append(releases, headRel)
	}

	for i := 0; i < len(tags)-1; i++ {
		rel, err := g.GenerateRelease(ctx, tags[i], tags[i+1], tags[i])
		if err == nil {
			releases = append(releases, rel)
		}
	}

	if len(tags) > 0 {
		firstRel, err := g.GenerateRelease(ctx, tags[len(tags)-1], "", tags[len(tags)-1])
		if err == nil {
			releases = append(releases, firstRel)
		}
	}

	return releases, nil
}

// FormatSummary -> produces a clean, structured release summary.
func (g *Generator) FormatSummary(releases []*Release) string {
	if len(releases) == 0 {
		return "No releases found."
	}

	var b strings.Builder

	for _, rel := range releases {
		// Count by tag
		counts := make(map[string]int)
		maxImpact := "patch"
		impactWeight := map[string]int{"patch": 0, "minor": 1, "major": 2}

		for _, e := range rel.Entries {
			tag := e.Tag
			if tag == "" || tag == "none" {
				tag = "other"
			}
			counts[tag]++
			if impactWeight[e.Impact] > impactWeight[maxImpact] {
				maxImpact = e.Impact
			}
		}

		// Header
		b.WriteString(fmt.Sprintf("Release %s\n", rel.Version))
		b.WriteString(fmt.Sprintf("Impact: %s\n", maxImpact))

		// Change counts
		var parts []string
		order := []struct{ tag, label string }{
			{"feature", "features"},
			{"fix", "fixes"},
			{"refactor", "refactors"},
			{"experiment", "experiments"},
			{"other", "other"},
		}
		for _, o := range order {
			if c, ok := counts[o.tag]; ok && c > 0 {
				parts = append(parts, fmt.Sprintf("%d %s", c, o.label))
			}
		}
		b.WriteString(fmt.Sprintf("Changes: %s\n", strings.Join(parts, ", ")))

		// Workspace (if consistent)
		wsSet := make(map[string]bool)
		for _, e := range rel.Entries {
			if e.Workspace != "" {
				wsSet[e.Workspace] = true
			}
		}
		if len(wsSet) == 1 {
			for ws := range wsSet {
				b.WriteString(fmt.Sprintf("Workspace: %s\n", ws))
			}
		} else if len(wsSet) > 1 {
			var names []string
			for ws := range wsSet {
				names = append(names, ws)
			}
			b.WriteString(fmt.Sprintf("Workspaces: %s\n", strings.Join(names, ", ")))
		}

		b.WriteString("───────────────────────\n")

		// Entries
		for _, e := range rel.Entries {
			tag := e.Tag
			if tag == "" || tag == "none" {
				tag = "other"
			}
			// Pad tag for alignment
			padded := fmt.Sprintf("%-10s", tag+":")
			b.WriteString(fmt.Sprintf("  %s %s (%s)\n", padded, e.Subject, e.ShortHash))
		}
		b.WriteString("\n")
	}

	return b.String()
}

// ExportJSON -> produces structured JSON for external tool consumption.
func (g *Generator) ExportJSON(releases []*Release) ([]byte, error) {
	return json.MarshalIndent(releases, "", "  ")
}

// ExportMarkdown -> produces a human-readable Markdown summary.
func (g *Generator) ExportMarkdown(releases []*Release) string {
	if len(releases) == 0 {
		return "# Release Summary\n\nNo releases found."
	}

	var b strings.Builder
	b.WriteString("# Release Summary\n\n")

	for _, rel := range releases {
		// Count by tag
		counts := make(map[string]int)
		maxImpact := "patch"
		impactWeight := map[string]int{"patch": 0, "minor": 1, "major": 2}

		for _, e := range rel.Entries {
			tag := e.Tag
			if tag == "" || tag == "none" {
				tag = "other"
			}
			counts[tag]++
			if impactWeight[e.Impact] > impactWeight[maxImpact] {
				maxImpact = e.Impact
			}
		}

		// Release Header
		b.WriteString(fmt.Sprintf("## %s\n", rel.Version))
		b.WriteString(fmt.Sprintf("- **Impact:** %s\n", maxImpact))

		// Change counts
		var parts []string
		order := []struct{ tag, label string }{
			{"feature", "features"},
			{"fix", "fixes"},
			{"refactor", "refactors"},
			{"experiment", "experiments"},
			{"other", "other"},
		}
		for _, o := range order {
			if c, ok := counts[o.tag]; ok && c > 0 {
				parts = append(parts, fmt.Sprintf("%d %s", c, o.label))
			}
		}
		b.WriteString(fmt.Sprintf("- **Changes:** %s\n", strings.Join(parts, ", ")))

		// Workspace (if consistent)
		wsSet := make(map[string]bool)
		for _, e := range rel.Entries {
			if e.Workspace != "" {
				wsSet[e.Workspace] = true
			}
		}
		if len(wsSet) == 1 {
			for ws := range wsSet {
				b.WriteString(fmt.Sprintf("- **Workspace:** %s\n", ws))
			}
		} else if len(wsSet) > 1 {
			var names []string
			for ws := range wsSet {
				names = append(names, ws)
			}
			b.WriteString(fmt.Sprintf("- **Workspaces:** %s\n", strings.Join(names, ", ")))
		}

		b.WriteString("\n---\n\n")

		// Entries
		for _, e := range rel.Entries {
			tag := e.Tag
			if tag == "" || tag == "none" {
				tag = "other"
			}
			b.WriteString(fmt.Sprintf("- **%s:** %s (`%s`)\n", tag, e.Subject, e.ShortHash))
		}
		b.WriteString("\n")
	}

	return b.String()
}
