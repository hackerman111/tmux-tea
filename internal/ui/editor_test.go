package ui

import (
	"reflect"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewTeaEditorModel(t *testing.T) {
	model := NewTeaEditorModel("Шен Пуэр", true)

	if model.title != "Новый чай" {
		t.Fatalf("title = %q, want %q", model.title, "Новый чай")
	}
	if model.nameInput.Value() != "Шен Пуэр" {
		t.Fatalf("name input = %q, want %q", model.nameInput.Value(), "Шен Пуэр")
	}
	if model.Result() != nil {
		t.Fatal("result should be nil on init")
	}
}

func TestTeaEditorModelSaveAndCancel(t *testing.T) {
	t.Run("save trimmed name", func(t *testing.T) {
		model := NewTeaEditorModel("  Да Хун Пао  ", false)
		updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
		model = updated.(TeaEditorModel)

		result := model.Result()
		if result == nil {
			t.Fatal("result should be set after save")
		}
		if !result.Saved {
			t.Fatal("Saved should be true")
		}
		if result.Name != "Да Хун Пао" {
			t.Fatalf("Name = %q, want %q", result.Name, "Да Хун Пао")
		}
	})

	t.Run("empty name does not save", func(t *testing.T) {
		model := NewTeaEditorModel("   ", true)
		updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
		model = updated.(TeaEditorModel)

		if model.Result() != nil {
			t.Fatalf("result = %#v, want nil", model.Result())
		}
	})

	t.Run("cancel", func(t *testing.T) {
		model := NewTeaEditorModel("Шен Пуэр", false)
		updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEsc})
		model = updated.(TeaEditorModel)

		result := model.Result()
		if result == nil {
			t.Fatal("result should be set after cancel")
		}
		if result.Saved {
			t.Fatal("Saved should be false on cancel")
		}
	})
}

func TestTeaEditorModelViewContainsTitleAndHelp(t *testing.T) {
	view := NewTeaEditorModel("Шен Пуэр", false).View()

	for _, want := range []string{
		"Редактировать чай",
		"Название:",
		"сохранить",
		"отмена",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("view should contain %q, got %q", want, view)
		}
	}
}

func TestNewScheduleEditorModel(t *testing.T) {
	model := NewScheduleEditorModel("Стандарт", []int{10, 15, 20}, true)

	if model.title != "Новое расписание" {
		t.Fatalf("title = %q, want %q", model.title, "Новое расписание")
	}
	if model.nameInput.Value() != "Стандарт" {
		t.Fatalf("name input = %q, want %q", model.nameInput.Value(), "Стандарт")
	}
	if model.poursInput.Value() != "10,15,20" {
		t.Fatalf("pours input = %q, want %q", model.poursInput.Value(), "10,15,20")
	}
	if model.focusIndex != 0 {
		t.Fatalf("focusIndex = %d, want 0", model.focusIndex)
	}
}

func TestParsePours(t *testing.T) {
	t.Run("valid values", func(t *testing.T) {
		got, err := parsePours("10, 15,20")
		if err != nil {
			t.Fatalf("parsePours returned error: %v", err)
		}
		want := []int{10, 15, 20}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("parsePours = %v, want %v", got, want)
		}
	})

	t.Run("empty input", func(t *testing.T) {
		_, err := parsePours("   ")
		if err == nil || !strings.Contains(err.Error(), "Введите хотя бы один пролив") {
			t.Fatalf("error = %v, want empty pours message", err)
		}
	})

	t.Run("non numeric value", func(t *testing.T) {
		_, err := parsePours("10, abc")
		if err == nil || !strings.Contains(err.Error(), "не число") {
			t.Fatalf("error = %v, want non numeric message", err)
		}
	})

	t.Run("non positive value", func(t *testing.T) {
		_, err := parsePours("10, 0")
		if err == nil || !strings.Contains(err.Error(), "пролив должен быть > 0 секунд") {
			t.Fatalf("error = %v, want positive validation message", err)
		}
	})
}

func TestScheduleEditorModelSaveValidationAndCancel(t *testing.T) {
	t.Run("save valid schedule", func(t *testing.T) {
		model := NewScheduleEditorModel("Стандарт", []int{10, 15}, false)
		updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
		model = updated.(ScheduleEditorModel)

		result := model.Result()
		if result == nil {
			t.Fatal("result should be set after save")
		}
		if !result.Saved {
			t.Fatal("Saved should be true")
		}
		if result.Name != "Стандарт" {
			t.Fatalf("Name = %q, want %q", result.Name, "Стандарт")
		}
		wantPours := []int{10, 15}
		if !reflect.DeepEqual(result.Pours, wantPours) {
			t.Fatalf("Pours = %v, want %v", result.Pours, wantPours)
		}
	})

	t.Run("tab switches focus", func(t *testing.T) {
		model := NewScheduleEditorModel("Стандарт", []int{10, 15}, false)
		updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyTab})
		model = updated.(ScheduleEditorModel)

		if model.focusIndex != 1 {
			t.Fatalf("focusIndex = %d, want 1", model.focusIndex)
		}
	})

	t.Run("empty name shows validation error", func(t *testing.T) {
		model := NewScheduleEditorModel("", []int{10, 15}, true)
		updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
		model = updated.(ScheduleEditorModel)

		if model.Result() != nil {
			t.Fatalf("result = %#v, want nil", model.Result())
		}
		if model.err != "Название не может быть пустым" {
			t.Fatalf("err = %q, want %q", model.err, "Название не может быть пустым")
		}
	})

	t.Run("invalid pours show validation error", func(t *testing.T) {
		model := NewScheduleEditorModel("Стандарт", nil, true)
		updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
		model = updated.(ScheduleEditorModel)

		if model.Result() != nil {
			t.Fatalf("result = %#v, want nil", model.Result())
		}
		if model.err != "Введите хотя бы один пролив" {
			t.Fatalf("err = %q, want %q", model.err, "Введите хотя бы один пролив")
		}
	})

	t.Run("cancel", func(t *testing.T) {
		model := NewScheduleEditorModel("Стандарт", []int{10, 15}, false)
		updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEsc})
		model = updated.(ScheduleEditorModel)

		result := model.Result()
		if result == nil {
			t.Fatal("result should be set after cancel")
		}
		if result.Saved {
			t.Fatal("Saved should be false on cancel")
		}
	})
}

func TestScheduleEditorModelViewContainsFieldsAndHelp(t *testing.T) {
	view := NewScheduleEditorModel("Стандарт", []int{10, 15, 20}, false).View()

	for _, want := range []string{
		"Редактировать расписание",
		"Название:",
		"Проливы",
		"tab поле",
		"enter сохранить",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("view should contain %q, got %q", want, view)
		}
	}
}
