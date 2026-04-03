package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/papayka/tmux-tea/internal/config"
)

type MenuResult struct {
	Tea      *config.Tea
	Action   string
	TeaIndex int
}

type MenuModel struct {
	teas     []config.Tea
	cursor   int
	result   *MenuResult
	quitting bool
}

func NewMenuModel(teas []config.Tea) MenuModel {
	return MenuModel{teas: teas}
}

func (m MenuModel) Result() *MenuResult {
	return m.result
}

func (m MenuModel) Init() tea.Cmd {
	return nil
}

func (m MenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("q", "esc"))):
			m.quitting = true
			m.result = &MenuResult{Action: "quit"}
			return m, tea.Quit

		case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
			if m.cursor > 0 {
				m.cursor--
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))):
			if m.cursor < len(m.teas)-1 {
				m.cursor++
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			if m.cursor < len(m.teas) {
				m.result = &MenuResult{
					Tea:      &m.teas[m.cursor],
					Action:   "select",
					TeaIndex: m.cursor,
				}
				return m, tea.Quit
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("a"))):
			m.result = &MenuResult{Action: "add"}
			return m, tea.Quit

		case key.Matches(msg, key.NewBinding(key.WithKeys("e"))):
			if m.cursor < len(m.teas) {
				m.result = &MenuResult{
					Tea:      &m.teas[m.cursor],
					Action:   "edit",
					TeaIndex: m.cursor,
				}
				return m, tea.Quit
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("d"))):
			if m.cursor < len(m.teas) && len(m.teas) > 1 {
				m.result = &MenuResult{
					Tea:      &m.teas[m.cursor],
					Action:   "delete",
					TeaIndex: m.cursor,
				}
				return m, tea.Quit
			}
		}
	}

	return m, nil
}

func (m MenuModel) View() string {
	s := TitleStyle.Render("Выберите чай:") + "\n\n"

	for i, currentTea := range m.teas {
		if i == m.cursor {
			s += SelectedStyle.Render(fmt.Sprintf("> %s", currentTea.Name)) + "\n"
		} else {
			s += NormalStyle.Render(currentTea.Name) + "\n"
		}
	}

	s += "\n"
	s += MutedStyle.Render("─────────────") + "\n"
	s += MutedStyle.Render("a добавить  e редакт.  d удалить") + "\n"
	s += HelpStyle.Render("  j/k выбор  enter старт  esc выход")

	return BorderStyle.Render(s)
}
