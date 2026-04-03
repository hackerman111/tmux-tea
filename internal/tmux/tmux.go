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
	return Run(buildArgs(
		"display-popup", "-E",
		"-w", fmt.Sprintf("%d", width),
		"-h", fmt.Sprintf("%d", height),
		command,
	)...)
}

func RunShellBg(command string) error {
	return Run(buildArgs("run-shell", "-b", command)...)
}

func SendBell() error {
	return Run(buildArgs("run-shell", "printf '\\a'")...)
}

func SetOption(option, value string) error {
	return Run(buildArgs("set-option", "-g", option, value)...)
}

func GetOption(option string) (string, error) {
	return RunOutput(buildArgs("show-option", "-gv", option)...)
}

func FormatStatus(info *StatusInfo) string {
	if info == nil {
		return ""
	}

	switch info.Status {
	case "ready":
		return fmt.Sprintf("tea %s [%d/%d] READY", info.TeaName, info.PourIndex+1, info.TotalPours)
	case "counting":
		mins := info.RemainingSec / 60
		secs := info.RemainingSec % 60
		return fmt.Sprintf("tea %s [%d/%d] %d:%02d", info.TeaName, info.PourIndex+1, info.TotalPours, mins, secs)
	default:
		return ""
	}
}
