package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type TeaEditorResult struct {
	Name  string
	Saved bool
}

type TeaEditorModel struct {
	nameInput textinput.Model
	result    *TeaEditorResult
	title     string
}

func NewTeaEditorModel(name string, isNew bool) TeaEditorModel {
	ti := textinput.New()
	ti.Placeholder = "Название чая"
	ti.SetValue(name)
	ti.Focus()
	ti.CharLimit = 50

	title := "Редактировать чай"
	if isNew {
		title = "Новый чай"
	}

	return TeaEditorModel{
		nameInput: ti,
		title:     title,
	}
}

func (m TeaEditorModel) Result() *TeaEditorResult {
	return m.result
}

func (m TeaEditorModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m TeaEditorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
			m.result = &TeaEditorResult{Saved: false}
			return m, tea.Quit

		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			name := strings.TrimSpace(m.nameInput.Value())
			if name != "" {
				m.result = &TeaEditorResult{Name: name, Saved: true}
				return m, tea.Quit
			}
		}
	}

	var cmd tea.Cmd
	m.nameInput, cmd = m.nameInput.Update(msg)
	return m, cmd
}

func (m TeaEditorModel) View() string {
	s := TitleStyle.Render(m.title) + "\n\n"
	s += "  Название: " + m.nameInput.View() + "\n\n"
	s += HelpStyle.Render("  enter сохранить  esc отмена")
	return BorderStyle.Render(s)
}

type ScheduleEditorResult struct {
	Name  string
	Pours []int
	Saved bool
}

type ScheduleEditorModel struct {
	nameInput  textinput.Model
	poursInput textinput.Model
	focusIndex int
	result     *ScheduleEditorResult
	title      string
	err        string
}

func NewScheduleEditorModel(name string, pours []int, isNew bool) ScheduleEditorModel {
	nameInput := textinput.New()
	nameInput.Placeholder = "Название расписания"
	nameInput.SetValue(name)
	nameInput.Focus()
	nameInput.CharLimit = 50

	poursInput := textinput.New()
	poursInput.Placeholder = "10,15,20,25,30"
	if len(pours) > 0 {
		parts := make([]string, 0, len(pours))
		for _, pour := range pours {
			parts = append(parts, strconv.Itoa(pour))
		}
		poursInput.SetValue(strings.Join(parts, ","))
	}
	poursInput.CharLimit = 200

	title := "Редактировать расписание"
	if isNew {
		title = "Новое расписание"
	}

	return ScheduleEditorModel{
		nameInput:  nameInput,
		poursInput: poursInput,
		title:      title,
	}
}

func (m ScheduleEditorModel) Result() *ScheduleEditorResult {
	return m.result
}

func (m ScheduleEditorModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m ScheduleEditorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
			m.result = &ScheduleEditorResult{Saved: false}
			return m, tea.Quit

		case key.Matches(msg, key.NewBinding(key.WithKeys("tab", "shift+tab"))):
			m.focusIndex = (m.focusIndex + 1) % 2
			m.err = ""
			if m.focusIndex == 0 {
				m.nameInput.Focus()
				m.poursInput.Blur()
			} else {
				m.nameInput.Blur()
				m.poursInput.Focus()
			}
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			name := strings.TrimSpace(m.nameInput.Value())
			if name == "" {
				m.err = "Название не может быть пустым"
				return m, nil
			}

			pours, err := parsePours(m.poursInput.Value())
			if err != nil {
				m.err = err.Error()
				return m, nil
			}

			m.result = &ScheduleEditorResult{
				Name:  name,
				Pours: pours,
				Saved: true,
			}
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	if m.focusIndex == 0 {
		m.nameInput, cmd = m.nameInput.Update(msg)
	} else {
		m.poursInput, cmd = m.poursInput.Update(msg)
	}
	m.err = ""
	return m, cmd
}

func (m ScheduleEditorModel) View() string {
	s := TitleStyle.Render(m.title) + "\n\n"
	s += "  Название: " + m.nameInput.View() + "\n"
	s += "  Проливы (сек, через запятую):\n"
	s += "  " + m.poursInput.View() + "\n"
	if m.err != "" {
		s += "\n" + lipglossRender(ColorDanger, "  "+m.err) + "\n"
	}
	s += "\n"
	s += HelpStyle.Render("  tab поле  enter сохранить  esc отмена")
	return BorderStyle.Render(s)
}

func lipglossRender(color lipgloss.Color, text string) string {
	return lipgloss.NewStyle().Foreground(color).Render(text)
}

func parsePours(s string) ([]int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, fmt.Errorf("Введите хотя бы один пролив")
	}

	parts := strings.Split(s, ",")
	pours := make([]int, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		value, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("'%s' — не число", part)
		}
		if value <= 0 {
			return nil, fmt.Errorf("пролив должен быть > 0 секунд")
		}

		pours = append(pours, value)
	}

	if len(pours) == 0 {
		return nil, fmt.Errorf("Введите хотя бы один пролив")
	}

	return pours, nil
}
