#!/usr/bin/env bash
set -euo pipefail

CURRENT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TEA_SCRIPT="$CURRENT_DIR/scripts/tmux-tea"

command -v bash >/dev/null 2>&1 || {
	echo "tmux-tea: bash is required" >&2
	exit 1
}

command -v tmux >/dev/null 2>&1 || {
	echo "tmux-tea: tmux is required" >&2
	exit 1
}

chmod +x "$TEA_SCRIPT"
echo "tmux-tea is ready: source $CURRENT_DIR/tea.tmux from tmux"
