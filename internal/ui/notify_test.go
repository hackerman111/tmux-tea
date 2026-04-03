package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewNotifyModel(t *testing.T) {
	model := NewNotifyModel("Шен Пуэр", 1, 5, false)

	if model.teaName != "Шен Пуэр" {
		t.Fatalf("teaName = %q, want %q", model.teaName, "Шен Пуэр")
	}
	if model.pourIndex != 1 {
		t.Fatalf("pourIndex = %d, want 1", model.pourIndex)
	}
	if model.totalPours != 5 {
		t.Fatalf("totalPours = %d, want 5", model.totalPours)
	}
	if model.finished {
		t.Fatal("finished should be false")
	}
	if model.Confirmed() {
		t.Fatal("confirmed should be false on init")
	}
}

func TestNotifyModelConfirmKeys(t *testing.T) {
	keys := []tea.KeyMsg{
		{Type: tea.KeyEnter},
		{Type: tea.KeyEsc},
		{Type: tea.KeySpace},
	}

	for _, keyMsg := range keys {
		model := NewNotifyModel("Шен Пуэр", 0, 3, false)
		updated, _ := model.Update(keyMsg)
		model = updated.(NotifyModel)

		if !model.Confirmed() {
			t.Fatalf("key %v should confirm notification", keyMsg.Type)
		}
	}
}

func TestNotifyModelViewShowsActivePour(t *testing.T) {
	view := NewNotifyModel("Шен Пуэр", 1, 5, false).View()

	for _, want := range []string{
		"Шен Пуэр",
		"пролив 2/5",
		"Enter закрыть, затем Prefix+T",
		"████████╗",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("view should contain %q, got %q", want, view)
		}
	}
}

func TestNotifyModelViewShowsFinishedState(t *testing.T) {
	view := NewNotifyModel("Шен Пуэр", 4, 5, true).View()

	for _, want := range []string{
		"Чаепитие завершено!",
		"Нажмите Enter",
		"██████╗",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("view should contain %q, got %q", want, view)
		}
	}
}
