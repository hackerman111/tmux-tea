package main

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/papayka/tmux-tea/internal/timer"
	"github.com/papayka/tmux-tea/internal/tmux"
)

func TestNewRootCmdContainsExpectedSubcommands(t *testing.T) {
	cmd := newRootCmd()

	var names []string
	for _, sub := range cmd.Commands() {
		names = append(names, sub.Name())
	}

	for _, want := range []string{"menu", "start", "status", "confirm", "stop", "notify"} {
		if !slices.Contains(names, want) {
			t.Fatalf("subcommands = %v, want %q", names, want)
		}
	}
}

func TestStartCmdRequiresFlags(t *testing.T) {
	cmd := startCmd()
	cmd.SetArgs(nil)

	if err := cmd.Execute(); err == nil {
		t.Fatal("start command should require flags")
	}
}

func TestStatusCmdNoStateOutputsNothing(t *testing.T) {
	restore := stubStatePath(t, filepath.Join(t.TempDir(), "state.json"))
	defer restore()

	cmd := statusCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("status command returned error: %v", err)
	}
	if out.String() != "" {
		t.Fatalf("output = %q, want empty", out.String())
	}
}

func TestStatusCmdOutputsFormattedStatusForRunningProcess(t *testing.T) {
	statePath := filepath.Join(t.TempDir(), "state.json")
	restore := stubStatePath(t, statePath)
	defer restore()

	state := &timer.State{
		PID:          os.Getpid(),
		TeaName:      "Шен Пуэр",
		PourIndex:    1,
		TotalPours:   5,
		RemainingSec: 42,
		Status:       timer.StatusCounting,
	}
	if err := timer.WriteState(state, statePath); err != nil {
		t.Fatalf("WriteState failed: %v", err)
	}

	cmd := statusCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("status command returned error: %v", err)
	}

	want := tmux.FormatStatus(&tmux.StatusInfo{
		TeaName:      state.TeaName,
		PourIndex:    state.PourIndex,
		TotalPours:   state.TotalPours,
		RemainingSec: state.RemainingSec,
		Status:       state.Status,
	})
	if out.String() != want {
		t.Fatalf("output = %q, want %q", out.String(), want)
	}
}

func TestConfirmCmdMarksReadyStateConfirmed(t *testing.T) {
	statePath := filepath.Join(t.TempDir(), "state.json")
	restore := stubStatePath(t, statePath)
	defer restore()

	state := &timer.State{
		PID:          os.Getpid(),
		TeaName:      "Шен Пуэр",
		PourIndex:    0,
		TotalPours:   5,
		RemainingSec: 0,
		Status:       timer.StatusReady,
	}
	if err := timer.WriteState(state, statePath); err != nil {
		t.Fatalf("WriteState failed: %v", err)
	}

	cmd := confirmCmd()
	if err := cmd.Execute(); err != nil {
		t.Fatalf("confirm command returned error: %v", err)
	}

	updated, err := timer.ReadState(statePath)
	if err != nil {
		t.Fatalf("ReadState failed: %v", err)
	}
	if updated == nil {
		t.Fatal("state should still exist")
	}
	if updated.Status != timer.StatusConfirmed {
		t.Fatalf("status = %q, want %q", updated.Status, timer.StatusConfirmed)
	}
}

func TestConfirmCmdStopsActiveCountingSession(t *testing.T) {
	statePath := filepath.Join(t.TempDir(), "state.json")
	restore := stubStatePath(t, statePath)
	defer restore()

	state := &timer.State{
		PID:          0,
		TeaName:      "Шен Пуэр",
		PourIndex:    0,
		TotalPours:   5,
		RemainingSec: 12,
		Status:       timer.StatusCounting,
	}
	if err := timer.WriteState(state, statePath); err != nil {
		t.Fatalf("WriteState failed: %v", err)
	}

	cmd := confirmCmd()
	if err := cmd.Execute(); err != nil {
		t.Fatalf("confirm command returned error: %v", err)
	}

	updated, err := timer.ReadState(statePath)
	if err != nil {
		t.Fatalf("ReadState failed: %v", err)
	}
	if updated != nil {
		t.Fatalf("state = %#v, want nil after stop", updated)
	}
}

func stubStatePath(t *testing.T, path string) func() {
	t.Helper()

	original := statePathFunc
	statePathFunc = func() string { return path }
	return func() {
		statePathFunc = original
	}
}
