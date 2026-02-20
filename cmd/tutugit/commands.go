package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"tutugit/internal/changelog"
	"tutugit/internal/diff"
	"tutugit/internal/git"
	"tutugit/internal/workspace"

	tea "github.com/charmbracelet/bubbletea"
)

// Summary commands
func (m model) fetchSummary() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		gen := changelog.NewGenerator(m.git, m.meta)
		rels, err := gen.GenerateFull(ctx)
		if err != nil {
			return errMsg(err)
		}
		content := gen.FormatSummary(rels)
		return summaryMsg(content)
	}
}

func (m model) exportMarkdown() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		gen := changelog.NewGenerator(m.git, m.meta)
		rels, err := gen.GenerateFull(ctx)
		if err != nil {
			return errMsg(err)
		}
		data := gen.ExportMarkdown(rels)
		path := filepath.Join(".tutugit", "release.md")
		if err := os.WriteFile(path, []byte(data), 0644); err != nil {
			return errMsg(err)
		}
		return successMsg("Exported to .tutugit/release.md!")
	}
}

// Fetch commands for data retrieval
func (m model) fetchHistory() tea.Msg {
	commits, err := m.git.GetLog(context.Background(), defaultHistoryLimit)
	if err != nil {
		return errMsg(err)
	}
	return historyMsg(commits)
}

func (m model) fetchReflog() tea.Msg {
	entries, err := m.git.GetReflog(context.Background(), defaultReflogLimit)
	if err != nil {
		return errMsg(err)
	}
	return reflogMsg(entries)
}

func (m model) fetchWorktrees() tea.Msg {
	worktrees, err := m.git.ListWorktrees(context.Background())
	if err != nil {
		return errMsg(err)
	}
	return worktreesMsg(worktrees)
}

func (m model) fetchRebaseStatus() tea.Msg {
	return rebaseStatusMsg(m.git.IsRebasing(context.Background()))
}

func (m model) fetchRebaseSteps(base string) tea.Cmd {
	return func() tea.Msg {
		commits, err := m.git.GetCommitsInRange(context.Background(), base, "HEAD")
		if err != nil {
			return errMsg(err)
		}
		steps := make([]git.RebaseStep, len(commits))
		for i, c := range commits {
			steps[i] = git.RebaseStep{
				Action:  git.ActionPick,
				Hash:    c.Hash,
				Message: c.Message,
			}
		}
		return rebaseStepsMsg(steps)
	}
}

func (m model) fetchBranch() tea.Msg {
	branch, err := m.git.GetCurrentBranch(context.Background())
	if err != nil {
		return errMsg(err)
	}
	return branchMsg(branch)
}

func (m model) fetchFiles() tea.Msg {
	files, err := m.git.ParseStatus(context.Background())
	if err != nil {
		return errMsg(err)
	}
	return filesMsg(files)
}

func (m model) fetchMeta() tea.Msg {
	meta, err := m.wsManager.Load()
	if err != nil {
		return errMsg(err)
	}
	return metaMsg(meta)
}

func (m model) fetchHygiene() tea.Msg {
	report, err := m.hygiene.GetReport(context.Background())
	if err != nil {
		return errMsg(err)
	}
	return reportMsg(report)
}

func (m model) fetchDiff(path string, staged bool) tea.Cmd {
	return func() tea.Msg {
		diff, err := m.git.GetDiff(context.Background(), path, staged)
		if err != nil {
			return errMsg(err)
		}
		return diffMsg(diff)
	}
}

// Action commands (stage, unstage, commit, etc.)
func (m model) stageFile(path string) tea.Cmd {
	return func() tea.Msg {
		err := m.git.StageFile(context.Background(), path)
		if err != nil {
			return errMsg(err)
		}
		return m.fetchFiles()
	}
}

func (m model) unstageFile(path string) tea.Cmd {
	return func() tea.Msg {
		err := m.git.UnstageFile(context.Background(), path)
		if err != nil {
			return errMsg(err)
		}
		return m.fetchFiles()
	}
}

func (m model) toggleStage(f git.FileStatus) tea.Cmd {
	if f.Staged {
		return m.unstageFile(f.Path)
	}
	return m.stageFile(f.Path)
}

func (m model) doCommit(msg string) tea.Cmd {
	return func() tea.Msg {
		// Pre-check: anything staged?
		hasStaged := false
		for _, f := range m.files {
			if f.Staged {
				hasStaged = true
				break
			}
		}
		if !hasStaged {
			return errMsg(fmt.Errorf("nothing to commit (stage your changes with [space] first)"))
		}

		ctx := context.Background()
		err := m.git.Commit(ctx, msg)
		if err != nil {
			return errMsg(err)
		}

		hash, err := m.git.GetLastCommitHash(ctx)
		if err != nil {
			return errMsg(err)
		}

		// Auto-detect semantic tag from message prefix
		tag := workspace.DetectTag(msg)
		if tag != "none" {
			if err := m.wsManager.AddTag(hash, tag); err != nil {
				return errMsg(err)
			}
		}

		// Auto-assign to active workspace
		if m.meta != nil && m.meta.ActiveWorkspace != "" {
			if err := m.wsManager.AddCommitToWorkspace(m.meta.ActiveWorkspace, hash); err != nil {
				return errMsg(err)
			}
		}

		// Save impact (decided in TUI)
		if err := m.wsManager.AddImpact(hash, m.decidedImpact); err != nil {
			return errMsg(err)
		}

		return successMsg("Commit done!")
	}
}

func (m model) createWorkspace(name, desc string) tea.Cmd {
	return func() tea.Msg {
		id := strings.ToLower(strings.ReplaceAll(name, " ", "-"))
		err := m.wsManager.CreateWorkspace(id, name, desc)
		if err != nil {
			return errMsg(err)
		}
		// Auto-activate the new workspace
		if err := m.wsManager.SetActiveWorkspace(id); err != nil {
			return errMsg(err)
		}
		return successMsg("Workspace created and activated!")
	}
}

func (m model) doApplyHunk(h diff.Hunk, filePath string) tea.Cmd {
	return func() tea.Msg {
		patch := h.ToPatch(filePath)
		err := m.git.ApplyHunk(context.Background(), patch)
		if err != nil {
			return errMsg(err)
		}
		return successMsg("Hunk applied!")
	}
}

func (m model) doResetHash(target string) tea.Cmd {
	return func() tea.Msg {
		err := m.git.ResetToHash(context.Background(), target)
		if err != nil {
			return errMsg(err)
		}
		return successMsg("Repository restored to " + target)
	}
}
