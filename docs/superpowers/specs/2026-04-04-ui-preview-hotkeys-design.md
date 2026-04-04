# tmux-tea: UI Preview, Notification Centering, and Hotkey Swap

## Overview

This change refines the existing tmux-tea UI without changing the core flow:

- the tea selection screen keeps its current list-based navigation
- pressing `Enter` on a tea still opens the schedule selection screen
- the empty lower area of the tea menu is replaced with a live preview of schedules and pour timings for the currently selected tea
- `TEA TIME` and `DONE` notifications remain in their own popup, but their content is centered within the available popup area
- tmux hotkeys are swapped:
  - `Prefix+t` confirms the current pour and advances to the next one
  - `Prefix+T` opens the main menu

## Scope

This design changes only presentation and tmux key bindings. It does not change:

- how teas and schedules are stored
- how schedules are chosen after selecting a tea
- timer state transitions
- the CLI command set

## Tea Menu Preview

### Goal

Use the empty lower portion of the tea list popup to show immediately useful information about the currently selected tea.

### Behavior

The menu keeps the tea list in the upper section. A preview section is rendered below it for the currently highlighted tea.

The preview section contains:

- a short section title such as `Варианты заваривания`
- up to a few schedule rows for the selected tea
- each row shows the schedule name and a compact preview of pour durations
- if there are more schedules than fit comfortably, the preview shows a final summary row such as `...ещё N`

Example structure:

```text
Выберите чай:

> Шен Пуэр
  Шу Пуэр
  Да Хун Пао

Варианты заваривания:
  Стандарт: 10, 15, 20, 25, 30...
  Быстрый: 5, 10, 10, 15, 15...
```

### Constraints

- the screen remains informational only
- `Enter` on the tea screen still opens the schedule screen
- the preview must update immediately when the tea cursor moves
- long names and long timing lists must be truncated to the panel width

## Notification Centering

### Goal

Make `TEA TIME` and `DONE` feel visually centered within the popup instead of sitting too high in the frame.

### Behavior

The notification popup keeps its own size. The rendered notification content is composed as a single block:

- ASCII title
- status subtitle
- help text

That block is then centered horizontally and vertically within the popup content area.

This applies to both:

- active-pour notification (`TEA TIME`)
- finished notification (`DONE`)

### Constraints

- the current meaning of keys inside the popup does not change
- help text may wrap, but must remain centered as a block
- the visible border and popup size remain separate from the main menu sizing

## Hotkey Swap

### Goal

Match the requested ergonomics by moving the timer action to the easier shortcut and the menu to the shifted shortcut.

### Behavior

In the tmux plugin entrypoint:

- `Prefix+t` calls `confirm`
- `Prefix+T` opens `menu`

No command names or timer logic change. Only the key bindings are swapped.

## Implementation Notes

### Files

- `internal/ui/menu.go`
  - add rendering of the preview section under the tea list
- `internal/ui/styles.go`
  - add or adjust helpers for preview layout if needed
- `internal/ui/notify.go`
  - center the notification block within the popup content area
- `tea.tmux`
  - swap `t` and `T` bindings
- `internal/ui/menu_test.go`
  - cover preview rendering and truncation
- `internal/ui/notify_test.go`
  - cover centered content rendering expectations
- `cmd/tmux-tea/tea_tmux_test.go`
  - cover swapped bindings

## Testing

Verification should cover:

- menu preview updates when moving between teas
- preview includes schedule names and pour timing summaries
- long preview rows truncate cleanly without breaking popup width
- `TEA TIME` and `DONE` views render centered content
- tmux bindings map `t` to `confirm` and `T` to `menu`

## Out of Scope

- one-screen tea and schedule selection
- changing popup sizes for notifications
- changing timer semantics
- changing how schedules are started from the tea screen
