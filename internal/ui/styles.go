package ui

import "github.com/charmbracelet/lipgloss"

var (
	ColorPrimary   = lipgloss.Color("#D4A574")
	ColorSecondary = lipgloss.Color("#8B7355")
	ColorAccent    = lipgloss.Color("#C8E6C9")
	ColorMuted     = lipgloss.Color("#666666")
	ColorDanger    = lipgloss.Color("#E57373")

	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			MarginBottom(1)

	SelectedStyle = lipgloss.NewStyle().
			Foreground(ColorAccent).
			Bold(true).
			PaddingLeft(2)

	NormalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			PaddingLeft(4)

	MutedStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			PaddingLeft(4)

	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			MarginTop(1)

	BorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorSecondary).
			Padding(1, 2)

	NotifyStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			Align(lipgloss.Center)
)
