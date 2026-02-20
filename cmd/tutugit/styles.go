package main

import (
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)

var (
	styleTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	styleBranch = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#43BF6D"))

	styleSelected = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Bold(true)

	styleStaged = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#43BF6D"))

	styleUnstaged = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E84855"))

	styleWS = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#2E3440")).
		Padding(0, 1)

	styleAlert = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EAD94C")).
			Bold(true)

	styleDiffAdd = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#43BF6D"))

	styleDiffDel = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E84855"))

	styleError = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E84855"))

	styleDim = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666"))

	styleVersion = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Italic(true)
)

// newStyledViewport creates a new viewport with consistent styling
func newStyledViewport(borderColor string) viewport.Model {
	vp := viewport.New(0, 0)
	vp.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(borderColor)).
		Padding(0, 1)
	return vp
}
