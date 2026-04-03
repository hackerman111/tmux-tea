package tmux

import "testing"

func TestBuildArgs(t *testing.T) {
	args := buildArgs("display-popup", "-E", "-w", "40", "-h", "15", "tmux-tea menu")
	if len(args) != 7 {
		t.Fatalf("expected 7 args, got %d: %v", len(args), args)
	}
	if args[0] != "display-popup" {
		t.Errorf("args[0] = %q, want 'display-popup'", args[0])
	}
}

func TestFormatStatusOutputEmpty(t *testing.T) {
	out := FormatStatus(nil)
	if out != "" {
		t.Errorf("FormatStatus(nil) = %q, want empty", out)
	}
}

func TestFormatStatusOutputCounting(t *testing.T) {
	out := FormatStatus(&StatusInfo{
		TeaName:      "Шен Пуэр",
		PourIndex:    3,
		TotalPours:   8,
		RemainingSec: 72,
		Status:       "counting",
	})
	if out == "" {
		t.Error("FormatStatus should not be empty for active timer")
	}
}

func TestFormatStatusOutputReady(t *testing.T) {
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
