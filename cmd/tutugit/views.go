package main

import (
	"fmt"
	"strings"

	"tutugit/internal/git"
	"tutugit/internal/workspace"

	"github.com/charmbracelet/lipgloss"
)

func safeShortHash(h string) string {
	if len(h) > 7 {
		return h[:7]
	}
	return h
}

func (m model) viewRebasePrepare() string {
	s := m.renderHeader()
	s += styleTitle.Render(" Interactive Rebase Planner ") + "\n\n"
	s += "Reorder and select actions for the commits:\n\n"

	for i, step := range m.rebaseSteps {
		cursor := "  "
		if m.rebaseCursor == i {
			cursor = "❯ "
		}

		actionStyle := styleSelected
		switch step.Action {
		case git.ActionPick:
			actionStyle = styleStaged
		case git.ActionDrop:
			actionStyle = styleUnstaged
		case git.ActionSquash, git.ActionFixup:
			actionStyle = styleAlert
		}

		line := fmt.Sprintf("%-7s %s %s",
			actionStyle.Render(string(step.Action)),
			styleSelected.Render(safeShortHash(step.Hash)),
			step.Message)

		if m.rebaseCursor == i {
			s += cursor + styleSelected.Render(line) + "\n"
		} else {
			s += cursor + line + "\n"
		}
	}

	s += "\nShortcuts: [j/k] navigate | [p/s/f/e/d/r] actions | [J/K] move | [enter] start | [esc] cancel\n"
	return s
}

func (m model) viewRebaseOngoing() string {
	s := m.renderHeader()
	s += styleTitle.Render(" Rebase in Progress ") + "\n\n"
	s += styleAlert.Render("⚠️  The Git is in the middle of a rebase.") + "\n"
	s += "Resolve the conflicts in the files if necessary.\n\n"
	s += "Shortcuts:\n"
	s += "  [c] Continue  - Continue after resolving conflicts\n"
	s += "  [s] Skip      - Skip the current patch\n"
	s += "  [a] Abort     - Abort the rebase\n"
	s += "  [esc/q]       - Back to main screen\n"
	return s
}

func (m model) viewGitWorktrees() string {
	s := m.renderHeader()
	s += styleTitle.Render(" Worktree Explorer ") + "\n\n"

	if len(m.worktrees) == 0 {
		if m.isUpdating {
			s += "  Loading worktrees...\n"
		} else {
			s += "  No worktrees (only main).\n"
		}
	} else {
		for i, wt := range m.worktrees {
			cursor := "  "
			if m.worktreeCursor == i {
				cursor = "❯ "
			}

			style := lipgloss.NewStyle()
			if wt.IsMain {
				style = styleBranch.Copy().Bold(true)
			}

			line := fmt.Sprintf("%s%s [%s] in %s",
				cursor,
				style.Render(wt.Branch),
				styleSelected.Render(safeShortHash(wt.Hash)),
				wt.Path)

			if m.worktreeCursor == i {
				s += styleSelected.Render(line) + "\n"
			} else {
				s += line + "\n"
			}
		}
	}

	s += "\nShortcuts: [j/k/up/down] navigate | [esc/q/t] back\n"
	return s
}

func (m model) viewReflogConfirm() string {
	if m.reflogCursor < 0 || m.reflogCursor >= len(m.reflogEntries) {
		return styleError.Render("Invalid reflog entry selected")
	}
	entry := m.reflogEntries[m.reflogCursor]
	s := styleTitle.Render(" WARNING: Restore Repository? ") + "\n\n"
	s += "You are about to perform a " + styleError.Render("RESET --HARD") + "\n"
	s += "to the state: " + styleSelected.Render(entry.Selector) + "\n"
	s += "Original action: " + styleAlert.Render(entry.Action) + "\n"
	s += "Message: " + entry.Message + "\n\n"
	s += styleError.Render("⚠️  THIS WILL OVERWRITE ALL UNCOMMITTED CHANGES!") + "\n\n"
	s += "Do you want to continue? [y] Yes / [n] No (Esc to cancel)\n"
	return s
}

func (m model) viewDiff() string {
	if m.cursor < 0 || m.cursor >= len(m.files) {
		return styleError.Render("No file selected")
	}
	f := m.files[m.cursor]
	s := m.renderHeader()
	s += styleTitle.Render(" Diff Full: "+f.Path) + " "
	if f.Staged {
		s += styleStaged.Render("[STAGED]")
	} else {
		s += styleUnstaged.Render("[UNSTAGED]")
	}

	help := "\nShortcuts: [j/k/up/down] scroll | [esc/q/d] back | [s] toggle stage\n"
	return s + "\n" + m.diffViewport.View() + help
}

func (m model) viewHunks() string {
	if m.cursor < 0 || m.cursor >= len(m.files) {
		return styleError.Render("No file selected")
	}
	f := m.files[m.cursor]
	s := m.renderHeader()
	s += styleTitle.Render(" Surgical Staging: "+f.Path) + "\n\n"

	if len(m.hunks) == 0 {
		if m.isUpdating {
			s += "  Loading hunks...\n"
		} else {
			s += "  No hunks detected.\n"
		}
	} else {
		for i, h := range m.hunks {
			cursor := "  "
			if m.hunkCursor == i {
				cursor = "> "
			}

			header := styleWS.Render(fmt.Sprintf("Hunk %d/%d", i+1, len(m.hunks)))
			if m.hunkCursor == i {
				s += cursor + styleSelected.Render(header) + "\n"
				// show hunk content for selected
				lines := strings.Split(h.Content, "\n")
				for _, l := range lines {
					if strings.HasPrefix(l, "+") {
						s += styleDiffAdd.Render("    "+l) + "\n"
					} else if strings.HasPrefix(l, "-") {
						s += styleDiffDel.Render("    "+l) + "\n"
					} else {
						s += "    " + l + "\n"
					}
				}
				s += "\n"
			} else {
				s += cursor + header + "\n"
			}
		}
	}

	s += "\nShortcuts: [j/k] navigate | [space/s] stage hunk | [esc/q] back\n"
	return s
}

func (m model) renderHeader() string {
	appName := styleTitle.Render(" tutugit ")
	vTag := styleVersion.Render(" v" + version)
	branch := styleBranch.Render(" " + m.branch)

	header := appName + vTag + branch
	if m.isUpdating {
		header += styleDim.Render(" (updating...)")
	}

	// add project info if present
	if m.cfg != nil && m.cfg.Project.Name != "" {
		header += "\n" + styleWS.Render(" "+m.cfg.Project.Name+" ")
		if m.cfg.Project.Description != "" {
			header += " " + styleDim.Render(m.cfg.Project.Description)
		}
	}

	return header + "\n\n"
}

func (m model) viewMain() string {
	s := m.renderHeader()

	// hygiene alerts section
	if m.report != nil {
		alerts := []string{}
		if len(m.report.WIPCommits) > 0 {
			alerts = append(alerts, fmt.Sprintf("! Detected %d WIP commits in history", len(m.report.WIPCommits)))
		}
		if len(m.report.SquashSuggestions) > 0 {
			alerts = append(alerts, fmt.Sprintf("! Workspaces suggested for squash: %s", strings.Join(m.report.SquashSuggestions, ", ")))
		}
		if len(m.report.StaleWorkspaces) > 0 {
			alerts = append(alerts, fmt.Sprintf("! Workspaces with invalid commits (stale): %s", strings.Join(m.report.StaleWorkspaces, ", ")))
		}

		if len(alerts) > 0 {
			s += "Hygiene Alerts:\n"
			for _, a := range alerts {
				s += styleAlert.Render("  "+a) + "\n"
			}
			s += "\n"
		}
	}

	s += "Git Status:\n"

	if len(m.files) == 0 {
		s += "  No pending changes\n"
	}

	for i, f := range m.files {
		cursor := "  "
		if m.cursor == i {
			cursor = "> "
		}

		statusChar := "U"
		stagedStyle := styleUnstaged
		if f.Staged {
			statusChar = "S"
			stagedStyle = styleStaged
		}

		line := fmt.Sprintf("%s [%s] %s", cursor, statusChar, f.Path)
		if m.cursor == i {
			s += styleSelected.Render(line) + "\n"
		} else {
			s += stagedStyle.Render(line) + "\n"
		}

		// inline diff if expanded
		if m.expandedFile == i {
			if m.fileDiff == "" && m.isUpdating {
				s += "    Loading diff...\n"
			} else if m.fileDiff == "" {
				s += "    (no changes to show)\n"
			} else {
				diffLines := strings.Split(m.fileDiff, "\n")
				// limit lines for preview
				displayLines := 10
				if len(diffLines) < displayLines {
					displayLines = len(diffLines)
				}
				for _, dl := range diffLines[:displayLines] {
					if strings.HasPrefix(dl, "+") {
						s += styleDiffAdd.Render("    "+dl) + "\n"
					} else if strings.HasPrefix(dl, "-") {
						s += styleDiffDel.Render("    "+dl) + "\n"
					} else {
						s += "    " + dl + "\n"
					}
				}
				if len(diffLines) > displayLines {
					s += fmt.Sprintf("    ... (+%d lines)\n", len(diffLines)-displayLines)
				}
			}
		}
	}

	s += "\nShortcuts: \n [q] quit \n [r] refresh \n [space] stage \n [c] commit \n [e/enter] diff \n [d] full diff \n [h] history \n [g] reflog \n [t] worktrees \n [w] workspaces \n [L] summary\n"

	return s
}

func (m *model) renderReflog() {
	var b strings.Builder
	for i, e := range m.reflogEntries {
		prefix := "  "
		if m.reflogCursor == i {
			prefix = "❯ "
		}

		color := "240"
		switch strings.ToLower(e.Action) {
		case "commit":
			color = "#43BF6D"
		case "rebase":
			color = "#7D56F4"
		case "reset":
			color = "#E84855"
		case "checkout":
			color = "#FAFAFA"
		}

		line := fmt.Sprintf("%s%s [%s] %s: %s",
			prefix,
			lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Bold(true).Render(e.Selector),
			styleSelected.Render(safeShortHash(e.Hash)),
			styleAlert.Render(e.Action),
			e.Message)

		if m.reflogCursor == i {
			b.WriteString(styleSelected.Render(line) + "\n")
		} else {
			b.WriteString(line + "\n")
		}
	}
	m.reflogViewport.SetContent(b.String())
}

func (m model) viewReflog() string {
	s := m.renderHeader()
	s += styleTitle.Render(" Time Machine (Reflog) ") + "\n"
	s += m.reflogViewport.View() + "\n"
	s += "Shortcuts: [j/k/up/down] scroll | [esc/q/g] back\n"
	return s
}

func (m *model) renderHistory() {
	var b strings.Builder
	for i, c := range m.commits {
		marker := "○"
		if len(c.Parents) > 1 {
			marker = "Φ" // merge symbol looks pro
		}

		prefix := "  "
		if m.historyCursor == i {
			prefix = "❯ "
		}

		line := fmt.Sprintf("%s%s [%s] %s (%s)",
			prefix,
			styleBranch.Render(marker),
			styleSelected.Render(c.ShortHash),
			c.Message,
			c.Date)

		if m.historyCursor == i {
			b.WriteString(styleSelected.Render(line) + "\n")
		} else {
			b.WriteString(line + "\n")
		}

		// extra details if expanded
		if m.expandedHistory[c.Hash] {
			b.WriteString(fmt.Sprintf("    %s Author: %s <%s>\n", styleAlert.Render(""), c.Author, c.Email))
			b.WriteString(fmt.Sprintf("    %s Hash: %s\n", styleAlert.Render(""), c.Hash))
		} else if m.historyCursor == i {
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("    [Tip: Press Enter for details]") + "\n")
		}

		if i < len(m.commits)-1 {
			b.WriteString("  │\n")
		}
	}
	m.historyViewport.SetContent(b.String())
}

func (m model) viewHistory() string {
	s := m.renderHeader()
	s += styleTitle.Render(" Visual History ") + "\n"
	s += m.historyViewport.View() + "\n"
	s += "Shortcuts: [j/k/up/down] scroll | [enter] details | [R] interactive rebase | [L] summary | [esc/q/h] back\n"
	return s
}

func (m model) viewWorkspaces() string {
	s := m.renderHeader()
	header := styleWS.Render(" Logical Workspaces ")
	if m.isUpdating {
		header += " (updating...)"
	}

	s += header + "\n\n"

	if m.meta == nil || len(m.meta.Workspaces) == 0 {
		s += "  No workspace created. Press [n] to create a new one.\n"
	} else {
		for i, w := range m.meta.Workspaces {
			cursor := "  "
			if m.cursor == i {
				cursor = "> "
			}
			active := ""
			if m.meta.ActiveWorkspace == w.ID {
				active = " \u2605"
			}
			line := fmt.Sprintf("%s %s (%d commits)%s", cursor, w.Name, len(w.Commits), active)
			if w.Description != "" {
				line += fmt.Sprintf(" \u2014 %s", w.Description)
			}
			if m.cursor == i {
				s += styleSelected.Render(line) + "\n"
			} else {
				s += line + "\n"
			}
		}
	}

	s += "\nShortcuts: [n] new | [a] activate | [w/esc] back\n"
	return s
}

func (m model) viewNewWorkspace() string {
	s := m.renderHeader()
	s += styleWS.Render(" New Workspace ") + "\n\n"
	s += "Name:\n"
	s += m.newWsName.View() + "\n\n"
	s += "Description:\n"
	s += m.newWsDesc.View() + "\n\n"
	s += "Shortcuts: [tab] switch field | [enter] create and activate | [esc] cancel\n"
	return s
}

func (m model) viewCommit() string {
	s := m.renderHeader()
	s += styleTitle.Render(" Commit ") + "\n\n"

	// show active workspace
	if m.meta != nil && m.meta.ActiveWorkspace != "" {
		wsName := m.wsManager.GetActiveWorkspaceName(m.meta)
		s += fmt.Sprintf("active workspace: %s\n", styleWS.Render(" "+wsName+" "))
	}

	// show auto-detected tag preview
	currentMsg := m.commitMsg.Value()
	if currentMsg != "" {
		detected := workspace.DetectTag(currentMsg)
		tagLabel := map[string]string{
			"feature": "Feature", "fix": "Fix", "refactor": "Refactor",
			"experiment": "Experiment", "none": "General",
		}
		s += fmt.Sprintf("detected tag: %s\n", tagLabel[detected])
	}

	// show impact (butterfly style)
	impactLabel := map[string]string{
		"patch": "🦋 [PATCH]", "minor": "🦋 [MINOR]", "major": "🦋 [MAJOR]",
	}
	info := "(suggested)"
	if m.manualImpact != "" {
		info = "(Manual)"
	}
	s += fmt.Sprintf("Impact: %s %s\n", impactLabel[m.decidedImpact], styleAlert.Render(info))

	s += "\n" + m.commitMsg.View() + "\n\n"
	s += "Tip: use prefixes like feat:, fix:, refactor: for auto-tagging\n"
	s += "Shortcuts: [enter] commit | [alt+i] impact | [esc] cancel\n"
	return s
}

func (m model) viewSummary() string {
	s := m.renderHeader()
	title := " Release Summary "
	if m.isUpdating {
		title += "(Generating...)"
	}
	s += styleTitle.Render(title) + "\n"
	s += m.summaryViewport.View() + "\n"
	s += "Shortcuts: [j/k/up/down] scroll | [E] export MD | [esc/q/L] back\n"
	return s
}
