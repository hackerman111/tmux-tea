package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSlugify(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Шен Пуэр", "shen-puer"},
		{"Hello World", "hello-world"},
		{"Да Хун Пао", "da-khun-pao"},
		{"simple", "simple"},
		{"  spaces  ", "spaces"},
	}
	for _, tt := range tests {
		got := Slugify(tt.input)
		if got == "" {
			t.Errorf("Slugify(%q) returned empty string", tt.input)
		}
		for _, r := range got {
			if r >= 'A' && r <= 'Z' {
				t.Errorf("Slugify(%q) = %q contains uppercase", tt.input, got)
			}
		}
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if len(cfg.Teas) != 3 {
		t.Fatalf("DefaultConfig should have 3 teas, got %d", len(cfg.Teas))
	}
	for _, tea := range cfg.Teas {
		if tea.ID == "" {
			t.Error("tea ID should not be empty")
		}
		if tea.Name == "" {
			t.Error("tea Name should not be empty")
		}
		if len(tea.Schedules) == 0 {
			t.Errorf("tea %q should have at least one schedule", tea.Name)
		}
		for _, schedule := range tea.Schedules {
			if len(schedule.Pours) == 0 {
				t.Errorf("schedule %q should have at least one pour", schedule.Name)
			}
			for i, pour := range schedule.Pours {
				if pour <= 0 {
					t.Errorf("schedule %q pour[%d] = %d, should be > 0", schedule.Name, i, pour)
				}
			}
		}
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "teas.json")

	cfg := DefaultConfig()
	if err := Save(cfg, path); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if len(loaded.Teas) != len(cfg.Teas) {
		t.Fatalf("loaded %d teas, want %d", len(loaded.Teas), len(cfg.Teas))
	}
	if loaded.Teas[0].Name != cfg.Teas[0].Name {
		t.Errorf("loaded name %q, want %q", loaded.Teas[0].Name, cfg.Teas[0].Name)
	}
}

func TestLoadOrCreateNewFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "subdir", "teas.json")

	cfg, err := LoadOrCreate(path)
	if err != nil {
		t.Fatalf("LoadOrCreate failed: %v", err)
	}
	if len(cfg.Teas) != 3 {
		t.Fatalf("expected default 3 teas, got %d", len(cfg.Teas))
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("file should have been created")
	}
}

func TestLoadOrCreateExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "teas.json")

	original := &Config{
		Teas: []Tea{
			{
				ID:   "test",
				Name: "Test Tea",
				Schedules: []Schedule{
					{ID: "s1", Name: "S1", Pours: []int{5, 10}},
				},
			},
		},
	}
	if err := Save(original, path); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := LoadOrCreate(path)
	if err != nil {
		t.Fatalf("LoadOrCreate failed: %v", err)
	}
	if len(loaded.Teas) != 1 {
		t.Fatalf("expected 1 tea, got %d", len(loaded.Teas))
	}
	if loaded.Teas[0].ID != "test" {
		t.Errorf("expected ID 'test', got %q", loaded.Teas[0].ID)
	}
}

func TestFindTea(t *testing.T) {
	cfg := DefaultConfig()
	tea := cfg.FindTea(cfg.Teas[0].ID)
	if tea == nil {
		t.Fatal("FindTea should find first tea")
	}
	if cfg.FindTea("nonexistent") != nil {
		t.Error("FindTea should return nil for nonexistent ID")
	}
}

func TestFindSchedule(t *testing.T) {
	cfg := DefaultConfig()
	tea := &cfg.Teas[0]
	schedule := tea.FindSchedule(tea.Schedules[0].ID)
	if schedule == nil {
		t.Fatal("FindSchedule should find first schedule")
	}
	if tea.FindSchedule("nonexistent") != nil {
		t.Error("FindSchedule should return nil for nonexistent ID")
	}
}
