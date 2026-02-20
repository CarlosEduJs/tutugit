package main

import (
	"fmt"
	"os"

	"tutugit/internal/config"
	"tutugit/internal/diff"
	"tutugit/internal/git"
	"tutugit/internal/hygiene"
	"tutugit/internal/workspace"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// model represents the application state
type model struct {
	git             git.GitProvider
	wsManager       *workspace.Manager
	cfgManager      *config.Manager
	cfg             *config.Config
	hygiene         *hygiene.Analyzer
	meta            *workspace.Meta
	report          *hygiene.HealthReport
	branch          string
	files           []git.FileStatus
	commits         []git.Commit
	reflogEntries   []git.ReflogEntry
	worktrees       []git.Worktree
	rebaseSteps     []git.RebaseStep
	cursor          int
	expandedFile    int // -1 if none
	fileDiff        string
	hunks           []diff.Hunk
	hunkCursor      int
	historyCursor   int
	reflogCursor    int
	worktreeCursor  int
	rebaseCursor    int
	expandedHistory map[string]bool
	diffViewport    viewport.Model
	historyViewport viewport.Model
	reflogViewport  viewport.Model
	err             error
	state           state
	commitMsg       textinput.Model
	newWsName       textinput.Model
	newWsDesc       textinput.Model
	selectedTag     int
	isUpdating      bool
	isRebasing      bool
	width           int
	height          int
	summaryContent  string
	summaryViewport viewport.Model
	manualImpact    string // empty if auto
	decidedImpact   string // the actual impact being used
	suggestedImpact string
	suggestedCount  int
}

// Message types for tea.Cmd
type branchMsg string
type filesMsg []git.FileStatus
type metaMsg *workspace.Meta
type reportMsg *hygiene.HealthReport
type diffMsg string
type historyMsg []git.Commit
type reflogMsg []git.ReflogEntry
type worktreesMsg []git.Worktree
type rebaseStatusMsg bool
type rebaseStepsMsg []git.RebaseStep
type summaryMsg string
type errMsg error
type successMsg string

// initialModel creates and initializes the application model
func initialModel() (model, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return model{}, fmt.Errorf("failed to get working directory: %w", err)
	}
	g := git.NewRunner(cwd)
	w := workspace.NewManager(cwd)
	c := config.NewManager(cwd)

	// Load config (defaults if not present)
	cfg, err := c.Load()
	if err != nil {
		cfg = config.DefaultConfig()
	}

	ti := textinput.New()
	ti.Placeholder = "Commit message..."
	ti.Focus()

	wn := textinput.New()
	wn.Placeholder = "Workspace name..."

	wd := textinput.New()
	wd.Placeholder = "Description (optional)..."

	vp := newStyledViewport("62")
	hp := newStyledViewport("63")
	rp := newStyledViewport("64")
	sp := newStyledViewport("39")

	return model{
		git:             g,
		wsManager:       w,
		cfgManager:      c,
		cfg:             cfg,
		hygiene:         hygiene.NewAnalyzer(g, w),
		state:           stateMain,
		commitMsg:       ti,
		newWsName:       wn,
		newWsDesc:       wd,
		expandedFile:    noFileSelected,
		diffViewport:    vp,
		historyViewport: hp,
		reflogViewport:  rp,
		summaryViewport: sp,
		expandedHistory: make(map[string]bool),
		isUpdating:      true,
		decidedImpact:   "patch",
	}, nil
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.fetchBranch,
		m.fetchFiles,
		m.fetchMeta,
		m.fetchHygiene,
		m.fetchRebaseStatus,
	)
}

// View dispatches to the appropriate view based on current state
func (m model) View() string {
	if m.err != nil {
		return styleError.Render(fmt.Sprintf("Error: %v", m.err)) + "\n\nPress 'q' to quit."
	}

	switch m.state {
	case stateCommit:
		return m.viewCommit()
	case stateWorkspaces:
		return m.viewWorkspaces()
	case stateNewWorkspace:
		return m.viewNewWorkspace()
	case stateHunks:
		return m.viewHunks()
	case stateDiff:
		return m.viewDiff()
	case stateHistory:
		return m.viewHistory()
	case stateReflog:
		return m.viewReflog()
	case stateReflogConfirm:
		return m.viewReflogConfirm()
	case stateGitWorktrees:
		return m.viewGitWorktrees()
	case stateRebasePrepare:
		return m.viewRebasePrepare()
	case stateRebaseOngoing:
		return m.viewRebaseOngoing()
	case stateSummary:
		return m.viewSummary()
	}

	return m.viewMain()
}
