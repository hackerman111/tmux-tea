# tmux-tea: Tmux Tea Ceremony Timer Plugin

## Overview

Плагин для tmux, реализующий таймер для чайной церемонии. Пользователь выбирает чай и расписание проливов через TUI-интерфейс, запускает таймер, и получает уведомления о каждом проливе.

## Tech Stack

- **Bash** — тонкая обёртка (`tea.tmux`) для регистрации хоткеев в tmux
- **Go** — вся логика: CLI, TUI, таймеры, хранение данных
- **Bubbletea/Lipgloss** — TUI-фреймворк для интерактивного интерфейса
- **JSON** — хранение данных о чаях и расписаниях (`~/.config/tmux-tea/teas.json`)

## Architecture

Монолитный Go-бинарник `tmux-tea` с субкомандами. Без демона — фоновый процесс запускается только на время активного таймера.

### Project Structure

```
tmux-tea/
├── tea.tmux                  # Bash: точка входа для TPM, регистрация хоткеев
├── scripts/
│   └── install.sh            # Сборка Go-бинарника
├── cmd/
│   └── tmux-tea/
│       └── main.go           # Точка входа, субкоманды (cobra)
├── internal/
│   ├── config/
│   │   └── config.go         # Загрузка/сохранение JSON, структуры данных
│   ├── timer/
│   │   └── timer.go          # Логика таймера, управление фоновым процессом
│   ├── ui/
│   │   ├── menu.go           # Главное меню (выбор чая)
│   │   ├── schedule.go       # Выбор/создание расписания
│   │   ├── editor.go         # Редактирование чая/расписания (формы)
│   │   ├── notify.go         # Экран "TEA TIME" с ASCII-артом
│   │   └── styles.go         # Lipgloss стили, цвета, тема
│   └── tmux/
│       └── tmux.go           # Обёртка над tmux CLI (display-popup, set-option, bell)
├── go.mod
├── go.sum
└── README.md
```

### Subcommands

| Команда | Описание |
|---------|----------|
| `tmux-tea menu` | Запускает TUI в tmux popup (выбор чая, расписания, CRUD) |
| `tmux-tea start --tea <id> --schedule <id>` | Запускает фоновый таймер |
| `tmux-tea status` | Выводит строку для tmux status bar |
| `tmux-tea confirm` | Подтверждение пролива, переход к следующему |
| `tmux-tea stop` | Остановка текущей сессии |

## Data Model

### Storage: `~/.config/tmux-tea/teas.json`

```json
{
  "teas": [
    {
      "id": "shen-puer",
      "name": "Шен Пуэр",
      "schedules": [
        {
          "id": "default",
          "name": "Стандарт",
          "pours": [10, 15, 20, 25, 30, 40, 50, 60]
        },
        {
          "id": "fast",
          "name": "Быстрый",
          "pours": [5, 10, 10, 15, 15, 20]
        }
      ]
    }
  ]
}
```

### Go Types

```go
type Tea struct {
    ID        string     `json:"id"`
    Name      string     `json:"name"`
    Schedules []Schedule `json:"schedules"`
}

type Schedule struct {
    ID    string `json:"id"`
    Name  string `json:"name"`
    Pours []int  `json:"pours"` // seconds
}

type Config struct {
    Teas []Tea `json:"teas"`
}
```

- `pours` — массив int, значения в секундах
- `id` генерируется из имени (slugify) при создании
- Первый запуск создаёт файл с 3 дефолтными чаями

### Timer State: `/tmp/tmux-tea-state.json`

```json
{
  "pid": 12345,
  "tea_name": "Шен Пуэр",
  "pour_index": 2,
  "total_pours": 8,
  "remaining_sec": 7,
  "started_at": "2026-04-03T15:30:00Z",
  "status": "counting"
}
```

`status`: `"counting"` | `"ready"` (пролив готов, ждёт подтверждения) | отсутствие файла = нет активной сессии.

## User Flow

```
1. Хоткей 1 (prefix + t)
   → tmux display-popup "tmux-tea menu"
   → TUI: список чаёв → выбор → список расписаний → выбор
   → TUI вызывает: tmux-tea start --tea <id> --schedule <id>
   → popup закрывается

2. Go-процесс запускается в фоне (tmux run-shell -b)
   → пишет состояние в /tmp/tmux-tea-state.json
   → tmux status bar показывает countdown

3. Таймер истёк
   → tmux display-popup "tmux-tea notify" + bell
   → ASCII-арт "TEA TIME" на экране

4. Хоткей 2 (prefix + T)
   → tmux-tea confirm
   → popup закрывается
   → таймер переходит к следующему проливу
   → цикл повторяется

5. После последнего пролива
   → popup "Чаепитие завершено!"
   → state очищается, status bar возвращается к обычному
```

## TUI Screens

### Main Menu (tmux-tea menu)

```
┌─ Главное меню ──────────────┐
│                              │
│  Выберите чай:               │
│                              │
│  > Шен Пуэр                 │
│    Шу Пуэр                  │
│    Да Хун Пао               │
│    ─────────────             │
│    + Добавить чай            │
│    e Редактировать           │
│                              │
│  j/k выбор  enter старт     │
└──────────────────────────────┘
```

### Schedule Selection

```
┌─ Расписание ────────────────┐
│                              │
│  Шен Пуэр — расписание:     │
│                              │
│  > Стандарт (10,15,20,25..) │
│    Быстрый  (5,10,10,15..)  │
│    ─────────────             │
│    + Новое расписание        │
│    e Редактировать           │
│                              │
│  esc назад  enter старт     │
└──────────────────────────────┘
```

### Tea Editor

```
┌─ Редактор чая ──────────────┐
│                              │
│  Название: [Шен Пуэр     ]  │
│                              │
│  tab поле  enter сохранить  │
└──────────────────────────────┘
```

### Schedule Editor

```
┌─ Редактор расписания ───────┐
│                              │
│  Название: [Стандарт      ]  │
│  Проливы (сек):              │
│  [10] [15] [20] [25] [30]   │
│  [40] [50] [60]             │
│  + добавить   x удалить     │
│                              │
│  tab поле  enter сохранить  │
└──────────────────────────────┘
```

### TEA TIME Notification

```
┌──────────────────────────────┐
│                              │
│  ████████╗███████╗ █████╗   │
│  ╚══██╔══╝██╔════╝██╔══██╗  │
│     ██║   █████╗  ███████║  │
│     ██║   ██╔══╝  ██╔══██║  │
│     ██║   ███████╗██║  ██║  │
│     ╚═╝   ╚══════╝╚═╝  ╚═╝ │
│                              │
│     Шен Пуэр · пролив 3/8   │
│     Нажмите Enter            │
│                              │
└──────────────────────────────┘
```

### Navigation

| Клавиша | Действие |
|---------|----------|
| `↑↓` / `j/k` | Перемещение по списку |
| `Enter` | Выбор / подтверждение |
| `Esc` | Назад / отмена |
| `d` | Удалить (с подтверждением) |
| `e` | Редактировать |

## Tmux Integration

### tea.tmux (TPM entry point)

```bash
#!/usr/bin/env bash
CURRENT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEA_BIN="$CURRENT_DIR/bin/tmux-tea"

# Hotkey 1: main menu
tmux bind-key t display-popup -E -w 40 -h 15 "$TEA_BIN menu"

# Hotkey 2: confirm pour
tmux bind-key T run-shell "$TEA_BIN confirm"

# Status bar integration
tmux set-option -ga status-right '#($TEA_BIN status)'
```

### Status Bar Format

| Состояние | Отображение |
|-----------|------------|
| Активный таймер | `🍵 Шен Пуэр [3/8] 0:12` |
| Ожидание пролива | `🍵 Шен Пуэр [3/8] READY` |
| Нет сессии | (пусто) |

### Status Interval

Плагин устанавливает `status-interval 1` при активном таймере для точного countdown, и восстанавливает оригинальное значение после завершения.

### Popup Sizes

| Экран | Размер |
|-------|--------|
| Меню | 40x15 |
| TEA TIME notify | 50x15 |
| Редакторы | 50x20 |

## Error Handling

| Ситуация | Поведение |
|----------|-----------|
| Таймер уже запущен, выбран новый чай | Подтверждение: "Остановить текущую сессию?" |
| Popup закрыт Esc вместо confirm | Таймер остаётся в READY, status bar показывает READY |
| Фоновый процесс убит | `tmux-tea status` обнаруживает orphaned state (pid мёртв) → очищает файл |
| Первый запуск, нет конфига | Создаёт `teas.json` с 3 дефолтными чаями |
| Невалидный JSON | Ошибка в stderr, предложение сбросить к дефолту |
| Удаление последнего чая | Запрещено: "Нельзя удалить последний чай" |

## Dependencies

- **Go 1.21+** — для сборки
- **tmux 3.3+** — для `display-popup` (поддержка с tmux 3.3)
- **charmbracelet/bubbletea** — TUI framework
- **charmbracelet/lipgloss** — стилизация
- **spf13/cobra** — CLI субкоманды
