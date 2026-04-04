#!/usr/bin/env bash

CURRENT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
RUNNER="$CURRENT_DIR/scripts/run.sh"

tmux bind-key t display-popup -E -w 72 -h 28 "$RUNNER menu"
tmux bind-key T run-shell "$RUNNER confirm"
tmux set-option -ga status-right '#('"$RUNNER"' status)'
