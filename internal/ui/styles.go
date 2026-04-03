package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	menuPopupWidth   = 56
	menuPopupHeight  = 16
	panelBodyWidth   = 46
	notifyPopupWidth = 58
)

var (
	ColorPrimary   = lipgloss.Color("#D4A574")
	ColorSecondary = lipgloss.Color("#8B7355")
	ColorAccent    = lipgloss.Color("#C8E6C9")
	ColorMuted     = lipgloss.Color("#666666")
	ColorDanger    = lipgloss.Color("#E57373")

	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary)

	SelectedStyle = lipgloss.NewStyle().
			Foreground(ColorAccent).
			Bold(true)

	NormalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF"))

	MutedStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	BorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorSecondary).
			Padding(1, 2)

	NotifyStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			Align(lipgloss.Center)
)

func renderPanel(lines ...string) string {
	return BorderStyle.Width(panelBodyWidth).Render(strings.Join(lines, "\n"))
}

func renderTitle(text string) string {
	return TitleStyle.Render(truncateText(text, panelBodyWidth))
}

func renderNormalLine(text string) string {
	return NormalStyle.Render("    " + truncateText(text, panelBodyWidth-4))
}

func renderSelectedLine(text string) string {
	return SelectedStyle.Render("  " + truncateText("> "+text, panelBodyWidth-2))
}

func renderMutedLine(text string) string {
	return MutedStyle.Render("    " + truncateText(text, panelBodyWidth-4))
}

func renderHelpLine(text string) string {
	return HelpStyle.Render("  " + truncateText(text, panelBodyWidth-2))
}

func truncateText(text string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	if lipgloss.Width(text) <= maxWidth {
		return text
	}

	const tail = "..."
	if maxWidth <= lipgloss.Width(tail) {
		return tail[:maxWidth]
	}

	budget := maxWidth - lipgloss.Width(tail)
	var b strings.Builder
	currentWidth := 0
	for _, r := range text {
		runeWidth := lipgloss.Width(string(r))
		if currentWidth+runeWidth > budget {
			break
		}
		b.WriteRune(r)
		currentWidth += runeWidth
	}
	return b.String() + tail
}
