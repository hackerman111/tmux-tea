package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/papayka/tmux-tea/internal/config"
)

func sampleTeas() []config.Tea {
	return []config.Tea{
		{ID: "shen-puer", Name: "Шен Пуэр"},
		{ID: "shu-puer", Name: "Шу Пуэр"},
	}
}

func TestNewMenuModelStartsAtFirstTea(t *testing.T) {
	model := NewMenuModel(sampleTeas())

	if model.cursor != 0 {
		t.Fatalf("cursor = %d, want 0", model.cursor)
	}
	if model.Result() != nil {
		t.Fatal("result should be nil on init")
	}
}

func TestMenuModelMovesCursorAndSelectsTea(t *testing.T) {
	model := NewMenuModel(sampleTeas())

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = updated.(MenuModel)
	if model.cursor != 1 {
		t.Fatalf("cursor = %d, want 1", model.cursor)
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(MenuModel)
	result := model.Result()
	if result == nil {
		t.Fatal("result should be set after enter")
	}
	if result.Action != "select" {
		t.Fatalf("action = %q, want select", result.Action)
	}
	if result.Tea == nil || result.Tea.ID != "shu-puer" {
		t.Fatalf("selected tea = %#v, want shu-puer", result.Tea)
	}
	if result.TeaIndex != 1 {
		t.Fatalf("TeaIndex = %d, want 1", result.TeaIndex)
	}
}

func TestMenuModelShortcuts(t *testing.T) {
	t.Run("quit", func(t *testing.T) {
		model := NewMenuModel(sampleTeas())
		updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEsc})
		model = updated.(MenuModel)

		if !model.quitting {
			t.Fatal("model should be quitting")
		}
		if model.Result() == nil || model.Result().Action != "quit" {
			t.Fatalf("result = %#v, want quit", model.Result())
		}
	})

	t.Run("add", func(t *testing.T) {
		model := NewMenuModel(sampleTeas())
		updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
		model = updated.(MenuModel)

		if model.Result() == nil || model.Result().Action != "add" {
			t.Fatalf("result = %#v, want add", model.Result())
		}
	})

	t.Run("edit", func(t *testing.T) {
		model := NewMenuModel(sampleTeas())
		updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
		model = updated.(MenuModel)

		if model.Result() == nil || model.Result().Action != "edit" {
			t.Fatalf("result = %#v, want edit", model.Result())
		}
		if model.Result().Tea == nil || model.Result().Tea.ID != "shen-puer" {
			t.Fatalf("result tea = %#v, want shen-puer", model.Result().Tea)
		}
	})

	t.Run("delete requires more than one tea", func(t *testing.T) {
		model := NewMenuModel([]config.Tea{{ID: "single", Name: "Single"}})
		updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
		model = updated.(MenuModel)

		if model.Result() != nil {
			t.Fatalf("result = %#v, want nil", model.Result())
		}
	})

	t.Run("delete", func(t *testing.T) {
		model := NewMenuModel(sampleTeas())
		updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
		model = updated.(MenuModel)

		if model.Result() == nil || model.Result().Action != "delete" {
			t.Fatalf("result = %#v, want delete", model.Result())
		}
	})
}

func TestMenuModelViewContainsTeaNamesAndHelp(t *testing.T) {
	view := NewMenuModel(sampleTeas()).View()

	for _, want := range []string{
		"Выберите чай:",
		"Шен Пуэр",
		"Шу Пуэр",
		"добавить",
		"enter старт",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("view should contain %q, got %q", want, view)
		}
	}
}

func TestMenuModelViewTruncatesLongTeaNames(t *testing.T) {
	longName := "Очень длинное название чая для узкого tmux popup окна"
	view := NewMenuModel([]config.Tea{{ID: "long", Name: longName}}).View()

	if strings.Contains(view, longName) {
		t.Fatalf("view should truncate long tea names, got %q", view)
	}
	if !strings.Contains(view, "...") {
		t.Fatalf("view should contain truncation marker, got %q", view)
	}

	for _, line := range strings.Split(strings.TrimRight(view, "\n"), "\n") {
		if width := lipgloss.Width(line); width > 56 {
			t.Fatalf("line width = %d, want <= 56: %q", width, line)
		}
	}
}
