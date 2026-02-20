package hygiene

import (
	"context"
	"strings"
	"tutugit/internal/git"
	"tutugit/internal/workspace"
)

// HealthReport -> contains the results of repo hygiene checks.
type HealthReport struct {
	WIPCommits        []string
	SquashSuggestions []string // Workspace IDs that should be squashed
	StaleWorkspaces   []string // Workspace names with missing/rebased commits
	DirtyFiles        bool
}

// Analyzer -> performs hygiene checks on the repository.
type Analyzer struct {
	Git git.GitProvider
	WS  *workspace.Manager
}

// NewAnalyzer -> creates a new hygiene analyzer.
func NewAnalyzer(g git.GitProvider, w *workspace.Manager) *Analyzer {
	return &Analyzer{Git: g, WS: w}
}

// GetReport -> generates a hygiene report for the current state.
func (a *Analyzer) GetReport(ctx context.Context) (*HealthReport, error) {
	report := &HealthReport{}

	// this checks for dirty files
	status, err := a.Git.GetStatus(ctx)
	if err == nil && status != "" {
		report.DirtyFiles = true
	}

	// this loads metadata for further checks
	meta, err := a.WS.Load()
	if err != nil {
		return nil, err
	}

	// this detects WIP commits in metadata or recent history
	// and checks recent commits (last 10) for "wip", "fixme", "temp"
	commits, err := a.Git.Run(ctx, "log", "-n", "10", "--format=%s")
	if err == nil {
		for _, msg := range strings.Split(commits, "\n") {
			lower := strings.ToLower(msg)
			if strings.Contains(lower, "wip") || strings.Contains(lower, "fixme") || strings.Contains(lower, "temp") {
				report.WIPCommits = append(report.WIPCommits, msg)
			}
		}
	}

	// this suggests squashes for workspaces with > 3 commits
	// also checks for stale commits in all workspaces
	for _, ws := range meta.Workspaces {
		if len(ws.Commits) > 3 {
			report.SquashSuggestions = append(report.SquashSuggestions, ws.Name)
		}

		// this checks if any commit in this workspace is stale
		stale := false
		for _, hash := range ws.Commits {
			if !a.Git.ValidateHash(ctx, hash) {
				stale = true
				break
			}
		}
		if stale {
			report.StaleWorkspaces = append(report.StaleWorkspaces, ws.Name)
		}
	}

	return report, nil
}
