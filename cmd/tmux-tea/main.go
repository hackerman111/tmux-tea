package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/papayka/tmux-tea/internal/config"
	"github.com/papayka/tmux-tea/internal/timer"
	"github.com/papayka/tmux-tea/internal/tmux"
	"github.com/papayka/tmux-tea/internal/ui"
)

const statusConfirmed = "confirmed"

var (
	configPathFunc = config.DefaultPath
	statePathFunc  = timer.StatePath
)

func main() {
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "tmux-tea",
		Short: "Tea ceremony timer for tmux",
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}

	rootCmd.AddCommand(menuCmd())
	rootCmd.AddCommand(startCmd())
	rootCmd.AddCommand(statusCmd())
	rootCmd.AddCommand(confirmCmd())
	rootCmd.AddCommand(stopCmd())
	rootCmd.AddCommand(notifyCmd())

	return rootCmd
}

func menuCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "menu",
		Short: "Open tea selection menu",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgPath := configPathFunc()
			cfg, err := config.LoadOrCreate(cfgPath)
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			for {
				menuModel := ui.NewMenuModel(cfg.Teas)
				finalModel, err := tea.NewProgram(menuModel).Run()
				if err != nil {
					return err
				}

				result := finalModel.(ui.MenuModel).Result()
				if result == nil || result.Action == "quit" {
					return nil
				}

				switch result.Action {
				case "select":
					action, err := runScheduleFlow(cfg, result.TeaIndex, cfgPath)
					if err != nil {
						return err
					}
					if action == "started" {
						return nil
					}

				case "add":
					finalModel, err := tea.NewProgram(ui.NewTeaEditorModel("", true)).Run()
					if err != nil {
						return err
					}
					editorResult := finalModel.(ui.TeaEditorModel).Result()
					if editorResult != nil && editorResult.Saved {
						cfg.Teas = append(cfg.Teas, config.Tea{
							ID:   config.Slugify(editorResult.Name),
							Name: editorResult.Name,
							Schedules: []config.Schedule{
								{ID: "default", Name: "Стандарт", Pours: []int{10, 15, 20, 25, 30}},
							},
						})
						if err := config.Save(cfg, cfgPath); err != nil {
							return err
						}
					}

				case "edit":
					finalModel, err := tea.NewProgram(ui.NewTeaEditorModel(result.Tea.Name, false)).Run()
					if err != nil {
						return err
					}
					editorResult := finalModel.(ui.TeaEditorModel).Result()
					if editorResult != nil && editorResult.Saved {
						cfg.Teas[result.TeaIndex].Name = editorResult.Name
						cfg.Teas[result.TeaIndex].ID = config.Slugify(editorResult.Name)
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
	selectedTea := &cfg.Teas[teaIndex]

	for {
		finalModel, err := tea.NewProgram(ui.NewScheduleModel(selectedTea.Name, selectedTea.Schedules)).Run()
		if err != nil {
			return "", err
		}

		result := finalModel.(ui.ScheduleModel).Result()
		if result == nil || result.Action == "back" {
			return "back", nil
		}

		switch result.Action {
		case "select":
			state, _ := timer.ReadState(statePathFunc())
			if state != nil && timer.IsProcessAlive(state.PID) {
				stopTimer(state.PID)
			}

			binPath, err := os.Executable()
			if err != nil {
				return "", err
			}

			command := fmt.Sprintf("%q start --tea %q --schedule %q", binPath, selectedTea.ID, result.Schedule.ID)
			if err := tmux.RunShellBg(command); err != nil {
				return "", err
			}
			return "started", nil

		case "add":
			finalModel, err := tea.NewProgram(ui.NewScheduleEditorModel("", nil, true)).Run()
			if err != nil {
				return "", err
			}
			editorResult := finalModel.(ui.ScheduleEditorModel).Result()
			if editorResult != nil && editorResult.Saved {
				selectedTea.Schedules = append(selectedTea.Schedules, config.Schedule{
					ID:    config.Slugify(editorResult.Name),
					Name:  editorResult.Name,
					Pours: editorResult.Pours,
				})
				cfg.Teas[teaIndex] = *selectedTea
				if err := config.Save(cfg, cfgPath); err != nil {
					return "", err
				}
			}

		case "edit":
			schedule := result.Schedule
			finalModel, err := tea.NewProgram(ui.NewScheduleEditorModel(schedule.Name, schedule.Pours, false)).Run()
			if err != nil {
				return "", err
			}
			editorResult := finalModel.(ui.ScheduleEditorModel).Result()
			if editorResult != nil && editorResult.Saved {
				selectedTea.Schedules[result.ScheduleIndex].Name = editorResult.Name
				selectedTea.Schedules[result.ScheduleIndex].ID = config.Slugify(editorResult.Name)
				selectedTea.Schedules[result.ScheduleIndex].Pours = editorResult.Pours
				cfg.Teas[teaIndex] = *selectedTea
				if err := config.Save(cfg, cfgPath); err != nil {
					return "", err
				}
			}

		case "delete":
			selectedTea.Schedules = append(
				selectedTea.Schedules[:result.ScheduleIndex],
				selectedTea.Schedules[result.ScheduleIndex+1:]...,
			)
			cfg.Teas[teaIndex] = *selectedTea
			if err := config.Save(cfg, cfgPath); err != nil {
				return "", err
			}
		}
	}
}

func startCmd() *cobra.Command {
	var teaID string
	var scheduleID string

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start tea timer",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadOrCreate(configPathFunc())
			if err != nil {
				return err
			}

			selectedTea := cfg.FindTea(teaID)
			if selectedTea == nil {
				return fmt.Errorf("tea %q not found", teaID)
			}

			selectedSchedule := selectedTea.FindSchedule(scheduleID)
			if selectedSchedule == nil {
				return fmt.Errorf("schedule %q not found", scheduleID)
			}

			originalInterval, _ := tmux.GetOption("status-interval")
			_ = tmux.SetOption("status-interval", "1")

			session := timer.NewSession(selectedTea.Name, selectedSchedule.Pours, statePathFunc())

			for !session.IsFinished() {
				binPath, err := os.Executable()
				if err != nil {
					return err
				}

				err = session.RunCurrentPour(func() error {
					_ = tmux.SendBell()
					popupCmd := fmt.Sprintf(
						"%q notify --tea-name %q --pour %d --total %d",
						binPath,
						session.TeaName,
						session.CurrentPour,
						session.TotalPours,
					)
					return tmux.DisplayPopup(50, 15, popupCmd)
				})
				if err != nil {
					break
				}

				session.AdvancePour()
			}

			if session.IsFinished() {
				binPath, err := os.Executable()
				if err == nil {
					popupCmd := fmt.Sprintf("%q notify --tea-name %q --finished", binPath, session.TeaName)
					_ = tmux.DisplayPopup(50, 15, popupCmd)
				}
			}

			timer.ClearState(statePathFunc())
			if originalInterval != "" {
				_ = tmux.SetOption("status-interval", originalInterval)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&teaID, "tea", "", "Tea ID")
	cmd.Flags().StringVar(&scheduleID, "schedule", "", "Schedule ID")
	_ = cmd.MarkFlagRequired("tea")
	_ = cmd.MarkFlagRequired("schedule")
	return cmd
}

func statusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Output status for tmux status bar",
		RunE: func(cmd *cobra.Command, args []string) error {
			state, err := timer.ReadState(statePathFunc())
			if err != nil || state == nil {
				return nil
			}

			if !timer.IsProcessAlive(state.PID) {
				timer.ClearState(statePathFunc())
				return nil
			}

			info := &tmux.StatusInfo{
				TeaName:      state.TeaName,
				PourIndex:    state.PourIndex,
				TotalPours:   state.TotalPours,
				RemainingSec: state.RemainingSec,
				Status:       state.Status,
			}
			_, err = fmt.Fprint(cmd.OutOrStdout(), tmux.FormatStatus(info))
			return err
		},
	}
}

func confirmCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "confirm",
		Short: "Confirm pour is done",
		RunE: func(cmd *cobra.Command, args []string) error {
			state, err := timer.ReadState(statePathFunc())
			if err != nil || state == nil {
				return nil
			}

			if state.Status == timer.StatusReady {
				state.Status = statusConfirmed
				return timer.WriteState(state, statePathFunc())
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
			state, _ := timer.ReadState(statePathFunc())
			if state != nil {
				stopTimer(state.PID)
			}
			timer.ClearState(statePathFunc())
			return nil
		},
	}
}

func notifyCmd() *cobra.Command {
	var teaName string
	var pour int
	var total int
	var finished bool

	cmd := &cobra.Command{
		Use:    "notify",
		Short:  "Show notification popup",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := tea.NewProgram(ui.NewNotifyModel(teaName, pour, total, finished)).Run()
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
	if pid <= 0 {
		return
	}

	process, err := os.FindProcess(pid)
	if err == nil {
		_ = process.Signal(os.Interrupt)
	}
}

func init() {
	cobra.EnableCommandSorting = false
}
