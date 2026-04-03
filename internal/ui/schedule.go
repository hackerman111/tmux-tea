package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/papayka/tmux-tea/internal/config"
)

type ScheduleResult struct {
	Schedule      *config.Schedule
	Action        string
	ScheduleIndex int
}

type ScheduleModel struct {
	teaName   string
	schedules []config.Schedule
	cursor    int
	result    *ScheduleResult
}

func NewScheduleModel(teaName string, schedules []config.Schedule) ScheduleModel {
	return ScheduleModel{teaName: teaName, schedules: schedules}
}

func (m ScheduleModel) Result() *ScheduleResult {
	return m.result
}

func (m ScheduleModel) Init() tea.Cmd {
	return nil
}

func (m ScheduleModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
			m.result = &ScheduleResult{Action: "back"}
			return m, tea.Quit

		case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
			if m.cursor > 0 {
				m.cursor--
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))):
			if m.cursor < len(m.schedules)-1 {
				m.cursor++
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			if m.cursor < len(m.schedules) {
				m.result = &ScheduleResult{
					Schedule:      &m.schedules[m.cursor],
					Action:        "select",
					ScheduleIndex: m.cursor,
				}
				return m, tea.Quit
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("a"))):
			m.result = &ScheduleResult{Action: "add"}
			return m, tea.Quit

		case key.Matches(msg, key.NewBinding(key.WithKeys("e"))):
			if m.cursor < len(m.schedules) {
				m.result = &ScheduleResult{
					Schedule:      &m.schedules[m.cursor],
					Action:        "edit",
					ScheduleIndex: m.cursor,
				}
				return m, tea.Quit
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("d"))):
			if m.cursor < len(m.schedules) && len(m.schedules) > 1 {
				m.result = &ScheduleResult{
					Schedule:      &m.schedules[m.cursor],
					Action:        "delete",
					ScheduleIndex: m.cursor,
				}
				return m, tea.Quit
			}
		}
	}

	return m, nil
}

func formatPours(pours []int) string {
	if len(pours) == 0 {
		return ""
	}

	parts := make([]string, 0, len(pours))
	for _, pour := range pours {
		parts = append(parts, fmt.Sprintf("%d", pour))
	}

	preview := strings.Join(parts, ",")
	if len(preview) > 20 {
		preview = preview[:20] + ".."
	}

	return "(" + preview + ")"
}

func (m ScheduleModel) View() string {
	lines := []string{
		renderTitle(fmt.Sprintf("%s — расписание:", m.teaName)),
		"",
	}

	for i, schedule := range m.schedules {
		label := fmt.Sprintf("%s %s", schedule.Name, formatPours(schedule.Pours))
		if i == m.cursor {
			lines = append(lines, renderSelectedLine(label))
		} else {
			lines = append(lines, renderNormalLine(label))
		}
	}

	lines = append(lines,
		"",
		renderMutedLine("─────────────"),
		renderMutedLine("a добавить  e редакт.  d удалить"),
		"",
		renderHelpLine("esc назад  enter старт"),
	)

	return renderPanel(lines...)
}
