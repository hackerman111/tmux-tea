package timer

import (
	"path/filepath"
	"testing"
	"time"
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
	if s.StatePath != statePath {
		t.Errorf("StatePath = %q, want %q", s.StatePath, statePath)
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

func TestSession_RunCurrentPourWritesReadyStateAndCallsOnReady(t *testing.T) {
	dir := t.TempDir()
	statePath := filepath.Join(dir, "state.json")

	oldSleep := sleep
	sleep = func(time.Duration) {}
	defer func() {
		sleep = oldSleep
	}()

	s := NewSession("Шен Пуэр", []int{1}, statePath)
	readyCalls := 0

	if err := s.RunCurrentPour(func() error {
		readyCalls++
		return nil
	}); err != nil {
		t.Fatalf("RunCurrentPour: %v", err)
	}

	if readyCalls != 1 {
		t.Fatalf("onReady calls = %d, want 1", readyCalls)
	}

	state, err := ReadState(statePath)
	if err != nil {
		t.Fatalf("ReadState: %v", err)
	}
	if state == nil {
		t.Fatal("state should not be nil")
	}
	if state.Status != StatusReady {
		t.Errorf("Status = %q, want %q", state.Status, StatusReady)
	}
	if state.RemainingSec != 0 {
		t.Errorf("RemainingSec = %d, want 0", state.RemainingSec)
	}
	if state.PourIndex != 0 {
		t.Errorf("PourIndex = %d, want 0", state.PourIndex)
	}
	if state.TotalPours != 1 {
		t.Errorf("TotalPours = %d, want 1", state.TotalPours)
	}
}

func TestSession_RunCurrentPourRejectsInvalidDuration(t *testing.T) {
	dir := t.TempDir()
	statePath := filepath.Join(dir, "state.json")

	s := NewSession("test", []int{0}, statePath)
	if err := s.RunCurrentPour(func() error {
		t.Fatal("onReady should not be called")
		return nil
	}); err == nil {
		t.Fatal("RunCurrentPour should fail for invalid duration")
	}
}
