package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const teaTimeASCII = `████████╗███████╗ █████╗
╚══██╔══╝██╔════╝██╔══██╗
   ██║   █████╗  ███████║
   ██║   ██╔══╝  ██╔══██║
   ██║   ███████╗██║  ██║
   ╚═╝   ╚══════╝╚═╝  ╚═╝`

const finishedASCII = `██████╗  ██████╗ ███╗   ██╗███████╗
██╔══██╗██╔═══██╗████╗  ██║██╔════╝
██║  ██║██║   ██║██╔██╗ ██║█████╗
██║  ██║██║   ██║██║╚██╗██║██╔══╝
██████╔╝╚██████╔╝██║ ╚████║███████╗
╚═════╝  ╚═════╝ ╚═╝  ╚═══╝╚══════╝`

type NotifyModel struct {
	teaName    string
	pourIndex  int
	totalPours int
	finished   bool
	confirmed  bool
}

func NewNotifyModel(teaName string, pourIndex, totalPours int, finished bool) NotifyModel {
	return NotifyModel{
		teaName:    teaName,
		pourIndex:  pourIndex,
		totalPours: totalPours,
		finished:   finished,
	}
}

func (m NotifyModel) Confirmed() bool {
	return m.confirmed
}

func (m NotifyModel) Init() tea.Cmd {
	return nil
}

func (m NotifyModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter", "esc", " "))):
			m.confirmed = true
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m NotifyModel) View() string {
	const contentWidth = 46

	var header, subtext, help string
	if m.finished {
		header = finishedASCII
		subtext = "Чаепитие завершено!"
		help = "Нажмите Enter"
	} else {
		header = teaTimeASCII
		subtext = fmt.Sprintf("%s · пролив %d/%d", m.teaName, m.pourIndex+1, m.totalPours)
		help = "Enter закрыть, затем Prefix+T"
	}

	body := NotifyStyle.Width(contentWidth).Render(header + "\n\n" + subtext)
	foot := lipgloss.NewStyle().
		Foreground(ColorMuted).
		Width(contentWidth).
		Align(lipgloss.Center).
		Render(help)

	return BorderStyle.Width(contentWidth).Render(body + "\n\n" + foot)
}
