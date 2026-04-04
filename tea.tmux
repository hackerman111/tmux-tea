#!/usr/bin/env bash

CURRENT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEA_BIN="$CURRENT_DIR/bin/tmux-tea"

if [ ! -f "$TEA_BIN" ]; then
	"$CURRENT_DIR/scripts/install.sh"
fi

tmux bind-key t display-popup -E -w 56 -h 24 "$TEA_BIN menu"
tmux bind-key T run-shell "$TEA_BIN confirm"
tmux set-option -ga status-right '#('"$TEA_BIN"' status)'
