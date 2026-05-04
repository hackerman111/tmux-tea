# UI Preview, Notification Centering, and Hotkey Swap Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add an informational schedule preview to the tea menu, center notification content within its popup, and swap the tmux hotkeys for menu vs confirm actions.

**Architecture:** Keep the current two-step tea -> schedule flow. Extend the existing `MenuModel` renderer with a lower preview section, keep notification behavior unchanged while changing only layout composition in `NotifyModel.View`, and swap only the tmux key bindings in the shell wrapper plus matching assertions in tests.

**Tech Stack:** Go 1.21+, Bubble Tea, Lip Gloss, tmux shell bindings, Go test

---

## File Map

| File | Responsibility |
|------|---------------|
| `internal/ui/menu.go` | Tea menu rendering and preview helpers |
| `internal/ui/styles.go` | Shared panel widths and layout helpers for menu/notify content |
| `internal/ui/notify.go` | Notification block composition and centering |
| `tea.tmux` | tmux key bindings for menu and confirm actions |
| `internal/ui/menu_test.go` | Preview rendering, truncation, and menu help expectations |
| `internal/ui/notify_test.go` | Centered notify layout and help text expectations |
| `cmd/tmux-tea/tea_tmux_test.go` | Swapped key binding assertions |

### Task 1: Add Tea Menu Preview

**Files:**
- Modify: `internal/ui/menu.go`
- Test: `internal/ui/menu_test.go`

- [ ] **Step 1: Write failing tea preview tests**

Add tests in `internal/ui/menu_test.go` that assert the menu view:

```go
func TestMenuModelViewShowsSelectedTeaSchedulePreview(t *testing.T) {
	view := NewMenuModel([]config.Tea{
		{
			ID:   "shen-puer",
			Name: "Шен Пуэр",
			Schedules: []config.Schedule{
				{ID: "default", Name: "Стандарт", Pours: []int{10, 15, 20, 25}},
				{ID: "fast", Name: "Быстрый", Pours: []int{5, 10, 10, 15}},
			},
		},
	}).View()

	for _, want := range []string{
		"Варианты заваривания",
		"Стандарт:",
		"10, 15, 20, 25",
		"Быстрый:",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("view should contain %q, got %q", want, view)
		}
	}
}

func TestMenuModelViewUpdatesPreviewForSelectedTea(t *testing.T) {
	model := NewMenuModel([]config.Tea{
		{ID: "shen", Name: "Шен", Schedules: []config.Schedule{{ID: "s1", Name: "Мягкий", Pours: []int{10, 15, 20}}}},
		{ID: "shu", Name: "Шу", Schedules: []config.Schedule{{ID: "s2", Name: "Плотный", Pours: []int{15, 20, 25}}}},
	})

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	view := updated.(MenuModel).View()

	if !strings.Contains(view, "Плотный:") {
		t.Fatalf("view should switch preview with cursor, got %q", view)
	}
	if strings.Contains(view, "Мягкий:") {
		t.Fatalf("view should not keep previous preview, got %q", view)
	}
}
```

- [ ] **Step 2: Run targeted menu tests and verify RED**

Run:

```bash
go test ./internal/ui -run 'TestMenuModelViewShowsSelectedTeaSchedulePreview|TestMenuModelViewUpdatesPreviewForSelectedTea' -v
```

Expected: FAIL because the current menu view has no preview section.

- [ ] **Step 3: Implement minimal preview rendering**

Update `internal/ui/menu.go` to:

```go
func (m MenuModel) View() string {
	lines := []string{
		renderTitle("Выберите чай:"),
		"",
	}

	for i, currentTea := range m.teas {
		if i == m.cursor {
			lines = append(lines, renderSelectedLine(currentTea.Name))
		} else {
			lines = append(lines, renderNormalLine(currentTea.Name))
		}
	}

	lines = append(lines, "", renderMutedLine("─────────────"))
	lines = append(lines, renderPreviewTitle("Варианты заваривания:"))
	lines = append(lines, renderTeaPreviewLines(m.selectedTea())...)
	lines = append(lines,
		"",
		renderMutedLine("a добавить  e редакт.  d удалить"),
		"",
		renderHelpLine("j/k выбор  enter выбор расписания  esc выход"),
	)

	return renderPanel(lines...)
}
```

Add focused helpers in the same file for `selectedTea`, compact pour formatting, and preview row truncation rather than pushing tea-specific formatting into `styles.go`.

- [ ] **Step 4: Run targeted menu tests and verify GREEN**

Run:

```bash
go test ./internal/ui -run 'TestMenuModelViewShowsSelectedTeaSchedulePreview|TestMenuModelViewUpdatesPreviewForSelectedTea|TestMenuModelViewContainsTeaNamesAndHelp|TestMenuModelViewTruncatesLongTeaNames' -v
```

Expected: PASS.

- [ ] **Step 5: Commit preview slice**

```bash
git add internal/ui/menu.go internal/ui/menu_test.go
git commit -m "feat: add tea menu schedule preview"
```

### Task 2: Center Notification Content

**Files:**
- Modify: `internal/ui/notify.go`
- Modify: `internal/ui/styles.go`
- Test: `internal/ui/notify_test.go`

- [ ] **Step 1: Write failing centering tests**

Add a test in `internal/ui/notify_test.go` that checks the notify view keeps vertical padding around the content block instead of rendering the ASCII header flush to the top:

```go
func TestNotifyModelViewCentersContentBlock(t *testing.T) {
	view := NewNotifyModel("Шен Пуэр", 0, 3, false).View()
	lines := strings.Split(view, "\n")

	nonEmpty := 0
	firstContent := -1
	for i, line := range lines {
		if strings.TrimSpace(line) != "" {
			nonEmpty++
			if strings.Contains(line, "████████╗") {
				firstContent = i
				break
			}
		}
	}

	if firstContent < 3 {
		t.Fatalf("notify content should start lower inside popup, got line %d in %q", firstContent, view)
	}
}
```

- [ ] **Step 2: Run targeted notify tests and verify RED**

Run:

```bash
go test ./internal/ui -run 'TestNotifyModelViewCentersContentBlock|TestNotifyModelViewShowsActivePour|TestNotifyModelViewShowsFinishedState' -v
```

Expected: FAIL because the current view renders the block directly at the top of the bordered panel.

- [ ] **Step 3: Implement centered notification layout**

Update `internal/ui/notify.go` and `internal/ui/styles.go` to compose the notification as a single block and center it with Lip Gloss:

```go
block := lipgloss.JoinVertical(
	lipgloss.Center,
	NotifyStyle.Width(contentWidth).Render(header),
	lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Center).Render(subtext),
	lipgloss.NewStyle().Foreground(ColorMuted).Width(contentWidth).Align(lipgloss.Center).Render(help),
)

content := lipgloss.Place(
	notifyPopupWidth,
	menuPopupHeight,
	lipgloss.Center,
	lipgloss.Center,
	block,
)

return BorderStyle.Width(contentWidth).Render(content)
```

Use the same popup-height constant as the menu popup so the content is visually centered in the live popup size without changing notify behavior or keys.

- [ ] **Step 4: Run targeted notify tests and verify GREEN**

Run:

```bash
go test ./internal/ui -run 'TestNotifyModelViewCentersContentBlock|TestNotifyModelViewShowsActivePour|TestNotifyModelViewShowsFinishedState' -v
```

Expected: PASS.

- [ ] **Step 5: Commit notify slice**

```bash
git add internal/ui/notify.go internal/ui/styles.go internal/ui/notify_test.go
git commit -m "feat: center tea notifications"
```

### Task 3: Swap tmux Hotkeys

**Files:**
- Modify: `tea.tmux`
- Test: `cmd/tmux-tea/tea_tmux_test.go`

- [ ] **Step 1: Write failing binding tests**

Update `cmd/tmux-tea/tea_tmux_test.go` to assert:

```go
func TestTeaTmuxSwapsMenuAndConfirmHotkeys(t *testing.T) {
	script := readTeaTmux(t)

	for _, want := range []string{
		`bind-key t run-shell "$RUNNER confirm"`,
		`bind-key T display-popup -E -w 72 -h 28 "$RUNNER menu"`,
	} {
		if !strings.Contains(script, want) {
			t.Fatalf("tea.tmux should contain %q, got:\n%s", want, script)
		}
	}
}
```

- [ ] **Step 2: Run targeted tmux tests and verify RED**

Run:

```bash
go test ./cmd/tmux-tea -run 'TestTeaTmuxSwapsMenuAndConfirmHotkeys|TestTeaTmuxUsesLargerMenuPopup' -v
```

Expected: FAIL because the current script still binds `t` to menu and `T` to confirm.

- [ ] **Step 3: Implement swapped tmux bindings**

Update `tea.tmux` to:

```bash
tmux bind-key t run-shell "$RUNNER confirm"
tmux bind-key T display-popup -E -w 72 -h 28 "$RUNNER menu"
tmux set-option -ga status-right '#('"$RUNNER"' status)'
```

- [ ] **Step 4: Run targeted tmux tests and verify GREEN**

Run:

```bash
go test ./cmd/tmux-tea -run 'TestTeaTmuxSwapsMenuAndConfirmHotkeys|TestTeaTmuxUsesLargerMenuPopup' -v
```

Expected: PASS.

- [ ] **Step 5: Commit tmux binding slice**

```bash
git add tea.tmux cmd/tmux-tea/tea_tmux_test.go
git commit -m "feat: swap tmux tea hotkeys"
```

### Task 4: Final Verification

**Files:**
- Verify: `internal/ui/menu.go`
- Verify: `internal/ui/notify.go`
- Verify: `tea.tmux`

- [ ] **Step 1: Run full automated verification**

```bash
go test ./...
bash -n tea.tmux
bash -n scripts/run.sh
```

Expected: all commands PASS.

- [ ] **Step 2: Re-apply tmux bindings in the current session**

```bash
bash tea.tmux
```

Expected: no tmux errors; `Prefix+t` now maps to confirm and `Prefix+T` opens the menu.

- [ ] **Step 3: Commit final polish**

```bash
git add internal/ui/menu.go internal/ui/menu_test.go internal/ui/notify.go internal/ui/notify_test.go internal/ui/styles.go tea.tmux cmd/tmux-tea/tea_tmux_test.go
git commit -m "feat: polish menu preview and tea popup layout"
```
