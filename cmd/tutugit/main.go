package main

import (
	"fmt"
	"os"
	"tutugit/internal/config"
	"tutugit/internal/workspace"

	tea "github.com/charmbracelet/bubbletea"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Dispatch non-KeyMsg messages to handlers
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.handleWindowResize(msg)
		return m, nil
	case branchMsg:
		m.handleBranchMsg(msg)
		return m, nil
	case filesMsg:
		m.handleFilesMsg(msg)
		return m, nil
	case metaMsg:
		m.handleMetaMsg(msg)
		return m, nil
	case reportMsg:
		m.handleReportMsg(msg)
		return m, nil
	case historyMsg:
		m.handleHistoryMsg(msg)
		return m, nil
	case reflogMsg:
		m.handleReflogMsg(msg)
		return m, nil
	case worktreesMsg:
		m.handleWorktreesMsg(msg)
		return m, nil
	case rebaseStatusMsg:
		m.handleRebaseStatusMsg(msg)
		return m, nil
	case rebaseStepsMsg:
		m.handleRebaseStepsMsg(msg)
		return m, nil
	case summaryMsg:
		m.handleSummaryMsg(msg)
		return m, nil
	case diffMsg:
		m.handleDiffMsg(msg)
		return m, nil
	case successMsg:
		return m.handleSuccessMsg()
	case errMsg:
		m.handleErrorMsg(msg)
		return m, nil
	case tea.KeyMsg:
		// Dispatch KeyMsg to state-specific keyboard handlers
		switch m.state {
		case stateSummary:
			return m.handleKeySummary(msg)
		case stateReflog:
			return m.handleKeyReflog(msg)
		case stateReflogConfirm:
			return m.handleKeyReflogConfirm(msg)
		case stateHistory:
			return m.handleKeyHistory(msg)
		case stateRebasePrepare:
			return m.handleKeyRebasePrepare(msg)
		case stateRebaseOngoing:
			return m.handleKeyRebaseOngoing(msg)
		case stateDiff:
			return m.handleKeyDiff(msg)
		case stateHunks:
			return m.handleKeyHunks(msg)
		case stateCommit:
			return m.handleKeyCommit(msg)
		case stateNewWorkspace:
			return m.handleKeyNewWorkspace(msg)
		case stateWorkspaces:
			return m.handleKeyWorkspaces(msg)
		case stateGitWorktrees:
			return m.handleKeyGitWorktrees(msg)
		case stateMain:
			return m.handleKeyMain(msg)
		}
	}

	return m, nil
}

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--version":
			fmt.Printf("tutugit version %s\n", version)
			return
		case "init":
			cwd, _ := os.Getwd()
			wsManager := workspace.NewManager(cwd)
			cfgManager := config.NewManager(cwd)

			cfg := config.DefaultConfig()
			if err := cfgManager.Save(cfg); err != nil {
				fmt.Printf("Error saving config: %v\n", err)
				os.Exit(1)
			}

			if err := wsManager.Bootstrap(); err != nil {
				fmt.Printf("Error initializing tutugit: %v\n", err)
				os.Exit(1)
			}

			fmt.Println("🚀 tutugit initialized successfully!")
			fmt.Println(".tutugit directory created with meta.json and config.yml")
			return
		case "demo":
			m := initialDemoModel()
			p := tea.NewProgram(m)
			if _, err := p.Run(); err != nil {
				fmt.Printf("Error starting demo: %v", err)
				os.Exit(1)
			}
			return
		}
	}

	m, err := initialModel()
	if err != nil {
		fmt.Printf("Failed to initialize: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error starting tutugit: %v", err)
		os.Exit(1)
	}
}
