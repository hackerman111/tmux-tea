# tmux-tea Shell Rewrite Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the Go/Bubbletea implementation with a lightweight Bash-only tmux plugin that loads cleanly via `tmux source-file tea.tmux`, provides menu, timer, status, confirm, stop, and edit flows, and keeps no build step.

**Architecture:** `tea.tmux` is a pure tmux config file. All runtime logic lives in one executable Bash script, `scripts/tmux-tea`, using tmux-native menus/popups and plain text state/config files. Active timers run as one short-lived background process only while tea is brewing.

**Tech Stack:** Bash, tmux CLI, POSIX-ish text files, no Go, no external TUI dependencies.

---

## File Structure

- Create `scripts/tmux-tea`: command dispatcher, config storage, menu/edit UI, timer state, notifications.
- Create `tests/smoke.sh`: fast smoke tests for parsing, config defaults, status output, tmux sourcing, and non-interactive timer behavior.
- Modify `tea.tmux`: valid tmux config, hotkeys, status integration.
- Modify `scripts/install.sh`: no build; make `scripts/tmux-tea` executable.
- Remove `scripts/run.sh`: Go build wrapper is obsolete.
- Remove Go project files: `go.mod`, `go.sum`, `cmd/`, `internal/`.
- Add `README.md`: usage, hotkeys, files, commands.

## Task 1: Shell Smoke Tests

**Files:**
- Create: `tests/smoke.sh`

- [ ] **Step 1: Write tests first**

Test coverage:
- `bash -n scripts/tmux-tea scripts/install.sh tests/smoke.sh`
- temp `HOME` seeds `~/.config/tmux-tea/teas.tsv`
- `scripts/tmux-tea status` is empty with no active state
- non-interactive `start` creates state and `stop` clears it
- isolated `tmux -L tmux-tea-smoke source-file tea.tmux` succeeds
- tmux bindings exist for `prefix + t` confirm and `prefix + T` menu

- [ ] **Step 2: Run the tests and confirm RED**

Run: `bash tests/smoke.sh`

Expected: FAIL because `scripts/tmux-tea` does not exist and `tea.tmux` is not a valid tmux config.

## Task 2: Bash Runtime

**Files:**
- Create: `scripts/tmux-tea`

- [ ] **Step 1: Implement command dispatcher**

Required commands:
- `menu`
- `schedule <tea_id>`
- `start <tea_id> <schedule_id>`
- `timer <tea_id> <schedule_id>`
- `confirm`
- `stop`
- `status`
- `notify <title> <line1> <line2>`
- `edit`
- `add-tea`
- `add-schedule <tea_id>`
- `delete-tea <tea_id>`
- `delete-schedule <tea_id> <schedule_id>`

- [ ] **Step 2: Store config as TSV**

Config path: `${XDG_CONFIG_HOME:-$HOME/.config}/tmux-tea/teas.tsv`

Line format:

```text
tea_id<TAB>tea_name<TAB>schedule_id<TAB>schedule_name<TAB>pours_csv
```

Default rows:

```text
shen-puer<TAB>Shen Puer<TAB>default<TAB>Classic<TAB>10,15,20,25,30,40
shu-puer<TAB>Shu Puer<TAB>default<TAB>Dense<TAB>8,12,16,20,25,30
da-hong-pao<TAB>Da Hong Pao<TAB>default<TAB>Rock<TAB>7,10,13,16,20
```

- [ ] **Step 3: Store timer state as shell assignments**

State path: `${TMPDIR:-/tmp}/tmux-tea-state.${UID:-$(id -u)}`

Fields:
- `pid`
- `tea_id`
- `tea_name`
- `schedule_id`
- `schedule_name`
- `pour_index`
- `total_pours`
- `remaining_sec`
- `status`

- [ ] **Step 4: Implement tmux-native UI**

Use:
- `tmux display-menu` for tea and schedule selection.
- `tmux display-popup -E` for the edit prompts and notifications.
- `tmux command-prompt` only for simple confirmation prompts where popup input is not needed.

- [ ] **Step 5: Run smoke tests and iterate to GREEN**

Run: `bash tests/smoke.sh`

Expected: PASS.

## Task 3: Tmux Entry Point

**Files:**
- Modify: `tea.tmux`
- Modify: `scripts/install.sh`
- Delete: `scripts/run.sh`

- [ ] **Step 1: Make `tea.tmux` pure tmux config**

Use tmux formats to resolve the plugin directory:

```tmux
set -g @tmux_tea_dir "#{d:#{current_file}}"
```

Bind:
- `prefix + t`: `run-shell -b "#{@tmux_tea_dir}/scripts/tmux-tea confirm"`
- `prefix + T`: `display-popup -E -w 72 -h 28 "#{@tmux_tea_dir}/scripts/tmux-tea menu"`

- [ ] **Step 2: Keep install no-build**

`scripts/install.sh` only checks Bash, tmux, and executable bit for `scripts/tmux-tea`.

- [ ] **Step 3: Re-run smoke tests**

Run: `bash tests/smoke.sh`

Expected: PASS.

## Task 4: Remove Go and Document Usage

**Files:**
- Delete: `go.mod`
- Delete: `go.sum`
- Delete: `cmd/tmux-tea/*`
- Delete: `internal/**/*`
- Create: `README.md`

- [ ] **Step 1: Remove Go files**

The repository should have no tracked Go source or Go module metadata.

- [ ] **Step 2: Document usage**

Document:
- TPM/source-file installation
- hotkeys
- config and state paths
- command list
- no build/no Go requirement

- [ ] **Step 3: Final verification**

Run:

```bash
bash tests/smoke.sh
rg --files | rg '(^go\.mod$|^go\.sum$|\.go$)' || true
tmux -L tmux-tea-check -f /dev/null new-session -d
tmux -L tmux-tea-check source-file tea.tmux
tmux -L tmux-tea-check kill-server
```

Expected:
- smoke tests pass
- Go file search prints nothing
- tmux source-file exits successfully
