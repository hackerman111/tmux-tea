#!/usr/bin/env bash
set -euo pipefail

CURRENT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
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

exec "$TEA_BIN" "$@"
