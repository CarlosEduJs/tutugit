package main

// version is the current version of tutugit. It is injected at build time by GoReleaser.
var version = "1.0.2-beta.1"

// state represents the current UI state of the application
type state int

const (
	stateMain state = iota
	stateCommit
	stateCommitTags
	stateCommitSelectWS
	stateWorkspaces
	stateNewWorkspace
	stateHunks
	stateDiff
	stateHistory
	stateReflog
	stateReflogConfirm
	stateGitWorktrees
	stateRebasePrepare
	stateRebaseOngoing
	stateSummary
)

// UI Constants
const (
	headerHeight          = 2
	footerHeight          = 2
	viewportPadding       = 2
	viewportReservedSpace = headerHeight + footerHeight + viewportPadding

	noFileSelected = -1

	minLogPartsCount = 7

	defaultHistoryLimit = 100
	defaultReflogLimit  = 50
)

// semanticTags are the available semantic commit types
var semanticTags = []string{"feature", "fix", "refactor", "experiment", "docs", "test", "chore", "none"}
