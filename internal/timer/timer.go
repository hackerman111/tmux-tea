package timer

import (
	"fmt"
	"os"
	"time"
)

var sleep = time.Sleep

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
		sleep(1 * time.Second)
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
	return onReady()
}
