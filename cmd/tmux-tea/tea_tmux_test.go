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
		"needs_build()",
		`-newer "$TEA_BIN"`,
		`"$CURRENT_DIR/scripts/install.sh"`,
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
		`"$TEA_BIN menu"`,
	} {
		if !strings.Contains(script, want) {
			t.Fatalf("tea.tmux should contain %q, got:\n%s", want, script)
		}
	}
}

func readTeaTmux(t *testing.T) string {
	t.Helper()

	path := filepath.Join("..", "..", "tea.tmux")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q): %v", path, err)
	}
	return string(data)
}
