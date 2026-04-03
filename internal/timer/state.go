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

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("parsing state: %w", err)
	}
	return &state, nil
}

func WriteState(state *State, path string) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling state: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("writing state: %w", err)
	}
	return nil
}

func ClearState(path string) {
	_ = os.Remove(path)
}

func IsProcessAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	return syscall.Kill(pid, 0) == nil
}
