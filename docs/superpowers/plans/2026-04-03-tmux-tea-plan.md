# tmux-tea Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a tmux plugin that provides a tea ceremony timer with TUI for selecting teas/schedules, countdown in status bar, and fullscreen ASCII notifications.

**Architecture:** Monolithic Go binary (`tmux-tea`) with subcommands (menu, start, status, confirm, stop). Bash wrapper (`tea.tmux`) registers tmux hotkeys. Timer state persisted to `/tmp/tmux-tea-state.json`, config to `~/.config/tmux-tea/teas.json`.

**Tech Stack:** Go 1.21+, Bubbletea, Lipgloss, Cobra, tmux 3.3+

---

## File Map

| File | Responsibility |
|------|---------------|
| `go.mod` | Module definition, dependencies |
| `cmd/tmux-tea/main.go` | CLI entry point, cobra root + subcommands |
| `internal/config/config.go` | Tea/Schedule/Config types, JSON load/save, defaults, slugify |
| `internal/timer/state.go` | TimerState type, state file read/write/cleanup |
| `internal/timer/timer.go` | Timer loop: countdown, notify trigger, confirm handler |
| `internal/tmux/tmux.go` | Exec wrappers: display-popup, run-shell, set-option, send bell |
| `internal/ui/styles.go` | Lipgloss styles, colors, borders |
| `internal/ui/menu.go` | Tea list screen (bubbletea Model) |
| `internal/ui/schedule.go` | Schedule list screen (bubbletea Model) |
| `internal/ui/editor.go` | Tea/Schedule editor forms (bubbletea Model) |
| `internal/ui/notify.go` | TEA TIME ASCII-art screen (bubbletea Model) |
| `tea.tmux` | TPM entry point, hotkey bindings, status-right |
| `scripts/install.sh` | Build Go binary to `bin/tmux-tea` |

---

### Task 1: Go Module & Config Package

**Files:**
- Create: `go.mod`
- Create: `internal/config/config.go`
- Create: `internal/config/config_test.go`

- [ ] **Step 1: Initialize Go module**

```bash
cd /home/papayka/my_plug/tmux-tea
go mod init github.com/papayka/tmux-tea
```

- [ ] **Step 2: Write config tests**

Create `internal/config/config_test.go`:

```go
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
		// Just check it's non-empty and lowercase — transliteration is best-effort
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
		for _, s := range tea.Schedules {
			if len(s.Pours) == 0 {
				t.Errorf("schedule %q should have at least one pour", s.Name)
			}
			for i, p := range s.Pours {
				if p <= 0 {
					t.Errorf("schedule %q pour[%d] = %d, should be > 0", s.Name, i, p)
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

func TestLoadOrCreate_NewFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "subdir", "teas.json")

	cfg, err := LoadOrCreate(path)
	if err != nil {
		t.Fatalf("LoadOrCreate failed: %v", err)
	}
	if len(cfg.Teas) != 3 {
		t.Fatalf("expected default 3 teas, got %d", len(cfg.Teas))
	}
	// File should now exist
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("file should have been created")
	}
}

func TestLoadOrCreate_ExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "teas.json")

	original := &Config{
		Teas: []Tea{
			{ID: "test", Name: "Test Tea", Schedules: []Schedule{
				{ID: "s1", Name: "S1", Pours: []int{5, 10}},
			}},
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
	s := tea.FindSchedule(tea.Schedules[0].ID)
	if s == nil {
		t.Fatal("FindSchedule should find first schedule")
	}
	if tea.FindSchedule("nonexistent") != nil {
		t.Error("FindSchedule should return nil for nonexistent ID")
	}
}
```

- [ ] **Step 3: Run tests to verify they fail**

```bash
cd /home/papayka/my_plug/tmux-tea
go test ./internal/config/ -v
```

Expected: compilation errors (types and functions not defined).

- [ ] **Step 4: Implement config package**

Create `internal/config/config.go`:

```go
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

type Schedule struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Pours []int  `json:"pours"`
}

type Tea struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Schedules []Schedule `json:"schedules"`
}

func (t *Tea) FindSchedule(id string) *Schedule {
	for i := range t.Schedules {
		if t.Schedules[i].ID == id {
			return &t.Schedules[i]
		}
	}
	return nil
}

type Config struct {
	Teas []Tea `json:"teas"`
}

func (c *Config) FindTea(id string) *Tea {
	for i := range c.Teas {
		if c.Teas[i].ID == id {
			return &c.Teas[i]
		}
	}
	return nil
}

var nonAlphaNum = regexp.MustCompile(`[^a-z0-9]+`)

func Slugify(s string) string {
	// Normalize unicode, transliterate cyrillic naively
	s = transliterate(s)
	s = strings.ToLower(s)
	s = nonAlphaNum.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}

func transliterate(s string) string {
	// NFD decomposition to strip accents
	t := transform.Chain(norm.NFD, transform.RemoveFunc(func(r rune) bool {
		return unicode.Is(unicode.Mn, r)
	}), norm.NFC)
	result, _, _ := transform.String(t, s)

	// Basic cyrillic transliteration
	cyrillic := map[rune]string{
		'а': "a", 'б': "b", 'в': "v", 'г': "g", 'д': "d", 'е': "e", 'ё': "yo",
		'ж': "zh", 'з': "z", 'и': "i", 'й': "y", 'к': "k", 'л': "l", 'м': "m",
		'н': "n", 'о': "o", 'п': "p", 'р': "r", 'с': "s", 'т': "t", 'у': "u",
		'ф': "f", 'х': "kh", 'ц': "ts", 'ч': "ch", 'ш': "sh", 'щ': "shch",
		'ъ': "", 'ы': "y", 'ь': "", 'э': "e", 'ю': "yu", 'я': "ya",
		'А': "A", 'Б': "B", 'В': "V", 'Г': "G", 'Д': "D", 'Е': "E", 'Ё': "Yo",
		'Ж': "Zh", 'З': "Z", 'И': "I", 'Й': "Y", 'К': "K", 'Л': "L", 'М': "M",
		'Н': "N", 'О': "O", 'П': "P", 'Р': "R", 'С': "S", 'Т': "T", 'У': "U",
		'Ф': "F", 'Х': "Kh", 'Ц': "Ts", 'Ч': "Ch", 'Ш': "Sh", 'Щ': "Shch",
		'Ъ': "", 'Ы': "Y", 'Ь': "", 'Э': "E", 'Ю': "Yu", 'Я': "Ya",
	}

	var b strings.Builder
	for _, r := range result {
		if tr, ok := cyrillic[r]; ok {
			b.WriteString(tr)
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func DefaultConfig() *Config {
	return &Config{
		Teas: []Tea{
			{
				ID:   "shen-puer",
				Name: "Шен Пуэр",
				Schedules: []Schedule{
					{ID: "default", Name: "Стандарт", Pours: []int{10, 15, 20, 25, 30, 40, 50, 60}},
					{ID: "fast", Name: "Быстрый", Pours: []int{5, 10, 10, 15, 15, 20}},
				},
			},
			{
				ID:   "shu-puer",
				Name: "Шу Пуэр",
				Schedules: []Schedule{
					{ID: "default", Name: "Стандарт", Pours: []int{10, 15, 15, 20, 25, 30, 40, 50}},
				},
			},
			{
				ID:   "da-khun-pao",
				Name: "Да Хун Пао",
				Schedules: []Schedule{
					{ID: "default", Name: "Стандарт", Pours: []int{15, 20, 25, 30, 35, 45, 60}},
				},
			},
		},
	}
}

func ConfigDir() string {
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return filepath.Join(dir, "tmux-tea")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "tmux-tea")
}

func DefaultPath() string {
	return filepath.Join(ConfigDir(), "teas.json")
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	return &cfg, nil
}

func Save(cfg *Config, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}
	return os.WriteFile(path, data, 0o644)
}

func LoadOrCreate(path string) (*Config, error) {
	cfg, err := Load(path)
	if err == nil {
		return cfg, nil
	}
	if !os.IsNotExist(err) {
		return nil, err
	}
	cfg = DefaultConfig()
	if err := Save(cfg, path); err != nil {
		return nil, err
	}
	return cfg, nil
}
```

- [ ] **Step 5: Add x/text dependency and run tests**

```bash
cd /home/papayka/my_plug/tmux-tea
go get golang.org/x/text
go test ./internal/config/ -v
```

Expected: all tests PASS.

- [ ] **Step 6: Commit**

```bash
git add go.mod go.sum internal/config/
git commit -m "feat: add config package with tea/schedule types, JSON persistence, defaults"
```

---

### Task 2: Timer State Package

**Files:**
- Create: `internal/timer/state.go`
- Create: `internal/timer/state_test.go`

- [ ] **Step 1: Write state tests**

Create `internal/timer/state_test.go`:

```go
package timer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteAndReadState(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")

	s := &State{
		PID:          12345,
		TeaName:      "Шен Пуэр",
		PourIndex:    2,
		TotalPours:   8,
		RemainingSec: 7,
		Status:       StatusCounting,
	}

	if err := WriteState(s, path); err != nil {
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

func TestReadState_NoFile(t *testing.T) {
	s, err := ReadState("/tmp/nonexistent-tmux-tea-test.json")
	if err != nil {
		t.Fatalf("ReadState should not error for missing file: %v", err)
	}
	if s != nil {
		t.Error("ReadState should return nil for missing file")
	}
}

func TestClearState(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")

	s := &State{PID: 1, TeaName: "test", Status: StatusCounting}
	if err := WriteState(s, path); err != nil {
		t.Fatalf("WriteState: %v", err)
	}

	ClearState(path)

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("ClearState should remove the file")
	}
}

func TestIsProcessAlive(t *testing.T) {
	// Current process should be alive
	if !IsProcessAlive(os.Getpid()) {
		t.Error("current process should be alive")
	}
	// PID 0 should not be considered alive in our check
	// Very high PID should not exist
	if IsProcessAlive(9999999) {
		t.Error("PID 9999999 should not be alive")
	}
}

func TestStatePath(t *testing.T) {
	p := StatePath()
	if p == "" {
		t.Error("StatePath should not be empty")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./internal/timer/ -v
```

Expected: compilation errors.

- [ ] **Step 3: Implement state package**

Create `internal/timer/state.go`:

```go
package timer

import (
	"encoding/json"
	"fmt"
	"os"
	"syscall"
)

const (
	StatusCounting = "counting"
	StatusReady    = "ready"
)

type State struct {
	PID          int    `json:"pid"`
	TeaName      string `json:"tea_name"`
	PourIndex    int    `json:"pour_index"`
	TotalPours   int    `json:"total_pours"`
	RemainingSec int    `json:"remaining_sec"`
	Status       string `json:"status"`
}

func StatePath() string {
	return "/tmp/tmux-tea-state.json"
}

func ReadState(path string) (*State, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading state: %w", err)
	}
	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parsing state: %w", err)
	}
	return &s, nil
}

func WriteState(s *State, path string) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling state: %w", err)
	}
	return os.WriteFile(path, data, 0o644)
}

func ClearState(path string) {
	os.Remove(path)
}

func IsProcessAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	err := syscall.Kill(pid, 0)
	return err == nil
}
```

- [ ] **Step 4: Run tests**

```bash
go test ./internal/timer/ -v
```

Expected: all tests PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/timer/
git commit -m "feat: add timer state read/write/clear with process liveness check"
```

---

### Task 3: Tmux Wrapper Package

**Files:**
- Create: `internal/tmux/tmux.go`
- Create: `internal/tmux/tmux_test.go`

- [ ] **Step 1: Write tests**

Create `internal/tmux/tmux_test.go`:

```go
package tmux

import "testing"

func TestBuildArgs(t *testing.T) {
	// Test that command building works correctly
	args := buildArgs("display-popup", "-E", "-w", "40", "-h", "15", "tmux-tea menu")
	if len(args) != 7 {
		t.Fatalf("expected 7 args, got %d: %v", len(args), args)
	}
	if args[0] != "display-popup" {
		t.Errorf("args[0] = %q, want 'display-popup'", args[0])
	}
}

func TestFormatStatusOutput_Empty(t *testing.T) {
	out := FormatStatus(nil)
	if out != "" {
		t.Errorf("FormatStatus(nil) = %q, want empty", out)
	}
}

func TestFormatStatusOutput_Counting(t *testing.T) {
	out := FormatStatus(&StatusInfo{
		TeaName:      "Шен Пуэр",
		PourIndex:    3,
		TotalPours:   8,
		RemainingSec: 72,
		Status:       "counting",
	})
	// Should contain tea name and time
	if out == "" {
		t.Error("FormatStatus should not be empty for active timer")
	}
}

func TestFormatStatusOutput_Ready(t *testing.T) {
	out := FormatStatus(&StatusInfo{
		TeaName:    "Шен Пуэр",
		PourIndex:  3,
		TotalPours: 8,
		Status:     "ready",
	})
	if out == "" {
		t.Error("FormatStatus should not be empty for ready state")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./internal/tmux/ -v
```

Expected: compilation errors.

- [ ] **Step 3: Implement tmux package**

Create `internal/tmux/tmux.go`:

```go
package tmux

import (
	"fmt"
	"os/exec"
)

type StatusInfo struct {
	TeaName      string
	PourIndex    int
	TotalPours   int
	RemainingSec int
	Status       string
}

func buildArgs(args ...string) []string {
	return args
}

func Run(args ...string) error {
	cmd := exec.Command("tmux", args...)
	return cmd.Run()
}

func RunOutput(args ...string) (string, error) {
	cmd := exec.Command("tmux", args...)
	out, err := cmd.Output()
	return string(out), err
}

func DisplayPopup(width, height int, command string) error {
	return Run("display-popup", "-E",
		"-w", fmt.Sprintf("%d", width),
		"-h", fmt.Sprintf("%d", height),
		command)
}

func RunShellBg(command string) error {
	return Run("run-shell", "-b", command)
}

func SendBell() error {
	return Run("run-shell", "printf '\\a'")
}

func SetOption(option, value string) error {
	return Run("set-option", "-g", option, value)
}

func GetOption(option string) (string, error) {
	return RunOutput("show-option", "-gv", option)
}

func FormatStatus(info *StatusInfo) string {
	if info == nil {
		return ""
	}
	switch info.Status {
	case "ready":
		return fmt.Sprintf("🍵 %s [%d/%d] READY",
			info.TeaName, info.PourIndex+1, info.TotalPours)
	case "counting":
		mins := info.RemainingSec / 60
		secs := info.RemainingSec % 60
		return fmt.Sprintf("🍵 %s [%d/%d] %d:%02d",
			info.TeaName, info.PourIndex+1, info.TotalPours, mins, secs)
	default:
		return ""
	}
}
```

- [ ] **Step 4: Run tests**

```bash
go test ./internal/tmux/ -v
```

Expected: all tests PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/tmux/
git commit -m "feat: add tmux wrapper with popup, bell, status formatting"
```

---

### Task 4: Timer Loop

**Files:**
- Create: `internal/timer/timer.go`
- Create: `internal/timer/timer_test.go`

- [ ] **Step 1: Write timer tests**

Create `internal/timer/timer_test.go`:

```go
package timer

import (
	"path/filepath"
	"testing"
)

func TestNewSession(t *testing.T) {
	dir := t.TempDir()
	statePath := filepath.Join(dir, "state.json")

	s := NewSession("Шен Пуэр", []int{10, 15, 20}, statePath)
	if s.TeaName != "Шен Пуэр" {
		t.Errorf("TeaName = %q, want 'Шен Пуэр'", s.TeaName)
	}
	if s.TotalPours != 3 {
		t.Errorf("TotalPours = %d, want 3", s.TotalPours)
	}
	if s.CurrentPour != 0 {
		t.Errorf("CurrentPour = %d, want 0", s.CurrentPour)
	}
}

func TestSession_NextPour(t *testing.T) {
	dir := t.TempDir()
	statePath := filepath.Join(dir, "state.json")

	s := NewSession("test", []int{5, 10}, statePath)
	if s.IsFinished() {
		t.Error("should not be finished at start")
	}

	dur := s.CurrentPourDuration()
	if dur != 5 {
		t.Errorf("first pour = %d, want 5", dur)
	}

	s.AdvancePour()
	if s.CurrentPour != 1 {
		t.Errorf("CurrentPour = %d, want 1", s.CurrentPour)
	}
	dur = s.CurrentPourDuration()
	if dur != 10 {
		t.Errorf("second pour = %d, want 10", dur)
	}

	s.AdvancePour()
	if !s.IsFinished() {
		t.Error("should be finished after all pours")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./internal/timer/ -v -run TestNewSession
```

Expected: compilation errors.

- [ ] **Step 3: Implement timer**

Create `internal/timer/timer.go`:

```go
package timer

import (
	"fmt"
	"os"
	"time"
)

type Session struct {
	TeaName     string
	Pours       []int
	CurrentPour int
	TotalPours  int
	StatePath   string
}

func NewSession(teaName string, pours []int, statePath string) *Session {
	return &Session{
		TeaName:     teaName,
		Pours:       pours,
		CurrentPour: 0,
		TotalPours:  len(pours),
		StatePath:   statePath,
	}
}

func (s *Session) CurrentPourDuration() int {
	if s.CurrentPour >= len(s.Pours) {
		return 0
	}
	return s.Pours[s.CurrentPour]
}

func (s *Session) AdvancePour() {
	s.CurrentPour++
}

func (s *Session) IsFinished() bool {
	return s.CurrentPour >= s.TotalPours
}

// Run executes the countdown for the current pour.
// It writes state every second and calls onReady when the pour is done.
// onReady should trigger the notification (popup + bell).
func (s *Session) RunCurrentPour(onReady func() error) error {
	duration := s.CurrentPourDuration()
	if duration <= 0 {
		return fmt.Errorf("invalid pour duration: %d", duration)
	}

	for remaining := duration; remaining > 0; remaining-- {
		state := &State{
			PID:          os.Getpid(),
			TeaName:      s.TeaName,
			PourIndex:    s.CurrentPour,
			TotalPours:   s.TotalPours,
			RemainingSec: remaining,
			Status:       StatusCounting,
		}
		if err := WriteState(state, s.StatePath); err != nil {
			return fmt.Errorf("writing state: %w", err)
		}
		time.Sleep(1 * time.Second)
	}

	// Pour is done — set status to ready
	state := &State{
		PID:          os.Getpid(),
		TeaName:      s.TeaName,
		PourIndex:    s.CurrentPour,
		TotalPours:   s.TotalPours,
		RemainingSec: 0,
		Status:       StatusReady,
	}
	if err := WriteState(state, s.StatePath); err != nil {
		return fmt.Errorf("writing state: %w", err)
	}

	return onReady()
}
```

- [ ] **Step 4: Run all timer tests**

```bash
go test ./internal/timer/ -v
```

Expected: all tests PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/timer/timer.go internal/timer/timer_test.go
git commit -m "feat: add timer session with pour countdown and state persistence"
```

---

### Task 5: UI Styles

**Files:**
- Create: `internal/ui/styles.go`

- [ ] **Step 1: Add bubbletea and lipgloss dependencies**

```bash
cd /home/papayka/my_plug/tmux-tea
go get github.com/charmbracelet/bubbletea
go get github.com/charmbracelet/lipgloss
go get github.com/charmbracelet/bubbles
```

- [ ] **Step 2: Create styles**

Create `internal/ui/styles.go`:

```go
package ui

import "github.com/charmbracelet/lipgloss"

var (
	ColorPrimary   = lipgloss.Color("#D4A574") // warm tea color
	ColorSecondary = lipgloss.Color("#8B7355")
	ColorAccent    = lipgloss.Color("#C8E6C9") // green tea accent
	ColorMuted     = lipgloss.Color("#666666")
	ColorDanger    = lipgloss.Color("#E57373")

	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			MarginBottom(1)

	SelectedStyle = lipgloss.NewStyle().
			Foreground(ColorAccent).
			Bold(true).
			PaddingLeft(2)

	NormalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			PaddingLeft(4)

	MutedStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			PaddingLeft(4)

	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			MarginTop(1)

	BorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorSecondary).
			Padding(1, 2)

	NotifyStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			Align(lipgloss.Center)
)
```

- [ ] **Step 3: Verify compilation**

```bash
go build ./internal/ui/
```

Expected: success, no errors.

- [ ] **Step 4: Commit**

```bash
git add internal/ui/styles.go go.mod go.sum
git commit -m "feat: add lipgloss styles and color theme for TUI"
```

---

### Task 6: Tea Menu Screen

**Files:**
- Create: `internal/ui/menu.go`

- [ ] **Step 1: Implement tea menu model**

Create `internal/ui/menu.go`:

```go
package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/papayka/tmux-tea/internal/config"
)

type MenuResult struct {
	Tea      *config.Tea
	Action   string // "select", "add", "edit", "delete", "quit"
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

func (m MenuModel) Result() *MenuResult { return m.result }

func (m MenuModel) Init() tea.Cmd { return nil }

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

	for i, t := range m.teas {
		if i == m.cursor {
			s += SelectedStyle.Render(fmt.Sprintf("> %s", t.Name)) + "\n"
		} else {
			s += NormalStyle.Render(t.Name) + "\n"
		}
	}

	s += "\n"
	s += MutedStyle.Render("─────────────") + "\n"
	s += MutedStyle.Render("a добавить  e редакт.  d удалить") + "\n"
	s += HelpStyle.Render("  j/k выбор  enter старт  esc выход")

	return BorderStyle.Render(s)
}
```

- [ ] **Step 2: Verify compilation**

```bash
go build ./internal/ui/
```

Expected: success.

- [ ] **Step 3: Commit**

```bash
git add internal/ui/menu.go
git commit -m "feat: add tea menu TUI screen with selection, add, edit, delete"
```

---

### Task 7: Schedule Selection Screen

**Files:**
- Create: `internal/ui/schedule.go`

- [ ] **Step 1: Implement schedule model**

Create `internal/ui/schedule.go`:

```go
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
	Action        string // "select", "add", "edit", "delete", "back"
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

func (m ScheduleModel) Result() *ScheduleResult { return m.result }

func (m ScheduleModel) Init() tea.Cmd { return nil }

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
	for _, p := range pours {
		parts = append(parts, fmt.Sprintf("%d", p))
	}
	preview := strings.Join(parts, ",")
	if len(preview) > 20 {
		preview = preview[:20] + ".."
	}
	return "(" + preview + ")"
}

func (m ScheduleModel) View() string {
	s := TitleStyle.Render(fmt.Sprintf("%s — расписание:", m.teaName)) + "\n\n"

	for i, sch := range m.schedules {
		label := fmt.Sprintf("%s %s", sch.Name, formatPours(sch.Pours))
		if i == m.cursor {
			s += SelectedStyle.Render("> "+label) + "\n"
		} else {
			s += NormalStyle.Render(label) + "\n"
		}
	}

	s += "\n"
	s += MutedStyle.Render("─────────────") + "\n"
	s += MutedStyle.Render("a добавить  e редакт.  d удалить") + "\n"
	s += HelpStyle.Render("  esc назад  enter старт")

	return BorderStyle.Render(s)
}
```

- [ ] **Step 2: Verify compilation**

```bash
go build ./internal/ui/
```

Expected: success.

- [ ] **Step 3: Commit**

```bash
git add internal/ui/schedule.go
git commit -m "feat: add schedule selection TUI screen"
```

---

### Task 8: Editor Screens (Tea & Schedule)

**Files:**
- Create: `internal/ui/editor.go`

- [ ] **Step 1: Implement editor models**

Create `internal/ui/editor.go`:

```go
package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// --- Tea Editor ---

type TeaEditorResult struct {
	Name   string
	Saved  bool
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

	return TeaEditorModel{nameInput: ti, title: title}
}

func (m TeaEditorModel) Result() *TeaEditorResult { return m.result }

func (m TeaEditorModel) Init() tea.Cmd { return textinput.Blink }

func (m TeaEditorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
			m.result = &TeaEditorResult{Saved: false}
			return m, tea.Quit

		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			val := strings.TrimSpace(m.nameInput.Value())
			if val != "" {
				m.result = &TeaEditorResult{Name: val, Saved: true}
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

// --- Schedule Editor ---

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
	ni := textinput.New()
	ni.Placeholder = "Название расписания"
	ni.SetValue(name)
	ni.Focus()
	ni.CharLimit = 50

	pi := textinput.New()
	pi.Placeholder = "10,15,20,25,30"
	if len(pours) > 0 {
		parts := make([]string, len(pours))
		for i, p := range pours {
			parts[i] = strconv.Itoa(p)
		}
		pi.SetValue(strings.Join(parts, ","))
	}
	pi.CharLimit = 200

	title := "Редактировать расписание"
	if isNew {
		title = "Новое расписание"
	}

	return ScheduleEditorModel{
		nameInput:  ni,
		poursInput: pi,
		title:      title,
	}
}

func (m ScheduleEditorModel) Result() *ScheduleEditorResult { return m.result }

func (m ScheduleEditorModel) Init() tea.Cmd { return textinput.Blink }

func (m ScheduleEditorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
			m.result = &ScheduleEditorResult{Saved: false}
			return m, tea.Quit

		case key.Matches(msg, key.NewBinding(key.WithKeys("tab", "shift+tab"))):
			m.focusIndex = (m.focusIndex + 1) % 2
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
			m.result = &ScheduleEditorResult{Name: name, Pours: pours, Saved: true}
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
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		v, err := strconv.Atoi(p)
		if err != nil {
			return nil, fmt.Errorf("'%s' — не число", p)
		}
		if v <= 0 {
			return nil, fmt.Errorf("пролив должен быть > 0 секунд")
		}
		pours = append(pours, v)
	}
	if len(pours) == 0 {
		return nil, fmt.Errorf("Введите хотя бы один пролив")
	}
	return pours, nil
}
```

- [ ] **Step 2: Verify compilation**

```bash
go build ./internal/ui/
```

Expected: success.

- [ ] **Step 3: Commit**

```bash
git add internal/ui/editor.go
git commit -m "feat: add tea and schedule editor TUI screens with input validation"
```

---

### Task 9: Notify Screen (TEA TIME)

**Files:**
- Create: `internal/ui/notify.go`

- [ ] **Step 1: Implement notify model**

Create `internal/ui/notify.go`:

```go
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

func (m NotifyModel) Confirmed() bool { return m.confirmed }

func (m NotifyModel) Init() tea.Cmd { return nil }

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
	s += HelpStyle.Render("  Нажмите Enter")

	return BorderStyle.Render(s)
}
```

- [ ] **Step 2: Verify compilation**

```bash
go build ./internal/ui/
```

Expected: success.

- [ ] **Step 3: Commit**

```bash
git add internal/ui/notify.go
git commit -m "feat: add TEA TIME and DONE notification screens with ASCII art"
```

---

### Task 10: CLI Entry Point (Cobra Subcommands)

**Files:**
- Create: `cmd/tmux-tea/main.go`

- [ ] **Step 1: Add cobra dependency**

```bash
go get github.com/spf13/cobra
```

- [ ] **Step 2: Implement main.go with all subcommands**

Create `cmd/tmux-tea/main.go`:

```go
package main

import (
	"fmt"
	"os"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/papayka/tmux-tea/internal/config"
	"github.com/papayka/tmux-tea/internal/timer"
	"github.com/papayka/tmux-tea/internal/tmux"
	"github.com/papayka/tmux-tea/internal/ui"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "tmux-tea",
		Short: "Tea ceremony timer for tmux",
	}

	rootCmd.AddCommand(menuCmd())
	rootCmd.AddCommand(startCmd())
	rootCmd.AddCommand(statusCmd())
	rootCmd.AddCommand(confirmCmd())
	rootCmd.AddCommand(stopCmd())
	rootCmd.AddCommand(notifyCmd())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func menuCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "menu",
		Short: "Open tea selection menu",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgPath := config.DefaultPath()
			cfg, err := config.LoadOrCreate(cfgPath)
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			for {
				// Tea selection
				menuModel := ui.NewMenuModel(cfg.Teas)
				p := tea.NewProgram(menuModel)
				finalModel, err := p.Run()
				if err != nil {
					return err
				}
				result := finalModel.(ui.MenuModel).Result()
				if result == nil || result.Action == "quit" {
					return nil
				}

				switch result.Action {
				case "select":
					// Schedule selection
					action, err := runScheduleFlow(cfg, result.TeaIndex, cfgPath)
					if err != nil {
						return err
					}
					if action == "started" {
						return nil
					}
					// action == "back" → continue loop

				case "add":
					editorModel := ui.NewTeaEditorModel("", true)
					p := tea.NewProgram(editorModel)
					final, err := p.Run()
					if err != nil {
						return err
					}
					res := final.(ui.TeaEditorModel).Result()
					if res != nil && res.Saved {
						newTea := config.Tea{
							ID:   config.Slugify(res.Name),
							Name: res.Name,
							Schedules: []config.Schedule{
								{ID: "default", Name: "Стандарт", Pours: []int{10, 15, 20, 25, 30}},
							},
						}
						cfg.Teas = append(cfg.Teas, newTea)
						if err := config.Save(cfg, cfgPath); err != nil {
							return err
						}
					}

				case "edit":
					editorModel := ui.NewTeaEditorModel(result.Tea.Name, false)
					p := tea.NewProgram(editorModel)
					final, err := p.Run()
					if err != nil {
						return err
					}
					res := final.(ui.TeaEditorModel).Result()
					if res != nil && res.Saved {
						cfg.Teas[result.TeaIndex].Name = res.Name
						cfg.Teas[result.TeaIndex].ID = config.Slugify(res.Name)
						if err := config.Save(cfg, cfgPath); err != nil {
							return err
						}
					}

				case "delete":
					cfg.Teas = append(cfg.Teas[:result.TeaIndex], cfg.Teas[result.TeaIndex+1:]...)
					if err := config.Save(cfg, cfgPath); err != nil {
						return err
					}
				}
			}
		},
	}
}

func runScheduleFlow(cfg *config.Config, teaIndex int, cfgPath string) (string, error) {
	t := &cfg.Teas[teaIndex]
	for {
		schedModel := ui.NewScheduleModel(t.Name, t.Schedules)
		p := tea.NewProgram(schedModel)
		final, err := p.Run()
		if err != nil {
			return "", err
		}
		result := final.(ui.ScheduleModel).Result()
		if result == nil || result.Action == "back" {
			return "back", nil
		}

		switch result.Action {
		case "select":
			// Check if timer is already running
			state, _ := timer.ReadState(timer.StatePath())
			if state != nil && timer.IsProcessAlive(state.PID) {
				// Stop existing timer
				stopTimer(state.PID)
			}

			// Launch timer in background
			binPath, _ := os.Executable()
			tmux.RunShellBg(fmt.Sprintf("%s start --tea %s --schedule %s",
				binPath, t.ID, result.Schedule.ID))
			return "started", nil

		case "add":
			editorModel := ui.NewScheduleEditorModel("", nil, true)
			p := tea.NewProgram(editorModel)
			final, err := p.Run()
			if err != nil {
				return "", err
			}
			res := final.(ui.ScheduleEditorModel).Result()
			if res != nil && res.Saved {
				newSch := config.Schedule{
					ID:    config.Slugify(res.Name),
					Name:  res.Name,
					Pours: res.Pours,
				}
				t.Schedules = append(t.Schedules, newSch)
				cfg.Teas[teaIndex] = *t
				if err := config.Save(cfg, cfgPath); err != nil {
					return "", err
				}
			}

		case "edit":
			sch := result.Schedule
			editorModel := ui.NewScheduleEditorModel(sch.Name, sch.Pours, false)
			p := tea.NewProgram(editorModel)
			final, err := p.Run()
			if err != nil {
				return "", err
			}
			res := final.(ui.ScheduleEditorModel).Result()
			if res != nil && res.Saved {
				t.Schedules[result.ScheduleIndex].Name = res.Name
				t.Schedules[result.ScheduleIndex].ID = config.Slugify(res.Name)
				t.Schedules[result.ScheduleIndex].Pours = res.Pours
				cfg.Teas[teaIndex] = *t
				if err := config.Save(cfg, cfgPath); err != nil {
					return "", err
				}
			}

		case "delete":
			t.Schedules = append(t.Schedules[:result.ScheduleIndex], t.Schedules[result.ScheduleIndex+1:]...)
			cfg.Teas[teaIndex] = *t
			if err := config.Save(cfg, cfgPath); err != nil {
				return "", err
			}
		}
	}
}

func startCmd() *cobra.Command {
	var teaID, scheduleID string

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start tea timer",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgPath := config.DefaultPath()
			cfg, err := config.LoadOrCreate(cfgPath)
			if err != nil {
				return err
			}

			t := cfg.FindTea(teaID)
			if t == nil {
				return fmt.Errorf("tea %q not found", teaID)
			}
			s := t.FindSchedule(scheduleID)
			if s == nil {
				return fmt.Errorf("schedule %q not found", scheduleID)
			}

			// Save original status-interval, set to 1
			origInterval, _ := tmux.GetOption("status-interval")
			tmux.SetOption("status-interval", "1")

			session := timer.NewSession(t.Name, s.Pours, timer.StatePath())

			for !session.IsFinished() {
				binPath, _ := os.Executable()

				err := session.RunCurrentPour(func() error {
					// Show notification popup + bell
					tmux.SendBell()
					popupCmd := fmt.Sprintf("%s notify --tea-name %q --pour %d --total %d",
						binPath, session.TeaName, session.CurrentPour, session.TotalPours)
					return tmux.DisplayPopup(50, 15, popupCmd)
				})
				if err != nil {
					break
				}

				session.AdvancePour()
			}

			// Show "done" notification
			if session.IsFinished() {
				binPath, _ := os.Executable()
				popupCmd := fmt.Sprintf("%s notify --tea-name %q --finished", binPath, session.TeaName)
				tmux.DisplayPopup(50, 15, popupCmd)
			}

			// Cleanup
			timer.ClearState(timer.StatePath())
			if origInterval != "" {
				tmux.SetOption("status-interval", origInterval)
			}

			return nil
		},
	}
	cmd.Flags().StringVar(&teaID, "tea", "", "Tea ID")
	cmd.Flags().StringVar(&scheduleID, "schedule", "", "Schedule ID")
	cmd.MarkFlagRequired("tea")
	cmd.MarkFlagRequired("schedule")
	return cmd
}

func statusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Output status for tmux status bar",
		RunE: func(cmd *cobra.Command, args []string) error {
			state, err := timer.ReadState(timer.StatePath())
			if err != nil {
				return nil // silent fail — don't pollute status bar
			}

			if state == nil {
				return nil // no active session
			}

			// Check if process is still alive
			if !timer.IsProcessAlive(state.PID) {
				timer.ClearState(timer.StatePath())
				return nil
			}

			info := &tmux.StatusInfo{
				TeaName:      state.TeaName,
				PourIndex:    state.PourIndex,
				TotalPours:   state.TotalPours,
				RemainingSec: state.RemainingSec,
				Status:       state.Status,
			}
			fmt.Print(tmux.FormatStatus(info))
			return nil
		},
	}
}

func confirmCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "confirm",
		Short: "Confirm pour is done",
		RunE: func(cmd *cobra.Command, args []string) error {
			state, err := timer.ReadState(timer.StatePath())
			if err != nil || state == nil {
				return nil
			}
			if state.Status == timer.StatusReady {
				// Signal the timer process by updating state
				state.Status = "confirmed"
				return timer.WriteState(state, timer.StatePath())
			}
			return nil
		},
	}
}

func stopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop current tea session",
		RunE: func(cmd *cobra.Command, args []string) error {
			state, _ := timer.ReadState(timer.StatePath())
			if state != nil {
				stopTimer(state.PID)
			}
			timer.ClearState(timer.StatePath())
			return nil
		},
	}
}

func notifyCmd() *cobra.Command {
	var teaName string
	var pour, total int
	var finished bool

	cmd := &cobra.Command{
		Use:    "notify",
		Short:  "Show notification popup",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			model := ui.NewNotifyModel(teaName, pour, total, finished)
			p := tea.NewProgram(model)
			_, err := p.Run()
			return err
		},
	}
	cmd.Flags().StringVar(&teaName, "tea-name", "", "Tea name")
	cmd.Flags().IntVar(&pour, "pour", 0, "Current pour index")
	cmd.Flags().IntVar(&total, "total", 0, "Total pours")
	cmd.Flags().BoolVar(&finished, "finished", false, "Show finished screen")
	return cmd
}

func stopTimer(pid int) {
	if pid > 0 {
		proc, err := os.FindProcess(pid)
		if err == nil {
			proc.Signal(os.Interrupt)
		}
	}
}

// Ensure binary is available for self-invocation
func init() {
	// Suppress cobra completion command
	cobra.EnableCommandSorting = false
}

// Suppress unused import
var _ = exec.Command
```

- [ ] **Step 3: Remove unused exec import and verify build**

```bash
cd /home/papayka/my_plug/tmux-tea
go build -o bin/tmux-tea ./cmd/tmux-tea/
```

Expected: binary created at `bin/tmux-tea`.

- [ ] **Step 4: Test basic commands**

```bash
./bin/tmux-tea --help
./bin/tmux-tea status
```

Expected: help text prints, status outputs nothing (no active session).

- [ ] **Step 5: Commit**

```bash
git add cmd/ bin/
echo "bin/tmux-tea" >> .gitignore
git add .gitignore
git commit -m "feat: add CLI entry point with menu, start, status, confirm, stop, notify subcommands"
```

---

### Task 11: Tmux Plugin Entry Point

**Files:**
- Create: `tea.tmux`
- Create: `scripts/install.sh`

- [ ] **Step 1: Create tea.tmux**

Create `tea.tmux`:

```bash
#!/usr/bin/env bash
CURRENT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEA_BIN="$CURRENT_DIR/bin/tmux-tea"

# Build if binary doesn't exist
if [ ! -f "$TEA_BIN" ]; then
    "$CURRENT_DIR/scripts/install.sh"
fi

# Hotkey 1: main menu (prefix + t)
tmux bind-key t display-popup -E -w 40 -h 15 "$TEA_BIN menu"

# Hotkey 2: confirm pour (prefix + T)
tmux bind-key T run-shell "$TEA_BIN confirm"

# Status bar integration
tmux set-option -ga status-right '#('"$TEA_BIN"' status)'
```

- [ ] **Step 2: Create install.sh**

Create `scripts/install.sh`:

```bash
#!/usr/bin/env bash
set -euo pipefail

CURRENT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$CURRENT_DIR"

echo "Building tmux-tea..."
mkdir -p bin
go build -o bin/tmux-tea ./cmd/tmux-tea/
echo "Done: bin/tmux-tea"
```

- [ ] **Step 3: Make scripts executable**

```bash
chmod +x tea.tmux scripts/install.sh
```

- [ ] **Step 4: Test build via install script**

```bash
./scripts/install.sh
```

Expected: "Done: bin/tmux-tea" printed.

- [ ] **Step 5: Commit**

```bash
git add tea.tmux scripts/
git commit -m "feat: add tmux plugin entry point and build script"
```

---

### Task 12: Integration Testing & Polish

**Files:**
- Modify: `cmd/tmux-tea/main.go` (if needed)

- [ ] **Step 1: Run all unit tests**

```bash
cd /home/papayka/my_plug/tmux-tea
go test ./... -v
```

Expected: all tests PASS.

- [ ] **Step 2: Run vet and basic checks**

```bash
go vet ./...
```

Expected: no issues.

- [ ] **Step 3: Manual integration test in tmux**

```bash
# In tmux, source the plugin:
tmux source-file tea.tmux

# Test hotkey: prefix + t → menu should open
# Select a tea, select a schedule → timer should start
# Check status bar for countdown
# Wait for notification → TEA TIME popup should appear
# prefix + T → confirm, next pour starts
```

- [ ] **Step 4: Final commit**

```bash
git add -A
git commit -m "chore: integration testing and final polish"
```
