package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTeaTmuxRebuildsWhenSourcesChange(t *testing.T) {
	script := readTeaTmux(t)

	for _, want := range []string{
		`RUNNER="$CURRENT_DIR/scripts/run.sh"`,
		`"$RUNNER menu"`,
		`"$RUNNER confirm"`,
		`'"$RUNNER"' status`,
	} {
		if !strings.Contains(script, want) {
			t.Fatalf("tea.tmux should contain %q, got:\n%s", want, script)
		}
	}
}

func TestTeaTmuxUsesLargerMenuPopup(t *testing.T) {
	script := readTeaTmux(t)

	for _, want := range []string{
		"bind-key t display-popup -E -w 72 -h 28",
		`"$RUNNER menu"`,
	} {
		if !strings.Contains(script, want) {
			t.Fatalf("tea.tmux should contain %q, got:\n%s", want, script)
		}
	}
}

func TestRunScriptRebuildsBeforeExecutingBinary(t *testing.T) {
	script := readScript(t, filepath.Join("..", "..", "scripts", "run.sh"))

	for _, want := range []string{
		"needs_build()",
		`-newer "$TEA_BIN"`,
		`"$CURRENT_DIR/scripts/install.sh"`,
		`exec "$TEA_BIN" "$@"`,
	} {
		if !strings.Contains(script, want) {
			t.Fatalf("scripts/run.sh should contain %q, got:\n%s", want, script)
		}
	}
}

func readTeaTmux(t *testing.T) string {
	t.Helper()

	return readScript(t, filepath.Join("..", "..", "tea.tmux"))
}

func readScript(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q): %v", path, err)
	}
	return string(data)
}
