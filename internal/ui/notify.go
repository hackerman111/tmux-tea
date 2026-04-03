package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

const teaTimeASCII = `
████████╗███████╗ █████╗ 
╚══██╔══╝██╔════╝██╔══██╗
   ██║   █████╗  ███████║
   ██║   ██╔══╝  ██╔══██║
   ██║   ███████╗██║  ██║
   ╚═╝   ╚══════╝╚═╝  ╚═╝`

const finishedASCII = `
██████╗  ██████╗ ███╗   ██╗███████╗
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
	var s string
	if m.finished {
		s = NotifyStyle.Render(finishedASCII) + "\n\n"
		s += NotifyStyle.Render("Чаепитие завершено!") + "\n"
	} else {
		s = NotifyStyle.Render(teaTimeASCII) + "\n\n"
		s += NotifyStyle.Render(
			fmt.Sprintf("%s · пролив %d/%d", m.teaName, m.pourIndex+1, m.totalPours),
		) + "\n"
	}

	s += "\n"
	if m.finished {
		s += HelpStyle.Render("  Нажмите Enter")
	} else {
		s += HelpStyle.Render("  Enter закрыть, затем Prefix+T")
	}
	return BorderStyle.Render(s)
}
