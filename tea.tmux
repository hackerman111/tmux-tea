#!/usr/bin/env bash

CURRENT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEA_BIN="$CURRENT_DIR/bin/tmux-tea"

needs_build() {
	if [ ! -x "$TEA_BIN" ]; then
		return 0
	fi

	find "$CURRENT_DIR" \
		-type f \
		\( -name '*.go' -o -name '*.sh' -o -name 'go.mod' -o -name 'go.sum' -o -name 'tea.tmux' \) \
		-newer "$TEA_BIN" | grep -q .
}

if needs_build; then
	"$CURRENT_DIR/scripts/install.sh"
fi

tmux bind-key t display-popup -E -w 72 -h 28 "$TEA_BIN menu"
tmux bind-key T run-shell "$TEA_BIN confirm"
tmux set-option -ga status-right '#('"$TEA_BIN"' status)'
