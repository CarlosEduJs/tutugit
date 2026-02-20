package main

import (
	"context"
	"strings"

	"tutugit/internal/diff"
	"tutugit/internal/git"
	"tutugit/internal/workspace"

	tea "github.com/charmbracelet/bubbletea"
)

// handleWindowResize handles terminal window resize events
func (m *model) handleWindowResize(msg tea.WindowSizeMsg) {
	m.width = msg.Width
	m.height = msg.Height
	m.diffViewport.Width = msg.Width
	m.diffViewport.Height = msg.Height - viewportReservedSpace
	m.historyViewport.Height = msg.Height - viewportReservedSpace
	m.historyViewport.Width = msg.Width
	m.reflogViewport.Height = msg.Height - viewportReservedSpace
	m.reflogViewport.Width = msg.Width
	m.summaryViewport.Height = msg.Height - viewportReservedSpace
	m.summaryViewport.Width = msg.Width
}

// handleBranchMsg handles branch name updates
func (m *model) handleBranchMsg(msg branchMsg) {
	m.branch = string(msg)
}

// handleFilesMsg handles file status updates
func (m *model) handleFilesMsg(msg filesMsg) {
	m.files = msg
	m.isUpdating = false
	if m.cursor >= len(m.files) {
		m.cursor = len(m.files) - 1
	}
	if m.cursor < 0 && len(m.files) > 0 {
		m.cursor = 0
	}
}

// handleMetaMsg handles workspace metadata updates
func (m *model) handleMetaMsg(msg metaMsg) {
	m.meta = msg
	if m.state == stateNewWorkspace || m.state == stateWorkspaces {
		m.state = stateWorkspaces
	}
	m.isUpdating = false
}

// handleReportMsg handles hygiene report updates
func (m *model) handleReportMsg(msg reportMsg) {
	m.report = msg
}

// handleHistoryMsg handles commit history updates
func (m *model) handleHistoryMsg(msg historyMsg) {
	m.commits = msg
	m.isUpdating = false
	m.renderHistory()
}

// handleReflogMsg handles reflog updates
func (m *model) handleReflogMsg(msg reflogMsg) {
	m.reflogEntries = msg
	m.isUpdating = false
	m.renderReflog()
}

// handleWorktreesMsg handles worktree list updates
func (m *model) handleWorktreesMsg(msg worktreesMsg) {
	m.worktrees = msg
	m.isUpdating = false
}

// handleRebaseStatusMsg handles rebase status updates
func (m *model) handleRebaseStatusMsg(msg rebaseStatusMsg) {
	m.isRebasing = bool(msg)
	if m.isRebasing && m.state == stateMain {
		m.state = stateRebaseOngoing
	}
}

// handleRebaseStepsMsg handles rebase step updates
func (m *model) handleRebaseStepsMsg(msg rebaseStepsMsg) {
	m.rebaseSteps = msg
	m.isUpdating = false
	m.state = stateRebasePrepare
	m.rebaseCursor = 0
}

// handleSummaryMsg handles summary content updates
func (m *model) handleSummaryMsg(msg summaryMsg) {
	m.summaryContent = string(msg)
	m.isUpdating = false
	m.state = stateSummary
	m.summaryViewport.SetContent(m.summaryContent)
}

// handleDiffMsg handles diff content updates
func (m *model) handleDiffMsg(msg diffMsg) {
	m.fileDiff = string(msg)
	m.isUpdating = false
	if m.state == stateHunks {
		fileDiffs := diff.ParseDiff(m.fileDiff)
		if len(fileDiffs) > 0 {
			m.hunks = fileDiffs[0].Hunks
			m.hunkCursor = 0
		}
	} else if m.state == stateDiff {
		// Apply styling to diff lines
		lines := strings.Split(m.fileDiff, "\n")
		var styledLines []string
		for _, l := range lines {
			if strings.HasPrefix(l, "+") {
				styledLines = append(styledLines, styleDiffAdd.Render(l))
			} else if strings.HasPrefix(l, "-") {
				styledLines = append(styledLines, styleDiffDel.Render(l))
			} else {
				styledLines = append(styledLines, l)
			}
		}
		m.diffViewport.SetContent(strings.Join(styledLines, "\n"))
	}
}

// handleSuccessMsg handles success messages
func (m *model) handleSuccessMsg() (model, tea.Cmd) {
	if m.state == stateHunks {
		// Stay in hunk mode but refresh files/diff
		if m.cursor >= 0 && m.cursor < len(m.files) {
			f := m.files[m.cursor]
			return *m, tea.Batch(m.fetchFiles, m.fetchDiff(f.Path, f.Staged), m.fetchHygiene)
		}
	}
	m.state = stateMain
	m.commitMsg.Reset()
	return *m, tea.Batch(m.fetchBranch, m.fetchFiles, m.fetchMeta, m.fetchHygiene)
}

// handleErrorMsg handles error messages
func (m *model) handleErrorMsg(msg errMsg) {
	m.err = msg
	m.isUpdating = false
}

// handleKeySummary handles keyboard input in summary view state
func (m *model) handleKeySummary(msg tea.KeyMsg) (model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q", "L":
		m.state = stateMain
		return *m, nil
	case "E":
		if !m.isUpdating {
			m.isUpdating = true
			return *m, m.exportMarkdown()
		}
	}
	var cmd tea.Cmd
	m.summaryViewport, cmd = m.summaryViewport.Update(msg)
	return *m, cmd
}

// handleKeyReflog handles keyboard input in reflog state
func (m *model) handleKeyReflog(msg tea.KeyMsg) (model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q", "g":
		m.state = stateMain
		return *m, nil
	case "up", "k":
		if m.reflogCursor > 0 {
			m.reflogCursor--
			m.renderReflog()
		}
	case "down", "j":
		if m.reflogCursor < len(m.reflogEntries)-1 {
			m.reflogCursor++
			m.renderReflog()
		}
	case "enter":
		if len(m.reflogEntries) > 0 {
			m.state = stateReflogConfirm
			return *m, nil
		}
	}
	var cmd tea.Cmd
	m.reflogViewport, cmd = m.reflogViewport.Update(msg)
	return *m, cmd
}

// handleKeyReflogConfirm handles keyboard input in reflog confirmation state
func (m *model) handleKeyReflogConfirm(msg tea.KeyMsg) (model, tea.Cmd) {
	switch msg.String() {
	case "esc", "n":
		m.state = stateReflog
		return *m, nil
	case "y", "enter":
		if m.reflogCursor >= 0 && m.reflogCursor < len(m.reflogEntries) {
			m.isUpdating = true
			entry := m.reflogEntries[m.reflogCursor]
			return *m, m.doResetHash(entry.Selector)
		}
	}
	return *m, nil
}

// handleKeyHistory handles keyboard input in history state
func (m *model) handleKeyHistory(msg tea.KeyMsg) (model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q", "h":
		m.state = stateMain
		return *m, nil
	case "R":
		if m.historyCursor >= 0 && m.historyCursor < len(m.commits) {
			hash := m.commits[m.historyCursor].Hash
			m.isUpdating = true
			return *m, m.fetchRebaseSteps(hash)
		}
	case "up", "k":
		if m.historyCursor > 0 {
			m.historyCursor--
			m.renderHistory()
		}
		return *m, nil
	case "down", "j":
		if m.historyCursor < len(m.commits)-1 {
			m.historyCursor++
			m.renderHistory()
		}
		return *m, nil
	case "enter":
		if m.historyCursor >= 0 && m.historyCursor < len(m.commits) {
			hash := m.commits[m.historyCursor].Hash
			m.expandedHistory[hash] = !m.expandedHistory[hash]
			m.renderHistory()
		}
		return *m, nil
	case "L":
		m.isUpdating = true
		m.summaryViewport.SetContent("Generating summary...")
		return *m, m.fetchSummary()
	}
	var cmd tea.Cmd
	m.historyViewport, cmd = m.historyViewport.Update(msg)
	return *m, cmd
}

// handleKeyRebasePrepare handles keyboard input in rebase prepare state
func (m *model) handleKeyRebasePrepare(msg tea.KeyMsg) (model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.state = stateHistory
		return *m, nil
	case "up", "k":
		if m.rebaseCursor > 0 {
			m.rebaseCursor--
		}
	case "down", "j":
		if m.rebaseCursor < len(m.rebaseSteps)-1 {
			m.rebaseCursor++
		}
	case "K": // Move up
		if m.rebaseCursor > 0 {
			m.rebaseSteps[m.rebaseCursor], m.rebaseSteps[m.rebaseCursor-1] = m.rebaseSteps[m.rebaseCursor-1], m.rebaseSteps[m.rebaseCursor]
			m.rebaseCursor--
		}
	case "J": // Move down
		if m.rebaseCursor < len(m.rebaseSteps)-1 {
			m.rebaseSteps[m.rebaseCursor], m.rebaseSteps[m.rebaseCursor+1] = m.rebaseSteps[m.rebaseCursor+1], m.rebaseSteps[m.rebaseCursor]
			m.rebaseCursor++
		}
	case "p":
		if m.rebaseCursor >= 0 && m.rebaseCursor < len(m.rebaseSteps) {
			m.rebaseSteps[m.rebaseCursor].Action = git.ActionPick
		}
	case "s":
		if m.rebaseCursor >= 0 && m.rebaseCursor < len(m.rebaseSteps) {
			m.rebaseSteps[m.rebaseCursor].Action = git.ActionSquash
		}
	case "f":
		if m.rebaseCursor >= 0 && m.rebaseCursor < len(m.rebaseSteps) {
			m.rebaseSteps[m.rebaseCursor].Action = git.ActionFixup
		}
	case "e":
		if m.rebaseCursor >= 0 && m.rebaseCursor < len(m.rebaseSteps) {
			m.rebaseSteps[m.rebaseCursor].Action = git.ActionEdit
		}
	case "d":
		if m.rebaseCursor >= 0 && m.rebaseCursor < len(m.rebaseSteps) {
			m.rebaseSteps[m.rebaseCursor].Action = git.ActionDrop
		}
	case "r":
		if m.rebaseCursor >= 0 && m.rebaseCursor < len(m.rebaseSteps) {
			m.rebaseSteps[m.rebaseCursor].Action = git.ActionReword
		}
	case "enter":
		if m.historyCursor >= 0 && m.historyCursor < len(m.commits) && len(m.rebaseSteps) > 0 {
			m.isUpdating = true
			base := m.commits[m.historyCursor].Hash
			return *m, func() tea.Msg {
				err := m.git.RunInteractiveRebase(context.Background(), base, m.rebaseSteps)
				if err != nil {
					return errMsg(err)
				}
				return successMsg("Rebase completed successfully!")
			}
		}
	}
	return *m, nil
}

// handleKeyRebaseOngoing handles keyboard input in rebase ongoing state
func (m *model) handleKeyRebaseOngoing(msg tea.KeyMsg) (model, tea.Cmd) {
	switch msg.String() {
	case "c":
		m.isUpdating = true
		return *m, func() tea.Msg {
			err := m.git.RebaseContinue(context.Background())
			if err != nil {
				return errMsg(err)
			}
			return successMsg("Rebase continued!")
		}
	case "a":
		m.isUpdating = true
		return *m, func() tea.Msg {
			err := m.git.RebaseAbort(context.Background())
			if err != nil {
				return errMsg(err)
			}
			return successMsg("Rebase aborted!")
		}
	case "s":
		m.isUpdating = true
		return *m, func() tea.Msg {
			err := m.git.RebaseSkip(context.Background())
			if err != nil {
				return errMsg(err)
			}
			return successMsg("Patch skipped!")
		}
	case "esc", "q":
		m.state = stateMain
		return *m, nil
	}
	return *m, nil
}

// handleKeyDiff handles keyboard input in diff view state
func (m *model) handleKeyDiff(msg tea.KeyMsg) (model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q", "d":
		m.state = stateMain
		return *m, nil
	}
	var cmd tea.Cmd
	m.diffViewport, cmd = m.diffViewport.Update(msg)
	return *m, cmd
}

// handleKeyHunks handles keyboard input in hunks state
func (m *model) handleKeyHunks(msg tea.KeyMsg) (model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.state = stateMain
		m.hunks = nil
		return *m, nil
	case "up", "k":
		if m.hunkCursor > 0 {
			m.hunkCursor--
		}
	case "down", "j":
		if m.hunkCursor < len(m.hunks)-1 {
			m.hunkCursor++
		}
	case "space", "s":
		if m.hunkCursor >= 0 && m.hunkCursor < len(m.hunks) &&
			m.cursor >= 0 && m.cursor < len(m.files) {
			m.isUpdating = true
			f := m.files[m.cursor]
			return *m, m.doApplyHunk(m.hunks[m.hunkCursor], f.Path)
		}
	}
	return *m, nil
}

// handleKeyCommit handles keyboard input in commit state
func (m *model) handleKeyCommit(msg tea.KeyMsg) (model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = stateMain
		return *m, nil
	case "alt+i":
		// Cycle impact: patch -> minor -> major -> "" (auto)
		switch m.manualImpact {
		case "":
			m.manualImpact = "patch"
		case "patch":
			m.manualImpact = "minor"
		case "minor":
			m.manualImpact = "major"
		case "major":
			m.manualImpact = "" // back to auto
		}
		m.decidedImpact = m.manualImpact
		if m.manualImpact == "" {
			m.decidedImpact = workspace.DetectImpact(m.commitMsg.Value())
		}
		return *m, nil
	case "enter":
		if m.commitMsg.Value() != "" {
			m.isUpdating = true
			return *m, m.doCommit(m.commitMsg.Value())
		}
	}

	oldVal := m.commitMsg.Value()
	var cmd tea.Cmd
	m.commitMsg, cmd = m.commitMsg.Update(msg)
	newVal := m.commitMsg.Value()

	// update impact if NOT manual and message changed
	if m.manualImpact == "" && oldVal != newVal {
		m.decidedImpact = workspace.DetectImpact(newVal)
	}
	return *m, cmd
}

// handleKeyNewWorkspace handles keyboard input in new workspace state
func (m *model) handleKeyNewWorkspace(msg tea.KeyMsg) (model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = stateWorkspaces
		return *m, nil
	case "tab":
		if m.newWsName.Focused() {
			m.newWsName.Blur()
			m.newWsDesc.Focus()
		} else {
			m.newWsDesc.Blur()
			m.newWsName.Focus()
		}
		return *m, nil
	case "enter":
		if m.newWsName.Value() != "" {
			m.isUpdating = true
			return *m, m.createWorkspace(m.newWsName.Value(), m.newWsDesc.Value())
		}
	}
	var cmd tea.Cmd
	if m.newWsName.Focused() {
		m.newWsName, cmd = m.newWsName.Update(msg)
	} else {
		m.newWsDesc, cmd = m.newWsDesc.Update(msg)
	}
	return *m, cmd
}

// handleKeyWorkspaces handles keyboard input in workspaces state
func (m *model) handleKeyWorkspaces(msg tea.KeyMsg) (model, tea.Cmd) {
	switch msg.String() {
	case "esc", "w":
		m.state = stateMain
		m.cursor = 0
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.meta.Workspaces)-1 {
			m.cursor++
		}
	case "n":
		m.state = stateNewWorkspace
		m.newWsName.Reset()
		m.newWsDesc.Reset()
		m.newWsName.Focus()
	case "a":
		if m.meta != nil && len(m.meta.Workspaces) > 0 && m.cursor < len(m.meta.Workspaces) {
			wsID := m.meta.Workspaces[m.cursor].ID
			m.isUpdating = true
			return *m, func() tea.Msg {
				if err := m.wsManager.SetActiveWorkspace(wsID); err != nil {
					return errMsg(err)
				}
				return successMsg("Workspace activated!")
			}
		}
	}
	return *m, nil
}

// handleKeyGitWorktrees handles keyboard input in git worktrees state
func (m *model) handleKeyGitWorktrees(msg tea.KeyMsg) (model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q", "t":
		m.state = stateMain
		return *m, nil
	case "up", "k":
		if m.worktreeCursor > 0 {
			m.worktreeCursor--
		}
	case "down", "j":
		if m.worktreeCursor < len(m.worktrees)-1 {
			m.worktreeCursor++
		}
	}
	return *m, nil
}

// handleKeyMain handles keyboard input in main state
func (m *model) handleKeyMain(msg tea.KeyMsg) (model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return *m, tea.Quit
	case "r":
		m.isUpdating = true
		return *m, tea.Batch(m.fetchBranch, m.fetchFiles, m.fetchMeta, m.fetchHygiene)
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.files)-1 {
			m.cursor++
		}
	case "space", "s":
		if m.cursor >= 0 && m.cursor < len(m.files) {
			f := m.files[m.cursor]
			m.isUpdating = true
			return *m, m.toggleStage(f)
		}
	case "enter", "e":
		if m.cursor >= 0 && m.cursor < len(m.files) {
			if m.expandedFile == m.cursor {
				m.expandedFile = -1
				m.fileDiff = ""
			} else {
				m.expandedFile = m.cursor
				m.isUpdating = true
				f := m.files[m.cursor]
				return *m, m.fetchDiff(f.Path, f.Staged)
			}
		}
	case "d":
		if m.cursor >= 0 && m.cursor < len(m.files) {
			m.state = stateDiff
			m.isUpdating = true
			f := m.files[m.cursor]
			m.diffViewport.SetContent("Loading diff...")
			return *m, m.fetchDiff(f.Path, f.Staged)
		}
	case "h":
		m.state = stateHistory
		m.isUpdating = true
		m.historyViewport.SetContent("Loading history...")
		return *m, m.fetchHistory
	case "g":
		m.state = stateReflog
		m.isUpdating = true
		m.reflogViewport.SetContent("Loading reflog...")
		return *m, m.fetchReflog
	case "t":
		m.state = stateGitWorktrees
		m.isUpdating = true
		return *m, m.fetchWorktrees
	case "p":
		if len(m.files) > 0 {
			m.state = stateHunks
			m.isUpdating = true
			f := m.files[m.cursor]
			return *m, m.fetchDiff(f.Path, f.Staged)
		}
	case "c":
		m.state = stateCommit
		m.commitMsg.Focus()
	case "w":
		m.state = stateWorkspaces
		m.cursor = 0
	case "L":
		m.isUpdating = true
		m.summaryViewport.SetContent("Generating summary...")
		return *m, m.fetchSummary()
	}
	return *m, nil
}
