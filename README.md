# tmux-tea

Lightweight tmux tea ceremony timer. One tmux config file and one Bash script.

## Install

With TPM:

```tmux
set -g @plugin 'hackerman111/tmux-tea'
```

Manual:

```tmux
source-file /home/papayka/my_plug/tmux-tea/tea.tmux
```


## Hotkeys

- `prefix + T`: open tea menu.
- `prefix + t`: confirm a ready pour, or show remaining time while counting.

The plugin also appends the active timer to `status-right`.

## Commands

```bash
scripts/tmux-tea menu
scripts/tmux-tea start <tea_id> <schedule_id>
scripts/tmux-tea confirm
scripts/tmux-tea stop
scripts/tmux-tea status
scripts/tmux-tea edit
scripts/tmux-tea add-tea "Tea Name" "Schedule Name" "10,15,20"
scripts/tmux-tea add-schedule <tea_id> "Schedule Name" "10,15,20"
scripts/tmux-tea delete-tea <tea_id>
scripts/tmux-tea delete-schedule <tea_id> <schedule_id>
```

## Data

Tea config:

```text
${XDG_CONFIG_HOME:-$HOME/.config}/tmux-tea/teas.tsv
```

Line format:

```text
tea_id<TAB>tea_name<TAB>schedule_id<TAB>schedule_name<TAB>pours_csv
```

Timer state:

```text
${TMPDIR:-/tmp}/tmux-tea-state.$UID
```

The timer is not a daemon. `start` launches one background Bash process for the
active session, and `stop` removes its state.

## Test

```bash
bash tests/smoke.sh
```
