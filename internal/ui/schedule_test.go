package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/papayka/tmux-tea/internal/config"
)

func sampleSchedules() []config.Schedule {
	return []config.Schedule{
		{ID: "default", Name: "Стандарт", Pours: []int{10, 15, 20}},
		{ID: "fast", Name: "Быстрый", Pours: []int{5, 10, 10, 15, 15, 20}},
	}
}

func TestNewScheduleModelStartsAtFirstSchedule(t *testing.T) {
	model := NewScheduleModel("Шен Пуэр", sampleSchedules())

	if model.cursor != 0 {
		t.Fatalf("cursor = %d, want 0", model.cursor)
	}
	if model.Result() != nil {
		t.Fatal("result should be nil on init")
	}
}

func TestFormatPours(t *testing.T) {
	if got := formatPours(nil); got != "" {
		t.Fatalf("formatPours(nil) = %q, want empty", got)
	}

	if got := formatPours([]int{10, 15, 20}); got != "(10,15,20)" {
		t.Fatalf("formatPours(short) = %q, want (10,15,20)", got)
	}

	got := formatPours([]int{5, 10, 10, 15, 15, 20, 25, 30, 35})
	if !strings.HasPrefix(got, "(") || !strings.HasSuffix(got, ")") {
		t.Fatalf("formatPours(long) = %q, want wrapped preview", got)
	}
	if !strings.Contains(got, "..") {
		t.Fatalf("formatPours(long) = %q, want truncated preview", got)
	}
}

func TestScheduleModelMovesCursorAndSelectsSchedule(t *testing.T) {
	model := NewScheduleModel("Шен Пуэр", sampleSchedules())

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = updated.(ScheduleModel)
	if model.cursor != 1 {
		t.Fatalf("cursor = %d, want 1", model.cursor)
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(ScheduleModel)
	result := model.Result()
	if result == nil {
		t.Fatal("result should be set after enter")
	}
	if result.Action != "select" {
		t.Fatalf("action = %q, want select", result.Action)
	}
	if result.Schedule == nil || result.Schedule.ID != "fast" {
		t.Fatalf("selected schedule = %#v, want fast", result.Schedule)
	}
	if result.ScheduleIndex != 1 {
		t.Fatalf("ScheduleIndex = %d, want 1", result.ScheduleIndex)
	}
}

func TestScheduleModelShortcuts(t *testing.T) {
	t.Run("back", func(t *testing.T) {
		model := NewScheduleModel("Шен Пуэр", sampleSchedules())
		updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEsc})
		model = updated.(ScheduleModel)

		if model.Result() == nil || model.Result().Action != "back" {
			t.Fatalf("result = %#v, want back", model.Result())
		}
	})

	t.Run("add", func(t *testing.T) {
		model := NewScheduleModel("Шен Пуэр", sampleSchedules())
		updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
		model = updated.(ScheduleModel)

		if model.Result() == nil || model.Result().Action != "add" {
			t.Fatalf("result = %#v, want add", model.Result())
		}
	})

	t.Run("edit", func(t *testing.T) {
		model := NewScheduleModel("Шен Пуэр", sampleSchedules())
		updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
		model = updated.(ScheduleModel)

		if model.Result() == nil || model.Result().Action != "edit" {
			t.Fatalf("result = %#v, want edit", model.Result())
		}
		if model.Result().Schedule == nil || model.Result().Schedule.ID != "default" {
			t.Fatalf("result schedule = %#v, want default", model.Result().Schedule)
		}
	})

	t.Run("delete requires more than one schedule", func(t *testing.T) {
		model := NewScheduleModel("Шен Пуэр", []config.Schedule{{ID: "single", Name: "Single", Pours: []int{10}}})
		updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
		model = updated.(ScheduleModel)

		if model.Result() != nil {
			t.Fatalf("result = %#v, want nil", model.Result())
		}
	})

	t.Run("delete", func(t *testing.T) {
		model := NewScheduleModel("Шен Пуэр", sampleSchedules())
		updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
		model = updated.(ScheduleModel)

		if model.Result() == nil || model.Result().Action != "delete" {
			t.Fatalf("result = %#v, want delete", model.Result())
		}
	})
}

func TestScheduleModelViewContainsTeaNamePreviewAndHelp(t *testing.T) {
	view := NewScheduleModel("Шен Пуэр", sampleSchedules()).View()

	for _, want := range []string{
		"Шен Пуэр",
		"Стандарт",
		"Быстрый",
		"добавить",
		"esc назад",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("view should contain %q, got %q", want, view)
		}
	}
}
