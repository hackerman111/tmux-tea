package timer

import (
	"context"
	"fmt"
	"os"
	"time"
)

var waitForDuration = func(ctx context.Context, duration time.Duration) error {
	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

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

func (s *Session) RunCurrentPour(ctx context.Context, onReady func() error) error {
	duration := s.CurrentPourDuration()
	if duration <= 0 {
		return fmt.Errorf("invalid pour duration: %d", duration)
	}

	for remaining := duration; remaining > 0; remaining-- {
		if err := ctx.Err(); err != nil {
			return err
		}

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
		if err := waitForDuration(ctx, 1*time.Second); err != nil {
			return err
		}
	}

	if err := ctx.Err(); err != nil {
		return err
	}

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

	if onReady == nil {
		return nil
	}
	if err := onReady(); err != nil {
		return err
	}

	// Re-write ready state after the popup closes to clear any confirmation
	// that may have been triggered via the hotkey while the popup was visible.
	state.Status = StatusReady
	return WriteState(state, s.StatePath)
}

func (s *Session) WaitForConfirmation(ctx context.Context) (bool, error) {
	for {
		if err := ctx.Err(); err != nil {
			return false, err
		}

		state, err := ReadState(s.StatePath)
		if err != nil {
			return false, fmt.Errorf("reading state: %w", err)
		}
		if state == nil {
			return false, nil
		}

		switch state.Status {
		case StatusConfirmed:
			return true, nil
		case StatusReady:
		default:
			return false, fmt.Errorf("unexpected state while waiting for confirmation: %s", state.Status)
		}

		if err := waitForDuration(ctx, 200*time.Millisecond); err != nil {
			return false, err
		}
	}
}
