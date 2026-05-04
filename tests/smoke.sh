#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

fail() {
	printf 'FAIL: %s\n' "$*" >&2
	exit 1
}

assert_contains() {
	local haystack="$1"
	local needle="$2"
	local label="$3"

	[[ "$haystack" == *"$needle"* ]] || fail "$label: expected to find '$needle' in '$haystack'"
}

run_tea() {
	HOME="$TEST_HOME" \
	XDG_CONFIG_HOME="$TEST_HOME/.config" \
	TMPDIR="$TEST_TMP" \
	TMUX_TEA_NO_POPUP=1 \
		"$ROOT/scripts/tmux-tea" "$@"
}

run_tea_input() {
	local input="$1"
	shift
	printf '%s' "$input" | HOME="$TEST_HOME" \
		XDG_CONFIG_HOME="$TEST_HOME/.config" \
		TMPDIR="$TEST_TMP" \
		TMUX_TEA_NO_POPUP=1 \
			"$ROOT/scripts/tmux-tea" "$@"
}

wait_for_file() {
	local file="$1"
	local i

	for i in {1..30}; do
		[[ -s "$file" ]] && return 0
		sleep 0.1
	done

	fail "timed out waiting for $file"
}

tmux_isolated() {
	TMUX= TMUX_PANE= tmux -L "$TMUX_SOCKET" "$@"
}

TEST_TMP="$(mktemp -d "${TMPDIR:-/tmp}/tmux-tea-smoke.XXXXXX")"
TEST_HOME="$TEST_TMP/home"
STATE_FILE="$TEST_TMP/tmux-tea-state.${UID:-$(id -u)}"
TMUX_SOCKET="tmux-tea-smoke-${RANDOM}-${RANDOM}"

cleanup() {
	TMPDIR="$TEST_TMP" "$ROOT/scripts/tmux-tea" stop >/dev/null 2>&1 || true
	tmux_isolated kill-server >/dev/null 2>&1 || true
	rm -rf "$TEST_TMP"
}
trap cleanup EXIT

mkdir -p "$TEST_HOME"

bash -n "$ROOT/scripts/tmux-tea" "$ROOT/scripts/install.sh" "$ROOT/tests/smoke.sh"

initial_status="$(run_tea status)"
[[ -z "$initial_status" ]] || fail "status should be empty without active timer"

run_tea start shen-puer default >/dev/null
wait_for_file "$STATE_FILE"

[[ -s "$TEST_HOME/.config/tmux-tea/teas.tsv" ]] || fail "default config was not created"
assert_contains "$(cat "$TEST_HOME/.config/tmux-tea/teas.tsv")" $'shen-puer\tShen Puer\tdefault' "default config"

run_tea add-tea "Test Tea" "Fast" "1,2" >/dev/null
assert_contains "$(cat "$TEST_HOME/.config/tmux-tea/teas.tsv")" $'test-tea\tTest Tea\tfast\tFast\t1,2' "add tea"

run_tea add-schedule test-tea "Slow" "3 4" >/dev/null
assert_contains "$(cat "$TEST_HOME/.config/tmux-tea/teas.tsv")" $'test-tea\tTest Tea\tslow\tSlow\t3,4' "add schedule"

run_tea_input $'DELETE\n' delete-schedule test-tea slow >/dev/null
if [[ "$(cat "$TEST_HOME/.config/tmux-tea/teas.tsv")" == *$'test-tea\tTest Tea\tslow\tSlow\t3,4'* ]]; then
	fail "delete schedule should remove selected row"
fi

run_tea_input $'DELETE\n' delete-tea test-tea >/dev/null
if [[ "$(cat "$TEST_HOME/.config/tmux-tea/teas.tsv")" == *$'test-tea\tTest Tea'* ]]; then
	fail "delete tea should remove all selected tea rows"
fi

active_status="$(run_tea status)"
assert_contains "$active_status" "tea Shen Puer" "active status"

run_tea stop >/dev/null
[[ ! -e "$STATE_FILE" ]] || fail "stop should remove state file"

tmux_isolated -f /dev/null new-session -d
tmux_isolated source-file "$ROOT/tea.tmux"

plugin_dir="$(tmux_isolated show-option -gqv @tmux_tea_dir)"
[[ "$plugin_dir" == "$ROOT" ]] || fail "tmux plugin dir mismatch: '$plugin_dir'"

confirm_binding="$(tmux_isolated list-keys -T prefix t)"
menu_binding="$(tmux_isolated list-keys -T prefix T)"
status_right="$(tmux_isolated show-option -gqv status-right)"

assert_contains "$confirm_binding" "tmux-tea confirm" "prefix+t binding"
assert_contains "$menu_binding" "display-popup" "prefix+T binding"
assert_contains "$menu_binding" "tmux-tea menu" "prefix+T command"
assert_contains "$status_right" "tmux-tea status" "status-right"

printf 'smoke tests passed\n'
