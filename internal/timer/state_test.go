package timer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteAndReadState(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")

	state := &State{
		PID:          12345,
		TeaName:      "Шен Пуэр",
		PourIndex:    2,
		TotalPours:   8,
		RemainingSec: 7,
		Status:       StatusCounting,
	}

	if err := WriteState(state, path); err != nil {
		t.Fatalf("WriteState: %v", err)
	}

	loaded, err := ReadState(path)
	if err != nil {
		t.Fatalf("ReadState: %v", err)
	}
	if loaded.PID != 12345 {
		t.Errorf("PID = %d, want 12345", loaded.PID)
	}
	if loaded.TeaName != "Шен Пуэр" {
		t.Errorf("TeaName = %q, want 'Шен Пуэр'", loaded.TeaName)
	}
	if loaded.PourIndex != 2 {
		t.Errorf("PourIndex = %d, want 2", loaded.PourIndex)
	}
	if loaded.Status != StatusCounting {
		t.Errorf("Status = %q, want %q", loaded.Status, StatusCounting)
	}
}

func TestReadStateNoFile(t *testing.T) {
	state, err := ReadState("/tmp/nonexistent-tmux-tea-test.json")
	if err != nil {
		t.Fatalf("ReadState should not error for missing file: %v", err)
	}
	if state != nil {
		t.Error("ReadState should return nil for missing file")
	}
}

func TestClearState(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")

	state := &State{PID: 1, TeaName: "test", Status: StatusCounting}
	if err := WriteState(state, path); err != nil {
		t.Fatalf("WriteState: %v", err)
	}

	ClearState(path)

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("ClearState should remove the file")
	}
}

func TestIsProcessAlive(t *testing.T) {
	if !IsProcessAlive(os.Getpid()) {
		t.Error("current process should be alive")
	}
	if IsProcessAlive(9999999) {
		t.Error("PID 9999999 should not be alive")
	}
}

func TestStatePath(t *testing.T) {
	path := StatePath()
	if path == "" {
		t.Error("StatePath should not be empty")
	}
}
